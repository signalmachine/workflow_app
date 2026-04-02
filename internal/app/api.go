package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"workflow_app/internal/accounting"
	"workflow_app/internal/ai"
	"workflow_app/internal/attachments"
	"workflow_app/internal/documents"
	"workflow_app/internal/identityaccess"
	"workflow_app/internal/intake"
	"workflow_app/internal/reporting"
	"workflow_app/internal/workflow"
)

const (
	sessionLoginPath           = "/api/session/login"
	sessionTokenPath           = "/api/session/token"
	sessionCurrentPath         = "/api/session"
	sessionRefreshPath         = "/api/session/refresh"
	sessionLogoutPath          = "/api/session/logout"
	webAppPath                 = "/app"
	webLoginPath               = "/app/login"
	webLogoutPath              = "/app/logout"
	webRouteCatalogPath        = "/app/routes"
	webSettingsPath            = "/app/settings"
	webAdminPath               = "/app/admin"
	webAdminAccountingPath     = "/app/admin/accounting"
	webAdminLedgerAccountsPath = "/app/admin/accounting/ledger-accounts"
	webAdminTaxCodesPath       = "/app/admin/accounting/tax-codes"
	webAdminPeriodsPath        = "/app/admin/accounting/periods"
	webOperationsPath          = "/app/operations"
	webOperationsFeedPath      = "/app/operations-feed"
	webAgentChatPath           = "/app/agent-chat"
	webReviewPath              = "/app/review"
	webSubmitInboundPagePath   = "/app/submit-inbound-request"
	webSubmitInboundPath       = "/app/inbound-requests"
	webInboundActionsPrefix    = "/app/inbound-requests/"
	webProcessNextQueuedPath   = "/app/agent/process-next-queued-inbound-request"
	webInboundDetailPrefix     = "/app/inbound-requests/"
	webInboundRequestsPath     = "/app/review/inbound-requests"
	webApprovalDecisionPrefix  = "/app/approvals/"
	webDocumentsPath           = "/app/review/documents"
	webDocumentDetailPrefix    = "/app/review/documents/"
	webAccountingPath          = "/app/review/accounting"
	webAccountingControlsPath  = "/app/review/accounting/control-accounts"
	webAccountingTaxesPath     = "/app/review/accounting/tax-summaries"
	webAccountingDetailPrefix  = "/app/review/accounting/"
	webApprovalsPath           = "/app/review/approvals"
	webApprovalDetailPrefix    = "/app/review/approvals/"
	webProposalsPath           = "/app/review/proposals"
	webProposalDetailPrefix    = "/app/review/proposals/"
	webInventoryHubPath        = "/app/inventory"
	webInventoryPath           = "/app/review/inventory"
	webInventoryItemsPath      = "/app/review/inventory/items"
	webInventoryLocationsPath  = "/app/review/inventory/locations"
	webInventoryDetailPrefix   = "/app/review/inventory/"
	webWorkOrdersPath          = "/app/review/work-orders"
	webAuditPath               = "/app/review/audit"
	webAuditDetailPrefix       = "/app/review/audit/"
	agentProcessNextQueuedPath = "/api/agent/process-next-queued-inbound-request"
	submitInboundRequestPath   = "/api/inbound-requests"
	inboundRequestActionPrefix = "/api/inbound-requests/"
	attachmentContentPrefix    = "/api/attachments/"
	reviewInboundRequestsPath  = "/api/review/inbound-requests"
	reviewInboundSummaryPath   = "/api/review/inbound-request-status-summary"
	reviewProposalListPath     = "/api/review/processed-proposals"
	reviewProposalActionPrefix = "/api/review/processed-proposals/"
	reviewProposalSummaryPath  = "/api/review/processed-proposal-status-summary"
	reviewApprovalQueuePath    = "/api/review/approval-queue"
	reviewDocumentsPath        = "/api/review/documents"
	reviewJournalEntriesPath   = "/api/review/accounting/journal-entries"
	reviewControlBalancesPath  = "/api/review/accounting/control-account-balances"
	reviewTaxSummariesPath     = "/api/review/accounting/tax-summaries"
	reviewInventoryStockPath   = "/api/review/inventory/stock"
	reviewInventoryMovesPath   = "/api/review/inventory/movements"
	reviewInventoryReconPath   = "/api/review/inventory/reconciliation"
	reviewWorkOrdersPath       = "/api/review/work-orders"
	reviewAuditEventsPath      = "/api/review/audit-events"
	adminLedgerAccountsPath    = "/api/admin/accounting/ledger-accounts"
	adminTaxCodesPath          = "/api/admin/accounting/tax-codes"
	adminPeriodsPath           = "/api/admin/accounting/periods"
	approvalDecisionPrefix     = "/api/approvals/"
	headerOrgID                = "X-Workflow-Org-ID"
	headerUserID               = "X-Workflow-User-ID"
	headerSessionID            = "X-Workflow-Session-ID"
	headerAuthorization        = "Authorization"
	sessionIDCookieName        = "workflow_session_id"
	refreshTokenCookieName     = "workflow_refresh_token"
)

var uuidPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

const browserSessionDuration = 24 * time.Hour
const accessTokenDuration = 15 * time.Minute

type ProcessNextQueuedInboundRequester interface {
	ProcessNextQueuedInboundRequest(ctx context.Context, input ProcessNextQueuedInboundRequestInput) (ProcessNextQueuedInboundRequestResult, error)
}

type queuedInboundRequestProcessorLoader func() (ProcessNextQueuedInboundRequester, error)

type inboundRequestSubmitter interface {
	SubmitInboundRequest(ctx context.Context, input SubmitInboundRequestInput) (SubmitInboundRequestResult, error)
	SaveInboundDraft(ctx context.Context, input SaveInboundDraftInput) (SaveInboundDraftResult, error)
	QueueInboundRequest(ctx context.Context, input QueueInboundRequestInput) (intake.InboundRequest, error)
	CancelInboundRequest(ctx context.Context, input CancelInboundRequestInput) (intake.InboundRequest, error)
	AmendInboundRequest(ctx context.Context, input AmendInboundRequestInput) (intake.InboundRequest, error)
	DeleteInboundDraft(ctx context.Context, input DeleteInboundDraftInput) error
	DownloadAttachment(ctx context.Context, input DownloadAttachmentInput) (attachments.AttachmentContent, error)
}

type operatorReviewReader interface {
	ListApprovalQueue(ctx context.Context, input reporting.ListApprovalQueueInput) ([]reporting.ApprovalQueueEntry, error)
	ListDocuments(ctx context.Context, input reporting.ListDocumentsInput) ([]reporting.DocumentReview, error)
	GetDocumentReview(ctx context.Context, input reporting.GetDocumentReviewInput) (reporting.DocumentReview, error)
	ListJournalEntries(ctx context.Context, input reporting.ListJournalEntriesInput) ([]reporting.JournalEntryReview, error)
	ListControlAccountBalances(ctx context.Context, input reporting.ListControlAccountBalancesInput) ([]reporting.ControlAccountBalance, error)
	ListTaxSummaries(ctx context.Context, input reporting.ListTaxSummariesInput) ([]reporting.TaxSummary, error)
	ListInventoryStock(ctx context.Context, input reporting.ListInventoryStockInput) ([]reporting.InventoryStockItem, error)
	ListInventoryMovements(ctx context.Context, input reporting.ListInventoryMovementsInput) ([]reporting.InventoryMovementReview, error)
	ListInventoryReconciliation(ctx context.Context, input reporting.ListInventoryReconciliationInput) ([]reporting.InventoryReconciliationItem, error)
	ListWorkOrders(ctx context.Context, input reporting.ListWorkOrdersInput) ([]reporting.WorkOrderReview, error)
	GetWorkOrderReview(ctx context.Context, input reporting.GetWorkOrderReviewInput) (reporting.WorkOrderReview, error)
	LookupAuditEvents(ctx context.Context, input reporting.LookupAuditEventsInput) ([]reporting.AuditEvent, error)
	ListInboundRequests(ctx context.Context, input reporting.ListInboundRequestsInput) ([]reporting.InboundRequestReview, error)
	GetInboundRequestDetail(ctx context.Context, input reporting.GetInboundRequestDetailInput) (reporting.InboundRequestDetail, error)
	ListInboundRequestStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.InboundRequestStatusSummary, error)
	ListProcessedProposals(ctx context.Context, input reporting.ListProcessedProposalsInput) ([]reporting.ProcessedProposalReview, error)
	ListProcessedProposalStatusSummary(ctx context.Context, actor identityaccess.Actor) ([]reporting.ProcessedProposalStatusSummary, error)
	GetWorkflowNavigationSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit int) (reporting.WorkflowNavigationSnapshot, error)
	GetOperationsFeedSnapshot(ctx context.Context, actor identityaccess.Actor, recentLimit int) (reporting.OperationsFeedSnapshot, error)
	GetOperationsLandingSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit, recentLimit int) (reporting.OperationsLandingSnapshot, error)
	GetDashboardSnapshot(ctx context.Context, actor identityaccess.Actor, pendingApprovalLimit, requestLimit, proposalLimit int) (reporting.DashboardSnapshot, error)
	GetAgentChatSnapshot(ctx context.Context, actor identityaccess.Actor, requestLimit, proposalLimit int) (reporting.AgentChatSnapshot, error)
	GetInventoryLandingSnapshot(ctx context.Context, actor identityaccess.Actor, recentLimit int) (reporting.InventoryLandingSnapshot, error)
}

type approvalDecisionService interface {
	DecideApproval(ctx context.Context, input workflow.DecideApprovalInput) (workflow.Approval, documents.Document, error)
}

type accountingAdminService interface {
	ListLedgerAccounts(ctx context.Context, input accounting.ListLedgerAccountsInput) ([]accounting.LedgerAccount, error)
	CreateLedgerAccount(ctx context.Context, input accounting.CreateLedgerAccountInput) (accounting.LedgerAccount, error)
	ListTaxCodes(ctx context.Context, input accounting.ListTaxCodesInput) ([]accounting.TaxCode, error)
	CreateTaxCode(ctx context.Context, input accounting.CreateTaxCodeInput) (accounting.TaxCode, error)
	ListAccountingPeriods(ctx context.Context, input accounting.ListAccountingPeriodsInput) ([]accounting.AccountingPeriod, error)
	CreateAccountingPeriod(ctx context.Context, input accounting.CreateAccountingPeriodInput) (accounting.AccountingPeriod, error)
	CloseAccountingPeriod(ctx context.Context, input accounting.CloseAccountingPeriodInput) (accounting.AccountingPeriod, error)
}

