package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// OpenRouterClient handles communication with OpenRouter API.
type OpenRouterClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// NewOpenRouterClient creates a new OpenRouter client.
func NewOpenRouterClient(apiKey string, logger *slog.Logger) *OpenRouterClient {
	return &OpenRouterClient{
		apiKey:  apiKey,
		baseURL: "https://openrouter.ai/api/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}
}

// ChatMessage represents a chat message.
type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// ChatCompletionRequest represents a chat completion request.
type ChatCompletionRequest struct {
	Model       string                 `json:"model"`
	Messages    []ChatMessage          `json:"messages"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	ImageURL    string                 `json:"image_url,omitempty"` // For vision models
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// ChatCompletionResponse represents the response from OpenRouter.
type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Model string `json:"model"`
}

// ChatCompletion performs a chat completion request.
func (c *OpenRouterClient) ChatCompletion(ctx context.Context, modelID string, messages []ChatMessage, imageURL string, maxTokens int, temperature float64) (*ChatCompletionResponse, error) {
	reqBody := ChatCompletionRequest{
		Model:       modelID,
		Messages:    messages,
		MaxTokens:   maxTokens,
		Temperature: temperature,
	}

	// Add image URL to last user message if provided
	if imageURL != "" && len(messages) > 0 {
		lastMsg := &messages[len(messages)-1]
		if lastMsg.Role == "user" {
			// For vision models, we need to format the content with image
			// OpenRouter supports image_url in content array
			reqBody.Extra = map[string]interface{}{
				"image_url": imageURL,
			}
		}
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/chat/completions", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("HTTP-Referer", "https://detectpricebyphoto.com") // Optional: for OpenRouter analytics
	req.Header.Set("X-Title", "Detect Price by Photo")

	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = c.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		if err != nil {
			c.logger.Warn("OpenRouter request failed, retrying", "attempt", i+1, "error", err)
		} else if resp.StatusCode == http.StatusTooManyRequests {
			c.logger.Warn("OpenRouter rate limited, retrying", "attempt", i+1)
			time.Sleep(time.Duration(i+1) * time.Second)
		} else {
			resp.Body.Close()
			return nil, fmt.Errorf("openrouter API error: status %d", resp.StatusCode)
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to execute request after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &response, nil
}

// EmbeddingRequest represents an embedding request.
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingResponse represents the embedding response.
type EmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// GenerateEmbedding generates embeddings for text using OpenAI ada-002 via OpenRouter.
func (c *OpenRouterClient) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	reqBody := EmbeddingRequest{
		Model: "openai/text-embedding-ada-002",
		Input: text,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/embeddings", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("HTTP-Referer", "https://detectpricebyphoto.com")
	req.Header.Set("X-Title", "Detect Price by Photo")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openrouter API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 || len(response.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("no embedding in response")
	}

	return response.Data[0].Embedding, nil
}

