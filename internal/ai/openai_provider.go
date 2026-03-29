package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"

	"workflow_app/internal/reporting"
)

const (
	openAIProviderTimeout             = 45 * time.Second
	openAIMaxCoordinatorToolLoops     = 3
	openAICoordinatorSummaryToolName  = "reporting_list_inbound_request_status_summary"
	openAICurrentRequestDetailTool    = "reporting_get_current_inbound_request_detail"
	openAICurrentRequestProposalsTool = "reporting_list_current_processed_proposals"
)

type openAIResponsesAPI interface {
	New(ctx context.Context, body responses.ResponseNewParams, opts ...option.RequestOption) (*responses.Response, error)
}

type coordinatorToolDefinition struct {
	ToolName     string
	Description  string
	MutatesState bool
	Parameters   map[string]any
	Execute      func(ctx context.Context, input CoordinatorProviderInput) (string, string, error)
}

type OpenAIProvider struct {
	responsesAPI      openAIResponsesAPI
	aiService         *Service
	reportingService  *reporting.Service
	model             string
	maxToolIterations int
}

type providerOutputParseResult struct {
	output         CoordinatorProviderOutput
	validationErr  error
	providerRespID string
	usage          coordinatorTokenUsage
}

func NewOpenAIProvider(db *sql.DB, config ProviderConfig) (*OpenAIProvider, error) {
	if !config.Enabled() {
		return nil, ErrInvalidProviderConfig
	}

	client := openai.NewClient(option.WithAPIKey(config.OpenAIAPIKey))
	return &OpenAIProvider{
		responsesAPI:      &client.Responses,
		aiService:         NewService(db),
		reportingService:  reporting.NewService(db),
		model:             config.OpenAIModel,
		maxToolIterations: openAIMaxCoordinatorToolLoops,
	}, nil
}

func (p *OpenAIProvider) ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error) {
	requestCtx, cancel := context.WithTimeout(ctx, openAIProviderTimeout)
	defer cancel()

	toolDefs := p.coordinatorToolDefinitions()
	params := p.newCoordinatorResponseParams(input, toolDefs)
	pendingInput := params.Input
	var latestResponseID string

	var (
		toolExecutions []CoordinatorToolExecution
		totalUsage     coordinatorTokenUsage
	)

	for iteration := 1; iteration <= p.normalizedMaxToolIterations(); iteration++ {
		params.Input = pendingInput

		resp, err := p.responsesAPI.New(requestCtx, params)
		if err != nil {
			var apiErr *openai.Error
			if errors.As(err, &apiErr) {
				return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api error (status %d): %w", apiErr.StatusCode, err)
			}
			return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api request failed: %w", err)
		}

		totalUsage.add(resp.Usage)
		if strings.TrimSpace(resp.ID) != "" {
			latestResponseID = resp.ID
		}
		if err := validateOpenAIResponse(resp); err != nil {
			return CoordinatorProviderOutput{}, err
		}

		functionCalls := extractFunctionCalls(resp)
		if len(functionCalls) == 0 {
			parseResult, err := p.parseCoordinatorResponse(input, resp, totalUsage, iteration, toolExecutions)
			if err != nil {
				return CoordinatorProviderOutput{}, err
			}
			latestResponseID = parseResult.providerRespID
			if parseResult.validationErr == nil {
				return parseResult.output, nil
			}
			repairedOutput, repairedResponseID, repairUsage, err := p.repairRequestCenteredOutput(requestCtx, input, parseResult.output)
			if err != nil {
				return CoordinatorProviderOutput{}, parseResult.validationErr
			}
			if strings.TrimSpace(repairedResponseID) != "" {
				latestResponseID = repairedResponseID
			}
			totalUsage.addUsage(repairUsage)
			repairedOutput.ProviderResponseID = latestResponseID
			repairedOutput.InputTokens = totalUsage.InputTokens
			repairedOutput.OutputTokens = totalUsage.OutputTokens
			repairedOutput.TotalTokens = totalUsage.TotalTokens
			repairedOutput.ToolLoopIterations = iteration
			repairedOutput.ToolExecutions = toolExecutions
			return repairedOutput, nil
		}

		if iteration == p.normalizedMaxToolIterations() {
			return CoordinatorProviderOutput{}, fmt.Errorf("openai coordinator tool loop exceeded %d iterations", p.normalizedMaxToolIterations())
		}

		toolOutputs := make([]responses.ResponseInputItemUnionParam, 0, len(functionCalls))
		for _, call := range functionCalls {
			output, execution := p.executeCoordinatorTool(requestCtx, input, toolDefs, iteration, call)
			toolExecutions = append(toolExecutions, execution)
			toolOutputs = append(toolOutputs, responses.ResponseInputItemUnionParam{
				OfFunctionCallOutput: &responses.ResponseInputItemFunctionCallOutputParam{
					CallID: call.CallID,
					Output: responses.ResponseInputItemFunctionCallOutputOutputUnionParam{
						OfString: openai.String(output),
					},
				},
			})
		}
		pendingInput = responses.ResponseNewParamsInputUnion{
			OfInputItemList: buildStatelessContinuationInput(resp, toolOutputs),
		}
	}

	return CoordinatorProviderOutput{}, fmt.Errorf("openai coordinator tool loop terminated without final output")
}

