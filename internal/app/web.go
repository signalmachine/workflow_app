package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"workflow_app/internal/attachments"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
)

type webAppDashboardData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	RoleHeadline     string
	RoleBody         string
	PrimaryActions   []webHomeAction
	SecondaryActions []webHomeAction
	InboundSummary   []reporting.InboundRequestStatusSummary
	ProposalSummary  []reporting.ProcessedProposalStatusSummary
	InboundRequests  []reporting.InboundRequestReview
	Proposals        []reporting.ProcessedProposalReview
	Approvals        []reporting.ApprovalQueueEntry
}

type webInboundDetailData struct {
	Session                identityaccess.SessionContext
	Notice                 string
	Error                  string
	Detail                 reporting.InboundRequestDetail
	EditableMessageID      string
	EditableMessageText    string
	EditableSubmitterLabel string
}

type webInboundRequestsData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	Status           string
	RequestReference string
	StatusSummary    []reporting.InboundRequestStatusSummary
	Requests         []reporting.InboundRequestReview
}

type webInboundSubmitData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	RequestReference string
	RequestStatus    string
}

type webOperationsFeedItem struct {
	OccurredAt     time.Time
	Kind           string
	Title          string
	Summary        string
	Status         string
	PrimaryLabel   string
	PrimaryHref    string
	SecondaryLabel string
	SecondaryHref  string
}

type webOperationsFeedData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Feed    []webOperationsFeedItem
}

type webOperationsLandingData struct {
	Session              identityaccess.SessionContext
	Notice               string
	Error                string
	QueuedRequestCount   int
	PendingApprovalCount int
	ProposalReviewCount  int
	RecentFeed           []webOperationsFeedItem
}

type webAgentChatData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	RequestReference string
	RequestStatus    string
	RecentRequests   []reporting.InboundRequestReview
	RecentProposals  []reporting.ProcessedProposalReview
}

type webDocumentsData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	DocumentID string
	TypeCode   string
	Status     string
	Documents  []reporting.DocumentReview
}

type webDocumentDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.DocumentReview
}

type webAccountingData struct {
	Session         identityaccess.SessionContext
	Notice          string
	Error           string
	StartOn         string
	EndOn           string
	AsOf            string
	EntryID         string
	DocumentID      string
	TaxType         string
	TaxCode         string
	ControlType     string
	AccountID       string
	JournalEntries  []reporting.JournalEntryReview
	ControlBalances []reporting.ControlAccountBalance
	TaxSummaries    []reporting.TaxSummary
}

type webAccountingDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.JournalEntryReview
}

type webControlAccountDetailData struct {
	Session          identityaccess.SessionContext
	Notice           string
	Error            string
	StartOn          string
	EndOn            string
	AsOf             string
	Balance          reporting.ControlAccountBalance
	RelatedSummaries []reporting.TaxSummary
}

type webTaxSummaryDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	StartOn string
	EndOn   string
	Summary reporting.TaxSummary
}

type webProposalsData struct {
	Session            identityaccess.SessionContext
	Notice             string
	Error              string
	RecommendationID   string
	Status             string
	RequestReference   string
	StatusSummary      []reporting.ProcessedProposalStatusSummary
	ProcessedProposals []reporting.ProcessedProposalReview
}

type webProposalDetailData struct {
	Session                identityaccess.SessionContext
	Notice                 string
	Error                  string
	Review                 reporting.ProcessedProposalReview
	ApprovalReason         string
	ApprovalQueueCodeDraft string
}

type webApprovalsData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	ApprovalID string
	Status     string
	QueueCode  string
	Approvals  []reporting.ApprovalQueueEntry
}

type webApprovalDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Entry   reporting.ApprovalQueueEntry
}

type webReviewLandingData struct {
	Session             identityaccess.SessionContext
	Notice              string
	Error               string
	InboundSummary      []reporting.InboundRequestStatusSummary
	ProposalSummary     []reporting.ProcessedProposalStatusSummary
	PendingApprovals    []reporting.ApprovalQueueEntry
	InboundRequestCount int
	ProposalCount       int
}

