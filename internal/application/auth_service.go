package application

import (
	"context"
	"fmt"
	"time"

	"github.com/yeecean/oneweb/internal/domain/operation"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// AuthService 编排交互式 OAuth 认证辅助流程。
type AuthService struct {
	OpService *OperationService
	CLI       *onedrive.CLIExecutor
}

// NewAuthService 创建认证服务。
func NewAuthService(opService *OperationService, cli *onedrive.CLIExecutor) *AuthService {
	return &AuthService{OpService: opService, CLI: cli}
}

// AuthResult 描述一次认证的结果。
type AuthResult struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	ExitCode int    `json:"exit_code"`
}

// StartAuth 启动认证操作：拉起 onedrive 子进程并监听授权 URL。
// 返回 Operation 与用于等待输入信号的 channel。
func (s *AuthService) StartAuth(ctx context.Context, profileID, confdir string) (*operation.Operation, error) {
	op, err := s.OpService.Create(profileID, operation.TypeAuth)
	if err != nil {
		return nil, err
	}
	if err := s.OpService.Start(op.ID); err != nil {
		return nil, err
	}

	handle, err := s.OpService.ProcessMgr.Run(ctx, s.CLI.BinaryPath, onedrive.ProcessOptions{
		Args:        []string{"--confdir=" + confdir},
		GracePeriod: 5 * time.Second,
	})
	if err != nil {
		_ = s.OpService.Fail(op.ID, map[string]interface{}{"error": err.Error()})
		return op, err
	}
	s.OpService.AttachProcess(op.ID, handle)

	parser := &onedrive.AuthParser{}
	go func() {
		for line := range handle.Stdout {
			s.OpService.AppendLog(op.ID, line, "stdout")
			if url, ok := parser.ParseAuthURL(line); ok {
				_ = s.OpService.TransitionToWaiting(op.ID, map[string]interface{}{
					"auth_url": url,
				})
			} else if code, ok := parser.ParseDeviceCode(line); ok {
				_ = s.OpService.TransitionToWaiting(op.ID, map[string]interface{}{
					"device_code": code,
				})
			}
		}
		for line := range handle.Stderr {
			s.OpService.AppendLog(op.ID, line, "stderr")
		}

		res := <-handle.Done
		current, _ := s.OpService.Get(op.ID)
		// 若已处于等待输入但进程结束（如设备码流程），标记失败
		if current != nil && current.Status == operation.OpWaitingForInput {
			_ = s.OpService.Fail(op.ID, map[string]interface{}{
				"auth_result": AuthResult{Success: false, Message: "process exited while waiting for input", ExitCode: res.ExitCode},
			})
			return
		}
		if res.Err != nil && res.ExitCode != 0 {
			_ = s.OpService.Fail(op.ID, map[string]interface{}{
				"auth_result": AuthResult{Success: false, Message: res.Err.Error(), ExitCode: res.ExitCode},
			})
			return
		}
		_ = s.OpService.Complete(op.ID, map[string]interface{}{
			"auth_result": AuthResult{Success: true, Message: "authentication completed", ExitCode: res.ExitCode},
		})
	}()

	return op, nil
}

// SubmitAuthInput 向等待输入的认证操作注入授权回调 URL。
func (s *AuthService) SubmitAuthInput(opID, callbackURL string) error {
	op, err := s.OpService.Get(opID)
	if err != nil {
		return err
	}
	if op.Status != operation.OpWaitingForInput {
		return ErrOperationNotWaiting
	}
	if err := s.OpService.SendInput(opID, callbackURL); err != nil {
		return err
	}
	return nil
}

// ValidateOperation 校验操作属于给定 Profile。
func (s *AuthService) ValidateOperation(opID, profileID string) error {
	op, err := s.OpService.Get(opID)
	if err != nil {
		return err
	}
	if op.ProfileID != profileID {
		return fmt.Errorf("application: operation %s does not belong to profile %s", opID, profileID)
	}
	return nil
}
