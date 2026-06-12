package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	db.SetMaxOpenConns(1)

	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	if _, err := db.Exec(defaultConfigSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("insert default config: %w", err)
	}

	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return s, nil
}

func (s *SQLiteStore) migrate() error {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('watched_users') WHERE name = 'tags'`).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		if _, err := s.db.Exec(`ALTER TABLE watched_users ADD COLUMN tags TEXT NOT NULL DEFAULT '[]'`); err != nil {
			return err
		}
	}

	var acctCount int
	err = s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='accounts'`).Scan(&acctCount)
	if err != nil {
		return err
	}
	if acctCount == 0 {
		if _, err := s.db.Exec(`CREATE TABLE accounts (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			name            TEXT NOT NULL UNIQUE,
			auth_token      TEXT NOT NULL,
			csrf_token      TEXT NOT NULL,
			cf_clearance    TEXT NOT NULL DEFAULT '',
			status          TEXT NOT NULL DEFAULT 'active',
			cooldown_until  DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			last_used_at    DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00'
		)`); err != nil {
			return err
		}
	}
	return nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

// --- Watched Users ---

func (s *SQLiteStore) AddWatchedUser(ctx context.Context, user WatchedUser) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO watched_users (user_id, screen_name, name, check_interval, tags) VALUES (?, ?, ?, ?, ?)`,
		user.UserID, user.ScreenName, user.Name, user.CheckInterval, encodeTags(user.Tags),
	)
	if err != nil {
		return fmt.Errorf("add watched user: %w", err)
	}
	return nil
}

func (s *SQLiteStore) RemoveWatchedUser(ctx context.Context, screenName string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM watched_users WHERE screen_name = ?`, screenName)
	if err != nil {
		return fmt.Errorf("remove watched user: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user @%s not found", screenName)
	}
	return nil
}

func (s *SQLiteStore) GetWatchedUsers(ctx context.Context) ([]WatchedUser, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, screen_name, name, check_interval, created_at, last_checked_at, latest_tweet_id, tags
		 FROM watched_users ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query watched users: %w", err)
	}
	defer rows.Close()
	return scanWatchedUsers(rows)
}

func (s *SQLiteStore) GetWatchedUser(ctx context.Context, screenName string) (*WatchedUser, error) {
	var u WatchedUser
	var tagsRaw string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, screen_name, name, check_interval, created_at, last_checked_at, latest_tweet_id, tags
		 FROM watched_users WHERE screen_name = ?`, screenName,
	).Scan(&u.ID, &u.UserID, &u.ScreenName, &u.Name,
		&u.CheckInterval, &u.CreatedAt, &u.LastCheckedAt, &u.LatestTweetID, &tagsRaw)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get watched user: %w", err)
	}
	u.Tags = decodeTags(tagsRaw)
	return &u, nil
}

func (s *SQLiteStore) GetWatchedUsersByTag(ctx context.Context, tag string) ([]WatchedUser, error) {
	users, err := s.GetWatchedUsers(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []WatchedUser
	for _, u := range users {
		if userHasTag(u.Tags, tag) {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}

func (s *SQLiteStore) AddUserTags(ctx context.Context, screenName string, tags []string) error {
	user, err := s.GetWatchedUser(ctx, screenName)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user @%s not found", screenName)
	}
	return s.SetUserTags(ctx, screenName, mergeTags(user.Tags, tags))
}

func (s *SQLiteStore) RemoveUserTags(ctx context.Context, screenName string, tags []string) error {
	user, err := s.GetWatchedUser(ctx, screenName)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user @%s not found", screenName)
	}
	return s.SetUserTags(ctx, screenName, removeTags(user.Tags, tags))
}

func (s *SQLiteStore) SetUserTags(ctx context.Context, screenName string, tags []string) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE watched_users SET tags = ? WHERE screen_name = ?`,
		encodeTags(tags), screenName,
	)
	if err != nil {
		return fmt.Errorf("set user tags: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user @%s not found", screenName)
	}
	return nil
}

func (s *SQLiteStore) ListTags(ctx context.Context) (map[string]int, error) {
	users, err := s.GetWatchedUsers(ctx)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, u := range users {
		for _, tag := range u.Tags {
			counts[tag]++
		}
	}
	return counts, nil
}

func scanWatchedUsers(rows *sql.Rows) ([]WatchedUser, error) {
	var users []WatchedUser
	for rows.Next() {
		var u WatchedUser
		var tagsRaw string
		if err := rows.Scan(&u.ID, &u.UserID, &u.ScreenName, &u.Name,
			&u.CheckInterval, &u.CreatedAt, &u.LastCheckedAt, &u.LatestTweetID, &tagsRaw); err != nil {
			return nil, fmt.Errorf("scan watched user: %w", err)
		}
		u.Tags = decodeTags(tagsRaw)
		users = append(users, u)
	}
	return users, rows.Err()
}

