package app

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"workflow_app/internal/identityaccess"
)

func (h *AgentAPIHandler) handleSessionLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionLoginPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}
	defer r.Body.Close()

	var req sessionLoginRequest
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}

	deviceLabel := strings.TrimSpace(req.DeviceLabel)
	if deviceLabel == "" {
		deviceLabel = "browser"
	}

	session, err := h.authService.StartBrowserSession(r.Context(), identityaccess.StartBrowserSessionInput{
		OrgSlug:     req.OrgSlug,
		Email:       req.Email,
		Password:    req.Password,
		DeviceLabel: deviceLabel,
		ExpiresAt:   time.Now().UTC().Add(browserSessionDuration),
	})
	if err != nil {
		switch {
		case errors.Is(err, identityaccess.ErrUnauthorized), errors.Is(err, identityaccess.ErrMembershipMissing):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid session credentials"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to start session"})
		}
		return
	}

	setSessionCookies(w, session.Session.ID, session.RefreshToken, session.Session.ExpiresAt)
	writeJSON(w, http.StatusCreated, mapSessionContext(identityaccess.SessionContext{
		Actor:           identityaccess.Actor{OrgID: session.Session.OrgID, UserID: session.Session.UserID, SessionID: session.Session.ID},
		Session:         session.Session,
		RoleCode:        session.RoleCode,
		OrgSlug:         session.OrgSlug,
		OrgName:         session.OrgName,
		UserEmail:       session.UserEmail,
		UserDisplayName: session.UserDisplayName,
	}))
}

func (h *AgentAPIHandler) handleSessionTokenLogin(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionTokenPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}
	defer r.Body.Close()

	var req sessionLoginRequest
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}

	deviceLabel := strings.TrimSpace(req.DeviceLabel)
	if deviceLabel == "" {
		deviceLabel = "non-browser"
	}

	session, err := h.authService.StartTokenSession(r.Context(), identityaccess.StartTokenSessionInput{
		OrgSlug:              req.OrgSlug,
		Email:                req.Email,
		Password:             req.Password,
		DeviceLabel:          deviceLabel,
		SessionExpiresAt:     time.Now().UTC().Add(browserSessionDuration),
		AccessTokenExpiresAt: time.Now().UTC().Add(accessTokenDuration),
	})
	if err != nil {
		switch {
		case errors.Is(err, identityaccess.ErrUnauthorized), errors.Is(err, identityaccess.ErrMembershipMissing):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid session credentials"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to start session"})
		}
		return
	}

	writeJSON(w, http.StatusCreated, mapTokenSession(session))
}

func (h *AgentAPIHandler) handleCurrentSession(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionCurrentPath {
		http.NotFound(w, r)
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}

	switch r.Method {
	case http.MethodGet:
		context, err := h.sessionContextFromRequest(r)
		if err != nil {
			if errors.Is(err, identityaccess.ErrUnauthorized) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
			return
		}
		if refreshToken := cookieValue(r, refreshTokenCookieName); refreshToken != "" {
			setSessionCookies(w, context.Session.ID, refreshToken, context.Session.ExpiresAt)
		}
		writeJSON(w, http.StatusOK, mapSessionContext(context))
	default:
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
	}
}

func (h *AgentAPIHandler) handleSessionRefresh(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionRefreshPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}
	defer r.Body.Close()

	var req sessionRefreshRequest
	if err := decodeJSONBody(r, &req, false); err != nil {
		writeJSONBodyError(w, err)
		return
	}

	session, err := h.authService.RefreshTokenSession(r.Context(), req.SessionID, req.RefreshToken, time.Now().UTC().Add(accessTokenDuration))
	if err != nil {
		switch {
		case errors.Is(err, identityaccess.ErrUnauthorized), errors.Is(err, identityaccess.ErrSessionNotActive):
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
		default:
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to refresh session"})
		}
		return
	}

	writeJSON(w, http.StatusOK, mapTokenSession(session))
}

func (h *AgentAPIHandler) handleSessionLogout(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != sessionLogoutPath {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		return
	}
	if h.authService == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "auth service unavailable"})
		return
	}

	switch {
	case bearerTokenFromRequest(r) != "":
		if err := h.authService.RevokeAccessTokenSession(r.Context(), bearerTokenFromRequest(r)); err != nil {
			if errors.Is(err, identityaccess.ErrUnauthorized) || errors.Is(err, identityaccess.ErrSessionNotActive) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to revoke session"})
			return
		}
	default:
		sessionID, refreshToken, ok := sessionCookiesFromRequest(r)
		if !ok {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
			return
		}
		if err := h.authService.RevokeAuthenticatedSession(r.Context(), sessionID, refreshToken); err != nil {
			if errors.Is(err, identityaccess.ErrUnauthorized) || errors.Is(err, identityaccess.ErrSessionNotActive) {
				writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "unauthorized"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "failed to revoke session"})
			return
		}
		clearSessionCookies(w)
	}

	writeJSON(w, http.StatusOK, struct {
		Revoked bool `json:"revoked"`
	}{Revoked: true})
}