type proposalApprovalService interface {
	RequestProcessedProposalApproval(ctx context.Context, input requestProcessedProposalApprovalInput) (workflow.Approval, reporting.ProcessedProposalReview, error)
}

type browserSessionService interface {
	StartBrowserSession(ctx context.Context, input identityaccess.StartBrowserSessionInput) (identityaccess.BrowserSession, error)
	StartTokenSession(ctx context.Context, input identityaccess.StartTokenSessionInput) (identityaccess.TokenSession, error)
	AuthenticateSession(ctx context.Context, sessionID, refreshToken string) (identityaccess.SessionContext, error)
	AuthenticateAccessToken(ctx context.Context, accessToken string) (identityaccess.SessionContext, error)
	RefreshTokenSession(ctx context.Context, sessionID, refreshToken string, accessTokenExpiresAt time.Time) (identityaccess.TokenSession, error)
	RevokeAuthenticatedSession(ctx context.Context, sessionID, refreshToken string) error
	RevokeAccessTokenSession(ctx context.Context, accessToken string) error
}

type processNextQueuedRequest struct {
	Channel string `json:"channel"`
}

type submitInboundRequestRequest struct {
	OriginType     string                              `json:"origin_type"`
	Channel        string                              `json:"channel"`
	Metadata       map[string]any                      `json:"metadata"`
	Message        submitInboundRequestMessageRequest  `json:"message"`
	Attachments    []submitInboundRequestAttachmentDTO `json:"attachments"`
	QueueForReview *bool                               `json:"queue_for_review,omitempty"`
}

type saveInboundDraftRequest struct {
	RequestID   string                              `json:"request_id,omitempty"`
	MessageID   string                              `json:"message_id,omitempty"`
	OriginType  string                              `json:"origin_type"`
	Channel     string                              `json:"channel"`
	Metadata    map[string]any                      `json:"metadata"`
	Message     submitInboundRequestMessageRequest  `json:"message"`
	Attachments []submitInboundRequestAttachmentDTO `json:"attachments"`
}

type inboundRequestActionRequest struct {
	Reason string `json:"reason,omitempty"`
}

type submitInboundRequestMessageRequest struct {
	MessageRole string `json:"message_role"`
	TextContent string `json:"text_content"`
}

type submitInboundRequestAttachmentDTO struct {
	OriginalFileName string `json:"original_file_name"`
	MediaType        string `json:"media_type"`
	ContentBase64    string `json:"content_base64"`
	LinkRole         string `json:"link_role"`
}

type processNextQueuedResponse struct {
	Processed             bool   `json:"processed"`
	RequestReference      string `json:"request_reference,omitempty"`
	RequestStatus         string `json:"request_status,omitempty"`
	RunID                 string `json:"run_id,omitempty"`
	RunStatus             string `json:"run_status,omitempty"`
	ArtifactID            string `json:"artifact_id,omitempty"`
	RecommendationID      string `json:"recommendation_id,omitempty"`
	RecommendationSummary string `json:"recommendation_summary,omitempty"`
}

type processNextQueuedErrorResponse struct {
	Error            string `json:"error"`
	RequestReference string `json:"request_reference,omitempty"`
	RunID            string `json:"run_id,omitempty"`
}

type submitInboundRequestResponse struct {
	RequestID           string     `json:"request_id"`
	RequestReference    string     `json:"request_reference"`
	Status              string     `json:"status"`
	MessageID           string     `json:"message_id,omitempty"`
	AttachmentIDs       []string   `json:"attachment_ids,omitempty"`
	CancellationReason  string     `json:"cancellation_reason,omitempty"`
	FailureReason       string     `json:"failure_reason,omitempty"`
	ReceivedAt          time.Time  `json:"received_at"`
	QueuedAt            *time.Time `json:"queued_at,omitempty"`
	ProcessingStartedAt *time.Time `json:"processing_started_at,omitempty"`
	ProcessedAt         *time.Time `json:"processed_at,omitempty"`
	ActedOnAt           *time.Time `json:"acted_on_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	FailedAt            *time.Time `json:"failed_at,omitempty"`
	CancelledAt         *time.Time `json:"cancelled_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

type decideApprovalRequest struct {
	Decision     string `json:"decision"`
	DecisionNote string `json:"decision_note"`
}

type requestProcessedProposalApprovalRequest struct {
	QueueCode string `json:"queue_code"`
	Reason    string `json:"reason"`
}

type sessionLoginRequest struct {
	OrgSlug     string `json:"org_slug"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DeviceLabel string `json:"device_label"`
}

type sessionRefreshRequest struct {
	SessionID    string `json:"session_id"`
	RefreshToken string `json:"refresh_token"`
}

type createLedgerAccountRequest struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	AccountClass        string `json:"account_class"`
	ControlType         string `json:"control_type"`
	AllowsDirectPosting bool   `json:"allows_direct_posting"`
	TaxCategoryCode     string `json:"tax_category_code"`
}

type createTaxCodeRequest struct {
	Code                string `json:"code"`
	Name                string `json:"name"`
	TaxType             string `json:"tax_type"`
	RateBasisPoints     int    `json:"rate_basis_points"`
	ReceivableAccountID string `json:"receivable_account_id"`
	PayableAccountID    string `json:"payable_account_id"`
}

type createAccountingPeriodRequest struct {
	PeriodCode string `json:"period_code"`
	StartOn    string `json:"start_on"`
	EndOn      string `json:"end_on"`
}

type ledgerAccountResponse struct {
	ID                  string    `json:"id"`
	Code                string    `json:"code"`
	Name                string    `json:"name"`
	AccountClass        string    `json:"account_class"`
	ControlType         string    `json:"control_type"`
	AllowsDirectPosting bool      `json:"allows_direct_posting"`
	Status              string    `json:"status"`
	TaxCategoryCode     *string   `json:"tax_category_code,omitempty"`
	CreatedByUserID     string    `json:"created_by_user_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type taxCodeResponse struct {
	ID                  string    `json:"id"`
	Code                string    `json:"code"`
	Name                string    `json:"name"`
	TaxType             string    `json:"tax_type"`
	RateBasisPoints     int       `json:"rate_basis_points"`
	ReceivableAccountID *string   `json:"receivable_account_id,omitempty"`
	PayableAccountID    *string   `json:"payable_account_id,omitempty"`
	Status              string    `json:"status"`
	CreatedByUserID     string    `json:"created_by_user_id"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type accountingPeriodResponse struct {
	ID              string     `json:"id"`
	PeriodCode      string     `json:"period_code"`
	StartOn         string     `json:"start_on"`
	EndOn           string     `json:"end_on"`
	Status          string     `json:"status"`
	ClosedByUserID  *string    `json:"closed_by_user_id,omitempty"`
	ClosedAt        *time.Time `json:"closed_at,omitempty"`
	CreatedByUserID string     `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type AgentAPIHandler struct {
	loadProcessor     queuedInboundRequestProcessorLoader
	submissionService inboundRequestSubmitter
	reviewService     operatorReviewReader
	approvalService   approvalDecisionService
	proposalApproval  proposalApprovalService
	accountingAdmin   accountingAdminService
	authService       browserSessionService
}

func NewAgentAPIHandler(db *sql.DB) http.Handler {
	documentService := documents.NewService(db)
	authService := identityaccess.NewService(db)
	accountingService := accounting.NewService(db, documentService)
	return newAgentAPIHandlerWithDependencies(func() (ProcessNextQueuedInboundRequester, error) {
		return NewOpenAIAgentProcessorFromEnv(db)
	}, NewSubmissionService(db), reporting.NewService(db), workflow.NewService(db, documentService), newProcessedProposalApprovalService(db), accountingService, authService)
}

func NewAgentAPIHandlerWithProcessorLoader(loader queuedInboundRequestProcessorLoader) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, nil, nil, nil, nil)
}

func NewAgentAPIHandlerWithServices(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter) http.Handler {
	return NewAgentAPIHandlerWithDependencies(loader, submissionService, nil, nil, nil)
}

func NewAgentAPIHandlerWithDependencies(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter, reviewService operatorReviewReader, approvalService approvalDecisionService, authService browserSessionService) http.Handler {
	return newAgentAPIHandlerWithDependencies(loader, submissionService, reviewService, approvalService, nil, nil, authService)
}

