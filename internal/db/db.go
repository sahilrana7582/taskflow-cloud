package db

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
)

// Conn opens a pgx connection pool from the DB_* environment variables and
// verifies it with a ping. The caller owns the pool and must Close it.
func Conn(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	host := cfg.Database.Host
	port := cfg.Database.Port
	user := cfg.Database.User
	password := cfg.Database.Password
	dbname := cfg.Database.Name

	if host == "" || port == 0 || user == "" || password == "" || dbname == "" {
		return nil, fmt.Errorf("database configuration is incomplete")
	}

	// Local Postgres has no TLS; set DB_SSLMODE=require (or verify-full) for RDS.
	sslmode := cfg.Database.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}

	// Build a URL rather than a key=value string so special characters in the
	// password are escaped correctly.
	dsn := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     net.JoinHostPort(host, strconv.Itoa(port)),
		Path:     "/" + dbname,
		RawQuery: url.Values{"sslmode": {sslmode}}.Encode(),
	}

	dbCfg, err := pgxpool.ParseConfig(dsn.String())
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
