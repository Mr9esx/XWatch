package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/mr9esx/xwatch/internal/config"
	"github.com/mr9esx/xwatch/internal/fetcher"
	"github.com/mr9esx/xwatch/internal/store"
)

type Deps struct {
	Store   store.Store
	Config  *config.Manager
	Fetcher fetcher.Fetcher
}

func InitDeps(dbPath string) (*Deps, error) {
	if dbPath == "" {
		dbPath = config.DefaultDBPath()
	}

	if err := config.EnsureDataDir(); err != nil {
		return nil, fmt.Errorf("ensure data dir: %w", err)
	}

	s, err := store.NewSQLiteStore(dbPath)
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}

	cfg := config.NewManager(s)
	ctx := context.Background()

	proxyURL := cfg.GetProxy(ctx)
	requestDelay := cfg.GetRequestDelay(ctx)

	var f fetcher.Fetcher

	accounts, _ := s.GetActiveAccounts(ctx)
	if len(accounts) > 0 {
		pool := fetcher.NewPoolFetcher(fetcher.PoolConfig{
			OnStatusChange: func(name, status string, cooldownUntil *time.Time) {
				_ = s.UpdateAccountStatus(ctx, name, status, cooldownUntil)
			},
		})
		for _, acct := range accounts {
			gf, err := fetcher.NewGraphQLFetcher(fetcher.GraphQLConfig{
				AuthToken:    acct.AuthToken,
				CSRFToken:    acct.CSRFToken,
				CFClearance:  acct.CFClearance,
				ProxyURL:     proxyURL,
				RequestDelay: requestDelay,
			})
			if err != nil {
				fmt.Printf("警告: 账号 %q 初始化失败: %v\n", acct.Name, err)
				continue
			}
			pool.AddAccount(acct.Name, gf)
		}
		if pool.Size() > 0 {
			f = pool
			fmt.Printf("账号池已启用 (%d 个账号)\n", pool.Size())
		}
	}

	if f == nil {
		authToken, _ := s.GetConfig(ctx, config.KeyXAuthToken)
		csrfToken, _ := s.GetConfig(ctx, config.KeyXCT0)
		cfClearance, _ := s.GetConfig(ctx, config.KeyXCFClearance)

		if authToken != "" && csrfToken != "" {
			gf, err := fetcher.NewGraphQLFetcher(fetcher.GraphQLConfig{
				AuthToken:    authToken,
				CSRFToken:    csrfToken,
				CFClearance:  cfClearance,
				ProxyURL:     proxyURL,
				RequestDelay: requestDelay,
			})
			if err != nil {
				fmt.Printf("警告: Twitter 认证初始化失败 (%v)，部分功能不可用\n", err)
			} else {
				f = gf
			}
		}
	}

	return &Deps{
		Store:   s,
		Config:  cfg,
		Fetcher: f,
	}, nil
}
