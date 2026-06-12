package fetcher

import (
	"context"
	"strings"
	"time"
)

type User struct {
	ID              string `json:"id"`
	ScreenName      string `json:"screen_name"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	FollowersCount  int    `json:"followers_count"`
	ProfileImageURL string `json:"profile_image_url"`
}

type Tweet struct {
	ID                  string    `json:"id"`
	UserID              string    `json:"user_id"`
	Text                string    `json:"text"`
	CreatedAt           time.Time `json:"created_at"`
	URL                 string    `json:"url"`
	MediaURLs           []string  `json:"media_urls"`
	IsRetweet           bool      `json:"is_retweet"`
	IsReply             bool      `json:"is_reply"`
	RetweetedScreenName string    `json:"retweeted_screen_name,omitempty"`
}

// Fetcher abstracts the tweet source.
type Fetcher interface {
	SearchUsers(ctx context.Context, query string, limit int) ([]User, error)
	GetUserByScreenName(ctx context.Context, screenName string) (*User, error)
	GetUserTweets(ctx context.Context, userID string, sinceID string, limit int) ([]Tweet, error)
}

func IsRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "429") ||
		strings.Contains(msg, "Rate limit") ||
		strings.Contains(msg, "rate limit") ||
		strings.Contains(msg, "Too Many Requests")
}

func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "401") ||
		strings.Contains(msg, "403") ||
		strings.Contains(msg, "Unauthorized") ||
		strings.Contains(msg, "Forbidden")
}
