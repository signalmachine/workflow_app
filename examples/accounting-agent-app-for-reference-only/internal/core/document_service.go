package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentService interface {
	CreateDraftDocument(ctx context.Context, companyID int, typeCode string, financialYear *int, branchID *int) (int, error)
	// PostDocument posts a document in its own transaction. Use for standalone calls.
	PostDocument(ctx context.Context, documentID int) error
	// PostDocumentTx posts a document using an existing transaction. Use when the caller
	// controls the transaction boundary (e.g. inside ledger.execute) to ensure
	// the document posting and the journal entry INSERT are fully atomic.
	PostDocumentTx(ctx context.Context, tx pgx.Tx, documentID int) error
}

type documentService struct {
	pool *pgxpool.Pool
}

func NewDocumentService(pool *pgxpool.Pool) DocumentService {
	return &documentService{pool: pool}
}

func (s *documentService) CreateDraftDocument(ctx context.Context, companyID int, typeCode string, financialYear *int, branchID *int) (int, error) {
	var id int
	query := `
		INSERT INTO documents (company_id, type_code, status, financial_year, branch_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	err := s.pool.QueryRow(ctx, query, companyID, typeCode, string(DocumentStatusDraft), financialYear, branchID).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create draft document: %w", err)
	}
	return id, nil
}

// PostDocument posts a document in its own standalone transaction.
func (s *documentService) PostDocument(ctx context.Context, documentID int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := postDocumentWithTx(ctx, tx, documentID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// PostDocumentTx posts a document inside the caller's existing transaction.
// The caller is responsible for committing or rolling back the transaction.
func (s *documentService) PostDocumentTx(ctx context.Context, tx pgx.Tx, documentID int) error {
	return postDocumentWithTx(ctx, tx, documentID)
}

func ensureGlobalNumberingPolicyTx(ctx context.Context, tx pgx.Tx, typeCode string) error {
	var numberingStrategy string
	var resetsEveryFY bool
	if err := tx.QueryRow(ctx, `
		SELECT numbering_strategy, resets_every_fy
		FROM document_types
		WHERE code = $1
	`, typeCode).Scan(&numberingStrategy, &resetsEveryFY); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("document type %s not found", typeCode)
		}
		return fmt.Errorf("failed to load document type %s: %w", typeCode, err)
	}

	if numberingStrategy != "global" || resetsEveryFY {
		return fmt.Errorf(
			"document type %s must use global numbering policy (strategy=global, resets_every_fy=false); got strategy=%s resets_every_fy=%t",
			typeCode, numberingStrategy, resetsEveryFY,
		)
	}
	return nil
}

// postDocumentWithTx contains the core posting logic and runs within a provided transaction.
func postDocumentWithTx(ctx context.Context, tx pgx.Tx, documentID int) error {
	var doc Document
	queryDoc := `
		SELECT company_id, type_code, status, financial_year, branch_id
		FROM documents
		WHERE id = $1
		FOR UPDATE
	`
	err := tx.QueryRow(ctx, queryDoc, documentID).Scan(
		&doc.CompanyID, &doc.TypeCode, &doc.Status, &doc.FinancialYear, &doc.BranchID,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("document not found: %d", documentID)
		}
		return fmt.Errorf("failed to read document for update: %w", err)
	}

	if doc.Status != DocumentStatusDraft {
		return fmt.Errorf("document must be in DRAFT status to be posted, current status: %s", doc.Status)
	}

	if err := ensureGlobalNumberingPolicyTx(ctx, tx, doc.TypeCode); err != nil {
		return err
	}

	// Global policy: sequence scope is always (company_id, type_code), independent of FY/branch.
	effectiveFY := (*int)(nil)
	effectiveBranch := (*int)(nil)

	// Concurrency-safe gapless sequence generation
	var lastNumber int64
	querySeq := `
		INSERT INTO document_sequences (company_id, type_code, financial_year, branch_id, last_number)
		VALUES ($1, $2, $3, $4, 1)
		ON CONFLICT (company_id, type_code)
		DO UPDATE SET last_number = document_sequences.last_number + 1
		RETURNING last_number
	`
	err = tx.QueryRow(ctx, querySeq, doc.CompanyID, doc.TypeCode, effectiveFY, effectiveBranch).Scan(&lastNumber)
	if err != nil {
		return fmt.Errorf("failed to generate gapless sequence number: %w", err)
	}

	// Format document number
	yearStr := "GLOBAL"
	if effectiveFY != nil {
		yearStr = fmt.Sprintf("%d", *effectiveFY)
	}
	branchStr := ""
	if effectiveBranch != nil {
		branchStr = fmt.Sprintf("B%d-", *effectiveBranch)
	}
	formattedNum := fmt.Sprintf("%s-%s%s-%05d", doc.TypeCode, branchStr, yearStr, lastNumber)

	updateDoc := `
		UPDATE documents
		SET status = $1, document_number = $2, posted_at = NOW(), financial_year = $3, branch_id = $4
		WHERE id = $5
	`
	_, err = tx.Exec(ctx, updateDoc, string(DocumentStatusPosted), formattedNum, effectiveFY, effectiveBranch, documentID)
	if err != nil {
		return fmt.Errorf("failed to update document status and number: %w", err)
	}

	return nil
}
