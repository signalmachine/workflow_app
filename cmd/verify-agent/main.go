package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"workflow_app/internal/ai"
	"workflow_app/internal/app"
	"workflow_app/internal/attachments"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/platform/envload"
)

const verifyTimeout = 2 * time.Minute
const verifyPassword = "verify-agent-password"

func main() {
	if err := envload.LoadDefaultIfPresent(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	var (
		databaseURL     string
		channel         string
		requestText     string
		approvalFlow    bool
		attachmentText  string
		attachmentName  string
		attachmentMedia string
		submitterLabel  string
	)

	flag.StringVar(&databaseURL, "database-url", firstNonEmpty(os.Getenv("TEST_DATABASE_URL"), os.Getenv("DATABASE_URL")), "PostgreSQL connection string for verification data")
	flag.StringVar(&channel, "channel", "browser", "inbound request channel to queue and process")
	flag.StringVar(&requestText, "request-text", "Customer reports an urgent warehouse pump failure and needs operator review.", "request message text to queue for live verification")
	flag.BoolVar(&approvalFlow, "approval-flow", false, "after live provider verification, create one deterministic approval-ready proposal and verify request -> proposal -> approval -> document continuity through the shared session/API path")
	flag.StringVar(&attachmentText, "attachment-text", "Voice note transcription: the warehouse pump is failing intermittently and needs urgent inspection.", "derived attachment text to include in the verification request")
	flag.StringVar(&attachmentName, "attachment-name", "verify-agent-note.txt", "attachment file name to persist for the verification request")
	flag.StringVar(&attachmentMedia, "attachment-media-type", "text/plain", "attachment media type to persist for the verification request")
	flag.StringVar(&submitterLabel, "submitter-label", "verify-agent", "submitter label stored in inbound request metadata")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("database URL is required; set TEST_DATABASE_URL or DATABASE_URL, or pass -database-url")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), verifyTimeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("ping database: %v", err)
	}

	processor, err := app.NewOpenAIAgentProcessorFromEnv(db)
	if err != nil {
		if errors.Is(err, app.ErrAgentProviderNotConfigured) {
			log.Fatal("OpenAI provider configuration is required; set OPENAI_API_KEY and OPENAI_MODEL")
		}
		log.Fatalf("initialize agent processor: %v", err)
	}

	identity, err := createVerificationIdentity(ctx, db)
	if err != nil {
		log.Fatalf("create verification identity: %v", err)
	}

	request, err := createVerificationRequest(ctx, db, identity.Operator.Actor, verificationRequestInput{
		Channel:         channel,
		RequestText:     requestText,
		AttachmentText:  attachmentText,
		AttachmentName:  attachmentName,
		AttachmentMedia: attachmentMedia,
		SubmitterLabel:  submitterLabel,
	})
	if err != nil {
		log.Fatalf("create verification request: %v", err)
	}

	result, err := processor.ProcessNextQueuedInboundRequest(ctx, app.ProcessNextQueuedInboundRequestInput{
		Channel: channel,
		Actor:   identity.Operator.Actor,
	})
	if err != nil {
		log.Fatalf("process queued request %s: %v", request.RequestReference, err)
	}

	fmt.Printf("request_reference=%s\n", result.Request.RequestReference)
	fmt.Printf("request_status=%s\n", result.Request.Status)
	fmt.Printf("coordinator_run_id=%s\n", result.Run.ID)
	fmt.Printf("coordinator_run_status=%s\n", result.Run.Status)
	if result.Delegation.ID != "" {
		fmt.Printf("delegation_id=%s\n", result.Delegation.ID)
	}
	if result.SpecialistRun.ID != "" {
		fmt.Printf("specialist_run_id=%s\n", result.SpecialistRun.ID)
		fmt.Printf("specialist_run_status=%s\n", result.SpecialistRun.Status)
	}
	fmt.Printf("artifact_id=%s\n", result.Artifact.ID)
	fmt.Printf("recommendation_id=%s\n", result.Recommendation.ID)
	fmt.Printf("recommendation_summary=%s\n", result.Recommendation.Summary)

	if approvalFlow {
		continuity, err := verifyApprovalContinuity(ctx, db, identity, request)
		if err != nil {
			log.Fatalf("verify approval continuity for %s: %v", request.RequestReference, err)
		}
		fmt.Printf("continuity_request_reference=%s\n", request.RequestReference)
		fmt.Printf("continuity_recommendation_id=%s\n", continuity.RecommendationID)
		fmt.Printf("continuity_document_id=%s\n", continuity.DocumentID)
		fmt.Printf("continuity_approval_id=%s\n", continuity.ApprovalID)
		fmt.Printf("continuity_approval_status=%s\n", continuity.ApprovalStatus)
		fmt.Printf("continuity_document_status=%s\n", continuity.DocumentStatus)
	}
}

