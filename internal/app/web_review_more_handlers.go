package app

import (
	"net/http"
	"net/url"
	"strings"

	"workflow_app/internal/reporting"
)

func (h *AgentAPIHandler) handleWebApprovals(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webApprovalsPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webApprovalsData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		ApprovalID: strings.TrimSpace(r.URL.Query().Get("approval_id")),
		Status:     strings.TrimSpace(r.URL.Query().Get("status")),
		QueueCode:  strings.TrimSpace(r.URL.Query().Get("queue_code")),
	}
	data.Approvals, err = h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: data.ApprovalID,
		Status:     data.Status,
		QueueCode:  data.QueueCode,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load approval queue"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webApprovalsPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Approvals:  &data,
	})
}

func (h *AgentAPIHandler) handleWebApprovalDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	approvalID, ok := parseChildPath(webApprovalsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	entries, err := h.reviewService.ListApprovalQueue(r.Context(), reporting.ListApprovalQueueInput{
		ApprovalID: approvalID,
		Limit:      2,
		Actor:      sessionContext.Actor,
	})
	if err != nil || len(entries) == 0 {
		http.Redirect(w, r, webApprovalsPath+"?error="+url.QueryEscape("failed to load approval"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webApprovalsPath,
		Session:    &sessionContext,
		ApprovalDetail: &webApprovalDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Entry:   entries[0],
		},
	})
}

func (h *AgentAPIHandler) handleWebInventory(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webInventoryPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webInventoryData{
		Session:               sessionContext,
		Notice:                strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:                 strings.TrimSpace(r.URL.Query().Get("error")),
		MovementID:            strings.TrimSpace(r.URL.Query().Get("movement_id")),
		ItemID:                strings.TrimSpace(r.URL.Query().Get("item_id")),
		LocationID:            strings.TrimSpace(r.URL.Query().Get("location_id")),
		DocumentID:            strings.TrimSpace(r.URL.Query().Get("document_id")),
		MovementType:          strings.TrimSpace(r.URL.Query().Get("movement_type")),
		OnlyPendingAccounting: strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_accounting")), "true"),
		OnlyPendingExecution:  strings.EqualFold(strings.TrimSpace(r.URL.Query().Get("only_pending_execution")), "true"),
	}

	data.Stock, err = h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		ItemID:     data.ItemID,
		LocationID: data.LocationID,
		Limit:      50,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load inventory stock"
	}
	if data.Movements, err = h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		MovementID:   data.MovementID,
		ItemID:       data.ItemID,
		LocationID:   data.LocationID,
		DocumentID:   data.DocumentID,
		MovementType: data.MovementType,
		Limit:        50,
		Actor:        sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load inventory movements"
	}
	if data.Reconciliation, err = h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		MovementID:            data.MovementID,
		ItemID:                data.ItemID,
		DocumentID:            data.DocumentID,
		OnlyPendingAccounting: data.OnlyPendingAccounting,
		OnlyPendingExecution:  data.OnlyPendingExecution,
		Limit:                 50,
		Actor:                 sessionContext.Actor,
	}); err != nil && data.Error == "" {
		data.Error = "failed to load inventory reconciliation"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInventoryPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Inventory:  &data,
	})
}

func (h *AgentAPIHandler) handleWebInventoryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	movementID, ok := parseChildPath(webInventoryPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	movements, err := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		MovementID: movementID,
		Limit:      2,
		Actor:      sessionContext.Actor,
	})
	if err != nil || len(movements) == 0 {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load inventory movement"), http.StatusSeeOther)
		return
	}

	reconciliation, err := h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		MovementID: movementID,
		Limit:      20,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load inventory reconciliation"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInventoryPath,
		Session:    &sessionContext,
		InventoryDetail: &webInventoryDetailData{
			Session:        sessionContext,
			Notice:         strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:          strings.TrimSpace(r.URL.Query().Get("error")),
			Review:         movements[0],
			Reconciliation: reconciliation,
		},
	})
}

func (h *AgentAPIHandler) handleWebInventoryItemDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	itemID, ok := parseChildPath(webInventoryItemsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	stock, err := h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		ItemID: itemID,
		Limit:  100,
		Actor:  sessionContext.Actor,
	})
	if err != nil || len(stock) == 0 {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load inventory item"), http.StatusSeeOther)
		return
	}

	movements, err := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		ItemID: itemID,
		Limit:  100,
		Actor:  sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load item movements"), http.StatusSeeOther)
		return
	}

	reconciliation, err := h.reviewService.ListInventoryReconciliation(r.Context(), reporting.ListInventoryReconciliationInput{
		ItemID: itemID,
		Limit:  100,
		Actor:  sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load item reconciliation"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInventoryPath,
		Session:    &sessionContext,
		InventoryItemDetail: &webInventoryItemDetailData{
			Session:        sessionContext,
			Notice:         strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:          strings.TrimSpace(r.URL.Query().Get("error")),
			ItemID:         itemID,
			ItemSKU:        stock[0].ItemSKU,
			ItemName:       stock[0].ItemName,
			ItemRole:       stock[0].ItemRole,
			Stock:          stock,
			Movements:      movements,
			Reconciliation: reconciliation,
		},
	})
}

