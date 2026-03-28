package core_test

import (
	"context"
	"testing"

	"accounting-agent/internal/core"
)

func TestRuleEngine_ResolveAccount(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// Seed rules for company 1
	_, err := pool.Exec(ctx, `
		INSERT INTO account_rules (company_id, rule_type, account_code, priority)
		VALUES
		  (1, 'AR',        '1200', 0),
		  (1, 'AP',        '2000', 0),
		  (1, 'HIGH_PRIO', '9999', 10),
		  (1, 'HIGH_PRIO', '8888', 0)
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("Failed to seed account_rules: %v", err)
	}

	re := core.NewRuleEngine(pool)

	t.Run("resolves AR", func(t *testing.T) {
		code, err := re.ResolveAccount(ctx, 1, "AR")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "1200" {
			t.Errorf("expected 1200, got %s", code)
		}
	})

	t.Run("resolves AP", func(t *testing.T) {
		code, err := re.ResolveAccount(ctx, 1, "AP")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "2000" {
			t.Errorf("expected 2000, got %s", code)
		}
	})

	t.Run("priority DESC picks highest priority row", func(t *testing.T) {
		code, err := re.ResolveAccount(ctx, 1, "HIGH_PRIO")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code != "9999" {
			t.Errorf("expected 9999 (priority 10), got %s", code)
		}
	})

	t.Run("missing rule returns descriptive error", func(t *testing.T) {
		_, err := re.ResolveAccount(ctx, 1, "NONEXISTENT_RULE")
		if err == nil {
			t.Error("expected error for missing rule, got nil")
		}
	})

	t.Run("company isolation — rule for company 1 not visible to company 2", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
			INSERT INTO companies (id, company_code, name, base_currency)
			VALUES (2, '2000', 'Company Two', 'USD')
			ON CONFLICT DO NOTHING;
		`)
		if err != nil {
			t.Fatalf("Failed to seed second company: %v", err)
		}

		_, err = re.ResolveAccount(ctx, 2, "AR")
		if err == nil {
			t.Error("expected error: company 2 has no AR rule")
		}
	})

	t.Run("respects effective_from and effective_to windows", func(t *testing.T) {
		_, err := pool.Exec(ctx, `
			INSERT INTO account_rules (company_id, rule_type, account_code, priority, effective_from, effective_to)
			VALUES
			  (1, 'TEMPORAL_RULE', '1200', 0, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '2 days'),
			  (1, 'TEMPORAL_RULE', '2000', 10, CURRENT_DATE + INTERVAL '1 day', NULL)
			ON CONFLICT DO NOTHING;
		`)
		if err != nil {
			t.Fatalf("seed temporal account rules: %v", err)
		}

		code, err := re.ResolveAccount(ctx, 1, "TEMPORAL_RULE")
		if err != nil {
			t.Fatalf("unexpected error resolving temporal rule: %v", err)
		}
		if code != "1200" {
			t.Errorf("expected currently-effective rule 1200, got %s", code)
		}
	})

	t.Run("prefers account_id-backed code when account_code is stale", func(t *testing.T) {
		var arAccountID int
		if err := pool.QueryRow(ctx, `
			SELECT id FROM accounts
			WHERE company_id = 1 AND code = '1200'
		`).Scan(&arAccountID); err != nil {
			t.Fatalf("lookup account id for code 1200: %v", err)
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO account_rules (company_id, rule_type, account_code, account_id, priority)
			VALUES (1, 'ACCOUNT_ID_PRECEDENCE', '2000', $1, 0)
		`, arAccountID)
		if err != nil {
			t.Fatalf("seed account_id precedence rule: %v", err)
		}

		code, err := re.ResolveAccount(ctx, 1, "ACCOUNT_ID_PRECEDENCE")
		if err != nil {
			t.Fatalf("unexpected error resolving ACCOUNT_ID_PRECEDENCE: %v", err)
		}
		if code != "1200" {
			t.Errorf("expected account code 1200 from account_id reference, got %s", code)
		}
	})
}
