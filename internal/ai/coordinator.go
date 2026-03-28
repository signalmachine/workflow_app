package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

const (
	DefaultCoordinatorCapabilityCode  = "inbound_request.coordination"
	coordinatorStepTypeProviderRun    = "provider_execution"
	specialistStepTypeDelegatedReview = "delegated_review"
	coordinatorArtifactType           = "provider_brief"
	coordinatorRecommendationType     = "operator_review"
)

var (
	ErrCoordinatorProviderMissing = errors.New("ai coordinator provider missing")
	ErrInvalidCoordinatorOutput   = errors.New("invalid ai coordinator output")
)

type CoordinatorProvider interface {
	ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error)
}

type CoordinatorProviderInput struct {
	CapabilityCode   string
	Actor            identityaccess.Actor
	RequestReference string
	Channel          string
	OriginType       string
	Metadata         json.RawMessage
	Messages         []CoordinatorMessage
	Attachments      []CoordinatorAttachment
	DerivedTexts     []CoordinatorDerivedText
}

type CoordinatorMessage struct {
	Role        string
	TextContent string
}

type CoordinatorAttachment struct {
	AttachmentID     string
	RequestMessageID string
	LinkRole         string
	OriginalFileName string
	MediaType        string
	SizeBytes        int64
}

type CoordinatorDerivedText struct {
	DerivedTextID      string
	SourceAttachmentID string
	RequestMessageID   string
	DerivativeType     string
	ContentText        string
}

type CoordinatorProviderOutput struct {
	ProviderResponseID   string
	ProviderName         string
	Model                string
	Summary              string
	Priority             string
	ArtifactTitle        string
	ArtifactBody         string
	Rationale            []string
	NextActions          []string
	InputTokens          int64
	OutputTokens         int64
	TotalTokens          int64
	ToolLoopIterations   int
	ToolExecutions       []CoordinatorToolExecution
	SpecialistDelegation *CoordinatorSpecialistDelegation
}

type CoordinatorSpecialistDelegation struct {
	CapabilityCode string `json:"capability_code"`
	Reason         string `json:"reason"`
}

type CoordinatorToolExecution struct {
	Iteration     int    `json:"iteration"`
	ToolName      string `json:"tool_name"`
	Policy        string `json:"policy"`
	Outcome       string `json:"outcome"`
	CallID        string `json:"call_id"`
	ArgumentsJSON string `json:"arguments_json"`
	ResultPreview string `json:"result_preview"`
}

type ProcessNextQueuedInput struct {
	Channel string
	Actor   identityaccess.Actor
}

type ProcessNextQueuedResult struct {
	Request        intake.InboundRequest
	Run            Run
	Step           RunStep
	Delegation     Delegation
	SpecialistRun  Run
	Artifact       Artifact
	Recommendation Recommendation
}

type Coordinator struct {
	intakeService   *intake.Service
	aiService       *Service
	provider        CoordinatorProvider
	capabilityCode  string
	requestLoaderDB *sql.DB
}

func NewCoordinator(db *sql.DB, provider CoordinatorProvider) *Coordinator {
	return &Coordinator{
		intakeService:   intake.NewService(db),
		aiService:       NewService(db),
		provider:        provider,
		capabilityCode:  DefaultCoordinatorCapabilityCode,
		requestLoaderDB: db,
	}
}

