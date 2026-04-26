package scraper

import (
	"context"
	"log"
	"strings"

	"github.com/unseen2004/grantmatch/internal/ai"
	"github.com/unseen2004/grantmatch/internal/models"
	"github.com/unseen2004/grantmatch/internal/repository"
)

type Source interface {
	Name() string
	Scrape(ctx context.Context) ([]models.Grant, error)
}

func Run(ctx context.Context, sources []Source, embedder *ai.Embedder, repo *repository.GrantRepository) {
	for _, source := range sources {
		grants, err := source.Scrape(ctx)
		if err != nil {
			log.Printf("[scraper] %s failed: %v", source.Name(), err)
			continue
		}
		log.Printf("[scraper] %s: scraped %d grants", source.Name(), len(grants))
		for _, g := range grants {
			embText := strings.TrimSpace(g.Title + " " + g.Description + " " + g.Eligibility)
			if embText == "" {
				continue
			}
			vec, err := embedder.Embed(ctx, embText)
			if err != nil {
				log.Printf("[scraper] embed failed for %q: %v", g.Title, err)
				continue
			}
			g.Embedding = vec
			if err := repo.Upsert(ctx, g); err != nil {
				log.Printf("[scraper] upsert failed for %q: %v", g.Title, err)
			}
		}
	}
}