func (s *SQLiteStore) UpdateLastChecked(ctx context.Context, screenName string, latestTweetID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE watched_users SET last_checked_at = ?, latest_tweet_id = ? WHERE screen_name = ?`,
		time.Now().UTC(), latestTweetID, screenName,
	)
	if err != nil {
		return fmt.Errorf("update last checked: %w", err)
	}
	return nil
}

func (s *SQLiteStore) UpdateCheckInterval(ctx context.Context, screenName string, interval int) error {
	result, err := s.db.ExecContext(ctx,
		`UPDATE watched_users SET check_interval = ? WHERE screen_name = ?`,
		interval, screenName,
	)
	if err != nil {
		return fmt.Errorf("update check interval: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user @%s not found", screenName)
	}
	return nil
}

// --- Tweets ---

func (s *SQLiteStore) SaveTweet(ctx context.Context, tweet StoredTweet) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO tweets (tweet_id, user_id, screen_name, original_text, translated_text, lang, tweet_url, media_urls, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tweet.TweetID, tweet.UserID, tweet.ScreenName,
		tweet.OriginalText, tweet.TranslatedText, tweet.Lang,
		tweet.TweetURL, tweet.MediaURLs, tweet.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("save tweet: %w", err)
	}
	return nil
}

func (s *SQLiteStore) TweetExists(ctx context.Context, tweetID string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(1) FROM tweets WHERE tweet_id = ?`, tweetID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check tweet exists: %w", err)
	}
	return count > 0, nil
}

func (s *SQLiteStore) GetTweets(ctx context.Context, screenName string, since time.Time, limit int) ([]StoredTweet, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tweet_id, user_id, screen_name, original_text, translated_text, lang, tweet_url, media_urls, created_at, fetched_at
		 FROM tweets WHERE screen_name = ? AND created_at >= ? ORDER BY created_at DESC LIMIT ?`,
		screenName, since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query tweets: %w", err)
	}
	defer rows.Close()
	return scanTweets(rows)
}

func (s *SQLiteStore) GetTweetsByTag(ctx context.Context, tag string, since time.Time, limit int) ([]StoredTweet, error) {
	users, err := s.GetWatchedUsersByTag(ctx, tag)
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, nil
	}
	screenNames := make([]string, len(users))
	for i, u := range users {
		screenNames[i] = u.ScreenName
	}
	return s.getTweetsByScreenNames(ctx, screenNames, since, limit)
}

func (s *SQLiteStore) getTweetsByScreenNames(ctx context.Context, screenNames []string, since time.Time, limit int) ([]StoredTweet, error) {
	query := `SELECT id, tweet_id, user_id, screen_name, original_text, translated_text, lang, tweet_url, media_urls, created_at, fetched_at
		 FROM tweets WHERE created_at >= ? AND screen_name IN (`
	args := []interface{}{since}
	for i, name := range screenNames {
		if i > 0 {
			query += ","
		}
		query += "?"
		args = append(args, name)
	}
	query += ") ORDER BY created_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tweets by screen names: %w", err)
	}
	defer rows.Close()
	return scanTweets(rows)
}

func (s *SQLiteStore) GetAllTweets(ctx context.Context, since time.Time, limit int) ([]StoredTweet, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, tweet_id, user_id, screen_name, original_text, translated_text, lang, tweet_url, media_urls, created_at, fetched_at
		 FROM tweets WHERE created_at >= ? ORDER BY created_at DESC LIMIT ?`,
		since, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query all tweets: %w", err)
	}
	defer rows.Close()
	return scanTweets(rows)
}

func scanTweets(rows *sql.Rows) ([]StoredTweet, error) {
	var tweets []StoredTweet
	for rows.Next() {
		var t StoredTweet
		if err := rows.Scan(&t.ID, &t.TweetID, &t.UserID, &t.ScreenName,
			&t.OriginalText, &t.TranslatedText, &t.Lang,
			&t.TweetURL, &t.MediaURLs, &t.CreatedAt, &t.FetchedAt); err != nil {
			return nil, fmt.Errorf("scan tweet: %w", err)
		}
		tweets = append(tweets, t)
	}
	return tweets, rows.Err()
}

// --- Config ---

func (s *SQLiteStore) GetConfig(ctx context.Context, key string) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx, `SELECT value FROM config WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("get config %s: %w", key, err)
	}
	return value, nil
}

func (s *SQLiteStore) SetConfig(ctx context.Context, key string, value string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO config (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value`,
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set config %s: %w", key, err)
	}
	return nil
}

func (s *SQLiteStore) GetAllConfig(ctx context.Context) (map[string]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key, value FROM config ORDER BY key`)
	if err != nil {
		return nil, fmt.Errorf("query all config: %w", err)
	}
	defer rows.Close()

	cfg := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, fmt.Errorf("scan config: %w", err)
		}
		cfg[k] = v
	}
	return cfg, rows.Err()
}
