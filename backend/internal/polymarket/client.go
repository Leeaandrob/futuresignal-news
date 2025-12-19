// Package polymarket provides a client for Polymarket's public APIs.
// Implements Data API and Gamma API for market data retrieval.
package polymarket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
)

const (
	// API endpoints
	GammaAPIBase = "https://gamma-api.polymarket.com"
	DataAPIBase  = "https://data-api.polymarket.com"
	CLOBAPIBase  = "https://clob.polymarket.com"

	// Rate limits (requests per 10 seconds)
	GammaRateLimit  = 750
	DataRateLimit   = 200
	MarketsLimit    = 125
	EventsLimit     = 100
)

// Client provides access to Polymarket APIs.
type Client struct {
	gamma *resty.Client
	data  *resty.Client
	clob  *resty.Client
}

// NewClient creates a new Polymarket client.
func NewClient() *Client {
	return &Client{
		gamma: resty.New().
			SetBaseURL(GammaAPIBase).
			SetTimeout(30 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second),
		data: resty.New().
			SetBaseURL(DataAPIBase).
			SetTimeout(30 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second),
		clob: resty.New().
			SetBaseURL(CLOBAPIBase).
			SetTimeout(30 * time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1 * time.Second),
	}
}

// JSONStringArray handles fields that come as JSON-encoded strings.
type JSONStringArray []string

func (j *JSONStringArray) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a regular array first
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*j = arr
		return nil
	}

	// Try to unmarshal as a string containing JSON array
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	// Parse the inner JSON array
	if str == "" {
		*j = []string{}
		return nil
	}

	if err := json.Unmarshal([]byte(str), &arr); err != nil {
		return err
	}
	*j = arr
	return nil
}

// Market represents a prediction market.
type Market struct {
	ID                    string          `json:"id"`
	Question              string          `json:"question"`
	ConditionID           string          `json:"conditionId"`
	Slug                  string          `json:"slug"`
	EndDate               string          `json:"endDate"`
	Description           string          `json:"description"`
	Outcomes              JSONStringArray `json:"outcomes"`
	OutcomePrices         JSONStringArray `json:"outcomePrices"`
	Volume                string          `json:"volume"`
	Volume24hr            float64         `json:"volume24hr"`
	Liquidity             string          `json:"liquidity"`
	Active                bool            `json:"active"`
	Closed                bool            `json:"closed"`
	MarketType            string          `json:"marketType"`
	GroupItemTitle        string          `json:"groupItemTitle"`
	GroupItemThreshold    string          `json:"groupItemThreshold"`
	Winner                string          `json:"winner"`
	VolumeNum             float64         `json:"volumeNum"`
	LiquidityNum          float64         `json:"liquidityNum"`
	CompetitorCount       int             `json:"competitorCount"`
	EnableOrderBook       bool            `json:"enableOrderBook"`
	AcceptingOrders       bool            `json:"acceptingOrders"`
	AcceptingOrdersTs     string          `json:"acceptingOrdersTimestamp"`
	ClobTokenIds          JSONStringArray `json:"clobTokenIds"`
	CreatedAt             time.Time       `json:"-"`
	UpdatedAt             time.Time       `json:"-"`

	// Computed fields
	YesPrice              float64         `json:"-"`
	NoPrice               float64         `json:"-"`
}

// Event represents a group of related markets.
type Event struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Slug            string    `json:"slug"`
	Description     string    `json:"description"`
	StartDate       string    `json:"startDate"`
	EndDate         string    `json:"endDate"`
	Image           string    `json:"image"`
	Icon            string    `json:"icon"`
	Active          bool      `json:"active"`
	Closed          bool      `json:"closed"`
	Archived        bool      `json:"archived"`
	Liquidity       float64   `json:"liquidity"`
	Volume          float64   `json:"volume"`
	Volume24hr      float64   `json:"volume24hr"`
	Markets         []Market  `json:"markets"`
	CompetitorCount int       `json:"competitorCount"`
	CommentCount    int       `json:"commentCount"`
	Tags            []Tag     `json:"tags"`
	CreatedAt       time.Time `json:"-"`
}

