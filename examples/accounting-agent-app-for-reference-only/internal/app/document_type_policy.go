package app

import (
	"accounting-agent/internal/core"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

// PostingIntent identifies the business intent behind a ledger proposal.
type PostingIntent string

const (
	PostingIntentUnknown          PostingIntent = "unknown"
	PostingIntentManualAdjustment PostingIntent = "manual_adjustment"
	PostingIntentSalesInvoice     PostingIntent = "sales_invoice"
	PostingIntentPurchaseInvoice  PostingIntent = "purchase_invoice"
	PostingIntentGoodsReceipt     PostingIntent = "goods_receipt"
	PostingIntentGoodsIssue       PostingIntent = "goods_issue"
	PostingIntentCustomerReceipt  PostingIntent = "customer_receipt"
	PostingIntentVendorPayment    PostingIntent = "vendor_payment"
)

func detectPostingIntent(proposal core.Proposal) PostingIntent {
	key := strings.ToLower(strings.TrimSpace(proposal.IdempotencyKey))
	switch {
	case strings.HasPrefix(key, "invoice-order-"):
		return PostingIntentSalesInvoice
	case strings.HasPrefix(key, "payment-order-"):
		return PostingIntentCustomerReceipt
	case strings.HasPrefix(key, "pay-vendor-po-"):
		return PostingIntentVendorPayment
	case strings.HasPrefix(key, "goods-receipt-"), strings.Contains(key, "-service-receipt"):
		return PostingIntentGoodsReceipt
	case strings.HasPrefix(key, "goods-issue-order-"):
		return PostingIntentGoodsIssue
	case strings.HasPrefix(key, "manual-"), strings.HasPrefix(key, "adjustment-"):
		return PostingIntentManualAdjustment
	default:
		return PostingIntentUnknown
	}
}

func expectedDocumentTypeForIntent(intent PostingIntent) (string, bool) {
	switch intent {
	case PostingIntentManualAdjustment:
		return "JE", true
	case PostingIntentSalesInvoice:
		return "SI", true
	case PostingIntentPurchaseInvoice:
		return "PI", true
	case PostingIntentGoodsReceipt:
		return "GR", true
	case PostingIntentGoodsIssue:
		return "GI", true
	case PostingIntentCustomerReceipt:
		return "RC", true
	case PostingIntentVendorPayment:
		return "PV", true
	default:
		return "", false
	}
}

func validateDocumentTypeForIntent(intent PostingIntent, docTypeCode string) error {
	expected, enforceable := expectedDocumentTypeForIntent(intent)
	if !enforceable {
		return nil
	}

	docTypeCode = strings.ToUpper(strings.TrimSpace(docTypeCode))
	if docTypeCode == expected {
		return nil
	}

	return fmt.Errorf("document type %s is not allowed for intent %s (expected %s)", docTypeCode, intent, expected)
}

func validateUnknownIntentDocumentTypePolicy(intent PostingIntent, docTypeCode string) error {
	if intent != PostingIntentUnknown {
		return nil
	}
	if strings.ToUpper(strings.TrimSpace(docTypeCode)) != "JE" {
		return nil
	}
	return fmt.Errorf("document type JE is not allowed for unknown intent; classify as manual_adjustment or use an operational document type")
}

func (s *appService) loadAllowedDocumentTypesForIntent(ctx context.Context, intent PostingIntent, source ProposalSource) (map[string]struct{}, error) {
	allowed := make(map[string]struct{})
	rows, err := s.pool.Query(ctx, `
		SELECT allowed_document_type
		FROM document_type_policies
		WHERE intent_code = $1
		  AND is_active = true
		  AND (source IS NULL OR source = $2)
	`, string(intent), string(source))
	if err != nil {
		var pgErr *pgconn.PgError
		// Backward-compatible fallback when migration 038 is not applied yet.
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			return nil, nil
		}
		return nil, fmt.Errorf("query document type policy table: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, fmt.Errorf("scan document type policy table: %w", err)
		}
		allowed[strings.ToUpper(strings.TrimSpace(code))] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate document type policy table: %w", err)
	}
	if len(allowed) == 0 {
		return nil, nil
	}
	return allowed, nil
}

