package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/assaidy/blink/app/db"
	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/workers"
	"github.com/charmbracelet/log"
)

var (
	queries = repo.New(db.GetPool())
)

func main() {
	logger := slog.New(log.NewWithOptions(os.Stderr,
		log.Options{ReportTimestamp: true},
	))

	workerManager := workers.NewWorkerManager(
		workers.WithLogger(logger),
	)

	workerManager.RegisterWorker(
		workers.NewWorker("clean expired sessions", cleanExpiredSessionsJob,
			workers.WithTick(1*time.Hour),
			workers.WithTimeout(10*time.Minute),
			workers.WithNRetries(2),
			workers.WithRetryDelay(30*time.Second),
		),
	)
	workerManager.RegisterWorker(
		workers.NewWorker("clean expired OTPs", cleanExpiredOtpsJob,
			workers.WithTick(5*time.Minute),
			workers.WithTimeout(30*time.Second),
			workers.WithNRetries(3),
			workers.WithRetryDelay(10*time.Second),
		),
	)
	workerManager.RegisterWorker(
		workers.NewWorker("clean deleted chat messages", cleanDeletedChatMessagesJob,
			workers.WithSchedules(workers.DailyAt(2, 0)),
			workers.WithTimeout(15*time.Minute),
			workers.WithNRetries(2),
			workers.WithRetryDelay(1*time.Minute),
		),
	)

	workerManager.Start()
}

func cleanExpiredSessionsJob(ctx context.Context, _ *slog.Logger) error {
	return queries.BatchDeleteExpriredSessions(ctx)
}

func cleanExpiredOtpsJob(ctx context.Context, _ *slog.Logger) error {
	return queries.BatchDeleteExpiredOtps(ctx)
}

func cleanDeletedChatMessagesJob(ctx context.Context, _ *slog.Logger) error {
	return queries.BatchDeleteChatMessages(ctx)
}
