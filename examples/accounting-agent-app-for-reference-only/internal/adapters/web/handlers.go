package web

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"accounting-agent/internal/app"
	webui "accounting-agent/web"

	"github.com/go-chi/chi/v5"
)

// Handler holds the ApplicationService, the chi router, and the pending action store.
type Handler struct {
	svc        app.ApplicationService
	router     chi.Router
	pending    *pendingStore
	jwtSecret  string
	fileServer http.Handler
	uploadDir  string // directory for temporary attachment uploads
}

// NewHandler creates and wires the chi router with all routes.
func NewHandler(ctx context.Context, svc app.ApplicationService, allowedOrigins, jwtSecret string) http.Handler {
	staticFS, err := fs.Sub(webui.Static, "static")
	if err != nil {
		panic("web/static embed sub-FS failed: " + err.Error())
	}

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = filepath.Join(os.TempDir(), "accounting-agent-uploads")
	}
	if err := os.MkdirAll(uploadDir, 0700); err != nil {
		log.Fatalf("cannot create upload dir: %v", err)
	}

	h := &Handler{
		svc:        svc,
		pending:    newPendingStore(),
		jwtSecret:  jwtSecret,
		fileServer: http.FileServer(http.FS(staticFS)),
		uploadDir:  uploadDir,
	}

	// Start background maintenance goroutines.
	h.pending.startPurge(ctx)
	h.startUploadCleanup(ctx)

	r := chi.NewRouter()
	r.Use(RequestID)
	r.Use(Logger)
	r.Use(Recoverer)
	r.Use(CORS(allowedOrigins))

	// ── Health (public) ───────────────────────────────────────────────────────
	r.Get("/api/health", h.health)

	// ── Auth (public API) ─────────────────────────────────────────────────────
	r.Post("/api/auth/login", h.login)
	r.Post("/api/auth/logout", h.logout)

	// ── Static files served at /static/* ─────────────────────────────────────
	r.Get("/static/*", func(w http.ResponseWriter, req *http.Request) {
		http.StripPrefix("/static", h.fileServer).ServeHTTP(w, req)
	})

	// ── Browser login/logout/register (public HTML) ──────────────────────────
	r.Get("/login", h.loginPage)
	r.Post("/login", h.loginFormSubmit)
	r.Post("/logout", h.logoutPage)
	r.Get("/register", h.registerPage)
	r.Post("/register", h.registerFormSubmit)

	// ── Protected browser routes (redirect to /login if unauthenticated) ─────
	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuthBrowser)
		r.Get("/", h.chatHome) // WF5: chat home is the primary interface
		r.Get("/dashboard", h.dashboardPage)
		r.Get("/sales", h.salesHomePage)
		r.Get("/purchases", h.purchasesHomePage)
		r.Get("/inventory", h.inventoryHomePage)
		r.Get("/reports", h.reportsHomePage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings", h.settingsHomePage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/rules", h.settingsRulesPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/inventory", h.settingsInventoryPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/chart-of-accounts", h.settingsChartOfAccountsPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/customers", h.settingsCustomersPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/vendors", h.settingsVendorsPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/ai-agent", h.settingsAIAgentPage)
		// WF4 accounting screens
		r.Get("/reports/trial-balance", h.trialBalancePage)
		r.Get("/reports/pl", h.plReportPage)
		r.Get("/reports/balance-sheet", h.balanceSheetPage)
		r.Get("/reports/control-account-reconciliation", h.controlAccountReconciliationPage)
		r.Get("/reports/document-type-governance", h.documentTypeGovernancePage)
		r.Get("/reports/statement", h.accountStatementPage)
		r.Get("/accounting/journal-entry", h.journalEntryPage)
		// WD0 — Sales / Inventory pages
		r.Get("/sales/customers", h.customersListPage)
		r.Get("/sales/customers/{code}", notImplementedPage) // detail — WD0 follow-on
		r.Get("/sales/orders", h.ordersListPage)
		r.Get("/sales/orders/new", h.orderWizardPage)
		r.Post("/sales/orders/new", h.orderCreateAction)
		r.Get("/sales/orders/{ref}", h.orderDetailPage)
		r.Get("/inventory/products", h.productsListPage)
		r.Get("/inventory/stock", h.stockPage)
		// WD1 — Purchases pages
		r.Get("/purchases/vendors", h.vendorsListPage)
		r.Get("/purchases/vendors/new", h.vendorCreatePage)
		r.Post("/purchases/vendors/new", h.vendorCreateAction)
		r.Get("/purchases/orders", h.purchaseOrdersListPage)
		r.Get("/purchases/orders/new", h.poWizardPage)
		r.Post("/purchases/orders/new", h.poCreateAction)
		r.Get("/purchases/orders/{id}", h.poDetailPage)
		// Settings — user management (ADMIN only)
		r.With(h.RequireRoleBrowser("ADMIN")).Get("/settings/users", h.usersPage)
		r.With(h.RequireRoleBrowser("ADMIN")).Post("/settings/users", h.usersCreateAction)
		r.With(h.RequireRoleBrowser("ADMIN")).Post("/settings/users/{id}/role", h.usersUpdateRoleAction)
		r.With(h.RequireRoleBrowser("ADMIN")).Post("/settings/users/{id}/active", h.usersToggleActiveAction)
		// About
		r.Get("/about", h.aboutPage)
	})

	// ── Protected API routes (return 401 JSON if unauthenticated) ────────────
	r.Group(func(r chi.Router) {
		r.Use(h.RequireAuth)

		// File upload: body limit is managed inside the handler (multipart, up to 50 MB).
		r.Post("/chat/upload", h.chatUpload)

		// All other protected endpoints: 1 MB body limit to prevent unbounded request abuse.
		r.Group(func(r chi.Router) {
			r.Use(RequestBodyLimit(1 << 20)) // 1 MB

			// Auth
			r.Get("/api/auth/me", h.me)

			// Chat — primary endpoints (WF5)
			r.Post("/chat", h.chatMessage)
			r.With(h.RequireRole("FINANCE_MANAGER", "ADMIN")).Post("/chat/confirm", h.chatConfirm)
			r.Post("/chat/clear", h.chatClear)

			// Chat — legacy endpoints (kept for backward compat with old static frontend)
			r.Post("/api/chat/message", h.chatMessage)
			r.With(h.RequireRole("FINANCE_MANAGER", "ADMIN")).Post("/api/chat/confirm", h.chatConfirm)

			// ── Accounting (WF4) ──────────────────────────────────────────────────
			r.Get("/api/companies/{code}/trial-balance", h.apiTrialBalance)
			r.Get("/api/companies/{code}/accounts/{accountCode}/statement", h.apiAccountStatement)
			r.Get("/api/companies/{code}/reports/pl", h.apiProfitAndLoss)
			r.Get("/api/companies/{code}/reports/balance-sheet", h.apiBalanceSheet)
			r.Get("/api/companies/{code}/reports/control-account-reconciliation", h.apiControlAccountReconciliation)
			r.Get("/api/companies/{code}/reports/document-type-governance", h.apiDocumentTypeGovernance)
			r.Get("/api/companies/{code}/reports/control-account-journal-entries", h.apiManualJEControlAccountHits)
			r.Post("/api/companies/{code}/journal-entries", h.apiPostJournalEntry)
			r.Post("/api/companies/{code}/journal-entries/validate", h.apiValidateJournalEntry)

			// ── Sales (WD0) ───────────────────────────────────────────────────────
			r.Get("/api/companies/{code}/customers", h.apiListCustomers)
			r.Get("/api/companies/{code}/orders", h.apiListOrders)
			r.Post("/api/companies/{code}/orders", h.apiCreateOrder)
			r.Get("/api/companies/{code}/orders/{ref}", h.apiGetOrder)
			r.Post("/api/companies/{code}/orders/{ref}/confirm", h.apiConfirmOrder)
			r.Post("/api/companies/{code}/orders/{ref}/ship", h.apiShipOrder)
			r.Post("/api/companies/{code}/orders/{ref}/invoice", h.apiInvoiceOrder)
			r.Post("/api/companies/{code}/orders/{ref}/payment", h.apiPaymentOrder)

			// ── Inventory (WD0) ───────────────────────────────────────────────────
			r.Get("/api/companies/{code}/products", h.apiListProducts)
			r.Get("/api/companies/{code}/warehouses", notImplemented)
			r.Get("/api/companies/{code}/stock", notImplemented)
			r.Post("/api/companies/{code}/stock/receive", notImplemented)

			// ── Purchases (WD1) ──────────────────────────────────────────────────
			r.Get("/api/companies/{code}/vendors", h.apiListVendors)
			r.Post("/api/companies/{code}/vendors", h.apiCreateVendor)
			r.Get("/api/companies/{code}/vendors/{vendorCode}", h.apiGetVendor)
			r.Get("/api/companies/{code}/purchase-orders", h.apiListPurchaseOrders)
			r.Post("/api/companies/{code}/purchase-orders", h.apiCreatePurchaseOrder)
			r.Get("/api/companies/{code}/purchase-orders/{id}", h.apiGetPurchaseOrder)
			r.With(h.RequireRole("FINANCE_MANAGER", "ADMIN")).Post("/api/companies/{code}/purchase-orders/{id}/approve", h.apiApprovePO)
			r.Post("/api/companies/{code}/purchase-orders/{id}/receive", h.apiReceivePO)
			r.Post("/api/companies/{code}/purchase-orders/{id}/invoice", h.apiInvoicePO)
			r.Post("/api/companies/{code}/purchase-orders/{id}/pay", h.apiPayPO)
			r.With(h.RequireRole("FINANCE_MANAGER", "ADMIN")).Post("/api/companies/{code}/purchase-orders/{id}/close", h.apiClosePO)
			r.Post("/api/companies/{code}/vendor-invoices", h.apiCreateVendorInvoice)
			r.Post("/api/companies/{code}/vendor-invoices/{id}/pay", h.apiPayVendorInvoice)

			// ── Users (ADMIN only) ────────────────────────────────────────────────
			r.With(h.RequireRole("ADMIN")).Get("/api/companies/{code}/users", h.apiListUsers)
			r.With(h.RequireRole("ADMIN")).Post("/api/companies/{code}/users", h.apiCreateUser)

			// ── AI (legacy admin endpoints — company-scoped) ──────────────────────
			r.Post("/api/companies/{code}/ai/interpret", notImplemented)
			r.Post("/api/companies/{code}/ai/validate", notImplemented)
			r.Post("/api/companies/{code}/ai/commit", notImplemented)
		})
	})

	h.router = r
	return r
}

