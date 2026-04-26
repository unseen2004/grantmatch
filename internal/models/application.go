package models

import "time"

type Application struct {
	ID        string    `json:"id"`
	SessionID string    `json:"session_id"`
	GrantID   string    `json:"grant_id"`
	Score     float64   `json:"score"`
	DraftText string    `json:"draft_text"`
	CreatedAt time.Time `json:"created_at"`
}
