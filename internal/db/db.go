package db

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Conn opens a pgx connection pool from the DB_* environment variables and
// verifies it with a ping. The caller owns the pool and must Close it.
func Conn(ctx context.Context) (*pgxpool.Pool, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	// Local Postgres has no TLS; set DB_SSLMODE=require (or verify-full) for RDS.
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}

	// Build a URL rather than a key=value string so special characters in the
	// password are escaped correctly.
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, port),
		Path:     "/" + dbname,
		RawQuery: url.Values{"sslmode": {sslmode}}.Encode(),
	}

	cfg, err := pgxpool.ParseConfig(dsn.String())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}

	cfg.MaxConns = 10
	cfg.MinConns = 2
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute
	cfg.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
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