type webInventoryData struct {
	Session               identityaccess.SessionContext
	Notice                string
	Error                 string
	MovementID            string
	ItemID                string
	LocationID            string
	DocumentID            string
	MovementType          string
	OnlyPendingAccounting bool
	OnlyPendingExecution  bool
	Stock                 []reporting.InventoryStockItem
	Movements             []reporting.InventoryMovementReview
	Reconciliation        []reporting.InventoryReconciliationItem
}

type webInventoryDetailData struct {
	Session        identityaccess.SessionContext
	Notice         string
	Error          string
	Review         reporting.InventoryMovementReview
	Reconciliation []reporting.InventoryReconciliationItem
}

type webInventoryItemDetailData struct {
	Session        identityaccess.SessionContext
	Notice         string
	Error          string
	ItemID         string
	ItemSKU        string
	ItemName       string
	ItemRole       string
	Stock          []reporting.InventoryStockItem
	Movements      []reporting.InventoryMovementReview
	Reconciliation []reporting.InventoryReconciliationItem
}

type webInventoryLocationDetailData struct {
	Session      identityaccess.SessionContext
	Notice       string
	Error        string
	LocationID   string
	LocationCode string
	LocationName string
	LocationRole string
	Stock        []reporting.InventoryStockItem
	Movements    []reporting.InventoryMovementReview
}

type webInventoryLandingData struct {
	Session                identityaccess.SessionContext
	Notice                 string
	Error                  string
	Stock                  []reporting.InventoryStockItem
	Movements              []reporting.InventoryMovementReview
	Reconciliation         []reporting.InventoryReconciliationItem
	PendingExecutionCount  int
	PendingAccountingCount int
}

type webWorkOrdersData struct {
	Session     identityaccess.SessionContext
	Notice      string
	Error       string
	WorkOrderID string
	Status      string
	DocumentID  string
	WorkOrders  []reporting.WorkOrderReview
}

type webWorkOrderDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Review  reporting.WorkOrderReview
}

type webAuditData struct {
	Session    identityaccess.SessionContext
	Notice     string
	Error      string
	EventID    string
	EntityType string
	EntityID   string
	Events     []reporting.AuditEvent
}

type webAuditDetailData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Event   reporting.AuditEvent
}

type webHomeAction struct {
	Title   string
	Summary string
	Href    string
	Badge   string
}

type webRouteCatalogEntry struct {
	Title        string
	Href         string
	Category     string
	Summary      string
	Keywords     string
	RequiresRole string
}

type webRouteCatalogData struct {
	Session identityaccess.SessionContext
	Notice  string
	Error   string
	Query   string
	Results []webRouteCatalogEntry
}

type webSettingsData struct {
	Session        identityaccess.SessionContext
	Notice         string
	Error          string
	PrimaryActions []webHomeAction
}

type webAdminData struct {
	Session      identityaccess.SessionContext
	Notice       string
	Error        string
	AdminActions []webHomeAction
}

func (h *AgentAPIHandler) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, webAppPath, http.StatusSeeOther)
}

func proposalQueueCode(review reporting.ProcessedProposalReview) string {
	if review.ApprovalQueueCode.Valid {
		return review.ApprovalQueueCode.String
	}
	if review.SuggestedQueueCode.Valid {
		return review.SuggestedQueueCode.String
	}
	return ""
}

