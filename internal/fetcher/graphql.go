package fetcher

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	twitterscraper "github.com/imperatrona/twitter-scraper"
)

type GraphQLFetcher struct {
	scraper *twitterscraper.Scraper
}

type GraphQLConfig struct {
	AuthToken    string
	CSRFToken    string
	CFClearance  string
	ProxyURL     string
	RequestDelay int // seconds between API requests, 0 = default (3s)
}

func NewGraphQLFetcher(cfg GraphQLConfig) (*GraphQLFetcher, error) {
	scraper := twitterscraper.New()

	if cfg.ProxyURL != "" {
		if err := scraper.SetProxy(cfg.ProxyURL); err != nil {
			return nil, fmt.Errorf("set proxy: %w", err)
		}
	}

	if cfg.AuthToken != "" && cfg.CSRFToken != "" {
		scraper.SetAuthToken(twitterscraper.AuthToken{
			Token:     cfg.AuthToken,
			CSRFToken: cfg.CSRFToken,
		})
		if cfg.CFClearance != "" {
			setCookie(scraper, "cf_clearance", cfg.CFClearance)
		}
		scraper.IsLoggedIn()
		scraper.SetSearchMode(twitterscraper.SearchUsers)
	}

	delay := int64(cfg.RequestDelay)
	if delay <= 0 {
		delay = 3
	}
	scraper.WithDelay(delay)

	return &GraphQLFetcher{scraper: scraper}, nil
}

func setCookie(scraper *twitterscraper.Scraper, name, value string) {
	expires := time.Date(2030, time.January, 1, 0, 0, 0, 0, time.UTC)
	scraper.SetCookies([]*http.Cookie{{
		Name: name, Value: value, Domain: "x.com", Path: "/",
		Expires: expires, Secure: true,
	}})
}

func (f *GraphQLFetcher) SearchUsers(ctx context.Context, query string, limit int) ([]User, error) {
	ch := f.scraper.SearchProfiles(ctx, query, limit)

	var users []User
	for profile := range ch {
		if profile.Error != nil {
			continue
		}
		users = append(users, User{
			ID:              profile.UserID,
			ScreenName:      profile.Username,
			Name:            profile.Name,
			Description:     profile.Biography,
			FollowersCount:  profile.FollowersCount,
			ProfileImageURL: profile.Avatar,
		})
		if len(users) >= limit {
			break
		}
	}

	if len(users) == 0 {
		screenName := strings.TrimPrefix(strings.TrimSpace(query), "@")
		if screenName != "" {
			profile, err := f.scraper.GetProfile(screenName)
			if err == nil && profile.Username != "" {
				users = append(users, userFromProfile(profile))
			}
		}
	}

	return users, nil
}

func userFromProfile(profile twitterscraper.Profile) User {
	return User{
		ID:              profile.UserID,
		ScreenName:      profile.Username,
		Name:            profile.Name,
		Description:     profile.Biography,
		FollowersCount:  profile.FollowersCount,
		ProfileImageURL: profile.Avatar,
	}
}

func (f *GraphQLFetcher) GetUserByScreenName(ctx context.Context, screenName string) (*User, error) {
	profile, err := f.scraper.GetProfile(screenName)
	if err != nil {
		return nil, fmt.Errorf("get profile @%s: %w", screenName, err)
	}

	user := userFromProfile(profile)
	return &user, nil
}

func (f *GraphQLFetcher) GetUserTweets(ctx context.Context, userID string, sinceID string, limit int) ([]Tweet, error) {
	screenName := userID
	tweetCh := f.scraper.GetTweets(ctx, screenName, limit)

	var tweets []Tweet
	for t := range tweetCh {
		if t.Error != nil {
			if len(tweets) > 0 {
				break
			}
			return nil, fmt.Errorf("get tweets: %w", t.Error)
		}

		if sinceID != "" && t.ID <= sinceID {
			continue
		}

		tweet := Tweet{
			ID:        t.ID,
			UserID:    t.UserID,
			Text:      t.Text,
			CreatedAt: time.Unix(t.Timestamp, 0),
			URL:       fmt.Sprintf("https://x.com/%s/status/%s", t.Username, t.ID),
			IsRetweet: t.IsRetweet,
			IsReply:   t.IsReply,
		}

		if t.IsRetweet && t.RetweetedStatus != nil {
			rt := t.RetweetedStatus
			tweet.RetweetedScreenName = rt.Username
			tweet.Text = rt.Text
			tweet.URL = fmt.Sprintf("https://x.com/%s/status/%s", rt.Username, rt.ID)
			for _, photo := range rt.Photos {
				tweet.MediaURLs = append(tweet.MediaURLs, photo.URL)
			}
			for _, video := range rt.Videos {
				tweet.MediaURLs = append(tweet.MediaURLs, video.Preview)
			}
		} else {
			for _, photo := range t.Photos {
				tweet.MediaURLs = append(tweet.MediaURLs, photo.URL)
			}
			for _, video := range t.Videos {
				tweet.MediaURLs = append(tweet.MediaURLs, video.Preview)
			}
		}

		tweets = append(tweets, tweet)
	}

	return tweets, nil
}
