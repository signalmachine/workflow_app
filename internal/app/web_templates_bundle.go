package app

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

//go:embed web_templates/*.tmpl
var webTemplateFS embed.FS

var webTemplateBundle = template.Must(template.New("web").Funcs(webTemplateFuncs()).ParseFS(webTemplateFS, "web_templates/*.tmpl"))

func webTemplateFuncs() template.FuncMap {
	return template.FuncMap{
		"formatTime":             formatTemplateTime,
		"prettyJSON":             prettyTemplateJSON,
		"statusClass":            templateStatusClass,
		"dashboardStatusBlurb":   templateDashboardStatusBlurb,
		"dashboardStatusAction":  templateDashboardStatusAction,
		"dashboardRequestAction": templateDashboardRequestAction,
		"inboundRequestHref":     templateInboundRequestHref,
		"inboundSectionHref":     templateInboundRequestSectionHref,
		"runSectionID":           templateAIRunSectionID,
		"stepSectionID":          templateAIStepSectionID,
		"delegationSectionID":    templateAIDelegationSectionID,
		"pageSectionHref":        templatePageSectionHref,
		"inboundRequestReview":   templateInboundRequestReviewHref,
		"inboundRequestsHref":    templateInboundRequestsReviewHref,
		"documentReviewHref":     templateDocumentReviewHref,
		"accountingReviewHref":   templateAccountingReviewHref,
		"accountingEntryHref":    templateAccountingEntryHref,
		"controlAccountHref":     templateControlAccountHref,
		"taxSummaryHref":         templateTaxSummaryHref,
		"approvalReviewHref":     templateApprovalReviewHref,
		"approvalQueueHref":      templateApprovalQueueHref,
		"proposalDetailHref":     templateProposalDetailHref,
		"proposalReviewHref":     templateProposalReviewHref,
		"workOrderReviewHref":    templateWorkOrderReviewHref,
		"inventoryReviewHref":    templateInventoryReviewHref,
		"inventoryItemHref":      templateInventoryItemHref,
		"inventoryLocationHref":  templateInventoryLocationHref,
		"inventoryMovementHref":  templateInventoryMovementHref,
		"auditEventHref":         templateAuditEventHref,
		"auditEntityHref":        templateAuditEntityHref,
		"auditEntityLabel":       templateAuditEntityLabel,
		"inboundActionHref":      templateInboundActionHref,
		"dict":                   templateDict,
		"navClass":               templateNavClass,
		"navSectionClass":        templateNavSectionClass,
		"reviewNavClass":         templateReviewNavClass,
		"startsWith":             strings.HasPrefix,
		"trimSpace":              strings.TrimSpace,
		"joinRequestSummary":     templateRequestSummary,
	}
}

func templateDict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires even number of arguments")
	}
	dict := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func templateNavClass(activePath, path string) string {
	if strings.TrimSpace(activePath) == strings.TrimSpace(path) {
		return "nav-link is-active"
	}
	return "nav-link"
}

func templateReviewNavClass(activePath string) string {
	if strings.HasPrefix(strings.TrimSpace(activePath), "/app/review/") {
		return "nav-link is-active"
	}
	return "nav-link"
}

func templateNavSectionClass(activePath, section string) string {
	activePath = strings.TrimSpace(activePath)

	match := false
	switch strings.TrimSpace(section) {
	case "home":
		match = activePath == webAppPath
	case "intake":
		match = activePath == webSubmitInboundPagePath
	case "operations":
		match = activePath == webOperationsPath || activePath == webOperationsFeedPath || activePath == webAgentChatPath
	case "review":
		match = activePath == webReviewPath ||
			activePath == webInboundRequestsPath ||
			activePath == webApprovalsPath ||
			activePath == webProposalsPath ||
			activePath == webDocumentsPath ||
			activePath == webAccountingPath ||
			activePath == webWorkOrdersPath ||
			activePath == webAuditPath
	case "inventory":
		match = activePath == webInventoryHubPath || activePath == webInventoryPath
	}

	if match {
		return "nav-link is-active"
	}
	return "nav-link"
}

func templateRequestSummary(item any) string {
	switch v := item.(type) {
	case webOperationsFeedItem:
		return strings.TrimSpace(v.Summary)
	default:
		return ""
	}
}

