package app

import (
	"net/http"
	"strings"

	"workflow_app/internal/identityaccess"
)

func accessAdminWebErrorMessage(err error, fallback string) string {
	switch {
	case err == nil:
		return ""
	case err == identityaccess.ErrUnauthorized:
		return "unauthorized"
	case err == identityaccess.ErrInvalidUser:
		return "invalid user"
	case err == identityaccess.ErrInvalidMembership:
		return "invalid membership"
	case err == identityaccess.ErrUserNotFound:
		return "user not found"
	case err == identityaccess.ErrMembershipNotFound:
		return "membership not found"
	case err == identityaccess.ErrProtectedMembership:
		return "membership update would remove current admin access"
	default:
		return fallback
	}
}

func adminAccessPathWithMessage(key, message string) string {
	if strings.TrimSpace(message) == "" {
		return webAdminAccessPath
	}
	return appendWebMessage(webAdminAccessPath, key, message)
}

func (h *AgentAPIHandler) handleWebAdminAccess(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminAccessPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accessAdmin != nil, "access admin service unavailable")
	if !ok {
		return
	}

	data := webAdminAccessData{
		Session: sessionContext,
		Notice:  strings.TrimSpace(r.URL.Query().Get("notice")),
		Error:   strings.TrimSpace(r.URL.Query().Get("error")),
		RoleOptions: []string{
			identityaccess.RoleAdmin,
			identityaccess.RoleOperator,
			identityaccess.RoleApprover,
		},
	}

	items, err := h.accessAdmin.ListOrgUsers(r.Context(), identityaccess.ListOrgUsersInput{Actor: sessionContext.Actor})
	if err != nil {
		data.Error = accessAdminWebErrorMessage(err, "failed to load org users")
	} else {
		data.Users = items
	}

	h.renderWebPage(w, webPageData{
		Title:       "workflow_app",
		ActivePath:  webAdminPath,
		Notice:      data.Notice,
		Error:       data.Error,
		Session:     &sessionContext,
		AdminAccess: &data,
	})
}

func (h *AgentAPIHandler) handleWebAdminAccessUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != webAdminAccessUsersPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accessAdmin != nil, "access admin service unavailable")
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, adminAccessPathWithMessage("error", "invalid user form"), http.StatusSeeOther)
		return
	}

	_, err := h.accessAdmin.ProvisionOrgUser(r.Context(), identityaccess.ProvisionOrgUserInput{
		Email:       strings.TrimSpace(r.FormValue("email")),
		DisplayName: strings.TrimSpace(r.FormValue("display_name")),
		RoleCode:    strings.TrimSpace(r.FormValue("role_code")),
		Password:    strings.TrimSpace(r.FormValue("password")),
		Actor:       sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminAccessPathWithMessage("error", accessAdminWebErrorMessage(err, "failed to provision org user")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccessPathWithMessage("notice", "Access membership saved."), http.StatusSeeOther)
}

func (h *AgentAPIHandler) handleWebAdminMembershipAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}

	membershipID, action, ok := parseChildActionPath(webAdminAccessUsersPath, r.URL.Path)
	if !ok || !strings.EqualFold(strings.TrimSpace(action), "role") {
		http.NotFound(w, r)
		return
	}

	sessionContext, ok := h.requireWebAdminSessionWithService(w, r, h.accessAdmin != nil, "access admin service unavailable")
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, adminAccessPathWithMessage("error", "invalid membership form"), http.StatusSeeOther)
		return
	}

	_, err := h.accessAdmin.UpdateMembershipRole(r.Context(), identityaccess.UpdateMembershipRoleInput{
		MembershipID: membershipID,
		RoleCode:     strings.TrimSpace(r.FormValue("role_code")),
		Actor:        sessionContext.Actor,
	})
	if err != nil {
		http.Redirect(w, r, adminAccessPathWithMessage("error", accessAdminWebErrorMessage(err, "failed to update membership role")), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, adminAccessPathWithMessage("notice", "Membership role updated."), http.StatusSeeOther)
}