type verificationRequestInput struct {
	Channel         string
	RequestText     string
	AttachmentText  string
	AttachmentName  string
	AttachmentMedia string
	SubmitterLabel  string
}

type verificationIdentity struct {
	OrgID    string
	OrgSlug  string
	Operator verificationUser
	Approver verificationUser
}

type verificationUser struct {
	Actor    identityaccess.Actor
	Email    string
	Password string
}

type continuityProposal struct {
	RecommendationID string
	DocumentID       string
}

type continuityResult struct {
	RecommendationID string
	DocumentID       string
	ApprovalID       string
	ApprovalStatus   string
	DocumentStatus   string
}

func createVerificationIdentity(ctx context.Context, db *sql.DB) (verificationIdentity, error) {
	orgSlug := "verify-agent-" + time.Now().UTC().Format("20060102-150405.000000000")

	var orgID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
		orgSlug,
		"Verify Agent Org",
	).Scan(&orgID); err != nil {
		return verificationIdentity{}, fmt.Errorf("insert org: %w", err)
	}

	authService := identityaccess.NewService(db)

	operator, err := createVerificationUser(ctx, db, authService, createVerificationUserInput{
		OrgID:       orgID,
		OrgSlug:     orgSlug,
		EmailPrefix: "verify-agent-operator",
		DisplayName: "Verify Agent Operator",
		RoleCode:    identityaccess.RoleOperator,
		DeviceLabel: "verify-agent-operator",
	})
	if err != nil {
		return verificationIdentity{}, err
	}

	approver, err := createVerificationUser(ctx, db, authService, createVerificationUserInput{
		OrgID:       orgID,
		OrgSlug:     orgSlug,
		EmailPrefix: "verify-agent-approver",
		DisplayName: "Verify Agent Approver",
		RoleCode:    identityaccess.RoleApprover,
		DeviceLabel: "verify-agent-approver",
	})
	if err != nil {
		return verificationIdentity{}, err
	}

	return verificationIdentity{
		OrgID:    orgID,
		OrgSlug:  orgSlug,
		Operator: operator,
		Approver: approver,
	}, nil
}

type createVerificationUserInput struct {
	OrgID       string
	OrgSlug     string
	EmailPrefix string
	DisplayName string
	RoleCode    string
	DeviceLabel string
}

