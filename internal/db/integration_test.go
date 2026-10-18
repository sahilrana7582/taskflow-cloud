package db

import (
	"cmp"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
)

// Integration tests run against a real PostgreSQL server and are skipped
// unless TASKFLOW_INTEGRATION=1. Each test creates its own throwaway database,
// so the database named by TASKFLOW_DATABASE_NAME is only used as a place to
// run CREATE DATABASE from and its data is never touched.
//
// Locally (connection settings come from ../../.env):
//
//	TASKFLOW_INTEGRATION=1 go test -v -count=1 -run Integration ./internal/db

func TestIntegrationConn(t *testing.T) {
	cfg := newTestDatabase(t)

	pool, err := Conn(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	t.Cleanup(pool.Close)

	var got string
	if err := pool.QueryRow(t.Context(), "SELECT current_database()").Scan(&got); err != nil {
		t.Fatalf("query current database: %v", err)
	}
	if got != cfg.Database.Name {
		t.Fatalf("connected to database %q, want %q", got, cfg.Database.Name)
	}
}

func TestIntegrationMigrations(t *testing.T) {
	cfg := newTestDatabase(t)
	migrationsDir := cfg.Database.Migrations.Path

	if err := RunMigrations(cfg); err != nil {
		t.Fatalf("RunMigrations on empty database: %v", err)
	}
	// The app migrates on every start, so a second run must be a no-op.
	if err := RunMigrations(cfg); err != nil {
		t.Fatalf("RunMigrations on migrated database: %v", err)
	}

	sqlDB, err := sql.Open("postgres", cfg.Database.URL())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })

	latest := latestMigrationVersion(t, migrationsDir)
	requireDBVersion(t, sqlDB, latest)

	// Rolling everything back and migrating up again proves every Down section
	// fully undoes its Up: a leftover table would make CREATE TABLE fail.
	if err := goose.Reset(sqlDB, migrationsDir); err != nil {
		t.Fatalf("roll back all migrations: %v", err)
	}
	requireDBVersion(t, sqlDB, 0)

	if err := RunMigrations(cfg); err != nil {
		t.Fatalf("RunMigrations after rollback: %v", err)
	}
	requireDBVersion(t, sqlDB, latest)
}

func TestIntegrationUsersConstraints(t *testing.T) {
	pool := newMigratedPool(t)
	ctx := t.Context()

	insertUser := func(email string) error {
		_, err := pool.Exec(ctx,
			`INSERT INTO users (email, password_hash, full_name) VALUES ($1, 'hash', 'Ada Lovelace')`,
			email)
		return err
	}

	if err := insertUser("ada@example.com"); err != nil {
		t.Fatalf("insert valid user: %v", err)
	}

	// 23514 = check_violation, 23505 = unique_violation.
	requirePgError(t, insertUser("Ada@example.com"), "23514", "users_email_format")
	requirePgError(t, insertUser("ada@example.com"), "23505", "users_email_unique")

	// The unique index only covers live accounts, so soft-deleting frees the email.
	if _, err := pool.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE email = $1`, "ada@example.com"); err != nil {
		t.Fatalf("soft-delete user: %v", err)
	}
	if err := insertUser("ada@example.com"); err != nil {
		t.Fatalf("reuse email of soft-deleted user: %v", err)
	}
}

// newTestDatabase creates an empty, uniquely named database for the calling
// test and drops it when the test finishes. The returned config points at it.
func newTestDatabase(t *testing.T) *config.Config {
	t.Helper()

	if os.Getenv("TASKFLOW_INTEGRATION") != "1" {
		t.Skip("set TASKFLOW_INTEGRATION=1 to run PostgreSQL integration tests")
	}

	// Local convenience only: variables that are already set (as in CI) win.
	_ = godotenv.Load("../../.env")

	port, err := strconv.Atoi(requireEnv(t, "TASKFLOW_DATABASE_PORT"))
	if err != nil {
		t.Fatalf("TASKFLOW_DATABASE_PORT: %v", err)
	}

	cfg := &config.Config{}
	cfg.Database.Host = requireEnv(t, "TASKFLOW_DATABASE_HOST")
	cfg.Database.Port = port
	cfg.Database.User = requireEnv(t, "TASKFLOW_DATABASE_USER")
	cfg.Database.Password = requireEnv(t, "TASKFLOW_DATABASE_PASSWORD")
	cfg.Database.Name = requireEnv(t, "TASKFLOW_DATABASE_NAME")
	cfg.Database.SSLMode = cmp.Or(os.Getenv("TASKFLOW_DATABASE_SSLMODE"), "disable")
	cfg.Database.Migrations.Path = "migrations"

	admin, err := pgx.Connect(t.Context(), cfg.Database.URL())
	if err != nil {
		t.Fatalf("connect to %s: %v", cfg.Database.Name, err)
	}
	t.Cleanup(func() { admin.Close(context.Background()) })

	name := "taskflow_test_" + strings.ToLower(rand.Text())
	ident := pgx.Identifier{name}.Sanitize()
	if _, err := admin.Exec(t.Context(), "CREATE DATABASE "+ident); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		// t.Context() is already canceled when cleanups run.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP DATABASE "+ident+" WITH (FORCE)"); err != nil {
			t.Errorf("drop test database %s: %v", name, err)
		}
	})

	cfg.Database.Name = name
	return cfg
}

// newMigratedPool returns a pool connected to a fresh test database with all
// migrations applied.
func newMigratedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	cfg := newTestDatabase(t)
	if err := RunMigrations(cfg); err != nil {
		t.Fatalf("RunMigrations: %v", err)
	}

	pool, err := Conn(t.Context(), cfg)
	if err != nil {
		t.Fatalf("Conn: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func requireEnv(t *testing.T, key string) string {
	t.Helper()

	v := os.Getenv(key)
	if v == "" {
		t.Fatalf("%s must be set to run integration tests", key)
	}
	return v
}

func latestMigrationVersion(t *testing.T, dir string) int64 {
	t.Helper()

	migrations, err := goose.CollectMigrations(dir, 0, goose.MaxVersion)
	if err != nil {
		t.Fatalf("collect migrations: %v", err)
	}
	last, err := migrations.Last()
	if err != nil {
		t.Fatalf("find latest migration: %v", err)
	}
	return last.Version
}

func requireDBVersion(t *testing.T, db *sql.DB, want int64) {
	t.Helper()

	got, err := goose.GetDBVersion(db)
	if err != nil {
		t.Fatalf("read migration version: %v", err)
	}
	if got != want {
		t.Fatalf("database is at migration %d, want %d", got, want)
	}
}

func requirePgError(t *testing.T, err error, code, constraint string) {
	t.Helper()

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("got error %v, want PostgreSQL error %s on %s", err, code, constraint)
	}
	if pgErr.Code != code || pgErr.ConstraintName != constraint {
		t.Fatalf("got error %s on %q, want %s on %q", pgErr.Code, pgErr.ConstraintName, code, constraint)
	}
}
