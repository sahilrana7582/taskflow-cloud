package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
)

// Conn opens a pgx connection pool to cfg.Database and verifies it with a
// ping. The caller owns the pool and must Close it.
func Conn(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	d := cfg.Database
	if d.Host == "" || d.Port == 0 || d.User == "" || d.Password == "" || d.Name == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	dbCfg, err := pgxpool.ParseConfig(d.URL())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	dbCfg.MaxConns = 10
	dbCfg.MinConns = 2
	dbCfg.MaxConnLifetime = time.Hour
	dbCfg.MaxConnIdleTime = 30 * time.Minute
	dbCfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, dbCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	// pgxpool connects lazily, so ping now to fail fast on bad credentials or
	// an unreachable host instead of on the first request.
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return pool, nil
}
