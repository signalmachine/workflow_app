package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/testsupport/dbtest"
)

func TestCoordinatorProcessNextQueuedWithOpenAIToolLoopIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedAIOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startAISession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	intakeService := intake.NewService(db)
	attachmentService := attachments.NewService(db)
	reportingService := reporting.NewService(db)

	request, err := intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: intake.OriginHuman,
		Channel:    "browser",
		Metadata: map[string]any{
			"submitter_label": "front desk",
		},
		Actor: operator,
	})
	if err != nil {
		t.Fatalf("create draft: %v", err)
	}

	message, err := intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: intake.MessageRoleRequest,
		TextContent: "Customer reported a failed pump and attached a voice note.",
		Actor:       operator,
	})
	if err != nil {
		t.Fatalf("add message: %v", err)
	}

	attachment, err := attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
		OriginalFileName: "voice-note.m4a",
		MediaType:        "audio/mp4",
		Content:          []byte("placeholder audio"),
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	if _, err := attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
		RequestMessageID: message.ID,
		AttachmentID:     attachment.ID,
		LinkRole:         attachments.LinkRoleSource,
		Actor:            operator,
	}); err != nil {
		t.Fatalf("link request message: %v", err)
	}

	if _, err := attachmentService.RecordDerivedText(ctx, attachments.RecordDerivedTextInput{
		SourceAttachmentID: attachment.ID,
		RequestMessageID:   message.ID,
		DerivativeType:     attachments.DerivativeTranscription,
		ContentText:        "Pump at the warehouse is failing intermittently and needs urgent inspection.",
		Actor:              operator,
	}); err != nil {
		t.Fatalf("record derived text: %v", err)
	}

	request, err = intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: request.ID,
		Actor:     operator,
	})
	if err != nil {
		t.Fatalf("queue request: %v", err)
	}

	provider := &OpenAIProvider{
		responsesAPI: &fakeOpenAIResponsesAPI{responses: []*responses.Response{
			mustResponseFromJSON(t, `{
				"id":"resp_tool_1",
				"created_at":1,
				"error":{},
				"incomplete_details":{},
				"instructions":"coordinator",
				"metadata":{},
				"model":"gpt-5.2",
				"object":"response",
				"output":[{"id":"fc_1","type":"function_call","status":"completed","call_id":"call_1","name":"reporting.list_inbound_request_status_summary","arguments":"{}"}],
				"parallel_tool_calls":false,
				"temperature":0.1,
				"tool_choice":"auto",
				"tools":[],
				"top_p":1,
				"status":"completed",
				"text":{"format":{"type":"json_schema","name":"inbound_request_review","schema":{"type":"object"}}},
				"usage":{"input_tokens":40,"input_tokens_details":{"cached_tokens":0},"output_tokens":10,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":50}
			}`),
			mustResponseFromJSON(t, `{
				"id":"resp_tool_2",
				"created_at":2,
				"error":{},
				"incomplete_details":{},
				"instructions":"coordinator",
				"metadata":{},
				"model":"gpt-5.2",
				"object":"response",
				"output":[{"id":"msg_1","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"{\"summary\":\"Operator review needed for an urgent equipment-failure request.\",\"priority\":\"urgent\",\"artifact_title\":\"Inbound request review brief\",\"artifact_body\":\"Customer reports a failing pump at the warehouse. Queue state confirms active demand and supports prioritizing urgent review.\",\"rationale\":[\"Equipment failure can affect active operations.\",\"Queue state shows the operator should treat the request as urgent.\"],\"next_actions\":[\"Review the request details and confirm the affected site.\",\"Create or route a work-order proposal after operator confirmation.\"]}","annotations":[]}]}],
				"parallel_tool_calls":false,
				"temperature":0.1,
				"tool_choice":"auto",
				"tools":[],
				"top_p":1,
				"previous_response_id":"resp_tool_1",
				"status":"completed",
				"text":{"format":{"type":"json_schema","name":"inbound_request_review","schema":{"type":"object"}}},
				"usage":{"input_tokens":30,"input_tokens_details":{"cached_tokens":0},"output_tokens":20,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":50}
			}`),
		}},
		aiService:         NewService(db),
		reportingService:  reporting.NewService(db),
		model:             "gpt-5.2",
		maxToolIterations: 3,
	}

	coordinator := NewCoordinator(db, provider)
	result, err := coordinator.ProcessNextQueued(ctx, ProcessNextQueuedInput{
		Channel: "browser",
		Actor:   operator,
	})
	if err != nil {
		t.Fatalf("process next queued request: %v", err)
	}

	if result.Request.Status != intake.StatusProcessed {
		t.Fatalf("unexpected request status: %s", result.Request.Status)
	}
	if result.Run.Status != RunStatusCompleted {
		t.Fatalf("unexpected run status: %s", result.Run.Status)
	}

	detail, err := reportingService.GetInboundRequestDetail(ctx, reporting.GetInboundRequestDetailInput{
		RequestReference: request.RequestReference,
		Actor:            operator,
	})
	if err != nil {
		t.Fatalf("get inbound request detail: %v", err)
	}
	if len(detail.Steps) != 1 {
		t.Fatalf("unexpected step count: %d", len(detail.Steps))
	}

	var stepPayload map[string]any
	if err := json.Unmarshal(detail.Steps[0].OutputPayload, &stepPayload); err != nil {
		t.Fatalf("unmarshal step payload: %v", err)
	}
	if got := int(stepPayload["tool_loop_iterations"].(float64)); got != 2 {
		t.Fatalf("unexpected tool loop iterations: %d", got)
	}
	toolExecutions, ok := stepPayload["tool_executions"].([]any)
	if !ok || len(toolExecutions) != 1 {
		t.Fatalf("unexpected tool executions payload: %+v", stepPayload["tool_executions"])
	}
	firstTool, ok := toolExecutions[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected tool execution shape: %+v", toolExecutions[0])
	}
	if firstTool["tool_name"] != openAICoordinatorSummaryToolName {
		t.Fatalf("unexpected tool name: %+v", firstTool)
	}
	if firstTool["outcome"] != "executed" {
		t.Fatalf("unexpected tool outcome: %+v", firstTool)
	}
	if firstTool["policy"] != PolicyAllow {
		t.Fatalf("unexpected tool policy: %+v", firstTool)
	}

	var artifactPayload map[string]any
	if err := json.Unmarshal(detail.Artifacts[0].Payload, &artifactPayload); err != nil {
		t.Fatalf("unmarshal artifact payload: %v", err)
	}
	if got := int(artifactPayload["tool_loop_iterations"].(float64)); got != 2 {
		t.Fatalf("unexpected artifact tool loop iterations: %d", got)
	}
}

