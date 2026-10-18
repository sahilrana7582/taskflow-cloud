package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testYAML = `
app:
  name: taskflow-cloud
  port: 8080
database:
  host: localhost
  port: 54432
  name: taskflow_dev
  user: taskflow
  sslmode: disable
  migrations:
    path: internal/db/migrations
`

// inTempDir runs the test from an empty directory, so Load sees no .env and
// only the config/local.yaml the test writes.
func inTempDir(t *testing.T, yaml string) {
	t.Helper()

	dir := t.TempDir()
	if yaml != "" {
		if err := os.Mkdir(filepath.Join(dir, "config"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "config", "local.yaml"), []byte(yaml), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Chdir(dir)
}

func setSecrets(t *testing.T) {
	t.Setenv("TASKFLOW_DATABASE_PASSWORD", "secret")
	t.Setenv("JWT_SECRET_KEY", "jwt-secret")
}

func TestLoadLocalReadsYAMLAndIgnoresSettingEnvVars(t *testing.T) {
	inTempDir(t, testYAML)
	setSecrets(t)
	t.Setenv("TASKFLOW_APP_ENVIRONMENT", "local")
	t.Setenv("TASKFLOW_DATABASE_HOST", "taskflow-db")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Database.Host != "localhost" || cfg.Database.Port != 54432 {
		t.Errorf("database = %s:%d, want localhost:54432 from YAML", cfg.Database.Host, cfg.Database.Port)
	}
	if cfg.Database.Password != "secret" || cfg.JwtSecretKey != "jwt-secret" {
		t.Errorf("secrets were not read from the environment")
	}
	if cfg.App.Environment != EnvLocal {
		t.Errorf("environment = %q, want %q", cfg.App.Environment, EnvLocal)
	}
}

func TestLoadProductionReadsOnlyEnvVars(t *testing.T) {
	// The YAML is present but must not be used.
	inTempDir(t, testYAML)
	setSecrets(t)
	t.Setenv("TASKFLOW_APP_ENVIRONMENT", "production")
	t.Setenv("TASKFLOW_APP_NAME", "taskflow-cloud")
	t.Setenv("TASKFLOW_APP_PORT", "9090")
	t.Setenv("TASKFLOW_DATABASE_HOST", "taskflow-db")
	t.Setenv("TASKFLOW_DATABASE_PORT", "5432")
	t.Setenv("TASKFLOW_DATABASE_NAME", "taskflow_prod")
	t.Setenv("TASKFLOW_DATABASE_USER", "taskflow")
	t.Setenv("TASKFLOW_DATABASE_SSLMODE", "require")
	t.Setenv("TASKFLOW_DATABASE_MIGRATIONS_PATH", "internal/db/migrations")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	want := "postgres://taskflow:secret@taskflow-db:5432/taskflow_prod?sslmode=require"
	if got := cfg.Database.URL(); got != want {
		t.Errorf("URL = %s, want %s", got, want)
	}
	if cfg.App.Port != 9090 {
		t.Errorf("port = %d, want 9090", cfg.App.Port)
	}
}

func TestLoadProductionNamesMissingEnvVars(t *testing.T) {
	inTempDir(t, "")
	setSecrets(t)
	t.Setenv("TASKFLOW_APP_ENVIRONMENT", "production")
	for _, key := range settingKeys {
		t.Setenv(settingEnvVar(key), "")
	}
	t.Setenv("TASKFLOW_DATABASE_HOST", "taskflow-db")

	_, err := Load()
	if err == nil {
		t.Fatal("Load succeeded with settings missing")
	}
	if !strings.Contains(err.Error(), "TASKFLOW_DATABASE_PORT") || strings.Contains(err.Error(), "TASKFLOW_DATABASE_HOST") {
		t.Errorf("error should name exactly the unset variables, got: %v", err)
	}
}

func TestLoadRejectsUnknownEnvironment(t *testing.T) {
	inTempDir(t, testYAML)
	setSecrets(t)
	t.Setenv("TASKFLOW_APP_ENVIRONMENT", "prod")

	if _, err := Load(); err == nil {
		t.Fatal("Load accepted an unknown environment")
	}
}