func (p *OpenAIProvider) newCoordinatorResponseParams(input CoordinatorProviderInput, toolDefs map[string]coordinatorToolDefinition) responses.ResponseNewParams {
	return responses.ResponseNewParams{
		Model:             p.model,
		Store:             openai.Bool(false),
		Temperature:       openai.Float(0.1),
		MaxOutputTokens:   openai.Int(900),
		MaxToolCalls:      openai.Int(1),
		ParallelToolCalls: openai.Bool(false),
		Instructions:      openai.String(strings.TrimSpace(p.coordinatorInstructions())),
		Include:           coordinatorResponseIncludes(p.model),
		Input:             responses.ResponseNewParamsInputUnion{OfString: openai.String(buildProviderPrompt(input))},
		Text:              responses.ResponseTextConfigParam{Format: coordinatorResponseFormat()},
		Tools:             coordinatorResponseTools(toolDefs),
	}
}

func coordinatorResponseIncludes(model string) []responses.ResponseIncludable {
	model = strings.ToLower(strings.TrimSpace(model))
	if strings.HasPrefix(model, "gpt-5") || strings.HasPrefix(model, "o") {
		return []responses.ResponseIncludable{responses.ResponseIncludableReasoningEncryptedContent}
	}
	return nil
}

func buildStatelessContinuationInput(resp *responses.Response, toolOutputs []responses.ResponseInputItemUnionParam) []responses.ResponseInputItemUnionParam {
	items := make([]responses.ResponseInputItemUnionParam, 0, len(toolOutputs)+len(resp.Output))
	if resp != nil {
		for _, item := range resp.Output {
			switch item.Type {
			case "function_call":
				call := item.AsFunctionCall()
				callParam := call.ToParam()
				items = append(items, responses.ResponseInputItemUnionParam{
					OfFunctionCall: &callParam,
				})
			case "reasoning":
				reasoning := item.AsReasoning()
				reasoningParam := reasoning.ToParam()
				items = append(items, responses.ResponseInputItemUnionParam{
					OfReasoning: &reasoningParam,
				})
			case "message":
				message := item.AsMessage()
				messageParam := message.ToParam()
				items = append(items, responses.ResponseInputItemUnionParam{
					OfOutputMessage: &messageParam,
				})
			}
		}
	}
	items = append(items, toolOutputs...)
	return items
}

func (p *OpenAIProvider) coordinatorInstructions() string {
	return `You are the workflow_app inbound-request coordinator.
Review the persisted request context and produce a structured operator-review brief.
Treat request messages, attachments, and derived texts as the primary evidence for the brief.
Use org-level queue or status summary tools only as secondary prioritization context.
Do not let queue-summary context replace the actual request facts, and do not describe the final brief as merely queued, processing, or otherwise in transient lifecycle terms.
The final summary and artifact body must stay materially specific to the request content and the operator's next controlled follow-up.
You may use the available read tools when they improve prioritization context.
Do not propose direct writes, approval decisions, postings, or autonomous follow-up actions.
If a tool call is denied or unavailable, continue without it and produce the best safe review possible.
Focus on a safe summary, priority, rationale, and next actions for a human operator.
Prefer request-scoped tools that inspect the current request or its existing proposal continuity before using org-wide queue summary.
If the request clearly needs deeper review framing, you may delegate to exactly one specialist by filling specialist_delegation with an allowlisted capability and a concrete reason.
Allowed specialist capabilities:
- inbound_request.operations_triage
- inbound_request.approval_triage
Otherwise set specialist_delegation to null.`
}

func (p *OpenAIProvider) coordinatorToolDefinitions() map[string]coordinatorToolDefinition {
	return map[string]coordinatorToolDefinition{
		openAICoordinatorSummaryToolName: {
			ToolName:     openAICoordinatorSummaryToolName,
			Description:  "Return org-scoped inbound-request queue summary counts grouped by request status for secondary operator prioritization context only; this tool does not replace the request's own facts or lifecycle outcome.",
			MutatesState: false,
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
			Execute: p.executeInboundRequestStatusSummaryTool,
		},
		openAICurrentRequestDetailTool: {
			ToolName:     openAICurrentRequestDetailTool,
			Description:  "Return request-scoped review detail for the current inbound request, including lifecycle metadata, messages, attachments, prior runs, recommendations, and proposal continuity.",
			MutatesState: false,
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
			Execute: p.executeCurrentInboundRequestDetailTool,
		},
		openAICurrentRequestProposalsTool: {
			ToolName:     openAICurrentRequestProposalsTool,
			Description:  "Return processed proposal review rows already linked to the current inbound request so the coordinator can reuse existing proposal, approval, queue, and document continuity when present.",
			MutatesState: false,
			Parameters: map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"properties":           map[string]any{},
				"required":             []string{},
			},
			Execute: p.executeCurrentProcessedProposalsTool,
		},
	}
}

