// Package systemd 实现基于 systemd --user 的运行时后端。
// 所有操作均以非特权用户身份执行，绝不调用 sudo（架构红线 5）。
package systemd

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/domain/runtime"
	"github.com/yeecean/oneweb/internal/infrastructure/journal"
)

// Backend 实现 runtime.RuntimeBackend，通过 systemctl --user 驱动。
type Backend struct {
	// UnitResolver 将 Profile 映射为 systemd unit 名；默认使用 RuntimeTarget。
	UnitResolver func(p profile.Profile) string
}

// unit 解析 Profile 对应的 systemd unit 名。
func (b *Backend) unit(p profile.Profile) string {
	if b.UnitResolver != nil {
		return b.UnitResolver(p)
	}
	if p.RuntimeTarget != "" {
		return p.RuntimeTarget
	}
	return "onedrive@" + p.ID + ".service"
}

// Detect 探测 systemd user 环境可用性与 linger 状态。
func (b *Backend) Detect(ctx context.Context) (runtime.Capability, error) {
	cap := runtime.Capability{}
	out, err := exec.CommandContext(ctx, "systemctl", "--user", "--version").Output()
	if err != nil {
		cap.Available = false
		return cap, nil
	}
	cap.Available = true
	cap.Version = strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	cap.Features = append(cap.Features, "user-systemd")

	if linger, err := DetectLinger(currentUser()); err == nil && linger {
		cap.Features = append(cap.Features, "linger-enabled")
	} else if err == nil {
		cap.Warnings = append(cap.Warnings, "user linger not enabled; service may stop after logout")
	}
	return cap, nil
}

// Status 查询 unit 的实时状态。
func (b *Backend) Status(ctx context.Context, p profile.Profile) (runtime.RuntimeStatus, error) {
	unit := b.unit(p)
	out, err := exec.CommandContext(ctx, "systemctl", "--user", "show", unit,
		"--property=ActiveState,SubState,MainPID,ExecMainStartTimestamp,MemoryCurrent").Output()
	if err != nil {
		return runtime.RuntimeStatus{}, fmt.Errorf("systemd: status %s: %w", unit, err)
	}
	return parseStatus(string(out))
}

// Start 启动 unit。
func (b *Backend) Start(ctx context.Context, p profile.Profile) error {
	return b.action(ctx, p, "start")
}

// Stop 停止 unit。
func (b *Backend) Stop(ctx context.Context, p profile.Profile) error {
	return b.action(ctx, p, "stop")
}

// Restart 重启 unit。
func (b *Backend) Restart(ctx context.Context, p profile.Profile) error {
	return b.action(ctx, p, "restart")
}

func (b *Backend) action(ctx context.Context, p profile.Profile, verb string) error {
	unit := b.unit(p)
	args := []string{"--user", verb}
	if verb == "start" || verb == "restart" {
		args = append(args, "--no-block")
	}
	args = append(args, unit)
	cmd := exec.CommandContext(ctx, "systemctl", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemd: %s %s failed: %s: %w", verb, unit, strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Logs 流式读取 unit 日志（journalctl --user-unit）。
func (b *Backend) Logs(ctx context.Context, p profile.Profile, opts runtime.LogOptions) (<-chan runtime.LogEntry, error) {
	return journal.StreamLogs(ctx, b.unit(p), opts)
}

// parseStatus 解析 systemctl --user show 输出为 RuntimeStatus。
func parseStatus(out string) (runtime.RuntimeStatus, error) {
	rs := runtime.RuntimeStatus{State: runtime.StateInactive}
	kv := map[string]string{}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			kv[parts[0]] = strings.TrimSpace(parts[1])
		}
	}
	switch kv["ActiveState"] {
	case "active":
		rs.State = runtime.StateRunning
	case "activating":
		rs.State = runtime.StateStarting
	case "failed":
		rs.State = runtime.StateFailed
	case "inactive", "deactivating":
		rs.State = runtime.StateStopped
	default:
		rs.State = runtime.StateInactive
	}
	rs.SubState = kv["SubState"]
	if pid, err := strconv.Atoi(kv["MainPID"]); err == nil && pid > 0 {
		rs.PID = pid
	}
	if mem, err := strconv.ParseUint(kv["MemoryCurrent"], 10, 64); err == nil {
		rs.MemoryBytes = mem
	}
	if ts := kv["ExecMainStartTimestamp"]; ts != "" && ts != "n/a" {
		if t, err := time.Parse("Mon 2006-01-02 15:04:05 MST", ts); err == nil {
			rs.StartedAt = t
		}
	}
	return rs, nil
}
