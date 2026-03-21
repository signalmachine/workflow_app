package ai

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/platform/audit"
)

var (
	ErrRunNotFound            = errors.New("ai run not found")
	ErrRunNotActive           = errors.New("ai run is not active")
	ErrToolNotFound           = errors.New("ai tool not found")
	ErrStepNotFound           = errors.New("ai step not found")
	ErrArtifactNotFound       = errors.New("ai artifact not found")
	ErrRecommendationNotFound = errors.New("ai recommendation not found")
	ErrApprovalNotFound       = errors.New("approval not found")
	ErrInvalidAgentRole       = errors.New("invalid agent role")
	ErrInvalidRunStatus       = errors.New("invalid run status")
	ErrInvalidStepStatus      = errors.New("invalid step status")
	ErrInvalidPolicy          = errors.New("invalid tool policy")
	ErrInvalidDelegation      = errors.New("invalid delegation")
)

const (
	RunRoleCoordinator = "coordinator"
	RunRoleSpecialist  = "specialist"

	RunStatusRunning   = "running"
	RunStatusCompleted = "completed"
	RunStatusFailed    = "failed"
	RunStatusCancelled = "cancelled"

	StepStatusCompleted = "completed"
	StepStatusFailed    = "failed"

	PolicyAllow            = "allow"
	PolicyApprovalRequired = "approval_required"
	PolicyDeny             = "deny"

	RecommendationStatusProposed          = "proposed"
	RecommendationStatusApprovalRequested = "approval_requested"
	RecommendationStatusAccepted          = "accepted"
	RecommendationStatusRejected          = "rejected"
)