// Tag represents a category tag.
type Tag struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Trade represents a single trade.
type Trade struct {
	ID            string    `json:"id"`
	TakerOrderID  string    `json:"taker_order_id"`
	MarketID      string    `json:"market"`
	AssetID       string    `json:"asset_id"`
	Side          string    `json:"side"`
	Size          string    `json:"size"`
	Price         string    `json:"price"`
	Outcome       string    `json:"outcome"`
	FeeRateBps    string    `json:"fee_rate_bps"`
	Timestamp     int64     `json:"timestamp"`
	TransactionID string    `json:"transaction_hash"`
}

// MarketFilters represents filters for market queries.
type MarketFilters struct {
	Active      *bool
	Closed      *bool
	Limit       int
	Offset      int
	Order       string // "volume", "liquidity", "created_at", etc.
	Ascending   bool
	TagSlug     string
	TextQuery   string
}

// EventFilters represents filters for event queries.
type EventFilters struct {
	Active    *bool
	Closed    *bool
	Archived  *bool
	Limit     int
	Offset    int
	Order     string
	Ascending bool
	TagSlug   string
	TextQuery string
}

// GetMarkets retrieves markets from Gamma API.
func (c *Client) GetMarkets(ctx context.Context, filters MarketFilters) ([]Market, error) {
	params := url.Values{}

	if filters.Active != nil {
		params.Set("active", strconv.FormatBool(*filters.Active))
	}
	if filters.Closed != nil {
		params.Set("closed", strconv.FormatBool(*filters.Closed))
	}
	if filters.Limit > 0 {
		params.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Offset > 0 {
		params.Set("offset", strconv.Itoa(filters.Offset))
	}
	if filters.Order != "" {
		params.Set("order", filters.Order)
	}
	// Always set ascending parameter when ordering - Polymarket API defaults to ascending=true
	if filters.Order != "" {
		params.Set("ascending", strconv.FormatBool(filters.Ascending))
	}
	if filters.TagSlug != "" {
		params.Set("tag_slug", filters.TagSlug)
	}
	if filters.TextQuery != "" {
		params.Set("_q", filters.TextQuery)
	}

	log.Debug().
		Str("endpoint", "/markets").
		Str("params", params.Encode()).
		Msg("Fetching markets from Gamma API")

	resp, err := c.gamma.R().
		SetContext(ctx).
		SetQueryParamsFromValues(params).
		Get("/markets")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch markets: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("markets API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var markets []Market
	if err := json.Unmarshal(resp.Body(), &markets); err != nil {
		return nil, fmt.Errorf("failed to parse markets: %w", err)
	}

	// Parse outcome prices
	for i := range markets {
		if len(markets[i].OutcomePrices) >= 2 {
			markets[i].YesPrice, _ = strconv.ParseFloat(markets[i].OutcomePrices[0], 64)
			markets[i].NoPrice, _ = strconv.ParseFloat(markets[i].OutcomePrices[1], 64)
		}
	}

	log.Debug().
		Int("count", len(markets)).
		Msg("Fetched markets")

	return markets, nil
}

// GetMarket retrieves a single market by ID.
func (c *Client) GetMarket(ctx context.Context, marketID string) (*Market, error) {
	resp, err := c.gamma.R().
		SetContext(ctx).
		Get("/markets/" + marketID)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch market: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("market API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var market Market
	if err := json.Unmarshal(resp.Body(), &market); err != nil {
		return nil, fmt.Errorf("failed to parse market: %w", err)
	}

	// Parse outcome prices
	if len(market.OutcomePrices) >= 2 {
		market.YesPrice, _ = strconv.ParseFloat(market.OutcomePrices[0], 64)
		market.NoPrice, _ = strconv.ParseFloat(market.OutcomePrices[1], 64)
	}

	return &market, nil
}

// GetEvents retrieves events from Gamma API.
func (c *Client) GetEvents(ctx context.Context, filters EventFilters) ([]Event, error) {
	params := url.Values{}

	if filters.Active != nil {
		params.Set("active", strconv.FormatBool(*filters.Active))
	}
	if filters.Closed != nil {
		params.Set("closed", strconv.FormatBool(*filters.Closed))
	}
	if filters.Archived != nil {
		params.Set("archived", strconv.FormatBool(*filters.Archived))
	}
	if filters.Limit > 0 {
		params.Set("limit", strconv.Itoa(filters.Limit))
	}
	if filters.Offset > 0 {
		params.Set("offset", strconv.Itoa(filters.Offset))
	}
	if filters.Order != "" {
		params.Set("order", filters.Order)
	}
	// Always set ascending parameter when ordering - Polymarket API defaults to ascending=true
	if filters.Order != "" {
		params.Set("ascending", strconv.FormatBool(filters.Ascending))
	}
	if filters.TagSlug != "" {
		params.Set("tag_slug", filters.TagSlug)
	}
	if filters.TextQuery != "" {
		params.Set("_q", filters.TextQuery)
	}

	resp, err := c.gamma.R().
		SetContext(ctx).
		SetQueryParamsFromValues(params).
		Get("/events")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch events: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("events API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var events []Event
	if err := json.Unmarshal(resp.Body(), &events); err != nil {
		return nil, fmt.Errorf("failed to parse events: %w", err)
	}

	log.Debug().
		Int("count", len(events)).
		Msg("Fetched events")

	return events, nil
}

// GetEvent retrieves a single event by slug.
func (c *Client) GetEvent(ctx context.Context, slug string) (*Event, error) {
	resp, err := c.gamma.R().
		SetContext(ctx).
		Get("/events/slug/" + slug)

	if err != nil {
		return nil, fmt.Errorf("failed to fetch event: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("event API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var event Event
	if err := json.Unmarshal(resp.Body(), &event); err != nil {
		return nil, fmt.Errorf("failed to parse event: %w", err)
	}

	return &event, nil
}

// GetTrades retrieves recent trades from Data API.
func (c *Client) GetTrades(ctx context.Context, marketID string, limit int) ([]Trade, error) {
	params := url.Values{}
	params.Set("market", marketID)
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	resp, err := c.data.R().
		SetContext(ctx).
		SetQueryParamsFromValues(params).
		Get("/trades")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch trades: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("trades API returned %d: %s", resp.StatusCode(), resp.String())
	}

	var trades []Trade
	if err := json.Unmarshal(resp.Body(), &trades); err != nil {
		return nil, fmt.Errorf("failed to parse trades: %w", err)
	}

	return trades, nil
}

// GetTopMarketsByVolume retrieves top markets by 24h volume.
func (c *Client) GetTopMarketsByVolume(ctx context.Context, limit int) ([]Market, error) {
	active := true
	closed := false

	return c.GetMarkets(ctx, MarketFilters{
		Active:    &active,
		Closed:    &closed,
		Limit:     limit,
		Order:     "volume24hr",
		Ascending: false, // Descending order to get highest volume first
	})
}

// GetActiveEventsByCategory retrieves active events for a category.
func (c *Client) GetActiveEventsByCategory(ctx context.Context, category string, limit int) ([]Event, error) {
	active := true
	closed := false

	return c.GetEvents(ctx, EventFilters{
		Active:  &active,
		Closed:  &closed,
		Limit:   limit,
		TagSlug: category,
		Order:   "volume24hr",
	})
}

// SearchMarkets searches for markets by text query.
func (c *Client) SearchMarkets(ctx context.Context, query string, limit int) ([]Market, error) {
	return c.GetMarkets(ctx, MarketFilters{
		TextQuery: query,
		Limit:     limit,
	})
}
