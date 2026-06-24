package main

//go:generate swag init -g main.go -d .,../../internal -o ../../docs

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"shared/config"
	"shared/health"
	sharedLogger "shared/logger"
	"shared/metrics"
	"shared/validation"
	"syscall"
	"time"
	_ "transaction-gw/docs"
	httpAdapter "transaction-gw/internal/adapters/http"
	"transaction-gw/internal/adapters/kafka"
	"transaction-gw/internal/adapters/postgres"
	"transaction-gw/internal/usecase"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	httpSwagger "github.com/swaggo/http-swagger"

	rateLimiterHttp "rate-limiter/adapters/http"
	rateLimiterDomain "rate-limiter/domain"
	rateLimiterLocal "rate-limiter/adapters/local"
	rateLimiterPorts "rate-limiter/ports"
	rateLimiterRedis "rate-limiter/adapters/redis"
)

// @title Transaction Gateway API
// @version 1.0
// @description Gateway de entrada para transações financeiras do Distributed Ledger.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

func main() {
	config.Load()
	sharedLogger.Init(config.GetEnv("LOG_LEVEL", "info"))
	slog.Info("Iniciando Transaction Gateway...")

	// Contexto para Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Inicializar Métricas Prometheus
	metricsRegistry, err := metrics.NewMetricsRegistry("transaction_gw")
	if err != nil {
		slog.Error("Falha ao inicializar métricas Prometheus", "error", err)
		os.Exit(1)
	}
	metricsMiddleware := metrics.NewMetricsMiddleware(metricsRegistry)

	// Inicializar Adaptadores
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := config.GetEnv("KAFKA_TOPIC", "transactions")

	broker := kafka.NewProducer([]string{kafkaBrokers}, kafkaTopic, metricsRegistry)
	defer broker.Close()

	// 1. Configurar Redis para Rate Limiter
	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// 2. Inicializar Rate Limiter (Projeto 2)
	redisLimiter := rateLimiterRedis.NewRedisRateLimiter(rdb)
	localLimiter := rateLimiterLocal.NewInMemoryRateLimiter()
	adaptiveLimiter := rateLimiterPorts.NewAdaptiveRateLimiter(redisLimiter, localLimiter)

	rlConfig := rateLimiterDomain.RateLimitConfig{
		Limit:  100, // 100 requisições
		Window: 60,  // por minuto
	}
	rlMiddleware := rateLimiterHttp.NewMiddleware(adaptiveLimiter, rlConfig)

	// 3. Configurar Postgres para Criação de Contas
	dbURL := config.GetEnv("DATABASE_URL", "postgres://staff_eng:staff_pwd@localhost:5432/ledger_db?sslmode=disable")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Falha ao abrir conexão com Postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Falha ao pingar Postgres", "error", err)
		os.Exit(1)
	}

	if err := postgres.RunMigrations(context.Background(), db); err != nil {
		slog.Error("Falha ao executar migrations", "error", err)
		os.Exit(1)
	}

	// Inicializar Repositórios
	accountRepo := postgres.NewAccountRepository(db)
	idempotencyRepo := postgres.NewIdempotencyRepository(db)
	transactionRepo := postgres.NewTransactionRepository(db, metricsRegistry)

	// Inicializar Casos de Uso
	submitTxUseCase := usecase.NewSubmitTransactionUseCase(broker)
	createAccountUseCase := usecase.NewCreateAccountUseCase(accountRepo)
	getTransactionUseCase := usecase.NewGetTransactionUseCase(transactionRepo)

	// Inicializar Validadores
	transactionValidator := validation.NewTransactionRequestValidator()
	accountValidator := validation.NewAccountRequestValidator()

	// Inicializar Handlers
	txHandler := httpAdapter.NewTransactionHandler(submitTxUseCase)
	accountHandler := httpAdapter.NewAccountHandler(createAccountUseCase)
	getTransactionHandler := httpAdapter.NewGetTransactionHandler(getTransactionUseCase)

	// Inicializar Middleware de Idempotência
	idempotencyMiddleware := httpAdapter.NewIdempotencyMiddleware(idempotencyRepo)



	// Servidor HTTP simples
	mux := http.NewServeMux()
	
	// GET /health - Health check
	// @Summary Health check
	// @Description Verifica se o serviço está funcionando corretamente
	// @Tags health
	// @Produce json
	// @Success 200 {object} map[string]string "Serviço saudável"
	// @Router /health [get]
	mux.HandleFunc("/health", health.Handler("transaction-gateway"))
	
	// POST /transactions com validação + idempotência + métricas + rate limit
	txValidatingHandler := httpAdapter.NewValidatingHandler(
		transactionValidator,
		http.HandlerFunc(txHandler.Handle),
	)
	mux.Handle("/transactions", 
		metricsMiddleware.Middleware(
			idempotencyMiddleware.Middleware(
				rlMiddleware.RateLimit(txValidatingHandler))))
	
	// POST /accounts com validação
	accountValidatingHandler := httpAdapter.NewValidatingHandler(
		accountValidator,
		http.HandlerFunc(accountHandler.Create),
	)
	mux.Handle("/accounts", accountValidatingHandler)
	
	// GET endpoint com métricas
	mux.HandleFunc("GET /transactions/{id}", 
		func(w http.ResponseWriter, r *http.Request) {
			metricsMiddleware.Middleware(http.HandlerFunc(getTransactionHandler.GetByID)).ServeHTTP(w, r)
		})
	
	mux.Handle("/swagger/", httpSwagger.WrapHandler)
	mux.Handle("/metrics", promhttp.Handler())
	
	// Servir OpenAPI spec completo
	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		// OpenAPI spec será servido daqui
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		slog.Info("Servidor HTTP iniciado na porta :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Falha ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Gateway pronto para receber transações")

	<-ctx.Done()

	slog.Info("Encerrando Gateway...", "motivo", ctx.Err())

	// Graceful shutdown timeout
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("Gateway finalizado com sucesso")
}
