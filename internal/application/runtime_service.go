package application

import (
	"context"
	"fmt"

	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/domain/runtime"
)

// ErrInvalidAction 表示不允许的运行时控制动作。
var ErrInvalidAction = fmt.Errorf("application: invalid runtime action")

// RuntimeService 根据 Profile 的运行时类型选择后端执行控制。
type RuntimeService struct {
	Backends map[profile.RuntimeType]runtime.RuntimeBackend
}

// NewRuntimeService 创建运行时服务。
func NewRuntimeService(backends map[profile.RuntimeType]runtime.RuntimeBackend) *RuntimeService {
	return &RuntimeService{Backends: backends}
}

// backend 获取 Profile 对应的后端；缺失时返回错误。
func (s *RuntimeService) backend(p *profile.Profile) (runtime.RuntimeBackend, error) {
	b, ok := s.Backends[p.RuntimeType]
	if !ok {
		return nil, fmt.Errorf("application: no backend for runtime type %q", p.RuntimeType)
	}
	return b, nil
}

// GetStatus 查询 Profile 运行时状态。
func (s *RuntimeService) GetStatus(ctx context.Context, p *profile.Profile) (*runtime.RuntimeStatus, error) {
	b, err := s.backend(p)
	if err != nil {
		return nil, err
	}
	st, err := b.Status(ctx, *p)
	if err != nil {
		return nil, err
	}
	return &st, nil
}

// ControlRuntime 控制运行时启停。action 仅接受 start/stop/restart。
func (s *RuntimeService) ControlRuntime(ctx context.Context, p *profile.Profile, action string) error {
	b, err := s.backend(p)
	if err != nil {
		return err
	}
	switch action {
	case "start":
		return b.Start(ctx, *p)
	case "stop":
		return b.Stop(ctx, *p)
	case "restart":
		return b.Restart(ctx, *p)
	default:
		return ErrInvalidAction
	}
}

// StreamLogs 流式读取运行时日志。
func (s *RuntimeService) StreamLogs(ctx context.Context, p *profile.Profile, opts runtime.LogOptions) (<-chan runtime.LogEntry, error) {
	b, err := s.backend(p)
	if err != nil {
		return nil, err
	}
	return b.Logs(ctx, *p, opts)
}

// DetectBackends 探测各后端能力。
func (s *RuntimeService) DetectBackends(ctx context.Context) map[profile.RuntimeType]runtime.Capability {
	out := map[profile.RuntimeType]runtime.Capability{}
	for rt, b := range s.Backends {
		c, err := b.Detect(ctx)
		if err != nil {
			c.Available = false
		}
		out[rt] = c
	}
	return out
}
