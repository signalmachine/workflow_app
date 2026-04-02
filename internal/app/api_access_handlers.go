package app

import (
	"net/http"
	"strings"

	"workflow_app/internal/identityaccess"
)

func (h *AgentAPIHandler) handleAdminAccessUsers(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != adminAccessUsersPath {
		http.NotFound(w, r)
		return
	}
	if h.accessAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "access admin service unavailable"})
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := h.accessAdmin.ListOrgUsers(r.Context(), identityaccess.ListOrgUsersInput{Actor: actor})
		if err != nil {
			handleAccessAdminError(w, err, "failed to list org users")
			return
		}
		response := struct {
			Items []orgUserMembershipResponse `json:"items"`
		}{Items: make([]orgUserMembershipResponse, 0, len(items))}
		for _, item := range items {
			response.Items = append(response.Items, mapOrgUserMembership(item))
		}
		writeJSON(w, http.StatusOK, response)
	case http.MethodPost:
		var req provisionOrgUserRequest
		if err := decodeJSONBody(r, &req, false); err != nil {
			writeJSONBodyError(w, err)
			return
		}
		item, err := h.accessAdmin.ProvisionOrgUser(r.Context(), identityaccess.ProvisionOrgUserInput{
			Email:       strings.TrimSpace(req.Email),
			DisplayName: strings.TrimSpace(req.DisplayName),
			RoleCode:    strings.TrimSpace(req.RoleCode),
			Password:    strings.TrimSpace(req.Password),
			Actor:       actor,
		})
		if err != nil {
			handleAccessAdminError(w, err, "failed to provision org user")
			return
		}
		writeJSON(w, http.StatusCreated, mapOrgUserMembership(item))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleAdminAccessMembershipAction(w http.ResponseWriter, r *http.Request) {
	if h.accessAdmin == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "access admin service unavailable"})
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}

	membershipID, action, ok := parseChildActionPath(adminAccessUsersPath, r.URL.Path)
	if !ok || !strings.EqualFold(strings.TrimSpace(action), "role") {
		http.NotFound(w, r)
		return
	}

	actor, err := h.adminActorFromRequest(r)
	if err != nil {
		writeAdminActorError(w, err)
		return
	}

	var req updateMembershipRoleRequest
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}

	item, err := h.accessAdmin.UpdateMembershipRole(r.Context(), identityaccess.UpdateMembershipRoleInput{
		MembershipID: membershipID,
		RoleCode:     strings.TrimSpace(req.RoleCode),
		Actor:        actor,
	})
	if err != nil {
		handleAccessAdminError(w, err, "failed to update membership role")
		return
	}
	writeJSON(w, http.StatusOK, mapOrgUserMembership(item))
}