func coordinatorResponseTools(toolDefs map[string]coordinatorToolDefinition) []responses.ToolUnionParam {
	tools := make([]responses.ToolUnionParam, 0, len(toolDefs))
	for _, tool := range toolDefs {
		tool := tool
		tools = append(tools, responses.ToolUnionParam{
			OfFunction: &responses.FunctionToolParam{
				Name:        tool.ToolName,
				Description: openai.String(tool.Description),
				Parameters:  tool.Parameters,
				Strict:      openai.Bool(true),
			},
		})
	}
	return tools
}

func (p *OpenAIProvider) executeInboundRequestStatusSummaryTool(ctx context.Context, input CoordinatorProviderInput) (string, string, error) {
	summaries, err := p.reportingService.ListInboundRequestStatusSummary(ctx, input.Actor)
	if err != nil {
		return "", "", fmt.Errorf("list inbound request status summary: %w", err)
	}

	type summaryItem struct {
		Status           string  `json:"status"`
		RequestCount     int     `json:"request_count"`
		MessageCount     int     `json:"message_count"`
		AttachmentCount  int     `json:"attachment_count"`
		LatestReceivedAt *string `json:"latest_received_at"`
		LatestQueuedAt   *string `json:"latest_queued_at"`
		LatestUpdatedAt  string  `json:"latest_updated_at"`
	}

	payload := make([]summaryItem, 0, len(summaries))
	for _, summary := range summaries {
		item := summaryItem{
			Status:          summary.Status,
			RequestCount:    summary.RequestCount,
			MessageCount:    summary.MessageCount,
			AttachmentCount: summary.AttachmentCount,
			LatestUpdatedAt: summary.LatestUpdatedAt.UTC().Format(time.RFC3339),
		}
		if summary.LatestReceivedAt.Valid {
			value := summary.LatestReceivedAt.Time.UTC().Format(time.RFC3339)
			item.LatestReceivedAt = &value
		}
		if summary.LatestQueuedAt.Valid {
			value := summary.LatestQueuedAt.Time.UTC().Format(time.RFC3339)
			item.LatestQueuedAt = &value
		}
		payload = append(payload, item)
	}

	body, err := json.Marshal(map[string]any{
		"request_reference": input.RequestReference,
		"status_summaries":  payload,
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal inbound request status summary: %w", err)
	}

	preview := fmt.Sprintf("returned %d status groups", len(payload))
	if len(payload) > 0 {
		preview = fmt.Sprintf("returned %d status groups; top status=%s", len(payload), payload[0].Status)
	}
	return string(body), preview, nil
}

func (p *OpenAIProvider) executeCurrentInboundRequestDetailTool(ctx context.Context, input CoordinatorProviderInput) (string, string, error) {
	detail, err := p.reportingService.GetInboundRequestDetail(ctx, reporting.GetInboundRequestDetailInput{
		RequestReference: input.RequestReference,
		Actor:            input.Actor,
	})
	if err != nil {
		return "", "", fmt.Errorf("get inbound request detail: %w", err)
	}

	type requestMetadata struct {
		RequestID                string         `json:"request_id"`
		RequestReference         string         `json:"request_reference"`
		Status                   string         `json:"status"`
		OriginType               string         `json:"origin_type"`
		Channel                  string         `json:"channel"`
		CancellationReason       string         `json:"cancellation_reason,omitempty"`
		FailureReason            string         `json:"failure_reason,omitempty"`
		ReceivedAt               string         `json:"received_at"`
		QueuedAt                 *string        `json:"queued_at"`
		ProcessingStartedAt      *string        `json:"processing_started_at"`
		ProcessedAt              *string        `json:"processed_at"`
		FailedAt                 *string        `json:"failed_at"`
		CancelledAt              *string        `json:"cancelled_at"`
		UpdatedAt                string         `json:"updated_at"`
		MessageCount             int            `json:"message_count"`
		AttachmentCount          int            `json:"attachment_count"`
		LastRunID                *string        `json:"last_run_id"`
		LastRunStatus            *string        `json:"last_run_status"`
		LastRecommendationID     *string        `json:"last_recommendation_id"`
		LastRecommendationStatus *string        `json:"last_recommendation_status"`
		Metadata                 map[string]any `json:"metadata"`
	}
	type messageItem struct {
		MessageID       string  `json:"message_id"`
		Role            string  `json:"role"`
		TextContent     string  `json:"text_content"`
		AttachmentCount int     `json:"attachment_count"`
		CreatedAt       string  `json:"created_at"`
		CreatedByUserID *string `json:"created_by_user_id"`
	}
	type attachmentItem struct {
		AttachmentID      string  `json:"attachment_id"`
		RequestMessageID  string  `json:"request_message_id"`
		LinkRole          string  `json:"link_role"`
		OriginalFileName  string  `json:"original_file_name"`
		MediaType         string  `json:"media_type"`
		SizeBytes         int64   `json:"size_bytes"`
		DerivedTextCount  int     `json:"derived_text_count"`
		LatestDerivedText *string `json:"latest_derived_text"`
	}
	type runItem struct {
		RunID          string  `json:"run_id"`
		AgentRole      string  `json:"agent_role"`
		CapabilityCode string  `json:"capability_code"`
		Status         string  `json:"status"`
		Summary        string  `json:"summary"`
		StartedAt      string  `json:"started_at"`
		CompletedAt    *string `json:"completed_at"`
	}
	type recommendationItem struct {
		RecommendationID   string  `json:"recommendation_id"`
		RunID              string  `json:"run_id"`
		RecommendationType string  `json:"recommendation_type"`
		Status             string  `json:"status"`
		Summary            string  `json:"summary"`
		ApprovalID         *string `json:"approval_id"`
		CreatedAt          string  `json:"created_at"`
		UpdatedAt          string  `json:"updated_at"`
	}

	metadata := map[string]any{}
	if len(detail.Request.Metadata) > 0 {
		if err := json.Unmarshal(detail.Request.Metadata, &metadata); err != nil {
			metadata = map[string]any{
				"raw_json": string(detail.Request.Metadata),
			}
		}
	}
	payload := map[string]any{
		"request": requestMetadata{
			RequestID:                detail.Request.RequestID,
			RequestReference:         detail.Request.RequestReference,
			Status:                   detail.Request.Status,
			OriginType:               detail.Request.OriginType,
			Channel:                  detail.Request.Channel,
			CancellationReason:       strings.TrimSpace(detail.Request.CancellationReason),
			FailureReason:            strings.TrimSpace(detail.Request.FailureReason),
			ReceivedAt:               detail.Request.ReceivedAt.UTC().Format(time.RFC3339),
			QueuedAt:                 formatNullTime(detail.Request.QueuedAt),
			ProcessingStartedAt:      formatNullTime(detail.Request.ProcessingStartedAt),
			ProcessedAt:              formatNullTime(detail.Request.ProcessedAt),
			FailedAt:                 formatNullTime(detail.Request.FailedAt),
			CancelledAt:              formatNullTime(detail.Request.CancelledAt),
			UpdatedAt:                detail.Request.UpdatedAt.UTC().Format(time.RFC3339),
			MessageCount:             detail.Request.MessageCount,
			AttachmentCount:          detail.Request.AttachmentCount,
			LastRunID:                formatNullString(detail.Request.LastRunID),
			LastRunStatus:            formatNullString(detail.Request.LastRunStatus),
			LastRecommendationID:     formatNullString(detail.Request.LastRecommendationID),
			LastRecommendationStatus: formatNullString(detail.Request.LastRecommendationStatus),
			Metadata:                 metadata,
		},
		"counts": map[string]any{
			"message_count":        len(detail.Messages),
			"attachment_count":     len(detail.Attachments),
			"run_count":            len(detail.Runs),
			"step_count":           len(detail.Steps),
			"delegation_count":     len(detail.Delegations),
			"artifact_count":       len(detail.Artifacts),
			"recommendation_count": len(detail.Recommendations),
			"proposal_count":       len(detail.Proposals),
		},
	}

	messages := make([]messageItem, 0, len(detail.Messages))
	for _, message := range detail.Messages {
		messages = append(messages, messageItem{
			MessageID:       message.MessageID,
			Role:            message.MessageRole,
			TextContent:     strings.TrimSpace(message.TextContent),
			AttachmentCount: message.AttachmentCount,
			CreatedAt:       message.CreatedAt.UTC().Format(time.RFC3339),
			CreatedByUserID: formatNullString(message.CreatedByUserID),
		})
	}
	payload["messages"] = messages

	attachments := make([]attachmentItem, 0, len(detail.Attachments))
	for _, attachment := range detail.Attachments {
		attachments = append(attachments, attachmentItem{
			AttachmentID:      attachment.AttachmentID,
			RequestMessageID:  attachment.RequestMessageID,
			LinkRole:          attachment.LinkRole,
			OriginalFileName:  attachment.OriginalFileName,
			MediaType:         attachment.MediaType,
			SizeBytes:         attachment.SizeBytes,
			DerivedTextCount:  attachment.DerivedTextCount,
			LatestDerivedText: formatNullString(attachment.LatestDerivedText),
		})
	}
	payload["attachments"] = attachments

	runs := make([]runItem, 0, len(detail.Runs))
	for _, run := range detail.Runs {
		runs = append(runs, runItem{
			RunID:          run.RunID,
			AgentRole:      run.AgentRole,
			CapabilityCode: run.CapabilityCode,
			Status:         run.Status,
			Summary:        strings.TrimSpace(run.Summary),
			StartedAt:      run.StartedAt.UTC().Format(time.RFC3339),
			CompletedAt:    formatNullTime(run.CompletedAt),
		})
	}
	payload["runs"] = runs

	recommendations := make([]recommendationItem, 0, len(detail.Recommendations))
	for _, recommendation := range detail.Recommendations {
		recommendations = append(recommendations, recommendationItem{
			RecommendationID:   recommendation.RecommendationID,
			RunID:              recommendation.RunID,
			RecommendationType: recommendation.RecommendationType,
			Status:             recommendation.Status,
			Summary:            strings.TrimSpace(recommendation.Summary),
			ApprovalID:         formatNullString(recommendation.ApprovalID),
			CreatedAt:          recommendation.CreatedAt.UTC().Format(time.RFC3339),
			UpdatedAt:          recommendation.UpdatedAt.UTC().Format(time.RFC3339),
		})
	}
	payload["recommendations"] = recommendations

	proposals, err := marshalCurrentProcessedProposals(detail.Proposals)
	if err != nil {
		return "", "", err
	}
	payload["processed_proposals"] = proposals

	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("marshal inbound request detail: %w", err)
	}

	preview := fmt.Sprintf(
		"returned request detail with %d messages, %d runs, and %d proposals",
		len(detail.Messages),
		len(detail.Runs),
		len(detail.Proposals),
	)
	return string(body), preview, nil
}