type Tool struct {
	ID           string
	ToolName     string
	DisplayName  string
	ModuleCode   string
	MutatesState bool
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Run struct {
	ID               string
	OrgID            string
	SessionID        string
	ActorUserID      string
	InboundRequestID sql.NullString
	AgentRole        string
	CapabilityCode   string
	Status           string
	RequestText      string
	Summary          string
	Metadata         json.RawMessage
	ParentRunID      sql.NullString
	StartedAt        time.Time
	CompletedAt      sql.NullTime
}

type RunStep struct {
	ID            string
	OrgID         string
	RunID         string
	StepIndex     int
	StepType      string
	StepTitle     string
	Status        string
	InputPayload  json.RawMessage
	OutputPayload json.RawMessage
	CreatedAt     time.Time
}

type ToolPolicy struct {
	ID              string
	OrgID           string
	CapabilityCode  string
	ToolID          string
	ToolName        string
	Policy          string
	Rationale       string
	CreatedByUserID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type Artifact struct {
	ID              string
	OrgID           string
	RunID           string
	StepID          sql.NullString
	ArtifactType    string
	Title           string
	Payload         json.RawMessage
	CreatedByUserID string
	CreatedAt       time.Time
}

type Recommendation struct {
	ID                 string
	OrgID              string
	RunID              string
	ArtifactID         sql.NullString
	ApprovalID         sql.NullString
	RecommendationType string
	Status             string
	Summary            string
	Payload            json.RawMessage
	CreatedByUserID    string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Delegation struct {
	ID                string
	OrgID             string
	ParentRunID       string
	ChildRunID        string
	RequestedByStepID sql.NullString
	CapabilityCode    string
	Reason            string
	CreatedAt         time.Time
}

type RegisterToolInput struct {
	ToolName     string
	DisplayName  string
	ModuleCode   string
	MutatesState bool
	Actor        identityaccess.Actor
}

type StartRunInput struct {
	AgentRole        string
	CapabilityCode   string
	InboundRequestID string
	RequestText      string
	Metadata         any
	ParentRunID      string
	Actor            identityaccess.Actor
}

type CompleteRunInput struct {
	RunID    string
	Status   string
	Summary  string
	Metadata any
	Actor    identityaccess.Actor
}

type AppendStepInput struct {
	RunID         string
	StepType      string
	StepTitle     string
	Status        string
	InputPayload  any
	OutputPayload any
	Actor         identityaccess.Actor
}

type SetToolPolicyInput struct {
	CapabilityCode string
	ToolName       string
	Policy         string
	Rationale      string
	Actor          identityaccess.Actor
}

type ResolveToolPolicyInput struct {
	CapabilityCode string
	ToolName       string
	DefaultPolicy  string
	Actor          identityaccess.Actor
}

type ResolvedToolPolicy struct {
	ToolName string
	Policy   string
	Source   string
}

type CreateArtifactInput struct {
	RunID        string
	StepID       string
	ArtifactType string
	Title        string
	Payload      any
	Actor        identityaccess.Actor
}

type CreateRecommendationInput struct {
	RunID              string
	ArtifactID         string
	RecommendationType string
	Summary            string
	Payload            any
	Actor              identityaccess.Actor
}

type LinkRecommendationApprovalInput struct {
	RecommendationID string
	ApprovalID       string
	Actor            identityaccess.Actor
}

type RecordDelegationInput struct {
	ParentRunID       string
	ChildRunID        string
	RequestedByStepID string
	CapabilityCode    string
	Reason            string
	Actor             identityaccess.Actor
}

type Service struct {
	db *sql.DB
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

func (s *Service) RegisterTool(ctx context.Context, input RegisterToolInput) (Tool, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Tool{}, fmt.Errorf("begin register tool: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Tool{}, err
	}

	tool, err := registerToolTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Tool{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.tool_registered",
		EntityType:  "ai.agent_tool",
		EntityID:    tool.ID,
		Payload: map[string]any{
			"tool_name":     tool.ToolName,
			"module_code":   tool.ModuleCode,
			"mutates_state": tool.MutatesState,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Tool{}, err
	}

	if err := tx.Commit(); err != nil {
		return Tool{}, fmt.Errorf("commit register tool: %w", err)
	}

	return tool, nil
}

func (s *Service) StartRun(ctx context.Context, input StartRunInput) (Run, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, fmt.Errorf("begin start run: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	run, err := startRunTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.run_started",
		EntityType:  "ai.agent_run",
		EntityID:    run.ID,
		Payload: map[string]any{
			"agent_role":         run.AgentRole,
			"capability_code":    run.CapabilityCode,
			"parent_run_id":      nullStringValue(run.ParentRunID),
			"inbound_request_id": nullStringValue(run.InboundRequestID),
		},
	}); err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	if err := tx.Commit(); err != nil {
		return Run{}, fmt.Errorf("commit start run: %w", err)
	}

	return run, nil
}

func (s *Service) CompleteRun(ctx context.Context, input CompleteRunInput) (Run, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Run{}, fmt.Errorf("begin complete run: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	run, err := completeRunTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.run_completed",
		EntityType:  "ai.agent_run",
		EntityID:    run.ID,
		Payload: map[string]any{
			"status": run.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Run{}, err
	}

	if err := tx.Commit(); err != nil {
		return Run{}, fmt.Errorf("commit complete run: %w", err)
	}

	return run, nil
}

func (s *Service) AppendStep(ctx context.Context, input AppendStepInput) (RunStep, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RunStep{}, fmt.Errorf("begin append step: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return RunStep{}, err
	}

	step, err := appendStepTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return RunStep{}, err
	}

	if err := tx.Commit(); err != nil {
		return RunStep{}, fmt.Errorf("commit append step: %w", err)
	}

	return step, nil
}

func (s *Service) SetToolPolicy(ctx context.Context, input SetToolPolicyInput) (ToolPolicy, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ToolPolicy{}, fmt.Errorf("begin set tool policy: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return ToolPolicy{}, err
	}

	policy, err := setToolPolicyTx(ctx, tx, input)
	if err != nil {
		_ = tx.Rollback()
		return ToolPolicy{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.tool_policy_set",
		EntityType:  "ai.agent_tool_policy",
		EntityID:    policy.ID,
		Payload: map[string]any{
			"capability_code": policy.CapabilityCode,
			"tool_name":       policy.ToolName,
			"policy":          policy.Policy,
		},
	}); err != nil {
		_ = tx.Rollback()
		return ToolPolicy{}, err
	}

	if err := tx.Commit(); err != nil {
		return ToolPolicy{}, fmt.Errorf("commit set tool policy: %w", err)
	}

	return policy, nil
}

func (s *Service) ResolveToolPolicy(ctx context.Context, input ResolveToolPolicyInput) (ResolvedToolPolicy, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ResolvedToolPolicy{}, fmt.Errorf("begin resolve tool policy: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return ResolvedToolPolicy{}, err
	}

	policy, found, err := resolveToolPolicyTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return ResolvedToolPolicy{}, err
	}
	if err := tx.Commit(); err != nil {
		return ResolvedToolPolicy{}, fmt.Errorf("commit resolve tool policy: %w", err)
	}

	if found {
		return policy, nil
	}
	return ResolvedToolPolicy{
		ToolName: normalizeRequired(input.ToolName),
		Policy:   normalizeDefaultToolPolicy(input.DefaultPolicy),
		Source:   "default",
	}, nil
}

func (s *Service) CreateArtifact(ctx context.Context, input CreateArtifactInput) (Artifact, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Artifact{}, fmt.Errorf("begin create artifact: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Artifact{}, err
	}

	artifact, err := createArtifactTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return Artifact{}, err
	}

	if err := tx.Commit(); err != nil {
		return Artifact{}, fmt.Errorf("commit create artifact: %w", err)
	}

	return artifact, nil
}

func (s *Service) CreateRecommendation(ctx context.Context, input CreateRecommendationInput) (Recommendation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Recommendation{}, fmt.Errorf("begin create recommendation: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	recommendation, err := createRecommendationTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.recommendation_created",
		EntityType:  "ai.agent_recommendation",
		EntityID:    recommendation.ID,
		Payload: map[string]any{
			"recommendation_type": recommendation.RecommendationType,
			"status":              recommendation.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	if err := tx.Commit(); err != nil {
		return Recommendation{}, fmt.Errorf("commit create recommendation: %w", err)
	}

	return recommendation, nil
}

func (s *Service) LinkRecommendationApproval(ctx context.Context, input LinkRecommendationApprovalInput) (Recommendation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Recommendation{}, fmt.Errorf("begin link recommendation approval: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	recommendation, err := linkRecommendationApprovalTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.recommendation_approval_linked",
		EntityType:  "ai.agent_recommendation",
		EntityID:    recommendation.ID,
		Payload: map[string]any{
			"approval_id": nullStringValue(recommendation.ApprovalID),
			"status":      recommendation.Status,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Recommendation{}, err
	}

	if err := tx.Commit(); err != nil {
		return Recommendation{}, fmt.Errorf("commit link recommendation approval: %w", err)
	}

	return recommendation, nil
}

func (s *Service) RecordDelegation(ctx context.Context, input RecordDelegationInput) (Delegation, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Delegation{}, fmt.Errorf("begin record delegation: %w", err)
	}

	if err := authorizeWriteTx(ctx, tx, input.Actor); err != nil {
		_ = tx.Rollback()
		return Delegation{}, err
	}

	delegation, err := recordDelegationTx(ctx, tx, input.Actor.OrgID, input)
	if err != nil {
		_ = tx.Rollback()
		return Delegation{}, err
	}

	if err := audit.WriteTx(ctx, tx, audit.Event{
		OrgID:       input.Actor.OrgID,
		ActorUserID: input.Actor.UserID,
		EventType:   "ai.delegation_recorded",
		EntityType:  "ai.agent_delegation",
		EntityID:    delegation.ID,
		Payload: map[string]any{
			"parent_run_id": delegation.ParentRunID,
			"child_run_id":  delegation.ChildRunID,
			"capability":    delegation.CapabilityCode,
		},
	}); err != nil {
		_ = tx.Rollback()
		return Delegation{}, err
	}

	if err := tx.Commit(); err != nil {
		return Delegation{}, fmt.Errorf("commit record delegation: %w", err)
	}

	return delegation, nil
}

func authorizeWriteTx(ctx context.Context, tx *sql.Tx, actor identityaccess.Actor) error {
	return identityaccess.AuthorizeTx(ctx, tx, actor, identityaccess.RoleAdmin, identityaccess.RoleOperator)
}

func registerToolTx(ctx context.Context, tx *sql.Tx, input RegisterToolInput) (Tool, error) {
	toolName := normalizeRequired(input.ToolName)
	displayName := strings.TrimSpace(input.DisplayName)
	moduleCode := normalizeRequired(input.ModuleCode)
	if toolName == "" || displayName == "" || moduleCode == "" {
		return Tool{}, ErrToolNotFound
	}

	const statement = `
INSERT INTO ai.agent_tools (
	tool_name,
	display_name,
	module_code,
	mutates_state,
	status
) VALUES ($1, $2, $3, $4, 'active')
ON CONFLICT (tool_name) DO UPDATE
SET display_name = EXCLUDED.display_name,
	module_code = EXCLUDED.module_code,
	mutates_state = EXCLUDED.mutates_state,
	status = 'active',
	updated_at = NOW()
RETURNING
	id,
	tool_name,
	display_name,
	module_code,
	mutates_state,
	status,
	created_at,
	updated_at;`

	return scanTool(tx.QueryRowContext(ctx, statement, toolName, displayName, moduleCode, input.MutatesState))
}

func startRunTx(ctx context.Context, tx *sql.Tx, input StartRunInput) (Run, error) {
	role := normalizeRequired(input.AgentRole)
	if role != RunRoleCoordinator && role != RunRoleSpecialist {
		return Run{}, ErrInvalidAgentRole
	}
	capabilityCode := normalizeRequired(input.CapabilityCode)
	if capabilityCode == "" {
		return Run{}, ErrInvalidDelegation
	}
	if input.ParentRunID != "" {
		parentRun, err := getRunForUpdate(ctx, tx, input.Actor.OrgID, input.ParentRunID)
		if err != nil {
			return Run{}, err
		}
		if parentRun.Status != RunStatusRunning {
			return Run{}, ErrRunNotActive
		}
	}
	if input.InboundRequestID != "" {
		requestStatus, err := getInboundRequestStatusForRunTx(ctx, tx, input.Actor.OrgID, input.InboundRequestID)
		if err != nil {
			return Run{}, err
		}
		if requestStatus != "processing" && requestStatus != "processed" && requestStatus != "acted_on" && requestStatus != "completed" {
			return Run{}, ErrRunNotActive
		}
	}

	metadata, err := marshalJSON(input.Metadata)
	if err != nil {
		return Run{}, err
	}

	const statement = `
INSERT INTO ai.agent_runs (
	org_id,
	session_id,
	actor_user_id,
	inbound_request_id,
	agent_role,
	capability_code,
	status,
	request_text,
	metadata,
	parent_run_id
) VALUES ($1, $2, $3, $4, $5, $6, 'running', $7, $8::jsonb, $9)
RETURNING
	id,
	org_id,
	session_id,
	actor_user_id,
	inbound_request_id,
	agent_role,
	capability_code,
	status,
	request_text,
	summary,
	metadata,
	parent_run_id,
	started_at,
	completed_at;`

	return scanRun(tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		input.Actor.SessionID,
		input.Actor.UserID,
		nullIfEmpty(input.InboundRequestID),
		role,
		capabilityCode,
		strings.TrimSpace(input.RequestText),
		string(metadata),
		nullIfEmpty(input.ParentRunID),
	))
}

func completeRunTx(ctx context.Context, tx *sql.Tx, orgID string, input CompleteRunInput) (Run, error) {
	status := normalizeRequired(input.Status)
	if status != RunStatusCompleted && status != RunStatusFailed && status != RunStatusCancelled {
		return Run{}, ErrInvalidRunStatus
	}

	run, err := getRunForUpdate(ctx, tx, orgID, input.RunID)
	if err != nil {
		return Run{}, err
	}
	if run.Status != RunStatusRunning {
		return Run{}, ErrRunNotActive
	}

	metadata, err := marshalJSON(input.Metadata)
	if err != nil {
		return Run{}, err
	}

	const statement = `
UPDATE ai.agent_runs
SET status = $3,
	summary = $4,
	metadata = $5::jsonb,
	completed_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	session_id,
	actor_user_id,
	inbound_request_id,
	agent_role,
	capability_code,
	status,
	request_text,
	summary,
	metadata,
	parent_run_id,
	started_at,
	completed_at;`

	return scanRun(tx.QueryRowContext(ctx, statement, orgID, input.RunID, status, strings.TrimSpace(input.Summary), string(metadata)))
}

func appendStepTx(ctx context.Context, tx *sql.Tx, orgID string, input AppendStepInput) (RunStep, error) {
	status := normalizeRequired(input.Status)
	if status != StepStatusCompleted && status != StepStatusFailed {
		return RunStep{}, ErrInvalidStepStatus
	}
	if strings.TrimSpace(input.StepType) == "" {
		return RunStep{}, ErrStepNotFound
	}

	run, err := getRunForUpdate(ctx, tx, orgID, input.RunID)
	if err != nil {
		return RunStep{}, err
	}
	if run.Status != RunStatusRunning {
		return RunStep{}, ErrRunNotActive
	}

	inputPayload, err := marshalJSON(input.InputPayload)
	if err != nil {
		return RunStep{}, err
	}
	outputPayload, err := marshalJSON(input.OutputPayload)
	if err != nil {
		return RunStep{}, err
	}

	const statement = `
WITH next_step AS (
	SELECT COALESCE(MAX(step_index), 0) + 1 AS step_index
	FROM ai.agent_run_steps
	WHERE run_id = $2
)
INSERT INTO ai.agent_run_steps (
	org_id,
	run_id,
	step_index,
	step_type,
	step_title,
	status,
	input_payload,
	output_payload
)
SELECT $1, $2, step_index, $3, $4, $5, $6::jsonb, $7::jsonb
FROM next_step
RETURNING
	id,
	org_id,
	run_id,
	step_index,
	step_type,
	step_title,
	status,
	input_payload,
	output_payload,
	created_at;`

	return scanRunStep(tx.QueryRowContext(
		ctx,
		statement,
		orgID,
		input.RunID,
		strings.TrimSpace(input.StepType),
		strings.TrimSpace(input.StepTitle),
		status,
		string(inputPayload),
		string(outputPayload),
	))
}

func setToolPolicyTx(ctx context.Context, tx *sql.Tx, input SetToolPolicyInput) (ToolPolicy, error) {
	capabilityCode := normalizeRequired(input.CapabilityCode)
	policyValue := normalizeRequired(input.Policy)
	if capabilityCode == "" {
		return ToolPolicy{}, ErrInvalidPolicy
	}
	if policyValue != PolicyAllow && policyValue != PolicyApprovalRequired && policyValue != PolicyDeny {
		return ToolPolicy{}, ErrInvalidPolicy
	}

	tool, err := getToolByNameForUpdate(ctx, tx, normalizeRequired(input.ToolName))
	if err != nil {
		return ToolPolicy{}, err
	}

	const statement = `
INSERT INTO ai.agent_tool_policies (
	org_id,
	capability_code,
	tool_id,
	policy,
	rationale,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (org_id, capability_code, tool_id) DO UPDATE
SET policy = EXCLUDED.policy,
	rationale = EXCLUDED.rationale,
	created_by_user_id = EXCLUDED.created_by_user_id,
	updated_at = NOW()
RETURNING id;`

	var policyID string
	if err := tx.QueryRowContext(
		ctx,
		statement,
		input.Actor.OrgID,
		capabilityCode,
		tool.ID,
		policyValue,
		strings.TrimSpace(input.Rationale),
		input.Actor.UserID,
	).Scan(&policyID); err != nil {
		return ToolPolicy{}, fmt.Errorf("upsert tool policy: %w", err)
	}

	return getToolPolicy(ctx, tx, input.Actor.OrgID, policyID)
}

func resolveToolPolicyTx(ctx context.Context, tx *sql.Tx, orgID string, input ResolveToolPolicyInput) (ResolvedToolPolicy, bool, error) {
	capabilityCode := normalizeRequired(input.CapabilityCode)
	toolName := normalizeRequired(input.ToolName)
	if capabilityCode == "" || toolName == "" {
		return ResolvedToolPolicy{}, false, ErrInvalidPolicy
	}

	const query = `
SELECT t.tool_name, p.policy
FROM ai.agent_tool_policies p
JOIN ai.agent_tools t
  ON t.id = p.tool_id
WHERE p.org_id = $1
  AND p.capability_code = $2
  AND t.tool_name = $3;`

	var resolved ResolvedToolPolicy
	if err := tx.QueryRowContext(ctx, query, orgID, capabilityCode, toolName).Scan(&resolved.ToolName, &resolved.Policy); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ResolvedToolPolicy{}, false, nil
		}
		return ResolvedToolPolicy{}, false, fmt.Errorf("load resolved tool policy: %w", err)
	}
	resolved.Source = "explicit"
	return resolved, true, nil
}

func createArtifactTx(ctx context.Context, tx *sql.Tx, orgID string, input CreateArtifactInput) (Artifact, error) {
	if strings.TrimSpace(input.ArtifactType) == "" {
		return Artifact{}, ErrArtifactNotFound
	}
	if _, err := getRunForUpdate(ctx, tx, orgID, input.RunID); err != nil {
		return Artifact{}, err
	}
	if input.StepID != "" {
		step, err := getStepForUpdate(ctx, tx, orgID, input.StepID)
		if err != nil {
			return Artifact{}, err
		}
		if step.RunID != input.RunID {
			return Artifact{}, ErrInvalidDelegation
		}
	}

	payload, err := marshalJSON(input.Payload)
	if err != nil {
		return Artifact{}, err
	}

	const statement = `
INSERT INTO ai.agent_artifacts (
	org_id,
	run_id,
	step_id,
	artifact_type,
	title,
	payload,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7)
RETURNING
	id,
	org_id,
	run_id,
	step_id,
	artifact_type,
	title,
	payload,
	created_by_user_id,
	created_at;`

	return scanArtifact(tx.QueryRowContext(
		ctx,
		statement,
		orgID,
		input.RunID,
		nullIfEmpty(input.StepID),
		strings.TrimSpace(input.ArtifactType),
		strings.TrimSpace(input.Title),
		string(payload),
		input.Actor.UserID,
	))
}

func createRecommendationTx(ctx context.Context, tx *sql.Tx, orgID string, input CreateRecommendationInput) (Recommendation, error) {
	if strings.TrimSpace(input.RecommendationType) == "" {
		return Recommendation{}, ErrRecommendationNotFound
	}
	if _, err := getRunForUpdate(ctx, tx, orgID, input.RunID); err != nil {
		return Recommendation{}, err
	}
	if input.ArtifactID != "" {
		artifact, err := getArtifactForUpdate(ctx, tx, orgID, input.ArtifactID)
		if err != nil {
			return Recommendation{}, err
		}
		if artifact.RunID != input.RunID {
			return Recommendation{}, ErrInvalidDelegation
		}
	}

	payload, err := marshalJSON(input.Payload)
	if err != nil {
		return Recommendation{}, err
	}

	const statement = `
INSERT INTO ai.agent_recommendations (
	org_id,
	run_id,
	artifact_id,
	recommendation_type,
	status,
	summary,
	payload,
	created_by_user_id
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8)
RETURNING
	id,
	org_id,
	run_id,
	artifact_id,
	approval_id,
	recommendation_type,
	status,
	summary,
	payload,
	created_by_user_id,
	created_at,
	updated_at;`

	return scanRecommendation(tx.QueryRowContext(
		ctx,
		statement,
		orgID,
		input.RunID,
		nullIfEmpty(input.ArtifactID),
		strings.TrimSpace(input.RecommendationType),
		RecommendationStatusProposed,
		strings.TrimSpace(input.Summary),
		string(payload),
		input.Actor.UserID,
	))
}

func linkRecommendationApprovalTx(ctx context.Context, tx *sql.Tx, orgID string, input LinkRecommendationApprovalInput) (Recommendation, error) {
	if err := ensureApprovalExists(ctx, tx, orgID, input.ApprovalID); err != nil {
		return Recommendation{}, err
	}
	recommendation, err := getRecommendationForUpdate(ctx, tx, orgID, input.RecommendationID)
	if err != nil {
		return Recommendation{}, err
	}

	const statement = `
UPDATE ai.agent_recommendations
SET approval_id = $3,
	status = 'approval_requested',
	updated_at = NOW()
WHERE org_id = $1
  AND id = $2
RETURNING
	id,
	org_id,
	run_id,
	artifact_id,
	approval_id,
	recommendation_type,
	status,
	summary,
	payload,
	created_by_user_id,
	created_at,
	updated_at;`

	updated, err := scanRecommendation(tx.QueryRowContext(ctx, statement, orgID, recommendation.ID, input.ApprovalID))
	if err != nil {
		return Recommendation{}, err
	}
	return updated, nil
}

func recordDelegationTx(ctx context.Context, tx *sql.Tx, orgID string, input RecordDelegationInput) (Delegation, error) {
	parentRun, err := getRunForUpdate(ctx, tx, orgID, input.ParentRunID)
	if err != nil {
		return Delegation{}, err
	}
	childRun, err := getRunForUpdate(ctx, tx, orgID, input.ChildRunID)
	if err != nil {
		return Delegation{}, err
	}
	if parentRun.AgentRole != RunRoleCoordinator || childRun.AgentRole != RunRoleSpecialist {
		return Delegation{}, ErrInvalidDelegation
	}
	if input.RequestedByStepID != "" {
		step, err := getStepForUpdate(ctx, tx, orgID, input.RequestedByStepID)
		if err != nil {
			return Delegation{}, err
		}
		if step.RunID != parentRun.ID {
			return Delegation{}, ErrInvalidDelegation
		}
	}
	capabilityCode := normalizeRequired(input.CapabilityCode)
	if capabilityCode == "" {
		return Delegation{}, ErrInvalidDelegation
	}

	const statement = `
INSERT INTO ai.agent_delegations (
	org_id,
	parent_run_id,
	child_run_id,
	requested_by_step_id,
	capability_code,
	reason
) VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
	id,
	org_id,
	parent_run_id,
	child_run_id,
	requested_by_step_id,
	capability_code,
	reason,
	created_at;`

	return scanDelegation(tx.QueryRowContext(
		ctx,
		statement,
		orgID,
		parentRun.ID,
		childRun.ID,
		nullIfEmpty(input.RequestedByStepID),
		capabilityCode,
		strings.TrimSpace(input.Reason),
	))
}

func getRunForUpdate(ctx context.Context, tx *sql.Tx, orgID, runID string) (Run, error) {
	const query = `
SELECT
	id,
	org_id,
	session_id,
	actor_user_id,
	inbound_request_id,
	agent_role,
	capability_code,
	status,
	request_text,
	summary,
	metadata,
	parent_run_id,
	started_at,
	completed_at
FROM ai.agent_runs
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	run, err := scanRun(tx.QueryRowContext(ctx, query, orgID, runID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Run{}, ErrRunNotFound
		}
		return Run{}, fmt.Errorf("load run: %w", err)
	}
	return run, nil
}

func getStepForUpdate(ctx context.Context, tx *sql.Tx, orgID, stepID string) (RunStep, error) {
	const query = `
SELECT
	id,
	org_id,
	run_id,
	step_index,
	step_type,
	step_title,
	status,
	input_payload,
	output_payload,
	created_at
FROM ai.agent_run_steps
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	step, err := scanRunStep(tx.QueryRowContext(ctx, query, orgID, stepID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return RunStep{}, ErrStepNotFound
		}
		return RunStep{}, fmt.Errorf("load step: %w", err)
	}
	return step, nil
}

func getArtifactForUpdate(ctx context.Context, tx *sql.Tx, orgID, artifactID string) (Artifact, error) {
	const query = `
SELECT
	id,
	org_id,
	run_id,
	step_id,
	artifact_type,
	title,
	payload,
	created_by_user_id,
	created_at
FROM ai.agent_artifacts
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	artifact, err := scanArtifact(tx.QueryRowContext(ctx, query, orgID, artifactID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Artifact{}, ErrArtifactNotFound
		}
		return Artifact{}, fmt.Errorf("load artifact: %w", err)
	}
	return artifact, nil
}

func getRecommendationForUpdate(ctx context.Context, tx *sql.Tx, orgID, recommendationID string) (Recommendation, error) {
	const query = `
SELECT
	id,
	org_id,
	run_id,
	artifact_id,
	approval_id,
	recommendation_type,
	status,
	summary,
	payload,
	created_by_user_id,
	created_at,
	updated_at
FROM ai.agent_recommendations
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	recommendation, err := scanRecommendation(tx.QueryRowContext(ctx, query, orgID, recommendationID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Recommendation{}, ErrRecommendationNotFound
		}
		return Recommendation{}, fmt.Errorf("load recommendation: %w", err)
	}
	return recommendation, nil
}

func getToolByNameForUpdate(ctx context.Context, tx *sql.Tx, toolName string) (Tool, error) {
	const query = `
SELECT
	id,
	tool_name,
	display_name,
	module_code,
	mutates_state,
	status,
	created_at,
	updated_at
FROM ai.agent_tools
WHERE tool_name = $1
FOR UPDATE;`

	tool, err := scanTool(tx.QueryRowContext(ctx, query, toolName))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Tool{}, ErrToolNotFound
		}
		return Tool{}, fmt.Errorf("load tool: %w", err)
	}
	return tool, nil
}

