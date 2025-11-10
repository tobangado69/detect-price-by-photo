package ai

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PriceEstimateResult represents the result of price estimation.
type PriceEstimateResult struct {
	EstimatedPriceMin   float64
	EstimatedPriceMax   float64
	EstimatedPriceMedian float64
	ConfidenceScore     float64
	Reasoning           string
	ModelUsed           string
	ProcessingTimeMs    int
	CostInUSD           float64
	MarketData          map[string]interface{}
}

// EstimatorService implements the decision tree logic for price estimation.
type EstimatorService struct {
	openRouterClient *OpenRouterClient
	embedder         *EmbedderService
	retriever        *RetrieverService
	modelRegistry    *ModelRegistry
	cacheClient      CacheClientInterface
	logger           *slog.Logger
}

// CacheClientInterface defines cache operations (to avoid circular dependency).
type CacheClientInterface interface {
	SetEstimation(ctx context.Context, key string, result interface{}, ttl time.Duration) error
	GetEstimation(ctx context.Context, key string, result interface{}) error
}

// NewEstimatorService creates a new estimator service.
func NewEstimatorService(
	openRouterClient *OpenRouterClient,
	embedder *EmbedderService,
	retriever *RetrieverService,
	modelRegistry *ModelRegistry,
	cacheClient CacheClientInterface,
	logger *slog.Logger,
) *EstimatorService {
	return &EstimatorService{
		openRouterClient: openRouterClient,
		embedder:         embedder,
		retriever:        retriever,
		modelRegistry:     modelRegistry,
		cacheClient:       cacheClient,
		logger:            logger,
	}
}

// EstimatePrice performs price estimation based on mode (fast/accurate/knowledge_based).
func (e *EstimatorService) EstimatePrice(ctx context.Context, imageURL string, productName string, condition string, mode string) (*PriceEstimateResult, error) {
	startTime := time.Now()

	// Generate cache key
	cacheKey := e.generateCacheKey(productName, condition, imageURL)

	// Check cache first
	var cachedResult PriceEstimateResult
	if err := e.cacheClient.GetEstimation(ctx, cacheKey, &cachedResult); err == nil {
		e.logger.Info("returning cached estimation", "product", productName, "condition", condition)
		return &cachedResult, nil
	}

	var result *PriceEstimateResult
	var err error

	// Decision tree based on mode
	switch mode {
	case "fast":
		result, err = e.estimateFast(ctx, imageURL, productName, condition)
	case "accurate":
		result, err = e.estimateAccurate(ctx, imageURL, productName, condition)
	case "knowledge_based":
		result, err = e.estimateKnowledgeBased(ctx, imageURL, productName, condition)
	default:
		// Fallback to fast mode
		e.logger.Warn("unknown mode, falling back to fast", "mode", mode)
		result, err = e.estimateFast(ctx, imageURL, productName, condition)
	}

	if err != nil {
		return nil, err
	}

	// Calculate processing time
	result.ProcessingTimeMs = int(time.Since(startTime).Milliseconds())

	// Cache result for 24 hours
	if err := e.cacheClient.SetEstimation(ctx, cacheKey, result, 24*time.Hour); err != nil {
		e.logger.Warn("failed to cache estimation result", "error", err)
	}

	return result, nil
}

// estimateFast uses quick LLM call with minimal context.
func (e *EstimatorService) estimateFast(ctx context.Context, imageURL string, productName string, condition string) (*PriceEstimateResult, error) {
	modelCfg := e.modelRegistry.GetModelForMode("fast")
	if modelCfg == nil {
		modelCfg = e.modelRegistry.GetDefaultModel()
	}

	prompt := e.buildFastPrompt(productName, condition)
	messages := []ChatMessage{
		{Role: "system", Content: "You are a price estimation expert for second-hand products in Indonesia. Provide price estimates in IDR."},
		{Role: "user", Content: prompt},
	}

	resp, err := e.openRouterClient.ChatCompletion(ctx, modelCfg.ID, messages, "", modelCfg.MaxTokens, modelCfg.Temperature)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	result := e.parseLLMResponse(resp.Choices[0].Message.Content, modelCfg.ID, resp.Usage.TotalTokens)
	return result, nil
}

