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
var webSlice1TemplateFS embed.FS

var webSlice1Template = template.Must(template.New("slice1").Funcs(webTemplateFuncs()).ParseFS(webSlice1TemplateFS, "web_templates/*.tmpl"))

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

func templateRequestSummary(item any) string {
	switch v := item.(type) {
	case webOperationsFeedItem:
		return strings.TrimSpace(v.Summary)
	default:
		return ""
	}
}

func slice1TemplateName(data webPageData) string {
	switch {
	case data.ShowLogin:
		return "slice1_login"
	case data.Dashboard != nil:
		return "slice1_dashboard"
	case data.InboundSubmit != nil:
		return "slice1_submit"
	case data.OperationsFeed != nil:
		return "slice1_operations_feed"
	case data.AgentChat != nil:
		return "slice1_agent_chat"
	default:
		return ""
	}
}

func (h *AgentAPIHandler) renderWebPage(w http.ResponseWriter, data webPageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if templateName := slice1TemplateName(data); templateName != "" {
		_ = webSlice1Template.ExecuteTemplate(w, templateName, data)
		return
	}
	_ = webAppTemplate.Execute(w, data)
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
