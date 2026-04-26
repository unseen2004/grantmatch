package repository

import (
	"context"
	"github.com/unseen2004/grantmatch/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, session models.NgoSession) (models.NgoSession, error) {
	query := `
		INSERT INTO ngo_sessions (name, mission, region, categories, budget)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, created_at
	`
	err := r.pool.QueryRow(ctx, query, session.Name, session.Mission, session.Region, session.Categories, session.Budget).Scan(&session.ID, &session.CreatedAt)
	return session, err
}

func (r *SessionRepository) UpdateEmbedding(ctx context.Context, id string, embedding []float32) error {
	query := `UPDATE ngo_sessions SET embedding = $1 WHERE id = $2`
	vec := pgvector.NewVector(embedding)
	_, err := r.pool.Exec(ctx, query, vec, id)
	return err
}

func (r *SessionRepository) GetByID(ctx context.Context, id string) (models.NgoSession, error) {
	query := `SELECT id::text, name, mission, region, categories, budget FROM ngo_sessions WHERE id = $1`
	var s models.NgoSession
	err := r.pool.QueryRow(ctx, query, id).Scan(&s.ID, &s.Name, &s.Mission, &s.Region, &s.Categories, &s.Budget)
	return s, err
}
