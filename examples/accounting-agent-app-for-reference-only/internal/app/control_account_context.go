package app

import "context"

type proposalSourceKey struct{}
type controlOverrideKey struct{}

// ProposalSource identifies the caller origin for JE policy handling.
type ProposalSource string

const (
	ProposalSourceManualWeb ProposalSource = "manual_web"
	ProposalSourceAIAgent   ProposalSource = "ai_agent"
	ProposalSourceCLI       ProposalSource = "cli"
	ProposalSourceREPL      ProposalSource = "repl"
)

type controlOverride struct {
	Enabled bool
	Reason  string
	Role    string
}

// WithProposalSource annotates context with proposal source.
func WithProposalSource(ctx context.Context, src ProposalSource) context.Context {
	return context.WithValue(ctx, proposalSourceKey{}, src)
}

func proposalSourceFromContext(ctx context.Context) ProposalSource {
	v, _ := ctx.Value(proposalSourceKey{}).(ProposalSource)
	return v
}

// WithControlAccountOverride annotates context with override metadata.
func WithControlAccountOverride(ctx context.Context, enabled bool, reason, role string) context.Context {
	return context.WithValue(ctx, controlOverrideKey{}, controlOverride{
		Enabled: enabled,
		Reason:  reason,
		Role:    role,
	})
}

func controlOverrideFromContext(ctx context.Context) controlOverride {
	v, _ := ctx.Value(controlOverrideKey{}).(controlOverride)
	return v
}