func newAgentAPIHandlerWithDependencies(loader queuedInboundRequestProcessorLoader, submissionService inboundRequestSubmitter, reviewService operatorReviewReader, approvalService approvalDecisionService, proposalApproval proposalApprovalService, accountingAdmin accountingAdminService, authService browserSessionService) http.Handler {
	handler := &AgentAPIHandler{
		loadProcessor:     loader,
		submissionService: submissionService,
		reviewService:     reviewService,
		approvalService:   approvalService,
		proposalApproval:  proposalApproval,
		accountingAdmin:   accountingAdmin,
		authService:       authService,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.handleRoot)
	mux.HandleFunc(webAppPath, handler.handleWebAppDashboard)
	mux.HandleFunc(webLoginPath, handler.handleWebLogin)
	mux.HandleFunc(webLogoutPath, handler.handleWebLogout)
	mux.HandleFunc(webRouteCatalogPath, handler.handleWebRouteCatalog)
	mux.HandleFunc(webSettingsPath, handler.handleWebSettings)
	mux.HandleFunc(webAdminPath, handler.handleWebAdmin)
	mux.HandleFunc(webAdminAccountingPath, handler.handleWebAdminAccounting)
	mux.HandleFunc(webAdminLedgerAccountsPath, handler.handleWebCreateLedgerAccount)
	mux.HandleFunc(webAdminTaxCodesPath, handler.handleWebCreateTaxCode)
	mux.HandleFunc(webAdminPeriodsPath, handler.handleWebAccountingPeriods)
	mux.HandleFunc(webAdminPeriodsPath+"/", handler.handleWebAccountingPeriodAction)
	mux.HandleFunc(webOperationsPath, handler.handleWebOperationsLanding)
	mux.HandleFunc(webOperationsFeedPath, handler.handleWebOperationsFeed)
	mux.HandleFunc(webAgentChatPath, handler.handleWebAgentChat)
	mux.HandleFunc(webReviewPath, handler.handleWebReviewLanding)
	mux.HandleFunc(webSubmitInboundPagePath, handler.handleWebSubmitInboundRequestPage)
	mux.HandleFunc(webSubmitInboundPath, handler.handleWebSubmitInboundRequest)
	mux.HandleFunc(webProcessNextQueuedPath, handler.handleWebProcessNextQueuedInboundRequest)
	mux.HandleFunc(webInboundDetailPrefix, handler.handleWebInboundRequestDetail)
	mux.HandleFunc(webInboundRequestsPath, handler.handleWebInboundRequests)
	mux.HandleFunc(webApprovalDecisionPrefix, handler.handleWebApprovalDecision)
	mux.HandleFunc(webDocumentsPath, handler.handleWebDocuments)
	mux.HandleFunc(webDocumentDetailPrefix, handler.handleWebDocumentDetail)
	mux.HandleFunc(webAccountingPath, handler.handleWebAccounting)
	mux.HandleFunc(webAccountingControlsPath+"/", handler.handleWebControlAccountDetail)
	mux.HandleFunc(webAccountingTaxesPath+"/", handler.handleWebTaxSummaryDetail)
	mux.HandleFunc(webAccountingDetailPrefix, handler.handleWebAccountingDetail)
	mux.HandleFunc(webApprovalsPath, handler.handleWebApprovals)
	mux.HandleFunc(webApprovalDetailPrefix, handler.handleWebApprovalDetail)
	mux.HandleFunc(webProposalsPath, handler.handleWebProposals)
	mux.HandleFunc(webProposalDetailPrefix, handler.handleWebProposalDetail)
	mux.HandleFunc(webInventoryHubPath, handler.handleWebInventoryLanding)
	mux.HandleFunc(webInventoryPath, handler.handleWebInventory)
	mux.HandleFunc(webInventoryItemsPath+"/", handler.handleWebInventoryItemDetail)
	mux.HandleFunc(webInventoryLocationsPath+"/", handler.handleWebInventoryLocationDetail)
	mux.HandleFunc(webInventoryDetailPrefix, handler.handleWebInventoryDetail)
	mux.HandleFunc(webWorkOrdersPath, handler.handleWebWorkOrders)
	mux.HandleFunc(webWorkOrdersPath+"/", handler.handleWebWorkOrderDetail)
	mux.HandleFunc(webAuditPath, handler.handleWebAudit)
	mux.HandleFunc(webAuditDetailPrefix, handler.handleWebAuditDetail)
	mux.HandleFunc(sessionLoginPath, handler.handleSessionLogin)
	mux.HandleFunc(sessionTokenPath, handler.handleSessionTokenLogin)
	mux.HandleFunc(sessionCurrentPath, handler.handleCurrentSession)
	mux.HandleFunc(sessionRefreshPath, handler.handleSessionRefresh)
	mux.HandleFunc(sessionLogoutPath, handler.handleSessionLogout)
	mux.HandleFunc(agentProcessNextQueuedPath, handler.handleProcessNextQueuedInboundRequest)
	mux.HandleFunc(submitInboundRequestPath, handler.handleSubmitInboundRequest)
	mux.HandleFunc(inboundRequestActionPrefix, handler.handleInboundRequestAction)
	mux.HandleFunc(attachmentContentPrefix, handler.handleDownloadAttachment)
	mux.HandleFunc(reviewInboundRequestsPath, handler.handleListInboundRequests)
	mux.HandleFunc(reviewInboundRequestsPath+"/", handler.handleGetInboundRequestDetail)
	mux.HandleFunc(reviewInboundSummaryPath, handler.handleListInboundRequestStatusSummary)
	mux.HandleFunc(reviewProposalListPath, handler.handleListProcessedProposals)
	mux.HandleFunc(reviewProposalActionPrefix, handler.handleProcessedProposalAction)
	mux.HandleFunc(reviewProposalSummaryPath, handler.handleListProcessedProposalStatusSummary)
	mux.HandleFunc(reviewApprovalQueuePath, handler.handleListApprovalQueue)
	mux.HandleFunc(reviewDocumentsPath, handler.handleListDocuments)
	mux.HandleFunc(reviewJournalEntriesPath, handler.handleListJournalEntries)
	mux.HandleFunc(reviewControlBalancesPath, handler.handleListControlAccountBalances)
	mux.HandleFunc(reviewTaxSummariesPath, handler.handleListTaxSummaries)
	mux.HandleFunc(reviewInventoryStockPath, handler.handleListInventoryStock)
	mux.HandleFunc(reviewInventoryMovesPath, handler.handleListInventoryMovements)
	mux.HandleFunc(reviewInventoryReconPath, handler.handleListInventoryReconciliation)
	mux.HandleFunc(reviewWorkOrdersPath, handler.handleListWorkOrders)
	mux.HandleFunc(reviewWorkOrdersPath+"/", handler.handleGetWorkOrderReview)
	mux.HandleFunc(reviewAuditEventsPath, handler.handleLookupAuditEvents)
	mux.HandleFunc(adminLedgerAccountsPath, handler.handleAdminLedgerAccounts)
	mux.HandleFunc(adminTaxCodesPath, handler.handleAdminTaxCodes)
	mux.HandleFunc(adminPeriodsPath, handler.handleAdminAccountingPeriods)
	mux.HandleFunc(adminPeriodsPath+"/", handler.handleAdminAccountingPeriodAction)
	mux.HandleFunc(approvalDecisionPrefix, handler.handleDecideApproval)
	return mux
}

func populateInboundRequestDetailLookup(input *reporting.GetInboundRequestDetailInput, lookup string) {
	if input == nil {
		return
	}
	switch {
	case strings.HasPrefix(strings.ToLower(lookup), "run:"):
		input.RunID = strings.TrimSpace(lookup[len("run:"):])
	case strings.HasPrefix(strings.ToLower(lookup), "delegation:"):
		input.DelegationID = strings.TrimSpace(lookup[len("delegation:"):])
	case strings.HasPrefix(strings.ToLower(lookup), "step:"):
		input.StepID = strings.TrimSpace(lookup[len("step:"):])
	case strings.HasPrefix(strings.ToUpper(lookup), "REQ-"):
		input.RequestReference = lookup
	default:
		input.RequestID = lookup
	}
}
func (h *AgentAPIHandler) actorFromRequest(r *http.Request) (identityaccess.Actor, error) {
	if h.authService == nil {
		if actor, err := actorFromHeaders(r); err == nil {
			return actor, nil
		}
		return identityaccess.Actor{}, fmt.Errorf("unauthorized")
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		return identityaccess.Actor{}, err
	}
	return sessionContext.Actor, nil
}

func actorFromHeaders(r *http.Request) (identityaccess.Actor, error) {
	orgID := strings.TrimSpace(r.Header.Get(headerOrgID))
	userID := strings.TrimSpace(r.Header.Get(headerUserID))
	sessionID := strings.TrimSpace(r.Header.Get(headerSessionID))
	if orgID == "" || userID == "" || sessionID == "" {
		return identityaccess.Actor{}, fmt.Errorf("missing required authentication headers")
	}
	if !uuidPattern.MatchString(orgID) || !uuidPattern.MatchString(userID) || !uuidPattern.MatchString(sessionID) {
		return identityaccess.Actor{}, fmt.Errorf("authentication headers must be UUIDs")
	}

	return identityaccess.Actor{
		OrgID:     orgID,
		UserID:    userID,
		SessionID: sessionID,
	}, nil
}

func (h *AgentAPIHandler) sessionContextFromRequest(r *http.Request) (identityaccess.SessionContext, error) {
	if accessToken := bearerTokenFromRequest(r); accessToken != "" {
		return h.authService.AuthenticateAccessToken(r.Context(), accessToken)
	}
	sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
	if !ok {
		if _, err := actorFromHeaders(r); err != nil {
			return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
		}
		return identityaccess.SessionContext{}, identityaccess.ErrUnauthorized
	}
	return h.authService.AuthenticateSession(r.Context(), sessionID, refreshToken)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONBodyError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errRequestBodyRequired):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "request body is required"})
	default:
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON request body"})
	}
}

var errRequestBodyRequired = errors.New("request body is required")

func decodeJSONBody(r *http.Request, dst any, allowEmpty bool) error {
	if r == nil || r.Body == nil {
		if allowEmpty {
			return nil
		}
		return errRequestBodyRequired
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		if allowEmpty && errors.Is(err, io.EOF) {
			return nil
		}
		if errors.Is(err, io.EOF) {
			return errRequestBodyRequired
		}
		return err
	}

	var trailing any
	if err := decoder.Decode(&trailing); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	}
	return errors.New("invalid JSON request body")
}

func setSessionCookies(w http.ResponseWriter, secure bool, sessionID, refreshToken string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionIDCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  expiresAt,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  expiresAt,
	})
}

func clearSessionCookies(w http.ResponseWriter, secure bool) {
	for _, name := range []string{sessionIDCookieName, refreshTokenCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   secure,
			MaxAge:   -1,
			Expires:  time.Unix(0, 0).UTC(),
		})
	}
}

func sessionCookiesShouldBeSecure(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	for _, header := range []string{"X-Forwarded-Proto", "X-Forwarded-Scheme"} {
		value := strings.TrimSpace(r.Header.Get(header))
		if value == "" {
			continue
		}
		if idx := strings.Index(value, ","); idx >= 0 {
			value = value[:idx]
		}
		if strings.EqualFold(strings.TrimSpace(value), "https") {
			return true
		}
	}
	return false
}

func sessionCookiesFromRequest(r *http.Request) (string, string, bool) {
	sessionIDCookie, err := r.Cookie(sessionIDCookieName)
	if err != nil {
		return "", "", false
	}
	refreshTokenCookie, err := r.Cookie(refreshTokenCookieName)
	if err != nil {
		return "", "", false
	}

	sessionID := strings.TrimSpace(sessionIDCookie.Value)
	refreshToken := strings.TrimSpace(refreshTokenCookie.Value)
	if !uuidPattern.MatchString(sessionID) || refreshToken == "" {
		return "", "", false
	}
	return sessionID, refreshToken, true
}

func cookieValue(r *http.Request, name string) string {
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return cookie.Value
}

func bearerTokenFromRequest(r *http.Request) string {
	authorization := strings.TrimSpace(r.Header.Get(headerAuthorization))
	if authorization == "" {
		return ""
	}
	if !strings.HasPrefix(strings.ToLower(authorization), "bearer ") {
		return ""
	}
	return strings.TrimSpace(authorization[len("Bearer "):])
}

func parseAttachmentContentPath(path string) (string, bool) {
	if !strings.HasPrefix(path, attachmentContentPrefix) || !strings.HasSuffix(path, "/content") {
		return "", false
	}
	attachmentID := strings.TrimSuffix(strings.TrimPrefix(path, attachmentContentPrefix), "/content")
	attachmentID = strings.Trim(attachmentID, "/")
	if attachmentID == "" {
		return "", false
	}
	return attachmentID, true
}