func getToolPolicy(ctx context.Context, tx *sql.Tx, orgID, policyID string) (ToolPolicy, error) {
	const query = `
SELECT
	p.id,
	p.org_id,
	p.capability_code,
	p.tool_id,
	t.tool_name,
	p.policy,
	p.rationale,
	p.created_by_user_id,
	p.created_at,
	p.updated_at
FROM ai.agent_tool_policies p
JOIN ai.agent_tools t
  ON t.id = p.tool_id
WHERE p.org_id = $1
  AND p.id = $2;`

	policy, err := scanToolPolicy(tx.QueryRowContext(ctx, query, orgID, policyID))
	if err != nil {
		return ToolPolicy{}, fmt.Errorf("load tool policy: %w", err)
	}
	return policy, nil
}

func ensureApprovalExists(ctx context.Context, tx *sql.Tx, orgID, approvalID string) error {
	const query = `
SELECT 1
FROM workflow.approvals
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	var found int
	if err := tx.QueryRowContext(ctx, query, orgID, approvalID).Scan(&found); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrApprovalNotFound
		}
		return fmt.Errorf("load approval: %w", err)
	}
	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanTool(row rowScanner) (Tool, error) {
	var tool Tool
	err := row.Scan(
		&tool.ID,
		&tool.ToolName,
		&tool.DisplayName,
		&tool.ModuleCode,
		&tool.MutatesState,
		&tool.Status,
		&tool.CreatedAt,
		&tool.UpdatedAt,
	)
	if err != nil {
		return Tool{}, err
	}
	return tool, nil
}

func scanRun(row rowScanner) (Run, error) {
	var (
		run      Run
		metadata []byte
	)
	err := row.Scan(
		&run.ID,
		&run.OrgID,
		&run.SessionID,
		&run.ActorUserID,
		&run.InboundRequestID,
		&run.AgentRole,
		&run.CapabilityCode,
		&run.Status,
		&run.RequestText,
		&run.Summary,
		&metadata,
		&run.ParentRunID,
		&run.StartedAt,
		&run.CompletedAt,
	)
	if err != nil {
		return Run{}, err
	}
	run.Metadata = metadata
	return run, nil
}

func getInboundRequestStatusForRunTx(ctx context.Context, tx *sql.Tx, orgID, requestID string) (string, error) {
	const query = `
SELECT status
FROM ai.inbound_requests
WHERE org_id = $1
  AND id = $2
FOR UPDATE;`

	var status string
	if err := tx.QueryRowContext(ctx, query, orgID, requestID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrRunNotFound
		}
		return "", fmt.Errorf("load inbound request for run: %w", err)
	}
	return status, nil
}

func scanRunStep(row rowScanner) (RunStep, error) {
	var (
		step          RunStep
		inputPayload  []byte
		outputPayload []byte
	)
	err := row.Scan(
		&step.ID,
		&step.OrgID,
		&step.RunID,
		&step.StepIndex,
		&step.StepType,
		&step.StepTitle,
		&step.Status,
		&inputPayload,
		&outputPayload,
		&step.CreatedAt,
	)
	if err != nil {
		return RunStep{}, err
	}
	step.InputPayload = inputPayload
	step.OutputPayload = outputPayload
	return step, nil
}

func scanToolPolicy(row rowScanner) (ToolPolicy, error) {
	var policy ToolPolicy
	err := row.Scan(
		&policy.ID,
		&policy.OrgID,
		&policy.CapabilityCode,
		&policy.ToolID,
		&policy.ToolName,
		&policy.Policy,
		&policy.Rationale,
		&policy.CreatedByUserID,
		&policy.CreatedAt,
		&policy.UpdatedAt,
	)
	if err != nil {
		return ToolPolicy{}, err
	}
	return policy, nil
}

func scanArtifact(row rowScanner) (Artifact, error) {
	var (
		artifact Artifact
		payload  []byte
	)
	err := row.Scan(
		&artifact.ID,
		&artifact.OrgID,
		&artifact.RunID,
		&artifact.StepID,
		&artifact.ArtifactType,
		&artifact.Title,
		&payload,
		&artifact.CreatedByUserID,
		&artifact.CreatedAt,
	)
	if err != nil {
		return Artifact{}, err
	}
	artifact.Payload = payload
	return artifact, nil
}

func scanRecommendation(row rowScanner) (Recommendation, error) {
	var (
		recommendation Recommendation
		payload        []byte
	)
	err := row.Scan(
		&recommendation.ID,
		&recommendation.OrgID,
		&recommendation.RunID,
		&recommendation.ArtifactID,
		&recommendation.ApprovalID,
		&recommendation.RecommendationType,
		&recommendation.Status,
		&recommendation.Summary,
		&payload,
		&recommendation.CreatedByUserID,
		&recommendation.CreatedAt,
		&recommendation.UpdatedAt,
	)
	if err != nil {
		return Recommendation{}, err
	}
	recommendation.Payload = payload
	return recommendation, nil
}

func scanDelegation(row rowScanner) (Delegation, error) {
	var delegation Delegation
	err := row.Scan(
		&delegation.ID,
		&delegation.OrgID,
		&delegation.ParentRunID,
		&delegation.ChildRunID,
		&delegation.RequestedByStepID,
		&delegation.CapabilityCode,
		&delegation.Reason,
		&delegation.CreatedAt,
	)
	if err != nil {
		return Delegation{}, err
	}
	return delegation, nil
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte(`{}`), nil
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal json payload: %w", err)
	}
	return payload, nil
}

func normalizeRequired(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func normalizeDefaultToolPolicy(value string) string {
	value = normalizeRequired(value)
	switch value {
	case PolicyAllow, PolicyApprovalRequired, PolicyDeny:
		return value
	default:
		return PolicyDeny
	}
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.TrimSpace(value)
}

func nullStringValue(value sql.NullString) any {
	if !value.Valid {
		return nil
	}
	return value.String
}
