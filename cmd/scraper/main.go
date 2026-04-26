package main

import (
	"context"
	"os"

	"github.com/unseen2004/grantmatch/internal/ai"
	"github.com/unseen2004/grantmatch/internal/config"
	"github.com/unseen2004/grantmatch/internal/db"
	"github.com/unseen2004/grantmatch/internal/repository"
	"github.com/unseen2004/grantmatch/internal/scraper"
	"github.com/unseen2004/grantmatch/internal/scraper/sources"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	_ = godotenv.Load()
	cfg := config.Load()

	ctx := context.Background()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer pool.Close()

	grantRepo := repository.NewGrantRepository(pool)
	embedder, err := ai.NewEmbedder(ctx, cfg.GeminiAPIKey)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init embedder")
	}

	log.Info().Msg("Starting one-off scrape...")
	scraper.Run(ctx, []scraper.Source{&sources.GrantsGov{}}, embedder, grantRepo)
	log.Info().Msg("Scrape process completed")
}