func (c *Coordinator) ProcessNextQueued(ctx context.Context, input ProcessNextQueuedInput) (result ProcessNextQueuedResult, err error) {
	if c.provider == nil {
		return result, ErrCoordinatorProviderMissing
	}

	request, err := c.intakeService.ClaimNextQueued(ctx, intake.ClaimNextQueuedInput{
		Channel: strings.TrimSpace(input.Channel),
		Actor:   input.Actor,
	})
	if err != nil {
		return result, err
	}
	result.Request = request

	run, err := c.aiService.StartRun(ctx, StartRunInput{
		AgentRole:        RunRoleCoordinator,
		CapabilityCode:   c.capabilityCode,
		InboundRequestID: request.ID,
		RequestText:      buildRunRequestText(request),
		Metadata: map[string]any{
			"channel":           request.Channel,
			"request_reference": request.RequestReference,
			"provider":          providerName(c.provider),
		},
		Actor: input.Actor,
	})
	if err != nil {
		c.markRequestFailed(ctx, request.ID, "failed to start coordinator run", input.Actor)
		return result, fmt.Errorf("start coordinator run: %w", err)
	}
	result.Run = run

	requestContext, err := c.loadRequestContext(ctx, input.Actor, request.ID)
	if err != nil {
		c.failRunAndRequest(ctx, run, "failed to load inbound request context", input.Actor)
		return result, fmt.Errorf("load inbound request context: %w", err)
	}

	providerOutput, execErr := c.provider.ExecuteInboundRequest(ctx, requestContext)
	if execErr != nil {
		step, stepErr := c.aiService.AppendStep(ctx, AppendStepInput{
			RunID:     run.ID,
			StepType:  coordinatorStepTypeProviderRun,
			StepTitle: "Provider execution failed",
			Status:    StepStatusFailed,
			InputPayload: map[string]any{
				"request_reference":  request.RequestReference,
				"message_count":      len(requestContext.Messages),
				"attachment_count":   len(requestContext.Attachments),
				"derived_text_count": len(requestContext.DerivedTexts),
			},
			OutputPayload: map[string]any{
				"error": execErr.Error(),
			},
			Actor: input.Actor,
		})
		if stepErr == nil {
			result.Step = step
		}
		c.failRunAndRequest(ctx, run, sanitizeFailureReason(execErr.Error()), input.Actor)
		return result, fmt.Errorf("execute provider-backed coordinator: %w", execErr)
	}

	if err := validateCoordinatorProviderOutput(providerOutput); err != nil {
		step, stepErr := c.aiService.AppendStep(ctx, AppendStepInput{
			RunID:     run.ID,
			StepType:  coordinatorStepTypeProviderRun,
			StepTitle: "Provider output validation failed",
			Status:    StepStatusFailed,
			InputPayload: map[string]any{
				"request_reference": request.RequestReference,
			},
			OutputPayload: map[string]any{
				"error": err.Error(),
			},
			Actor: input.Actor,
		})
		if stepErr == nil {
			result.Step = step
		}
		c.failRunAndRequest(ctx, run, sanitizeFailureReason(err.Error()), input.Actor)
		return result, err
	}
	if err := validateRequestCenteredCoordinatorOutput(requestContext, providerOutput); err != nil {
		step, stepErr := c.aiService.AppendStep(ctx, AppendStepInput{
			RunID:     run.ID,
			StepType:  coordinatorStepTypeProviderRun,
			StepTitle: "Provider output request-centering validation failed",
			Status:    StepStatusFailed,
			InputPayload: map[string]any{
				"request_reference": request.RequestReference,
			},
			OutputPayload: map[string]any{
				"error": err.Error(),
			},
			Actor: input.Actor,
		})
		if stepErr == nil {
			result.Step = step
		}
		c.failRunAndRequest(ctx, run, sanitizeFailureReason(err.Error()), input.Actor)
		return result, err
	}
	providerOutput.Priority = normalizePriority(providerOutput.Priority)

	step, err := c.aiService.AppendStep(ctx, AppendStepInput{
		RunID:     run.ID,
		StepType:  coordinatorStepTypeProviderRun,
		StepTitle: "Execute provider-backed coordinator review",
		Status:    StepStatusCompleted,
		InputPayload: map[string]any{
			"request_reference":  request.RequestReference,
			"message_count":      len(requestContext.Messages),
			"attachment_count":   len(requestContext.Attachments),
			"derived_text_count": len(requestContext.DerivedTexts),
		},
		OutputPayload: map[string]any{
			"provider":              providerOutput.ProviderName,
			"provider_response_id":  providerOutput.ProviderResponseID,
			"model":                 providerOutput.Model,
			"priority":              providerOutput.Priority,
			"input_tokens":          providerOutput.InputTokens,
			"output_tokens":         providerOutput.OutputTokens,
			"total_tokens":          providerOutput.TotalTokens,
			"tool_loop_iterations":  providerOutput.ToolLoopIterations,
			"tool_executions":       providerOutput.ToolExecutions,
			"specialist_delegation": providerOutput.SpecialistDelegation,
		},
		Actor: input.Actor,
	})
	if err != nil {
		c.failRunAndRequest(ctx, run, "failed to record provider execution step", input.Actor)
		return result, fmt.Errorf("append coordinator step: %w", err)
	}
	result.Step = step

	artifactRun := run
	artifactStep := step
	if providerOutput.SpecialistDelegation != nil {
		var delegation Delegation
		artifactRun, artifactStep, delegation, err = c.createDelegatedSpecialistRun(ctx, request, run, step, requestContext, providerOutput, input.Actor)
		if err != nil {
			c.failRunAndRequest(ctx, run, sanitizeFailureReason(err.Error()), input.Actor)
			return result, fmt.Errorf("create delegated specialist run: %w", err)
		}
		result.Delegation = delegation
		result.SpecialistRun = artifactRun
		if artifactRun.ParentRunID.Valid {
			result.Run = run
		}
	}

	artifact, err := c.aiService.CreateArtifact(ctx, CreateArtifactInput{
		RunID:        artifactRun.ID,
		StepID:       artifactStep.ID,
		ArtifactType: coordinatorArtifactType,
		Title:        providerOutput.ArtifactTitle,
		Payload: map[string]any{
			"provider":              providerOutput.ProviderName,
			"provider_response_id":  providerOutput.ProviderResponseID,
			"model":                 providerOutput.Model,
			"request_reference":     request.RequestReference,
			"summary":               providerOutput.Summary,
			"priority":              providerOutput.Priority,
			"body":                  providerOutput.ArtifactBody,
			"rationale":             providerOutput.Rationale,
			"next_actions":          providerOutput.NextActions,
			"tool_loop_iterations":  providerOutput.ToolLoopIterations,
			"tool_executions":       providerOutput.ToolExecutions,
			"specialist_delegation": providerOutput.SpecialistDelegation,
		},
		Actor: input.Actor,
	})
	if err != nil {
		c.failRunAndRequest(ctx, run, "failed to persist provider artifact", input.Actor)
		return result, fmt.Errorf("create coordinator artifact: %w", err)
	}
	result.Artifact = artifact

	recommendation, err := c.aiService.CreateRecommendation(ctx, CreateRecommendationInput{
		RunID:              artifactRun.ID,
		ArtifactID:         artifact.ID,
		RecommendationType: coordinatorRecommendationType,
		Summary:            providerOutput.Summary,
		Payload: map[string]any{
			"provider":              providerOutput.ProviderName,
			"model":                 providerOutput.Model,
			"request_reference":     request.RequestReference,
			"priority":              providerOutput.Priority,
			"next_actions":          providerOutput.NextActions,
			"rationale":             providerOutput.Rationale,
			"tool_loop_iterations":  providerOutput.ToolLoopIterations,
			"tool_executions":       providerOutput.ToolExecutions,
			"specialist_delegation": providerOutput.SpecialistDelegation,
		},
		Actor: input.Actor,
	})
	if err != nil {
		c.failRunAndRequest(ctx, run, "failed to persist provider recommendation", input.Actor)
		return result, fmt.Errorf("create coordinator recommendation: %w", err)
	}
	result.Recommendation = recommendation

	if providerOutput.SpecialistDelegation != nil {
		specialistRun, err := c.aiService.CompleteRun(ctx, CompleteRunInput{
			RunID:   artifactRun.ID,
			Status:  RunStatusCompleted,
			Summary: providerOutput.Summary,
			Metadata: map[string]any{
				"provider":              providerOutput.ProviderName,
				"provider_response_id":  providerOutput.ProviderResponseID,
				"model":                 providerOutput.Model,
				"priority":              providerOutput.Priority,
				"recommendation_id":     recommendation.ID,
				"specialist_delegation": providerOutput.SpecialistDelegation,
			},
			Actor: input.Actor,
		})
		if err != nil {
			c.failRunAndRequest(ctx, run, "failed to complete specialist run", input.Actor)
			return result, fmt.Errorf("complete specialist run: %w", err)
		}
		result.SpecialistRun = specialistRun
	}

	runSummary := providerOutput.Summary
	runMetadata := map[string]any{
		"provider":             providerOutput.ProviderName,
		"provider_response_id": providerOutput.ProviderResponseID,
		"model":                providerOutput.Model,
		"priority":             providerOutput.Priority,
		"recommendation_id":    recommendation.ID,
	}
	if providerOutput.SpecialistDelegation != nil {
		runSummary = fmt.Sprintf(
			"Delegated inbound request %s to %s",
			request.RequestReference,
			providerOutput.SpecialistDelegation.CapabilityCode,
		)
		runMetadata["specialist_delegation"] = providerOutput.SpecialistDelegation
		if result.SpecialistRun.ID != "" {
			runMetadata["child_run_id"] = result.SpecialistRun.ID
		}
		if result.Delegation.ID != "" {
			runMetadata["delegation_id"] = result.Delegation.ID
		}
	}
	run, err = c.aiService.CompleteRun(ctx, CompleteRunInput{
		RunID:    run.ID,
		Status:   RunStatusCompleted,
		Summary:  runSummary,
		Metadata: runMetadata,
		Actor:    input.Actor,
	})
	if err != nil {
		c.markRequestFailed(ctx, request.ID, "failed to complete coordinator run", input.Actor)
		return result, fmt.Errorf("complete coordinator run: %w", err)
	}
	result.Run = run

	request, err = c.intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID: request.ID,
		Status:    intake.StatusProcessed,
		Actor:     input.Actor,
	})
	if err != nil {
		return result, fmt.Errorf("mark inbound request processed: %w", err)
	}
	result.Request = request

	return result, nil
}

