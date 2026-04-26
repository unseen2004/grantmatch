package main

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/unseen2004/grantmatch/internal/ai"
	"github.com/unseen2004/grantmatch/internal/config"
	"github.com/unseen2004/grantmatch/internal/db"
	"github.com/unseen2004/grantmatch/internal/handlers"
	"github.com/unseen2004/grantmatch/internal/rag"
	"github.com/unseen2004/grantmatch/internal/repository"
	"github.com/unseen2004/grantmatch/internal/scraper"
	"github.com/unseen2004/grantmatch/internal/scraper/sources"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()

	// Run DB migrations
	if err := db.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	// Connect DB pool
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool failed: %v", err)
	}
	defer pool.Close()

	// Init AI clients
	embedder, err := ai.NewEmbedder(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("embedder init failed: %v", err)
	}
	generator, err := ai.NewGenerator(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatalf("generator init failed: %v", err)
	}

	// Init repos and matcher
	grantRepo := repository.NewGrantRepository(pool)
	sessionRepo := repository.NewSessionRepository(pool)
	appRepo := repository.NewApplicationRepository(pool)
	matcher := rag.NewMatcher(pool)

	// Seed grants on startup in background
	go func() {
		scraper.Run(ctx,
			[]scraper.Source{&sources.GrantsGov{}},
			embedder,
			grantRepo,
		)
	}()

	// Parse templates with custom functions
	tmpl, err := template.New("").Funcs(template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
	}).ParseGlob(filepath.Join("internal", "templates", "*.html"))
	if err != nil {
		log.Fatalf("template parse failed: %v", err)
	}

	// Init handlers
	h := handlers.NewHandlers(sessionRepo, grantRepo, appRepo, embedder, generator, matcher, tmpl)

	// Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	r.Get("/", h.Home)
	r.Get("/grants", h.ListGrants)
	r.Post("/match", h.HandleMatch)
	r.Post("/draft", h.HandleDraft)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
