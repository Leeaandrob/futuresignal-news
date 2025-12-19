// Package xtracker provides integration with the XTracker API for social signal tracking.
package xtracker

import (
	"fmt"
	"time"
)

// TrackedUser represents a user being tracked on XTracker.
type TrackedUser struct {
	ID         string    `json:"id"`
	Handle     string    `json:"handle"`
	Name       string    `json:"name"`
	PlatformID string    `json:"platformId"`
	AvatarURL  string    `json:"avatarUrl"`
	Bio        string    `json:"bio"`
	Verified   bool      `json:"verified"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	PostCount  int       `json:"-"` // Populated from _count.posts
}

// Post represents a tracked post/tweet from XTracker.
type Post struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	PlatformID string    `json:"platformId"` // Tweet ID for URL construction
	Content    string    `json:"content"`
	CreatedAt  time.Time `json:"createdAt"`
	ImportedAt time.Time `json:"importedAt"`
}

// TweetURL returns the full URL to the tweet on X/Twitter.
func (p *Post) TweetURL(handle string) string {
	return fmt.Sprintf("https://x.com/%s/status/%s", handle, p.PlatformID)
}

// Tracking represents a tracking period for a user (related to Polymarket markets).
type Tracking struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Title      string    `json:"title"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	IsActive   bool      `json:"isActive"`
	MarketLink string    `json:"marketLink,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// DailyMetric represents daily tweet count metrics.
type DailyMetric struct {
	ID     string    `json:"id"`
	UserID string    `json:"userId"`
	Date   time.Time `json:"date"`
	Type   string    `json:"type"` // "daily", "hourly"
	Data   struct {
		Count      int    `json:"count"`
		Cumulative int    `json:"cumulative"`
		TrackingID string `json:"trackingId"`
	} `json:"data"`
}

// TrackingStats represents statistics for a tracking period.
type TrackingStats struct {
	Total           int     `json:"total"`
	Cumulative      int     `json:"cumulative"`
	Pace            int     `json:"pace"`
	PercentComplete float64 `json:"percentComplete"`
	DaysElapsed     int     `json:"daysElapsed"`
	DaysRemaining   int     `json:"daysRemaining"`
	DaysTotal       int     `json:"daysTotal"`
	IsComplete      bool    `json:"isComplete"`
}

// Note: SocialSignal and MarketMovement types are defined in models package
// to avoid circular dependencies. Use models.SocialSignal and models.MarketMovement.

// API Response types

type UsersResponse struct {
	Success bool          `json:"success"`
	Data    []TrackedUser `json:"data"`
}

type UserResponse struct {
	Success bool        `json:"success"`
	Data    TrackedUser `json:"data"`
}

type PostsResponse struct {
	Success bool   `json:"success"`
	Data    []Post `json:"data"`
}

type TrackingsResponse struct {
	Success bool       `json:"success"`
	Data    []Tracking `json:"data"`
}

type MetricsResponse struct {
	Success bool          `json:"success"`
	Data    []DailyMetric `json:"data"`
}

// UserWithCount is used for parsing the nested _count structure.
type UserWithCount struct {
	TrackedUser
	Count struct {
		Posts int `json:"posts"`
	} `json:"_count"`
}

func (u *UserWithCount) ToTrackedUser() TrackedUser {
	user := u.TrackedUser
	user.PostCount = u.Count.Posts
	return user
}
