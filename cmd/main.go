package main

import (
	"context"
	"log"
	"time"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
	"github.com/sahilrana7582/taskflow-cloud/internal/db"
	"github.com/sahilrana7582/taskflow-cloud/internal/routes"
	"github.com/sahilrana7582/taskflow-cloud/internal/server"
)

func main() {
	// Load Config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Validate the JWT settings before touching the database, so a missing or
	// weak secret stops startup immediately.
	tokens, err := auth.NewTokenManager(cfg.JwtSecretKey, cfg.App.Name, 4*time.Hour)
	if err != nil {
		log.Fatal("Failed to configure JWT:", err)
	}

	// 1. Connect to PostgreSQL.
	pool, err := db.Conn(context.Background(), cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// 2. Run database migrations.
	if err := db.RunMigrations(cfg); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 3. Wire dependencies and register routes.
	userRepo := auth.NewUserRepository(pool)
	authService := auth.NewAuthService(userRepo, tokens)
	authHandler := auth.NewAuthHandler(authService)

	mux := routes.NewServeMux(authHandler, tokens)

	// 4. Start the HTTP server.
	srv := server.NewServer("8080", *mux)

	log.Println("Starting server on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
