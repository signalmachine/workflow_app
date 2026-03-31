# Workflow Approvals And Task Model

Date: 2026-03-31
Status: Active technical guide
Purpose: explain the approval queue, approval decision path, and task ownership model in the workflow package.

## 1. What this package owns

`internal/workflow` owns the review-control layer:

1. approvals
2. approval queues
3. approval decisions
4. task records
5. approval-triggered document state changes

This package is not a generic workflow engine. It is a controlled business review layer.

## 2. Approval lifecycle

Approvals are created when a submitted document needs explicit review.

The usual sequence is:

1. request approval for a submitted document
2. place the approval in a queue
3. review and decide
4. update the document state through the document package

Example:

```go
approval, err := s.RequestApproval(ctx, workflow.RequestApprovalInput{
	DocumentID: document.ID,
	QueueCode:  "appr-ops",
	Reason:     "operator review required",
	Actor:      actor,
})
```

## 3. Decision handling

An approval decision is not just a label change.

It also updates the related document state.

```go
approval, document, err := s.DecideApproval(ctx, workflow.DecideApprovalInput{
	ApprovalID: approval.ID,
	Decision:   "approved",
	Actor:      actor,
})
```

That shape is important because the review record and the shared document record must stay in sync.

## 4. Queue codes

Approvals are queue-aware.

That means different approval pathways can be separated by queue code without inventing separate workflow systems. Queue code is part of the routing and review model.

## 5. Task model

Tasks are the unit of accountable work inside workflow.

The task record includes:

1. context type
2. context ID
3. queue code
4. accountable worker
5. status

That gives the application a small but durable way to track human-controlled work that is not yet a posted business record.

## 6. Why this package is separate from accounting

Workflow owns the decision to approve or reject. Accounting owns the posting of financial truth.

This split avoids a common failure mode:

1. a system posts accounting entries too early
2. approval becomes a side effect instead of a control boundary

`workflow_app` deliberately keeps those roles distinct.

## 7. What to keep stable

Be careful with:

1. approval state transitions
2. queue code handling
3. task ownership
4. document-state synchronization
5. role checks for approvers versus operators

