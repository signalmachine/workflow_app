package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"workflow_app/internal/ai"
	"workflow_app/internal/identityaccess"
)

var ErrAgentProviderNotConfigured = errors.New("ai provider not configured")

type ProcessNextQueuedInboundRequestInput struct {
	Channel string
	Actor   identityaccess.Actor
}

type ProcessNextQueuedInboundRequestResult = ai.ProcessNextQueuedResult

type AgentProcessor struct {
	coordinator *ai.Coordinator
}

func NewAgentProcessor(db *sql.DB, provider ai.CoordinatorProvider) (*AgentProcessor, error) {
	if db == nil {
		return nil, fmt.Errorf("agent processor requires database")
	}
	if provider == nil {
		return nil, fmt.Errorf("%w: coordinator provider missing", ErrAgentProviderNotConfigured)
	}

	return &AgentProcessor{
		coordinator: ai.NewCoordinator(db, provider),
	}, nil
}

func NewOpenAIAgentProcessorFromEnv(db *sql.DB) (*AgentProcessor, error) {
	config, err := ai.LoadProviderConfigFromEnv()
	if err != nil {
		return nil, err
	}
	if !config.Enabled() {
		return nil, ErrAgentProviderNotConfigured
	}

	provider, err := ai.NewOpenAIProvider(db, config)
	if err != nil {
		return nil, err
	}

	return NewAgentProcessor(db, provider)
}

func (p *AgentProcessor) ProcessNextQueuedInboundRequest(ctx context.Context, input ProcessNextQueuedInboundRequestInput) (ProcessNextQueuedInboundRequestResult, error) {
	if p == nil || p.coordinator == nil {
		return ProcessNextQueuedInboundRequestResult{}, fmt.Errorf("agent processor is not initialized")
	}

	return p.coordinator.ProcessNextQueued(ctx, ai.ProcessNextQueuedInput{
		Channel: strings.TrimSpace(input.Channel),
		Actor:   input.Actor,
	})
}
