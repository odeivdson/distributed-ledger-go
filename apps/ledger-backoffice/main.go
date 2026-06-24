package main

import (
	"context"
	"database/sql"
	"ledger-backoffice/handlers"
	"ledger-backoffice/repository"
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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
)

func main() {
	config.Load()
	sharedLogger.Init(config.GetEnv("LOG_LEVEL", "info"))

	dbURL := config.GetEnv("DATABASE_URL", "postgres://staff_eng:staff_pwd@localhost:5432/ledger_db?sslmode=disable")
	redisAddr := config.GetEnv("REDIS_ADDR", "localhost:6379")
	port := config.GetEnv("PORT", "8081")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Falha ao abrir banco de dados", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		slog.Error("Falha ao conectar ao banco de dados", "error", err)
		os.Exit(1)
	}
	slog.Info("Conectado ao PostgreSQL com sucesso")

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Warn("Falha ao conectar ao Redis", "error", err)
	} else {
		slog.Info("Conectado ao Redis com sucesso")
	}

	repo := repository.NewReaderRepository(db, rdb)
	backofficeHandler := handlers.NewBackofficeHandler(repo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", health.Handler("ledger-backoffice"))
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/backoffice", func(r chi.Router) {
		r.Get("/dashboard", backofficeHandler.Dashboard)
		r.Get("/accounts/{id}", backofficeHandler.AccountDetail)
		r.Get("/accounts/search", backofficeHandler.AccountDetail)
		r.Get("/dlq", backofficeHandler.DLQ)
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/backoffice/dashboard", http.StatusMovedPermanently)
	})

	server := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		slog.Info("Servidor Backoffice iniciado", "port", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Falha ao iniciar servidor", "error", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Encerrando servidor...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Falha no graceful shutdown", "error", err)
	}
	slog.Info("Servidor encerrado com sucesso")
}
