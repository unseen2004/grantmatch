package repository

import (
	"context"
	"github.com/hackathon/grantmatch/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ApplicationRepository struct {
	pool *pgxpool.Pool
}

func NewApplicationRepository(pool *pgxpool.Pool) *ApplicationRepository {
	return &ApplicationRepository{pool: pool}
}

func (r *ApplicationRepository) Create(ctx context.Context, app models.Application) error {
	query := `
		INSERT INTO applications (session_id, grant_id, score, draft_text)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.pool.Exec(ctx, query, app.SessionID, app.GrantID, app.Score, app.DraftText)
	return err
}
