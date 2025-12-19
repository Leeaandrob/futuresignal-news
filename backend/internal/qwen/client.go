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

// GenerateNarrative generates a narrative for a market signal using Bloomberg-style journalism.
func (c *Client) GenerateNarrative(ctx context.Context, signal SignalData) (*Narrative, error) {
	// Bloomberg-style editorial guidelines
	systemPrompt := `You are a senior financial journalist at a major news wire service.

EDITORIAL STANDARDS (The Bloomberg Way):
1. ACCURACY FIRST: Every fact must be precise. Use exact numbers, not approximations.
2. INTEGRATE DATA: Weave statistics into prose naturally (e.g., "surged 15 points to 78%" not "increased significantly")
3. EXPLAIN THE STAKES: Always answer "so what?" - why should sophisticated readers care?
4. SHORT & DIRECT: Prefer short sentences. Cut unnecessary words. One idea per sentence.
5. SPECIFIC OVER VAGUE: Name names, cite figures, be concrete.
6. FORWARD-LOOKING: What happens next? What are the implications?

STRUCTURE (Four-Paragraph Lead):
- LEAD: Hook with the most newsworthy development
- DETAILS: Supporting facts with integrated data
- NUT GRAPH: What's at stake for markets, policy, or the broader economy
- OUTLOOK: Forward-looking analysis

VOICE:
- Authoritative but not arrogant
- Objective - never advocate positions
- Professional wire service tone
- NO financial advice or recommendations

Respond ONLY with valid JSON.`

	// Determine the movement narrative
	change := signal.CurrentProb - signal.PreviousProb
	moveVerb := "moved"
	if change > 0.10 {
		moveVerb = "surged"
	} else if change > 0.05 {
		moveVerb = "jumped"
	} else if change > 0.02 {
		moveVerb = "rose"
	} else if change < -0.10 {
		moveVerb = "plunged"
	} else if change < -0.05 {
		moveVerb = "tumbled"
	} else if change < -0.02 {
		moveVerb = "fell"
	} else if change > 0 {
		moveVerb = "edged higher"
	} else if change < 0 {
		moveVerb = "slipped"
	}

	// Build social signals section if available
	socialSignalsSection := ""
	if signal.SocialSignalsContext != "" {
		socialSignalsSection = fmt.Sprintf(`

Social Signals (Tracked Influencer Posts):
%s
`, signal.SocialSignalsContext)
	}

	userPrompt := fmt.Sprintf(`Generate a Bloomberg-style news article for this prediction market signal.

═══════════════════════════════════════════════════════════════
MARKET DATA
═══════════════════════════════════════════════════════════════
Question: %s
Event: %s
Category: %s

Price Movement:
• Previous: %.1f%% → Current: %.1f%% (%s %+.1f points)
• 24h Volume: $%s
• Total Volume: $%s
• Timeframe: %s

External Context:
%s%s

═══════════════════════════════════════════════════════════════
OUTPUT REQUIREMENTS
═══════════════════════════════════════════════════════════════

Generate JSON with this structure:

{
  "headline": "Sharp, active-voice headline. Lead with action verb when possible. Max 90 chars. Example: 'Trump Election Odds Surge Past 70%% as Polling Gap Widens'",

  "subheadline": "One sentence capturing the key takeaway with specific data. Example: 'Prediction markets price in 15-point swing after debate, marking largest single-day move since June'",

  "what_changed": "THE LEAD + DETAILS (2-3 punchy sentences). Start with the news hook. Integrate exact figures. What specifically happened and when? Include the probability change, volume, and any catalysts. If social signals are present, mention the influencer commentary as supporting context.",

  "why_it_matters": "THE NUT GRAPH (2-3 sentences). Answer 'so what?' for sophisticated readers. What are the stakes? Economic implications? Policy consequences? How does this fit the bigger picture? Connect to broader market/political themes.",

  "market_context": "BROADER CONTEXT (2 sentences). Five Easy Pieces approach - connect to markets, economy, policy, or industry. What else is happening that relates to this? Historical context if relevant. Reference any relevant social signals as primary sources.",

  "what_to_watch": "FORWARD OUTLOOK (2 sentences). What catalysts could move this next? Key dates, events, or data releases to monitor. Be specific about triggers.",

  "tags": ["3-5 relevant SEO tags"],
  "sentiment": "bullish|bearish|neutral",
  "significance": "low|medium|high|breaking"
}

QUALITY CHECKLIST:
✓ Headline uses active voice and specific numbers
✓ Every sentence contains concrete information
✓ Data is woven into narrative, not listed separately
✓ "So what?" is clearly answered
✓ Forward-looking element included
✓ No hedge words (might, could, possibly) without substance
✓ If social signals are available, cite influencers as sources (e.g., "according to @handle")`,
		signal.MarketTitle,
		signal.EventTitle,
		signal.Category,
		signal.PreviousProb*100,
		signal.CurrentProb*100,
		moveVerb,
		change*100,
		formatVolume(signal.Volume24h),
		formatVolume(signal.TotalVolume),
		signal.TimeFrame,
		getContextOrDefault(signal.ExternalContext),
		socialSignalsSection,
	)

	var narrative Narrative
	err := c.ChatJSON(ctx, ChatRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Temperature:  0.4, // Slightly higher for more natural writing
		MaxTokens:    1200,
	}, &narrative)

	if err != nil {
		return nil, err
	}

	return &narrative, nil
}

func getContextOrDefault(ctx string) string {
	if ctx == "" {
		return "No additional context available. Focus on the market data and its implications."
	}
	return ctx
}

// SignalData represents market signal data for narrative generation.
type SignalData struct {
	MarketTitle          string
	EventTitle           string
	Category             string
	PreviousProb         float64
	CurrentProb          float64
	TimeFrame            string
	Volume24h            float64
	TotalVolume          float64
	ExternalContext      string
	SocialSignalsContext string // Context from XTracker influencer posts
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
