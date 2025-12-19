// Package qwen provides a client for Alibaba Qwen Cloud (DashScope) API.
// Uses OpenAI-compatible endpoint for chat completions.
package qwen

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
	openai "github.com/sashabaranov/go-openai"
)

const (
	// DashScope OpenAI-compatible endpoint
	DefaultEndpoint = "https://dashscope-intl.aliyuncs.com/compatible-mode/v1"

	// Available models
	ModelQwenPlus    = "qwen-plus"
	ModelQwenTurbo   = "qwen-turbo"
	ModelQwenMax     = "qwen-max"
	ModelQwenLong    = "qwen-long"
)

// Client wraps the OpenAI SDK configured for DashScope.
type Client struct {
	client *openai.Client
	model  string
}

// Config holds the configuration for the Qwen client.
type Config struct {
	APIKey   string
	Endpoint string
	Model    string
}

// NewClient creates a new Qwen client.
func NewClient(cfg Config) *Client {
	if cfg.Endpoint == "" {
		cfg.Endpoint = DefaultEndpoint
	}
	if cfg.Model == "" {
		cfg.Model = ModelQwenPlus
	}

	config := openai.DefaultConfig(cfg.APIKey)
	config.BaseURL = cfg.Endpoint

	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  cfg.Model,
	}
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	SystemPrompt string
	UserPrompt   string
	Temperature  float32
	MaxTokens    int
	JSONMode     bool
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	Content      string
	FinishReason string
	TokensUsed   TokenUsage
}

// TokenUsage represents token usage statistics.
type TokenUsage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Chat sends a chat completion request to Qwen.
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	messages := []openai.ChatCompletionMessage{}

	if req.SystemPrompt != "" {
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: req.SystemPrompt,
		})
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: req.UserPrompt,
	})

	chatReq := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: req.Temperature,
	}

	if req.MaxTokens > 0 {
		chatReq.MaxTokens = req.MaxTokens
	}

	if req.JSONMode {
		chatReq.ResponseFormat = &openai.ChatCompletionResponseFormat{
			Type: openai.ChatCompletionResponseFormatTypeJSONObject,
		}
	}

	log.Debug().
		Str("model", c.model).
		Int("messages", len(messages)).
		Bool("json_mode", req.JSONMode).
		Msg("Sending chat request to Qwen")

	resp, err := c.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("qwen chat completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	return &ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		FinishReason: string(resp.Choices[0].FinishReason),
		TokensUsed: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// ChatJSON sends a chat request and parses the response as JSON.
func (c *Client) ChatJSON(ctx context.Context, req ChatRequest, result interface{}) error {
	req.JSONMode = true

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(resp.Content), result); err != nil {
		return fmt.Errorf("failed to parse JSON response: %w", err)
	}

	return nil
}

// GenerateNarrative generates a narrative for a market signal.
func (c *Client) GenerateNarrative(ctx context.Context, signal SignalData) (*Narrative, error) {
	systemPrompt := `You are a financial markets analyst who explains prediction market movements in clear, editorial language.
Your tone is professional, objective, and informative - like Bloomberg or Reuters.
Never give financial advice or recommend actions.
Always explain the "why" behind market movements.
Respond ONLY with valid JSON.`

	userPrompt := fmt.Sprintf(`Analyze this prediction market signal and generate an editorial narrative.

MARKET: %s
EVENT: %s
CATEGORY: %s

SIGNAL DATA:
- Previous Probability: %.1f%%
- Current Probability: %.1f%%
- Change: %+.1f%% in %s
- Volume (24h): $%s
- Total Volume: $%s

CONTEXT (if available):
%s

Generate a JSON response:
{
  "headline": "Compelling headline (max 100 chars)",
  "subheadline": "One sentence explaining the key takeaway",
  "what_changed": "2-3 sentences explaining what happened",
  "why_it_matters": "2-3 sentences on significance",
  "market_context": "1-2 sentences on broader context",
  "what_to_watch": "1-2 sentences on what could move markets next",
  "tags": ["relevant", "tags", "for", "seo"],
  "sentiment": "bullish|bearish|neutral",
  "significance": "low|medium|high|breaking"
}`,
		signal.MarketTitle,
		signal.EventTitle,
		signal.Category,
		signal.PreviousProb*100,
		signal.CurrentProb*100,
		(signal.CurrentProb-signal.PreviousProb)*100,
		signal.TimeFrame,
		formatVolume(signal.Volume24h),
		formatVolume(signal.TotalVolume),
		signal.ExternalContext,
	)

	var narrative Narrative
	err := c.ChatJSON(ctx, ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.3,
		MaxTokens:    1000,
	}, &narrative)

	if err != nil {
		return nil, err
	}

	return &narrative, nil
}

// SignalData represents market signal data for narrative generation.
type SignalData struct {
	MarketTitle     string
	EventTitle      string
	Category        string
	PreviousProb    float64
	CurrentProb     float64
	TimeFrame       string
	Volume24h       float64
	TotalVolume     float64
	ExternalContext string
}

// Narrative represents a generated narrative.
type Narrative struct {
	Headline      string   `json:"headline"`
	Subheadline   string   `json:"subheadline"`
	WhatChanged   string   `json:"what_changed"`
	WhyItMatters  string   `json:"why_it_matters"`
	MarketContext string   `json:"market_context"`
	WhatToWatch   string   `json:"what_to_watch"`
	Tags          []string `json:"tags"`
	Sentiment     string   `json:"sentiment"`
	Significance  string   `json:"significance"`
}

func formatVolume(v float64) string {
	switch {
	case v >= 1_000_000:
		return fmt.Sprintf("%.1fM", v/1_000_000)
	case v >= 1_000:
		return fmt.Sprintf("%.1fK", v/1_000)
	default:
		return fmt.Sprintf("%.0f", v)
	}
}
