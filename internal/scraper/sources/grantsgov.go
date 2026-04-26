package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/unseen2004/grantmatch/internal/models"
)

type GrantsGov struct{}

func (g *GrantsGov) Name() string { return "grants.gov" }

type grantsGovResponse struct {
	OppHits []struct {
		ID          int    `json:"id"`
		Title       string `json:"title"`
		AgencyName  string `json:"agencyName"`
		Synopsis    string `json:"synopsis"`
		CloseDate   string `json:"closeDate"` // "MMDDYYYY"
		AwardCeiling int64  `json:"awardCeiling"`
		AwardFloor   int64  `json:"awardFloor"`
		OppNumber    string `json:"number"`
	} `json:"oppHits"`
}

func (g *GrantsGov) Scrape(ctx context.Context) ([]models.Grant, error) {
	url := "https://apply07.grants.gov/grantsws/rest/opportunities/search/"
	payload := `{"keyword":"nonprofit NGO social","oppStatuses":"forecasted|posted","rows":100,"sortBy":"openDate|desc"}`

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result grantsGovResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	var grants []models.Grant
	for _, hit := range result.OppHits {
		deadline := parseGrantsGovDate(hit.CloseDate)
		grants = append(grants, models.Grant{
			Title:       hit.Title,
			Description: hit.Synopsis,
			Funder:      hit.AgencyName,
			AmountMin:   &hit.AwardFloor,
			AmountMax:   &hit.AwardCeiling,
			Currency:    "USD",
			Deadline:    deadline,
			URL:         fmt.Sprintf("https://www.grants.gov/search-results-detail/%d", hit.ID),
			Region:      "usa",
			Categories:  []string{"government", "nonprofit"},
			Eligibility: "",
			Source:      "grants.gov",
		})
	}
	return grants, nil
}

// parseGrantsGovDate parses "MMDDYYYY" → *time.Time
func parseGrantsGovDate(s string) *time.Time {
	if len(s) != 8 {
		return nil
	}
	t, err := time.Parse("01022006", s)
	if err != nil {
		return nil
	}
	return &t
}
