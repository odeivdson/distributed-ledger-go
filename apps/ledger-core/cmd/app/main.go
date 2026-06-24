package main

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
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ledger-core/internal/adapters/kafka"
	"ledger-core/internal/adapters/postgres"
	"ledger-core/internal/usecase"
	"ledger-core/internal/worker"

	_ "github.com/lib/pq"
)

func main() {
	config.Load()
	sharedLogger.Init(config.GetEnv("LOG_LEVEL", "info"))
	slog.Info("Iniciando Ledger Core...")

	// 1. Configurar Infraestrutura (Adapters)
	dbURL := config.GetEnv("DATABASE_URL", "postgres://staff_eng:super_secret_password@localhost:5432/ledger_db?sslmode=disable")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Falha ao abrir conexão com o banco", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Configurações de pool de conexões
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		slog.Error("Falha ao pingar o banco", "error", err)
		os.Exit(1)
	}

	// Executar migrações automaticamente
	if err := postgres.RunMigrations(context.Background(), db); err != nil {
		slog.Error("Falha ao executar migrações", "error", err)
		os.Exit(1)
	}

	// Inicializar Métricas Prometheus
	metricsRegistry, err := metrics.NewMetricsRegistry("ledger_core")
	if err != nil {
		slog.Error("Falha ao inicializar métricas Prometheus", "error", err)
		os.Exit(1)
	}

	uow := postgres.NewUnitOfWork(db)
	repo := postgres.NewLedgerRepository(db, metricsRegistry)



	// 2. Inicializar Adaptadores de Saída (Kafka)
	kafkaBrokersStr := config.GetEnv("KAFKA_BROKERS", "localhost:9092")
	notificationTopic := config.GetEnv("NOTIFICATION_TOPIC", "notifications")

	notifier := kafka.NewNotifier([]string{kafkaBrokersStr}, notificationTopic, metricsRegistry)
	defer notifier.Close()

	// 3. Injetar dependências no Caso de Uso (Core)
	ledgerUC := usecase.NewLedgerUseCase(repo, repo, repo, uow)

	// 3.1. Inicializar Outbox Worker
	outboxWorker := worker.NewOutboxWorker(
		repo,
		notifier,
		500*time.Millisecond,
		100,
		notificationTopic,
		5,
	)

	// 4. Inicializar Adaptador de Entrada (Kafka Consumer)
	kafkaTopic := config.GetEnv("KAFKA_TOPIC", "transactions")
	kafkaGroupID := config.GetEnv("KAFKA_GROUP_ID", "ledger-core-group")

	consumer := kafka.NewConsumer([]string{kafkaBrokersStr}, kafkaTopic, kafkaGroupID, ledgerUC)
	defer consumer.Close()

	// 5. Servidor de Health Check e Métricas (Side-car)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.Handler("ledger-core"))
	healthMux.Handle("/metrics", promhttp.Handler())
	healthServer := &http.Server{
		Addr:    ":8082",
		Handler: healthMux,
	}

	go func() {
		slog.Info("Health Check iniciado na porta :8082")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Falha ao iniciar health check", "error", err)
		}
	}()

	// 6. Orquestrar a execução com Graceful Shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go outboxWorker.Start(ctx)

	go func() {
		if err := consumer.Consume(ctx); err != nil {
			slog.Error("Erro fatal no consumidor Kafka", "error", err)
			stop()
		}
	}()

	slog.Info("Ledger Core pronto para processar transações")

	<-ctx.Done()

	slog.Info("Encerrando Ledger Core...", "motivo", ctx.Err())

	// Graceful shutdown timeout
	ctxShut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := healthServer.Shutdown(ctxShut); err != nil {
		slog.Error("Erro ao encerrar health check server", "error", err)
	}

	slog.Info("Ledger Core finalizado com sucesso")
}
