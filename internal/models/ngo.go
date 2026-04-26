package models

import "time"

type NgoSession struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Mission    string    `json:"mission"`
	Region     string    `json:"region"`
	Categories []string  `json:"categories"`
	Budget     *int64    `json:"budget"`
	Embedding  []float32 `json:"embedding"`
	CreatedAt  time.Time `json:"created_at"`
}
