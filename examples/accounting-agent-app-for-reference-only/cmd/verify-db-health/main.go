package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type issue struct {
	CompanyCode string
	RuleType    string
	AccountCode string
	AccountName *string
	AccountType *string
}

type accountRuleIDCodeMismatch struct {
	CompanyCode string
	RuleType    string
	AccountCode string
	IDCode      string
}

func main() {
	_ = godotenv.Load()

	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatal("[CONFIG] DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	hasErrors := false

	log.Println("[CHECK] account_rules references existing accounts...")
	missingRefs, err := queryIssues(ctx, pool, `
		SELECT c.company_code, ar.rule_type, ar.account_code, NULL::text AS account_name, NULL::text AS account_type
		FROM account_rules ar
		JOIN companies c ON c.id = ar.company_id
		LEFT JOIN accounts a ON a.company_id = ar.company_id AND a.code = ar.account_code
		WHERE a.id IS NULL
		ORDER BY c.company_code, ar.rule_type`)
	if err != nil {
		log.Fatalf("query missing account_rules refs: %v", err)
	}
	if len(missingRefs) > 0 {
		hasErrors = true
		log.Printf("[ERROR] found %d account_rules rows referencing missing accounts", len(missingRefs))
		for _, it := range missingRefs {
			log.Printf("  company=%s rule=%s account_code=%s", it.CompanyCode, it.RuleType, it.AccountCode)
		}
	} else {
		log.Println("[OK] all account_rules references are valid.")
	}

	log.Println("[CHECK] core rule types map to expected account types...")
	typeMismatch, err := queryIssues(ctx, pool, `
		WITH typed AS (
			SELECT c.company_code, ar.rule_type, ar.account_code, a.name AS account_name, a.type AS account_type
			FROM account_rules ar
			JOIN companies c ON c.id = ar.company_id
			LEFT JOIN accounts a ON a.company_id = ar.company_id AND a.code = ar.account_code
		)
		SELECT company_code, rule_type, account_code, account_name, account_type
		FROM typed
		WHERE (rule_type = 'AR' AND account_type IS DISTINCT FROM 'asset')
		   OR (rule_type = 'AP' AND account_type IS DISTINCT FROM 'liability')
		   OR (rule_type = 'INVENTORY' AND account_type IS DISTINCT FROM 'asset')
		   OR (rule_type = 'COGS' AND account_type IS DISTINCT FROM 'expense')
		   OR (rule_type = 'BANK_DEFAULT' AND account_type IS DISTINCT FROM 'asset')
		   OR (rule_type = 'RECEIPT_CREDIT' AND account_type IS DISTINCT FROM 'liability')
		ORDER BY company_code, rule_type`)
	if err != nil {
		log.Fatalf("query rule/account type mismatch: %v", err)
	}
	if len(typeMismatch) > 0 {
		hasErrors = true
		log.Printf("[ERROR] found %d rule/account type mismatches", len(typeMismatch))
		for _, it := range typeMismatch {
			log.Printf("  company=%s rule=%s account=%s (%s, %s)",
				it.CompanyCode, it.RuleType, it.AccountCode, val(it.AccountName), val(it.AccountType))
		}
	} else {
		log.Println("[OK] core rules map to expected account types.")
	}

	log.Println("[CHECK] required core rules exist per company (AR/AP/INVENTORY/COGS/BANK_DEFAULT/RECEIPT_CREDIT)...")
	missingRules, err := queryIssues(ctx, pool, `
		WITH required(rule_type) AS (
			VALUES ('AR'), ('AP'), ('INVENTORY'), ('COGS'), ('BANK_DEFAULT'), ('RECEIPT_CREDIT')
		)
		SELECT c.company_code, req.rule_type, ''::text AS account_code, NULL::text AS account_name, NULL::text AS account_type
		FROM companies c
		CROSS JOIN required req
		LEFT JOIN account_rules ar
		  ON ar.company_id = c.id AND ar.rule_type = req.rule_type
		WHERE ar.id IS NULL
		ORDER BY c.company_code, req.rule_type`)
	if err != nil {
		log.Fatalf("query missing required rules: %v", err)
	}
	if len(missingRules) > 0 {
		hasErrors = true
		log.Printf("[ERROR] found %d missing required rule mappings", len(missingRules))
		for _, it := range missingRules {
			log.Printf("  company=%s missing_rule=%s", it.CompanyCode, it.RuleType)
		}
	} else {
		log.Println("[OK] all companies have required core rules.")
	}

	log.Println("[CHECK] account_rules account_id/account_code drift (warning only)...")
	idCodeMismatch, err := queryIDCodeMismatches(ctx, pool)
	if err != nil {
		log.Fatalf("query account_id/account_code drift: %v", err)
	}
	if len(idCodeMismatch) > 0 {
		log.Printf("[WARN] found %d account_rules rows where account_id points to a different code than account_code", len(idCodeMismatch))
		for _, it := range idCodeMismatch {
			log.Printf("  company=%s rule=%s account_code=%s account_id_code=%s",
				it.CompanyCode, it.RuleType, it.AccountCode, it.IDCode)
		}
	} else {
		log.Println("[OK] no account_rules account_id/account_code drift detected.")
	}

	log.Println("[CHECK] default company 1000 seeded CoA labels (warning only)...")
	seedNameDrift, err := queryIssues(ctx, pool, `
		WITH expected(code, expected_name, expected_type) AS (
			VALUES
			  ('1000','Cash','asset'),
			  ('1100','Bank Account','asset'),
			  ('1200','Accounts Receivable','asset'),
			  ('1300','Furniture & Fixtures','asset'),
			  ('1400','Inventory','asset'),
			  ('2000','Accounts Payable','liability'),
			  ('2100','Short-Term Loans','liability'),
			  ('3000','Owner Capital','equity'),
			  ('3100','Retained Earnings','equity'),
			  ('4000','Sales Revenue','revenue'),
			  ('4100','Service Revenue','revenue'),
			  ('5000','Cost of Goods Sold','expense'),
			  ('5100','Rent Expense','expense'),
			  ('5200','Salary Expense','expense'),
			  ('5300','Utilities Expense','expense')
		)
		SELECT c.company_code, e.code, e.code, a.name, a.type
		FROM companies c
		JOIN expected e ON true
		LEFT JOIN accounts a ON a.company_id = c.id AND a.code = e.code
		WHERE c.company_code = '1000'
		  AND (a.id IS NULL OR a.name IS DISTINCT FROM e.expected_name OR a.type IS DISTINCT FROM e.expected_type)
		ORDER BY e.code`)
	if err != nil {
		log.Fatalf("query seeded coa drift: %v", err)
	}
	if len(seedNameDrift) > 0 {
		log.Printf("[WARN] company 1000 has %d seeded CoA label/type differences", len(seedNameDrift))
		for _, it := range seedNameDrift {
			log.Printf("  company=%s account_code=%s actual_name=%s actual_type=%s",
				it.CompanyCode, it.RuleType, val(it.AccountName), val(it.AccountType))
		}
	} else {
		log.Println("[OK] company 1000 CoA labels match seeded defaults.")
	}

	log.Println("[CHECK] go-live document types enforce global numbering (blocking)...")
	var docTypePresenceCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_types
		WHERE code IN ('JE', 'SI', 'PI', 'SO', 'PO', 'GR', 'GI', 'RC', 'PV')
	`).Scan(&docTypePresenceCount); err != nil {
		log.Fatalf("query go-live document types presence: %v", err)
	}
	if docTypePresenceCount != 9 {
		hasErrors = true
		log.Printf("[ERROR] expected 9 go-live document types (JE, SI, PI, SO, PO, GR, GI, RC, PV), found %d", docTypePresenceCount)
	} else {
		log.Println("[OK] all go-live document types are present.")
	}

	var numberingDriftCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_types
		WHERE code IN ('JE', 'SI', 'PI', 'SO', 'PO', 'GR', 'GI', 'RC', 'PV')
		  AND (numbering_strategy <> 'global' OR resets_every_fy IS TRUE)
	`).Scan(&numberingDriftCount); err != nil {
		log.Fatalf("query go-live document numbering policy: %v", err)
	}
	if numberingDriftCount > 0 {
		hasErrors = true
		log.Printf("[ERROR] found %d go-live document types not configured as global + no FY reset", numberingDriftCount)
	} else {
		log.Println("[OK] go-live document types are global + no FY reset.")
	}

	log.Println("[CHECK] document type policy seed rows (blocking)...")
	var policySeedCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_type_policies
		WHERE is_active = true
		  AND (intent_code, allowed_document_type) IN (
		      ('manual_adjustment', 'JE'),
		      ('sales_invoice', 'SI'),
		      ('purchase_invoice', 'PI'),
		      ('goods_receipt', 'GR'),
		      ('goods_issue', 'GI'),
		      ('customer_receipt', 'RC'),
		      ('vendor_payment', 'PV')
		  )
	`).Scan(&policySeedCount); err != nil {
		log.Fatalf("query document type policy seed rows: %v", err)
	}
	if policySeedCount < 7 {
		hasErrors = true
		log.Printf("[ERROR] expected at least 7 active document type policy seed rows, found %d", policySeedCount)
	} else {
		log.Println("[OK] document type policy seed rows are present.")
	}

	log.Println("[CHECK] document type policy violation audit table exists (blocking)...")
	var hasPolicyViolationAuditTable bool
	if err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public'
			  AND table_name = 'document_type_policy_violation_audits'
		)
	`).Scan(&hasPolicyViolationAuditTable); err != nil {
		log.Fatalf("query document_type_policy_violation_audits table presence: %v", err)
	}
	if !hasPolicyViolationAuditTable {
		hasErrors = true
		log.Printf("[ERROR] missing required audit table document_type_policy_violation_audits")
	} else {
		log.Println("[OK] document type policy violation audit table is present.")
	}

	log.Println("[CHECK] document sequence uniqueness index shape (blocking)...")
	if ok, indexDef, err := checkIndexContains(ctx, pool, "document_sequences_unique_idx", []string{
		"on public.document_sequences",
		"(company_id, type_code)",
	}); err != nil {
		log.Fatalf("query document_sequences_unique_idx definition: %v", err)
	} else if !ok {
		hasErrors = true
		log.Printf("[ERROR] document_sequences_unique_idx is not global-scope. indexdef=%s", indexDef)
	} else {
		log.Println("[OK] document_sequences_unique_idx is global-scope.")
	}

	log.Println("[CHECK] document number uniqueness index shape (blocking)...")
	if ok, indexDef, err := checkIndexContains(ctx, pool, "documents_unique_number_idx", []string{
		"on public.documents",
		"(company_id, type_code, document_number)",
		"where (document_number is not null)",
	}); err != nil {
		log.Fatalf("query documents_unique_number_idx definition: %v", err)
	} else if !ok {
		hasErrors = true
		log.Printf("[ERROR] documents_unique_number_idx is not global uniqueness with non-null filter. indexdef=%s", indexDef)
	} else {
		log.Println("[OK] documents_unique_number_idx enforces global uniqueness on posted numbers.")
	}

	if hasErrors {
		log.Fatal("[FAIL] database health checks found blocking issues.")
	}
	log.Println("[PASS] database health checks passed.")
}