func (p *OpenAIProvider) executeCurrentProcessedProposalsTool(ctx context.Context, input CoordinatorProviderInput) (string, string, error) {
	proposals, err := p.reportingService.ListProcessedProposals(ctx, reporting.ListProcessedProposalsInput{
		RequestReference: input.RequestReference,
		Actor:            input.Actor,
	})
	if err != nil {
		return "", "", fmt.Errorf("list processed proposals for current request: %w", err)
	}

	payload, err := marshalCurrentProcessedProposals(proposals)
	if err != nil {
		return "", "", err
	}

	body, err := json.Marshal(map[string]any{
		"request_reference":   input.RequestReference,
		"processed_proposals": payload,
	})
	if err != nil {
		return "", "", fmt.Errorf("marshal current processed proposals: %w", err)
	}

	preview := fmt.Sprintf("returned %d processed proposals", len(proposals))
	return string(body), preview, nil
}

func marshalCurrentProcessedProposals(proposals []reporting.ProcessedProposalReview) ([]map[string]any, error) {
	payload := make([]map[string]any, 0, len(proposals))
	for _, proposal := range proposals {
		payload = append(payload, map[string]any{
			"request_id":            proposal.RequestID,
			"request_reference":     proposal.RequestReference,
			"request_status":        proposal.RequestStatus,
			"recommendation_id":     proposal.RecommendationID,
			"run_id":                proposal.RunID,
			"recommendation_type":   proposal.RecommendationType,
			"recommendation_status": proposal.RecommendationStatus,
			"summary":               strings.TrimSpace(proposal.Summary),
			"suggested_queue_code":  formatNullString(proposal.SuggestedQueueCode),
			"approval_id":           formatNullString(proposal.ApprovalID),
			"approval_status":       formatNullString(proposal.ApprovalStatus),
			"approval_queue_code":   formatNullString(proposal.ApprovalQueueCode),
			"document_id":           formatNullString(proposal.DocumentID),
			"document_type_code":    formatNullString(proposal.DocumentTypeCode),
			"document_title":        formatNullString(proposal.DocumentTitle),
			"document_number":       formatNullString(proposal.DocumentNumber),
			"document_status":       formatNullString(proposal.DocumentStatus),
			"created_at":            proposal.CreatedAt.UTC().Format(time.RFC3339),
		})
	}
	return payload, nil
}

