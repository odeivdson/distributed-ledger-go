package worker

import (
	"context"
	"log/slog"
	"sync"
)

type Job struct {
	NotificationID string
	Payload        []byte
	Target         string
}

type WorkerPool struct {
	maxWorkers int
	jobChannel chan Job
	wg         sync.WaitGroup
	handler    func(ctx context.Context, job Job) error
}

func NewWorkerPool(maxWorkers int, bufferSize int, handler func(ctx context.Context, job Job) error) *WorkerPool {
	return &WorkerPool{
		maxWorkers: maxWorkers,
		jobChannel: make(chan Job, bufferSize),
		handler:    handler,
	}
}

func (wp *WorkerPool) Start(ctx context.Context) {
	for i := 1; i <= wp.maxWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	slog.Info("Worker iniciado", "worker_id", id)

	for {
		select {
		case job, ok := <-wp.jobChannel:
			if !ok {
				return
			}
			err := wp.handler(ctx, job)
			if err != nil {
				slog.Error("Erro ao processar job", "worker_id", id, "notification_id", job.NotificationID, "error", err)
			}
		case <-ctx.Done():
			slog.Info("Encerrando worker por contexto", "worker_id", id)
			return
		}
	}
}

func (wp *WorkerPool) Submit(job Job) {
	wp.jobChannel <- job
}

func (wp *WorkerPool) Shutdown() {
	close(wp.jobChannel)
	wp.wg.Wait()
	slog.Info("Todos os workers do pool foram finalizados.")
}