func createVerificationUser(ctx context.Context, db *sql.DB, authService *identityaccess.Service, input createVerificationUserInput) (verificationUser, error) {
	email := input.EmailPrefix + "-" + time.Now().UTC().Format("150405.000000000") + "@example.com"

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, $2) RETURNING id`,
		email,
		input.DisplayName,
	).Scan(&userID); err != nil {
		return verificationUser{}, fmt.Errorf("insert user: %w", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		input.OrgID,
		userID,
		input.RoleCode,
	); err != nil {
		return verificationUser{}, fmt.Errorf("insert membership: %w", err)
	}

	if err := authService.SetUserPassword(ctx, identityaccess.SetUserPasswordInput{
		UserID:    userID,
		Password:  verifyPassword,
		UpdatedAt: time.Now().UTC(),
	}); err != nil {
		return verificationUser{}, fmt.Errorf("set verification password: %w", err)
	}

	session, err := authService.StartBrowserSession(ctx, identityaccess.StartBrowserSessionInput{
		OrgSlug:     input.OrgSlug,
		Email:       email,
		Password:    verifyPassword,
		DeviceLabel: input.DeviceLabel,
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return verificationUser{}, fmt.Errorf("start browser session: %w", err)
	}

	return verificationUser{
		Actor: identityaccess.Actor{
			OrgID:     input.OrgID,
			UserID:    userID,
			SessionID: session.Session.ID,
		},
		Email:    email,
		Password: verifyPassword,
	}, nil
}

func createVerificationRequest(ctx context.Context, db *sql.DB, actor identityaccess.Actor, input verificationRequestInput) (intake.InboundRequest, error) {
	intakeService := intake.NewService(db)
	attachmentService := attachments.NewService(db)

	request, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    input.Channel,
		Metadata: map[string]any{
			"submitter_label": input.SubmitterLabel,
			"source":          "cmd/verify-agent",
		},
		Actor: actor,
	})
	if err != nil {
		return intake.InboundRequest{}, fmt.Errorf("create draft: %w", err)
	}

	message, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: input.RequestText,
		Actor:       actor,
	})
	if err != nil {
		return intake.InboundRequest{}, fmt.Errorf("add request message: %w", err)
	}

	if input.AttachmentText != "" {
		attachment, err := attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
			OriginalFileName: input.AttachmentName,
			MediaType:        input.AttachmentMedia,
			Content:          []byte(input.AttachmentText),
			Actor:            actor,
		})
		if err != nil {
			return intake.InboundRequest{}, fmt.Errorf("create attachment: %w", err)
		}

		if _, err := attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
			RequestMessageID: message.ID,
			AttachmentID:     attachment.ID,
			LinkRole:         attachments.LinkRoleSource,
			Actor:            actor,
		}); err != nil {
			return intake.InboundRequest{}, fmt.Errorf("link attachment to request message: %w", err)
		}

		if _, err := attachmentService.RecordDerivedText(ctx, attachments.RecordDerivedTextInput{
			SourceAttachmentID: attachment.ID,
			RequestMessageID:   message.ID,
			DerivativeType:     attachments.DerivativeTranscription,
			ContentText:        input.AttachmentText,
			Actor:              actor,
		}); err != nil {
			return intake.InboundRequest{}, fmt.Errorf("record derived text: %w", err)
		}
	}

	request, err = intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     actor,
	})
	if err != nil {
		return intake.InboundRequest{}, fmt.Errorf("queue request: %w", err)
	}

	return request, nil
}

func verifyApprovalContinuity(ctx context.Context, db *sql.DB, identity verificationIdentity, request intake.InboundRequest) (continuityResult, error) {
	proposal, err := createApprovalContinuityProposal(ctx, db, identity.Operator.Actor, request)
	if err != nil {
		return continuityResult{}, err
	}

	handler := app.NewServedAgentAPIHandler(db)
	operatorCookies, err := issueBrowserSessionCookies(handler, identity.OrgSlug, identity.Operator.Email, identity.Operator.Password, "verify-agent-browser")
	if err != nil {
		return continuityResult{}, err
	}
	approverCookies, err := issueBrowserSessionCookies(handler, identity.OrgSlug, identity.Approver.Email, identity.Approver.Password, "verify-agent-approver-browser")
	if err != nil {
		return continuityResult{}, err
	}

	var proposalDetail struct {
		RecommendationID string  `json:"recommendation_id"`
		DocumentID       *string `json:"document_id"`
		RequestReference string  `json:"request_reference"`
	}
	if err := performJSON(handler, operatorCookies, http.MethodGet, "/api/review/processed-proposals/"+proposal.RecommendationID, nil, http.StatusOK, &proposalDetail); err != nil {
		return continuityResult{}, fmt.Errorf("load processed proposal detail: %w", err)
	}
	if proposalDetail.RecommendationID != proposal.RecommendationID || proposalDetail.DocumentID == nil || *proposalDetail.DocumentID != proposal.DocumentID {
		return continuityResult{}, fmt.Errorf("processed proposal detail did not expose expected recommendation/document continuity")
	}

	var approvalResponse struct {
		RecommendationID     string  `json:"recommendation_id"`
		RecommendationStatus string  `json:"recommendation_status"`
		ApprovalID           string  `json:"approval_id"`
		ApprovalStatus       string  `json:"approval_status"`
		DocumentID           string  `json:"document_id"`
		DocumentStatus       *string `json:"document_status"`
	}
	if err := performJSON(
		handler,
		operatorCookies,
		http.MethodPost,
		"/api/review/processed-proposals/"+proposal.RecommendationID+"/request-approval",
		map[string]string{"reason": "verify-agent approval continuity"},
		http.StatusCreated,
		&approvalResponse,
	); err != nil {
		return continuityResult{}, fmt.Errorf("request proposal approval: %w", err)
	}
	if approvalResponse.RecommendationID != proposal.RecommendationID || approvalResponse.ApprovalID == "" || approvalResponse.DocumentID != proposal.DocumentID {
		return continuityResult{}, fmt.Errorf("request approval response did not expose expected ids")
	}

	var decisionResponse struct {
		ApprovalID     string `json:"approval_id"`
		Status         string `json:"status"`
		DocumentID     string `json:"document_id"`
		DocumentStatus string `json:"document_status"`
	}
	if err := performJSON(
		handler,
		approverCookies,
		http.MethodPost,
		"/api/approvals/"+approvalResponse.ApprovalID+"/decision",
		map[string]string{"decision": "approved", "decision_note": "Approved from cmd/verify-agent continuity flow."},
		http.StatusOK,
		&decisionResponse,
	); err != nil {
		return continuityResult{}, fmt.Errorf("decide approval: %w", err)
	}
	if decisionResponse.ApprovalID != approvalResponse.ApprovalID || decisionResponse.DocumentID != proposal.DocumentID {
		return continuityResult{}, fmt.Errorf("approval decision response did not expose expected ids")
	}

	var requestDetail struct {
		Request struct {
			RequestReference string `json:"request_reference"`
		} `json:"request"`
		Proposals []struct {
			RecommendationID string  `json:"recommendation_id"`
			ApprovalID       *string `json:"approval_id"`
			DocumentID       *string `json:"document_id"`
		} `json:"proposals"`
	}
	if err := performJSON(handler, operatorCookies, http.MethodGet, "/api/review/inbound-requests/"+request.RequestReference, nil, http.StatusOK, &requestDetail); err != nil {
		return continuityResult{}, fmt.Errorf("load inbound request detail: %w", err)
	}
	if requestDetail.Request.RequestReference != request.RequestReference {
		return continuityResult{}, fmt.Errorf("request detail did not expose the expected request reference")
	}
	if !proposalPresent(requestDetail.Proposals, proposal.RecommendationID, approvalResponse.ApprovalID, proposal.DocumentID) {
		return continuityResult{}, fmt.Errorf("request detail did not expose the expected proposal -> approval -> document continuity")
	}

	var approvalDetail struct {
		ApprovalID     string `json:"approval_id"`
		ApprovalStatus string `json:"approval_status"`
		DocumentID     string `json:"document_id"`
	}
	if err := performJSON(handler, approverCookies, http.MethodGet, "/api/review/approval-queue/"+approvalResponse.ApprovalID, nil, http.StatusOK, &approvalDetail); err != nil {
		return continuityResult{}, fmt.Errorf("load approval detail: %w", err)
	}
	if approvalDetail.ApprovalID != approvalResponse.ApprovalID || approvalDetail.DocumentID != proposal.DocumentID {
		return continuityResult{}, fmt.Errorf("approval detail did not expose the expected continuity")
	}

	var documentDetail struct {
		DocumentID     string  `json:"document_id"`
		Status         string  `json:"status"`
		ApprovalID     *string `json:"approval_id"`
		RequestRef     *string `json:"request_reference"`
		Recommendation *string `json:"recommendation_id"`
	}
	if err := performJSON(handler, approverCookies, http.MethodGet, "/api/review/documents/"+proposal.DocumentID, nil, http.StatusOK, &documentDetail); err != nil {
		return continuityResult{}, fmt.Errorf("load document detail: %w", err)
	}
	if documentDetail.DocumentID != proposal.DocumentID || documentDetail.ApprovalID == nil || *documentDetail.ApprovalID != approvalResponse.ApprovalID {
		return continuityResult{}, fmt.Errorf("document detail did not expose the expected approval continuity")
	}
	if documentDetail.RequestRef == nil || *documentDetail.RequestRef != request.RequestReference || documentDetail.Recommendation == nil || *documentDetail.Recommendation != proposal.RecommendationID {
		return continuityResult{}, fmt.Errorf("document detail did not expose the expected upstream request/proposal continuity")
	}

	return continuityResult{
		RecommendationID: proposal.RecommendationID,
		DocumentID:       proposal.DocumentID,
		ApprovalID:       approvalResponse.ApprovalID,
		ApprovalStatus:   approvalDetail.ApprovalStatus,
		DocumentStatus:   decisionResponse.DocumentStatus,
	}, nil
}

func createApprovalContinuityProposal(ctx context.Context, db *sql.DB, actor identityaccess.Actor, request intake.InboundRequest) (continuityProposal, error) {
	documentService := documents.NewService(db)
	aiService := ai.NewService(db)

	doc, err := documentService.CreateDraft(ctx, documents.CreateDraftInput{
		TypeCode: "invoice",
		Title:    "Verify agent continuity document",
		Actor:    actor,
	})
	if err != nil {
		return continuityProposal{}, fmt.Errorf("create continuity document: %w", err)
	}
	doc, err = documentService.Submit(ctx, documents.SubmitInput{
		DocumentID: doc.ID,
		Actor:      actor,
	})
	if err != nil {
		return continuityProposal{}, fmt.Errorf("submit continuity document: %w", err)
	}

	run, err := aiService.StartRun(ctx, ai.StartRunInput{
		InboundRequestID: request.ID,
		AgentRole:        ai.RunRoleSpecialist,
		CapabilityCode:   "workflow.approvals",
		RequestText:      "verify request approval continuity for the submitted verification document",
		Metadata: map[string]any{
			"request_reference": request.RequestReference,
			"source":            "cmd/verify-agent",
		},
		Actor: actor,
	})
	if err != nil {
		return continuityProposal{}, fmt.Errorf("start continuity run: %w", err)
	}

	recommendation, err := aiService.CreateRecommendation(ctx, ai.CreateRecommendationInput{
		RunID:              run.ID,
		RecommendationType: "request_approval",
		Summary:            "Request finance approval for the verification continuity document.",
		Payload: map[string]any{
			"document_id": doc.ID,
			"queue_code":  "finance-review",
		},
		Actor: actor,
	})
	if err != nil {
		return continuityProposal{}, fmt.Errorf("create continuity recommendation: %w", err)
	}

	return continuityProposal{
		RecommendationID: recommendation.ID,
		DocumentID:       doc.ID,
	}, nil
}

func issueBrowserSessionCookies(handler http.Handler, orgSlug, email, password, deviceLabel string) ([]*http.Cookie, error) {
	body, err := json.Marshal(map[string]string{
		"org_slug":     orgSlug,
		"email":        email,
		"password":     password,
		"device_label": deviceLabel,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal login body: %w", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/session/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusCreated {
		return nil, fmt.Errorf("login failed: status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	return recorder.Result().Cookies(), nil
}

func performJSON(handler http.Handler, cookies []*http.Cookie, method, path string, body any, wantStatus int, target any) error {
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		payload, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	if recorder.Code != wantStatus {
		return fmt.Errorf("unexpected status for %s %s: got %d body=%s", method, path, recorder.Code, recorder.Body.String())
	}
	if target == nil {
		return nil
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func proposalPresent(proposals []struct {
	RecommendationID string  `json:"recommendation_id"`
	ApprovalID       *string `json:"approval_id"`
	DocumentID       *string `json:"document_id"`
}, recommendationID, approvalID, documentID string) bool {
	for _, proposal := range proposals {
		if proposal.RecommendationID != recommendationID {
			continue
		}
		if proposal.ApprovalID == nil || *proposal.ApprovalID != approvalID {
			return false
		}
		if proposal.DocumentID == nil || *proposal.DocumentID != documentID {
			return false
		}
		return true
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