func formatNullString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	trimmed := strings.TrimSpace(value.String)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func formatNullTime(value sql.NullTime) *string {
	if !value.Valid {
		return nil
	}
	formatted := value.Time.UTC().Format(time.RFC3339)
	return &formatted
}

func (p *OpenAIProvider) executeCoordinatorTool(ctx context.Context, input CoordinatorProviderInput, toolDefs map[string]coordinatorToolDefinition, iteration int, call responses.ResponseFunctionToolCall) (string, CoordinatorToolExecution) {
	execution := CoordinatorToolExecution{
		Iteration:     iteration,
		ToolName:      call.Name,
		CallID:        call.CallID,
		ArgumentsJSON: strings.TrimSpace(call.Arguments),
	}

	toolDef, ok := toolDefs[call.Name]
	if !ok {
		execution.Policy = PolicyDeny
		execution.Outcome = "unknown_tool"
		execution.ResultPreview = "requested tool is not registered in the coordinator runtime"
		return marshalToolOutput(map[string]any{
			"status":  "error",
			"tool":    call.Name,
			"message": "requested tool is not available",
		}), execution
	}

	defaultPolicy := PolicyAllow
	if toolDef.MutatesState {
		defaultPolicy = PolicyApprovalRequired
	}
	resolvedPolicy, err := p.aiService.ResolveToolPolicy(ctx, ResolveToolPolicyInput{
		CapabilityCode: input.CapabilityCode,
		ToolName:       toolDef.ToolName,
		DefaultPolicy:  defaultPolicy,
		Actor:          input.Actor,
	})
	if err != nil {
		execution.Policy = PolicyDeny
		execution.Outcome = "policy_lookup_failed"
		execution.ResultPreview = sanitizeToolPreview(err.Error())
		return marshalToolOutput(map[string]any{
			"status":  "error",
			"tool":    toolDef.ToolName,
			"message": "tool policy lookup failed",
			"error":   sanitizeToolPreview(err.Error()),
		}), execution
	}
	execution.Policy = resolvedPolicy.Policy

	if resolvedPolicy.Policy != PolicyAllow {
		execution.Outcome = "blocked_by_policy"
		execution.ResultPreview = fmt.Sprintf("execution blocked by %s policy", resolvedPolicy.Policy)
		return marshalToolOutput(map[string]any{
			"status":        "blocked",
			"tool":          toolDef.ToolName,
			"policy":        resolvedPolicy.Policy,
			"policy_source": resolvedPolicy.Source,
			"message":       "tool execution blocked by policy",
		}), execution
	}

	output, preview, err := toolDef.Execute(ctx, input)
	if err != nil {
		execution.Outcome = "execution_failed"
		execution.ResultPreview = sanitizeToolPreview(err.Error())
		return marshalToolOutput(map[string]any{
			"status":  "error",
			"tool":    toolDef.ToolName,
			"message": "tool execution failed",
			"error":   sanitizeToolPreview(err.Error()),
		}), execution
	}

	execution.Outcome = "executed"
	execution.ResultPreview = sanitizeToolPreview(preview)
	return output, execution
}

