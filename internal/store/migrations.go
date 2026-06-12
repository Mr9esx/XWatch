package store

const schemaSQL = `
CREATE TABLE IF NOT EXISTS watched_users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id         TEXT NOT NULL UNIQUE,
    screen_name     TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL DEFAULT '',
    check_interval  INTEGER NOT NULL DEFAULT 0,
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_checked_at DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
    latest_tweet_id TEXT NOT NULL DEFAULT '',
    tags            TEXT NOT NULL DEFAULT '[]'
);

CREATE TABLE IF NOT EXISTS tweets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    tweet_id        TEXT NOT NULL UNIQUE,
    user_id         TEXT NOT NULL,
    screen_name     TEXT NOT NULL,
    original_text   TEXT NOT NULL,
    translated_text TEXT NOT NULL DEFAULT '',
    lang            TEXT NOT NULL DEFAULT '',
    tweet_url       TEXT NOT NULL DEFAULT '',
    media_urls      TEXT NOT NULL DEFAULT '[]',
    created_at      DATETIME NOT NULL,
    fetched_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tweets_user_id ON tweets(user_id);
CREATE INDEX IF NOT EXISTS idx_tweets_created_at ON tweets(created_at);
CREATE INDEX IF NOT EXISTS idx_tweets_tweet_id ON tweets(tweet_id);

CREATE TABLE IF NOT EXISTS config (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL UNIQUE,
    auth_token      TEXT NOT NULL,
    csrf_token      TEXT NOT NULL,
    cf_clearance    TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'active',
    cooldown_until  DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00',
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at    DATETIME NOT NULL DEFAULT '1970-01-01 00:00:00'
);
`

const defaultConfigSQL = `
INSERT OR IGNORE INTO config (key, value) VALUES
    ('default_check_interval', '300'),
    ('proxy', ''),
    ('x_auth_token', ''),
    ('x_ct0', ''),
    ('x_username', ''),
    ('x_password', '');
`
