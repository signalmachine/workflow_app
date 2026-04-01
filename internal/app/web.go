package app

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
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

var webAppTemplate = template.Must(template.New("app").Funcs(webTemplateFuncs()).Parse(webAppHTML))

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

const webAppHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>{{.Title}}</title>
  <style>
    @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap');
    :root {
      --bg: #f8fafc;
      --bg-muted: #f1f5f9;
      --panel: rgba(255, 255, 255, 0.85);
      --panel-strong: rgba(255, 255, 255, 1);
      --ink: #0f172a;
      --ink-soft: #334155;
      --muted: #64748b;
      --line: rgba(226, 232, 240, 0.8);
      --line-strong: #94a3b8;
      --accent: #3b82f6;
      --accent-strong: #2563eb;
      --accent-soft: #eff6ff;
      --accent-faint: rgba(239, 246, 255, 0.5);
      --warn: #c2410c;
      --warn-soft: #ffedd5;
      --bad: #b91c1c;
      --bad-soft: #fee2e2;
      --good: #15803d;
      --good-soft: #dcfce7;
      --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
      --shadow: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
      --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
      --radius-lg: 20px;
      --radius-md: 14px;
      --radius-sm: 10px;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: 'Inter', system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
      color: var(--ink);
      background: linear-gradient(135deg, #f0f4f8 0%, #e1e8f0 100%);
      min-height: 100vh;
      line-height: 1.6;
    }
    a {
      color: var(--accent);
      text-decoration: none;
    }
    a:hover { color: var(--accent-strong); }
    .shell {
      width: min(1240px, calc(100% - 28px));
      margin: 20px auto 48px;
    }
    .masthead, .panel {
      background: var(--panel);
      backdrop-filter: blur(16px);
      -webkit-backdrop-filter: blur(16px);
      border: 1px solid var(--line);
      box-shadow: var(--shadow);
      border-radius: var(--radius-lg);
      transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1), box-shadow 200ms cubic-bezier(0.16, 1, 0.3, 1);
    }
    .masthead {
      padding: 24px;
      margin-bottom: 18px;
    }
    .masthead h1 {
      margin: 0;
      font-size: clamp(1.9rem, 4vw, 3rem);
      line-height: 1;
      letter-spacing: -0.04em;
    }
    .masthead p, .meta {
      margin: 0;
      color: var(--muted);
    }
    .brand-row {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: start;
    }
    .brand-mark {
      display: inline-flex;
      align-items: center;
      gap: 10px;
      padding: 7px 11px;
      margin-bottom: 12px;
      border-radius: 999px;
      background: rgba(23, 92, 211, 0.06);
      border: 1px solid rgba(23, 92, 211, 0.12);
      color: var(--accent-strong);
      font-size: 0.76rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.1em;
    }
    .brand-mark::before {
      content: "";
      width: 10px;
      height: 10px;
      border-radius: 999px;
      background: linear-gradient(135deg, var(--accent), #4f8bff);
      box-shadow: 0 0 0 4px rgba(23, 92, 211, 0.08);
    }
    .masthead-copy {
      max-width: 760px;
      position: relative;
      z-index: 1;
    }
    .masthead-copy p {
      margin-top: 10px;
      max-width: 62ch;
      font-size: 1rem;
      color: var(--ink-soft);
    }
    .masthead-status {
      min-width: 240px;
      padding: 14px 16px;
      border-radius: 16px;
      background: rgba(245, 248, 252, 0.96);
      border: 1px solid var(--line);
    }
    .masthead-status strong {
      display: block;
      margin-bottom: 5px;
      font-size: 0.78rem;
      text-transform: uppercase;
      letter-spacing: 0.12em;
      color: var(--accent-strong);
    }
    .nav {
      margin-top: 18px;
      padding-top: 16px;
      border-top: 1px solid var(--line);
      display: flex;
      flex-wrap: wrap;
      gap: 14px;
      align-items: end;
      justify-content: space-between;
    }
    .nav-links {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      max-width: 100%;
      overflow-x: auto;
      padding-bottom: 2px;
    }
    .nav-links a {
      padding: 7px 11px;
      border-radius: 10px;
      color: var(--ink-soft);
      background: rgba(255, 255, 255, 0.86);
      border: 1px solid var(--line);
      transition: all 200ms cubic-bezier(0.16, 1, 0.3, 1);
      white-space: nowrap;
    }
    .nav-links a:hover,
    .nav-links a.active {
      color: var(--accent-strong);
      background: var(--accent-faint);
      border-color: rgba(59, 130, 246, 0.25);
      transform: translateY(-2px);
      box-shadow: var(--shadow-sm);
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
      gap: 16px;
      align-items: start;
    }
    .panel { padding: 20px; }
    .panel h2, .panel h3 {
      margin-top: 0;
      margin-bottom: 8px;
      font-size: 1.12rem;
      letter-spacing: -0.02em;
    }
    .panel > p.meta,
    .panel > .meta {
      margin-bottom: 18px;
    }
    .page-header {
      display: flex;
      flex-wrap: wrap;
      justify-content: space-between;
      align-items: start;
      gap: 16px;
      margin-bottom: 14px;
    }
    .page-header h2 {
      margin-bottom: 4px;
      font-size: clamp(1.45rem, 2vw, 1.9rem);
    }
    .eyebrow {
      margin-bottom: 8px;
      color: var(--accent-strong);
      font-size: 0.78rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.14em;
    }
    .page-actions {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      align-items: center;
    }
    .pill-link {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      padding: 7px 11px;
      border-radius: 10px;
      border: 1px solid var(--line);
      background: rgba(255, 255, 255, 0.9);
      color: var(--ink-soft);
      transition: all 200ms cubic-bezier(0.16, 1, 0.3, 1);
    }
    .pill-link:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow-sm);
      border-color: rgba(59, 130, 246, 0.25);
    }
    .hero-grid {
      display: grid;
      grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.9fr);
      gap: 16px;
      align-items: stretch;
    }
    .hero-card {
      padding: 18px;
      border-radius: 16px;
      border: 1px solid var(--line);
      background: #fff;
      box-shadow: none;
    }
    .hero-card h3 {
      margin-bottom: 6px;
      font-size: 1.12rem;
    }
    .hero-card p {
      margin: 0;
      color: var(--ink-soft);
    }
    .hero-metrics {
      margin-top: 14px;
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
      gap: 10px;
    }
    .metric-tile {
      padding: 12px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: #fff;
    }
    .metric-tile strong {
      display: block;
      font-size: 1.35rem;
      line-height: 1.1;
      margin-bottom: 6px;
    }
    .section-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
      gap: 16px;
    }
    .section-note {
      padding: 13px 14px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: #f6f9fc;
      color: var(--ink-soft);
    }
    .notice, .error {
      padding: 12px 14px;
      border-radius: var(--radius-sm);
      margin-bottom: 14px;
      border: 1px solid transparent;
      box-shadow: none;
    }
    .notice {
      background: var(--accent-soft);
      border-color: rgba(23, 92, 211, 0.14);
    }
    .error {
      background: var(--warn-soft);
      border-color: rgba(154,52,18,0.22);
      color: var(--warn);
    }
    form { display: grid; gap: 12px; }
    label {
      display: grid;
      gap: 6px;
      font-weight: 600;
      color: var(--ink-soft);
    }
    input, textarea, select, button {
      width: 100%;
      font: inherit;
      padding: 11px 13px;
      border-radius: var(--radius-sm);
      border: 1px solid var(--line-strong);
      background: rgba(255,255,255,0.96);
      color: var(--ink);
      transition: border-color 120ms ease, box-shadow 120ms ease, background 120ms ease;
    }
    input:focus, textarea:focus, select:focus {
      outline: none;
      border-color: rgba(23, 92, 211, 0.48);
      box-shadow: 0 0 0 4px rgba(23, 92, 211, 0.12);
    }
    textarea { min-height: 132px; resize: vertical; }
    button {
      width: auto;
      min-width: 160px;
      background: linear-gradient(180deg, var(--accent) 0%, var(--accent-strong) 100%);
      color: #fff;
      border: 1px solid rgba(19, 73, 165, 0.22);
      cursor: pointer;
      font-weight: 700;
      box-shadow: 0 4px 14px rgba(59, 130, 246, 0.4);
      transition: all 200ms cubic-bezier(0.16, 1, 0.3, 1);
    }
    button:hover {
      transform: translateY(-2px);
      box-shadow: 0 6px 20px rgba(59, 130, 246, 0.5);
    }
    button.secondary {
      background: linear-gradient(180deg, #475467 0%, #344054 100%);
      border-color: rgba(52, 64, 84, 0.24);
      box-shadow: 0 4px 14px rgba(52, 64, 84, 0.3);
    }
    button.ghost {
      background: rgba(255,255,255,0.82);
      color: var(--ink-soft);
      border: 1px solid var(--line);
      box-shadow: var(--shadow-sm);
    }
    table {
      width: 100%;
      border-collapse: collapse;
      font-size: 0.96rem;
      min-width: 0;
    }
    th, td {
      text-align: left;
      padding: 12px 10px;
      border-top: 1px solid rgba(207, 218, 229, 0.8);
      vertical-align: top;
    }
    thead th { border-top: none; }
    th {
      color: var(--muted);
      font-size: 0.79rem;
      text-transform: uppercase;
      letter-spacing: 0.1em;
      background: rgba(237, 243, 248, 0.75);
    }
    tbody tr { transition: background 150ms ease, transform 150ms ease; }
    tbody tr:hover td { background: rgba(239, 246, 255, 0.7); }
    .hero-card, .metric-tile, .summary-card, .detail-card, .detail-block {
      transition: transform 200ms cubic-bezier(0.16, 1, 0.3, 1), box-shadow 200ms cubic-bezier(0.16, 1, 0.3, 1);
    }
    .hero-card:hover, .metric-tile:hover, .summary-card:hover {
      transform: translateY(-2px);
      box-shadow: var(--shadow);
    }
    .table-wrap {
      overflow-x: auto;
      margin: 0 -4px;
      padding: 0 4px;
    }
    .table-wrap > table {
      min-width: 640px;
    }
    .status-pill {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      padding: 5px 10px;
      border-radius: 999px;
      font-size: 0.82rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      white-space: nowrap;
    }
    .status-good { background: var(--good-soft); color: var(--good); }
    .status-bad { background: var(--bad-soft); color: var(--bad); }
    .status-neutral { background: var(--neutral-soft); color: var(--ink-soft); }
    pre {
      margin: 0;
      white-space: pre-wrap;
      word-break: break-word;
      background: #f5f8fb;
      border: 1px solid var(--line);
      border-radius: var(--radius-sm);
      padding: 14px;
      overflow-x: auto;
      color: var(--ink-soft);
    }
    .split {
      display: grid;
      grid-template-columns: 1.2fr 0.8fr;
      gap: 16px;
    }
    .summary-list {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 12px;
    }
    .summary-card {
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 12px;
      background: #fff;
      box-shadow: none;
    }
    .summary-card strong {
      display: block;
      font-size: 1.45rem;
      margin-bottom: 5px;
    }
    .inline-form {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
      align-items: center;
    }
    .inline-form > * {
      flex: 1 1 180px;
    }
    .inline-form button,
    .inline-form a {
      flex: 0 0 auto;
    }
    .stack { display: grid; gap: 20px; }
    .detail-block + .detail-block { margin-top: 16px; }
    .detail-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 14px;
      margin-top: 16px;
    }
    .detail-card,
    .detail-block {
      padding: 13px 14px;
      border-radius: 12px;
      border: 1px solid var(--line);
      background: #fff;
    }
    .detail-card strong,
    .detail-block strong {
      color: var(--ink-soft);
    }
    .card-stack {
      display: grid;
      gap: 14px;
    }
    .empty-state {
      padding: 16px;
      border: 1px dashed var(--line-strong);
      border-radius: 12px;
      color: var(--muted);
      background: rgba(255,255,255,0.7);
    }
    .filter-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
      gap: 12px;
      align-items: end;
    }
    .filter-grid .filter-actions {
      display: flex;
      align-items: end;
      gap: 10px;
      flex-wrap: wrap;
    }
    .login-panel {
      max-width: 620px;
      margin: 0 auto;
    }
    .login-panel form {
      margin-top: 18px;
    }
    .login-panel .section-note {
      margin-top: 16px;
    }
    @media (max-width: 880px) {
      .split { grid-template-columns: 1fr; }
      .hero-grid { grid-template-columns: 1fr; }
      .brand-row { flex-direction: column; }
      .masthead-status { min-width: 0; width: 100%; }
      .shell { width: min(100% - 20px, 1240px); }
      .masthead, .panel { border-radius: 18px; }
    }
  </style>
