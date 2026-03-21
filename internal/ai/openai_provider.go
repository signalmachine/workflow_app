package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

const openAIProviderTimeout = 45 * time.Second

type OpenAIProvider struct {
	client openai.Client
	model  string
}

func NewOpenAIProvider(config ProviderConfig) (*OpenAIProvider, error) {
	if !config.Enabled() {
		return nil, ErrInvalidProviderConfig
	}

	return &OpenAIProvider{
		client: openai.NewClient(option.WithAPIKey(config.OpenAIAPIKey)),
		model:  config.OpenAIModel,
	}, nil
}

func (p *OpenAIProvider) ExecuteInboundRequest(ctx context.Context, input CoordinatorProviderInput) (CoordinatorProviderOutput, error) {
	requestCtx, cancel := context.WithTimeout(ctx, openAIProviderTimeout)
	defer cancel()

	params := responses.ResponseNewParams{
		Model:           p.model,
		Store:           openai.Bool(false),
		Temperature:     openai.Float(0.1),
		MaxOutputTokens: openai.Int(900),
		Instructions: openai.String(strings.TrimSpace(`You are the workflow_app inbound-request coordinator.
Review the persisted request context and produce a structured operator-review brief.
Do not propose direct writes, approval decisions, postings, or autonomous follow-up actions.
Focus on a safe summary, priority, rationale, and next actions for a human operator.`)),
		Input: responses.ResponseNewParamsInputUnion{
			OfString: openai.String(buildProviderPrompt(input)),
		},
		Text: responses.ResponseTextConfigParam{
			Format: coordinatorResponseFormat(),
		},
	}

	resp, err := p.client.Responses.New(requestCtx, params)
	if err != nil {
		var apiErr *openai.Error
		if errors.As(err, &apiErr) {
			return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api error (status %d): %w", apiErr.StatusCode, err)
		}
		return CoordinatorProviderOutput{}, fmt.Errorf("openai responses api request failed: %w", err)
	}

	var parsed openAICoordinatorPayload
	if err := json.Unmarshal([]byte(resp.OutputText()), &parsed); err != nil {
		return CoordinatorProviderOutput{}, fmt.Errorf("decode openai coordinator output: %w", err)
	}

	return CoordinatorProviderOutput{
		ProviderResponseID: resp.ID,
		ProviderName:       "openai",
		Model:              string(resp.Model),
		Summary:            strings.TrimSpace(parsed.Summary),
		Priority:           normalizePriority(parsed.Priority),
		ArtifactTitle:      strings.TrimSpace(parsed.ArtifactTitle),
		ArtifactBody:       strings.TrimSpace(parsed.ArtifactBody),
		Rationale:          trimList(parsed.Rationale),
		NextActions:        trimList(parsed.NextActions),
		InputTokens:        resp.Usage.InputTokens,
		OutputTokens:       resp.Usage.OutputTokens,
		TotalTokens:        resp.Usage.TotalTokens,
	}, nil
}

type openAICoordinatorPayload struct {
	Summary       string   `json:"summary"`
	Priority      string   `json:"priority"`
	ArtifactTitle string   `json:"artifact_title"`
	ArtifactBody  string   `json:"artifact_body"`
	Rationale     []string `json:"rationale"`
	NextActions   []string `json:"next_actions"`
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
		},
		"required": []string{
			"summary",
			"priority",
			"artifact_title",
			"artifact_body",
			"rationale",
			"next_actions",
		},
	}
}

func buildProviderPrompt(input CoordinatorProviderInput) string {
	var b strings.Builder

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
