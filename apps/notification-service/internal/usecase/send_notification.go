package usecase

import (
	"context"
	"log/slog"
	"math/rand"
	"notification-service/internal/ports"
	"notification-service/internal/worker"
	"time"
)

type NotificationUseCase struct {
	provider ports.NotificationProvider
}

func NewNotificationUseCase(p ports.NotificationProvider) *NotificationUseCase {
	return &NotificationUseCase{
		provider: p,
	}
}

func (uc *NotificationUseCase) HandleJob(ctx context.Context, job worker.Job) error {
	maxRetries := 3
	baseDelay := 1 * time.Second

	for i := 0; i <= maxRetries; i++ {
		err := uc.provider.Send(ctx, job.Target, job.Payload)
		if err == nil {
			slog.Info("Notificação enviada com sucesso", "notification_id", job.NotificationID)
			return nil
		}

		if i == maxRetries {
			slog.Error("Falha definitiva ao enviar notificação após retentativas", 
				"notification_id", job.NotificationID, "error", err)
			return err
		}

		// Exponential Backoff com Jitter
		delay := time.Duration(1<<uint(i)) * baseDelay
		jitter := time.Duration(rand.Int63n(int64(delay / 2)))
		totalDelay := delay + jitter

		slog.Warn("Falha ao enviar notificação, agendando retentativa", 
			"notification_id", job.NotificationID, 
			"tentativa", i+1, 
			"atraso", totalDelay, 
			"error", err)

		select {
		case <-time.After(totalDelay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
