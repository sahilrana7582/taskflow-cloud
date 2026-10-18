package config

import (
	"cmp"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Environments accepted in TASKFLOW_APP_ENVIRONMENT.
const (
	EnvLocal      = "local"
	EnvProduction = "production"
)

// settingKeys are the non-secret settings every environment must provide.
// Locally they come from config/local.yaml; in production each one is read
// from its environment variable (see settingEnvVar).
var settingKeys = []string{
	"app.name",
	"app.port",
	"database.host",
	"database.port",
	"database.name",
	"database.user",
	"database.sslmode",
	"database.migrations.path",
}

// secrets never live in YAML, so every environment reads them from the
// environment. Locally .env fills them in.
var secrets = []struct{ key, envVar string }{
	{"database.password", "TASKFLOW_DATABASE_PASSWORD"},
	{"jwt_secret_key", "JWT_SECRET_KEY"},
}

// Load builds the Config for the environment named by TASKFLOW_APP_ENVIRONMENT
// ("local" when unset):
//
//   - local: settings come from config/local.yaml.
//   - production: settings come only from environment variables, as set by the
//     container platform (docker compose here, the ECS task definition on AWS).
//     The YAML file is never read, so a local value can't leak into production.
func Load() (*Config, error) {
	// Local development convenience. .env is not copied into the image, and
	// variables that are already set are never overwritten.
	_ = godotenv.Load()

	environment := cmp.Or(os.Getenv("TASKFLOW_APP_ENVIRONMENT"), EnvLocal)

	v := viper.New()

	switch environment {
	case EnvLocal:
		v.SetConfigName("local")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	case EnvProduction:
		for _, key := range settingKeys {
			_ = v.BindEnv(key, settingEnvVar(key))
		}
	default:
		return nil, fmt.Errorf("TASKFLOW_APP_ENVIRONMENT must be %q or %q, got %q", EnvLocal, EnvProduction, environment)
	}

	for _, s := range secrets {
		_ = v.BindEnv(s.key, s.envVar)
	}

	// Report everything that's missing at once, named the way it has to be set,
	// so a misconfigured deploy fails on startup with a fixable message.
	var missing []string
	for _, key := range settingKeys {
		if v.GetString(key) == "" {
			if environment == EnvProduction {
				key = settingEnvVar(key)
			}
			missing = append(missing, key)
		}
	}
	for _, s := range secrets {
		if v.GetString(s.key) == "" {
			missing = append(missing, s.envVar)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%s config is missing: %s", environment, strings.Join(missing, ", "))
	}

	// Unmarshal only after every source is registered: it copies the values
	// viper knows at this moment, and later bindings would not reach cfg.
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.App.Environment = environment

	return &cfg, nil
}

// settingEnvVar is the environment variable a setting is read from in
// production, e.g. database.host -> TASKFLOW_DATABASE_HOST.
func settingEnvVar(key string) string {
	return "TASKFLOW_" + strings.ToUpper(strings.ReplaceAll(key, ".", "_"))
}
