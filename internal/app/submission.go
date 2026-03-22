package app

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
)

var ErrAttachmentContentEncoding = errors.New("invalid attachment content encoding")

type SubmitInboundRequestInput struct {
	OriginType     string
	Channel        string
	Metadata       map[string]any
	MessageRole    string
	MessageText    string
	Attachments    []SubmitInboundRequestAttachmentInput
	QueueForReview bool
	Actor          identityaccess.Actor
}

type SubmitInboundRequestAttachmentInput struct {
	OriginalFileName string
	MediaType        string
	ContentBase64    string
	LinkRole         string
}

type SubmitInboundRequestResult struct {
	Request     intake.InboundRequest
	Message     intake.Message
	Attachments []attachments.Attachment
}

type DownloadAttachmentInput struct {
	AttachmentID string
	Actor        identityaccess.Actor
}

type SubmissionService struct {
	intakeService     *intake.Service
	attachmentService *attachments.Service
}

func NewSubmissionService(db *sql.DB) *SubmissionService {
	return &SubmissionService{
		intakeService:     intake.NewService(db),
		attachmentService: attachments.NewService(db),
	}
}

func (s *SubmissionService) SubmitInboundRequest(ctx context.Context, input SubmitInboundRequestInput) (_ SubmitInboundRequestResult, err error) {
	if s == nil || s.intakeService == nil || s.attachmentService == nil {
		return SubmitInboundRequestResult{}, fmt.Errorf("submission service is not initialized")
	}

	queued := input.QueueForReview
	if !queued {
		queued = true
	}

	if strings.TrimSpace(input.MessageText) == "" && len(input.Attachments) == 0 {
		return SubmitInboundRequestResult{}, intake.ErrInvalidInboundRequest
	}

	request, err := s.intakeService.CreateDraft(ctx, intake.CreateDraftInput{
		OriginType: input.OriginType,
		Channel:    input.Channel,
		Metadata:   input.Metadata,
		Actor:      input.Actor,
	})
	if err != nil {
		return SubmitInboundRequestResult{}, err
	}

	defer func() {
		if err == nil {
			return
		}
		_ = s.intakeService.DeleteDraft(ctx, intake.DeleteDraftInput{
			RequestID: request.ID,
			Actor:     input.Actor,
		})
	}()

	message, err := s.intakeService.AddMessage(ctx, intake.AddMessageInput{
		RequestID:   request.ID,
		MessageRole: input.MessageRole,
		TextContent: input.MessageText,
		Actor:       input.Actor,
	})
	if err != nil {
		return SubmitInboundRequestResult{}, err
	}

	result := SubmitInboundRequestResult{
		Request: request,
		Message: message,
	}

	for _, attachmentInput := range input.Attachments {
		content, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(attachmentInput.ContentBase64))
		if decodeErr != nil {
			return SubmitInboundRequestResult{}, fmt.Errorf("%w: %v", ErrAttachmentContentEncoding, decodeErr)
		}

		attachment, createErr := s.attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
			OriginalFileName: attachmentInput.OriginalFileName,
			MediaType:        attachmentInput.MediaType,
			Content:          content,
			Actor:            input.Actor,
		})
		if createErr != nil {
			return SubmitInboundRequestResult{}, createErr
		}

		if _, linkErr := s.attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
			RequestMessageID: message.ID,
			AttachmentID:     attachment.ID,
			LinkRole:         attachmentInput.LinkRole,
			Actor:            input.Actor,
		}); linkErr != nil {
			return SubmitInboundRequestResult{}, linkErr
		}

		result.Attachments = append(result.Attachments, attachment)
	}

	if queued {
		request, err = s.intakeService.QueueRequest(ctx, intake.QueueRequestInput{
			RequestID: request.ID,
			Actor:     input.Actor,
		})
		if err != nil {
			return SubmitInboundRequestResult{}, err
		}
		result.Request = request
	}

	return result, nil
}

func (s *SubmissionService) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (attachments.AttachmentContent, error) {
	if s == nil || s.attachmentService == nil {
		return attachments.AttachmentContent{}, fmt.Errorf("submission service is not initialized")
	}
	return s.attachmentService.GetAttachmentContent(ctx, input.AttachmentID, input.Actor)
}
