package rest

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/profile"
)

// ConfigHandler 提供配置管理端点。
type ConfigHandler struct {
	Profiles *application.ProfileService
	Config   *application.ConfigService
}

// Get 处理 GET /api/v1/profiles/{id}/config。
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProfile(w, r)
	if err != nil {
		return
	}
	resp, err := h.Config.ReadConfig(p.ConfDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// Put 处理 PUT /api/v1/profiles/{id}/config。
func (h *ConfigHandler) Put(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProfile(w, r)
	if err != nil {
		return
	}
	var req application.SaveConfigRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body: "+err.Error(), nil)
		return
	}
	validation, err := h.Config.SaveConfig(r.Context(), p.ConfDir, req)
	if err != nil {
		if errors.Is(err, application.ErrConflict) {
			// 附上最新指纹帮助前端合并
			details := map[string]interface{}{}
			if resp, rerr := h.Config.ReadConfig(p.ConfDir); rerr == nil && resp.VersionMeta != nil {
				details["current_sha256"] = resp.VersionMeta.SHA256
			}
			writeError(w, http.StatusConflict, "CONFLICT", "config file has been modified externally", details)
			return
		}
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error(), map[string]interface{}{"warnings": validation})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "saved",
		"warnings": validation,
	})
}

// Validate 处理 POST /api/v1/profiles/{id}/config/validate。
func (h *ConfigHandler) Validate(w http.ResponseWriter, r *http.Request) {
	p, err := h.loadProfile(w, r)
	if err != nil {
		return
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "read body failed", nil)
		return
	}
	validation, err := h.Config.ValidateConfig(context.Background(), p.ConfDir, string(body))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, "VALIDATION", err.Error(), map[string]interface{}{"warnings": validation})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "valid", "warnings": validation})
}

// loadProfile 加载路径中的 Profile，失败时写出错误并返回 err。
func (h *ConfigHandler) loadProfile(w http.ResponseWriter, r *http.Request) (*profile.Profile, error) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return nil, err
	}
	return p, nil
}