func (c *Coordinator) createDelegatedSpecialistRun(
	ctx context.Context,
	request intake.InboundRequest,
	parentRun Run,
	parentStep RunStep,
	requestContext CoordinatorProviderInput,
	providerOutput CoordinatorProviderOutput,
	actor identityaccess.Actor,
) (Run, RunStep, Delegation, error) {
	specialist := providerOutput.SpecialistDelegation
	if specialist == nil {
		return Run{}, RunStep{}, Delegation{}, ErrInvalidCoordinatorOutput
	}

	childRun, err := c.aiService.StartRun(ctx, StartRunInput{
		AgentRole:        RunRoleSpecialist,
		CapabilityCode:   specialist.CapabilityCode,
		InboundRequestID: request.ID,
		ParentRunID:      parentRun.ID,
		RequestText:      fmt.Sprintf("Handle inbound request %s via delegated specialist review", request.RequestReference),
		Metadata: map[string]any{
			"request_reference":     request.RequestReference,
			"parent_run_id":         parentRun.ID,
			"provider":              providerOutput.ProviderName,
			"provider_response_id":  providerOutput.ProviderResponseID,
			"specialist_delegation": specialist,
		},
		Actor: actor,
	})
	if err != nil {
		return Run{}, RunStep{}, Delegation{}, err
	}

	delegation, err := c.aiService.RecordDelegation(ctx, RecordDelegationInput{
		ParentRunID:       parentRun.ID,
		ChildRunID:        childRun.ID,
		RequestedByStepID: parentStep.ID,
		CapabilityCode:    specialist.CapabilityCode,
		Reason:            specialist.Reason,
		Actor:             actor,
	})
	if err != nil {
		_, _ = c.aiService.CompleteRun(ctx, CompleteRunInput{
			RunID:   childRun.ID,
			Status:  RunStatusFailed,
			Summary: "failed to record delegation",
			Metadata: map[string]any{
				"failure_reason": "failed to record delegation",
			},
			Actor: actor,
		})
		return Run{}, RunStep{}, Delegation{}, err
	}

	step, err := c.aiService.AppendStep(ctx, AppendStepInput{
		RunID:     childRun.ID,
		StepType:  specialistStepTypeDelegatedReview,
		StepTitle: "Produce delegated specialist review",
		Status:    StepStatusCompleted,
		InputPayload: map[string]any{
			"request_reference":  request.RequestReference,
			"parent_run_id":      parentRun.ID,
			"delegation_id":      delegation.ID,
			"delegation_reason":  specialist.Reason,
			"message_count":      len(requestContext.Messages),
			"attachment_count":   len(requestContext.Attachments),
			"derived_text_count": len(requestContext.DerivedTexts),
		},
		OutputPayload: map[string]any{
			"provider":              providerOutput.ProviderName,
			"provider_response_id":  providerOutput.ProviderResponseID,
			"model":                 providerOutput.Model,
			"priority":              providerOutput.Priority,
			"summary":               providerOutput.Summary,
			"tool_loop_iterations":  providerOutput.ToolLoopIterations,
			"tool_executions":       providerOutput.ToolExecutions,
			"specialist_delegation": specialist,
		},
		Actor: actor,
	})
	if err != nil {
		_, _ = c.aiService.CompleteRun(ctx, CompleteRunInput{
			RunID:   childRun.ID,
			Status:  RunStatusFailed,
			Summary: "failed to record delegated specialist step",
			Metadata: map[string]any{
				"failure_reason": "failed to record delegated specialist step",
			},
			Actor: actor,
		})
		return Run{}, RunStep{}, Delegation{}, err
	}

	return childRun, step, delegation, nil
}