type webPageData struct {
	Title                   string
	ActivePath              string
	Notice                  string
	Error                   string
	ShowLogin               bool
	LoginPath               string
	Session                 *identityaccess.SessionContext
	Dashboard               *webAppDashboardData
	RouteCatalog            *webRouteCatalogData
	Settings                *webSettingsData
	Admin                   *webAdminData
	OperationsLanding       *webOperationsLandingData
	OperationsFeed          *webOperationsFeedData
	AgentChat               *webAgentChatData
	InboundSubmit           *webInboundSubmitData
	InboundRequests         *webInboundRequestsData
	Detail                  *webInboundDetailData
	Documents               *webDocumentsData
	DocumentDetail          *webDocumentDetailData
	Accounting              *webAccountingData
	AccountingDetail        *webAccountingDetailData
	ControlAccountDetail    *webControlAccountDetailData
	TaxSummaryDetail        *webTaxSummaryDetailData
	Approvals               *webApprovalsData
	ApprovalDetail          *webApprovalDetailData
	ReviewLanding           *webReviewLandingData
	Proposals               *webProposalsData
	ProposalDetail          *webProposalDetailData
	InventoryLanding        *webInventoryLandingData
	Inventory               *webInventoryData
	InventoryDetail         *webInventoryDetailData
	InventoryItemDetail     *webInventoryItemDetailData
	InventoryLocationDetail *webInventoryLocationDetailData
	WorkOrders              *webWorkOrdersData
	WorkOrderDetail         *webWorkOrderDetailData
	Audit                   *webAuditData
	AuditDetail             *webAuditDetailData
}

func sanitizeWebReturnPath(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, webAppPath) {
		return ""
	}
	if strings.Contains(raw, "://") {
		return ""
	}
	return raw
}

func appendWebMessage(target, key, message string) string {
	separator := "?"
	if strings.Contains(target, "?") {
		separator = "&"
	}
	return target + separator + key + "=" + url.QueryEscape(message)
}

const (
	inboundRequestChannelBrowser   = "browser"
	inboundRequestChannelAgentChat = "agent_chat"
)

func normalizeWebInboundChannel(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case inboundRequestChannelAgentChat:
		return inboundRequestChannelAgentChat
	default:
		return inboundRequestChannelBrowser
	}
}

func buildOperationsFeedFromRequests(items []reporting.InboundRequestReview) []webOperationsFeedItem {
	feed := make([]webOperationsFeedItem, 0, len(items))
	for _, item := range items {
		summary := fmt.Sprintf("%s via %s with %d messages and %d attachments.", item.RequestReference, item.Channel, item.MessageCount, item.AttachmentCount)
		switch {
		case item.FailureReason != "":
			summary = item.FailureReason
		case item.CancellationReason != "":
			summary = item.CancellationReason
		}
		secondaryLabel := ""
		secondaryHref := ""
		if item.LastRecommendationID.Valid {
			secondaryLabel = "Open proposal"
			secondaryHref = templateProposalDetailHref(item.LastRecommendationID.String)
		} else if item.LastRunID.Valid {
			secondaryLabel = "Open latest run"
			secondaryHref = templateInboundRequestSectionHref("run:"+item.LastRunID.String, templateAIRunSectionID(item.LastRunID.String))
		}
		feed = append(feed, webOperationsFeedItem{
			OccurredAt:     item.UpdatedAt,
			Kind:           "Request status",
			Title:          fmt.Sprintf("%s moved through %s", item.RequestReference, item.Status),
			Summary:        summary,
			Status:         item.Status,
			PrimaryLabel:   "Open request",
			PrimaryHref:    templateInboundRequestHref(item.RequestReference),
			SecondaryLabel: secondaryLabel,
			SecondaryHref:  secondaryHref,
		})
	}
	return feed
}

func buildOperationsFeedFromProposals(items []reporting.ProcessedProposalReview) []webOperationsFeedItem {
	feed := make([]webOperationsFeedItem, 0, len(items))
	for _, item := range items {
		summary := fmt.Sprintf("%s for %s", item.Summary, item.RequestReference)
		secondaryLabel := "Open request"
		secondaryHref := templateInboundRequestHref(item.RequestReference)
		if item.DocumentID.Valid {
			secondaryLabel = "Open document"
			secondaryHref = templateDocumentReviewHref(item.DocumentID.String)
		} else if item.ApprovalID.Valid {
			secondaryLabel = "Open approval"
			secondaryHref = templateApprovalReviewHref(item.ApprovalID.String)
		}
		feed = append(feed, webOperationsFeedItem{
			OccurredAt:     item.CreatedAt,
			Kind:           "Coordinator proposal",
			Title:          fmt.Sprintf("%s proposal for %s", item.RecommendationStatus, item.RequestReference),
			Summary:        summary,
			Status:         item.RecommendationStatus,
			PrimaryLabel:   "Open proposal",
			PrimaryHref:    templateProposalDetailHref(item.RecommendationID),
			SecondaryLabel: secondaryLabel,
			SecondaryHref:  secondaryHref,
		})
	}
	return feed
}

