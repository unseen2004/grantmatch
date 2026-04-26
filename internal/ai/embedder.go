package ai

import (
	"context"
	"google.golang.org/genai"
)

type Embedder struct {
	client *genai.Client
	model  string
}

func NewEmbedder(ctx context.Context, apiKey string) (*Embedder, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, err
	}
	return &Embedder{client: client, model: "text-embedding-004"}, nil
}

// Embed generates a 768-dim vector for the given text.
func (e *Embedder) Embed(ctx context.Context, text string) ([]float32, error) {
	result, err := e.client.Models.EmbedContent(ctx, e.model,
		genai.Text(text), nil)
	if err != nil {
		return nil, err
	}
	return result.Embeddings[0].Values, nil
}
