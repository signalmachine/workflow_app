# Attachments And Derived Text

Date: 2026-03-31
Status: Active technical guide
Purpose: explain how `workflow_app` stores attachment bytes, links them to requests, and records derived text such as transcription.

## 1. What attachments own

`internal/attachments` owns:

1. attachment bytes
2. media-type validation
3. attachment links to request messages
4. derived text records
5. attachment content retrieval

Attachments are intentionally part of the intake foundation rather than a separate file-management product.

## 2. Storage model

Thin-v1 stores attachment content in PostgreSQL.

That is acceptable for the current phase because the design still leaves a clean path to move binary storage to object storage later.

The important point is that the content is still durably tied to the request workflow.

## 3. Attachment creation

```go
attachment, err := s.CreateAttachment(ctx, attachments.CreateAttachmentInput{
	OriginalFileName: "note.txt",
	MediaType:        "text/plain",
	Content:          []byte("voice transcription"),
	Actor:            actor,
})
```

The service validates size, media type, and other attachment constraints before persistence.

## 4. Request-message links

An attachment becomes workflow-visible when it is linked to a request message.

```go
_, err := s.LinkRequestMessage(ctx, attachments.LinkRequestMessageInput{
	RequestMessageID: message.ID,
	AttachmentID:     attachment.ID,
	LinkRole:         attachments.LinkRoleSource,
	Actor:            actor,
})
```

That link is what lets the intake detail page show the attachment in context rather than as an orphaned file.

## 5. Derived text

Derived text is the durable derivative of an attachment, such as a transcription.

The key rule is that derived text does not replace the original attachment. It supplements it.

That means the app can preserve:

1. the original upload
2. the derivative text
3. the provenance between them

## 6. Why this matters for AI

The coordinator can use attachments and derived text as evidence when creating a review brief.

That is only safe if the original artifact stays durable and discoverable. The attachments package exists to make that possible.

## 7. What to keep stable

Be careful with:

1. size limits
2. media-type validation
3. link roles
4. derived-text provenance
5. content retrieval headers