func buildOperationsFeedFromApprovals(items []reporting.ApprovalQueueEntry) []webOperationsFeedItem {
	feed := make([]webOperationsFeedItem, 0, len(items))
	for _, item := range items {
		occurredAt := item.RequestedAt
		if item.ClosedAt.Valid {
			occurredAt = item.ClosedAt.Time
		}
		summary := fmt.Sprintf("%s on %s", item.QueueCode, item.DocumentTitle)
		if item.RequestReference.Valid {
			summary += " for " + item.RequestReference.String
		}
		feed = append(feed, webOperationsFeedItem{
			OccurredAt:     occurredAt,
			Kind:           "Approval queue",
			Title:          fmt.Sprintf("%s approval is %s", item.DocumentTitle, item.ApprovalStatus),
			Summary:        summary,
			Status:         item.ApprovalStatus,
			PrimaryLabel:   "Open approval",
			PrimaryHref:    templateApprovalReviewHref(item.ApprovalID),
			SecondaryLabel: "Open document",
			SecondaryHref:  templateDocumentReviewHref(item.DocumentID),
		})
	}
	return feed
}

func sumInboundRequestCount(rows []reporting.InboundRequestStatusSummary) int {
	total := 0
	for _, row := range rows {
		total += row.RequestCount
	}
	return total
}

func sumProposalCount(rows []reporting.ProcessedProposalStatusSummary) int {
	total := 0
	for _, row := range rows {
		total += row.ProposalCount
	}
	return total
}

func countQueuedRequests(rows []reporting.InboundRequestStatusSummary) int {
	for _, row := range rows {
		if strings.EqualFold(strings.TrimSpace(row.Status), "queued") {
			return row.RequestCount
		}
	}
	return 0
}

func countPendingReconciliation(rows []reporting.InventoryReconciliationItem, field string) int {
	count := 0
	for _, row := range rows {
		switch field {
		case "execution":
			if row.ExecutionLinkStatus.Valid && strings.EqualFold(strings.TrimSpace(row.ExecutionLinkStatus.String), "pending") {
				count++
			}
		case "accounting":
			if row.AccountingHandoffStatus.Valid && strings.EqualFold(strings.TrimSpace(row.AccountingHandoffStatus.String), "pending") {
				count++
			}
		}
	}
	return count
}

func parseMultipartAttachments(form *multipart.Form) ([]SubmitInboundRequestAttachmentInput, error) {
	var files []SubmitInboundRequestAttachmentInput
	if form == nil {
		return files, nil
	}
	for _, fileHeader := range form.File["attachments"] {
		file, openErr := fileHeader.Open()
		if openErr != nil {
			return nil, openErr
		}
		content, readErr := io.ReadAll(file)
		_ = file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if len(content) == 0 {
			continue
		}

		mediaType := strings.TrimSpace(fileHeader.Header.Get("Content-Type"))
		if mediaType == "" {
			mediaType = "application/octet-stream"
		}
		files = append(files, SubmitInboundRequestAttachmentInput{
			OriginalFileName: fileHeader.Filename,
			MediaType:        mediaType,
			ContentBase64:    base64.StdEncoding.EncodeToString(content),
			LinkRole:         attachments.LinkRoleEvidence,
		})
	}
	return files, nil
}

