package ai

import (
	"context"
	"fmt"
	"google.golang.org/genai"
)

type Generator struct {
	client *genai.Client
}

func NewGenerator(ctx context.Context, apiKey string) (*Generator, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, err
	}
	return &Generator{client: client}, nil
}

func (g *Generator) GenerateDraft(ctx context.Context, ngoMission, grantTitle, grantDesc, eligibility string) (string, error) {
	prompt := fmt.Sprintf(`You are an expert grant writer. Write a compelling grant application letter for the following:

NGO MISSION: %s

GRANT: %s
GRANT DESCRIPTION: %s
ELIGIBILITY CRITERIA: %s

Write a professional 400-600 word application letter. Structure:
1. Opening paragraph: Who we are and why we are applying
2. Problem statement: The social problem we address
3. Proposed project: How we will use the grant
4. Impact metrics: Measurable outcomes we commit to
5. Closing: Why we are the ideal recipient

Be specific, persuasive, and align the NGO mission tightly with the grant criteria.`, ngoMission, grantTitle, grantDesc, eligibility)

	resp, err := g.client.Models.GenerateContent(ctx, "gemini-1.5-flash", genai.Text(prompt), nil)
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		part := resp.Candidates[0].Content.Parts[0]
		if part.Text != "" {
			return part.Text, nil
		}
	}
	return "", fmt.Errorf("failed to generate draft")
}