func parseApprovalDecisionPath(path string) (string, bool) {
	if !strings.HasPrefix(path, approvalDecisionPrefix) || !strings.HasSuffix(path, "/decision") {
		return "", false
	}
	approvalID := strings.TrimSuffix(strings.TrimPrefix(path, approvalDecisionPrefix), "/decision")
	approvalID = strings.Trim(approvalID, "/")
	if approvalID == "" {
		return "", false
	}
	return approvalID, true
}

func parseChildPath(prefix, path string) (string, bool) {
	if !strings.HasPrefix(path, prefix+"/") {
		return "", false
	}
	segment := strings.TrimPrefix(path, prefix+"/")
	segment = strings.Trim(segment, "/")
	if segment == "" || strings.Contains(segment, "/") {
		return "", false
	}
	unescaped, err := url.PathUnescape(segment)
	if err != nil || strings.TrimSpace(unescaped) == "" {
		return "", false
	}
	return strings.TrimSpace(unescaped), true
}

func parseChildActionPath(prefix, path string) (string, string, bool) {
	if !strings.HasPrefix(path, prefix) {
		return "", "", false
	}
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	if trimmed == "" {
		return "", "", false
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) != 2 {
		return "", "", false
	}
	child, err := url.PathUnescape(parts[0])
	if err != nil || strings.TrimSpace(child) == "" {
		return "", "", false
	}
	action, err := url.PathUnescape(parts[1])
	if err != nil || strings.TrimSpace(action) == "" {
		return "", "", false
	}
	return strings.TrimSpace(child), strings.TrimSpace(action), true
}

func parseLimit(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	var limit int
	if _, err := fmt.Sscanf(raw, "%d", &limit); err != nil {
		return -1
	}
	return limit
}

func parseOptionalDate(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.DateOnly, raw)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func contentDisposition(fileName string) string {
	encoded := url.PathEscape(fileName)
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", fileName, encoded)
}

type approvalQueueEntryResponse struct {
	QueueEntryID         string     `json:"queue_entry_id"`
	ApprovalID           string     `json:"approval_id"`
	QueueCode            string     `json:"queue_code"`
	QueueStatus          string     `json:"queue_status"`
	EnqueuedAt           time.Time  `json:"enqueued_at"`
	ClosedAt             *time.Time `json:"closed_at,omitempty"`
	ApprovalStatus       string     `json:"approval_status"`
	RequestedAt          time.Time  `json:"requested_at"`
	RequestedByUserID    string     `json:"requested_by_user_id"`
	DecidedAt            *time.Time `json:"decided_at,omitempty"`
	DecidedByUserID      *string    `json:"decided_by_user_id,omitempty"`
	DocumentID           string     `json:"document_id"`
	DocumentTypeCode     string     `json:"document_type_code"`
	DocumentTitle        string     `json:"document_title"`
	DocumentNumber       *string    `json:"document_number,omitempty"`
	DocumentStatus       string     `json:"document_status"`
	JournalEntryID       *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber   *int64     `json:"journal_entry_number,omitempty"`
	JournalEntryPostedAt *time.Time `json:"journal_entry_posted_at,omitempty"`
}

type documentReviewResponse struct {
	DocumentID           string     `json:"document_id"`
	TypeCode             string     `json:"type_code"`
	Title                string     `json:"title"`
	NumberValue          *string    `json:"number_value,omitempty"`
	Status               string     `json:"status"`
	SourceDocumentID     *string    `json:"source_document_id,omitempty"`
	CreatedByUserID      string     `json:"created_by_user_id"`
	SubmittedByUserID    *string    `json:"submitted_by_user_id,omitempty"`
	SubmittedAt          *time.Time `json:"submitted_at,omitempty"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	RejectedAt           *time.Time `json:"rejected_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	ApprovalID           *string    `json:"approval_id,omitempty"`
	ApprovalStatus       *string    `json:"approval_status,omitempty"`
	ApprovalQueueCode    *string    `json:"approval_queue_code,omitempty"`
	ApprovalRequestedAt  *time.Time `json:"approval_requested_at,omitempty"`
	ApprovalDecidedAt    *time.Time `json:"approval_decided_at,omitempty"`
	JournalEntryID       *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber   *int64     `json:"journal_entry_number,omitempty"`
	JournalEntryPostedAt *time.Time `json:"journal_entry_posted_at,omitempty"`
}

type journalEntryReviewResponse struct {
	EntryID              string    `json:"entry_id"`
	EntryNumber          int64     `json:"entry_number"`
	EntryKind            string    `json:"entry_kind"`
	SourceDocumentID     *string   `json:"source_document_id,omitempty"`
	ReversalOfEntryID    *string   `json:"reversal_of_entry_id,omitempty"`
	CurrencyCode         string    `json:"currency_code"`
	TaxScopeCode         string    `json:"tax_scope_code"`
	Summary              string    `json:"summary"`
	ReversalReason       *string   `json:"reversal_reason,omitempty"`
	PostedByUserID       string    `json:"posted_by_user_id"`
	EffectiveOn          time.Time `json:"effective_on"`
	PostedAt             time.Time `json:"posted_at"`
	CreatedAt            time.Time `json:"created_at"`
	DocumentTypeCode     *string   `json:"document_type_code,omitempty"`
	DocumentNumber       *string   `json:"document_number,omitempty"`
	DocumentStatus       *string   `json:"document_status,omitempty"`
	ApprovalID           *string   `json:"approval_id,omitempty"`
	ApprovalStatus       *string   `json:"approval_status,omitempty"`
	ApprovalQueueCode    *string   `json:"approval_queue_code,omitempty"`
	RequestID            *string   `json:"request_id,omitempty"`
	RequestReference     *string   `json:"request_reference,omitempty"`
	RecommendationID     *string   `json:"recommendation_id,omitempty"`
	RecommendationStatus *string   `json:"recommendation_status,omitempty"`
	RunID                *string   `json:"run_id,omitempty"`
	LineCount            int       `json:"line_count"`
	TotalDebitMinor      int64     `json:"total_debit_minor"`
	TotalCreditMinor     int64     `json:"total_credit_minor"`
	HasReversal          bool      `json:"has_reversal"`
}

type controlAccountBalanceResponse struct {
	AccountID        string     `json:"account_id"`
	AccountCode      string     `json:"account_code"`
	AccountName      string     `json:"account_name"`
	AccountClass     string     `json:"account_class"`
	ControlType      string     `json:"control_type"`
	TotalDebitMinor  int64      `json:"total_debit_minor"`
	TotalCreditMinor int64      `json:"total_credit_minor"`
	NetMinor         int64      `json:"net_minor"`
	LastEffectiveOn  *time.Time `json:"last_effective_on,omitempty"`
}

type taxSummaryResponse struct {
	TaxType               string     `json:"tax_type"`
	TaxCode               string     `json:"tax_code"`
	TaxName               string     `json:"tax_name"`
	RateBasisPoints       int        `json:"rate_basis_points"`
	EntryCount            int        `json:"entry_count"`
	DocumentCount         int        `json:"document_count"`
	TotalDebitMinor       int64      `json:"total_debit_minor"`
	TotalCreditMinor      int64      `json:"total_credit_minor"`
	NetMinor              int64      `json:"net_minor"`
	ReceivableAccountID   *string    `json:"receivable_account_id,omitempty"`
	ReceivableAccountCode *string    `json:"receivable_account_code,omitempty"`
	ReceivableAccountName *string    `json:"receivable_account_name,omitempty"`
	PayableAccountID      *string    `json:"payable_account_id,omitempty"`
	PayableAccountCode    *string    `json:"payable_account_code,omitempty"`
	PayableAccountName    *string    `json:"payable_account_name,omitempty"`
	LastEffectiveOn       *time.Time `json:"last_effective_on,omitempty"`
}

type inventoryStockResponse struct {
	ItemID       string `json:"item_id"`
	ItemSKU      string `json:"item_sku"`
	ItemName     string `json:"item_name"`
	ItemRole     string `json:"item_role"`
	LocationID   string `json:"location_id"`
	LocationCode string `json:"location_code"`
	LocationName string `json:"location_name"`
	LocationRole string `json:"location_role"`
	OnHandMilli  int64  `json:"on_hand_milli"`
}

type inventoryMovementResponse struct {
	MovementID              string    `json:"movement_id"`
	MovementNumber          int64     `json:"movement_number"`
	DocumentID              *string   `json:"document_id,omitempty"`
	DocumentTypeCode        *string   `json:"document_type_code,omitempty"`
	DocumentTitle           *string   `json:"document_title,omitempty"`
	DocumentNumber          *string   `json:"document_number,omitempty"`
	DocumentStatus          *string   `json:"document_status,omitempty"`
	ApprovalID              *string   `json:"approval_id,omitempty"`
	ApprovalStatus          *string   `json:"approval_status,omitempty"`
	ApprovalQueueCode       *string   `json:"approval_queue_code,omitempty"`
	RequestID               *string   `json:"request_id,omitempty"`
	RequestReference        *string   `json:"request_reference,omitempty"`
	RecommendationID        *string   `json:"recommendation_id,omitempty"`
	RecommendationStatus    *string   `json:"recommendation_status,omitempty"`
	RunID                   *string   `json:"run_id,omitempty"`
	ItemID                  string    `json:"item_id"`
	ItemSKU                 string    `json:"item_sku"`
	ItemName                string    `json:"item_name"`
	ItemRole                string    `json:"item_role"`
	MovementType            string    `json:"movement_type"`
	MovementPurpose         string    `json:"movement_purpose"`
	UsageClassification     string    `json:"usage_classification"`
	SourceLocationID        *string   `json:"source_location_id,omitempty"`
	SourceLocationCode      *string   `json:"source_location_code,omitempty"`
	SourceLocationName      *string   `json:"source_location_name,omitempty"`
	SourceLocationRole      *string   `json:"source_location_role,omitempty"`
	DestinationLocationID   *string   `json:"destination_location_id,omitempty"`
	DestinationLocationCode *string   `json:"destination_location_code,omitempty"`
	DestinationLocationName *string   `json:"destination_location_name,omitempty"`
	DestinationLocationRole *string   `json:"destination_location_role,omitempty"`
	QuantityMilli           int64     `json:"quantity_milli"`
	ReferenceNote           string    `json:"reference_note"`
	CreatedByUserID         string    `json:"created_by_user_id"`
	CreatedAt               time.Time `json:"created_at"`
}