func editableInboundMessageID(detail reporting.InboundRequestDetail) string {
	for _, message := range detail.Messages {
		if message.MessageRole == intake.MessageRoleRequest {
			return message.MessageID
		}
	}
	if len(detail.Messages) == 0 {
		return ""
	}
	return detail.Messages[0].MessageID
}

func editableInboundMessageText(detail reporting.InboundRequestDetail) string {
	messageID := editableInboundMessageID(detail)
	for _, message := range detail.Messages {
		if message.MessageID == messageID {
			return message.TextContent
		}
	}
	return ""
}

func inboundRequestMetadataString(raw json.RawMessage, key string) string {
	if len(raw) == 0 || strings.TrimSpace(key) == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	value, ok := payload[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func formatTemplateTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format("2006-01-02 15:04:05 UTC")
}

func formatDateInput(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.DateOnly)
}

func prettyTemplateJSON(raw any) string {
	switch v := raw.(type) {
	case nil:
		return "{}"
	case []byte:
		if len(v) == 0 {
			return "{}"
		}
		var out bytes.Buffer
		if err := json.Indent(&out, v, "", "  "); err == nil {
			return out.String()
		}
		return string(v)
	default:
		b, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(b)
	}
}

func templateStatusClass(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "processed", "completed", "approved":
		return "status-good"
	case "failed", "rejected", "cancelled":
		return "status-bad"
	default:
		return "status-neutral"
	}
}

func templateDashboardStatusBlurb(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft":
		return "Resume parked drafts before they enter the queue."
	case "queued":
		return "Review queued requests and open lifecycle controls before pickup."
	case "processing":
		return "Watch active coordinator work and continue into the latest run."
	case "failed":
		return "Inspect failed requests, understand the break, and restart follow-up work."
	case "cancelled":
		return "Return cancelled pre-processing requests to draft when they should be resubmitted."
	case "processed", "acted_on", "completed":
		return "Continue from completed intake into proposals, approvals, and downstream review."
	default:
		return "Open the filtered request review for this state."
	}
}

func templateDashboardStatusAction(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft":
		return "Continue drafts"
	case "queued":
		return "Open queued requests"
	case "processing":
		return "Watch in-flight requests"
	case "failed":
		return "Review failures"
	case "cancelled":
		return "Recover cancellations"
	case "processed", "acted_on", "completed":
		return "Review outcomes"
	default:
		return "Open requests"
	}
}

func templateDashboardRequestAction(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft":
		return "Continue draft"
	case "queued":
		return "Open lifecycle actions"
	case "processing":
		return "Watch execution"
	case "failed":
		return "Inspect failure"
	case "cancelled":
		return "Amend back to draft"
	case "processed", "acted_on", "completed":
		return "Open request outcome"
	default:
		return "Open request detail"
	}
}

func sortInboundRequestStatusSummaries(rows []reporting.InboundRequestStatusSummary) {
	statusOrder := map[string]int{
		"draft":      0,
		"queued":     1,
		"processing": 2,
		"failed":     3,
		"cancelled":  4,
		"processed":  5,
		"acted_on":   6,
		"completed":  7,
	}
	sort.SliceStable(rows, func(i, j int) bool {
		left := strings.ToLower(strings.TrimSpace(rows[i].Status))
		right := strings.ToLower(strings.TrimSpace(rows[j].Status))
		leftRank, leftOK := statusOrder[left]
		rightRank, rightOK := statusOrder[right]
		switch {
		case leftOK && rightOK && leftRank != rightRank:
			return leftRank < rightRank
		case leftOK != rightOK:
			return leftOK
		case rows[i].LatestUpdatedAt != rows[j].LatestUpdatedAt:
			return rows[i].LatestUpdatedAt.After(rows[j].LatestUpdatedAt)
		default:
			return left < right
		}
	})
}

func templateInboundRequestHref(lookup string) string {
	lookup = strings.TrimSpace(lookup)
	if lookup == "" {
		return webInboundRequestsPath
	}
	return webInboundDetailPrefix + url.PathEscape(lookup)
}

