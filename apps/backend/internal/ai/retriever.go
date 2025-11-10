package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

// MarketRecord represents a market knowledge record.
type MarketRecord struct {
	ID          string
	ProductName string
	Condition   string
	MarketData  map[string]interface{}
	Source      string
	Similarity  float64 // Cosine similarity score
}

// RetrieverService performs vector similarity search on market knowledge base.
type RetrieverService struct {
	pgPool *pgxpool.Pool
	logger *slog.Logger
}

// NewRetrieverService creates a new retriever service.
func NewRetrieverService(pgPool *pgxpool.Pool, logger *slog.Logger) *RetrieverService {
	return &RetrieverService{
		pgPool: pgPool,
		logger: logger,
	}
}

// RetrieveSimilarProducts performs vector similarity search and returns top-N similar products.
func (r *RetrieverService) RetrieveSimilarProducts(ctx context.Context, embedding []float64, condition string, limit int) ([]*MarketRecord, error) {
	if limit <= 0 || limit > 20 {
		limit = 5 // Default to 5
	}

	// Convert embedding to PostgreSQL vector format
	embeddingStr := formatVector(embedding)

	// Query with cosine similarity (<=> operator)
	// Filter by condition if provided
	query := `
		SELECT 
			id, 
			product_name, 
			condition, 
			market_data, 
			source,
			1 - (embedding <=> $1::vector) as similarity
		FROM public.market_knowledge
		WHERE 1 - (embedding <=> $1::vector) > 0.3
	`
	args := []interface{}{embeddingStr}
	
	if condition != "" {
		query += " AND condition = $2"
		args = append(args, condition)
		query += fmt.Sprintf(" ORDER BY embedding <=> $1::vector LIMIT $%d", len(args)+1)
	} else {
		query += fmt.Sprintf(" ORDER BY embedding <=> $1::vector LIMIT $%d", len(args)+1)
	}
	
	args = append(args, limit)

	rows, err := r.pgPool.Query(ctx, query, args...)
	if err != nil {
		r.logger.Error("failed to query market knowledge", "error", err)
		return nil, fmt.Errorf("failed to query market knowledge: %w", err)
	}
	defer rows.Close()

	var records []*MarketRecord
	for rows.Next() {
		var record MarketRecord
		var marketDataJSON []byte
		
		err := rows.Scan(
			&record.ID,
			&record.ProductName,
			&record.Condition,
			&marketDataJSON,
			&record.Source,
			&record.Similarity,
		)
		if err != nil {
			r.logger.Error("failed to scan market record", "error", err)
			continue
		}

		// Parse JSONB market_data
		if err := json.Unmarshal(marketDataJSON, &record.MarketData); err != nil {
			r.logger.Warn("failed to unmarshal market_data", "error", err)
			record.MarketData = make(map[string]interface{})
		}

		records = append(records, &record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	r.logger.Info("retrieved similar products", "count", len(records), "condition", condition)
	return records, nil
}

// formatVector converts a float64 slice to PostgreSQL vector string format.
func formatVector(embedding []float64) string {
	if len(embedding) == 0 {
		return "[]"
	}
	
	str := "["
	for i, v := range embedding {
		if i > 0 {
			str += ","
		}
		str += fmt.Sprintf("%.6f", v)
	}
	str += "]"
	return str
}

// StoreMarketRecord stores a new market knowledge record with embedding.
func (r *RetrieverService) StoreMarketRecord(ctx context.Context, productName string, condition string, embedding []float64, marketData map[string]interface{}, source string) error {
	embeddingStr := formatVector(embedding)
	
	marketDataJSON, err := json.Marshal(marketData)
	if err != nil {
		return fmt.Errorf("failed to marshal market data: %w", err)
	}

	query := `
		INSERT INTO public.market_knowledge (product_name, condition, embedding, market_data, source)
		VALUES ($1, $2, $3::vector, $4::jsonb, $5)
		ON CONFLICT DO NOTHING
	`

	_, err = r.pgPool.Exec(ctx, query, productName, condition, embeddingStr, marketDataJSON, source)
	if err != nil {
		r.logger.Error("failed to store market record", "error", err)
		return fmt.Errorf("failed to store market record: %w", err)
	}

	return nil
}

