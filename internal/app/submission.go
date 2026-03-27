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

type SaveInboundDraftInput struct {
	RequestID   string
	MessageID   string
	OriginType  string
	Channel     string
	Metadata    map[string]any
	MessageRole string
	MessageText string
	Attachments []SubmitInboundRequestAttachmentInput
	Actor       identityaccess.Actor
}

type SaveInboundDraftResult struct {
	Request     intake.InboundRequest
	Message     intake.Message
	Attachments []attachments.Attachment
}

type QueueInboundRequestInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type CancelInboundRequestInput struct {
	RequestID string
	Reason    string
	Actor     identityaccess.Actor
}

type AmendInboundRequestInput struct {
	RequestID string
	Actor     identityaccess.Actor
}

type DeleteInboundDraftInput struct {
	RequestID string
	Actor     identityaccess.Actor
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

func (s *SubmissionService) SaveInboundDraft(ctx context.Context, input SaveInboundDraftInput) (_ SaveInboundDraftResult, err error) {
	if s == nil || s.intakeService == nil || s.attachmentService == nil {
		return SaveInboundDraftResult{}, fmt.Errorf("submission service is not initialized")
	}

	requestID := strings.TrimSpace(input.RequestID)
	role := strings.TrimSpace(input.MessageRole)
	if role == "" {
		role = intake.MessageRoleRequest
	}

	var request intake.InboundRequest
	if requestID == "" {
		request, err = s.intakeService.CreateDraft(ctx, intake.CreateDraftInput{
			OriginType: input.OriginType,
			Channel:    input.Channel,
			Metadata:   input.Metadata,
			Actor:      input.Actor,
		})
		if err != nil {
			return SaveInboundDraftResult{}, err
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
	} else {
		request = intake.InboundRequest{ID: requestID}
	}

	var message intake.Message
	messageID := strings.TrimSpace(input.MessageID)
	messageText := strings.TrimSpace(input.MessageText)
	switch {
	case messageID != "":
		message, err = s.intakeService.UpdateMessage(ctx, intake.UpdateMessageInput{
			MessageID:   messageID,
			TextContent: messageText,
			MessageRole: role,
			Actor:       input.Actor,
		})
		if err != nil {
			return SaveInboundDraftResult{}, err
		}
	case messageText != "":
		message, err = s.intakeService.AddMessage(ctx, intake.AddMessageInput{
			RequestID:   request.ID,
			MessageRole: role,
			TextContent: messageText,
			Actor:       input.Actor,
		})
		if err != nil {
			return SaveInboundDraftResult{}, err
		}
	}

	result := SaveInboundDraftResult{
		Request: request,
		Message: message,
	}
	for _, attachmentInput := range input.Attachments {
		content, decodeErr := base64.StdEncoding.DecodeString(strings.TrimSpace(attachmentInput.ContentBase64))
		if decodeErr != nil {
			return SaveInboundDraftResult{}, fmt.Errorf("%w: %v", ErrAttachmentContentEncoding, decodeErr)
		}

		attachment, createErr := s.attachmentService.CreateAttachment(ctx, attachments.CreateAttachmentInput{
			OriginalFileName: attachmentInput.OriginalFileName,
			MediaType:        attachmentInput.MediaType,
			Content:          content,
			Actor:            input.Actor,
		})
		if createErr != nil {
			return SaveInboundDraftResult{}, createErr
		}
		if message.ID == "" {
			message, err = s.intakeService.AddMessage(ctx, intake.AddMessageInput{
				RequestID:   request.ID,
				MessageRole: role,
				TextContent: "",
				Actor:       input.Actor,
			})
			if err != nil {
				return SaveInboundDraftResult{}, err
			}
			result.Message = message
		}

		if _, linkErr := s.attachmentService.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
			RequestMessageID: message.ID,
			AttachmentID:     attachment.ID,
			LinkRole:         attachmentInput.LinkRole,
			Actor:            input.Actor,
		}); linkErr != nil {
			return SaveInboundDraftResult{}, linkErr
		}
		result.Attachments = append(result.Attachments, attachment)
	}

	return result, nil
}

func (s *SubmissionService) QueueInboundRequest(ctx context.Context, input QueueInboundRequestInput) (intake.InboundRequest, error) {
	if s == nil || s.intakeService == nil {
		return intake.InboundRequest{}, fmt.Errorf("submission service is not initialized")
	}
	return s.intakeService.QueueRequest(ctx, intake.QueueRequestInput{
		RequestID: input.RequestID,
		Actor:     input.Actor,
	})
}

func (s *SubmissionService) CancelInboundRequest(ctx context.Context, input CancelInboundRequestInput) (intake.InboundRequest, error) {
	if s == nil || s.intakeService == nil {
		return intake.InboundRequest{}, fmt.Errorf("submission service is not initialized")
	}
	return s.intakeService.CancelRequest(ctx, intake.CancelRequestInput{
		RequestID: input.RequestID,
		Reason:    input.Reason,
		Actor:     input.Actor,
	})
}

func (s *SubmissionService) AmendInboundRequest(ctx context.Context, input AmendInboundRequestInput) (intake.InboundRequest, error) {
	if s == nil || s.intakeService == nil {
		return intake.InboundRequest{}, fmt.Errorf("submission service is not initialized")
	}
	return s.intakeService.AmendRequest(ctx, intake.AmendRequestInput{
		RequestID: input.RequestID,
		Actor:     input.Actor,
	})
}

func (s *SubmissionService) DeleteInboundDraft(ctx context.Context, input DeleteInboundDraftInput) error {
	if s == nil || s.intakeService == nil {
		return fmt.Errorf("submission service is not initialized")
	}
	return s.intakeService.DeleteDraft(ctx, intake.DeleteDraftInput{
		RequestID: input.RequestID,
		Actor:     input.Actor,
	})
}

func (s *SubmissionService) DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (attachments.AttachmentContent, error) {
	if s == nil || s.attachmentService == nil {
		return attachments.AttachmentContent{}, fmt.Errorf("submission service is not initialized")
	}
	return s.attachmentService.GetAttachmentContent(ctx, input.AttachmentID, input.Actor)
}
