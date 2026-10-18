package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/sahilrana27582/taskflow/internal/db"
	"github.com/sahilrana27582/taskflow/internal/routes"
	"github.com/sahilrana27582/taskflow/internal/server"
)

func init() {
	fmt.Println("Running INIT")
	godotenv.Load()
}

func main() {
	// 1. Connect to PostgreSQL.
	conn, err := db.Conn()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close(context.Background())

	// 2. Run database migrations.
	if err := db.RunMigrations(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 3. Register routes.
	mux := routes.NewServeMux()

	// 4. Start the HTTP server.
	srv := server.NewServer("8080", mux)

	log.Println("Starting server on port 8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal("Server failed:", err)
	}
}