func templateInboundRequestSectionHref(lookup, sectionID string) string {
	target := templateInboundRequestHref(lookup)
	sectionID = strings.TrimSpace(sectionID)
	if sectionID == "" {
		return target
	}
	return target + "#" + sectionID
}

func templateInboundActionHref(requestID, action string) string {
	requestID = strings.TrimSpace(requestID)
	action = strings.TrimSpace(action)
	if requestID == "" || action == "" {
		return webInboundRequestsPath
	}
	return webInboundDetailPrefix + url.PathEscape(requestID) + "/" + url.PathEscape(action)
}

func templateAIRunSectionID(runID string) string {
	return templateSectionID("run", runID)
}

func templateAIStepSectionID(stepID string) string {
	return templateSectionID("step", stepID)
}

func templateAIDelegationSectionID(delegationID string) string {
	return templateSectionID("delegation", delegationID)
}

func templatePageSectionHref(sectionID string) string {
	sectionID = strings.TrimSpace(sectionID)
	if sectionID == "" {
		return "#"
	}
	return "#" + sectionID
}

func templateSectionID(prefix, value string) string {
	prefix = strings.ToLower(strings.TrimSpace(prefix))
	if prefix == "" {
		prefix = "section"
	}
	var builder strings.Builder
	builder.Grow(len(prefix) + len(value) + 1)
	builder.WriteString(prefix)
	builder.WriteByte('-')
	lastDash := true
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r + ('a' - 'A'))
			lastDash = false
		default:
			if !lastDash {
				builder.WriteByte('-')
				lastDash = true
			}
		}
	}
	result := strings.TrimSuffix(builder.String(), "-")
	if result == prefix {
		return prefix
	}
	return result
}

func templateInboundRequestReviewHref(requestReference string) string {
	requestReference = strings.TrimSpace(requestReference)
	if requestReference == "" {
		return webInboundRequestsPath
	}
	return webInboundRequestsPath + "?request_reference=" + url.QueryEscape(requestReference)
}

func templateInboundRequestsReviewHref(status, requestReference string) string {
	values := url.Values{}
	if strings.TrimSpace(status) != "" {
		values.Set("status", strings.TrimSpace(status))
	}
	if strings.TrimSpace(requestReference) != "" {
		values.Set("request_reference", strings.TrimSpace(requestReference))
	}
	if encoded := values.Encode(); encoded != "" {
		return webInboundRequestsPath + "?" + encoded
	}
	return webInboundRequestsPath
}

func templateDocumentReviewHref(documentID string) string {
	documentID = strings.TrimSpace(documentID)
	if documentID == "" {
		return webDocumentsPath
	}
	return webDocumentDetailPrefix + url.PathEscape(documentID)
}

func templateAccountingEntryHref(entryID string) string {
	entryID = strings.TrimSpace(entryID)
	if entryID == "" {
		return webAccountingPath
	}
	return webAccountingDetailPrefix + url.PathEscape(entryID)
}

func templateAccountingReviewHref(startOn, endOn, asOf, entryID, documentID, taxType, taxCode, controlType, accountID, anchor string) string {
	values := url.Values{}
	if strings.TrimSpace(startOn) != "" {
		values.Set("start_on", strings.TrimSpace(startOn))
	}
	if strings.TrimSpace(endOn) != "" {
		values.Set("end_on", strings.TrimSpace(endOn))
	}
	if strings.TrimSpace(asOf) != "" {
		values.Set("as_of", strings.TrimSpace(asOf))
	}
	if strings.TrimSpace(entryID) != "" {
		values.Set("entry_id", strings.TrimSpace(entryID))
	}
	if strings.TrimSpace(documentID) != "" {
		values.Set("document_id", strings.TrimSpace(documentID))
	}
	if strings.TrimSpace(taxType) != "" {
		values.Set("tax_type", strings.TrimSpace(taxType))
	}
	if strings.TrimSpace(taxCode) != "" {
		values.Set("tax_code", strings.TrimSpace(taxCode))
	}
	if strings.TrimSpace(controlType) != "" {
		values.Set("control_type", strings.TrimSpace(controlType))
	}
	if strings.TrimSpace(accountID) != "" {
		values.Set("account_id", strings.TrimSpace(accountID))
	}

	target := webAccountingPath
	if encoded := values.Encode(); encoded != "" {
		target += "?" + encoded
	}
	anchor = strings.TrimSpace(anchor)
	if anchor != "" {
		target += "#" + strings.TrimPrefix(anchor, "#")
	}
	return target
}