func validateOpenAIResponse(resp *responses.Response) error {
	if resp == nil {
		return errors.New("openai responses api returned nil response")
	}
	switch resp.Status {
	case responses.ResponseStatusCompleted:
	case responses.ResponseStatusIncomplete:
		reason := strings.TrimSpace(resp.IncompleteDetails.Reason)
		if reason == "" {
			reason = "unknown"
		}
		return fmt.Errorf("openai response incomplete: %s", reason)
	case responses.ResponseStatusFailed:
		return fmt.Errorf("openai response failed")
	default:
		return fmt.Errorf("openai response ended with status %s", resp.Status)
	}

	if hasResponseRefusal(resp) {
		return fmt.Errorf("openai response refused to complete the coordinator task")
	}

	if len(extractFunctionCalls(resp)) == 0 && strings.TrimSpace(resp.OutputText()) == "" {
		return fmt.Errorf("openai response did not return structured output text")
	}

	return nil
}

func extractFunctionCalls(resp *responses.Response) []responses.ResponseFunctionToolCall {
	if resp == nil {
		return nil
	}
	functionCalls := make([]responses.ResponseFunctionToolCall, 0)
	for _, item := range resp.Output {
		if item.Type != "function_call" {
			continue
		}
		functionCalls = append(functionCalls, item.AsFunctionCall())
	}
	return functionCalls
}

func hasResponseRefusal(resp *responses.Response) bool {
	if resp == nil {
		return false
	}
	for _, item := range resp.Output {
		for _, content := range item.Content {
			if content.Type == "refusal" && strings.TrimSpace(content.Refusal) != "" {
				return true
			}
		}
	}
	return false
}

func marshalToolOutput(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return `{"status":"error","message":"failed to marshal tool output"}`
	}
	return string(body)
}

func sanitizeToolPreview(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) > 200 {
		return value[:200]
	}
	return value
}

func (p *OpenAIProvider) normalizedMaxToolIterations() int {
	if p.maxToolIterations <= 0 {
		return openAIMaxCoordinatorToolLoops
	}
	return p.maxToolIterations
}

type coordinatorTokenUsage struct {
	InputTokens  int64
	OutputTokens int64
	TotalTokens  int64
}

func (u *coordinatorTokenUsage) add(usage responses.ResponseUsage) {
	u.InputTokens += usage.InputTokens
	u.OutputTokens += usage.OutputTokens
	u.TotalTokens += usage.TotalTokens
}

func (u *coordinatorTokenUsage) addUsage(other coordinatorTokenUsage) {
	u.InputTokens += other.InputTokens
	u.OutputTokens += other.OutputTokens
	u.TotalTokens += other.TotalTokens
}

