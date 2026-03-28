package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"workflow_app/internal/ai"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/reporting"
	"workflow_app/internal/workflow"
)

var (
	ErrProcessedProposalNotFound        = errors.New("processed proposal not found")
	ErrProcessedProposalApprovalExists  = errors.New("processed proposal already linked to approval")
	ErrProcessedProposalDocumentMissing = errors.New("processed proposal document is required")
)

type requestProcessedProposalApprovalInput struct {
	RecommendationID string
	QueueCode        string
	Reason           string
	Actor            identityaccess.Actor
}

type processedProposalApprovalRequester interface {
	RequestProcessedProposalApproval(ctx context.Context, input requestProcessedProposalApprovalInput) (workflow.Approval, reporting.ProcessedProposalReview, error)
}

type processedProposalApprovalService struct {
	db        *sql.DB
	review    *reporting.Service
	workflow  *workflow.Service
	aiService *ai.Service
}

func newProcessedProposalApprovalService(db *sql.DB) *processedProposalApprovalService {
	documentService := documents.NewService(db)
	return &processedProposalApprovalService{
		db:        db,
		review:    reporting.NewService(db),
		workflow:  workflow.NewService(db, documentService),
		aiService: ai.NewService(db),
	}
}

func (s *processedProposalApprovalService) RequestProcessedProposalApproval(ctx context.Context, input requestProcessedProposalApprovalInput) (workflow.Approval, reporting.ProcessedProposalReview, error) {
	proposals, err := s.review.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		RecommendationID: strings.TrimSpace(input.RecommendationID),
		Limit:            2,
		Actor:            input.Actor,
	})
	if err != nil {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, err
	}
	if len(proposals) == 0 {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, ErrProcessedProposalNotFound
	}

	proposal := proposals[0]
	if proposal.ApprovalID.Valid {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, ErrProcessedProposalApprovalExists
	}
	if !proposal.DocumentID.Valid {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, ErrProcessedProposalDocumentMissing
	}

	queueCode := strings.TrimSpace(input.QueueCode)
	if queueCode == "" && proposal.SuggestedQueueCode.Valid {
		queueCode = proposal.SuggestedQueueCode.String
	}
	if queueCode == "" {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, workflow.ErrApprovalQueueRequired
	}

	reason := strings.TrimSpace(input.Reason)
	if reason == "" {
		reason = fmt.Sprintf("request approval for processed proposal %s", proposal.RecommendationID)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, fmt.Errorf("begin request proposal approval: %w", err)
	}

	approval, err := s.workflow.RequestApprovalTx(ctx, tx, workflow.RequestApprovalInput{
		DocumentID: proposal.DocumentID.String,
		QueueCode:  queueCode,
		Reason:     reason,
		Actor:      input.Actor,
	})
	if err != nil {
		_ = tx.Rollback()
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, err
	}

	if _, err := s.aiService.LinkRecommendationApprovalTx(ctx, tx, ai.LinkRecommendationApprovalInput{
		RecommendationID: proposal.RecommendationID,
		ApprovalID:       approval.ID,
		Actor:            input.Actor,
	}); err != nil {
		_ = tx.Rollback()
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, err
	}

	if err := tx.Commit(); err != nil {
		return workflow.Approval{}, reporting.ProcessedProposalReview{}, fmt.Errorf("commit request proposal approval: %w", err)
	}

	proposal.ApprovalID = sql.NullString{String: approval.ID, Valid: true}
	proposal.ApprovalStatus = sql.NullString{String: approval.Status, Valid: true}
	proposal.ApprovalQueueCode = sql.NullString{String: approval.QueueCode, Valid: true}
	return approval, proposal, nil
}
