package rest

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	wsgw "github.com/yeecean/oneweb/internal/api/websocket"
	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// Dependencies 聚合 REST 层所需的全部服务依赖。
type Dependencies struct {
	Version      string
	ProfileSvc   *application.ProfileService
	ConfigSvc    *application.ConfigService
	RuntimeSvc   *application.RuntimeService
	OperationSvc *application.OperationService
	AuthSvc      *application.AuthService
	SyncSvc      *application.SyncService
	WSGateway    *wsgw.Gateway
	CLI          *onedrive.CLIExecutor
}

// NewRouter 构建 chi 路由。
func NewRouter(deps *Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	systemH := &SystemHandler{Version: deps.Version, CLI: deps.CLI, RuntimeSvc: deps.RuntimeSvc}
	profileH := &ProfileHandler{Svc: deps.ProfileSvc}
	configH := &ConfigHandler{Profiles: deps.ProfileSvc, Config: deps.ConfigSvc}
	runtimeH := &RuntimeHandler{Profiles: deps.ProfileSvc, Runtime: deps.RuntimeSvc}
	opH := &OperationHandler{
		Profiles: deps.ProfileSvc,
		Ops:      deps.OperationSvc,
		Auth:     deps.AuthSvc,
		Gateway:  deps.WSGateway,
	}
	syncH := &SyncListHandler{Profiles: deps.ProfileSvc, Sync: deps.SyncSvc}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/system/info", systemH.Info)
		r.Get("/system/capabilities", systemH.Capabilities)
		r.Route("/profiles", func(r chi.Router) {
			r.Get("/", profileH.List)
			r.Post("/", profileH.Create)
			r.Route("/{profileID}", func(r chi.Router) {
				r.Get("/", profileH.Get)
				r.Patch("/", profileH.Update)
				r.Delete("/", profileH.Delete)
				r.Get("/config", configH.Get)
				r.Put("/config", configH.Put)
				r.Post("/config/validate", configH.Validate)
				r.Get("/runtime", runtimeH.GetStatus)
				r.Post("/runtime/actions", runtimeH.Action)
				r.Get("/runtime/logs", runtimeH.Logs)
				r.Get("/operations", opH.List)
				r.Post("/operations", opH.Create)
				r.Get("/sync-list", syncH.Get)
				r.Put("/sync-list", syncH.Put)
			})
		})
		r.Route("/operations", func(r chi.Router) {
			r.Route("/{operationID}", func(r chi.Router) {
				r.Get("/", opH.Get)
				r.Post("/input", opH.Input)
				r.Post("/cancel", opH.Cancel)
			})
		})
	})

	// WebSocket 网关
	if deps.WSGateway != nil {
		r.HandleFunc("/ws/v1/operations/{operationID}", deps.WSGateway.HandleOperationWS)
	}

	// SPA 静态资源（前端构建产物通过 embed 打包）
	r.NotFound(func(w http.ResponseWriter, req *http.Request) {
		// API 路径未命中时返回 JSON 404，其余交给 SPA index.html
		if strings.HasPrefix(req.URL.Path, "/api/") || strings.HasPrefix(req.URL.Path, "/ws/") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "resource not found", nil)
			return
		}
		staticHandler().ServeHTTP(w, req)
	})
	return r
}