func templateControlAccountHref(accountID string) string {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return webAccountingPath + "#control-accounts"
	}
	return webAccountingControlsPath + "/" + url.PathEscape(accountID)
}

func templateTaxSummaryHref(taxCode string) string {
	taxCode = strings.TrimSpace(taxCode)
	if taxCode == "" {
		return webAccountingPath + "#tax-summaries"
	}
	return webAccountingTaxesPath + "/" + url.PathEscape(taxCode)
}

func templateApprovalReviewHref(approvalID string) string {
	approvalID = strings.TrimSpace(approvalID)
	if approvalID == "" {
		return webApprovalsPath
	}
	return webApprovalDetailPrefix + url.PathEscape(approvalID)
}

func templateApprovalQueueHref(queueCode, status string) string {
	values := url.Values{}
	if strings.TrimSpace(queueCode) != "" {
		values.Set("queue_code", strings.TrimSpace(queueCode))
	}
	status = strings.ToLower(strings.TrimSpace(status))
	switch status {
	case "pending", "closed":
		values.Set("status", status)
	case "approved", "rejected":
		values.Set("status", "closed")
	}
	if encoded := values.Encode(); encoded != "" {
		return webApprovalsPath + "?" + encoded
	}
	return webApprovalsPath
}

func templateProposalReviewHref(recommendationID, status, requestReference string) string {
	values := url.Values{}
	if strings.TrimSpace(recommendationID) != "" {
		values.Set("recommendation_id", strings.TrimSpace(recommendationID))
	}
	if strings.TrimSpace(status) != "" {
		values.Set("status", strings.TrimSpace(status))
	}
	if strings.TrimSpace(requestReference) != "" {
		values.Set("request_reference", strings.TrimSpace(requestReference))
	}
	if encoded := values.Encode(); encoded != "" {
		return webProposalsPath + "?" + encoded
	}
	return webProposalsPath
}

func templateProposalDetailHref(recommendationID string) string {
	recommendationID = strings.TrimSpace(recommendationID)
	if recommendationID == "" {
		return webProposalsPath
	}
	return webProposalDetailPrefix + url.PathEscape(recommendationID)
}

func templateWorkOrderReviewHref(workOrderID, status, documentID string) string {
	values := url.Values{}
	if strings.TrimSpace(workOrderID) != "" {
		values.Set("work_order_id", strings.TrimSpace(workOrderID))
	}
	if strings.TrimSpace(status) != "" {
		values.Set("status", strings.TrimSpace(status))
	}
	if strings.TrimSpace(documentID) != "" {
		values.Set("document_id", strings.TrimSpace(documentID))
	}
	if encoded := values.Encode(); encoded != "" {
		return webWorkOrdersPath + "?" + encoded
	}
	return webWorkOrdersPath
}