type inventoryReconciliationResponse struct {
	DocumentID              string     `json:"document_id"`
	DocumentTypeCode        string     `json:"document_type_code"`
	DocumentTitle           string     `json:"document_title"`
	DocumentNumber          *string    `json:"document_number,omitempty"`
	DocumentStatus          string     `json:"document_status"`
	ApprovalID              *string    `json:"approval_id,omitempty"`
	ApprovalStatus          *string    `json:"approval_status,omitempty"`
	ApprovalQueueCode       *string    `json:"approval_queue_code,omitempty"`
	RequestID               *string    `json:"request_id,omitempty"`
	RequestReference        *string    `json:"request_reference,omitempty"`
	RecommendationID        *string    `json:"recommendation_id,omitempty"`
	RecommendationStatus    *string    `json:"recommendation_status,omitempty"`
	RunID                   *string    `json:"run_id,omitempty"`
	DocumentLineID          string     `json:"document_line_id"`
	LineNumber              int        `json:"line_number"`
	MovementID              string     `json:"movement_id"`
	MovementNumber          int64      `json:"movement_number"`
	MovementType            string     `json:"movement_type"`
	MovementPurpose         string     `json:"movement_purpose"`
	UsageClassification     string     `json:"usage_classification"`
	ItemID                  string     `json:"item_id"`
	ItemSKU                 string     `json:"item_sku"`
	ItemName                string     `json:"item_name"`
	ItemRole                string     `json:"item_role"`
	SourceLocationID        *string    `json:"source_location_id,omitempty"`
	SourceLocationCode      *string    `json:"source_location_code,omitempty"`
	SourceLocationName      *string    `json:"source_location_name,omitempty"`
	DestinationLocationID   *string    `json:"destination_location_id,omitempty"`
	DestinationLocationCode *string    `json:"destination_location_code,omitempty"`
	DestinationLocationName *string    `json:"destination_location_name,omitempty"`
	QuantityMilli           int64      `json:"quantity_milli"`
	ExecutionLinkID         *string    `json:"execution_link_id,omitempty"`
	ExecutionContextType    *string    `json:"execution_context_type,omitempty"`
	ExecutionContextID      *string    `json:"execution_context_id,omitempty"`
	ExecutionLinkStatus     *string    `json:"execution_link_status,omitempty"`
	WorkOrderID             *string    `json:"work_order_id,omitempty"`
	WorkOrderCode           *string    `json:"work_order_code,omitempty"`
	WorkOrderStatus         *string    `json:"work_order_status,omitempty"`
	AccountingHandoffID     *string    `json:"accounting_handoff_id,omitempty"`
	AccountingHandoffStatus *string    `json:"accounting_handoff_status,omitempty"`
	CostMinor               *int64     `json:"cost_minor,omitempty"`
	CostCurrencyCode        *string    `json:"cost_currency_code,omitempty"`
	JournalEntryID          *string    `json:"journal_entry_id,omitempty"`
	JournalEntryNumber      *int64     `json:"journal_entry_number,omitempty"`
	AccountingPostedAt      *time.Time `json:"accounting_posted_at,omitempty"`
	MovementCreatedAt       time.Time  `json:"movement_created_at"`
}

type workOrderReviewResponse struct {
	WorkOrderID              string     `json:"work_order_id"`
	DocumentID               string     `json:"document_id"`
	DocumentStatus           string     `json:"document_status"`
	DocumentNumber           *string    `json:"document_number,omitempty"`
	WorkOrderCode            string     `json:"work_order_code"`
	Title                    string     `json:"title"`
	Summary                  string     `json:"summary"`
	Status                   string     `json:"status"`
	ClosedAt                 *time.Time `json:"closed_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	LastStatusChangedAt      time.Time  `json:"last_status_changed_at"`
	OpenTaskCount            int        `json:"open_task_count"`
	CompletedTaskCount       int        `json:"completed_task_count"`
	LaborEntryCount          int        `json:"labor_entry_count"`
	TotalLaborMinutes        int        `json:"total_labor_minutes"`
	TotalLaborCostMinor      int64      `json:"total_labor_cost_minor"`
	PostedLaborEntryCount    int        `json:"posted_labor_entry_count"`
	PostedLaborCostMinor     int64      `json:"posted_labor_cost_minor"`
	MaterialUsageCount       int        `json:"material_usage_count"`
	MaterialQuantityMilli    int64      `json:"material_quantity_milli"`
	PostedMaterialUsageCount int        `json:"posted_material_usage_count"`
	PostedMaterialCostMinor  int64      `json:"posted_material_cost_minor"`
	LastAccountingPostedAt   *time.Time `json:"last_accounting_posted_at,omitempty"`
}

type auditEventResponse struct {
	ID          string          `json:"id"`
	OrgID       *string         `json:"org_id,omitempty"`
	ActorUserID *string         `json:"actor_user_id,omitempty"`
	EventType   string          `json:"event_type"`
	EntityType  string          `json:"entity_type"`
	EntityID    string          `json:"entity_id"`
	Payload     json.RawMessage `json:"payload"`
	OccurredAt  time.Time       `json:"occurred_at"`
}

type sessionContextResponse struct {
	SessionID       string    `json:"session_id"`
	OrgID           string    `json:"org_id"`
	OrgSlug         string    `json:"org_slug"`
	OrgName         string    `json:"org_name"`
	UserID          string    `json:"user_id"`
	UserEmail       string    `json:"user_email"`
	UserDisplayName string    `json:"user_display_name"`
	RoleCode        string    `json:"role_code"`
	DeviceLabel     string    `json:"device_label"`
	ExpiresAt       time.Time `json:"expires_at"`
	IssuedAt        time.Time `json:"issued_at"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

type tokenSessionResponse struct {
	SessionID             string    `json:"session_id"`
	OrgID                 string    `json:"org_id"`
	OrgSlug               string    `json:"org_slug"`
	OrgName               string    `json:"org_name"`
	UserID                string    `json:"user_id"`
	UserEmail             string    `json:"user_email"`
	UserDisplayName       string    `json:"user_display_name"`
	RoleCode              string    `json:"role_code"`
	DeviceLabel           string    `json:"device_label"`
	ExpiresAt             time.Time `json:"expires_at"`
	IssuedAt              time.Time `json:"issued_at"`
	LastSeenAt            time.Time `json:"last_seen_at"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

func mapSessionContext(context identityaccess.SessionContext) sessionContextResponse {
	return sessionContextResponse{
		SessionID:       context.Session.ID,
		OrgID:           context.Actor.OrgID,
		OrgSlug:         context.OrgSlug,
		OrgName:         context.OrgName,
		UserID:          context.Actor.UserID,
		UserEmail:       context.UserEmail,
		UserDisplayName: context.UserDisplayName,
		RoleCode:        context.RoleCode,
		DeviceLabel:     context.Session.DeviceLabel,
		ExpiresAt:       context.Session.ExpiresAt,
		IssuedAt:        context.Session.IssuedAt,
		LastSeenAt:      context.Session.LastSeenAt,
	}
}

func mapTokenSession(session identityaccess.TokenSession) tokenSessionResponse {
	return tokenSessionResponse{
		SessionID:             session.Session.ID,
		OrgID:                 session.Session.OrgID,
		OrgSlug:               session.OrgSlug,
		OrgName:               session.OrgName,
		UserID:                session.Session.UserID,
		UserEmail:             session.UserEmail,
		UserDisplayName:       session.UserDisplayName,
		RoleCode:              session.RoleCode,
		DeviceLabel:           session.Session.DeviceLabel,
		ExpiresAt:             session.Session.ExpiresAt,
		IssuedAt:              session.Session.IssuedAt,
		LastSeenAt:            session.Session.LastSeenAt,
		AccessToken:           session.AccessToken,
		AccessTokenExpiresAt:  session.AccessTokenExpiresAt,
		RefreshToken:          session.RefreshToken,
		RefreshTokenExpiresAt: session.RefreshTokenExpiresAt,
	}
}

type inboundRequestReviewResponse struct {
	RequestID                string          `json:"request_id"`
	RequestReference         string          `json:"request_reference"`
	SessionID                *string         `json:"session_id,omitempty"`
	ActorUserID              *string         `json:"actor_user_id,omitempty"`
	OriginType               string          `json:"origin_type"`
	Channel                  string          `json:"channel"`
	Status                   string          `json:"status"`
	Metadata                 json.RawMessage `json:"metadata"`
	CancellationReason       string          `json:"cancellation_reason,omitempty"`
	FailureReason            string          `json:"failure_reason,omitempty"`
	ReceivedAt               time.Time       `json:"received_at"`
	QueuedAt                 *time.Time      `json:"queued_at,omitempty"`
	ProcessingStartedAt      *time.Time      `json:"processing_started_at,omitempty"`
	ProcessedAt              *time.Time      `json:"processed_at,omitempty"`
	ActedOnAt                *time.Time      `json:"acted_on_at,omitempty"`
	CompletedAt              *time.Time      `json:"completed_at,omitempty"`
	FailedAt                 *time.Time      `json:"failed_at,omitempty"`
	CancelledAt              *time.Time      `json:"cancelled_at,omitempty"`
	CreatedAt                time.Time       `json:"created_at"`
	UpdatedAt                time.Time       `json:"updated_at"`
	MessageCount             int             `json:"message_count"`
	AttachmentCount          int             `json:"attachment_count"`
	LastRunID                *string         `json:"last_run_id,omitempty"`
	LastRunStatus            *string         `json:"last_run_status,omitempty"`
	LastRecommendationID     *string         `json:"last_recommendation_id,omitempty"`
	LastRecommendationStatus *string         `json:"last_recommendation_status,omitempty"`
}

type inboundRequestStatusSummaryResponse struct {
	Status           string     `json:"status"`
	RequestCount     int        `json:"request_count"`
	MessageCount     int        `json:"message_count"`
	AttachmentCount  int        `json:"attachment_count"`
	LatestReceivedAt *time.Time `json:"latest_received_at,omitempty"`
	LatestQueuedAt   *time.Time `json:"latest_queued_at,omitempty"`
	LatestUpdatedAt  time.Time  `json:"latest_updated_at"`
}

type inboundRequestMessageResponse struct {
	MessageID       string    `json:"message_id"`
	MessageIndex    int       `json:"message_index"`
	MessageRole     string    `json:"message_role"`
	TextContent     string    `json:"text_content"`
	CreatedByUserID *string   `json:"created_by_user_id,omitempty"`
	AttachmentCount int       `json:"attachment_count"`
	CreatedAt       time.Time `json:"created_at"`
}

type requestAttachmentResponse struct {
	AttachmentID         string    `json:"attachment_id"`
	RequestMessageID     string    `json:"request_message_id"`
	LinkRole             string    `json:"link_role"`
	OriginalFileName     string    `json:"original_file_name"`
	MediaType            string    `json:"media_type"`
	SizeBytes            int64     `json:"size_bytes"`
	UploadedByUserID     *string   `json:"uploaded_by_user_id,omitempty"`
	LatestDerivedText    *string   `json:"latest_derived_text,omitempty"`
	LatestDerivedByRunID *string   `json:"latest_derived_by_run_id,omitempty"`
	DerivedTextCount     int       `json:"derived_text_count"`
	CreatedAt            time.Time `json:"created_at"`
}

type aiRunResponse struct {
	RunID          string     `json:"run_id"`
	AgentRole      string     `json:"agent_role"`
	CapabilityCode string     `json:"capability_code"`
	Status         string     `json:"status"`
	Summary        string     `json:"summary"`
	StartedAt      time.Time  `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
}

type aiStepResponse struct {
	StepID        string          `json:"step_id"`
	RunID         string          `json:"run_id"`
	StepIndex     int             `json:"step_index"`
	StepType      string          `json:"step_type"`
	StepTitle     string          `json:"step_title"`
	Status        string          `json:"status"`
	InputPayload  json.RawMessage `json:"input_payload"`
	OutputPayload json.RawMessage `json:"output_payload"`
	CreatedAt     time.Time       `json:"created_at"`
}

type aiDelegationResponse struct {
	DelegationID        string    `json:"delegation_id"`
	ParentRunID         string    `json:"parent_run_id"`
	ChildRunID          string    `json:"child_run_id"`
	RequestedByStepID   *string   `json:"requested_by_step_id,omitempty"`
	CapabilityCode      string    `json:"capability_code"`
	Reason              string    `json:"reason"`
	ChildAgentRole      string    `json:"child_agent_role"`
	ChildCapabilityCode string    `json:"child_capability_code"`
	ChildRunStatus      string    `json:"child_run_status"`
	CreatedAt           time.Time `json:"created_at"`
}

type aiArtifactResponse struct {
	ArtifactID      string          `json:"artifact_id"`
	RunID           string          `json:"run_id"`
	StepID          *string         `json:"step_id,omitempty"`
	ArtifactType    string          `json:"artifact_type"`
	Title           string          `json:"title"`
	Payload         json.RawMessage `json:"payload"`
	CreatedByUserID string          `json:"created_by_user_id"`
	CreatedAt       time.Time       `json:"created_at"`
}

type aiRecommendationResponse struct {
	RecommendationID   string          `json:"recommendation_id"`
	RunID              string          `json:"run_id"`
	ArtifactID         *string         `json:"artifact_id,omitempty"`
	ApprovalID         *string         `json:"approval_id,omitempty"`
	RecommendationType string          `json:"recommendation_type"`
	Status             string          `json:"status"`
	Summary            string          `json:"summary"`
	Payload            json.RawMessage `json:"payload"`
	CreatedByUserID    string          `json:"created_by_user_id"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type processedProposalReviewResponse struct {
	RequestID            string    `json:"request_id"`
	RequestReference     string    `json:"request_reference"`
	RequestStatus        string    `json:"request_status"`
	RecommendationID     string    `json:"recommendation_id"`
	RunID                string    `json:"run_id"`
	RecommendationType   string    `json:"recommendation_type"`
	RecommendationStatus string    `json:"recommendation_status"`
	Summary              string    `json:"summary"`
	SuggestedQueueCode   *string   `json:"suggested_queue_code,omitempty"`
	ApprovalID           *string   `json:"approval_id,omitempty"`
	ApprovalStatus       *string   `json:"approval_status,omitempty"`
	ApprovalQueueCode    *string   `json:"approval_queue_code,omitempty"`
	DocumentID           *string   `json:"document_id,omitempty"`
	DocumentTypeCode     *string   `json:"document_type_code,omitempty"`
	DocumentTitle        *string   `json:"document_title,omitempty"`
	DocumentNumber       *string   `json:"document_number,omitempty"`
	DocumentStatus       *string   `json:"document_status,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
}

type processedProposalStatusSummaryResponse struct {
	RecommendationStatus string    `json:"recommendation_status"`
	ProposalCount        int       `json:"proposal_count"`
	RequestCount         int       `json:"request_count"`
	DocumentCount        int       `json:"document_count"`
	LatestCreatedAt      time.Time `json:"latest_created_at"`
}

type inboundRequestDetailResponse struct {
	Request         inboundRequestReviewResponse      `json:"request"`
	Messages        []inboundRequestMessageResponse   `json:"messages"`
	Attachments     []requestAttachmentResponse       `json:"attachments"`
	Runs            []aiRunResponse                   `json:"runs"`
	Steps           []aiStepResponse                  `json:"steps"`
	Delegations     []aiDelegationResponse            `json:"delegations"`
	Artifacts       []aiArtifactResponse              `json:"artifacts"`
	Recommendations []aiRecommendationResponse        `json:"recommendations"`
	Proposals       []processedProposalReviewResponse `json:"proposals"`
}

type approvalDecisionResponse struct {
	Error           string     `json:"error,omitempty"`
	ApprovalID      string     `json:"approval_id"`
	Status          string     `json:"status"`
	QueueCode       string     `json:"queue_code"`
	DocumentID      string     `json:"document_id"`
	DocumentStatus  string     `json:"document_status"`
	DecisionNote    *string    `json:"decision_note,omitempty"`
	DecidedByUserID *string    `json:"decided_by_user_id,omitempty"`
	DecidedAt       *time.Time `json:"decided_at,omitempty"`
}

type processedProposalApprovalResponse struct {
	RecommendationID     string    `json:"recommendation_id"`
	RecommendationStatus string    `json:"recommendation_status"`
	ApprovalID           string    `json:"approval_id"`
	ApprovalStatus       string    `json:"approval_status"`
	ApprovalQueueCode    string    `json:"approval_queue_code"`
	DocumentID           string    `json:"document_id"`
	DocumentStatus       *string   `json:"document_status,omitempty"`
	RequestedAt          time.Time `json:"requested_at"`
}

func handleReviewError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, identityaccess.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	case errors.Is(err, reporting.ErrInvalidReviewFilter):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid review filter"})
	case errors.Is(err, workflow.ErrApprovalQueueRequired):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "approval queue is required"})
	case errors.Is(err, ErrProcessedProposalDocumentMissing):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "processed proposal document is required"})
	case errors.Is(err, ErrProcessedProposalApprovalExists), errors.Is(err, ai.ErrRecommendationApprovalLinked):
		writeJSON(w, http.StatusConflict, errorResponse{Error: "processed proposal already linked to approval"})
	case errors.Is(err, ErrProcessedProposalNotFound), errors.Is(err, ai.ErrRecommendationNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "record not found"})
	case errors.Is(err, sql.ErrNoRows), errors.Is(err, reporting.ErrWorkOrderNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "record not found"})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: fallback})
	}
}

