package config

import (
	"fmt"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func Load() (*Config, error) {
	// Load .env for local development.
	// Existing environment variables are not overwritten.
	_ = godotenv.Load()

	v := viper.New()
	v.SetConfigName("local")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// Environment variables take precedence over YAML.
	v.SetEnvPrefix("TASKFLOW")
	v.AutomaticEnv()

	// Bind environment variables explicitly for predictable mapping.
	_ = v.BindEnv("database.url", "DATABASE_URL")
	_ = v.BindEnv("database.password", "TASKFLOW_DATABASE_PASSWORD")

	cfg.Database.Password = v.GetString("database.password")

	cfg.DatabaseURL = v.GetString("database.url")
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	fmt.Println(cfg)

	return &cfg, nil
}
