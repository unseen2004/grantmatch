package rag

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"time"
)

type GrantMatch struct {
	GrantID     string
	Title       string
	Funder      string
	Description string
	Deadline    *time.Time
	AmountMax   *int64
	URL         string
	Score       float64
}

type Matcher struct {
	pool *pgxpool.Pool
}

func NewMatcher(pool *pgxpool.Pool) *Matcher {
	return &Matcher{pool: pool}
}

// FindMatches performs cosine similarity search against the grants table.
// topK controls how many results to return.
func (m *Matcher) FindMatches(ctx context.Context, ngoEmbedding []float32, topK int) ([]GrantMatch, error) {
	query := `
        SELECT id::text, title, funder, description, 
               deadline, amount_max, url,
               1 - (embedding <=> $1) AS score
        FROM grants
        WHERE deadline >= CURRENT_DATE OR deadline IS NULL
        ORDER BY embedding <=> $1
        LIMIT $2
    `
	vec := pgvector.NewVector(ngoEmbedding)
	rows, err := m.pool.Query(ctx, query, vec, topK)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []GrantMatch
	for rows.Next() {
		var m GrantMatch
		if err := rows.Scan(&m.GrantID, &m.Title, &m.Funder,
			&m.Description, &m.Deadline, &m.AmountMax, &m.URL, &m.Score); err != nil {
			continue
		}
		matches = append(matches, m)
	}
	return matches, nil
}
