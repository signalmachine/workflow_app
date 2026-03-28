package core

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RuleEngine resolves configurable account mappings from the account_rules table.
// It replaces hardcoded account constants in domain services.
type RuleEngine interface {
	ResolveAccount(ctx context.Context, companyID int, ruleType string) (string, error)
}

type ruleEngine struct {
	pool *pgxpool.Pool
}

// NewRuleEngine constructs a RuleEngine backed by the account_rules table.
func NewRuleEngine(pool *pgxpool.Pool) RuleEngine {
	return &ruleEngine{pool: pool}
}

// ResolveAccount returns the account code for (companyID, ruleType), highest priority first.
// Returns a descriptive error if no active rule exists.
func (r *ruleEngine) ResolveAccount(ctx context.Context, companyID int, ruleType string) (string, error) {
	var accountCode string
	err := r.pool.QueryRow(ctx, `
		SELECT COALESCE(a.code, ar.account_code)
		FROM account_rules ar
		LEFT JOIN accounts a
		  ON a.id = ar.account_id
		 AND a.company_id = ar.company_id
		WHERE ar.company_id = $1
		  AND ar.rule_type = $2
		  AND ar.effective_from <= CURRENT_DATE
		  AND (ar.effective_to IS NULL OR ar.effective_to >= CURRENT_DATE)
		ORDER BY ar.priority DESC, ar.effective_from DESC, ar.id DESC
		LIMIT 1
	`, companyID, ruleType).Scan(&accountCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("no account rule found for company_id %d, rule_type %q — seed account_rules or run migrations", companyID, ruleType)
		}
		return "", fmt.Errorf("failed to resolve account rule (company_id=%d, rule_type=%q): %w", companyID, ruleType, err)
	}
	return accountCode, nil
}
