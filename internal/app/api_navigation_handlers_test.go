package app

import (
	"testing"
	"time"

	"workflow_app/internal/identityaccess"
	"workflow_app/internal/reporting"
)

func TestFilterRouteCatalogEntriesRespectsRoleAndMultiTermRanking(t *testing.T) {
	adminSession := identityaccess.SessionContext{RoleCode: identityaccess.RoleAdmin}
	operatorSession := identityaccess.SessionContext{RoleCode: identityaccess.RoleOperator}

	adminResults := filterRouteCatalogEntries(adminSession, "pending approvals")
	if len(adminResults) == 0 {
		t.Fatalf("expected admin route-catalog results")
	}
	if adminResults[0].Href != webApprovalsPath {
		t.Fatalf("expected approvals route to rank first for pending approvals, got %+v", adminResults[0])
	}

	operatorResults := filterRouteCatalogEntries(operatorSession, "admin")
	for _, entry := range operatorResults {
		if entry.RequiresRole == identityaccess.RoleAdmin {
			t.Fatalf("expected operator catalog to exclude admin-only entries, found %+v", entry)
		}
	}
}

func TestBuildHomeActionsPrioritizesApproverQueues(t *testing.T) {
	session := identityaccess.SessionContext{RoleCode: identityaccess.RoleApprover}
	inboundSummary := []reporting.InboundRequestStatusSummary{
		{Status: "draft", RequestCount: 2},
	}
	proposalSummary := []reporting.ProcessedProposalStatusSummary{
		{RecommendationStatus: "approval_requested", ProposalCount: 3},
	}
	approvals := []reporting.ApprovalQueueEntry{
		{ApprovalID: "approval-1"},
		{ApprovalID: "approval-2"},
	}

	primary, secondary := buildHomeActions(session, inboundSummary, proposalSummary, approvals)
	if len(primary) < 2 {
		t.Fatalf("expected primary approver actions, got %+v", primary)
	}
	if primary[0].Href != webApprovalsPath+"?status=pending" || primary[0].Badge != "2" {
		t.Fatalf("expected pending approvals to lead approver actions, got %+v", primary[0])
	}
	if primary[1].Href != webProposalsPath+"?status=approval_requested" || primary[1].Badge != "3" {
		t.Fatalf("expected approval-ready proposals to follow, got %+v", primary[1])
	}
	if len(secondary) == 0 || secondary[0].Href != webInboundRequestsPath+"?status=draft" {
		t.Fatalf("expected drafts continuity in secondary actions, got %+v", secondary)
	}
}

func TestSortInboundRequestStatusSummariesUsesWorkflowPriority(t *testing.T) {
	now := time.Now().UTC()
	rows := []reporting.InboundRequestStatusSummary{
		{Status: "processed", LatestUpdatedAt: now.Add(-1 * time.Minute)},
		{Status: "queued", LatestUpdatedAt: now.Add(-10 * time.Minute)},
		{Status: "failed", LatestUpdatedAt: now},
		{Status: "draft", LatestUpdatedAt: now.Add(-20 * time.Minute)},
	}

	sortInboundRequestStatusSummaries(rows)

	got := []string{rows[0].Status, rows[1].Status, rows[2].Status, rows[3].Status}
	want := []string{"draft", "queued", "failed", "processed"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected summary order: got %v want %v", got, want)
		}
	}
}
