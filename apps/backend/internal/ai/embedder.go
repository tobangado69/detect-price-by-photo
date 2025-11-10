package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// EmbedderService generates embeddings for product queries.
type EmbedderService struct {
	openRouterClient *OpenRouterClient
	logger           *slog.Logger
}

// NewEmbedderService creates a new embedder service.
func NewEmbedderService(openRouterClient *OpenRouterClient, logger *slog.Logger) *EmbedderService {
	return &EmbedderService{
		openRouterClient: openRouterClient,
		logger:           logger,
	}
}

// GenerateEmbedding generates an embedding for a product query.
// Combines product name and condition into a searchable text.
func (e *EmbedderService) GenerateEmbedding(ctx context.Context, productName string, condition string) ([]float64, error) {
	// Create searchable text: "product_name in condition condition"
	searchText := fmt.Sprintf("%s in %s condition", productName, condition)
	
	embedding, err := e.openRouterClient.GenerateEmbedding(ctx, searchText)
	if err != nil {
		e.logger.Error("failed to generate embedding", "product", productName, "condition", condition, "error", err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Validate embedding dimension (should be 1536 for ada-002)
	if len(embedding) != 1536 {
		return nil, fmt.Errorf("unexpected embedding dimension: got %d, expected 1536", len(embedding))
	}

	return embedding, nil
}

// GenerateEmbeddingFromText generates an embedding from arbitrary text.
func (e *EmbedderService) GenerateEmbeddingFromText(ctx context.Context, text string) ([]float64, error) {
	embedding, err := e.openRouterClient.GenerateEmbedding(ctx, text)
	if err != nil {
		e.logger.Error("failed to generate embedding from text", "text", text, "error", err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embedding) != 1536 {
		return nil, fmt.Errorf("unexpected embedding dimension: got %d, expected 1536", len(embedding))
	}

	return embedding, nil
}

// NormalizeProductName normalizes product name for better search results.
func (e *EmbedderService) NormalizeProductName(productName string) string {
	// Convert to lowercase, trim spaces
	normalized := strings.ToLower(strings.TrimSpace(productName))
	// Remove extra spaces
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

