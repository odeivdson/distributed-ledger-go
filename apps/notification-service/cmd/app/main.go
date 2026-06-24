package main

import (
	"context"
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

	"notification-service/internal/adapters/kafka"
	httpAdapter "notification-service/internal/adapters/http"
	"notification-service/internal/usecase"
	"notification-service/internal/worker"
)

func main() {
	config.Load()
	sharedLogger.Init(config.GetEnv("LOG_LEVEL", "info"))
	slog.Info("Iniciando Notification Service...")

	// 1. Configurar Adapters
	provider := httpAdapter.NewNotificationProvider()

	// 2. Configurar Casos de Uso
	nc := usecase.NewNotificationUseCase(provider)

	// 3. Inicializar Worker Pool
	// maxWorkers=10, bufferSize=100
	wp := worker.NewWorkerPool(10, 100, nc.HandleJob)

	// 4. Inicializar Consumidor Kafka
	kafkaBrokers := config.GetEnv("KAFKA_BROKERS", "localhost:9092")
	kafkaTopic := config.GetEnv("NOTIFICATION_TOPIC", "notifications")
	kafkaGroupID := config.GetEnv("KAFKA_GROUP_ID", "notification-service-group")

	consumer := kafka.NewConsumer([]string{kafkaBrokers}, kafkaTopic, kafkaGroupID, wp)
	defer consumer.Close()

	// 5. Servidor de Health Check (Side-car)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", health.Handler("notification-service"))
	healthMux.Handle("/metrics", promhttp.Handler())
	healthServer := &http.Server{
		Addr:    ":8083",
		Handler: healthMux,
	}

	go func() {
		slog.Info("Health Check iniciado na porta :8083")
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Falha ao iniciar health check", "error", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Iniciar Worker Pool
	wp.Start(ctx)
	slog.Info("Worker Pool iniciado", "workers", 10)

	// Iniciar Consumo
	go func() {
		if err := consumer.Consume(ctx); err != nil {
			slog.Error("Erro fatal no consumidor Kafka", "error", err)
			stop()
		}
	}()

	<-ctx.Done()
	slog.Info("Encerrando Notification Service...")

	// Graceful shutdown do health check
	ctxShut, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	healthServer.Shutdown(ctxShut)

	wp.Shutdown()
	slog.Info("Notification Service encerrado com sucesso.")
}
