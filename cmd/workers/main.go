package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/assaidy/blink/app/config"
	"github.com/assaidy/blink/app/db"
	"github.com/assaidy/blink/app/repo"
	"github.com/assaidy/workers"
	"github.com/charmbracelet/log"
)

var (
	conf    = config.Load()
	queries = repo.New(db.GetPostgresConnectionPool(conf.DBUrl))
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
			workers.WithRetryDelay(30*time.Second),
		),
	)
	workerManager.RegisterWorker(
		workers.NewWorker("clean expired OTPs", cleanExpiredOtpsJob,
			workers.WithTick(5*time.Minute),
			workers.WithTimeout(30*time.Second),
			workers.WithRetryDelay(10*time.Second),
		),
	)
	workerManager.RegisterWorker(
		workers.NewWorker("clean deleted chat messages", cleanDeletedChatMessagesJob,
			workers.WithSchedules(workers.DailyAt(2, 0)),
			workers.WithTimeout(15*time.Minute),
			workers.WithRetryDelay(1*time.Minute),
		),
	)

	workerManager.Start()
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