func handleAccountingAdminError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, identityaccess.ErrUnauthorized):
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
	case errors.Is(err, accounting.ErrInvalidAccount):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid ledger account"})
	case errors.Is(err, accounting.ErrInvalidTaxCode):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid tax code"})
	case errors.Is(err, accounting.ErrInvalidAccountingPeriod):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid accounting period"})
	case errors.Is(err, accounting.ErrLedgerAccountNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "ledger account not found"})
	case errors.Is(err, accounting.ErrTaxCodeNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "tax code not found"})
	case errors.Is(err, accounting.ErrAccountingPeriodNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: "accounting period not found"})
	case errors.Is(err, accounting.ErrAccountingPeriodOverlap), errors.Is(err, accounting.ErrAccountingPeriodNotOpen):
		writeJSON(w, http.StatusConflict, errorResponse{Error: err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: fallback})
	}
}

func mapLedgerAccount(account accounting.LedgerAccount) ledgerAccountResponse {
	return ledgerAccountResponse{
		ID:                  account.ID,
		Code:                account.Code,
		Name:                account.Name,
		AccountClass:        account.AccountClass,
		ControlType:         account.ControlType,
		AllowsDirectPosting: account.AllowsDirectPosting,
		Status:              account.Status,
		TaxCategoryCode:     stringPtr(account.TaxCategoryCode),
		CreatedByUserID:     account.CreatedByUserID,
		CreatedAt:           account.CreatedAt,
		UpdatedAt:           account.UpdatedAt,
	}
}

func mapTaxCode(code accounting.TaxCode) taxCodeResponse {
	return taxCodeResponse{
		ID:                  code.ID,
		Code:                code.Code,
		Name:                code.Name,
		TaxType:             code.TaxType,
		RateBasisPoints:     code.RateBasisPoints,
		ReceivableAccountID: stringPtr(code.ReceivableAccountID),
		PayableAccountID:    stringPtr(code.PayableAccountID),
		Status:              code.Status,
		CreatedByUserID:     code.CreatedByUserID,
		CreatedAt:           code.CreatedAt,
		UpdatedAt:           code.UpdatedAt,
	}
}

func mapAccountingPeriod(period accounting.AccountingPeriod) accountingPeriodResponse {
	return accountingPeriodResponse{
		ID:              period.ID,
		PeriodCode:      period.PeriodCode,
		StartOn:         period.StartOn.Format(time.DateOnly),
		EndOn:           period.EndOn.Format(time.DateOnly),
		Status:          period.Status,
		ClosedByUserID:  stringPtr(period.ClosedByUserID),
		ClosedAt:        timePtr(period.ClosedAt),
		CreatedByUserID: period.CreatedByUserID,
		CreatedAt:       period.CreatedAt,
		UpdatedAt:       period.UpdatedAt,
	}
}

func mapApprovalQueueEntry(entry reporting.ApprovalQueueEntry) approvalQueueEntryResponse {
	return approvalQueueEntryResponse{
		QueueEntryID:         entry.QueueEntryID,
		ApprovalID:           entry.ApprovalID,
		QueueCode:            entry.QueueCode,
		QueueStatus:          entry.QueueStatus,
		EnqueuedAt:           entry.EnqueuedAt,
		ClosedAt:             timePtr(entry.ClosedAt),
		ApprovalStatus:       entry.ApprovalStatus,
		RequestedAt:          entry.RequestedAt,
		RequestedByUserID:    entry.RequestedByUserID,
		DecidedAt:            timePtr(entry.DecidedAt),
		DecidedByUserID:      stringPtr(entry.DecidedByUserID),
		DocumentID:           entry.DocumentID,
		DocumentTypeCode:     entry.DocumentTypeCode,
		DocumentTitle:        entry.DocumentTitle,
		DocumentNumber:       stringPtr(entry.DocumentNumber),
		DocumentStatus:       entry.DocumentStatus,
		JournalEntryID:       stringPtr(entry.JournalEntryID),
		JournalEntryNumber:   int64Ptr(entry.JournalEntryNumber),
		JournalEntryPostedAt: timePtr(entry.JournalEntryPostedAt),
	}
}

