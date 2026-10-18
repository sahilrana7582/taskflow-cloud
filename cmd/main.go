package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sahilrana7582/taskflow-cloud/internal/auth"
	"github.com/sahilrana7582/taskflow-cloud/internal/db"
	"github.com/sahilrana7582/taskflow-cloud/internal/routes"
	"github.com/sahilrana7582/taskflow-cloud/internal/server"
)

func init() {
	fmt.Println("Running INIT")
	godotenv.Load()
}

func main() {
	// 1. Connect to PostgreSQL.
	pool, err := db.Conn(context.Background())
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	// 2. Run database migrations.
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 3. Wire dependencies and register routes.
	userRepo := auth.NewUserRepository(pool)
	authService := auth.NewAuthService(userRepo)
	authHandler := auth.NewAuthHandler(authService)

	mux := routes.NewServeMux(authHandler)

	// 4. Start the HTTP server.
	srv := server.NewServer("8080", mux)

	log.Println("Starting server on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
