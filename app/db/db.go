package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/assaidy/blink/app/env"
	_ "github.com/lib/pq"
)

var pool *sql.DB

func GetPool() *sql.DB {
	if pool != nil {
		return pool
	}

	var err error
	if pool, err = sql.Open("postgres", env.DBUrl); err != nil {
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

	return pool
}
