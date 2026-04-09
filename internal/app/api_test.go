package app

import (
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/reporting"
)

func TestMapProcessedProposalReviewIncludesExactAccountingEntryContinuity(t *testing.T) {
	createdAt := time.Now().UTC()
	item := reporting.ProcessedProposalReview{
		RequestID:            "request-1",
		RequestReference:     "REQ-1001",
		RequestStatus:        "processed",
		RecommendationID:     "proposal-1",
		RunID:                "run-1",
		RecommendationType:   "proposal",
		RecommendationStatus: "approved",
		Summary:              "Posted downstream accounting continuity is ready.",
		DocumentID:           sql.NullString{String: "document-1", Valid: true},
		JournalEntryID:       sql.NullString{String: "entry-1", Valid: true},
		JournalEntryNumber:   sql.NullInt64{Int64: 42, Valid: true},
		CreatedAt:            createdAt,
	}

	got := mapProcessedProposalReview(item)

	if got.JournalEntryID == nil || *got.JournalEntryID != "entry-1" {
		t.Fatalf("expected journal entry id to be mapped, got %+v", got.JournalEntryID)
	}
	if got.JournalEntryNumber == nil || *got.JournalEntryNumber != 42 {
		t.Fatalf("expected journal entry number to be mapped, got %+v", got.JournalEntryNumber)
	}
	if got.DocumentID == nil || *got.DocumentID != "document-1" {
		t.Fatalf("expected document id to stay mapped, got %+v", got.DocumentID)
	}
}
