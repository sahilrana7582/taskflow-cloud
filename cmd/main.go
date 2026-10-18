package main

import (
	"context"
	"log"
	"time"

	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/config"
	"github.com/sahilrana7582/taskflow-cloud/internal/db"
	"github.com/sahilrana7582/taskflow-cloud/internal/job"
	"github.com/sahilrana7582/taskflow-cloud/internal/project"
	"github.com/sahilrana7582/taskflow-cloud/internal/routes"
	"github.com/sahilrana7582/taskflow-cloud/internal/server"
	"github.com/sahilrana7582/taskflow-cloud/internal/workspace"
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

	// Projects and jobs ask the workspace service whether the caller may act.
	workspaceService := workspace.NewService(workspace.NewRepository(pool))
	projectService := project.NewService(project.NewRepository(pool), workspaceService)
	jobService := job.NewService(job.NewRepository(pool), workspaceService)

	mux := routes.NewServeMux(routes.Handlers{
		Auth:      authHandler,
		Workspace: workspace.NewHandler(workspaceService),
		Project:   project.NewHandler(projectService),
		Job:       job.NewHandler(jobService),
	}, tokens)

	// 4. Start the HTTP server.
	srv := server.NewServer("8080", mux)

	log.Println("Starting server on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