// health returns service status based on critical dependency checks.
func (h *Handler) health(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Status  string `json:"status"`
		Message string `json:"message,omitempty"`
	}

	if err := h.svc.HealthCheck(r.Context()); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		writeJSON(w, response{Status: "error", Message: "database unavailable"})
		return
	}

	writeJSON(w, response{Status: "ok"})
}

// companyCode extracts the {code} URL parameter.
func companyCode(r *http.Request) string {
	return chi.URLParam(r, "code")
}

// requireCompanyAccess verifies that the authenticated user belongs to the company
// identified by requestedCode. Returns false and writes a 403 response if not.
// Must only be called after RequireAuth or RequireAuthBrowser middleware.
func (h *Handler) requireCompanyAccess(w http.ResponseWriter, r *http.Request, requestedCode string) bool {
	claims := authFromContext(r.Context())
	if claims == nil {
		writeError(w, r, "authentication required", "UNAUTHORIZED", http.StatusUnauthorized)
		return false
	}
	if claims.CompanyCode != requestedCode {
		writeError(w, r, "access denied: company mismatch", "FORBIDDEN", http.StatusForbidden)
		return false
	}
	return true
}

// decodeJSON decodes the request body into v and returns false + writes an appropriate
// error response on failure. Returns HTTP 413 when the body exceeds the size limit set
// by RequestBodyLimit middleware; HTTP 400 for all other decode errors.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, r, "request body too large", "REQUEST_TOO_LARGE", http.StatusRequestEntityTooLarge)
			return false
		}
		writeError(w, r, "invalid JSON body: "+err.Error(), "BAD_REQUEST", http.StatusBadRequest)
		return false
	}
	return true
}
