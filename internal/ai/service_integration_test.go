package ai_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"workflow_app/internal/ai"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/testsupport/dbtest"
	"workflow_app/internal/workflow"
)

func TestAIServiceCoordinatorDelegationFlowIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startSession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	documentService := documents.NewService(db)
	workflowService := workflow.NewService(db, documentService)
	aiService := ai.NewService(db)

	tool, err := aiService.RegisterTool(ctx, ai.RegisterToolInput{
		ToolName:     "workflow.request_approval",
		DisplayName:  "Request Approval",
		ModuleCode:   "workflow",
		MutatesState: true,
		Actor:        operator,
	})
	if err != nil {
		t.Fatalf("register tool: %v", err)
	}
	if tool.ToolName != "workflow.request_approval" {
		t.Fatalf("unexpected tool name: %s", tool.ToolName)
	}

	policy, err := aiService.SetToolPolicy(ctx, ai.SetToolPolicyInput{
		CapabilityCode: "workflow.coordination",
		ToolName:       tool.ToolName,
		Policy:         ai.PolicyApprovalRequired,
		Rationale:      "document submissions require shared human approval",
		Actor:          operator,
	})
	if err != nil {
		t.Fatalf("set tool policy: %v", err)
	}
	if policy.Policy != ai.PolicyApprovalRequired {
		t.Fatalf("unexpected policy: %s", policy.Policy)
	}

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Invoice proposed by coordinator",
		Actor:    operator,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("submit document: %v", err)
	}

	coordinatorRun, err := aiService.StartRun(ctx, ai.StartRunInput{
		AgentRole:      ai.RunRoleCoordinator,
		CapabilityCode: "workflow.coordination",
		RequestText:    "prepare an approval request for the submitted invoice",
		Metadata: map[string]any{
			"document_id": doc.ID,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("start coordinator run: %v", err)
	}

	coordinatorStep, err := aiService.AppendStep(ctx, ai.AppendStepInput{
		RunID:     coordinatorRun.ID,
		StepType:  "tool_plan",
		StepTitle: "Choose specialist capability",
		Status:    ai.StepStatusCompleted,
		InputPayload: map[string]any{
			"document_id": doc.ID,
		},
		OutputPayload: map[string]any{
			"capability": "workflow.approvals",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("append coordinator step: %v", err)
	}

	specialistRun, err := aiService.StartRun(ctx, ai.StartRunInput{
		AgentRole:      ai.RunRoleSpecialist,
		CapabilityCode: "workflow.approvals",
		RequestText:    "request approval for the submitted invoice",
		Metadata: map[string]any{
			"document_id": doc.ID,
		},
		ParentRunID: coordinatorRun.ID,
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("start specialist run: %v", err)
	}

	delegation, err := aiService.RecordDelegation(ctx, ai.RecordDelegationInput{
		ParentRunID:       coordinatorRun.ID,
		ChildRunID:        specialistRun.ID,
		RequestedByStepID: coordinatorStep.ID,
		CapabilityCode:    "workflow.approvals",
		Reason:            "submitted documents route through the approval specialist",
		Actor:             operator,
	})
	if err != nil {
		t.Fatalf("record delegation: %v", err)
	}
	if delegation.ChildRunID != specialistRun.ID {
		t.Fatalf("unexpected child run id: %s", delegation.ChildRunID)
	}

	specialistStep, err := aiService.AppendStep(ctx, ai.AppendStepInput{
		RunID:     specialistRun.ID,
		StepType:  "analysis",
		StepTitle: "Review approval need",
		Status:    ai.StepStatusCompleted,
		InputPayload: map[string]any{
			"document_status": doc.Status,
		},
		OutputPayload: map[string]any{
			"queue_code": "finance-review",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("append specialist step: %v", err)
	}

	artifact, err := aiService.CreateArtifact(ctx, ai.CreateArtifactInput{
		RunID:        specialistRun.ID,
		StepID:       specialistStep.ID,
		ArtifactType: "approval_brief",
		Title:        "Approval brief for submitted invoice",
		Payload: map[string]any{
			"document_id": doc.ID,
			"reason":      "invoice requires review before posting",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create artifact: %v", err)
	}

	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              specialistRun.ID,
		ArtifactID:         artifact.ID,
		RecommendationType: "request_approval",
		Summary:            "Route the invoice through the finance approval queue",
		Payload: map[string]any{
			"document_id": doc.ID,
			"queue_code":  "finance-review",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create recommendation: %v", err)
	}
	if recommendation.Status != ai.RecommendationStatusProposed {
		t.Fatalf("unexpected recommendation status: %s", recommendation.Status)
	}

	approval, err := workflowService.RequestApproval(ctx, workflow.RequestApprovalInput{
		DocumentID: doc.ID,
		QueueCode:  "finance-review",
		Reason:     "invoice requires human review",
		Actor:      operator,
	})
	if err != nil {
		t.Fatalf("request approval: %v", err)
	}

	recommendation, err = aiService.LinkRecommendationApproval(ctx, ai.LinkRecommendationApprovalInput{
		RecommendationID: recommendation.ID,
		ApprovalID:       approval.ID,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("link recommendation approval: %v", err)
	}
	if recommendation.Status != ai.RecommendationStatusApprovalRequested {
		t.Fatalf("unexpected linked recommendation status: %s", recommendation.Status)
	}

	coordinatorRun, err = aiService.CompleteRun(ctx, ai.CompleteRunInput{
		RunID:   coordinatorRun.ID,
		Status:  ai.RunStatusCompleted,
		Summary: "delegated the request to the approval specialist",
		Metadata: map[string]any{
			"delegation_id": delegation.ID,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("complete coordinator run: %v", err)
	}

	specialistRun, err = aiService.CompleteRun(ctx, ai.CompleteRunInput{
		RunID:   specialistRun.ID,
		Status:  ai.RunStatusCompleted,
		Summary: "created an approval recommendation and linked it to workflow approval",
		Metadata: map[string]any{
			"approval_id": approval.ID,
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("complete specialist run: %v", err)
	}

	var approvalID sql.NullString
	if err := db.QueryRowContext(ctx, `SELECT approval_id FROM ai.agent_recommendations WHERE id = $1`, recommendation.ID).Scan(&approvalID); err != nil {
		t.Fatalf("load recommendation approval id: %v", err)
	}
	if !approvalID.Valid || approvalID.String != approval.ID {
		t.Fatalf("unexpected approval link: %+v want %s", approvalID, approval.ID)
	}

	var delegationCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai.agent_delegations WHERE parent_run_id = $1`, coordinatorRun.ID).Scan(&delegationCount); err != nil {
		t.Fatalf("count delegations: %v", err)
	}
	if delegationCount != 1 {
		t.Fatalf("unexpected delegation count: %d", delegationCount)
	}

	var auditCount int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM platform.audit_events WHERE org_id = $1 AND event_type LIKE 'ai.%'`, orgID).Scan(&auditCount); err != nil {
		t.Fatalf("count ai audit events: %v", err)
	}
	if auditCount != 9 {
		t.Fatalf("unexpected ai audit count: got %d want 9", auditCount)
	}

	if coordinatorRun.Status != ai.RunStatusCompleted {
		t.Fatalf("unexpected coordinator status: %s", coordinatorRun.Status)
	}
	if specialistRun.Status != ai.RunStatusCompleted {
		t.Fatalf("unexpected specialist status: %s", specialistRun.Status)
	}
}

func seedOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode, existingOrgID string) (string, string) {
	t.Helper()

	orgID := existingOrgID
	if orgID == "" {
		if err := db.QueryRowContext(
			ctx,
			`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
			uniqueSlug("acme"),
			"Acme",
		).Scan(&orgID); err != nil {
			t.Fatalf("insert org: %v", err)
		}
	}

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		uniqueEmail(),
	).Scan(&userID); err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		roleCode,
	); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	return orgID, userID
}

func startSession(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) identityaccess.Session {
	t.Helper()

	service := identityaccess.NewService(db)
	session, err := service.StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "test-device",
		RefreshTokenHash: uniqueTokenHash(),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	return session
}

func uniqueSlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("150405.000000000")
}

func uniqueEmail() string {
	return "user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"
}

func uniqueTokenHash() string {
	return "token-" + time.Now().UTC().Format("150405.000000000")
}