func (c *Coordinator) loadRequestContext(ctx context.Context, actor identityaccess.Actor, requestID string) (CoordinatorProviderInput, error) {
	request, err := c.loadRequestRow(ctx, actor.OrgID, requestID)
	if err != nil {
		return CoordinatorProviderInput{}, err
	}

	messages, err := c.loadRequestMessages(ctx, actor.OrgID, requestID)
	if err != nil {
		return CoordinatorProviderInput{}, err
	}

	attachments, err := c.loadRequestAttachments(ctx, actor.OrgID, requestID)
	if err != nil {
		return CoordinatorProviderInput{}, err
	}

	derivedTexts, err := c.loadRequestDerivedTexts(ctx, actor.OrgID, requestID)
	if err != nil {
		return CoordinatorProviderInput{}, err
	}

	return CoordinatorProviderInput{
		CapabilityCode:   c.capabilityCode,
		Actor:            actor,
		RequestReference: request.RequestReference,
		Channel:          request.Channel,
		OriginType:       request.OriginType,
		Metadata:         request.Metadata,
		Messages:         messages,
		Attachments:      attachments,
		DerivedTexts:     derivedTexts,
	}, nil
}

func (c *Coordinator) loadRequestRow(ctx context.Context, orgID, requestID string) (intake.InboundRequest, error) {
	const query = `
SELECT
	id,
	org_id,
	request_number,
	request_reference,
	session_id,
	actor_user_id,
	origin_type,
	channel,
	status,
	metadata,
	cancellation_reason,
	failure_reason,
	received_at,
	queued_at,
	processing_started_at,
	processed_at,
	acted_on_at,
	completed_at,
	failed_at,
	cancelled_at,
	created_at,
	updated_at
FROM ai.inbound_requests
WHERE org_id = $1
  AND id = $2;`

	return scanInboundRequestContext(c.requestLoaderDB.QueryRowContext(ctx, query, orgID, requestID))
}