</head>
<body>
  <div class="shell">
    <section class="masthead">
      <div class="brand-row">
        <div class="masthead-copy">
          <div class="brand-mark">Control surface</div>
          <h1>workflow_app</h1>
          <p>AI-agent-first intake, review, approvals, and operator control on one shared browser seam with durable request continuity.</p>
        </div>
        <div class="masthead-status">
          <strong>Current posture</strong>
          <div class="meta">Browser control for intake, proposals, approvals, and review continuity.</div>
        </div>
      </div>
      {{if .Session}}
      <div class="nav">
        <div>
          <div class="meta">Signed in as {{.Session.UserEmail}} in {{.Session.OrgName}} ({{.Session.RoleCode}})</div>
          <div class="nav-links" style="margin-top:12px;">
            <a href="/app" {{if eq .ActivePath "/app"}}class="active"{{end}}>Home</a>
            <a href="/app/operations-feed" {{if eq .ActivePath "/app/operations-feed"}}class="active"{{end}}>Operations feed</a>
            <a href="/app/agent-chat" {{if eq .ActivePath "/app/agent-chat"}}class="active"{{end}}>Agent chat</a>
            <a href="/app/submit-inbound-request" {{if eq .ActivePath "/app/submit-inbound-request"}}class="active"{{end}}>Submit request</a>
            <a href="/app/review/inbound-requests" {{if eq .ActivePath "/app/review/inbound-requests"}}class="active"{{end}}>Inbound requests</a>
            <a href="/app/review/documents" {{if eq .ActivePath "/app/review/documents"}}class="active"{{end}}>Documents</a>
            <a href="/app/review/accounting" {{if eq .ActivePath "/app/review/accounting"}}class="active"{{end}}>Accounting</a>
            <a href="/app/review/approvals" {{if eq .ActivePath "/app/review/approvals"}}class="active"{{end}}>Approvals</a>
            <a href="/app/review/proposals" {{if eq .ActivePath "/app/review/proposals"}}class="active"{{end}}>Proposals</a>
            <a href="/app/review/inventory" {{if eq .ActivePath "/app/review/inventory"}}class="active"{{end}}>Inventory</a>
            <a href="/app/review/work-orders" {{if eq .ActivePath "/app/review/work-orders"}}class="active"{{end}}>Work orders</a>
            <a href="/app/review/audit" {{if eq .ActivePath "/app/review/audit"}}class="active"{{end}}>Audit</a>
          </div>
        </div>
        <form method="post" action="/app/logout" style="display:inline-grid;">
          <button type="submit" class="secondary">Sign out</button>
        </form>
      </div>
      {{end}}
    </section>

    {{if .ShowLogin}}
    <section class="panel login-panel">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <div class="page-header">
        <div>
          <div class="eyebrow">Browser session</div>
          <h2>Sign in</h2>
          <p class="meta">Use the shared operator account to enter the dashboard, review requests, and supervise the coordinator queue.</p>
        </div>
      </div>
      <form method="post" action="{{.LoginPath}}">
        <label>Org slug
          <input type="text" name="org_slug" autocomplete="organization" required>
        </label>
        <label>User email
          <input type="email" name="email" autocomplete="email" required>
        </label>
        <label>Password
          <input type="password" name="password" autocomplete="current-password" required>
        </label>
        <label>Device label
          <input type="text" name="device_label" value="browser">
        </label>
        <button type="submit">Start browser session</button>
      </form>
      <div class="section-note">Thin v1 keeps one shared backend seam for the browser and later non-browser clients. This sign-in surface only issues the browser-session path.</div>
    </section>
    {{end}}

    {{with .Dashboard}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="hero-grid">
          <div class="hero-card">
            <div class="eyebrow">Operations home</div>
            <h3>Review queue posture, launch the next run, and keep request continuity visible.</h3>
            <p>Home stays dashboard-first: use the dedicated submission page for new intake, then return here for current status, approvals, and downstream review continuity.</p>
            <div class="hero-metrics">
              <div class="metric-tile">
                <strong>{{len .InboundSummary}}</strong>
                <div class="meta">Tracked request states</div>
              </div>
              <div class="metric-tile">
                <strong>{{len .InboundRequests}}</strong>
                <div class="meta">Recent inbound requests</div>
              </div>
              <div class="metric-tile">
                <strong>{{len .Approvals}}</strong>
                <div class="meta">Pending approvals shown</div>
              </div>
            </div>
          </div>
          <div class="card-stack">
            <div class="hero-card">
              <div class="eyebrow">Quick links</div>
              <h3>Move directly into intake or review</h3>
              <p>Start new persisted intake on its own page, then continue through request detail, proposal review, approvals, and downstream inspection without mixing those jobs back into the dashboard.</p>
              <div class="page-actions" style="margin-top:16px;">
                <a href="/app/operations-feed" class="pill-link">Open operations feed</a>
                <a href="/app/agent-chat" class="pill-link">Open agent chat</a>
                <a href="/app/submit-inbound-request" class="pill-link">Open submission page</a>
                <a href="/app/review/inbound-requests" class="pill-link">Open inbound review</a>
              </div>
            </div>
            <div class="hero-card">
              <div class="eyebrow">Coordinator queue</div>
              <h3>Run the next queued request</h3>
              <p>Process one queued request through the provider-backed coordinator using the same backend seam exercised by the API.</p>
              <form method="post" action="/app/agent/process-next-queued-inbound-request" style="margin-top:16px;">
                <button type="submit">Process next queued request</button>
              </form>
            </div>
            <div class="section-note">
              Use the submission page for new intake and the review surfaces below to jump from draft, queued, failed, or processed requests without losing the request reference.
            </div>
          </div>
        </div>
      </section>

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Starting points</div>
            <h2>Operator starting points</h2>
            <p class="meta">Start from parked, in-flight, failed, or cancelled request states without reopening broad review pages first.</p>
          </div>
        </div>
        <div class="summary-list">
          {{range .InboundSummary}}
          <div class="summary-card">
            <strong>{{.RequestCount}}</strong>
            <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
            <div class="meta">Messages: {{.MessageCount}} | Attachments: {{.AttachmentCount}}</div>
            <div class="meta">{{dashboardStatusBlurb .Status}}</div>
            <div class="meta">Updated: {{formatTime .LatestUpdatedAt}}</div>
            <div class="meta"><a href="{{inboundRequestsHref .Status ""}}">{{dashboardStatusAction .Status}}</a></div>
          </div>
          {{else}}
          <div class="summary-card">No inbound requests yet.</div>
          {{end}}
        </div>
      </section>

      <div class="grid">
        <section class="panel">
          <div class="page-header">
            <div>
              <div class="eyebrow">Recent intake</div>
              <h2>Recent inbound requests</h2>
            </div>
            <div class="page-actions">
              <a href="/app/review/inbound-requests" class="pill-link">Open full inbound-request review</a>
            </div>
          </div>
          <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Reference</th>
                <th>Status</th>
                <th>Channel</th>
                <th>Messages</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {{range .InboundRequests}}
              <tr>
                <td><a href="{{inboundRequestHref .RequestReference}}">{{.RequestReference}}</a></td>
                <td>
                  <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
                  {{if .CancelledAt.Valid}}<div class="meta">Cancelled: {{formatTime .CancelledAt.Time}}</div>{{end}}
                  {{if .CancellationReason}}<div class="meta">{{.CancellationReason}}</div>{{end}}
                  {{if .FailedAt.Valid}}<div class="meta">Failed: {{formatTime .FailedAt.Time}}</div>{{end}}
                  {{if .FailureReason}}<div class="meta">{{.FailureReason}}</div>{{end}}
                </td>
                <td>{{.Channel}}</td>
                <td>
                  {{.MessageCount}} messages / {{.AttachmentCount}} files
                  <div class="meta"><a href="{{inboundRequestHref .RequestReference}}">{{dashboardRequestAction .Status}}</a></div>
                  {{if .LastRunID.Valid}}
                  <div class="meta"><a href="{{inboundSectionHref (printf "run:%s" .LastRunID.String) (runSectionID .LastRunID.String)}}">Open latest run</a></div>
                  {{end}}
                  {{if .LastRecommendationID.Valid}}
                  <div class="meta"><a href="{{proposalDetailHref .LastRecommendationID.String}}">Open latest proposal</a></div>
                  {{end}}
                </td>
                <td>{{formatTime .UpdatedAt}}</td>
              </tr>
              {{else}}
              <tr><td colspan="5">No inbound requests available.</td></tr>
              {{end}}
            </tbody>
          </table>
          </div>
        </section>

        <section class="panel">
          <div class="page-header">
            <div>
              <div class="eyebrow">Control boundary</div>
              <h2>Pending approvals</h2>
            </div>
            <div class="page-actions">
              <a href="/app/review/approvals?status=pending" class="pill-link">Open full approval review</a>
            </div>
          </div>
          <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Queue</th>
                <th>Document</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {{range .Approvals}}
              <tr>
                <td>
                  <a href="{{approvalQueueHref .QueueCode .QueueStatus}}">{{.QueueCode}}</a>
                  <div class="meta"><a href="{{approvalReviewHref .ApprovalID}}">Open exact approval</a></div>
                </td>
                <td>
                  <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a>
                  <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a></div>
                </td>
                <td>
                  <div class="status-pill {{statusClass .ApprovalStatus}}">{{.ApprovalStatus}}</div>
                  <form method="post" action="/app/approvals/{{.ApprovalID}}/decision" style="margin-top:8px;">
                    <input type="hidden" name="return_to" value="/app">
                    <input type="text" name="decision_note" placeholder="Decision note">
                    <div class="inline-form">
                      <button type="submit" name="decision" value="approved">Approve</button>
                      <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
                    </div>
                  </form>
                </td>
              </tr>
              {{else}}
              <tr><td colspan="3">No pending approvals.</td></tr>
              {{end}}
            </tbody>
          </table>
          </div>
        </section>
      </div>

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Downstream review</div>
            <h2>Processed proposals</h2>
            <p class="meta">Watch the handoff from request intake into recommendation, approval, and document continuity.</p>
          </div>
          <div class="page-actions">
            <a href="/app/review/proposals" class="pill-link">Open full proposal review</a>
          </div>
        </div>
        <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Request</th>
              <th>Recommendation</th>
              <th>Approval</th>
              <th>Document</th>
            </tr>
          </thead>
          <tbody>
            {{range .Proposals}}
            <tr>
              <td><a href="{{inboundRequestHref .RequestReference}}">{{.RequestReference}}</a></td>
              <td>
                <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
                <div>{{.Summary}}</div>
                <div class="meta"><a href="{{proposalDetailHref .RecommendationID}}">Open exact proposal</a></div>
              </td>
              <td>
                {{if .ApprovalID.Valid}}
                <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>
                <div class="status-pill {{statusClass .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</div>
                {{else}}
                {{.ApprovalStatus.String}}
                {{end}}
              </td>
              <td>{{if .DocumentID.Valid}}<a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No processed proposals available.</td></tr>
            {{end}}
          </tbody>
        </table>
        </div>
      </section>
    </div>
    {{end}}

    {{with .OperationsFeed}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="hero-grid">
          <div class="hero-card">
            <div class="eyebrow">Durable operations feed</div>
            <h3>Review one-way coordinator and system updates without turning the application into a chat stream.</h3>
            <p>This page is event-driven and durable. Follow the linked request, proposal, approval, and document surfaces when action is required.</p>
          </div>
          <div class="card-stack">
            <div class="detail-card">
              <strong>Feed rule</strong>
              <div class="meta">The operations feed is one-way status and control visibility. Workflow actions still happen on the linked exact pages.</div>
            </div>
            <div class="detail-card">
              <strong>Conversation rule</strong>
              <div class="meta">Use the separate agent-chat surface for guidance-oriented messages that should enter the persisted intake queue.</div>
            </div>
          </div>
        </div>
      </section>

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Latest events</div>
            <h2>Operations feed</h2>
            <p class="meta">Recent request lifecycle changes, coordinator proposals, and approval-boundary events on one durable page.</p>
          </div>
          <div class="page-actions">
            <a href="/app/agent-chat" class="pill-link">Open agent chat</a>
            <a href="/app/review/inbound-requests" class="pill-link">Open inbound review</a>
          </div>
        </div>
        <div class="card-stack">
          {{range .Feed}}
          <div class="detail-card">
            <div class="page-header">
              <div>
                <div class="eyebrow">{{.Kind}}</div>
                <h3>{{.Title}}</h3>
                <p class="meta">{{formatTime .OccurredAt}}</p>
              </div>
              <div class="page-actions">
                <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
              </div>
            </div>
            <p class="meta">{{.Summary}}</p>
            <div class="page-actions">
              <a href="{{.PrimaryHref}}" class="pill-link">{{.PrimaryLabel}}</a>
              {{if .SecondaryHref}}<a href="{{.SecondaryHref}}" class="pill-link">{{.SecondaryLabel}}</a>{{end}}
            </div>
          </div>
          {{else}}
          <div class="empty-state">No durable operations updates are available yet.</div>
          {{end}}
        </div>
      </section>
    </div>
    {{end}}

    {{with .AgentChat}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="hero-grid">
          <div class="hero-card">
            <div class="eyebrow">Coordinator chat</div>
            <h3>Queue a guidance-oriented coordinator request without collapsing intake, feed, and review into one surface.</h3>
            <p>Agent chat stays on the same persisted inbound-request model. It is interactive in entry shape, but the response path remains queue-oriented and durable rather than an immediate transient assistant reply.</p>
          </div>
          <div class="card-stack">
            <div class="detail-card">
              <strong>Persistence rule</strong>
              <div class="meta">Each chat message becomes one exact request reference on the shared intake foundation.</div>
            </div>
            <div class="detail-card">
              <strong>Control rule</strong>
              <div class="meta">Chat does not bypass approvals, downstream review, or the request detail trace.</div>
            </div>
          </div>
        </div>
      </section>

      {{if .RequestReference}}
      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Queued conversation</div>
            <h2>{{.RequestReference}}</h2>
            <p class="meta">The coordinator conversation is now on the persisted request path in <span class="status-pill {{statusClass .RequestStatus}}">{{.RequestStatus}}</span> state.</p>
          </div>
        </div>
        <div class="page-actions">
          <a href="/app/inbound-requests/{{.RequestReference}}" class="pill-link">Open exact request detail</a>
          <a href="/app/operations-feed" class="pill-link">Open operations feed</a>
          <a href="/app/review/proposals?request_reference={{.RequestReference}}" class="pill-link">Open proposal continuity</a>
        </div>
      </section>
      {{end}}

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">New coordinator request</div>
            <h2>Start agent chat</h2>
            <p class="meta">Ask for clarification, guidance, or issue-oriented follow-up. Queue it as durable intake instead of expecting an immediate inline assistant answer.</p>
          </div>
        </div>
        <div class="split">
          <div>
            <form method="post" action="/app/inbound-requests" enctype="multipart/form-data">
              <input type="hidden" name="return_to" value="/app/agent-chat">
              <input type="hidden" name="channel" value="agent_chat">
              <label>Operator label
                <input type="text" name="submitter_label" placeholder="dispatch desk">
              </label>
              <label>Coordinator message
                <textarea name="message_text" required placeholder="Explain the issue, the question, and any workflow page or request reference that should stay in view."></textarea>
              </label>
              <label>Attachments
                <input type="file" name="attachments" multiple>
              </label>
              <div class="inline-form">
                <button type="submit" name="intent" value="queue">Queue coordinator request</button>
                <button type="submit" name="intent" value="save_draft" class="secondary">Save chat draft</button>
              </div>
            </form>
          </div>
          <div class="card-stack">
            <div class="detail-card">
              <strong>Good use</strong>
              <div class="meta">Clarification, request scoping, issue triage, and workflow guidance that should stay visible through exact request detail and later proposal review.</div>
            </div>
            <div class="detail-card">
              <strong>Do not use chat for</strong>
              <div class="meta">Bypassing approval decisions, treating the feed as a reply thread, or replacing the dedicated request-submission page for normal operational intake.</div>
            </div>
          </div>
        </div>
      </section>

      <div class="grid">
        <section class="panel">
          <div class="page-header">
            <div>
              <div class="eyebrow">Recent conversations</div>
              <h2>Recent agent-chat requests</h2>
            </div>
            <div class="page-actions">
              <a href="/app/review/inbound-requests" class="pill-link">Open full inbound review</a>
            </div>
          </div>
          <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Reference</th>
                <th>Status</th>
                <th>Messages</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {{range .RecentRequests}}
              <tr>
                <td><a href="{{inboundRequestHref .RequestReference}}">{{.RequestReference}}</a></td>
                <td>
                  <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
                  {{if .FailureReason}}<div class="meta">{{.FailureReason}}</div>{{end}}
                  {{if .CancellationReason}}<div class="meta">{{.CancellationReason}}</div>{{end}}
                </td>
                <td>
                  {{.MessageCount}} messages / {{.AttachmentCount}} files
                  {{if .LastRecommendationID.Valid}}<div class="meta"><a href="{{proposalDetailHref .LastRecommendationID.String}}">Open latest proposal</a></div>{{end}}
                </td>
                <td>{{formatTime .UpdatedAt}}</td>
              </tr>
              {{else}}
              <tr><td colspan="4">No recent agent-chat requests yet.</td></tr>
              {{end}}
            </tbody>
          </table>
          </div>
        </section>

        <section class="panel">
          <div class="page-header">
            <div>
              <div class="eyebrow">Proposal continuity</div>
              <h2>Recent agent-chat proposals</h2>
            </div>
            <div class="page-actions">
              <a href="/app/operations-feed" class="pill-link">Open operations feed</a>
            </div>
          </div>
          <div class="table-wrap">
          <table>
            <thead>
              <tr>
                <th>Request</th>
                <th>Recommendation</th>
                <th>Approval</th>
              </tr>
            </thead>
            <tbody>
              {{range .RecentProposals}}
              <tr>
                <td><a href="{{inboundRequestHref .RequestReference}}">{{.RequestReference}}</a></td>
                <td>
                  <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
                  <div>{{.Summary}}</div>
                  <div class="meta"><a href="{{proposalDetailHref .RecommendationID}}">Open proposal</a></div>
                </td>
                <td>
                  {{if .ApprovalID.Valid}}
                  <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>
                  {{else}}
                  -
                  {{end}}
                </td>
              </tr>
              {{else}}
              <tr><td colspan="3">No recent proposals from agent-chat requests yet.</td></tr>
              {{end}}
            </tbody>
          </table>
          </div>
        </section>
      </div>
    </div>
    {{end}}

    {{with .InboundSubmit}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="hero-grid">
          <div class="hero-card">
            <div class="eyebrow">Dedicated intake</div>
            <h3>Submit one inbound request without diluting the dashboard.</h3>
            <p>Capture the request message, supporting evidence, and operator label on a focused page that keeps queue-oriented intake separate from review and approval work.</p>
            <div class="page-actions" style="margin-top:16px;">
              <a href="/app" class="pill-link">Back to dashboard</a>
              <a href="/app/review/inbound-requests" class="pill-link">Open inbound review</a>
            </div>
          </div>
          <div class="card-stack">
            <div class="detail-card">
              <strong>Draft path</strong>
              <div class="meta">Save a parked request first when the operator still expects more detail, attachments, or amendment before queueing.</div>
            </div>
            <div class="detail-card">
              <strong>Queued path</strong>
              <div class="meta">Queue the request when the coordinator should pick it up as durable intake truth rather than an immediate chat response.</div>
            </div>
            <div class="detail-card">
              <strong>Continuity rule</strong>
              <div class="meta">Keep the exact REQ-... reference visible across detail, proposal, approval, and downstream review surfaces.</div>
            </div>
          </div>
        </div>
      </section>

      {{if .RequestReference}}
      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Submission result</div>
            <h2>{{.RequestReference}}</h2>
            <p class="meta">The request is now in <span class="status-pill {{statusClass .RequestStatus}}">{{.RequestStatus}}</span> state on the persisted intake path.</p>
          </div>
        </div>
        <div class="page-actions">
          <a href="/app/inbound-requests/{{.RequestReference}}" class="pill-link">Open exact request detail</a>
          <a href="/app" class="pill-link">Return to dashboard</a>
          <a href="/app/review/inbound-requests?request_reference={{.RequestReference}}" class="pill-link">Open request review</a>
        </div>
      </section>
      {{end}}

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Request intake</div>
            <h2>Submit inbound request</h2>
            <p class="meta">Capture the request, supporting evidence, and the exact message that should enter the queue.</p>
          </div>
        </div>
        <div class="split">
          <div>
            <form method="post" action="/app/inbound-requests" enctype="multipart/form-data">
              <input type="hidden" name="return_to" value="/app/submit-inbound-request">
              <input type="hidden" name="channel" value="browser">
              <label>Submitter label
                <input type="text" name="submitter_label" placeholder="front desk">
              </label>
              <label>Request message
                <textarea name="message_text" required placeholder="Describe the request, evidence, and expected follow-up."></textarea>
              </label>
              <label>Attachments
                <input type="file" name="attachments" multiple>
              </label>
              <div class="inline-form">
                <button type="submit" name="intent" value="queue">Queue inbound request</button>
                <button type="submit" name="intent" value="save_draft" class="secondary">Save draft</button>
              </div>
            </form>
          </div>
          <div class="card-stack">
            <div class="detail-card">
              <strong>What success looks like</strong>
              <div class="meta">A successful submit or draft save returns the exact REQ-... reference here so the operator can continue into detail or review without guessing.</div>
            </div>
            <div class="detail-card">
              <strong>Workflow boundary</strong>
              <div class="meta">This page persists intake truth only. Later approval, proposal, and execution actions still happen on their dedicated workflow surfaces.</div>
            </div>
          </div>
        </div>
      </section>
    </div>
    {{end}}

    {{with .InboundRequests}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Inbound review</div>
            <h2>Inbound-request review</h2>
            <p class="meta">Filter by request state or exact REQ-... continuity when operators need the intake queue without broad dashboard context.</p>
          </div>
        </div>
        <form method="get" action="/app/review/inbound-requests" class="filter-grid">
          <label>Status
            <input type="text" name="status" value="{{.Status}}" placeholder="queued">
          </label>
          <label>Request reference
            <input type="text" name="request_reference" value="{{.RequestReference}}" placeholder="REQ-... reference">
          </label>
          <div class="filter-actions">
            <button type="submit">Filter requests</button>
            <a href="/app/review/inbound-requests" class="pill-link">Clear filters</a>
          </div>
        </form>
      </section>

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Queue posture</div>
            <h2>Request status summary</h2>
          </div>
        </div>
        <div class="summary-list">
          {{range .StatusSummary}}
          <div class="summary-card">
            <strong>{{.RequestCount}}</strong>
            <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
            <div class="meta">Messages: {{.MessageCount}} | Attachments: {{.AttachmentCount}}</div>
            <div class="meta">Updated: {{formatTime .LatestUpdatedAt}}</div>
            <div class="meta"><a href="{{inboundRequestsHref .Status ""}}">Open {{.Status}}</a></div>
          </div>
          {{else}}
          <div class="summary-card">No inbound requests yet.</div>
          {{end}}
        </div>
      </section>

      <section class="panel">
        <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Reference</th>
              <th>Status</th>
              <th>Channel</th>
              <th>Messages</th>
              <th>AI</th>
              <th>Updated</th>
            </tr>
          </thead>
          <tbody>
            {{range .Requests}}
            <tr>
              <td>
                <a href="{{inboundRequestHref .RequestReference}}">{{.RequestReference}}</a>
                <div class="meta">{{.RequestID}}</div>
                {{if eq .Status "draft"}}<div class="meta"><a href="{{inboundRequestHref .RequestReference}}">Continue draft</a></div>{{end}}
              </td>
              <td>
                <span class="status-pill {{statusClass .Status}}">{{.Status}}</span>
                {{if .CancelledAt.Valid}}<div class="meta">Cancelled: {{formatTime .CancelledAt.Time}}</div>{{end}}
                {{if .CancellationReason}}<div class="meta">{{.CancellationReason}}</div>{{end}}
                {{if .FailedAt.Valid}}<div class="meta">Failed: {{formatTime .FailedAt.Time}}</div>{{end}}
                {{if .FailureReason}}<div class="meta">{{.FailureReason}}</div>{{end}}
                {{if or (eq .Status "queued") (eq .Status "cancelled")}}<div class="meta"><a href="{{inboundRequestHref .RequestReference}}">Manage lifecycle</a></div>{{end}}
              </td>
              <td>{{.Channel}}<div class="meta">{{.OriginType}}</div></td>
              <td>{{.MessageCount}} messages / {{.AttachmentCount}} files</td>
              <td>
                {{if .LastRunID.Valid}}
                <div><a href="{{inboundSectionHref (printf "run:%s" .LastRunID.String) (runSectionID .LastRunID.String)}}"><span class="status-pill {{statusClass .LastRunStatus.String}}">{{.LastRunStatus.String}}</span></a></div>
                {{else}}
                -
                {{end}}
                {{if .LastRecommendationStatus.Valid}}
                <div class="meta">
                  {{if .LastRecommendationID.Valid}}
                  <a href="{{proposalDetailHref .LastRecommendationID.String}}">{{.LastRecommendationStatus.String}}</a>
                  {{else}}
                  {{.LastRecommendationStatus.String}}
                  {{end}}
                </div>
                {{end}}
              </td>
              <td>{{formatTime .UpdatedAt}}</td>
            </tr>
            {{else}}
            <tr><td colspan="6">No inbound requests available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
        </div>
      </section>
    </div>
    {{end}}

    {{with .Approvals}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Approval review</h2>
        <form method="get" action="/app/review/approvals" class="inline-form">
          <input type="text" name="approval_id" value="{{.ApprovalID}}" placeholder="approval id">
          <input type="text" name="status" value="{{.Status}}" placeholder="pending or closed">
          <input type="text" name="queue_code" value="{{.QueueCode}}" placeholder="queue code">
          <button type="submit">Filter approvals</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Queue</th>
              <th>Document</th>
              <th>Approval</th>
              <th>Posting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Approvals}}
            <tr>
              <td>
                <a href="{{approvalQueueHref .QueueCode .QueueStatus}}">{{.QueueCode}}</a>
                <div class="meta">Enqueued: {{formatTime .EnqueuedAt}}</div>
                <div class="meta"><span class="status-pill {{statusClass .QueueStatus}}">{{.QueueStatus}}</span></div>
                {{if .RequestReference.Valid}}
                <div class="meta">
                  <a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>
                  {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">Proposal</a>{{end}}
                  {{if .RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a>{{end}}
                </div>
                {{end}}
              </td>
              <td>
                <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a>
                <div class="meta">{{.DocumentTypeCode}} | <span class="status-pill {{statusClass .DocumentStatus}}">{{.DocumentStatus}}</span></div>
                <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a></div>
              </td>
              <td>
                <div class="status-pill {{statusClass .ApprovalStatus}}">{{.ApprovalStatus}}</div>
                <div class="meta">Requested: {{formatTime .RequestedAt}}</div>
                {{if eq .QueueStatus "pending"}}
                <form method="post" action="/app/approvals/{{.ApprovalID}}/decision" style="margin-top:8px;">
                  <input type="hidden" name="return_to" value="{{approvalQueueHref $.Approvals.QueueCode $.Approvals.Status}}">
                  <input type="text" name="decision_note" placeholder="Decision note">
                  <div class="inline-form">
                    <button type="submit" name="decision" value="approved">Approve</button>
                    <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
                  </div>
                </form>
                {{else}}
                <div class="meta">Closed: {{if .ClosedAt.Valid}}{{formatTime .ClosedAt.Time}}{{else}}-{{end}}</div>
                {{end}}
              </td>
              <td>
                {{if .JournalEntryID.Valid}}
                <a href="{{accountingEntryHref .JournalEntryID.String}}">Entry #{{.JournalEntryNumber.Int64}}</a>
                <div class="meta">{{if .JournalEntryPostedAt.Valid}}{{formatTime .JournalEntryPostedAt.Time}}{{else}}Not posted{{end}}</div>
                {{else if .JournalEntryNumber.Valid}}
                <a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>
                <div class="meta">{{if .JournalEntryPostedAt.Valid}}{{formatTime .JournalEntryPostedAt.Time}}{{else}}Not posted{{end}}</div>
                {{else}}
                -
                {{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No approval queue rows available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .ApprovalDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Approval {{.Entry.ApprovalID}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Entry.ApprovalStatus}}">{{.Entry.ApprovalStatus}}</span>
          <p><strong>{{.Entry.DocumentTitle}}</strong></p>
          <p class="meta">{{.Entry.QueueCode}} | queue {{.Entry.QueueStatus}} | requested {{formatTime .Entry.RequestedAt}}</p>
          <p class="meta">
            <a href="{{approvalQueueHref .Entry.QueueCode .Entry.QueueStatus}}">Filtered queue view</a> |
            <a href="{{documentReviewHref .Entry.DocumentID}}">Open document</a> |
            <a href="/app/review/audit?entity_type=workflow.approval&amp;entity_id={{.Entry.ApprovalID}}">Audit trail</a>
            {{if .Entry.RequestReference.Valid}} | <a href="{{inboundRequestHref .Entry.RequestReference.String}}">{{.Entry.RequestReference.String}}</a>{{end}}
            {{if .Entry.RecommendationID.Valid}} | <a href="{{proposalDetailHref .Entry.RecommendationID.String}}">Proposal</a>{{end}}
            {{if .Entry.RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .Entry.RunID.String) (runSectionID .Entry.RunID.String)}}">AI run</a>{{end}}
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Decision</h2>
          <table>
            <tbody>
              <tr><th>Queue status</th><td><span class="status-pill {{statusClass .Entry.QueueStatus}}">{{.Entry.QueueStatus}}</span></td></tr>
              <tr><th>Requested by</th><td>{{.Entry.RequestedByUserID}}</td></tr>
              <tr><th>Requested at</th><td>{{formatTime .Entry.RequestedAt}}</td></tr>
              <tr><th>Decided by</th><td>{{if .Entry.DecidedByUserID.Valid}}{{.Entry.DecidedByUserID.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Decided at</th><td>{{if .Entry.DecidedAt.Valid}}{{formatTime .Entry.DecidedAt.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
          {{if eq .Entry.QueueStatus "pending"}}
          <form method="post" action="/app/approvals/{{.Entry.ApprovalID}}/decision" style="margin-top:12px;">
            <input type="hidden" name="return_to" value="{{approvalReviewHref .Entry.ApprovalID}}">
            <input type="text" name="decision_note" placeholder="Decision note">
            <div class="inline-form">
              <button type="submit" name="decision" value="approved">Approve</button>
              <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
            </div>
          </form>
          {{end}}
        </section>
        <section class="panel">
          <h2>Linked record</h2>
          <table>
            <tbody>
              <tr><th>Document</th><td><a href="{{documentReviewHref .Entry.DocumentID}}">{{.Entry.DocumentTitle}}</a></td></tr>
              <tr><th>Type</th><td>{{.Entry.DocumentTypeCode}}</td></tr>
              <tr><th>Status</th><td><span class="status-pill {{statusClass .Entry.DocumentStatus}}">{{.Entry.DocumentStatus}}</span></td></tr>
              <tr><th>Document number</th><td>{{if .Entry.DocumentNumber.Valid}}{{.Entry.DocumentNumber.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Request</th><td>{{if .Entry.RequestReference.Valid}}<a href="{{inboundRequestHref .Entry.RequestReference.String}}">{{.Entry.RequestReference.String}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>Proposal</th><td>{{if .Entry.RecommendationID.Valid}}<a href="{{proposalDetailHref .Entry.RecommendationID.String}}">{{if .Entry.RecommendationStatus.Valid}}{{.Entry.RecommendationStatus.String}}{{else}}proposal{{end}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>AI run</th><td>{{if .Entry.RunID.Valid}}<a href="{{inboundSectionHref (printf "run:%s" .Entry.RunID.String) (runSectionID .Entry.RunID.String)}}">{{.Entry.RunID.String}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>Posting</th><td>{{if .Entry.JournalEntryID.Valid}}<a href="{{accountingEntryHref .Entry.JournalEntryID.String}}">Entry #{{.Entry.JournalEntryNumber.Int64}}</a>{{else if .Entry.JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.Entry.DocumentID}}">Entry #{{.Entry.JournalEntryNumber.Int64}}</a>{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Proposals}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Downstream review</div>
            <h2>Proposal review</h2>
            <p class="meta">Track the coordinator handoff from exact inbound request reference into recommendation, approval, and draft document continuity.</p>
          </div>
        </div>
        <form method="get" action="/app/review/proposals" class="filter-grid">
          <label>Recommendation ID
            <input type="text" name="recommendation_id" value="{{.RecommendationID}}" placeholder="recommendation id">
          </label>
          <label>Status
            <input type="text" name="status" value="{{.Status}}" placeholder="approval_requested">
          </label>
          <label>Request reference
            <input type="text" name="request_reference" value="{{.RequestReference}}" placeholder="REQ-... reference">
          </label>
          <div class="filter-actions">
            <button type="submit">Filter proposals</button>
            <a href="/app/review/proposals" class="pill-link">Clear filters</a>
          </div>
        </form>
      </section>
      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Proposal posture</div>
            <h2>Proposal status summary</h2>
          </div>
        </div>
        <div class="summary-list">
          {{range .StatusSummary}}
          <div class="summary-card">
            <strong>{{.ProposalCount}}</strong>
            <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
            <div class="meta">Requests: {{.RequestCount}} | Documents: {{.DocumentCount}}</div>
            <div class="meta">Updated: {{formatTime .LatestCreatedAt}}</div>
            <div class="meta"><a href="{{proposalReviewHref "" .RecommendationStatus ""}}">Open {{.RecommendationStatus}}</a></div>
          </div>
          {{else}}
          <div class="summary-card">No processed proposals yet.</div>
          {{end}}
        </div>
      </section>
      <section class="panel">
        <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Request</th>
              <th>Recommendation</th>
              <th>Approval</th>
              <th>Document</th>
            </tr>
          </thead>
          <tbody>
            {{range .ProcessedProposals}}
            <tr>
              <td>
                <a href="/app/inbound-requests/{{.RequestReference}}">{{.RequestReference}}</a>
                <div class="meta"><span class="status-pill {{statusClass .RequestStatus}}">{{.RequestStatus}}</span></div>
              </td>
              <td>
                <span class="status-pill {{statusClass .RecommendationStatus}}">{{.RecommendationStatus}}</span>
                <div>{{.Summary}}</div>
                <div class="meta">Created: {{formatTime .CreatedAt}}</div>
                <div class="meta"><a href="{{proposalDetailHref .RecommendationID}}">Open exact proposal</a></div>
              </td>
              <td>
                {{if .ApprovalID.Valid}}
                <div><a href="{{approvalQueueHref .ApprovalQueueCode.String .ApprovalStatus.String}}">{{.ApprovalQueueCode.String}}</a></div>
                <div class="status-pill {{statusClass .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</div>
                <div class="meta"><a href="{{approvalReviewHref .ApprovalID.String}}">Open exact approval</a></div>
                {{else}}
                -
                {{end}}
              </td>
              <td>
                {{if .DocumentID.Valid}}
                <a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>
                <div class="meta">{{.DocumentTypeCode.String}} | <span class="status-pill {{statusClass .DocumentStatus.String}}">{{.DocumentStatus.String}}</span></div>
                {{else}}
                -
                {{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No processed proposals available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
        </div>
      </section>
    </div>
    {{end}}

    {{with .ProposalDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Proposal {{.Review.RecommendationID}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.RecommendationStatus}}">{{.Review.RecommendationStatus}}</span>
          <p><strong>{{.Review.Summary}}</strong></p>
          <p class="meta">Request <a href="/app/inbound-requests/{{.Review.RequestReference}}">{{.Review.RequestReference}}</a> | run {{.Review.RunID}} | created {{formatTime .Review.CreatedAt}}</p>
          <p class="meta">
            <a href="{{proposalReviewHref .Review.RecommendationID .Review.RecommendationStatus .Review.RequestReference}}">Filtered proposal view</a> |
            <a href="/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id={{.Review.RecommendationID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Control chain</h2>
          <table>
            <tbody>
              <tr><th>Request</th><td><a href="/app/inbound-requests/{{.Review.RequestReference}}">{{.Review.RequestReference}}</a> | <span class="status-pill {{statusClass .Review.RequestStatus}}">{{.Review.RequestStatus}}</span></td></tr>
              <tr><th>Approval</th><td>{{if .Review.ApprovalID.Valid}}<a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{if .Review.ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .Review.ApprovalStatus.String}}">{{.Review.ApprovalStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
              <tr><th>Document</th><td>{{if .Review.DocumentID.Valid}}<a href="{{documentReviewHref .Review.DocumentID.String}}">{{.Review.DocumentTitle.String}}</a>{{if .Review.DocumentStatus.Valid}} | <span class="status-pill {{statusClass .Review.DocumentStatus.String}}">{{.Review.DocumentStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
              <tr><th>Suggested queue</th><td>{{if .Review.SuggestedQueueCode.Valid}}{{.Review.SuggestedQueueCode.String}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Identifiers</h2>
          <table>
            <tbody>
              <tr><th>Recommendation</th><td>{{.Review.RecommendationID}}</td></tr>
              <tr><th>Run</th><td>{{.Review.RunID}}</td></tr>
              <tr><th>Type</th><td>{{.Review.RecommendationType}}</td></tr>
              <tr><th>Document number</th><td>{{if .Review.DocumentNumber.Valid}}{{.Review.DocumentNumber.String}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
      {{if and (not .Review.ApprovalID.Valid) .Review.DocumentID.Valid}}
      <section class="panel">
        <h2>Request approval</h2>
        <form method="post" action="/app/review/proposals/{{.Review.RecommendationID}}/request-approval">
          <input type="hidden" name="return_to" value="/app/review/proposals/{{.Review.RecommendationID}}">
          <input type="text" name="queue_code" value="{{.ApprovalQueueCodeDraft}}" placeholder="queue code">
          <input type="text" name="reason" value="{{.ApprovalReason}}" placeholder="reason">
          <div class="inline-form">
            <button type="submit">Request approval</button>
            <a href="{{documentReviewHref .Review.DocumentID.String}}">Open document</a>
          </div>
        </form>
      </section>
      {{end}}
    </div>
    {{end}}

    {{with .Documents}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Document review</h2>
        <form method="get" action="/app/review/documents" class="inline-form">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <input type="text" name="type_code" value="{{.TypeCode}}" placeholder="type code">
          <input type="text" name="status" value="{{.Status}}" placeholder="status">
          <button type="submit">Filter documents</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Type</th>
              <th>Title</th>
              <th>Status</th>
              <th>Approval</th>
              <th>Posting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Documents}}
            <tr>
              <td>{{.TypeCode}}</td>
              <td>
                <strong><a href="{{documentReviewHref .DocumentID}}">{{.Title}}</a></strong>
                <div class="meta">{{.DocumentID}}</div>
                <div class="meta">
                  <a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a>
                  {{if eq .TypeCode "work_order"}} | <a href="/app/review/work-orders?document_id={{.DocumentID}}">Execution review</a>{{end}}
                  {{if or (eq .TypeCode "inventory_receipt") (eq .TypeCode "inventory_issue") (eq .TypeCode "inventory_adjustment")}} | <a href="/app/review/inventory?document_id={{.DocumentID}}">Inventory review</a>{{end}}
                  {{if .RequestReference.Valid}} | <a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>{{end}}
                  {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">Proposal</a>{{end}}
                </div>
              </td>
              <td><span class="status-pill {{statusClass .Status}}">{{.Status}}</span></td>
              <td>
                {{if .ApprovalID.Valid}}
                <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>
                {{if .ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</span>{{end}}
                {{else if .ApprovalQueueCode.Valid}}
                <a href="{{approvalQueueHref .ApprovalQueueCode.String .ApprovalStatus.String}}">{{.ApprovalStatus.String}}</a>
                {{else}}
                {{.ApprovalStatus.String}}
                {{end}}
              </td>
              <td>{{if .JournalEntryID.Valid}}<a href="{{accountingEntryHref .JournalEntryID.String}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="5">No documents available for the selected filters.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .DocumentDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Document {{if .Review.NumberValue.Valid}}{{.Review.NumberValue.String}}{{else}}{{.Review.TypeCode}}{{end}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.Status}}">{{.Review.Status}}</span>
          <p><strong>{{.Review.Title}}</strong></p>
          <p class="meta">{{.Review.DocumentID}}</p>
          <p class="meta">Type: {{.Review.TypeCode}} | Created: {{formatTime .Review.CreatedAt}} | Updated: {{formatTime .Review.UpdatedAt}}</p>
          <p class="meta">
            <a href="/app/review/documents?document_id={{.Review.DocumentID}}">Filtered list view</a> |
            <a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.Review.DocumentID}}">Audit trail</a>
            {{if .Review.SourceDocumentID.Valid}} | <a href="{{documentReviewHref .Review.SourceDocumentID.String}}">Source document</a>{{end}}
            {{if .Review.RequestReference.Valid}} | <a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{end}}
            {{if .Review.RecommendationID.Valid}} | <a href="{{proposalDetailHref .Review.RecommendationID.String}}">Proposal</a>{{end}}
            {{if .Review.RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">AI run</a>{{end}}
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Control chain</h2>
          <table>
            <tbody>
              <tr>
                <th>Approval</th>
                <td>
                  {{if .Review.ApprovalID.Valid}}
                  <a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>
                  {{if .Review.ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .Review.ApprovalStatus.String}}">{{.Review.ApprovalStatus.String}}</span>{{end}}
                  {{else}}
                  -
                  {{end}}
                </td>
              </tr>
              <tr>
                <th>Request</th>
                <td>{{if .Review.RequestReference.Valid}}<a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{else}}-{{end}}</td>
              </tr>
              <tr>
                <th>Proposal</th>
                <td>{{if .Review.RecommendationID.Valid}}<a href="{{proposalDetailHref .Review.RecommendationID.String}}">{{if .Review.RecommendationStatus.Valid}}{{.Review.RecommendationStatus.String}}{{else}}proposal{{end}}</a>{{else}}-{{end}}</td>
              </tr>
              <tr>
                <th>AI run</th>
                <td>{{if .Review.RunID.Valid}}<a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">{{.Review.RunID.String}}</a>{{else}}-{{end}}</td>
              </tr>
              <tr>
                <th>Accounting</th>
                <td>
                  {{if .Review.JournalEntryID.Valid}}
                  <a href="{{accountingEntryHref .Review.JournalEntryID.String}}">Entry #{{.Review.JournalEntryNumber.Int64}}</a>
                  {{if .Review.JournalEntryPostedAt.Valid}} | {{formatTime .Review.JournalEntryPostedAt.Time}}{{end}}
                  {{else if .Review.JournalEntryNumber.Valid}}
                  <a href="/app/review/accounting?document_id={{.Review.DocumentID}}">Entry #{{.Review.JournalEntryNumber.Int64}}</a>
                  {{if .Review.JournalEntryPostedAt.Valid}} | {{formatTime .Review.JournalEntryPostedAt.Time}}{{end}}
                  {{else}}
                  -
                  {{end}}
                </td>
              </tr>
              <tr>
                <th>Execution</th>
                <td>{{if eq .Review.TypeCode "work_order"}}<a href="/app/review/work-orders?document_id={{.Review.DocumentID}}">Work-order review</a>{{else}}-{{end}}</td>
              </tr>
              <tr>
                <th>Inventory</th>
                <td>{{if or (eq .Review.TypeCode "inventory_receipt") (eq .Review.TypeCode "inventory_issue") (eq .Review.TypeCode "inventory_adjustment")}}<a href="/app/review/inventory?document_id={{.Review.DocumentID}}">Inventory review</a>{{else}}-{{end}}</td>
              </tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Lifecycle</h2>
          <table>
            <tbody>
              <tr><th>Created by</th><td>{{.Review.CreatedByUserID}}</td></tr>
              <tr><th>Submitted by</th><td>{{if .Review.SubmittedByUserID.Valid}}{{.Review.SubmittedByUserID.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Submitted at</th><td>{{if .Review.SubmittedAt.Valid}}{{formatTime .Review.SubmittedAt.Time}}{{else}}-{{end}}</td></tr>
              <tr><th>Approved at</th><td>{{if .Review.ApprovedAt.Valid}}{{formatTime .Review.ApprovedAt.Time}}{{else}}-{{end}}</td></tr>
              <tr><th>Rejected at</th><td>{{if .Review.RejectedAt.Valid}}{{formatTime .Review.RejectedAt.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Accounting}}
    {{$accounting := .}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Accounting review</h2>
        <form method="get" action="/app/review/accounting" class="inline-form">
          <input type="date" name="start_on" value="{{.StartOn}}">
          <input type="date" name="end_on" value="{{.EndOn}}">
          <input type="date" name="as_of" value="{{.AsOf}}">
          <input type="text" name="entry_id" value="{{.EntryID}}" placeholder="journal entry id">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="source document id">
          <select name="tax_type">
            <option value="">all tax types</option>
            <option value="gst" {{if eq .TaxType "gst"}}selected{{end}}>gst</option>
            <option value="tds" {{if eq .TaxType "tds"}}selected{{end}}>tds</option>
          </select>
          <input type="text" name="tax_code" value="{{.TaxCode}}" placeholder="tax code">
          <select name="control_type">
            <option value="">all control accounts</option>
            <option value="receivable" {{if eq .ControlType "receivable"}}selected{{end}}>receivable</option>
            <option value="payable" {{if eq .ControlType "payable"}}selected{{end}}>payable</option>
            <option value="gst_input" {{if eq .ControlType "gst_input"}}selected{{end}}>gst_input</option>
            <option value="gst_output" {{if eq .ControlType "gst_output"}}selected{{end}}>gst_output</option>
            <option value="tds_receivable" {{if eq .ControlType "tds_receivable"}}selected{{end}}>tds_receivable</option>
            <option value="tds_payable" {{if eq .ControlType "tds_payable"}}selected{{end}}>tds_payable</option>
          </select>
          <input type="text" name="account_id" value="{{.AccountID}}" placeholder="control account id">
          <button type="submit">Apply filters</button>
        </form>
      </section>
      <div class="grid">
        <section class="panel" id="journal-entries">
          <h2>Journal entries</h2>
          <table>
            <thead>
              <tr>
                <th>Entry</th>
                <th>Scope</th>
                <th>Summary</th>
                <th>Totals</th>
              </tr>
            </thead>
            <tbody>
            {{range .JournalEntries}}
            <tr>
              <td>
                  <a href="{{accountingEntryHref .EntryID}}">#{{.EntryNumber}}</a>
                  <div class="meta">{{.EntryKind}} | {{formatTime .PostedAt}}</div>
                  <div class="meta"><a href="/app/review/audit?entity_type=accounting.journal_entry&amp;entity_id={{.EntryID}}">Audit trail</a></div>
              </td>
              <td>{{.TaxScopeCode}}</td>
              <td>
                  {{.Summary}}
                  {{if .SourceDocumentID.Valid}}
                  <div class="meta">
                    <a href="{{documentReviewHref .SourceDocumentID.String}}">Source document</a>
                    {{if .DocumentStatus.Valid}} | <span class="status-pill {{statusClass .DocumentStatus.String}}">{{.DocumentStatus.String}}</span>{{end}}
                  </div>
                  {{if or .RequestReference.Valid .RecommendationID.Valid .ApprovalID.Valid .RunID.Valid}}
                  <div class="meta">
                    {{if .RequestReference.Valid}}<a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>{{end}}
                    {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">{{if .RecommendationStatus.Valid}}{{.RecommendationStatus.String}}{{else}}Proposal{{end}}</a>{{end}}
                    {{if .ApprovalID.Valid}} | <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
                    {{if .RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a>{{end}}
                  </div>
                  {{end}}
                  <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.SourceDocumentID.String}}">Document audit</a></div>
                  {{end}}
                </td>
                <td>Dr {{.TotalDebitMinor}} / Cr {{.TotalCreditMinor}}</td>
              </tr>
              {{else}}
              <tr><td colspan="4">No journal entries available.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
        <section class="panel" id="control-accounts">
          <h2>Control accounts</h2>
          <table>
            <thead>
              <tr>
                <th>Code</th>
                <th>Type</th>
                <th>Last effective</th>
                <th>Net</th>
              </tr>
            </thead>
            <tbody>
              {{range .ControlBalances}}
              <tr>
                <td>
                  <a href="{{controlAccountHref .AccountID}}">{{.AccountCode}}</a>
                  <div class="meta">{{.AccountName}}</div>
                </td>
                <td>
                  <a href="{{accountingReviewHref $accounting.StartOn $accounting.EndOn $accounting.AsOf $accounting.EntryID $accounting.DocumentID $accounting.TaxType $accounting.TaxCode .ControlType "" "control-accounts"}}">{{.ControlType}}</a>
                </td>
                <td>{{if .LastEffectiveOn.Valid}}{{formatTime .LastEffectiveOn.Time}}{{else}}-{{end}}</td>
                <td>{{.NetMinor}}</td>
              </tr>
              {{else}}
              <tr><td colspan="4">No control accounts available.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>
      <section class="panel" id="tax-summaries">
        <h2>Tax summaries</h2>
        <table>
          <thead>
            <tr>
              <th>Tax code</th>
              <th>Type</th>
              <th>Entries</th>
              <th>Linked control accounts</th>
              <th>Net</th>
            </tr>
          </thead>
          <tbody>
            {{range .TaxSummaries}}
            <tr>
              <td>
                <a href="{{taxSummaryHref .TaxCode}}">{{.TaxCode}}</a>
                <div class="meta">{{.TaxName}}</div>
              </td>
              <td><a href="{{accountingReviewHref $accounting.StartOn $accounting.EndOn $accounting.AsOf $accounting.EntryID $accounting.DocumentID .TaxType .TaxCode $accounting.ControlType $accounting.AccountID "tax-summaries"}}">{{.TaxType}}</a></td>
              <td>{{.EntryCount}}</td>
              <td>
                {{if .ReceivableAccountID.Valid}}<a href="{{controlAccountHref .ReceivableAccountID.String}}">{{.ReceivableAccountCode.String}}</a>{{else}}-{{end}}
                /
                {{if .PayableAccountID.Valid}}<a href="{{controlAccountHref .PayableAccountID.String}}">{{.PayableAccountCode.String}}</a>{{else}}-{{end}}
              </td>
              <td>{{.NetMinor}}</td>
            </tr>
            {{else}}
            <tr><td colspan="5">No tax summaries available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .AccountingDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Journal entry #{{.Review.EntryNumber}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.EntryKind}}">{{.Review.EntryKind}}</span>
          <p><strong>{{.Review.Summary}}</strong></p>
          <p class="meta">{{.Review.EntryID}}</p>
          <p class="meta">Effective: {{formatTime .Review.EffectiveOn}} | Posted: {{formatTime .Review.PostedAt}} | Tax scope: {{.Review.TaxScopeCode}}</p>
          <p class="meta">
            <a href="/app/review/accounting?entry_id={{.Review.EntryID}}">Filtered accounting view</a> |
            <a href="/app/review/audit?entity_type=accounting.journal_entry&amp;entity_id={{.Review.EntryID}}">Audit trail</a>
            {{if .Review.SourceDocumentID.Valid}} | <a href="{{documentReviewHref .Review.SourceDocumentID.String}}">Source document</a>{{end}}
            {{if .Review.RequestReference.Valid}} | <a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{end}}
            {{if .Review.RecommendationID.Valid}} | <a href="{{proposalDetailHref .Review.RecommendationID.String}}">Proposal</a>{{end}}
            {{if .Review.RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">AI run</a>{{end}}
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Posting detail</h2>
          <table>
            <tbody>
              <tr><th>Entry kind</th><td>{{.Review.EntryKind}}</td></tr>
              <tr><th>Currency</th><td>{{.Review.CurrencyCode}}</td></tr>
              <tr><th>Lines</th><td>{{.Review.LineCount}}</td></tr>
              <tr><th>Debit total</th><td>{{.Review.TotalDebitMinor}}</td></tr>
              <tr><th>Credit total</th><td>{{.Review.TotalCreditMinor}}</td></tr>
              <tr><th>Posted by</th><td>{{.Review.PostedByUserID}}</td></tr>
              <tr><th>Created</th><td>{{formatTime .Review.CreatedAt}}</td></tr>
              <tr><th>Reversal</th><td>{{if .Review.ReversalOfEntryID.Valid}}<a href="{{accountingEntryHref .Review.ReversalOfEntryID.String}}">Reversal of prior entry</a>{{else if .Review.HasReversal}}Reversed by a later entry{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Control chain</h2>
          <table>
            <tbody>
              <tr><th>Source document</th><td>{{if .Review.SourceDocumentID.Valid}}<a href="{{documentReviewHref .Review.SourceDocumentID.String}}">{{if .Review.DocumentNumber.Valid}}{{.Review.DocumentNumber.String}}{{else}}{{.Review.DocumentTypeCode.String}}{{end}}</a>{{if .Review.DocumentStatus.Valid}} | <span class="status-pill {{statusClass .Review.DocumentStatus.String}}">{{.Review.DocumentStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
              <tr><th>Document type</th><td>{{if .Review.DocumentTypeCode.Valid}}{{.Review.DocumentTypeCode.String}}{{else}}-{{end}}</td></tr>
              <tr><th>Document filter</th><td>{{if .Review.SourceDocumentID.Valid}}<a href="/app/review/accounting?document_id={{.Review.SourceDocumentID.String}}">All entries for source document</a>{{else}}-{{end}}</td></tr>
              <tr><th>Approval</th><td>{{if .Review.ApprovalID.Valid}}<a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{if .Review.ApprovalStatus.Valid}} | <span class="status-pill {{statusClass .Review.ApprovalStatus.String}}">{{.Review.ApprovalStatus.String}}</span>{{end}}{{else}}-{{end}}</td></tr>
              <tr><th>Request</th><td>{{if .Review.RequestReference.Valid}}<a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>Proposal</th><td>{{if .Review.RecommendationID.Valid}}<a href="{{proposalDetailHref .Review.RecommendationID.String}}">{{if .Review.RecommendationStatus.Valid}}{{.Review.RecommendationStatus.String}}{{else}}proposal{{end}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>AI run</th><td>{{if .Review.RunID.Valid}}<a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">{{.Review.RunID.String}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>Original entry</th><td>{{if .Review.ReversalOfEntryID.Valid}}<a href="{{accountingEntryHref .Review.ReversalOfEntryID.String}}">{{.Review.ReversalOfEntryID.String}}</a>{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .ControlAccountDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Control account {{.Balance.AccountCode}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Balance.ControlType}}">{{.Balance.ControlType}}</span>
          <p><strong>{{.Balance.AccountName}}</strong></p>
          <p class="meta">{{.Balance.AccountID}}</p>
          <p class="meta">
            <a href="{{accountingReviewHref .StartOn .EndOn .AsOf "" "" "" "" .Balance.ControlType .Balance.AccountID "control-accounts"}}">Filtered accounting view</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Balance detail</h2>
          <table>
            <tbody>
              <tr><th>Account class</th><td>{{.Balance.AccountClass}}</td></tr>
              <tr><th>Total debit</th><td>{{.Balance.TotalDebitMinor}}</td></tr>
              <tr><th>Total credit</th><td>{{.Balance.TotalCreditMinor}}</td></tr>
              <tr><th>Net</th><td>{{.Balance.NetMinor}}</td></tr>
              <tr><th>Last effective</th><td>{{if .Balance.LastEffectiveOn.Valid}}{{formatTime .Balance.LastEffectiveOn.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Linked tax summaries</h2>
          <table>
            <thead>
              <tr>
                <th>Tax code</th>
                <th>Type</th>
                <th>Net</th>
              </tr>
            </thead>
            <tbody>
              {{range .RelatedSummaries}}
              <tr>
                <td><a href="{{taxSummaryHref .TaxCode}}">{{.TaxCode}}</a></td>
                <td>{{.TaxType}}</td>
                <td>{{.NetMinor}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No linked tax summaries reference this control account.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .TaxSummaryDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Tax summary {{.Summary.TaxCode}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Summary.TaxType}}">{{.Summary.TaxType}}</span>
          <p><strong>{{.Summary.TaxName}}</strong></p>
          <p class="meta">Rate: {{.Summary.RateBasisPoints}} bps</p>
          <p class="meta">
            <a href="{{accountingReviewHref .StartOn .EndOn "" "" "" .Summary.TaxType .Summary.TaxCode "" "" "tax-summaries"}}">Filtered accounting view</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Summary totals</h2>
          <table>
            <tbody>
              <tr><th>Entry count</th><td>{{.Summary.EntryCount}}</td></tr>
              <tr><th>Document count</th><td>{{.Summary.DocumentCount}}</td></tr>
              <tr><th>Total debit</th><td>{{.Summary.TotalDebitMinor}}</td></tr>
              <tr><th>Total credit</th><td>{{.Summary.TotalCreditMinor}}</td></tr>
              <tr><th>Net</th><td>{{.Summary.NetMinor}}</td></tr>
              <tr><th>Last effective</th><td>{{if .Summary.LastEffectiveOn.Valid}}{{formatTime .Summary.LastEffectiveOn.Time}}{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Linked control accounts</h2>
          <table>
            <tbody>
              <tr><th>Receivable</th><td>{{if .Summary.ReceivableAccountID.Valid}}<a href="{{controlAccountHref .Summary.ReceivableAccountID.String}}">{{.Summary.ReceivableAccountCode.String}}</a>{{else}}-{{end}}</td></tr>
              <tr><th>Payable</th><td>{{if .Summary.PayableAccountID.Valid}}<a href="{{controlAccountHref .Summary.PayableAccountID.String}}">{{.Summary.PayableAccountCode.String}}</a>{{else}}-{{end}}</td></tr>
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Inventory}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Inventory review</h2>
        <form method="get" action="/app/review/inventory" class="inline-form">
          <input type="text" name="movement_id" value="{{.MovementID}}" placeholder="movement id">
          <input type="text" name="item_id" value="{{.ItemID}}" placeholder="item id">
          <input type="text" name="location_id" value="{{.LocationID}}" placeholder="location id">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <input type="text" name="movement_type" value="{{.MovementType}}" placeholder="movement type">
          <label><input type="checkbox" name="only_pending_accounting" value="true" {{if .OnlyPendingAccounting}}checked{{end}}> pending accounting</label>
          <label><input type="checkbox" name="only_pending_execution" value="true" {{if .OnlyPendingExecution}}checked{{end}}> pending execution</label>
          <button type="submit">Apply filters</button>
        </form>
      </section>
      <section class="panel" id="stock-balances">
        <h2>Stock balances</h2>
        <table>
          <thead>
            <tr>
              <th>Item</th>
              <th>Role</th>
              <th>Location</th>
              <th>On hand</th>
            </tr>
          </thead>
          <tbody>
            {{range .Stock}}
            <tr>
              <td>
                <a href="{{inventoryItemHref .ItemID}}">{{.ItemSKU}} | {{.ItemName}}</a>
                <div class="meta">
                  <a href="{{inventoryReviewHref "" .ItemID .LocationID "" "" false false "movement-history"}}">Movement history</a> |
                  <a href="{{inventoryReviewHref "" .ItemID "" "" "" false false "reconciliation"}}">Reconciliation</a>
                </div>
              </td>
              <td>{{.ItemRole}}</td>
              <td>
                <a href="{{inventoryLocationHref .LocationID}}">{{.LocationCode}} | {{.LocationName}}</a>
                <div class="meta"><a href="{{inventoryReviewHref "" "" .LocationID "" "" false false "movement-history"}}">Location movements</a></div>
              </td>
              <td>{{.OnHandMilli}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No stock balances available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
      <section class="panel" id="movement-history">
        <h2>Movement history</h2>
        <table>
          <thead>
            <tr>
              <th>Movement</th>
              <th>Item</th>
              <th>Route</th>
              <th>Quantity</th>
            </tr>
          </thead>
          <tbody>
            {{range .Movements}}
            <tr>
              <td>
                <a href="{{inventoryMovementHref .MovementID}}">#{{.MovementNumber}} | {{.MovementType}}</a>
                <div class="meta"><a href="/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id={{.MovementID}}">Audit trail</a></div>
                {{if or .RequestReference.Valid .RecommendationID.Valid .ApprovalID.Valid .RunID.Valid}}
                <div class="meta">
                  {{if .RequestReference.Valid}}<a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>{{end}}
                  {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">Proposal</a>{{end}}
                  {{if .ApprovalID.Valid}} | <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
                  {{if .RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a>{{end}}
                </div>
                {{end}}
              </td>
              <td>
                {{.ItemSKU}} | {{.ItemName}}
                <div class="meta"><a href="{{inventoryReviewHref "" .ItemID "" "" "" false false "movement-history"}}">Filter by item</a></div>
              </td>
              <td>{{if .SourceLocationCode.Valid}}{{.SourceLocationCode.String}}{{else}}-{{end}} -> {{if .DestinationLocationCode.Valid}}{{.DestinationLocationCode.String}}{{else}}-{{end}}</td>
              <td>
                {{.QuantityMilli}}
                {{if .DocumentID.Valid}}<div class="meta"><a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a></div>{{end}}
              </td>
            </tr>
            {{else}}
            <tr><td colspan="4">No inventory movements available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
      <section class="panel" id="reconciliation">
        <h2>Reconciliation</h2>
        <table>
          <thead>
            <tr>
              <th>Document</th>
              <th>Item</th>
              <th>Execution</th>
              <th>Accounting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Reconciliation}}
            <tr>
              <td>
                <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a> line {{.LineNumber}}
                <div class="meta"><a href="/app/review/audit?entity_type=documents.document&amp;entity_id={{.DocumentID}}">Audit trail</a></div>
                {{if or .RequestReference.Valid .RecommendationID.Valid .ApprovalID.Valid .RunID.Valid}}
                <div class="meta">
                  {{if .RequestReference.Valid}}<a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>{{end}}
                  {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">Proposal</a>{{end}}
                  {{if .ApprovalID.Valid}} | <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
                  {{if .RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a>{{end}}
                </div>
                {{end}}
              </td>
              <td>{{.ItemSKU}} | {{.ItemName}}</td>
              <td>{{if .WorkOrderID.Valid}}<a href="/app/review/work-orders/{{.WorkOrderID.String}}">{{.WorkOrderCode.String}}</a>{{else}}-{{end}} / {{if .ExecutionLinkStatus.Valid}}{{.ExecutionLinkStatus.String}}{{else}}-{{end}}</td>
              <td>{{if .JournalEntryID.Valid}}<a href="{{accountingEntryHref .JournalEntryID.String}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}} / {{if .AccountingHandoffStatus.Valid}}{{.AccountingHandoffStatus.String}}{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No reconciliation rows available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .InventoryDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Inventory movement #{{.Review.MovementNumber}}</h2>
        <p class="meta">
          <a href="/app/review/inventory?movement_id={{.Review.MovementID}}">Filtered inventory view</a> |
          <a href="/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id={{.Review.MovementID}}">Audit trail</a> |
          <a href="{{inventoryItemHref .Review.ItemID}}">Open item review</a> |
          <a href="{{inventoryReviewHref "" .Review.ItemID "" "" "" false false "movement-history"}}">Item movement history</a>
          {{if .Review.DocumentID.Valid}} | <a href="{{inventoryReviewHref "" "" "" .Review.DocumentID.String "" false false "reconciliation"}}">Document reconciliation</a>{{end}}
          {{if .Review.RequestReference.Valid}} | <a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{end}}
          {{if .Review.RecommendationID.Valid}} | <a href="{{proposalDetailHref .Review.RecommendationID.String}}">Proposal</a>{{end}}
          {{if .Review.ApprovalID.Valid}} | <a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
          {{if .Review.RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">AI run</a>{{end}}
        </p>
        <div class="detail-grid">
          <div class="detail-block"><strong>Movement ID</strong><br>{{.Review.MovementID}}</div>
          <div class="detail-block"><strong>Movement type</strong><br>{{.Review.MovementType}}</div>
          <div class="detail-block"><strong>Purpose</strong><br>{{.Review.MovementPurpose}}</div>
          <div class="detail-block"><strong>Usage</strong><br>{{.Review.UsageClassification}}</div>
          <div class="detail-block">
            <strong>Item</strong><br>
            <a href="{{inventoryItemHref .Review.ItemID}}">{{.Review.ItemSKU}} | {{.Review.ItemName}}</a>
            <div class="meta">
              <a href="{{inventoryReviewHref "" .Review.ItemID "" "" "" false false "stock-balances"}}">Stock balances</a> |
              <a href="{{inventoryReviewHref "" .Review.ItemID "" "" "" false false "movement-history"}}">Item movements</a> |
              <a href="{{inventoryReviewHref "" .Review.ItemID "" "" "" false false "reconciliation"}}">Item reconciliation</a>
            </div>
          </div>
          <div class="detail-block"><strong>Item role</strong><br>{{.Review.ItemRole}}</div>
          <div class="detail-block">
            <strong>Source</strong><br>
            {{if .Review.SourceLocationCode.Valid}}
            <a href="{{inventoryLocationHref .Review.SourceLocationID.String}}">{{.Review.SourceLocationCode.String}} | {{.Review.SourceLocationName.String}}</a>
            <div class="meta"><a href="{{inventoryReviewHref "" "" .Review.SourceLocationID.String "" "" false false "movement-history"}}">Location movements</a></div>
            {{else}}
            -
            {{end}}
          </div>
          <div class="detail-block">
            <strong>Destination</strong><br>
            {{if .Review.DestinationLocationCode.Valid}}
            <a href="{{inventoryLocationHref .Review.DestinationLocationID.String}}">{{.Review.DestinationLocationCode.String}} | {{.Review.DestinationLocationName.String}}</a>
            <div class="meta"><a href="{{inventoryReviewHref "" "" .Review.DestinationLocationID.String "" "" false false "movement-history"}}">Location movements</a></div>
            {{else}}
            -
            {{end}}
          </div>
          <div class="detail-block"><strong>Quantity</strong><br>{{.Review.QuantityMilli}}</div>
          <div class="detail-block"><strong>Created</strong><br>{{formatTime .Review.CreatedAt}}</div>
          <div class="detail-block"><strong>Created by</strong><br>{{.Review.CreatedByUserID}}</div>
          <div class="detail-block"><strong>Reference note</strong><br>{{if .Review.ReferenceNote}}{{.Review.ReferenceNote}}{{else}}-{{end}}</div>
        </div>
        {{if .Review.DocumentID.Valid}}
        <div class="detail-block">
          <strong>Source document</strong><br>
          <a href="{{documentReviewHref .Review.DocumentID.String}}">{{if .Review.DocumentTitle.Valid}}{{.Review.DocumentTitle.String}}{{else}}Document{{end}}</a>
          {{if .Review.DocumentNumber.Valid}} | {{.Review.DocumentNumber.String}}{{end}}
          {{if .Review.DocumentStatus.Valid}} | {{.Review.DocumentStatus.String}}{{end}}
          <div class="meta">
            <a href="{{inventoryReviewHref "" "" "" .Review.DocumentID.String "" false false "reconciliation"}}">Document reconciliation</a> |
            <a href="{{accountingReviewHref "" "" "" "" .Review.DocumentID.String "" "" "" "" ""}}">Accounting review</a>
            {{if .Review.RequestReference.Valid}} | <a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a>{{end}}
            {{if .Review.RecommendationID.Valid}} | <a href="{{proposalDetailHref .Review.RecommendationID.String}}">Proposal</a>{{end}}
            {{if .Review.ApprovalID.Valid}} | <a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalQueueCode.Valid}}{{.Review.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
            {{if .Review.RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">AI run</a>{{end}}
          </div>
        </div>
        {{end}}
      </section>
      <section class="panel">
        <h2>Reconciliation links</h2>
        <table>
          <thead>
            <tr>
              <th>Document line</th>
              <th>Execution</th>
              <th>Accounting</th>
              <th>Movement timing</th>
            </tr>
          </thead>
          <tbody>
            {{range .Reconciliation}}
            <tr>
              <td>
                <a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a> line {{.LineNumber}}
                <div class="meta">{{.DocumentTypeCode}} | {{.DocumentStatus}}</div>
                {{if or .RequestReference.Valid .RecommendationID.Valid .ApprovalID.Valid .RunID.Valid}}
                <div class="meta">
                  {{if .RequestReference.Valid}}<a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a>{{end}}
                  {{if .RecommendationID.Valid}} | <a href="{{proposalDetailHref .RecommendationID.String}}">Proposal</a>{{end}}
                  {{if .ApprovalID.Valid}} | <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
                  {{if .RunID.Valid}} | <a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a>{{end}}
                </div>
                {{end}}
              </td>
              <td>{{if .WorkOrderID.Valid}}<a href="/app/review/work-orders/{{.WorkOrderID.String}}">{{.WorkOrderCode.String}}</a>{{else}}-{{end}} / {{if .ExecutionLinkStatus.Valid}}{{.ExecutionLinkStatus.String}}{{else}}-{{end}}</td>
              <td>{{if .JournalEntryID.Valid}}<a href="{{accountingEntryHref .JournalEntryID.String}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}} / {{if .AccountingHandoffStatus.Valid}}{{.AccountingHandoffStatus.String}}{{else}}-{{end}}</td>
              <td>{{formatTime .MovementCreatedAt}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No reconciliation rows available for this movement.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .InventoryItemDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Inventory item {{.ItemSKU}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .ItemRole}}">{{.ItemRole}}</span>
          <p><strong>{{.ItemName}}</strong></p>
          <p class="meta">{{.ItemID}}</p>
          <p class="meta">
            <a href="{{inventoryReviewHref "" .ItemID "" "" "" false false "stock-balances"}}">Filtered inventory view</a> |
            <a href="/app/review/audit?entity_type=inventory_ops.item&amp;entity_id={{.ItemID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Stock balances</h2>
          <table>
            <thead>
              <tr>
                <th>Location</th>
                <th>Role</th>
                <th>On hand</th>
              </tr>
            </thead>
            <tbody>
              {{range .Stock}}
              <tr>
                <td><a href="{{inventoryLocationHref .LocationID}}">{{.LocationCode}} | {{.LocationName}}</a></td>
                <td>{{.LocationRole}}</td>
                <td>{{.OnHandMilli}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No stock balances available for this item.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Movement history</h2>
          <table>
            <thead>
              <tr>
                <th>Movement</th>
                <th>Route</th>
                <th>Quantity</th>
              </tr>
            </thead>
            <tbody>
              {{range .Movements}}
              <tr>
                <td>
                  <a href="{{inventoryMovementHref .MovementID}}">#{{.MovementNumber}} | {{.MovementType}}</a>
                  <div class="meta"><a href="/app/review/audit?entity_type=inventory_ops.movement&amp;entity_id={{.MovementID}}">Audit trail</a></div>
                </td>
                <td>
                  {{if .SourceLocationID.Valid}}<a href="{{inventoryLocationHref .SourceLocationID.String}}">{{.SourceLocationCode.String}}</a>{{else}}-{{end}}
                  ->
                  {{if .DestinationLocationID.Valid}}<a href="{{inventoryLocationHref .DestinationLocationID.String}}">{{.DestinationLocationCode.String}}</a>{{else}}-{{end}}
                </td>
                <td>{{.QuantityMilli}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No movements available for this item.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>
      <section class="panel">
        <h2>Reconciliation</h2>
        <table>
          <thead>
            <tr>
              <th>Document</th>
              <th>Movement</th>
              <th>Execution</th>
              <th>Accounting</th>
            </tr>
          </thead>
          <tbody>
            {{range .Reconciliation}}
            <tr>
              <td><a href="{{documentReviewHref .DocumentID}}">{{.DocumentTitle}}</a> line {{.LineNumber}}</td>
              <td><a href="{{inventoryMovementHref .MovementID}}">#{{.MovementNumber}}</a></td>
              <td>{{if .WorkOrderID.Valid}}<a href="/app/review/work-orders/{{.WorkOrderID.String}}">{{.WorkOrderCode.String}}</a>{{else}}-{{end}}</td>
              <td>{{if .JournalEntryID.Valid}}<a href="{{accountingEntryHref .JournalEntryID.String}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else if .JournalEntryNumber.Valid}}<a href="/app/review/accounting?document_id={{.DocumentID}}">Entry #{{.JournalEntryNumber.Int64}}</a>{{else}}-{{end}}</td>
            </tr>
            {{else}}
            <tr><td colspan="4">No reconciliation rows available for this item.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .InventoryLocationDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Inventory location {{.LocationCode}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .LocationRole}}">{{.LocationRole}}</span>
          <p><strong>{{.LocationName}}</strong></p>
          <p class="meta">{{.LocationID}}</p>
          <p class="meta">
            <a href="{{inventoryReviewHref "" "" .LocationID "" "" false false "stock-balances"}}">Filtered inventory view</a> |
            <a href="/app/review/audit?entity_type=inventory_ops.location&amp;entity_id={{.LocationID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Stock balances</h2>
          <table>
            <thead>
              <tr>
                <th>Item</th>
                <th>Role</th>
                <th>On hand</th>
              </tr>
            </thead>
            <tbody>
              {{range .Stock}}
              <tr>
                <td><a href="{{inventoryItemHref .ItemID}}">{{.ItemSKU}} | {{.ItemName}}</a></td>
                <td>{{.ItemRole}}</td>
                <td>{{.OnHandMilli}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No stock balances available for this location.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
        <section class="panel">
          <h2>Movement history</h2>
          <table>
            <thead>
              <tr>
                <th>Movement</th>
                <th>Item</th>
                <th>Route</th>
              </tr>
            </thead>
            <tbody>
              {{range .Movements}}
              <tr>
                <td><a href="{{inventoryMovementHref .MovementID}}">#{{.MovementNumber}} | {{.MovementType}}</a></td>
                <td><a href="{{inventoryItemHref .ItemID}}">{{.ItemSKU}} | {{.ItemName}}</a></td>
                <td>{{if .SourceLocationCode.Valid}}{{.SourceLocationCode.String}}{{else}}-{{end}} -> {{if .DestinationLocationCode.Valid}}{{.DestinationLocationCode.String}}{{else}}-{{end}}</td>
              </tr>
              {{else}}
              <tr><td colspan="3">No movements available for this location.</td></tr>
              {{end}}
            </tbody>
          </table>
        </section>
      </div>
    </div>
    {{end}}

    {{with .WorkOrders}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Work-order review</h2>
        <form method="get" action="/app/review/work-orders" class="inline-form">
          <input type="text" name="work_order_id" value="{{.WorkOrderID}}" placeholder="work order id">
          <input type="text" name="status" value="{{.Status}}" placeholder="status">
          <input type="text" name="document_id" value="{{.DocumentID}}" placeholder="document id">
          <button type="submit">Filter work orders</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Code</th>
              <th>Status</th>
              <th>Tasks</th>
              <th>Labor</th>
              <th>Material</th>
            </tr>
          </thead>
          <tbody>
            {{range .WorkOrders}}
            <tr>
              <td>
                <a href="/app/review/work-orders/{{.WorkOrderID}}">{{.WorkOrderCode}}</a>
                <div>{{.Title}}</div>
                <div class="meta">
                  <a href="{{documentReviewHref .DocumentID}}">Source document</a> |
                  {{if .RequestReference.Valid}}<a href="{{inboundRequestHref .RequestReference.String}}">{{.RequestReference.String}}</a> | {{end}}
                  {{if .RecommendationID.Valid}}<a href="{{proposalDetailHref .RecommendationID.String}}">{{if .RecommendationStatus.Valid}}{{.RecommendationStatus.String}}{{else}}Proposal{{end}}</a> | {{end}}
                  {{if .ApprovalID.Valid}}<a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalStatus.Valid}}{{.ApprovalStatus.String}}{{else}}Approval{{end}}</a> | {{end}}
                  {{if .RunID.Valid}}<a href="{{inboundSectionHref (printf "run:%s" .RunID.String) (runSectionID .RunID.String)}}">AI run</a> | {{end}}
                  <a href="/app/review/audit?entity_type=work_orders.work_order&amp;entity_id={{.WorkOrderID}}">Audit trail</a>
                </div>
              </td>
              <td><span class="status-pill {{statusClass .Status}}">{{.Status}}</span></td>
              <td>{{.OpenTaskCount}} open / {{.CompletedTaskCount}} done</td>
              <td>{{.LaborEntryCount}} entries / {{.TotalLaborMinutes}} min</td>
              <td>{{.MaterialUsageCount}} usages / {{.MaterialQuantityMilli}}</td>
            </tr>
            {{else}}
            <tr><td colspan="5">No work orders available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .WorkOrderDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Work order {{.Review.WorkOrderCode}}</h2>
        <div class="detail-block">
          <span class="status-pill {{statusClass .Review.Status}}">{{.Review.Status}}</span>
          <p>{{.Review.Title}}</p>
          <p class="meta">{{.Review.Summary}}</p>
          <p class="meta">
            <a href="{{workOrderReviewHref .Review.WorkOrderID "" ""}}">Filtered list view</a> |
            <a href="{{documentReviewHref .Review.DocumentID}}">Source document</a> |
            {{if .Review.RequestReference.Valid}}<a href="{{inboundRequestHref .Review.RequestReference.String}}">{{.Review.RequestReference.String}}</a> | {{end}}
            {{if .Review.RecommendationID.Valid}}<a href="{{proposalDetailHref .Review.RecommendationID.String}}">{{if .Review.RecommendationStatus.Valid}}{{.Review.RecommendationStatus.String}}{{else}}Proposal{{end}}</a> | {{end}}
            {{if .Review.ApprovalID.Valid}}<a href="{{approvalReviewHref .Review.ApprovalID.String}}">{{if .Review.ApprovalStatus.Valid}}{{.Review.ApprovalStatus.String}}{{else}}Approval{{end}}</a> | {{end}}
            {{if .Review.RunID.Valid}}<a href="{{inboundSectionHref (printf "run:%s" .Review.RunID.String) (runSectionID .Review.RunID.String)}}">AI run</a> | {{end}}
            <a href="{{accountingReviewHref "" "" "" "" .Review.DocumentID "" "" "" "" ""}}">Accounting review</a> |
            <a href="/app/review/audit?entity_type=work_orders.work_order&amp;entity_id={{.Review.WorkOrderID}}">Audit trail</a>
          </p>
        </div>
      </section>
      <div class="grid">
        <section class="panel">
          <h2>Execution rollup</h2>
          <div class="detail-block">Tasks: {{.Review.OpenTaskCount}} open / {{.Review.CompletedTaskCount}} completed</div>
          <div class="detail-block">Labor: {{.Review.LaborEntryCount}} entries / {{.Review.TotalLaborMinutes}} minutes / {{.Review.TotalLaborCostMinor}} minor</div>
          <div class="detail-block">Material: {{.Review.MaterialUsageCount}} usages / {{.Review.MaterialQuantityMilli}} milli / {{.Review.PostedMaterialCostMinor}} posted cost</div>
        </section>
        <section class="panel">
          <h2>Accounting linkage</h2>
          <div class="detail-block">Document status: {{.Review.DocumentStatus}}</div>
          <div class="detail-block">Posted labor entries: {{.Review.PostedLaborEntryCount}} / {{.Review.PostedLaborCostMinor}}</div>
          <div class="detail-block">Posted material usages: {{.Review.PostedMaterialUsageCount}}</div>
          <div class="detail-block">Last accounting post: {{if .Review.LastAccountingPostedAt.Valid}}<a href="/app/review/accounting?document_id={{.Review.DocumentID}}">{{formatTime .Review.LastAccountingPostedAt.Time}}</a>{{else}}-{{end}}</div>
        </section>
      </div>
    </div>
    {{end}}

    {{with .Audit}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Audit lookup</h2>
        <form method="get" action="/app/review/audit" class="inline-form">
          <input type="text" name="event_id" value="{{.EventID}}" placeholder="event id">
          <input type="text" name="entity_type" value="{{.EntityType}}" placeholder="entity type">
          <input type="text" name="entity_id" value="{{.EntityID}}" placeholder="entity id">
          <button type="submit">Search audit</button>
        </form>
      </section>
      <section class="panel">
        <table>
          <thead>
            <tr>
              <th>Occurred</th>
              <th>Event</th>
              <th>Entity</th>
              <th>Payload</th>
            </tr>
          </thead>
          <tbody>
            {{range .Events}}
            <tr>
              <td>{{formatTime .OccurredAt}}</td>
              <td>
                <strong><a href="{{auditEventHref .ID}}">{{.EventType}}</a></strong>
                <div class="meta">{{.ID}}</div>
              </td>
              <td>
                {{.EntityType}} / {{.EntityID}}
                {{if auditEntityHref .EntityType .EntityID}}
                <div class="meta"><a href="{{auditEntityHref .EntityType .EntityID}}">{{auditEntityLabel .EntityType}}</a></div>
                {{end}}
              </td>
              <td><pre>{{prettyJSON .Payload}}</pre></td>
            </tr>
            {{else}}
            <tr><td colspan="4">No audit events available.</td></tr>
            {{end}}
          </tbody>
        </table>
      </section>
    </div>
    {{end}}

    {{with .AuditDetail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}
      <section class="panel">
        <h2>Audit event {{.Event.ID}}</h2>
        <p class="meta">
          <a href="/app/review/audit?event_id={{.Event.ID}}">Filtered audit view</a>
          {{if auditEntityHref .Event.EntityType .Event.EntityID}} |
          <a href="{{auditEntityHref .Event.EntityType .Event.EntityID}}">{{auditEntityLabel .Event.EntityType}}</a>
          {{end}}
        </p>
        <table>
          <tbody>
            <tr><th>Occurred</th><td>{{formatTime .Event.OccurredAt}}</td></tr>
            <tr><th>Event type</th><td>{{.Event.EventType}}</td></tr>
            <tr><th>Entity</th><td>{{.Event.EntityType}} / {{.Event.EntityID}}</td></tr>
            <tr><th>Actor user</th><td>{{if .Event.ActorUserID.Valid}}{{.Event.ActorUserID.String}}{{else}}-{{end}}</td></tr>
          </tbody>
        </table>
      </section>
      <section class="panel">
        <h3>Payload</h3>
        <pre>{{prettyJSON .Event.Payload}}</pre>
      </section>
    </div>
    {{end}}

    {{with .Detail}}
    <div class="stack">
      {{if .Notice}}<div class="notice">{{.Notice}}</div>{{end}}
      {{if .Error}}<div class="error">{{.Error}}</div>{{end}}

      <section class="panel">
        <div class="page-header">
          <div>
            <div class="eyebrow">Request detail</div>
            <h2>Inbound request {{.Detail.Request.RequestReference}}</h2>
            <p class="meta">Review the persisted request, message and attachment evidence, lifecycle controls, and AI execution trace for one exact intake reference.</p>
          </div>
          <div class="page-actions">
            <a href="{{inboundRequestReview .Detail.Request.RequestReference}}" class="pill-link">Filtered request review</a>
            <a href="/app/review/audit?entity_type=ai.inbound_request&amp;entity_id={{.Detail.Request.RequestID}}" class="pill-link">Audit trail</a>
          </div>
        </div>
        <div class="detail-grid">
          <div class="detail-card">
            <strong>Status</strong><br>
            <span class="status-pill {{statusClass .Detail.Request.Status}}">{{.Detail.Request.Status}}</span>
          </div>
          <div class="detail-card">
            <strong>Channel</strong><br>{{.Detail.Request.Channel}}
          </div>
          <div class="detail-card">
            <strong>Origin</strong><br>{{.Detail.Request.OriginType}}
          </div>
          <div class="detail-card">
            <strong>Received</strong><br>{{formatTime .Detail.Request.ReceivedAt}}
          </div>
        </div>
        {{if .Detail.Request.CancelledAt.Valid}}<div class="detail-block">Cancelled: {{formatTime .Detail.Request.CancelledAt.Time}}{{if .Detail.Request.CancellationReason}} | {{.Detail.Request.CancellationReason}}{{end}}</div>{{end}}
        {{if .Detail.Request.FailedAt.Valid}}<div class="detail-block">Failed: {{formatTime .Detail.Request.FailedAt.Time}}{{if .Detail.Request.FailureReason}} | {{.Detail.Request.FailureReason}}{{end}}</div>{{end}}
        {{if eq .Detail.Request.Status "draft"}}
        <div class="detail-block">
          <h3>Edit draft</h3>
          <form method="post" action="/app/inbound-requests" enctype="multipart/form-data">
            <input type="hidden" name="request_id" value="{{.Detail.Request.RequestID}}">
            <input type="hidden" name="message_id" value="{{.EditableMessageID}}">
            <input type="hidden" name="return_to" value="/app/inbound-requests/{{.Detail.Request.RequestReference}}">
            <input type="hidden" name="channel" value="{{.Detail.Request.Channel}}">
            <label>Submitter label
              <input type="text" name="submitter_label" value="{{.EditableSubmitterLabel}}">
            </label>
            <label>Request message
              <textarea name="message_text" required>{{.EditableMessageText}}</textarea>
            </label>
            <label>Add attachments
              <input type="file" name="attachments" multiple>
            </label>
            <div class="inline-form">
              <button type="submit" name="intent" value="save_draft">Save draft</button>
              <button type="submit" name="intent" value="queue">Queue request</button>
            </div>
          </form>
          <form method="post" action="{{inboundActionHref .Detail.Request.RequestID "delete"}}" style="margin-top:10px;">
            <input type="hidden" name="return_to" value="/app/inbound-requests/{{.Detail.Request.RequestReference}}">
            <button type="submit" class="secondary">Delete draft</button>
          </form>
        </div>
        {{else if eq .Detail.Request.Status "queued"}}
        <div class="detail-block">
          <h3>Queued request actions</h3>
          <div class="inline-form">
            <form method="post" action="{{inboundActionHref .Detail.Request.RequestID "cancel"}}">
              <input type="hidden" name="return_to" value="/app/inbound-requests/{{.Detail.Request.RequestReference}}">
              <input type="text" name="reason" placeholder="Cancellation reason">
              <button type="submit" class="secondary">Cancel request</button>
            </form>
            <form method="post" action="{{inboundActionHref .Detail.Request.RequestID "amend"}}">
              <input type="hidden" name="return_to" value="/app/inbound-requests/{{.Detail.Request.RequestReference}}">
              <button type="submit">Return to draft</button>
            </form>
          </div>
        </div>
        {{else if eq .Detail.Request.Status "cancelled"}}
        <div class="detail-block">
          <h3>Cancelled request actions</h3>
          <form method="post" action="{{inboundActionHref .Detail.Request.RequestID "amend"}}">
            <input type="hidden" name="return_to" value="/app/inbound-requests/{{.Detail.Request.RequestReference}}">
            <button type="submit">Amend back to draft</button>
          </form>
        </div>
        {{end}}
        <div class="detail-block">
          <h3>Metadata</h3>
          <pre>{{prettyJSON .Detail.Request.Metadata}}</pre>
        </div>
      </section>

      <div class="grid">
        <section class="panel">
          <h2>Messages</h2>
          {{range .Detail.Messages}}
          <div class="detail-block">
            <strong>#{{.MessageIndex}} {{.MessageRole}}</strong>
            <p>{{.TextContent}}</p>
            <div class="meta">{{formatTime .CreatedAt}}</div>
          </div>
          {{else}}
          <div class="empty-state">No messages.</div>
          {{end}}
        </section>

        <section class="panel">
          <h2>Attachments</h2>
          {{range .Detail.Attachments}}
          <div class="detail-block">
            <div><a href="/api/attachments/{{.AttachmentID}}/content">{{.OriginalFileName}}</a></div>
            <div class="meta">{{.MediaType}} | {{.SizeBytes}} bytes | {{.LinkRole}}</div>
            {{if .LatestDerivedText.Valid}}
            <pre>{{.LatestDerivedText.String}}</pre>
            {{end}}
          </div>
          {{else}}
          <div class="empty-state">No attachments.</div>
          {{end}}
        </section>
      </div>

      <div class="grid">
        <section class="panel">
          <h2>AI runs</h2>
          {{range .Detail.Runs}}
          <div class="detail-block" id="{{runSectionID .RunID}}">
            <div><strong>{{.AgentRole}}</strong> / {{.CapabilityCode}}</div>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <p>{{.Summary}}</p>
            <div class="meta">{{.RunID}}</div>
            <div class="meta">
              Started: {{formatTime .StartedAt}}
              {{if .CompletedAt.Valid}} | Completed: {{formatTime .CompletedAt.Time}}{{end}} |
              <a href="/app/review/audit?entity_type=ai.agent_run&amp;entity_id={{.RunID}}">Audit trail</a>
            </div>
          </div>
          {{else}}
          <div class="empty-state">No AI runs yet.</div>
          {{end}}
        </section>

        <section class="panel">
          <h2>AI steps</h2>
          {{range .Detail.Steps}}
          <div class="detail-block" id="{{stepSectionID .StepID}}">
            <strong>#{{.StepIndex}} {{.StepTitle}}</strong>
            <div class="meta">
              {{.StepType}} |
              <a href="{{pageSectionHref (runSectionID .RunID)}}">Run {{.RunID}}</a>
            </div>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <div class="meta">
              Step: {{.StepID}} | Created: {{formatTime .CreatedAt}} |
              <a href="/app/review/audit?entity_type=ai.agent_run_step&amp;entity_id={{.StepID}}">Audit trail</a>
            </div>
            <details style="margin-top:10px;">
              <summary>Step payloads</summary>
              <div class="detail-block">
                <div class="meta">Input</div>
                <pre>{{prettyJSON .InputPayload}}</pre>
              </div>
              <div class="detail-block">
                <div class="meta">Output</div>
                <pre>{{prettyJSON .OutputPayload}}</pre>
              </div>
            </details>
          </div>
          {{else}}
          <div class="empty-state">No AI steps yet.</div>
          {{end}}
        </section>

        <section class="panel">
          <h2>Artifacts</h2>
          {{range .Detail.Artifacts}}
          <div class="detail-block">
            <strong>{{.Title}}</strong>
            <div class="meta">{{.ArtifactType}} | {{formatTime .CreatedAt}}</div>
            <pre>{{prettyJSON .Payload}}</pre>
          </div>
          {{else}}
          <div class="empty-state">No artifacts yet.</div>
          {{end}}
        </section>
      </div>

      <div class="grid">
        <section class="panel">
          <h2>Delegations</h2>
          {{range .Detail.Delegations}}
          <div class="detail-block" id="{{delegationSectionID .DelegationID}}">
            <strong>{{.CapabilityCode}}</strong>
            <div class="meta">Delegation: {{.DelegationID}}</div>
            <div class="meta">Parent run: <a href="{{pageSectionHref (runSectionID .ParentRunID)}}">{{.ParentRunID}}</a></div>
            <div class="meta">Child run: <a href="{{pageSectionHref (runSectionID .ChildRunID)}}">{{.ChildRunID}}</a> | {{.ChildAgentRole}} / {{.ChildCapabilityCode}}</div>
            {{if .RequestedByStepID.Valid}}<div class="meta">Requested by step: <a href="{{pageSectionHref (stepSectionID .RequestedByStepID.String)}}">{{.RequestedByStepID.String}}</a></div>{{end}}
            <div class="status-pill {{statusClass .ChildRunStatus}}">{{.ChildRunStatus}}</div>
            <p>{{.Reason}}</p>
            <div class="meta">
              Created: {{formatTime .CreatedAt}} |
              <a href="/app/review/audit?entity_type=ai.agent_delegation&amp;entity_id={{.DelegationID}}">Audit trail</a>
            </div>
          </div>
          {{else}}
          <div class="empty-state">No delegations yet.</div>
          {{end}}
        </section>

        <section class="panel">
          <h2>Recommendations</h2>
          {{range .Detail.Recommendations}}
          <div class="detail-block">
            <strong>{{.Summary}}</strong>
            <div class="status-pill {{statusClass .Status}}">{{.Status}}</div>
            <div class="meta">
              <a href="{{proposalDetailHref .RecommendationID}}">Open exact proposal</a> |
              <a href="/app/review/proposals?recommendation_id={{.RecommendationID}}">Filtered proposal review</a> |
              <a href="/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id={{.RecommendationID}}">Audit trail</a>
              {{if .ApprovalID.Valid}} | <a href="{{approvalReviewHref .ApprovalID.String}}">Exact approval</a>{{end}}
            </div>
            <pre>{{prettyJSON .Payload}}</pre>
          </div>
          {{else}}
          <div class="empty-state">No recommendations yet.</div>
          {{end}}
        </section>

        <section class="panel">
          <h2>Proposals</h2>
          {{range .Detail.Proposals}}
          <div class="detail-block">
            <strong><a href="{{proposalDetailHref .RecommendationID}}">{{.Summary}}</a></strong>
            <div class="meta">Recommendation: {{.RecommendationStatus}} | Approval: {{.ApprovalStatus.String}}</div>
            <div class="meta">
              Request: <a href="{{inboundRequestReview .RequestReference}}">{{.RequestReference}}</a> |
              Audit: <a href="/app/review/audit?entity_type=ai.agent_recommendation&amp;entity_id={{.RecommendationID}}">proposal trail</a>
            </div>
            <div class="meta">
              Document: {{if .DocumentID.Valid}}<a href="{{documentReviewHref .DocumentID.String}}">{{.DocumentTitle.String}}</a>{{else}}{{.DocumentTitle.String}}{{end}}
              {{if .ApprovalID.Valid}} | Approval: <a href="{{approvalReviewHref .ApprovalID.String}}">{{if .ApprovalQueueCode.Valid}}{{.ApprovalQueueCode.String}}{{else}}approval{{end}}</a>{{end}}
            </div>
            {{if .ApprovalID.Valid}}
            <form method="post" action="/app/approvals/{{.ApprovalID.String}}/decision">
              <input type="hidden" name="return_to" value="{{inboundRequestHref $.Detail.Request.RequestReference}}">
              <input type="text" name="decision_note" placeholder="Decision note">
              <div class="inline-form">
                <button type="submit" name="decision" value="approved">Approve</button>
                <button type="submit" name="decision" value="rejected" class="secondary">Reject</button>
              </div>
            </form>
            {{end}}
          </div>
          {{else}}
          <div class="empty-state">No downstream proposals yet.</div>
          {{end}}
        </section>
      </div>
    </div>
    {{end}}
  </div>
</body>
</html>`
