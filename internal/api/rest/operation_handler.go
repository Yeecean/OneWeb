package rest

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	wsgw "github.com/yeecean/oneweb/internal/api/websocket"
	"github.com/yeecean/oneweb/internal/application"
	"github.com/yeecean/oneweb/internal/domain/operation"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// OperationHandler 提供异步操作端点。
type OperationHandler struct {
	Profiles *application.ProfileService
	Ops      *application.OperationService
	Auth     *application.AuthService
	Gateway  *wsgw.Gateway
}

// Create 处理 POST /api/v1/profiles/{id}/operations。
func (h *OperationHandler) Create(w http.ResponseWriter, r *http.Request) {
	p, err := h.Profiles.GetProfile(chi.URLParam(r, "profileID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "profile not found", nil)
		return
	}
	var req struct {
		Type operation.OperationType `json:"type"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}
	switch req.Type {
	case operation.TypeAuth:
		op, err := h.Auth.StartAuth(r.Context(), p.ID, p.ConfDir)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
			return
		}
		writeJSON(w, http.StatusAccepted, op)
	case operation.TypeDryRun, operation.TypeSync, operation.TypeResync, operation.TypeValidate:
		op, err := h.Ops.Create(p.ID, req.Type)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
			return
		}
		if err := h.Ops.Start(op.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
			return
		}
		// 启动 CLI 子进程并跟踪完成（使用后台上下文，不受请求生命周期影响）
		go h.runCLIOperation(context.Background(), op.ID, p.ConfDir, req.Type)
		writeJSON(w, http.StatusAccepted, op)
	default:
		writeError(w, http.StatusBadRequest, "INVALID_TYPE", "operation type must be auth|sync|dry-run|resync|validate", nil)
	}
}

// List 处理 GET /api/v1/profiles/{id}/operations。
func (h *OperationHandler) List(w http.ResponseWriter, r *http.Request) {
	profileID := chi.URLParam(r, "profileID")
	ops, err := h.Ops.List(profileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, ops)
}

// Get 处理 GET /api/v1/operations/{id}。
func (h *OperationHandler) Get(w http.ResponseWriter, r *http.Request) {
	op, err := h.Ops.Get(chi.URLParam(r, "operationID"))
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "operation not found", nil)
		return
	}
	writeJSON(w, http.StatusOK, op)
}

// Input 处理 POST /api/v1/operations/{id}/input。
func (h *OperationHandler) Input(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_REQUEST", "invalid JSON body", nil)
		return
	}
	if err := h.Auth.SubmitAuthInput(chi.URLParam(r, "operationID"), req.URL); err != nil {
		if err == application.ErrOperationNotWaiting {
			writeError(w, http.StatusConflict, "NOT_WAITING", "operation is not waiting for input", nil)
			return
		}
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "input submitted"})
}

// Cancel 处理 POST /api/v1/operations/{id}/cancel。
func (h *OperationHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if err := h.Ops.Cancel(chi.URLParam(r, "operationID")); err != nil {
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error(), nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cancel requested"})
}

// runCLIOperation 在后台运行 onedrive CLI 子进程并跟踪操作完成。
func (h *OperationHandler) runCLIOperation(ctx context.Context, opID, confdir string, typ operation.OperationType) {
	if h.Ops.CLI == nil {
		_ = h.Ops.Fail(opID, map[string]interface{}{"error": "onedrive binary not available"})
		return
	}
	args := []string{"--confdir=" + confdir}
	switch typ {
	case operation.TypeDryRun:
		args = append(args, "--dry-run")
	case operation.TypeResync:
		args = append(args, "--resync", "--sync")
	case operation.TypeSync:
		args = append(args, "--sync")
	case operation.TypeValidate:
		args = append(args, "--display-config")
	}
	handle, err := h.Ops.ProcessMgr.Run(ctx, h.Ops.CLI.BinaryPath, onedrive.ProcessOptions{
		Args:        args,
		GracePeriod: 5 * time.Second,
	})
	if err != nil {
		_ = h.Ops.Fail(opID, map[string]interface{}{"error": err.Error()})
		return
	}
	h.Ops.AttachProcess(opID, handle)
	go func() {
		for line := range handle.Stdout {
			h.Ops.AppendLog(opID, line, "stdout")
		}
		for line := range handle.Stderr {
			h.Ops.AppendLog(opID, line, "stderr")
		}
		res := <-handle.Done
		if res.ExitCode != 0 {
			_ = h.Ops.Fail(opID, map[string]interface{}{
				"error":     "CLI exited with non-zero code",
				"exit_code": res.ExitCode,
			})
			return
		}
		_ = h.Ops.Complete(opID, map[string]interface{}{
			"exit_code": res.ExitCode,
		})
	}()
}

// OperationWS 处理 /ws/v1/operations/{id}。
func (h *OperationHandler) OperationWS(w http.ResponseWriter, r *http.Request) {
	h.Gateway.HandleOperationWS(w, r)
}
