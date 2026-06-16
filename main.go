package main

import (
	"log"
	"net/http"
	"time"

	"who-owes-me/actual"
	"who-owes-me/auth"
	"who-owes-me/db"
	"who-owes-me/handlers"
	"who-owes-me/internal/envutil"

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

	actual.InitCache(5 * time.Minute)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	handlers.RegisterRoutes(r)

	port := envutil.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s\n", port)
	if err := http.ListenAndServe("0.0.0.0:"+port, r); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
