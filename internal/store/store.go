package store

import (
	"context"
	"time"
)

type Account struct {
	ID            int64
	Name          string // user-chosen label, e.g. "alt1"
	AuthToken     string
	CSRFToken     string
	CFClearance   string
	Status        string // "active", "rate_limited", "disabled"
	CooldownUntil time.Time
	CreatedAt     time.Time
	LastUsedAt    time.Time
}

type WatchedUser struct {
	ID            int64
	UserID        string
	ScreenName    string
	Name          string
	Tags          []string
	CheckInterval int // seconds, 0 = use global default
	CreatedAt     time.Time
	LastCheckedAt time.Time
	LatestTweetID string
}

type StoredTweet struct {
	ID             int64
	TweetID        string
	UserID         string
	ScreenName     string
	OriginalText   string
	TranslatedText string
	Lang           string
	TweetURL       string
	MediaURLs      string // JSON array
	CreatedAt      time.Time
	FetchedAt      time.Time
}

type Store interface {
	AddWatchedUser(ctx context.Context, user WatchedUser) error
	RemoveWatchedUser(ctx context.Context, screenName string) error
	GetWatchedUsers(ctx context.Context) ([]WatchedUser, error)
	GetWatchedUser(ctx context.Context, screenName string) (*WatchedUser, error)
	UpdateLastChecked(ctx context.Context, screenName string, latestTweetID string) error
	UpdateCheckInterval(ctx context.Context, screenName string, interval int) error
	AddUserTags(ctx context.Context, screenName string, tags []string) error
	RemoveUserTags(ctx context.Context, screenName string, tags []string) error
	SetUserTags(ctx context.Context, screenName string, tags []string) error
	ListTags(ctx context.Context) (map[string]int, error)
	GetWatchedUsersByTag(ctx context.Context, tag string) ([]WatchedUser, error)

	SaveTweet(ctx context.Context, tweet StoredTweet) error
	TweetExists(ctx context.Context, tweetID string) (bool, error)
	GetTweets(ctx context.Context, screenName string, since time.Time, limit int) ([]StoredTweet, error)
	GetTweetsByTag(ctx context.Context, tag string, since time.Time, limit int) ([]StoredTweet, error)
	GetAllTweets(ctx context.Context, since time.Time, limit int) ([]StoredTweet, error)

	AddAccount(ctx context.Context, acct Account) error
	RemoveAccount(ctx context.Context, name string) error
	GetAccount(ctx context.Context, name string) (*Account, error)
	GetAccounts(ctx context.Context) ([]Account, error)
	GetActiveAccounts(ctx context.Context) ([]Account, error)
	UpdateAccountStatus(ctx context.Context, name string, status string, cooldownUntil *time.Time) error

	GetConfig(ctx context.Context, key string) (string, error)
	SetConfig(ctx context.Context, key string, value string) error
	GetAllConfig(ctx context.Context) (map[string]string, error)

	Close() error
}
