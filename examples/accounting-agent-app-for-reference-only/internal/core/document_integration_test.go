package core_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"accounting-agent/internal/core"
)

func TestDocumentService_ConcurrentPosting(t *testing.T) {
	pool := setupTestDB(t) // Skips if TEST_DATABASE_URL is not set
	defer pool.Close()

	// Ensure documents table has the valid type for tests since setupTestDB clears it
	_, err := pool.Exec(context.Background(), `
		INSERT INTO document_types (code, name, affects_inventory, affects_gl, affects_ar, affects_ap, numbering_strategy, resets_every_fy) 
		VALUES ('JE', 'Journal Entry', false, true, false, false, 'global', false)
		ON CONFLICT DO NOTHING;
	`)
	if err != nil {
		t.Fatalf("failed to seed document type: %v", err)
	}

	docService := core.NewDocumentService(pool)
	ctx := context.Background()

	// 1. Create 10 draft documents
	var docIDs []int
	for i := 0; i < 10; i++ {
		// Using 'JE' type which is seeded globally by migration 005
		id, err := docService.CreateDraftDocument(ctx, 1, "JE", nil, nil)
		if err != nil {
			t.Fatalf("failed to create draft document: %v", err)
		}
		docIDs = append(docIDs, id)
	}

	// 2. Post all documents concurrently
	var wg sync.WaitGroup
	errCh := make(chan error, len(docIDs))

	for _, id := range docIDs {
		wg.Add(1)
		go func(docID int) {
			defer wg.Done()
			if err := docService.PostDocument(ctx, docID); err != nil {
				errCh <- err
			}
		}(id)
	}

	wg.Wait()
	close(errCh)

	// Catch any errors from the goroutines
	for err := range errCh {
		t.Errorf("concurrent post error: %v", err)
	}

	// 3. Verify exactly 10 POSTED documents and exactly 10 unique document_numbers
	var count int
	err = pool.QueryRow(ctx, "SELECT count(DISTINCT document_number) FROM documents WHERE company_id = 1 AND type_code = 'JE' AND status = 'POSTED'").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count unique document numbers: %v", err)
	}

	if count != 10 {
		t.Errorf("expected 10 unique document numbers, got %d", count)
	}
}

func TestDocumentService_NumberingStrategyEnforcement(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ctx := context.Background()

	fy2026 := 2026
	branch1 := 1
	globalDocID, err := docService.CreateDraftDocument(ctx, 1, "JE", &fy2026, &branch1)
	if err != nil {
		t.Fatalf("failed to create JE draft: %v", err)
	}
	if err := docService.PostDocument(ctx, globalDocID); err != nil {
		t.Fatalf("failed to post JE document: %v", err)
	}
	var globalFY *int
	var globalBranch *int
	if err := pool.QueryRow(ctx, "SELECT financial_year, branch_id FROM documents WHERE id = $1", globalDocID).Scan(&globalFY, &globalBranch); err != nil {
		t.Fatalf("failed to read JE document metadata: %v", err)
	}
	if globalFY != nil || globalBranch != nil {
		t.Fatalf("global strategy should clear FY/branch, got fy=%v branch=%v", globalFY, globalBranch)
	}

	// Force a drifted configuration and assert posting blocks.
	_, err = pool.Exec(ctx, `
		UPDATE document_types
		SET numbering_strategy = 'per_fy', resets_every_fy = true
		WHERE code = 'SI'
	`)
	if err != nil {
		t.Fatalf("failed to update SI numbering policy: %v", err)
	}

	driftedDocID, err := docService.CreateDraftDocument(ctx, 1, "SI", nil, nil)
	if err != nil {
		t.Fatalf("failed to create SI draft: %v", err)
	}
	err = docService.PostDocument(ctx, driftedDocID)
	if err == nil || !strings.Contains(err.Error(), "must use global numbering policy") {
		t.Fatalf("expected global policy enforcement error, got: %v", err)
	}
}

func TestDocumentService_CrossYearSequenceContinues(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	docService := core.NewDocumentService(pool)
	ctx := context.Background()

	fy2026 := 2026
	fy2027 := 2027
	docA, err := docService.CreateDraftDocument(ctx, 1, "SI", &fy2026, nil)
	if err != nil {
		t.Fatalf("failed to create SI draft for FY2026: %v", err)
	}
	docB, err := docService.CreateDraftDocument(ctx, 1, "SI", &fy2027, nil)
	if err != nil {
		t.Fatalf("failed to create SI draft for FY2027: %v", err)
	}

	if err := docService.PostDocument(ctx, docA); err != nil {
		t.Fatalf("failed to post FY2026 SI: %v", err)
	}
	if err := docService.PostDocument(ctx, docB); err != nil {
		t.Fatalf("failed to post FY2027 SI: %v", err)
	}

	var numberA, numberB string
	if err := pool.QueryRow(ctx, "SELECT document_number FROM documents WHERE id = $1", docA).Scan(&numberA); err != nil {
		t.Fatalf("failed to read first SI number: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT document_number FROM documents WHERE id = $1", docB).Scan(&numberB); err != nil {
		t.Fatalf("failed to read second SI number: %v", err)
	}

	if !strings.HasSuffix(numberA, "-00001") {
		t.Fatalf("expected first SI number to end with -00001, got %s", numberA)
	}
	if !strings.HasSuffix(numberB, "-00002") {
		t.Fatalf("expected second SI number to end with -00002, got %s", numberB)
	}
	if strings.Contains(numberA, "2026") || strings.Contains(numberB, "2027") {
		t.Fatalf("sequence should not include FY semantics in number: got %s and %s", numberA, numberB)
	}
}

func TestDocumentService_GoLiveTypesAreGlobalNoFYReset(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
		INSERT INTO document_types (code, name, affects_inventory, affects_gl, affects_ar, affects_ap, numbering_strategy, resets_every_fy)
		VALUES
			('SO', 'Sales Order', false, false, true, false, 'global', false),
			('PO', 'Purchase Order', false, false, false, false, 'global', false),
			('GR', 'Goods Receipt', true, true, false, true, 'global', false),
			('GI', 'Goods Issue', true, true, false, false, 'global', false)
		ON CONFLICT (code) DO UPDATE
		SET numbering_strategy = EXCLUDED.numbering_strategy,
		    resets_every_fy = EXCLUDED.resets_every_fy
	`)
	if err != nil {
		t.Fatalf("failed to seed go-live document types: %v", err)
	}

	var driftCount int
	if err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM document_types
		WHERE code IN ('JE', 'SI', 'PI', 'SO', 'PO', 'GR', 'GI')
		  AND (numbering_strategy <> 'global' OR resets_every_fy IS TRUE)
	`).Scan(&driftCount); err != nil {
		t.Fatalf("failed to verify numbering policy drift: %v", err)
	}
	if driftCount != 0 {
		t.Fatalf("go-live numbering policy drift detected for %d document types", driftCount)
	}
}
