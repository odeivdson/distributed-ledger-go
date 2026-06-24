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
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"ledger-reconciler/internal/adapters/postgres"
	"ledger-reconciler/internal/usecase"

	_ "github.com/lib/pq"
)

func main() {
	config.Load()
	sharedLogger.Init(config.GetEnv("LOG_LEVEL", "info"))
	slog.Info("Iniciando Ledger Reconciler (Batch Job)...")

	dbURL := config.GetEnv("DATABASE_URL", "postgres://staff_eng:super_secret_password@localhost:5432/ledger_db?sslmode=disable")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Falha ao abrir conexão com o banco", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Falha ao pingar o banco", "error", err)
		os.Exit(1)
	}

	repo := postgres.NewReconcilerRepository(db)
	uc := usecase.NewReconcileUseCase(repo)

	// 5. Servidor de Health Check (Side-car)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.Handler("ledger-reconciler"))
	healthMux.Handle("/metrics", promhttp.Handler())
	healthServer := &http.Server{
		Addr:    ":8084",
		Handler: healthMux,
	}

	go func() {
		slog.Info("Health Check iniciado na porta :8084")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Falha ao iniciar health check", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Configurações do Job
	chunkSize := 1000
	maxParallel := 10

	start := time.Now()
	if err := uc.RunBatch(ctx, chunkSize, maxParallel); err != nil {
		slog.Error("Erro fatal durante a reconciliação", "error", err)
		os.Exit(1)
	}

	slog.Info("Reconciliação concluída com sucesso", "duration", time.Since(start))

	// Graceful shutdown do health check
	ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	healthServer.Shutdown(ctxShut)
}
