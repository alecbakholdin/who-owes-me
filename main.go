package main

import (
	"log"
	"net/http"
	"os"

	"who-owes-me/auth"
	"who-owes-me/db"
	"who-owes-me/handlers"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.dev")

	db.InitDB()

	if err := auth.InitOIDC(); err != nil {
		log.Printf("WARNING: OIDC not configured (%v) — running without authentication", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handlers.RegisterRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