type openAICoordinatorPayload struct {
	Summary              string                           `json:"summary"`
	Priority             string                           `json:"priority"`
	ArtifactTitle        string                           `json:"artifact_title"`
	ArtifactBody         string                           `json:"artifact_body"`
	Rationale            []string                         `json:"rationale"`
	NextActions          []string                         `json:"next_actions"`
	SpecialistDelegation *CoordinatorSpecialistDelegation `json:"specialist_delegation"`
}

func coordinatorResponseFormat() responses.ResponseFormatTextConfigUnionParam {
	return responses.ResponseFormatTextConfigUnionParam{
		OfJSONSchema: &responses.ResponseFormatTextJSONSchemaConfigParam{
			Name:        "inbound_request_review",
			Description: openai.String("Structured operator-review brief for a queued inbound request"),
			Schema:      coordinatorResponseSchema(),
			Strict:      openai.Bool(true),
		},
	}
}

func coordinatorResponseSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"properties": map[string]any{
			"summary": map[string]any{
				"type": "string",
			},
			"priority": map[string]any{
				"type": "string",
				"enum": []string{"low", "normal", "high", "urgent"},
			},
			"artifact_title": map[string]any{
				"type": "string",
			},
			"artifact_body": map[string]any{
				"type": "string",
			},
			"rationale": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"next_actions": map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
			"specialist_delegation": map[string]any{
				"anyOf": []any{
					map[string]any{
						"type":                 "object",
						"additionalProperties": false,
						"properties": map[string]any{
							"capability_code": map[string]any{
								"type": "string",
								"enum": []string{
									"inbound_request.operations_triage",
									"inbound_request.approval_triage",
								},
							},
							"reason": map[string]any{
								"type": "string",
							},
						},
						"required": []string{
							"capability_code",
							"reason",
						},
					},
					map[string]any{
						"type": "null",
					},
				},
			},
		},
		"required": []string{
			"summary",
			"priority",
			"artifact_title",
			"artifact_body",
			"rationale",
			"next_actions",
			"specialist_delegation",
		},
	}
}

func buildProviderPrompt(input CoordinatorProviderInput) string {
	var b strings.Builder

	keywords := requestEvidenceKeywords(input)

	b.WriteString("Inbound request review task\n")
	b.WriteString("Base the review on the concrete request evidence below.\n")
	b.WriteString("Treat queue-wide status context as secondary only if you decide to use the read tool.\n")
	b.WriteString("Do not describe the final review as merely queued or processing.\n\n")

	if len(keywords) > 0 {
		limit := len(keywords)
		if limit > 8 {
			limit = 8
		}
		b.WriteString("Concrete request details to preserve in the summary or artifact body:\n")
		for _, keyword := range keywords[:limit] {
			b.WriteString(fmt.Sprintf("- %s\n", keyword))
		}
		b.WriteString("\n")
	}

	b.WriteString("Inbound request context\n")
	b.WriteString(fmt.Sprintf("Request reference: %s\n", input.RequestReference))
	b.WriteString(fmt.Sprintf("Channel: %s\n", input.Channel))
	b.WriteString(fmt.Sprintf("Origin type: %s\n", input.OriginType))

	metadata := strings.TrimSpace(string(input.Metadata))
	if metadata == "" {
		metadata = "{}"
	}
	b.WriteString("Metadata:\n")
	b.WriteString(metadata)
	b.WriteString("\n\nMessages:\n")
	if len(input.Messages) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, message := range input.Messages {
			b.WriteString(fmt.Sprintf("- [%s] %s\n", message.Role, strings.TrimSpace(message.TextContent)))
		}
	}

	b.WriteString("\nAttachments:\n")
	if len(input.Attachments) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, attachment := range input.Attachments {
			b.WriteString(fmt.Sprintf(
				"- message=%s file=%s media_type=%s size_bytes=%d link_role=%s\n",
				attachment.RequestMessageID,
				attachment.OriginalFileName,
				attachment.MediaType,
				attachment.SizeBytes,
				attachment.LinkRole,
			))
		}
	}

	b.WriteString("\nDerived texts:\n")
	if len(input.DerivedTexts) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, derived := range input.DerivedTexts {
			b.WriteString(fmt.Sprintf(
				"- attachment=%s message=%s type=%s text=%s\n",
				derived.SourceAttachmentID,
				derived.RequestMessageID,
				derived.DerivativeType,
				strings.TrimSpace(derived.ContentText),
			))
		}
	}

	return b.String()
}

