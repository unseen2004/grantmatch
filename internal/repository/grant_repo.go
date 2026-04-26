package repository

import (
	"context"
	"github.com/hackathon/grantmatch/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type GrantRepository struct {
	pool *pgxpool.Pool
}

func NewGrantRepository(pool *pgxpool.Pool) *GrantRepository {
	return &GrantRepository{pool: pool}
}

func (r *GrantRepository) Upsert(ctx context.Context, grant models.Grant) error {
	// Simple check if URL exists
	if grant.URL != "" {
		var exists bool
		err := r.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM grants WHERE url=$1)", grant.URL).Scan(&exists)
		if err == nil && exists {
			return nil // Skip
		}
	}

	query := `
		INSERT INTO grants (title, description, funder, amount_min, amount_max, currency, deadline, url, region, categories, eligibility, embedding, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	vec := pgvector.NewVector(grant.Embedding)
	_, err := r.pool.Exec(ctx, query,
		grant.Title, grant.Description, grant.Funder, grant.AmountMin, grant.AmountMax, grant.Currency,
		grant.Deadline, grant.URL, grant.Region, grant.Categories, grant.Eligibility, vec, grant.Source,
	)
	return err
}

func (r *GrantRepository) GetByID(ctx context.Context, id string) (models.Grant, error) {
	query := `SELECT id::text, title, description, funder, amount_min, amount_max, currency, deadline, url, region, categories, eligibility, source FROM grants WHERE id = $1`
	var g models.Grant
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&g.ID, &g.Title, &g.Description, &g.Funder, &g.AmountMin, &g.AmountMax, &g.Currency,
		&g.Deadline, &g.URL, &g.Region, &g.Categories, &g.Eligibility, &g.Source,
	)
	return g, err
}

func (r *GrantRepository) GetAll(ctx context.Context) ([]models.Grant, error) {
	query := `SELECT id::text, title, description, funder, deadline, amount_max, url FROM grants ORDER BY created_at DESC LIMIT 50`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []models.Grant
	for rows.Next() {
		var g models.Grant
		if err := rows.Scan(&g.ID, &g.Title, &g.Description, &g.Funder, &g.Deadline, &g.AmountMax, &g.URL); err == nil {
			grants = append(grants, g)
		}
	}
	return grants, nil
}