// estimateAccurate uses full LLM call with image analysis.
func (e *EstimatorService) estimateAccurate(ctx context.Context, imageURL string, productName string, condition string) (*PriceEstimateResult, error) {
	modelCfg := e.modelRegistry.GetModelForMode("accurate")
	if modelCfg == nil {
		modelCfg = e.modelRegistry.GetDefaultModel()
	}

	prompt := e.buildAccuratePrompt(productName, condition)
	messages := []ChatMessage{
		{Role: "system", Content: "You are a price estimation expert for second-hand products in Indonesia. Analyze the product image and provide detailed price estimates in IDR."},
		{Role: "user", Content: prompt},
	}

	resp, err := e.openRouterClient.ChatCompletion(ctx, modelCfg.ID, messages, imageURL, modelCfg.MaxTokens, modelCfg.Temperature)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	result := e.parseLLMResponse(resp.Choices[0].Message.Content, modelCfg.ID, resp.Usage.TotalTokens)
	return result, nil
}

// estimateKnowledgeBased uses RAG + LLM synthesis.
func (e *EstimatorService) estimateKnowledgeBased(ctx context.Context, imageURL string, productName string, condition string) (*PriceEstimateResult, error) {
	// 1. Generate embedding for product query
	embedding, err := e.embedder.GenerateEmbedding(ctx, productName, condition)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// 2. Retrieve similar market records
	marketRecords, err := e.retriever.RetrieveSimilarProducts(ctx, embedding, condition, 5)
	if err != nil {
		e.logger.Warn("failed to retrieve market records, proceeding without RAG", "error", err)
		marketRecords = []*MarketRecord{}
	}

	// 3. Build prompt with market data
	prompt := e.buildKnowledgePrompt(productName, condition, marketRecords)
	
	modelCfg := e.modelRegistry.GetModelForMode("knowledge_based")
	if modelCfg == nil {
		modelCfg = e.modelRegistry.GetDefaultModel()
	}

	messages := []ChatMessage{
		{Role: "system", Content: "You are a price estimation expert for second-hand products in Indonesia. Use the provided market data to synthesize accurate price estimates in IDR."},
		{Role: "user", Content: prompt},
	}

	resp, err := e.openRouterClient.ChatCompletion(ctx, modelCfg.ID, messages, imageURL, modelCfg.MaxTokens, modelCfg.Temperature)
	if err != nil {
		return nil, fmt.Errorf("failed to get LLM response: %w", err)
	}

	result := e.parseLLMResponse(resp.Choices[0].Message.Content, modelCfg.ID, resp.Usage.TotalTokens)
	
	// Add market data to result
	if len(marketRecords) > 0 {
		result.MarketData = map[string]interface{}{
			"total_listings_found": len(marketRecords),
			"similar_products":     marketRecords,
		}
	}

	return result, nil
}

// buildFastPrompt creates a prompt for fast mode.
func (e *EstimatorService) buildFastPrompt(productName string, condition string) string {
	return fmt.Sprintf(`Estimate the market price for a %s in %s condition in Indonesia (IDR).

Provide your response in JSON format:
{
  "min_price": <number>,
  "max_price": <number>,
  "median_price": <number>,
  "confidence": <0.0-1.0>,
  "reasoning": "<brief explanation>"
}`, productName, condition)
}

// buildAccuratePrompt creates a prompt for accurate mode.
func (e *EstimatorService) buildAccuratePrompt(productName string, condition string) string {
	return fmt.Sprintf(`Analyze the product image and estimate the market price for a %s in %s condition in Indonesia (IDR).

Consider:
- Physical condition visible in the image
- Brand and model identification
- Market trends for this product category
- Condition-specific pricing adjustments

Provide your response in JSON format:
{
  "min_price": <number>,
  "max_price": <number>,
  "median_price": <number>,
  "confidence": <0.0-1.0>,
  "reasoning": "<detailed explanation>"
}`, productName, condition)
}

// buildKnowledgePrompt creates a prompt for knowledge-based mode with market data.
func (e *EstimatorService) buildKnowledgePrompt(productName string, condition string, marketRecords []*MarketRecord) string {
	var marketDataStr strings.Builder
	if len(marketRecords) > 0 {
		marketDataStr.WriteString("\n\nMarket Data from Similar Products:\n")
		for i, record := range marketRecords {
			marketDataStr.WriteString(fmt.Sprintf("%d. %s (%s) - Similarity: %.2f\n", i+1, record.ProductName, record.Condition, record.Similarity))
			if price, ok := record.MarketData["average_price"].(float64); ok {
				marketDataStr.WriteString(fmt.Sprintf("   Average Price: IDR %.0f\n", price))
			}
		}
	}

	return fmt.Sprintf(`Estimate the market price for a %s in %s condition in Indonesia (IDR) using the provided market data.%s

Synthesize the market data with your knowledge to provide accurate estimates.

Provide your response in JSON format:
{
  "min_price": <number>,
  "max_price": <number>,
  "median_price": <number>,
  "confidence": <0.0-1.0>,
  "reasoning": "<explanation referencing market data>"
}`, productName, condition, marketDataStr.String())
}

