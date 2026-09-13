package rest

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yeecean/oneweb/internal/application"
)

// SyncListHandler 提供选择性同步规则端点。
type SyncListHandler struct {
	Profiles *application.ProfileService
	Sync     *application.SyncService
}

// Get 处理 GET /api/v1/profiles/{id}/sync-list。
func (h *SyncListHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	resp, err := h.Sync.GetSyncList(p.ConfDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Put 处理 PUT /api/v1/profiles/{id}/sync-list。
func (h *SyncListHandler) Put(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	var req application.SaveSyncListRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body: "+err.Error(), nil)
		return
	}
	resp, err := h.Sync.SaveSyncList(p.ConfDir, req)
	if err != nil {
		if errors.Is(err, application.ErrSyncListConflict) {
			writeError(w, http.StatusConflict, "CONFLICT", err.Error(), nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