func (c *Coordinator) loadRequestMessages(ctx context.Context, orgID, requestID string) ([]CoordinatorMessage, error) {
	const query = `
SELECT message_role, text_content
FROM ai.inbound_request_messages
WHERE org_id = $1
  AND request_id = $2
ORDER BY message_index ASC;`

	rows, err := c.requestLoaderDB.QueryContext(ctx, query, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("query inbound request messages: %w", err)
	}
	defer rows.Close()

	var messages []CoordinatorMessage
	for rows.Next() {
		var message CoordinatorMessage
		if err := rows.Scan(&message.Role, &message.TextContent); err != nil {
			return nil, fmt.Errorf("scan inbound request message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound request messages: %w", err)
	}
	return messages, nil
}

func (c *Coordinator) loadRequestAttachments(ctx context.Context, orgID, requestID string) ([]CoordinatorAttachment, error) {
	const query = `
SELECT
	a.id,
	rml.request_message_id,
	rml.link_role,
	a.original_file_name,
	a.media_type,
	a.size_bytes
FROM attachments.request_message_links rml
JOIN ai.inbound_request_messages m
  ON m.org_id = rml.org_id
 AND m.id = rml.request_message_id
JOIN attachments.attachments a
  ON a.org_id = rml.org_id
 AND a.id = rml.attachment_id
WHERE rml.org_id = $1
  AND m.request_id = $2
ORDER BY a.created_at ASC, a.id ASC;`

	rows, err := c.requestLoaderDB.QueryContext(ctx, query, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("query inbound request attachments: %w", err)
	}
	defer rows.Close()

	var attachments []CoordinatorAttachment
	for rows.Next() {
		var attachment CoordinatorAttachment
		if err := rows.Scan(
			&attachment.AttachmentID,
			&attachment.RequestMessageID,
			&attachment.LinkRole,
			&attachment.OriginalFileName,
			&attachment.MediaType,
			&attachment.SizeBytes,
		); err != nil {
			return nil, fmt.Errorf("scan inbound request attachment: %w", err)
		}
		attachments = append(attachments, attachment)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound request attachments: %w", err)
	}
	return attachments, nil
}

func (c *Coordinator) loadRequestDerivedTexts(ctx context.Context, orgID, requestID string) ([]CoordinatorDerivedText, error) {
	const query = `
SELECT
	dt.id,
	dt.source_attachment_id,
	COALESCE(dt.request_message_id::text, ''),
	dt.derivative_type,
	dt.content_text
FROM attachments.derived_texts dt
WHERE dt.org_id = $1
  AND (
	(dt.request_message_id IS NOT NULL AND EXISTS (
		SELECT 1
		FROM ai.inbound_request_messages m
		WHERE m.org_id = dt.org_id
		  AND m.id = dt.request_message_id
		  AND m.request_id = $2
	))
	OR EXISTS (
		SELECT 1
		FROM attachments.request_message_links rml
		JOIN ai.inbound_request_messages m
		  ON m.org_id = rml.org_id
		 AND m.id = rml.request_message_id
		WHERE rml.org_id = dt.org_id
		  AND rml.attachment_id = dt.source_attachment_id
		  AND m.request_id = $2
	)
  )
ORDER BY dt.created_at ASC, dt.id ASC;`

	rows, err := c.requestLoaderDB.QueryContext(ctx, query, orgID, requestID)
	if err != nil {
		return nil, fmt.Errorf("query inbound request derived texts: %w", err)
	}
	defer rows.Close()

	var derivedTexts []CoordinatorDerivedText
	for rows.Next() {
		var derived CoordinatorDerivedText
		if err := rows.Scan(
			&derived.DerivedTextID,
			&derived.SourceAttachmentID,
			&derived.RequestMessageID,
			&derived.DerivativeType,
			&derived.ContentText,
		); err != nil {
			return nil, fmt.Errorf("scan inbound request derived text: %w", err)
		}
		derivedTexts = append(derivedTexts, derived)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate inbound request derived texts: %w", err)
	}
	return derivedTexts, nil
}

func (c *Coordinator) failRunAndRequest(ctx context.Context, run Run, reason string, actor identityaccess.Actor) {
	_, _ = c.aiService.CompleteRun(ctx, CompleteRunInput{
		RunID:   run.ID,
		Status:  RunStatusFailed,
		Summary: reason,
		Metadata: map[string]any{
			"failure_reason": reason,
		},
		Actor: actor,
	})
	c.markRequestFailed(ctx, run.InboundRequestID.String, reason, actor)
}

func (c *Coordinator) markRequestFailed(ctx context.Context, requestID, reason string, actor identityaccess.Actor) {
	if strings.TrimSpace(requestID) == "" {
		return
	}
	_, _ = c.intakeService.AdvanceRequest(ctx, intake.AdvanceRequestInput{
		RequestID:     requestID,
		Status:        intake.StatusFailed,
		FailureReason: sanitizeFailureReason(reason),
		Actor:         actor,
	})
}

func buildRunRequestText(request intake.InboundRequest) string {
	return fmt.Sprintf("Process inbound request %s from channel %s", request.RequestReference, request.Channel)
}

func validateCoordinatorProviderOutput(output CoordinatorProviderOutput) error {
	if strings.TrimSpace(output.ProviderName) == "" {
		return fmt.Errorf("%w: provider name is required", ErrInvalidCoordinatorOutput)
	}
	if strings.TrimSpace(output.Model) == "" {
		return fmt.Errorf("%w: model is required", ErrInvalidCoordinatorOutput)
	}
	if strings.TrimSpace(output.Summary) == "" {
		return fmt.Errorf("%w: summary is required", ErrInvalidCoordinatorOutput)
	}
	if strings.TrimSpace(output.ArtifactTitle) == "" {
		return fmt.Errorf("%w: artifact title is required", ErrInvalidCoordinatorOutput)
	}
	if strings.TrimSpace(output.ArtifactBody) == "" {
		return fmt.Errorf("%w: artifact body is required", ErrInvalidCoordinatorOutput)
	}
	if output.ToolLoopIterations < 0 {
		return fmt.Errorf("%w: tool loop iterations cannot be negative", ErrInvalidCoordinatorOutput)
	}

	priority := normalizePriority(output.Priority)
	if priority == "" {
		return fmt.Errorf("%w: priority must be low, normal, high, or urgent", ErrInvalidCoordinatorOutput)
	}
	if output.SpecialistDelegation != nil {
		if strings.TrimSpace(output.SpecialistDelegation.Reason) == "" {
			return fmt.Errorf("%w: specialist delegation reason is required", ErrInvalidCoordinatorOutput)
		}
		if !isAllowedSpecialistCapability(output.SpecialistDelegation.CapabilityCode) {
			return fmt.Errorf("%w: unsupported specialist capability %q", ErrInvalidCoordinatorOutput, output.SpecialistDelegation.CapabilityCode)
		}
	}

	return nil
}

func validateRequestCenteredCoordinatorOutput(input CoordinatorProviderInput, output CoordinatorProviderOutput) error {
	combinedOutput := strings.ToLower(strings.Join(append([]string{
		output.Summary,
		output.ArtifactTitle,
		output.ArtifactBody,
	}, append(output.Rationale, output.NextActions...)...), " "))

	if containsTransientLifecycleWording(combinedOutput) {
		return fmt.Errorf("%w: output uses transient lifecycle wording instead of a stable operator brief", ErrInvalidCoordinatorOutput)
	}

	keywords := requestEvidenceKeywords(input)
	if len(keywords) == 0 {
		return nil
	}
	for _, keyword := range keywords {
		if strings.Contains(combinedOutput, keyword) {
			return nil
		}
	}

	return fmt.Errorf("%w: output does not stay materially specific to the request content", ErrInvalidCoordinatorOutput)
}

func containsTransientLifecycleWording(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return false
	}

	patterns := []string{
		"currently processing",
		"still processing",
		"request is processing",
		"request remains processing",
		"in processing",
		"merely processing",
		"only processing",
	}
	for _, pattern := range patterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

func requestEvidenceKeywords(input CoordinatorProviderInput) []string {
	candidates := make(map[string]struct{})
	addKeywords := func(text string) {
		for _, token := range strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsDigit(r)
		}) {
			token = strings.TrimSpace(token)
			if len(token) < 4 {
				continue
			}
			if isCoordinatorStopword(token) {
				continue
			}
			candidates[token] = struct{}{}
		}
	}

	for _, message := range input.Messages {
		addKeywords(message.TextContent)
	}
	for _, derived := range input.DerivedTexts {
		addKeywords(derived.ContentText)
	}

	keywords := make([]string, 0, len(candidates))
	for keyword := range candidates {
		keywords = append(keywords, keyword)
	}
	sort.Slice(keywords, func(i, j int) bool {
		if len(keywords[i]) == len(keywords[j]) {
			return keywords[i] < keywords[j]
		}
		return len(keywords[i]) > len(keywords[j])
	})
	return keywords
}

