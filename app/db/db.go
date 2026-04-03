package db

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/assaidy/blink/app/config"
	_ "github.com/lib/pq"
)

var (
	pool *sql.DB
	once sync.Once
)

func Get() *sql.DB {
	once.Do(func() {
		var err error
		if pool, err = sql.Open("postgres", config.Get().PostgresUrl); err != nil {
			panic(fmt.Sprintf("failed to open db connection: %v", err))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := pool.PingContext(ctx); err != nil {
			panic(fmt.Sprintf("couldn't ping db: %v", err))
		}

		pool.SetMaxOpenConns(20)
		pool.SetMaxIdleConns(20)
		pool.SetConnMaxLifetime(20 * time.Minute)
		pool.SetConnMaxIdleTime(5 * time.Minute)
	})

	return pool
}