func mapDocumentReview(review reporting.DocumentReview) documentReviewResponse {
	return documentReviewResponse{
		DocumentID:           review.DocumentID,
		TypeCode:             review.TypeCode,
		Title:                review.Title,
		NumberValue:          stringPtr(review.NumberValue),
		Status:               review.Status,
		SourceDocumentID:     stringPtr(review.SourceDocumentID),
		CreatedByUserID:      review.CreatedByUserID,
		SubmittedByUserID:    stringPtr(review.SubmittedByUserID),
		SubmittedAt:          timePtr(review.SubmittedAt),
		ApprovedAt:           timePtr(review.ApprovedAt),
		RejectedAt:           timePtr(review.RejectedAt),
		CreatedAt:            review.CreatedAt,
		UpdatedAt:            review.UpdatedAt,
		ApprovalID:           stringPtr(review.ApprovalID),
		ApprovalStatus:       stringPtr(review.ApprovalStatus),
		ApprovalQueueCode:    stringPtr(review.ApprovalQueueCode),
		ApprovalRequestedAt:  timePtr(review.ApprovalRequestedAt),
		ApprovalDecidedAt:    timePtr(review.ApprovalDecidedAt),
		JournalEntryID:       stringPtr(review.JournalEntryID),
		JournalEntryNumber:   int64Ptr(review.JournalEntryNumber),
		JournalEntryPostedAt: timePtr(review.JournalEntryPostedAt),
	}
}

func mapJournalEntryReview(review reporting.JournalEntryReview) journalEntryReviewResponse {
	return journalEntryReviewResponse{
		EntryID:              review.EntryID,
		EntryNumber:          review.EntryNumber,
		EntryKind:            review.EntryKind,
		SourceDocumentID:     stringPtr(review.SourceDocumentID),
		ReversalOfEntryID:    stringPtr(review.ReversalOfEntryID),
		CurrencyCode:         review.CurrencyCode,
		TaxScopeCode:         review.TaxScopeCode,
		Summary:              review.Summary,
		ReversalReason:       stringPtr(review.ReversalReason),
		PostedByUserID:       review.PostedByUserID,
		EffectiveOn:          review.EffectiveOn,
		PostedAt:             review.PostedAt,
		CreatedAt:            review.CreatedAt,
		DocumentTypeCode:     stringPtr(review.DocumentTypeCode),
		DocumentNumber:       stringPtr(review.DocumentNumber),
		DocumentStatus:       stringPtr(review.DocumentStatus),
		ApprovalID:           stringPtr(review.ApprovalID),
		ApprovalStatus:       stringPtr(review.ApprovalStatus),
		ApprovalQueueCode:    stringPtr(review.ApprovalQueueCode),
		RequestID:            stringPtr(review.RequestID),
		RequestReference:     stringPtr(review.RequestReference),
		RecommendationID:     stringPtr(review.RecommendationID),
		RecommendationStatus: stringPtr(review.RecommendationStatus),
		RunID:                stringPtr(review.RunID),
		LineCount:            review.LineCount,
		TotalDebitMinor:      review.TotalDebitMinor,
		TotalCreditMinor:     review.TotalCreditMinor,
		HasReversal:          review.HasReversal,
	}
}

func mapControlAccountBalance(balance reporting.ControlAccountBalance) controlAccountBalanceResponse {
	return controlAccountBalanceResponse{
		AccountID:        balance.AccountID,
		AccountCode:      balance.AccountCode,
		AccountName:      balance.AccountName,
		AccountClass:     balance.AccountClass,
		ControlType:      balance.ControlType,
		TotalDebitMinor:  balance.TotalDebitMinor,
		TotalCreditMinor: balance.TotalCreditMinor,
		NetMinor:         balance.NetMinor,
		LastEffectiveOn:  timePtr(balance.LastEffectiveOn),
	}
}

func mapTaxSummary(summary reporting.TaxSummary) taxSummaryResponse {
	return taxSummaryResponse{
		TaxType:               summary.TaxType,
		TaxCode:               summary.TaxCode,
		TaxName:               summary.TaxName,
		RateBasisPoints:       summary.RateBasisPoints,
		EntryCount:            summary.EntryCount,
		DocumentCount:         summary.DocumentCount,
		TotalDebitMinor:       summary.TotalDebitMinor,
		TotalCreditMinor:      summary.TotalCreditMinor,
		NetMinor:              summary.NetMinor,
		ReceivableAccountID:   stringPtr(summary.ReceivableAccountID),
		ReceivableAccountCode: stringPtr(summary.ReceivableAccountCode),
		ReceivableAccountName: stringPtr(summary.ReceivableAccountName),
		PayableAccountID:      stringPtr(summary.PayableAccountID),
		PayableAccountCode:    stringPtr(summary.PayableAccountCode),
		PayableAccountName:    stringPtr(summary.PayableAccountName),
		LastEffectiveOn:       timePtr(summary.LastEffectiveOn),
	}
}

func mapInventoryStock(item reporting.InventoryStockItem) inventoryStockResponse {
	return inventoryStockResponse{
		ItemID:       item.ItemID,
		ItemSKU:      item.ItemSKU,
		ItemName:     item.ItemName,
		ItemRole:     item.ItemRole,
		LocationID:   item.LocationID,
		LocationCode: item.LocationCode,
		LocationName: item.LocationName,
		LocationRole: item.LocationRole,
		OnHandMilli:  item.OnHandMilli,
	}
}

func mapInventoryMovement(review reporting.InventoryMovementReview) inventoryMovementResponse {
	return inventoryMovementResponse{
		MovementID:              review.MovementID,
		MovementNumber:          review.MovementNumber,
		DocumentID:              stringPtr(review.DocumentID),
		DocumentTypeCode:        stringPtr(review.DocumentTypeCode),
		DocumentTitle:           stringPtr(review.DocumentTitle),
		DocumentNumber:          stringPtr(review.DocumentNumber),
		DocumentStatus:          stringPtr(review.DocumentStatus),
		ApprovalID:              stringPtr(review.ApprovalID),
		ApprovalStatus:          stringPtr(review.ApprovalStatus),
		ApprovalQueueCode:       stringPtr(review.ApprovalQueueCode),
		RequestID:               stringPtr(review.RequestID),
		RequestReference:        stringPtr(review.RequestReference),
		RecommendationID:        stringPtr(review.RecommendationID),
		RecommendationStatus:    stringPtr(review.RecommendationStatus),
		RunID:                   stringPtr(review.RunID),
		ItemID:                  review.ItemID,
		ItemSKU:                 review.ItemSKU,
		ItemName:                review.ItemName,
		ItemRole:                review.ItemRole,
		MovementType:            review.MovementType,
		MovementPurpose:         review.MovementPurpose,
		UsageClassification:     review.UsageClassification,
		SourceLocationID:        stringPtr(review.SourceLocationID),
		SourceLocationCode:      stringPtr(review.SourceLocationCode),
		SourceLocationName:      stringPtr(review.SourceLocationName),
		SourceLocationRole:      stringPtr(review.SourceLocationRole),
		DestinationLocationID:   stringPtr(review.DestinationLocationID),
		DestinationLocationCode: stringPtr(review.DestinationLocationCode),
		DestinationLocationName: stringPtr(review.DestinationLocationName),
		DestinationLocationRole: stringPtr(review.DestinationLocationRole),
		QuantityMilli:           review.QuantityMilli,
		ReferenceNote:           review.ReferenceNote,
		CreatedByUserID:         review.CreatedByUserID,
		CreatedAt:               review.CreatedAt,
	}
}

func mapInventoryReconciliation(item reporting.InventoryReconciliationItem) inventoryReconciliationResponse {
	return inventoryReconciliationResponse{
		DocumentID:              item.DocumentID,
		DocumentTypeCode:        item.DocumentTypeCode,
		DocumentTitle:           item.DocumentTitle,
		DocumentNumber:          stringPtr(item.DocumentNumber),
		DocumentStatus:          item.DocumentStatus,
		ApprovalID:              stringPtr(item.ApprovalID),
		ApprovalStatus:          stringPtr(item.ApprovalStatus),
		ApprovalQueueCode:       stringPtr(item.ApprovalQueueCode),
		RequestID:               stringPtr(item.RequestID),
		RequestReference:        stringPtr(item.RequestReference),
		RecommendationID:        stringPtr(item.RecommendationID),
		RecommendationStatus:    stringPtr(item.RecommendationStatus),
		RunID:                   stringPtr(item.RunID),
		DocumentLineID:          item.DocumentLineID,
		LineNumber:              item.LineNumber,
		MovementID:              item.MovementID,
		MovementNumber:          item.MovementNumber,
		MovementType:            item.MovementType,
		MovementPurpose:         item.MovementPurpose,
		UsageClassification:     item.UsageClassification,
		ItemID:                  item.ItemID,
		ItemSKU:                 item.ItemSKU,
		ItemName:                item.ItemName,
		ItemRole:                item.ItemRole,
		SourceLocationID:        stringPtr(item.SourceLocationID),
		SourceLocationCode:      stringPtr(item.SourceLocationCode),
		SourceLocationName:      stringPtr(item.SourceLocationName),
		DestinationLocationID:   stringPtr(item.DestinationLocationID),
		DestinationLocationCode: stringPtr(item.DestinationLocationCode),
		DestinationLocationName: stringPtr(item.DestinationLocationName),
		QuantityMilli:           item.QuantityMilli,
		ExecutionLinkID:         stringPtr(item.ExecutionLinkID),
		ExecutionContextType:    stringPtr(item.ExecutionContextType),
		ExecutionContextID:      stringPtr(item.ExecutionContextID),
		ExecutionLinkStatus:     stringPtr(item.ExecutionLinkStatus),
		WorkOrderID:             stringPtr(item.WorkOrderID),
		WorkOrderCode:           stringPtr(item.WorkOrderCode),
		WorkOrderStatus:         stringPtr(item.WorkOrderStatus),
		AccountingHandoffID:     stringPtr(item.AccountingHandoffID),
		AccountingHandoffStatus: stringPtr(item.AccountingHandoffStatus),
		CostMinor:               int64Ptr(item.CostMinor),
		CostCurrencyCode:        stringPtr(item.CostCurrencyCode),
		JournalEntryID:          stringPtr(item.JournalEntryID),
		JournalEntryNumber:      int64Ptr(item.JournalEntryNumber),
		AccountingPostedAt:      timePtr(item.AccountingPostedAt),
		MovementCreatedAt:       item.MovementCreatedAt,
	}
}