func validateDocumentTypeWithTablePolicy(docTypeCode string, allowed map[string]struct{}) error {
	docTypeCode = strings.ToUpper(strings.TrimSpace(docTypeCode))
	if _, ok := allowed[docTypeCode]; ok {
		return nil
	}
	codes := make([]string, 0, len(allowed))
	for c := range allowed {
		codes = append(codes, c)
	}
	return fmt.Errorf("document type %s is not allowed by table policy (allowed: %s)", docTypeCode, strings.Join(codes, ","))
}

func documentTypePolicyModeFromEnv() string {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("DOCUMENT_TYPE_POLICY_MODE")))
	switch mode {
	case "warn", "enforce":
		return mode
	default:
		return "off"
	}
}

func (s *appService) recordDocumentTypePolicyViolation(ctx context.Context, proposal core.Proposal, source ProposalSource, intent PostingIntent, mode string, violation error, isEnforced bool) {
	var companyID int
	if err := s.pool.QueryRow(ctx, "SELECT id FROM companies WHERE company_code = $1", proposal.CompanyCode).Scan(&companyID); err != nil {
		log.Printf("[WARN] DOCUMENT_TYPE_POLICY_AUDIT resolve_company_failed company=%s err=%v", proposal.CompanyCode, err)
		return
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO document_type_policy_violation_audits
		    (company_id, source, policy_mode, intent_code, document_type_code, idempotency_key, violation_message, is_enforced)
		VALUES
		    ($1, $2, $3, $4, $5, $6, $7, $8)
	`, companyID, string(source), mode, string(intent), strings.ToUpper(strings.TrimSpace(proposal.DocumentTypeCode)), proposal.IdempotencyKey, violation.Error(), isEnforced)
	if err != nil {
		var pgErr *pgconn.PgError
		// Backward-compatible fallback when migration 041 is not applied yet.
		if errors.As(err, &pgErr) && pgErr.Code == "42P01" {
			log.Printf("[WARN] DOCUMENT_TYPE_POLICY_AUDIT table_missing company=%s source=%s mode=%s idempotency_key=%s",
				proposal.CompanyCode, source, mode, proposal.IdempotencyKey)
			return
		}
		log.Printf("[WARN] DOCUMENT_TYPE_POLICY_AUDIT insert_failed company=%s source=%s mode=%s idempotency_key=%s err=%v",
			proposal.CompanyCode, source, mode, proposal.IdempotencyKey, err)
	}
}

func (s *appService) enforceDocumentTypePolicy(ctx context.Context, proposal core.Proposal) error {
	intent := detectPostingIntent(proposal)
	source := proposalSourceFromContext(ctx)

	var validationErr error
	if allowed, err := s.loadAllowedDocumentTypesForIntent(ctx, intent, source); err != nil {
		return err
	} else if len(allowed) > 0 {
		validationErr = validateDocumentTypeWithTablePolicy(proposal.DocumentTypeCode, allowed)
	} else {
		validationErr = validateDocumentTypeForIntent(intent, proposal.DocumentTypeCode)
	}
	if validationErr == nil {
		validationErr = validateUnknownIntentDocumentTypePolicy(intent, proposal.DocumentTypeCode)
	}
	if validationErr == nil {
		return nil
	}

	mode := documentTypePolicyModeFromEnv()
	if mode == "off" {
		return nil
	}

	if mode == "warn" {
		s.recordDocumentTypePolicyViolation(ctx, proposal, source, intent, mode, validationErr, false)
		log.Printf("[WARN] DOCUMENT_TYPE_POLICY mode=warn source=%s company=%s idempotency_key=%s violation=%q",
			source, proposal.CompanyCode, proposal.IdempotencyKey, validationErr.Error())
		return nil
	}

	s.recordDocumentTypePolicyViolation(ctx, proposal, source, intent, mode, validationErr, true)
	return fmt.Errorf("DOCUMENT_TYPE_POLICY_ENFORCED: %w", validationErr)
}
