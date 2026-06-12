package config

import (
	"context"
	"os"
	"path/filepath"
	"strconv"

	"github.com/mr9esx/xwatch/internal/store"
)

const (
	KeyDefaultCheckInterval = "default_check_interval"
	KeyProxy                = "proxy"
	KeyXAuthToken           = "x_auth_token"
	KeyXCT0                 = "x_ct0"
	KeyXCFClearance         = "cf_clearance"
	KeyXUsername             = "x_username"
	KeyXPassword            = "x_password"
	KeyRequestDelay         = "request_delay"
)

func DefaultDBPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".xwatch", "xwatch.db")
}

func EnsureDataDir() error {
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, ".xwatch")
	return os.MkdirAll(dir, 0755)
}

type Manager struct {
	store store.Store
}

func NewManager(s store.Store) *Manager {
	return &Manager{store: s}
}

func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	return m.store.GetConfig(ctx, key)
}

func (m *Manager) Set(ctx context.Context, key, value string) error {
	return m.store.SetConfig(ctx, key, value)
}

func (m *Manager) GetAll(ctx context.Context) (map[string]string, error) {
	return m.store.GetAllConfig(ctx)
}

func (m *Manager) GetInt(ctx context.Context, key string, defaultVal int) int {
	v, err := m.store.GetConfig(ctx, key)
	if err != nil || v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

func (m *Manager) GetDefaultCheckInterval(ctx context.Context) int {
	return m.GetInt(ctx, KeyDefaultCheckInterval, 300)
}

func (m *Manager) GetProxy(ctx context.Context) string {
	v, _ := m.store.GetConfig(ctx, KeyProxy)
	return v
}

func (m *Manager) GetRequestDelay(ctx context.Context) int {
	return m.GetInt(ctx, KeyRequestDelay, 3)
}
