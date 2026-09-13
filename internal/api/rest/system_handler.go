package rest

import (
	"context"
	"net/http"

	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/capability"
	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
	"github.com/yeecean/oneweb/internal/infrastructure/runtime/systemd"
)

// SystemInfoResponse 描述宿主系统与运行时能力信息。
type SystemInfoResponse struct {
	Version        string                                    `json:"version"`
	Hostname       string                                    `json:"hostname"`
	OnedriveBinary string                                    `json:"onedrive_binary,omitempty"`
	ClientVersion  string                                    `json:"client_version,omitempty"`
	Capabilities   capability.FeatureSupport                 `json:"capabilities,omitempty"`
	Runtimes       map[profile.RuntimeType]runtimeCapability `json:"runtimes"`
}

type runtimeCapability struct {
	Available bool     `json:"available"`
	Version   string   `json:"version,omitempty"`
	Features  []string `json:"features,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// SystemHandler 提供系统级端点。
type SystemHandler struct {
	Version    string
	CLI        *onedrive.CLIExecutor
	RuntimeSvc *application.RuntimeService
}

// Info 处理 GET /api/v1/system/info。
func (h *SystemHandler) Info(w http.ResponseWriter, r *http.Request) {
	resp := SystemInfoResponse{
		Version:  h.Version,
		Runtimes: map[profile.RuntimeType]runtimeCapability{},
	}

	// 探测各运行时后端
	if h.RuntimeSvc != nil {
		detected := h.RuntimeSvc.DetectBackends(r.Context())
		for rt, cap := range detected {
			resp.Runtimes[rt] = runtimeCapability{
				Available: cap.Available,
				Version:   cap.Version,
				Features:  cap.Features,
				Warnings:  cap.Warnings,
			}
		}
	}
	// 兼容探测 systemd 后端（当未注入 RuntimeService 时）
	if _, ok := resp.Runtimes[profile.RuntimeSystemd]; !ok {
		if cap, err := (&systemd.Backend{}).Detect(r.Context()); err == nil {
			resp.Runtimes[profile.RuntimeSystemd] = runtimeCapability{Available: cap.Available, Version: cap.Version, Features: cap.Features, Warnings: cap.Warnings}
		}
	}

	// 探测 onedrive 客户端版本与能力
	if h.CLI != nil {
		if v, err := h.CLI.Version(r.Context()); err == nil {
			resp.ClientVersion = v
			if fs, err := capability.CapabilitiesOf(v); err == nil {
				resp.Capabilities = fs
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// Capabilities 处理 GET /api/v1/system/capabilities。
func (h *SystemHandler) Capabilities(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	resp := map[string]interface{}{
		"client":   map[string]interface{}{"available": false},
		"runtimes": map[string]interface{}{},
	}
	if h.CLI != nil {
		if v, err := h.CLI.Version(ctx); err == nil {
			fs, _ := capability.CapabilitiesOf(v)
			resp["client"] = map[string]interface{}{
				"available": true,
				"version":   v,
				"features":  fs,
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
}