func isCoordinatorStopword(token string) bool {
	switch token {
	case "about", "after", "attached", "because", "browser", "channel", "confirm", "content", "customer", "details", "derived", "front", "human", "immediately", "issue", "label", "message", "metadata", "needs", "note", "operator", "please", "queue", "queued", "request", "review", "role", "safe", "site", "source", "status", "submitter", "summary", "texts", "through", "urgent", "voice", "with":
		return true
	default:
		return false
	}
}

func isAllowedSpecialistCapability(capabilityCode string) bool {
	switch strings.TrimSpace(capabilityCode) {
	case "inbound_request.operations_triage", "inbound_request.approval_triage":
		return true
	default:
		return false
	}
}

func normalizePriority(priority string) string {
	switch strings.ToLower(strings.TrimSpace(priority)) {
	case "low", "normal", "high", "urgent":
		return strings.ToLower(strings.TrimSpace(priority))
	default:
		return ""
	}
}

func providerName(provider CoordinatorProvider) string {
	switch provider.(type) {
	case *OpenAIProvider:
		return "openai"
	default:
		return "custom"
	}
}

func sanitizeFailureReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "provider-backed coordinator execution failed"
	}
	if len(reason) > 500 {
		return reason[:500]
	}
	return reason
}

type requestContextScanner interface {
	Scan(dest ...any) error
}

func scanInboundRequestContext(row requestContextScanner) (intake.InboundRequest, error) {
	var request intake.InboundRequest
	var metadata []byte
	err := row.Scan(
		&request.ID,
		&request.OrgID,
		&request.RequestNumber,
		&request.RequestReference,
		&request.SessionID,
		&request.ActorUserID,
		&request.OriginType,
		&request.Channel,
		&request.Status,
		&metadata,
		&request.CancellationReason,
		&request.FailureReason,
		&request.ReceivedAt,
		&request.QueuedAt,
		&request.ProcessingStartedAt,
		&request.ProcessedAt,
		&request.ActedOnAt,
		&request.CompletedAt,
		&request.FailedAt,
		&request.CancelledAt,
		&request.CreatedAt,
		&request.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return intake.InboundRequest{}, intake.ErrInboundRequestNotFound
		}
		return intake.InboundRequest{}, err
	}
	request.Metadata = append(request.Metadata[:0], metadata...)
	return request, nil
}