func templateInventoryReviewHref(movementID, itemID, locationID, documentID, movementType string, onlyPendingAccounting, onlyPendingExecution bool, anchor string) string {
	values := url.Values{}
	if strings.TrimSpace(movementID) != "" {
		values.Set("movement_id", strings.TrimSpace(movementID))
	}
	if strings.TrimSpace(itemID) != "" {
		values.Set("item_id", strings.TrimSpace(itemID))
	}
	if strings.TrimSpace(locationID) != "" {
		values.Set("location_id", strings.TrimSpace(locationID))
	}
	if strings.TrimSpace(documentID) != "" {
		values.Set("document_id", strings.TrimSpace(documentID))
	}
	if strings.TrimSpace(movementType) != "" {
		values.Set("movement_type", strings.TrimSpace(movementType))
	}
	if onlyPendingAccounting {
		values.Set("only_pending_accounting", "true")
	}
	if onlyPendingExecution {
		values.Set("only_pending_execution", "true")
	}

	target := webInventoryPath
	if encoded := values.Encode(); encoded != "" {
		target += "?" + encoded
	}
	anchor = strings.TrimSpace(anchor)
	if anchor != "" {
		target += "#" + strings.TrimPrefix(anchor, "#")
	}
	return target
}

func templateInventoryMovementHref(movementID string) string {
	movementID = strings.TrimSpace(movementID)
	if movementID == "" {
		return webInventoryPath
	}
	return webInventoryDetailPrefix + url.PathEscape(movementID)
}

func templateInventoryItemHref(itemID string) string {
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return webInventoryPath + "#stock-balances"
	}
	return webInventoryItemsPath + "/" + url.PathEscape(itemID)
}

func templateInventoryLocationHref(locationID string) string {
	locationID = strings.TrimSpace(locationID)
	if locationID == "" {
		return webInventoryPath + "#stock-balances"
	}
	return webInventoryLocationsPath + "/" + url.PathEscape(locationID)
}

func templateAuditEventHref(eventID string) string {
	eventID = strings.TrimSpace(eventID)
	if eventID == "" {
		return webAuditPath
	}
	return webAuditDetailPrefix + url.PathEscape(eventID)
}

func templateAuditEntityHref(entityType, entityID string) string {
	entityType = strings.TrimSpace(entityType)
	entityID = strings.TrimSpace(entityID)
	if entityType == "" || entityID == "" {
		return ""
	}

	switch entityType {
	case "documents.document":
		return templateDocumentReviewHref(entityID)
	case "ai.inbound_request":
		return templateInboundRequestHref(entityID)
	case "workflow.approval":
		return templateApprovalReviewHref(entityID)
	case "ai.agent_recommendation":
		return templateProposalDetailHref(entityID)
	case "ai.agent_run":
		return templateInboundRequestSectionHref("run:"+entityID, templateAIRunSectionID(entityID))
	case "ai.agent_step", "ai.agent_run_step":
		return templateInboundRequestSectionHref("step:"+entityID, templateAIStepSectionID(entityID))
	case "ai.agent_delegation":
		return templateInboundRequestSectionHref("delegation:"+entityID, templateAIDelegationSectionID(entityID))
	case "accounting.journal_entry":
		return templateAccountingEntryHref(entityID)
	case "work_orders.work_order":
		return webWorkOrdersPath + "/" + url.PathEscape(entityID)
	case "inventory_ops.item":
		return templateInventoryItemHref(entityID)
	case "inventory_ops.location":
		return templateInventoryLocationHref(entityID)
	case "inventory_ops.movement":
		return templateInventoryMovementHref(entityID)
	default:
		return ""
	}
}

func templateAuditEntityLabel(entityType string) string {
	switch strings.TrimSpace(entityType) {
	case "documents.document":
		return "Open document"
	case "ai.inbound_request":
		return "Open inbound request detail"
	case "workflow.approval":
		return "Open approval review"
	case "ai.agent_recommendation":
		return "Open proposal review"
	case "ai.agent_run":
		return "Open inbound request execution detail"
	case "ai.agent_step", "ai.agent_run_step":
		return "Open inbound request step detail"
	case "ai.agent_delegation":
		return "Open inbound request delegation detail"
	case "accounting.journal_entry":
		return "Open journal entry"
	case "work_orders.work_order":
		return "Open work order"
	case "inventory_ops.item":
		return "Open inventory item review"
	case "inventory_ops.location":
		return "Open inventory location review"
	case "inventory_ops.movement":
		return "Open movement detail"
	default:
		return ""
	}
}

