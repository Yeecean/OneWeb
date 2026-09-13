// Package onedrive 封装对 onedrive CLI 二进制的子进程调用。
package onedrive

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// CLIExecutor 封装对 onedrive 二进制的命令执行。
type CLIExecutor struct {
	BinaryPath string
}

// NewCLIExecutor 创建执行器，二进制路径为空时探测 PATH。
func NewCLIExecutor(binaryPath string) (*CLIExecutor, error) {
	if binaryPath == "" {
		binaryPath = "onedrive"
	}
	// 快速探测可用性
	if _, err := exec.LookPath(binaryPath); err != nil {
		return nil, fmt.Errorf("onedrive: binary %q not found in PATH: %w", binaryPath, err)
	}
	return &CLIExecutor{BinaryPath: binaryPath}, nil
}

// Version 执行 onedrive --version 返回版本字符串。
func (c *CLIExecutor) Version(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("onedrive: version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// DisplayConfig 在给定 confdir 下执行 onedrive --display-config。
func (c *CLIExecutor) DisplayConfig(ctx context.Context, confdir string) (string, error) {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "--confdir="+confdir, "--display-config")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("onedrive: display-config: %w", err)
	}
	return string(out), nil
}

// DryRun 在给定 confdir 下执行 onedrive --dry-run，返回 stdout 与 stderr。
func (c *CLIExecutor) DryRun(ctx context.Context, confdir string) (string, string, error) {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "--confdir="+confdir, "--dry-run")
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Sandbox 创建一次性沙箱目录用于配置校验。
// 返回沙箱路径与清理函数。
func (c *CLIExecutor) Sandbox(runtimeDir, prefix string) (string, func(), error) {
	base := runtimeDir
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "oneweb", prefix)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", nil, fmt.Errorf("onedrive: create sandbox: %w", err)
	}
	return dir, func() { _ = os.RemoveAll(dir) }, nil
}

// ValidateInSandbox 将候选 config 内容写入沙箱并执行三级校验。
// validateDisplay: 执行 --display-config；validateDryRun: 执行 --dry-run。
func (c *CLIExecutor) ValidateInSandbox(ctx context.Context, candidate []byte, runtimeDir string, validateDisplay, validateDryRun bool) ([]string, error) {
	dir, cleanup, err := c.Sandbox(runtimeDir, "op_sandbox")
	if err != nil {
		return nil, err
	}
	defer cleanup()

	if err := os.WriteFile(filepath.Join(dir, "config"), candidate, 0o600); err != nil {
		return nil, fmt.Errorf("onedrive: write sandbox config: %w", err)
	}
	var messages []string
	if validateDisplay {
		out, err := c.DisplayConfig(ctx, dir)
		messages = append(messages, out)
		if err != nil {
			return messages, fmt.Errorf("onedrive: sandbox display-config: %w", err)
		}
	}
	if validateDryRun {
		stdout, stderr, err := c.DryRun(ctx, dir)
		messages = append(messages, stdout, stderr)
		if err != nil {
			return messages, fmt.Errorf("onedrive: sandbox dry-run: %w", err)
		}
	}
	return messages, nil
}
