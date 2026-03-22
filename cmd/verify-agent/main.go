package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"workflow_app/internal/app"
	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

const verifyTimeout = 2 * time.Minute

func main() {
	var (
		databaseURL     string
		channel         string
		requestText     string
		attachmentText  string
		attachmentName  string
		attachmentMedia string
		submitterLabel  string
	)

	flag.StringVar(&databaseURL, "database-url", firstNonEmpty(os.Getenv("TEST_DATABASE_URL"), os.Getenv("DATABASE_URL")), "PostgreSQL connection string for verification data")
	flag.StringVar(&channel, "channel", "browser", "inbound request channel to queue and process")
	flag.StringVar(&requestText, "request-text", "Customer reports an urgent warehouse pump failure and needs operator review.", "request message text to queue for live verification")
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

	actor, err := createVerificationActor(ctx, db)
	if err != nil {
		log.Fatalf("create verification actor: %v", err)
	}

	request, err := createVerificationRequest(ctx, db, actor, verificationRequestInput{
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
		Actor:   actor,
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
}

type verificationRequestInput struct {
	Channel         string
	RequestText     string
	AttachmentText  string
	AttachmentName  string
	AttachmentMedia string
	SubmitterLabel  string
}

func createVerificationActor(ctx context.Context, db *sql.DB) (identityaccess.Actor, error) {
	orgSlug := "verify-agent-" + time.Now().UTC().Format("20060102-150405.000000000")
	email := "verify-agent-" + time.Now().UTC().Format("150405.000000000") + "@example.com"

	var orgID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.orgs (slug, name) VALUES ($1, $2) RETURNING id`,
		orgSlug,
		"Verify Agent Org",
	).Scan(&orgID); err != nil {
		return identityaccess.Actor{}, fmt.Errorf("insert org: %w", err)
	}

	var userID string
	if err := db.QueryRowContext(
		ctx,
		`INSERT INTO identityaccess.users (email, display_name) VALUES ($1, $2) RETURNING id`,
		email,
		"Verify Agent Operator",
	).Scan(&userID); err != nil {
		return identityaccess.Actor{}, fmt.Errorf("insert user: %w", err)
	}

	if _, err := db.ExecContext(
		ctx,
		`INSERT INTO identityaccess.memberships (org_id, user_id, role_code) VALUES ($1, $2, $3)`,
		orgID,
		userID,
		identityaccess.RoleOperator,
	); err != nil {
		return identityaccess.Actor{}, fmt.Errorf("insert membership: %w", err)
	}

	session, err := identityaccess.NewService(db).StartSession(ctx, identityaccess.StartSessionInput{
		OrgID:            orgID,
		UserID:           userID,
		DeviceLabel:      "verify-agent",
		RefreshTokenHash: "verify-agent-token-" + time.Now().UTC().Format("150405.000000000"),
		ExpiresAt:        time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		return identityaccess.Actor{}, fmt.Errorf("start session: %w", err)
	}

	return identityaccess.Actor{
		OrgID:     orgID,
		UserID:    userID,
		SessionID: session.ID,
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

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
