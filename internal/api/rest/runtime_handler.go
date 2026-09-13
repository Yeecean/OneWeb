package rest

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/runtime"
)

// RuntimeHandler 提供运行时生命周期控制端点。
type RuntimeHandler struct {
	Profiles *application.ProfileService
	Runtime  *application.RuntimeService
}

// GetStatus 处理 GET /api/v1/profiles/{id}/runtime。
func (h *RuntimeHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	st, err := h.Runtime.GetStatus(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

// Action 处理 POST /api/v1/profiles/{id}/runtime/actions。
func (h *RuntimeHandler) Action(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}
	if err := h.Runtime.ControlRuntime(r.Context(), p, req.Action); err != nil {
		if err == application.ErrInvalidAction {
			writeError(w, http.StatusBadRequest, "INVALID_ACTION", "action must be start|stop|restart", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	st, _ := h.Runtime.GetStatus(r.Context(), p)
	writeJSON(w, http.StatusOK, st)
}

// Logs 处理 GET /api/v1/profiles/{id}/runtime/logs?follow=true&lines=50。
func (h *RuntimeHandler) Logs(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	q := r.URL.Query()
	opts := runtime.LogOptions{Follow: q.Get("follow") == "true"}
	if lines := parseIntSafe(q.Get("lines")); lines > 0 {
		opts.Lines = lines
	}
	ch, err := h.Runtime.StreamLogs(r.Context(), p, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	for entry := range ch {
		_ = enc.Encode(entry)
	}
}

func parseIntSafe(s string) int {
	n := 0
	if s == "" {
		return 0
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}
