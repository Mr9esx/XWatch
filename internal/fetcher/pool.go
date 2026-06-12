package fetcher

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type AccountEntry struct {
	Name        string
	Fetcher     *GraphQLFetcher
	RateLimited bool
	CooldownAt  time.Time
}

type PoolFetcher struct {
	mu       sync.Mutex
	accounts []*AccountEntry
	cursor   int
	onStatus func(name, status string, cooldownUntil *time.Time)
}

type PoolConfig struct {
	OnStatusChange func(name, status string, cooldownUntil *time.Time)
}

func NewPoolFetcher(cfg PoolConfig) *PoolFetcher {
	return &PoolFetcher{
		onStatus: cfg.OnStatusChange,
	}
}

func (p *PoolFetcher) AddAccount(name string, f *GraphQLFetcher) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.accounts = append(p.accounts, &AccountEntry{Name: name, Fetcher: f})
}

func (p *PoolFetcher) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.accounts)
}

func (p *PoolFetcher) pick() (*AccountEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	n := len(p.accounts)
	if n == 0 {
		return nil, fmt.Errorf("账号池为空")
	}

	now := time.Now()
	for i := 0; i < n; i++ {
		idx := (p.cursor + i) % n
		entry := p.accounts[idx]
		if entry.RateLimited && now.Before(entry.CooldownAt) {
			continue
		}
		if entry.RateLimited {
			entry.RateLimited = false
		}
		p.cursor = (idx + 1) % n
		return entry, nil
	}

	earliest := p.accounts[0].CooldownAt
	for _, e := range p.accounts[1:] {
		if e.CooldownAt.Before(earliest) {
			earliest = e.CooldownAt
		}
	}
	return nil, fmt.Errorf("所有账号均被限流，最早恢复: %s", earliest.Local().Format("15:04:05"))
}

func (p *PoolFetcher) markRateLimited(entry *AccountEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry.RateLimited = true
	entry.CooldownAt = time.Now().Add(15 * time.Minute)
	if p.onStatus != nil {
		cd := entry.CooldownAt
		p.onStatus(entry.Name, "rate_limited", &cd)
	}
}

func (p *PoolFetcher) markDisabled(entry *AccountEntry) {
	p.mu.Lock()
	defer p.mu.Unlock()
	entry.RateLimited = true
	entry.CooldownAt = time.Now().Add(24 * time.Hour)
	if p.onStatus != nil {
		cd := entry.CooldownAt
		p.onStatus(entry.Name, "disabled", &cd)
	}
}

func (p *PoolFetcher) SearchUsers(ctx context.Context, query string, limit int) ([]User, error) {
	return withRetry(p, func(f *GraphQLFetcher) ([]User, error) {
		return f.SearchUsers(ctx, query, limit)
	})
}

func (p *PoolFetcher) GetUserByScreenName(ctx context.Context, screenName string) (*User, error) {
	return withRetry(p, func(f *GraphQLFetcher) (*User, error) {
		return f.GetUserByScreenName(ctx, screenName)
	})
}

func (p *PoolFetcher) GetUserTweets(ctx context.Context, userID string, sinceID string, limit int) ([]Tweet, error) {
	return withRetry(p, func(f *GraphQLFetcher) ([]Tweet, error) {
		return f.GetUserTweets(ctx, userID, sinceID, limit)
	})
}

func withRetry[T any](p *PoolFetcher, fn func(*GraphQLFetcher) (T, error)) (T, error) {
	tried := 0
	maxTries := p.Size()
	if maxTries < 1 {
		maxTries = 1
	}

	for tried < maxTries {
		entry, err := p.pick()
		if err != nil {
			var zero T
			return zero, err
		}

		result, err := fn(entry.Fetcher)
		if err == nil {
			return result, nil
		}

		if IsRateLimitError(err) {
			p.markRateLimited(entry)
			tried++
			continue
		}
		if IsAuthError(err) {
			p.markDisabled(entry)
			tried++
			continue
		}

		var zero T
		return zero, err
	}

	var zero T
	return zero, fmt.Errorf("所有账号均不可用")
}
