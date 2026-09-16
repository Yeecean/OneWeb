package rest

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/profile"
)

// ProfileHandler 提供 Profile CRUD 端点。
type ProfileHandler struct {
	Svc *application.ProfileService
}

// List 处理 GET /api/v1/profiles。
func (h *ProfileHandler) List(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.Svc.ListProfiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

// Discover 处理 POST /api/v1/profiles/discover，重新扫描宿主机。
func (h *ProfileHandler) Discover(w http.ResponseWriter, r *http.Request) {
	profiles, err := h.Svc.AutoDiscoverProfiles()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, profiles)
}

// Create 处理 POST /api/v1/profiles。
func (h *ProfileHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req application.CreateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body: "+err.Error(), nil)
		return
	}
	p, err := h.Svc.CreateProfile(req)
	if err != nil {
		switch {
		case errors.Is(err, profile.ErrEmptyID), errors.Is(err, profile.ErrInvalidID),
			errors.Is(err, profile.ErrEmptyConfDir), errors.Is(err, profile.ErrInvalidRuntime):
			writeError(w, http.StatusBadRequest, "INVALID_PROFILE", err.Error(), nil)
		case errors.Is(err, application.ErrConfDirNotExist):
			writeError(w, http.StatusBadRequest, "INVALID_CONFDIR", err.Error(), nil)
		default:
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		}
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

// Get 处理 GET /api/v1/profiles/{id}。
func (h *ProfileHandler) Get(w http.ResponseWriter, r *http.Request) {
	p, err := h.Svc.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Update 处理 PATCH /api/v1/profiles/{id}。
func (h *ProfileHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req application.UpdateProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body: "+err.Error(), nil)
		return
	}
	p, err := h.Svc.UpdateProfile(chi.URLParam(r, "profileID"), req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, p)
}

// Delete 处理 DELETE /api/v1/profiles/{id}。
func (h *ProfileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.Svc.DeleteProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