func webTemplateName(data webPageData) string {
	switch {
	case data.ShowLogin:
		return "web_login"
	case data.Dashboard != nil:
		return "web_dashboard"
	case data.RouteCatalog != nil:
		return "web_route_catalog"
	case data.Settings != nil:
		return "web_settings"
	case data.Admin != nil:
		return "web_admin"
	case data.AdminAccounting != nil:
		return "web_admin_accounting"
	case data.AdminParties != nil:
		return "web_admin_parties"
	case data.OperationsLanding != nil:
		return "web_operations_landing"
	case data.InboundSubmit != nil:
		return "web_submit"
	case data.OperationsFeed != nil:
		return "web_operations_feed"
	case data.AgentChat != nil:
		return "web_agent_chat"
	case data.InboundRequests != nil:
		return "web_inbound_requests"
	case data.Detail != nil:
		return "web_inbound_detail"
	case data.Approvals != nil:
		return "web_approvals"
	case data.ApprovalDetail != nil:
		return "web_approval_detail"
	case data.ReviewLanding != nil:
		return "web_review_landing"
	case data.Proposals != nil:
		return "web_proposals"
	case data.ProposalDetail != nil:
		return "web_proposal_detail"
	case data.Documents != nil:
		return "web_documents"
	case data.DocumentDetail != nil:
		return "web_document_detail"
	case data.Accounting != nil:
		return "web_accounting"
	case data.AccountingDetail != nil:
		return "web_accounting_detail"
	case data.ControlAccountDetail != nil:
		return "web_control_account_detail"
	case data.TaxSummaryDetail != nil:
		return "web_tax_summary_detail"
	case data.InventoryLanding != nil:
		return "web_inventory_landing"
	case data.Inventory != nil:
		return "web_inventory"
	case data.InventoryDetail != nil:
		return "web_inventory_detail"
	case data.InventoryItemDetail != nil:
		return "web_inventory_item_detail"
	case data.InventoryLocationDetail != nil:
		return "web_inventory_location_detail"
	case data.WorkOrders != nil:
		return "web_work_orders"
	case data.WorkOrderDetail != nil:
		return "web_work_order_detail"
	case data.Audit != nil:
		return "web_audit"
	case data.AuditDetail != nil:
		return "web_audit_detail"
	default:
		return ""
	}
}

func normalizeWebPageFlash(data webPageData) webPageData {
	if strings.TrimSpace(data.Notice) != "" || strings.TrimSpace(data.Error) != "" {
		return data
	}

	switch {
	case data.Detail != nil:
		data.Notice = strings.TrimSpace(data.Detail.Notice)
		data.Error = strings.TrimSpace(data.Detail.Error)
	case data.DocumentDetail != nil:
		data.Notice = strings.TrimSpace(data.DocumentDetail.Notice)
		data.Error = strings.TrimSpace(data.DocumentDetail.Error)
	case data.AccountingDetail != nil:
		data.Notice = strings.TrimSpace(data.AccountingDetail.Notice)
		data.Error = strings.TrimSpace(data.AccountingDetail.Error)
	case data.ControlAccountDetail != nil:
		data.Notice = strings.TrimSpace(data.ControlAccountDetail.Notice)
		data.Error = strings.TrimSpace(data.ControlAccountDetail.Error)
	case data.TaxSummaryDetail != nil:
		data.Notice = strings.TrimSpace(data.TaxSummaryDetail.Notice)
		data.Error = strings.TrimSpace(data.TaxSummaryDetail.Error)
	case data.ApprovalDetail != nil:
		data.Notice = strings.TrimSpace(data.ApprovalDetail.Notice)
		data.Error = strings.TrimSpace(data.ApprovalDetail.Error)
	case data.ProposalDetail != nil:
		data.Notice = strings.TrimSpace(data.ProposalDetail.Notice)
		data.Error = strings.TrimSpace(data.ProposalDetail.Error)
	case data.InventoryDetail != nil:
		data.Notice = strings.TrimSpace(data.InventoryDetail.Notice)
		data.Error = strings.TrimSpace(data.InventoryDetail.Error)
	case data.InventoryItemDetail != nil:
		data.Notice = strings.TrimSpace(data.InventoryItemDetail.Notice)
		data.Error = strings.TrimSpace(data.InventoryItemDetail.Error)
	case data.InventoryLocationDetail != nil:
		data.Notice = strings.TrimSpace(data.InventoryLocationDetail.Notice)
		data.Error = strings.TrimSpace(data.InventoryLocationDetail.Error)
	case data.WorkOrderDetail != nil:
		data.Notice = strings.TrimSpace(data.WorkOrderDetail.Notice)
		data.Error = strings.TrimSpace(data.WorkOrderDetail.Error)
	case data.AuditDetail != nil:
		data.Notice = strings.TrimSpace(data.AuditDetail.Notice)
		data.Error = strings.TrimSpace(data.AuditDetail.Error)
	}

	return data
}

func (h *AgentAPIHandler) renderWebPage(w http.ResponseWriter, data webPageData) {
	data = normalizeWebPageFlash(data)

	templateName := webTemplateName(data)
	if templateName == "" {
		http.Error(w, "web page template not configured", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = webTemplateBundle.ExecuteTemplate(w, templateName, data)
}

func loginFormAction(data webPageData) string {
	action := strings.TrimSpace(data.LoginPath)
	if action == "" {
		return webLoginPath
	}
	return action
}

func sanitizeActiveQuery(target string) string {
	target = strings.TrimSpace(target)
	if target == "" {
		return ""
	}
	if strings.Contains(target, "://") {
		return ""
	}
	return target
}

func shellActionURL(basePath, key, value string) string {
	basePath = strings.TrimSpace(basePath)
	if basePath == "" {
		basePath = webAppPath
	}
	values := url.Values{}
	if strings.TrimSpace(value) != "" {
		values.Set(strings.TrimSpace(key), strings.TrimSpace(value))
	}
	if encoded := values.Encode(); encoded != "" {
		return basePath + "?" + encoded
	}
	return basePath
}
