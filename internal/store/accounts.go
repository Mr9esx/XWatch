package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func (s *SQLiteStore) AddAccount(ctx context.Context, acct Account) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO accounts (name, auth_token, csrf_token, cf_clearance) VALUES (?, ?, ?, ?)`,
		acct.Name, acct.AuthToken, acct.CSRFToken, acct.CFClearance,
	)
	if err != nil {
		return fmt.Errorf("add account: %w", err)
	}
	return nil
}

func (s *SQLiteStore) RemoveAccount(ctx context.Context, name string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM accounts WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("remove account: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("account %q not found", name)
	}
	return nil
}

func (s *SQLiteStore) GetAccount(ctx context.Context, name string) (*Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, auth_token, csrf_token, cf_clearance, status, cooldown_until, created_at, last_used_at
		 FROM accounts WHERE name = ?`, name,
	).Scan(&a.ID, &a.Name, &a.AuthToken, &a.CSRFToken, &a.CFClearance,
		&a.Status, &a.CooldownUntil, &a.CreatedAt, &a.LastUsedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get account: %w", err)
	}
	return &a, nil
}

func (s *SQLiteStore) GetAccounts(ctx context.Context) ([]Account, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, auth_token, csrf_token, cf_clearance, status, cooldown_until, created_at, last_used_at
		 FROM accounts ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("query accounts: %w", err)
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *SQLiteStore) GetActiveAccounts(ctx context.Context) ([]Account, error) {
	now := time.Now().UTC()
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, auth_token, csrf_token, cf_clearance, status, cooldown_until, created_at, last_used_at
		 FROM accounts
		 WHERE status = 'active' OR (status = 'rate_limited' AND cooldown_until <= ?)
		 ORDER BY last_used_at ASC`, now)
	if err != nil {
		return nil, fmt.Errorf("query active accounts: %w", err)
	}
	defer rows.Close()
	return scanAccounts(rows)
}

func (s *SQLiteStore) UpdateAccountStatus(ctx context.Context, name string, status string, cooldownUntil *time.Time) error {
	cd := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if cooldownUntil != nil {
		cd = *cooldownUntil
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE accounts SET status = ?, cooldown_until = ?, last_used_at = ? WHERE name = ?`,
		status, cd, time.Now().UTC(), name,
	)
	if err != nil {
		return fmt.Errorf("update account status: %w", err)
	}
	return nil
}

func scanAccounts(rows *sql.Rows) ([]Account, error) {
	var accounts []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.Name, &a.AuthToken, &a.CSRFToken, &a.CFClearance,
			&a.Status, &a.CooldownUntil, &a.CreatedAt, &a.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan account: %w", err)
		}
		accounts = append(accounts, a)
	}
	return accounts, rows.Err()
}
