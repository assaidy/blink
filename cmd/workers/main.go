package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/assaidy/blink/app/db"
	"github.com/assaidy/blink/app/repo"
	"github.com/charmbracelet/log"
)

func main() {
	workers := []Worker{
		{
			name:       "clean expired sessions",
			job:        cleanExpiredSessionsJob,
			tick:       24 * time.Hour,
			timeout:    30 * time.Minute,
			retries:    5,
			retryDelay: 1 * time.Minute,
		},
		{
			name:       "clean expired OTPs",
			job:        cleanExpiredOtpsJob,
			tick:       12 * time.Hour,
			timeout:    30 * time.Minute,
			retries:    5,
			retryDelay: 1 * time.Minute,
		},
		{
			name:       "clean deleted chat messages",
			job:        cleanDeletedChatMessagesJob,
			tick:       1 * time.Hour,
			timeout:    30 * time.Minute,
			retries:    5,
			retryDelay: 1 * time.Minute,
		},
	}

	quitCtx, quitCtxCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	workersWg := sync.WaitGroup{}
	defer workersWg.Wait()
	for _, w := range workers {
		workersWg.Go(func() {
			w.start(quitCtx)
		})
	}

	<-quitCtx.Done()
	quitCtxCancel()
	logger.Info("gracefully stopping workers. press Ctrl-c to force stop.")
}

var (
	logger  = log.NewWithOptions(os.Stderr, log.Options{ReportTimestamp: true})
	queries = repo.New(db.GetPool())
)

type Worker struct {
	name       string
	job        func(ctx context.Context) error
	tick       time.Duration
	timeout    time.Duration
	retries    uint
	retryDelay time.Duration
}

func (me Worker) start(workerCtx context.Context) {
	logger.Info("worker started", "worker", me.name)
	defer logger.Info("worker stopped", "worker", me.name)

	ticker := time.NewTicker(me.tick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("job started", "worker", me.name)

			jobCtx, jobCtxCancel := context.WithTimeout(context.Background(), me.timeout)
			if err := me.job(jobCtx); err != nil {
				logger.Error("job failed", "worker", me.name, "error", err)

				for i := uint(1); i <= me.retries; i++ {
					<-time.After(me.retryDelay)

					logger.Error("job retry started", "worker", me.name, "retry", i)
					if err := me.job(jobCtx); err != nil {
						logger.Error("job retry failed", "worker", me.name, "retry", i)
						continue
					}
					logger.Error("job retry finished successfully", "worker", me.name, "retry", i)
				}
			} else {
				logger.Info("job finished successfully", "worker", me.name)
			}
			jobCtxCancel()

		case <-workerCtx.Done():
			return
		}
	}
}

func cleanExpiredSessionsJob(ctx context.Context) error {
	return queries.BatchDeleteExpriredSessions(ctx)
}

func cleanExpiredOtpsJob(ctx context.Context) error {
	return queries.BatchDeleteExpiredOtps(ctx)
}

func cleanDeletedChatMessagesJob(ctx context.Context) error {
	return queries.BatchDeleteChatMessages(ctx)
}
