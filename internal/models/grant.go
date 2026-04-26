package models

import "time"

type Grant struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Funder      string     `json:"funder"`
	AmountMin   *int64     `json:"amount_min"`
	AmountMax   *int64     `json:"amount_max"`
	Currency    string     `json:"currency"`
	Deadline    *time.Time `json:"deadline"`
	URL         string     `json:"url"`
	Region      string     `json:"region"`
	Categories  []string   `json:"categories"`
	Eligibility string     `json:"eligibility"`
	Embedding   []float32  `json:"embedding"`
	Source      string     `json:"source"`
	ScrapedAt   time.Time  `json:"scraped_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