func TestOpenAIProviderBlocksDeniedToolPolicyIntegration(t *testing.T) {
	db := dbtest.Open(t)
	defer db.Close()
	dbtest.Reset(t, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	orgID, operatorUserID := seedAIOrgAndUser(t, ctx, db, identityaccess.RoleOperator, "")
	session := startAISession(t, ctx, db, orgID, operatorUserID)
	operator := identityaccess.Actor{OrgID: orgID, UserID: operatorUserID, SessionID: session.ID}

	aiService := NewService(db)
	if _, err := aiService.RegisterTool(ctx, RegisterToolInput{
		ToolName:     openAICoordinatorSummaryToolName,
		DisplayName:  "Inbound Request Status Summary",
		ModuleCode:   "reporting",
		MutatesState: false,
		Actor:        operator,
	}); err != nil {
		t.Fatalf("register tool: %v", err)
	}
	if _, err := aiService.SetToolPolicy(ctx, SetToolPolicyInput{
		CapabilityCode: DefaultCoordinatorCapabilityCode,
		ToolName:       openAICoordinatorSummaryToolName,
		Policy:         PolicyDeny,
		Rationale:      "test policy denies automatic queue-summary reads",
		Actor:          operator,
	}); err != nil {
		t.Fatalf("set tool policy: %v", err)
	}

	provider := &OpenAIProvider{
		responsesAPI: &fakeOpenAIResponsesAPI{responses: []*responses.Response{
			mustResponseFromJSON(t, `{
				"id":"resp_deny_1",
				"created_at":1,
				"error":{},
				"incomplete_details":{},
				"instructions":"coordinator",
				"metadata":{},
				"model":"gpt-5.2",
				"object":"response",
				"output":[{"id":"fc_1","type":"function_call","status":"completed","call_id":"call_1","name":"reporting.list_inbound_request_status_summary","arguments":"{}"}],
				"parallel_tool_calls":false,
				"temperature":0.1,
				"tool_choice":"auto",
				"tools":[],
				"top_p":1,
				"status":"completed",
				"text":{"format":{"type":"json_schema","name":"inbound_request_review","schema":{"type":"object"}}},
				"usage":{"input_tokens":20,"input_tokens_details":{"cached_tokens":0},"output_tokens":10,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":30}
			}`),
			mustResponseFromJSON(t, `{
				"id":"resp_deny_2",
				"created_at":2,
				"error":{},
				"incomplete_details":{},
				"instructions":"coordinator",
				"metadata":{},
				"model":"gpt-5.2",
				"object":"response",
				"output":[{"id":"msg_1","type":"message","status":"completed","role":"assistant","content":[{"type":"output_text","text":"{\"summary\":\"Operator review is still required without the denied queue-summary tool.\",\"priority\":\"high\",\"artifact_title\":\"Inbound request review brief\",\"artifact_body\":\"The queue-summary read tool was blocked by policy, so the recommendation is based only on the persisted request context.\",\"rationale\":[\"The request still describes an equipment issue requiring human review.\"],\"next_actions\":[\"Review the request details directly.\"]}","annotations":[]}]}],
				"parallel_tool_calls":false,
				"temperature":0.1,
				"tool_choice":"auto",
				"tools":[],
				"top_p":1,
				"previous_response_id":"resp_deny_1",
				"status":"completed",
				"text":{"format":{"type":"json_schema","name":"inbound_request_review","schema":{"type":"object"}}},
				"usage":{"input_tokens":15,"input_tokens_details":{"cached_tokens":0},"output_tokens":15,"output_tokens_details":{"reasoning_tokens":0},"total_tokens":30}
			}`),
		}},
		aiService:         aiService,
		reportingService:  reporting.NewService(db),
		model:             "gpt-5.2",
		maxToolIterations: 3,
	}

	output, err := provider.ExecuteInboundRequest(ctx, CoordinatorProviderInput{
		CapabilityCode:   DefaultCoordinatorCapabilityCode,
		Actor:            operator,
		RequestReference: "REQ-000001",
		Channel:          "browser",
		OriginType:       intake.OriginHuman,
		Metadata:         json.RawMessage(`{"submitter_label":"front desk"}`),
		Messages: []CoordinatorMessage{
			{Role: intake.MessageRoleRequest, TextContent: "Pump failure reported from the warehouse."},
		},
	})
	if err != nil {
		t.Fatalf("execute inbound request: %v", err)
	}

	if output.ToolLoopIterations != 2 {
		t.Fatalf("unexpected tool loop iterations: %d", output.ToolLoopIterations)
	}
	if len(output.ToolExecutions) != 1 {
		t.Fatalf("unexpected tool execution count: %d", len(output.ToolExecutions))
	}
	if output.ToolExecutions[0].Outcome != "blocked_by_policy" {
		t.Fatalf("unexpected tool outcome: %+v", output.ToolExecutions[0])
	}
	if output.ToolExecutions[0].Policy != PolicyDeny {
		t.Fatalf("unexpected tool policy: %+v", output.ToolExecutions[0])
	}
	if output.Priority != "high" {
		t.Fatalf("unexpected priority: %s", output.Priority)
	}
}

type fakeOpenAIResponsesAPI struct {
	responses []*responses.Response
}

func (f *fakeOpenAIResponsesAPI) New(context.Context, responses.ResponseNewParams, ...option.RequestOption) (*responses.Response, error) {
	if len(f.responses) == 0 {
		return nil, errors.New("unexpected extra response request")
	}
	resp := f.responses[0]
	f.responses = f.responses[1:]
	return resp, nil
}

func mustResponseFromJSON(t *testing.T, body string) *responses.Response {
	t.Helper()

	var resp responses.Response
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal fake response: %v", err)
	}
	return &resp
}

func seedAIOrgAndUser(t *testing.T, ctx context.Context, db *sql.DB, roleCode, existingOrgID string) (string, string) {
	t.Helper()

	orgID := existingOrgID
	if orgID == "" {
		if err := db.QueryRowContext(
			ctx,
			`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
			uniqueAISlug("acme"),
			"Acme",
		).Scan(&orgID); err != nil {
			t.Fatalf("insert org: %v", err)
		}
	}

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, 'Example User') RETURNING id`,
		uniqueAIEmail(),
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

func startAISession(t *testing.T, ctx context.Context, db *sql.DB, orgID, userID string) identityaccess.Session {
	t.Helper()

	service := identityaccess.NewService(db)
	session, err := service.StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "test-device",
		RefreshTokenHash: uniqueAITokenHash(),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("start session: %v", err)
	}

	return session
}

func uniqueAISlug(prefix string) string {
	return prefix + "-" + time.Now().UTC().Format("150405.000000000")
}

func uniqueAIEmail() string {
	return "user-" + time.Now().UTC().Format("150405.000000000") + "@example.com"
}

func uniqueAITokenHash() string {
	return "token-" + time.Now().UTC().Format("150405.000000000")
}