func (p *OpenAIProvider) parseCoordinatorResponse(input CoordinatorProviderInput, resp *responses.Response, usage coordinatorTokenUsage, iteration int, toolExecutions []CoordinatorToolExecution) (providerOutputParseResult, error) {
	var parsed openAICoordinatorPayload
	if err := json.Unmarshal([]byte(resp.OutputText()), &parsed); err != nil {
		return providerOutputParseResult{}, fmt.Errorf("decode openai coordinator output: %w", err)
	}

	output := CoordinatorProviderOutput{
		ProviderResponseID:   strings.TrimSpace(resp.ID),
		ProviderName:         "openai",
		Model:                string(resp.Model),
		Summary:              strings.TrimSpace(parsed.Summary),
		Priority:             normalizePriority(parsed.Priority),
		ArtifactTitle:        strings.TrimSpace(parsed.ArtifactTitle),
		ArtifactBody:         strings.TrimSpace(parsed.ArtifactBody),
		Rationale:            trimList(parsed.Rationale),
		NextActions:          trimList(parsed.NextActions),
		InputTokens:          usage.InputTokens,
		OutputTokens:         usage.OutputTokens,
		TotalTokens:          usage.TotalTokens,
		ToolLoopIterations:   iteration,
		ToolExecutions:       toolExecutions,
		SpecialistDelegation: normalizeSpecialistDelegation(parsed.SpecialistDelegation),
	}

	return providerOutputParseResult{
		output:         output,
		validationErr:  validateRequestCenteredCoordinatorOutput(input, output),
		providerRespID: strings.TrimSpace(resp.ID),
		usage: coordinatorTokenUsage{
			InputTokens:  resp.Usage.InputTokens,
			OutputTokens: resp.Usage.OutputTokens,
			TotalTokens:  resp.Usage.TotalTokens,
		},
	}, nil
}

func (p *OpenAIProvider) repairRequestCenteredOutput(ctx context.Context, input CoordinatorProviderInput, priorOutput CoordinatorProviderOutput) (CoordinatorProviderOutput, string, coordinatorTokenUsage, error) {
	params := responses.ResponseNewParams{
		Model:           p.model,
		Store:           openai.Bool(false),
		Temperature:     openai.Float(0),
		MaxOutputTokens: openai.Int(900),
		Instructions: openai.String(`Revise the operator-review brief so it stays materially specific to the inbound request.
Preserve the controlled workflow stance.
Mention at least one concrete detail from the request evidence in the summary or artifact body.
Do not describe the request as merely queued or processing.
Return only the corrected structured output.`),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(buildRequestCenteredRepairPrompt(input, priorOutput)),
		},
		Text: responses.ResponseTextConfigParam{Format: coordinatorResponseFormat()},
	}

	resp, err := p.responsesAPI.New(ctx, params)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) {
			return CoordinatorProviderOutput{}, "", coordinatorTokenUsage{}, fmt.Errorf("openai responses api error (status %d): %w", apiErr.StatusCode, err)
		}
		return CoordinatorProviderOutput{}, "", coordinatorTokenUsage{}, fmt.Errorf("openai responses api request failed: %w", err)
	}
	if err := validateOpenAIResponse(resp); err != nil {
		return CoordinatorProviderOutput{}, "", coordinatorTokenUsage{}, err
	}
	parseResult, err := p.parseCoordinatorResponse(input, resp, coordinatorTokenUsage{
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
		TotalTokens:  resp.Usage.TotalTokens,
	}, priorOutput.ToolLoopIterations, priorOutput.ToolExecutions)
	if err != nil {
		return CoordinatorProviderOutput{}, "", coordinatorTokenUsage{}, err
	}
	if parseResult.validationErr != nil {
		return CoordinatorProviderOutput{}, "", coordinatorTokenUsage{}, parseResult.validationErr
	}
	return parseResult.output, parseResult.providerRespID, parseResult.usage, nil
}

func buildRequestCenteredRepairPrompt(input CoordinatorProviderInput, priorOutput CoordinatorProviderOutput) string {
	var b strings.Builder
	b.WriteString("Original request evidence:\n")
	b.WriteString(buildProviderPrompt(input))
	b.WriteString("\nOriginal structured brief to revise:\n")
	body, err := json.Marshal(map[string]any{
		"summary":               priorOutput.Summary,
		"priority":              priorOutput.Priority,
		"artifact_title":        priorOutput.ArtifactTitle,
		"artifact_body":         priorOutput.ArtifactBody,
		"rationale":             priorOutput.Rationale,
		"next_actions":          priorOutput.NextActions,
		"specialist_delegation": priorOutput.SpecialistDelegation,
	})
	if err != nil {
		b.WriteString("{}")
		return b.String()
	}
	b.Write(body)
	return b.String()
}

func trimList(items []string) []string {
	trimmed := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		trimmed = append(trimmed, item)
	}
	return trimmed
}

func normalizeSpecialistDelegation(delegation *CoordinatorSpecialistDelegation) *CoordinatorSpecialistDelegation {
	if delegation == nil {
		return nil
	}
	return &CoordinatorSpecialistDelegation{
		CapabilityCode: strings.TrimSpace(delegation.CapabilityCode),
		Reason:         strings.TrimSpace(delegation.Reason),
	}
}
