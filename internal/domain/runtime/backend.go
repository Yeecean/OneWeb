// Package runtime 定义运行时控制面的领域抽象。
// 此文件只定义接口，不依赖任何具体实现包（架构红线 5）。
package runtime

import (
	"context"
	"time"

	"github.com/yeecean/oneweb/internal/domain/profile"
)

// RuntimeState 是运行时状态的宽泛枚举。
type RuntimeState string

const (
	StateRunning  RuntimeState = "running"
	StateStarting RuntimeState = "starting"
	StateStopped  RuntimeState = "stopped"
	StateFailed   RuntimeState = "failed"
	StateInactive RuntimeState = "inactive"
)

// RuntimeStatus 描述守护进程的运行状态快照。
type RuntimeStatus struct {
	State       RuntimeState `json:"state"`
	SubState    string       `json:"sub_state,omitempty"`
	PID         int          `json:"pid,omitempty"`
	StartedAt   time.Time    `json:"started_at,omitempty"`
	MemoryBytes uint64       `json:"memory_bytes,omitempty"`
}

// LogStream 标识日志来源流。
type LogStream string

const (
	StreamStdout LogStream = "stdout"
	StreamStderr LogStream = "stderr"
)

// LogEntry 表示一条运行时日志。
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Stream    LogStream `json:"stream"`
	Level     string    `json:"level,omitempty"`
	Message   string    `json:"message"`
}

// LogOptions 描述日志读取参数。
type LogOptions struct {
	Since  *time.Time
	Follow bool
	Lines  int
}

// Capability 描述运行时后端的能力探测结果。
type Capability struct {
	Available bool     `json:"available"`
	Version   string   `json:"version,omitempty"`
	Features  []string `json:"features,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

// RuntimeBackend 抽象不同运行时后端的统一控制接口。
type RuntimeBackend interface {
	Detect(ctx context.Context) (Capability, error)
	Start(ctx context.Context, p profile.Profile) error
	Stop(ctx context.Context, p profile.Profile) error
	Restart(ctx context.Context, p profile.Profile) error
	Status(ctx context.Context, p profile.Profile) (RuntimeStatus, error)
	Logs(ctx context.Context, p profile.Profile, opts LogOptions) (<-chan LogEntry, error)
}