func mapWorkOrderReview(review reporting.WorkOrderReview) workOrderReviewResponse {
	return workOrderReviewResponse{
		WorkOrderID:              review.WorkOrderID,
		DocumentID:               review.DocumentID,
		DocumentStatus:           review.DocumentStatus,
		DocumentNumber:           stringPtr(review.DocumentNumber),
		WorkOrderCode:            review.WorkOrderCode,
		Title:                    review.Title,
		Summary:                  review.Summary,
		Status:                   review.Status,
		ClosedAt:                 timePtr(review.ClosedAt),
		CreatedAt:                review.CreatedAt,
		UpdatedAt:                review.UpdatedAt,
		LastStatusChangedAt:      review.LastStatusChangedAt,
		OpenTaskCount:            review.OpenTaskCount,
		CompletedTaskCount:       review.CompletedTaskCount,
		LaborEntryCount:          review.LaborEntryCount,
		TotalLaborMinutes:        review.TotalLaborMinutes,
		TotalLaborCostMinor:      review.TotalLaborCostMinor,
		PostedLaborEntryCount:    review.PostedLaborEntryCount,
		PostedLaborCostMinor:     review.PostedLaborCostMinor,
		MaterialUsageCount:       review.MaterialUsageCount,
		MaterialQuantityMilli:    review.MaterialQuantityMilli,
		PostedMaterialUsageCount: review.PostedMaterialUsageCount,
		PostedMaterialCostMinor:  review.PostedMaterialCostMinor,
		LastAccountingPostedAt:   timePtr(review.LastAccountingPostedAt),
	}
}

func mapAuditEvent(event reporting.AuditEvent) auditEventResponse {
	return auditEventResponse{
		ID:          event.ID,
		OrgID:       stringPtr(event.OrgID),
		ActorUserID: stringPtr(event.ActorUserID),
		EventType:   event.EventType,
		EntityType:  event.EntityType,
		EntityID:    event.EntityID,
		Payload:     event.Payload,
		OccurredAt:  event.OccurredAt,
	}
}

func mapInboundRequestReview(review reporting.InboundRequestReview) inboundRequestReviewResponse {
	return inboundRequestReviewResponse{
		RequestID:                review.RequestID,
		RequestReference:         review.RequestReference,
		SessionID:                stringPtr(review.SessionID),
		ActorUserID:              stringPtr(review.ActorUserID),
		OriginType:               review.OriginType,
		Channel:                  review.Channel,
		Status:                   review.Status,
		Metadata:                 review.Metadata,
		CancellationReason:       review.CancellationReason,
		FailureReason:            review.FailureReason,
		ReceivedAt:               review.ReceivedAt,
		QueuedAt:                 timePtr(review.QueuedAt),
		ProcessingStartedAt:      timePtr(review.ProcessingStartedAt),
		ProcessedAt:              timePtr(review.ProcessedAt),
		ActedOnAt:                timePtr(review.ActedOnAt),
		CompletedAt:              timePtr(review.CompletedAt),
		FailedAt:                 timePtr(review.FailedAt),
		CancelledAt:              timePtr(review.CancelledAt),
		CreatedAt:                review.CreatedAt,
		UpdatedAt:                review.UpdatedAt,
		MessageCount:             review.MessageCount,
		AttachmentCount:          review.AttachmentCount,
		LastRunID:                stringPtr(review.LastRunID),
		LastRunStatus:            stringPtr(review.LastRunStatus),
		LastRecommendationID:     stringPtr(review.LastRecommendationID),
		LastRecommendationStatus: stringPtr(review.LastRecommendationStatus),
	}
}

func mapInboundRequestMutationResponse(request intake.InboundRequest) submitInboundRequestResponse {
	return submitInboundRequestResponse{
		RequestID:           request.ID,
		RequestReference:    request.RequestReference,
		Status:              request.Status,
		CancellationReason:  request.CancellationReason,
		FailureReason:       request.FailureReason,
		ReceivedAt:          request.ReceivedAt,
		QueuedAt:            timePtr(request.QueuedAt),
		ProcessingStartedAt: timePtr(request.ProcessingStartedAt),
		ProcessedAt:         timePtr(request.ProcessedAt),
		ActedOnAt:           timePtr(request.ActedOnAt),
		CompletedAt:         timePtr(request.CompletedAt),
		FailedAt:            timePtr(request.FailedAt),
		CancelledAt:         timePtr(request.CancelledAt),
		CreatedAt:           request.CreatedAt,
		UpdatedAt:           request.UpdatedAt,
	}
}

func mapInboundRequestDetail(detail reporting.InboundRequestDetail) inboundRequestDetailResponse {
	response := inboundRequestDetailResponse{
		Request:         mapInboundRequestReview(detail.Request),
		Messages:        make([]inboundRequestMessageResponse, 0, len(detail.Messages)),
		Attachments:     make([]requestAttachmentResponse, 0, len(detail.Attachments)),
		Runs:            make([]aiRunResponse, 0, len(detail.Runs)),
		Steps:           make([]aiStepResponse, 0, len(detail.Steps)),
		Delegations:     make([]aiDelegationResponse, 0, len(detail.Delegations)),
		Artifacts:       make([]aiArtifactResponse, 0, len(detail.Artifacts)),
		Recommendations: make([]aiRecommendationResponse, 0, len(detail.Recommendations)),
		Proposals:       make([]processedProposalReviewResponse, 0, len(detail.Proposals)),
	}
	for _, item := range detail.Messages {
		response.Messages = append(response.Messages, inboundRequestMessageResponse{
			MessageID:       item.MessageID,
			MessageIndex:    item.MessageIndex,
			MessageRole:     item.MessageRole,
			TextContent:     item.TextContent,
			CreatedByUserID: stringPtr(item.CreatedByUserID),
			AttachmentCount: item.AttachmentCount,
			CreatedAt:       item.CreatedAt,
		})
	}
	for _, item := range detail.Attachments {
		response.Attachments = append(response.Attachments, requestAttachmentResponse{
			AttachmentID:         item.AttachmentID,
			RequestMessageID:     item.RequestMessageID,
			LinkRole:             item.LinkRole,
			OriginalFileName:     item.OriginalFileName,
			MediaType:            item.MediaType,
			SizeBytes:            item.SizeBytes,
			UploadedByUserID:     stringPtr(item.UploadedByUserID),
			LatestDerivedText:    stringPtr(item.LatestDerivedText),
			LatestDerivedByRunID: stringPtr(item.LatestDerivedByRunID),
			DerivedTextCount:     item.DerivedTextCount,
			CreatedAt:            item.CreatedAt,
		})
	}
	for _, item := range detail.Runs {
		response.Runs = append(response.Runs, aiRunResponse{
			RunID:          item.RunID,
			AgentRole:      item.AgentRole,
			CapabilityCode: item.CapabilityCode,
			Status:         item.Status,
			Summary:        item.Summary,
			StartedAt:      item.StartedAt,
			CompletedAt:    timePtr(item.CompletedAt),
		})
	}
	for _, item := range detail.Steps {
		response.Steps = append(response.Steps, aiStepResponse{
			StepID:        item.StepID,
			RunID:         item.RunID,
			StepIndex:     item.StepIndex,
			StepType:      item.StepType,
			StepTitle:     item.StepTitle,
			Status:        item.Status,
			InputPayload:  item.InputPayload,
			OutputPayload: item.OutputPayload,
			CreatedAt:     item.CreatedAt,
		})
	}
	for _, item := range detail.Delegations {
		response.Delegations = append(response.Delegations, aiDelegationResponse{
			DelegationID:        item.DelegationID,
			ParentRunID:         item.ParentRunID,
			ChildRunID:          item.ChildRunID,
			RequestedByStepID:   stringPtr(item.RequestedByStepID),
			CapabilityCode:      item.CapabilityCode,
			Reason:              item.Reason,
			ChildAgentRole:      item.ChildAgentRole,
			ChildCapabilityCode: item.ChildCapabilityCode,
			ChildRunStatus:      item.ChildRunStatus,
			CreatedAt:           item.CreatedAt,
		})
	}
	for _, item := range detail.Artifacts {
		response.Artifacts = append(response.Artifacts, aiArtifactResponse{
			ArtifactID:      item.ArtifactID,
			RunID:           item.RunID,
			StepID:          stringPtr(item.StepID),
			ArtifactType:    item.ArtifactType,
			Title:           item.Title,
			Payload:         item.Payload,
			CreatedByUserID: item.CreatedByUserID,
			CreatedAt:       item.CreatedAt,
		})
	}
	for _, item := range detail.Recommendations {
		response.Recommendations = append(response.Recommendations, aiRecommendationResponse{
			RecommendationID:   item.RecommendationID,
			RunID:              item.RunID,
			ArtifactID:         stringPtr(item.ArtifactID),
			ApprovalID:         stringPtr(item.ApprovalID),
			RecommendationType: item.RecommendationType,
			Status:             item.Status,
			Summary:            item.Summary,
			Payload:            item.Payload,
			CreatedByUserID:    item.CreatedByUserID,
			CreatedAt:          item.CreatedAt,
			UpdatedAt:          item.UpdatedAt,
		})
	}
	for _, item := range detail.Proposals {
		response.Proposals = append(response.Proposals, mapProcessedProposalReview(item))
	}
	return response
}

func mapProcessedProposalReview(item reporting.ProcessedProposalReview) processedProposalReviewResponse {
	return processedProposalReviewResponse{
		RequestID:            item.RequestID,
		RequestReference:     item.RequestReference,
		RequestStatus:        item.RequestStatus,
		RecommendationID:     item.RecommendationID,
		RunID:                item.RunID,
		RecommendationType:   item.RecommendationType,
		RecommendationStatus: item.RecommendationStatus,
		Summary:              item.Summary,
		SuggestedQueueCode:   stringPtr(item.SuggestedQueueCode),
		ApprovalID:           stringPtr(item.ApprovalID),
		ApprovalStatus:       stringPtr(item.ApprovalStatus),
		ApprovalQueueCode:    stringPtr(item.ApprovalQueueCode),
		DocumentID:           stringPtr(item.DocumentID),
		DocumentTypeCode:     stringPtr(item.DocumentTypeCode),
		DocumentTitle:        stringPtr(item.DocumentTitle),
		DocumentNumber:       stringPtr(item.DocumentNumber),
		DocumentStatus:       stringPtr(item.DocumentStatus),
		CreatedAt:            item.CreatedAt,
	}
}

func stringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	v := value.Time
	return &v
}

func int64Ptr(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	v := value.Int64
	return &v
}