func queryIssues(ctx context.Context, pool *pgxpool.Pool, sql string) ([]issue, error) {
	rows, err := pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []issue
	for rows.Next() {
		var it issue
		if err := rows.Scan(&it.CompanyCode, &it.RuleType, &it.AccountCode, &it.AccountName, &it.AccountType); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func queryIDCodeMismatches(ctx context.Context, pool *pgxpool.Pool) ([]accountRuleIDCodeMismatch, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.company_code, ar.rule_type, ar.account_code, a.code
		FROM account_rules ar
		JOIN companies c ON c.id = ar.company_id
		JOIN accounts a
		  ON a.id = ar.account_id
		 AND a.company_id = ar.company_id
		WHERE ar.account_id IS NOT NULL
		  AND ar.account_code IS NOT NULL
		  AND ar.account_code <> a.code
		ORDER BY c.company_code, ar.rule_type, ar.id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []accountRuleIDCodeMismatch
	for rows.Next() {
		var it accountRuleIDCodeMismatch
		if err := rows.Scan(&it.CompanyCode, &it.RuleType, &it.AccountCode, &it.IDCode); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func val(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

func checkIndexContains(ctx context.Context, pool *pgxpool.Pool, indexName string, requiredSubstrings []string) (bool, string, error) {
	var indexDef string
	if err := pool.QueryRow(ctx, `
		SELECT pg_get_indexdef(c.oid)
		FROM pg_class c
		WHERE c.relkind = 'i'
		  AND c.relname = $1
	`, indexName).Scan(&indexDef); err != nil {
		return false, "", err
	}
	indexDefLower := strings.ToLower(indexDef)
	for _, token := range requiredSubstrings {
		if !strings.Contains(indexDefLower, strings.ToLower(token)) {
			return false, indexDef, nil
		}
	}
	return true, indexDef, nil
}