// parseLLMResponse parses LLM JSON response into PriceEstimateResult.
func (e *EstimatorService) parseLLMResponse(content string, modelID string, tokensUsed int) *PriceEstimateResult {
	result := &PriceEstimateResult{
		ModelUsed: modelID,
		CostInUSD: e.calculateCost(modelID, tokensUsed),
	}

	// Try to extract JSON from response
	jsonMatch := regexp.MustCompile(`\{[^}]+\}`).FindString(content)
	if jsonMatch == "" {
		// Fallback: try to extract numbers
		result = e.extractPricesFromText(content, modelID, tokensUsed)
		result.Reasoning = content
		return result
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonMatch), &data); err != nil {
		e.logger.Warn("failed to parse LLM JSON response, using fallback", "error", err)
		result = e.extractPricesFromText(content, modelID, tokensUsed)
		result.Reasoning = content
		return result
	}

	// Extract values
	if min, ok := data["min_price"].(float64); ok {
		result.EstimatedPriceMin = min
	}
	if max, ok := data["max_price"].(float64); ok {
		result.EstimatedPriceMax = max
	}
	if median, ok := data["median_price"].(float64); ok {
		result.EstimatedPriceMedian = median
	} else if result.EstimatedPriceMin > 0 && result.EstimatedPriceMax > 0 {
		result.EstimatedPriceMedian = (result.EstimatedPriceMin + result.EstimatedPriceMax) / 2
	}
	if conf, ok := data["confidence"].(float64); ok {
		result.ConfidenceScore = conf
	} else {
		result.ConfidenceScore = 0.7 // Default confidence
	}
	if reasoning, ok := data["reasoning"].(string); ok {
		result.Reasoning = reasoning
	} else {
		result.Reasoning = content
	}

	return result
}

// extractPricesFromText extracts price estimates from unstructured text (fallback).
func (e *EstimatorService) extractPricesFromText(text string, modelID string, tokensUsed int) *PriceEstimateResult {
	result := &PriceEstimateResult{
		ModelUsed:        modelID,
		CostInUSD:        e.calculateCost(modelID, tokensUsed),
		ConfidenceScore:  0.5, // Lower confidence for unstructured parsing
		Reasoning:         text,
		EstimatedPriceMin: 0,
		EstimatedPriceMax: 0,
	}

	// Try to find price ranges in text (e.g., "5,000,000 - 7,000,000 IDR")
	pricePattern := regexp.MustCompile(`(\d{1,3}(?:[.,]\d{3})*(?:[.,]\d{3})*)`)
	matches := pricePattern.FindAllString(text, -1)
	if len(matches) >= 2 {
		if min, err := parsePrice(matches[0]); err == nil {
			result.EstimatedPriceMin = min
		}
		if max, err := parsePrice(matches[len(matches)-1]); err == nil {
			result.EstimatedPriceMax = max
		}
		if result.EstimatedPriceMin > 0 && result.EstimatedPriceMax > 0 {
			result.EstimatedPriceMedian = (result.EstimatedPriceMin + result.EstimatedPriceMax) / 2
		}
	}

	return result
}

// parsePrice parses price string to float64.
func parsePrice(s string) (float64, error) {
	// Remove commas and dots (thousand separators)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	return strconv.ParseFloat(s, 64)
}

// calculateCost estimates cost in USD based on model and tokens.
func (e *EstimatorService) calculateCost(modelID string, tokensUsed int) float64 {
	// Rough cost estimates per 1K tokens (input + output average)
	costPer1K := map[string]float64{
		"openai/gpt-4o-mini":           0.00015, // $0.15 per 1M tokens
		"anthropic/claude-3-5-haiku":    0.00025, // $0.25 per 1M tokens
		"openai/gpt-4o":                0.005,   // $5 per 1M tokens
		"anthropic/claude-3-5-sonnet":   0.003,   // $3 per 1M tokens
	}

	cost, ok := costPer1K[modelID]
	if !ok {
		cost = 0.001 // Default estimate
	}

	return float64(tokensUsed) / 1000.0 * cost
}

// generateCacheKey generates a cache key for estimation results.
func (e *EstimatorService) generateCacheKey(productName string, condition string, imageURL string) string {
	// Use hash of product name + condition + image URL hash
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s", productName, condition, imageURL)))
	return fmt.Sprintf("estimate:%x", hash[:8])
}