func (h *AgentAPIHandler) handleWebInventoryLocationDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	locationID, ok := parseChildPath(webInventoryLocationsPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	stock, err := h.reviewService.ListInventoryStock(r.Context(), reporting.ListInventoryStockInput{
		LocationID: locationID,
		Limit:      100,
		Actor:      sessionContext.Actor,
	})
	if err != nil || len(stock) == 0 {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load inventory location"), http.StatusSeeOther)
		return
	}

	movements, err := h.reviewService.ListInventoryMovements(r.Context(), reporting.ListInventoryMovementsInput{
		LocationID: locationID,
		Limit:      100,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webInventoryPath+"?error="+url.QueryEscape("failed to load location movements"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webInventoryPath,
		Session:    &sessionContext,
		InventoryLocationDetail: &webInventoryLocationDetailData{
			Session:      sessionContext,
			Notice:       strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:        strings.TrimSpace(r.URL.Query().Get("error")),
			LocationID:   locationID,
			LocationCode: stock[0].LocationCode,
			LocationName: stock[0].LocationName,
			LocationRole: stock[0].LocationRole,
			Stock:        stock,
			Movements:    movements,
		},
	})
}

func (h *AgentAPIHandler) handleWebWorkOrders(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webWorkOrdersPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webWorkOrdersData{
		Session:     sessionContext,
		Notice:      strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:       strings.TrimSpace(r.URL.Query().Get("error")),
		WorkOrderID: strings.TrimSpace(r.URL.Query().Get("work_order_id")),
		Status:      strings.TrimSpace(r.URL.Query().Get("status")),
		DocumentID:  strings.TrimSpace(r.URL.Query().Get("document_id")),
	}
	data.WorkOrders, err = h.reviewService.ListWorkOrders(r.Context(), reporting.ListWorkOrdersInput{
		WorkOrderID: data.WorkOrderID,
		Status:      data.Status,
		DocumentID:  data.DocumentID,
		Limit:       50,
		Actor:       sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load work orders"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webWorkOrdersPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		WorkOrders: &data,
	})
}

func (h *AgentAPIHandler) handleWebWorkOrderDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}
	workOrderID, ok := parseChildPath(webWorkOrdersPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	review, err := h.reviewService.GetWorkOrderReview(r.Context(), reporting.GetWorkOrderReviewInput{
		WorkOrderID: workOrderID,
		Actor:       sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, webWorkOrdersPath+"?error="+url.QueryEscape("failed to load work order"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webWorkOrdersPath,
		Session:    &sessionContext,
		WorkOrderDetail: &webWorkOrderDetailData{
			Session: sessionContext,
			Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
			Error:   strings.TrimSpace(r.URL.Query().Get("error")),
			Review:  review,
		},
	})
}

func (h *AgentAPIHandler) handleWebAudit(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAuditPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	data := webAuditData{
		Session:    sessionContext,
		Notice:     strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:      strings.TrimSpace(r.URL.Query().Get("error")),
		EventID:    strings.TrimSpace(r.URL.Query().Get("event_id")),
		EntityType: strings.TrimSpace(r.URL.Query().Get("entity_type")),
		EntityID:   strings.TrimSpace(r.URL.Query().Get("entity_id")),
	}
	data.Events, err = h.reviewService.LookupAuditEvents(r.Context(), reporting.LookupAuditEventsInput{
		EventID:    data.EventID,
		EntityType: data.EntityType,
		EntityID:   data.EntityID,
		Limit:      100,
		Actor:      sessionContext.Actor,
	})
	if err != nil {
		data.Error = "failed to load audit events"
	}

	h.renderWebPage(w, webPageData{
		Title:      "workflow_app",
		ActivePath: webAuditPath,
		Notice:     data.Notice,
		Error:      data.Error,
		Session:    &sessionContext,
		Audit:      &data,
	})
}

func (h *AgentAPIHandler) handleWebAuditDetail(w http.ResponseWriter, r *http.Request) {
	eventID, ok := parseChildPath(webAuditPath, r.URL.Path)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, err := h.sessionContextFromRequest(r)
	if err != nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("Please sign in."), http.StatusSeeOther)
		return
	}
	if h.reviewService == nil {
		http.Redirect(w, r, webAppPath+"?error="+url.QueryEscape("review service unavailable"), http.StatusSeeOther)
		return
	}

	events, err := h.reviewService.LookupAuditEvents(r.Context(), reporting.LookupAuditEventsInput{
		EventID: eventID,
		Limit:   1,
		Actor:   sessionContext.Actor,
	})
	if err != nil || len(events) == 0 {
		http.Redirect(w, r, webAuditPath+"?error="+url.QueryEscape("failed to load audit event"), http.StatusSeeOther)
		return
	}

	h.renderWebPage(w, webPageData{
		Title:       "workflow_app",
		ActivePath:  webAuditPath,
		Session:     &sessionContext,
		AuditDetail: &webAuditDetailData{Session: sessionContext, Event: events[0]},
	})
}
