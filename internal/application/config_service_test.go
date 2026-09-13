package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// newTestConfigService 构造无需 onedrive 二进制的测试服务（CLI 可空）。
func newTestConfigService(t *testing.T) *ConfigService {
	t.Helper()
	var cli *onedrive.CLIExecutor
	// 若系统存在 onedrive 则加载，否则用 nil（跳过沙箱校验）
	if _, err := onedrive.NewCLIExecutor(""); err == nil {
		cli, _ = onedrive.NewCLIExecutor("")
	}
	svc, err := NewConfigService(cli)
	if err != nil {
		t.Fatal(err)
	}
	return svc
}

func writeConfig(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, "config")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestReadConfigMissingFile(t *testing.T) {
	svc := newTestConfigService(t)
	resp, err := svc.ReadConfig(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatal(err)
	}
	if resp.FileExists {
		t.Fatal("FileExists should be false")
	}
	if resp.EffectiveConfig["threads"] != "8" {
		t.Fatalf("effective threads default = %q", resp.EffectiveConfig["threads"])
	}
}

func TestReadConfigParsesFile(t *testing.T) {
	svc := newTestConfigService(t)
	dir := t.TempDir()
	writeConfig(t, dir, "sync_dir = \"/od\"\nthreads = 4\n")
	resp, err := svc.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !resp.FileExists {
		t.Fatal("FileExists should be true")
	}
	if resp.FileConfig["threads"] != "4" {
		t.Fatalf("file threads = %q", resp.FileConfig["threads"])
	}
	if resp.VersionMeta == nil || resp.VersionMeta.SHA256 == "" {
		t.Fatal("version meta missing")
	}
	if resp.EffectiveConfig["threads"] != "4" {
		t.Fatalf("effective threads = %q", resp.EffectiveConfig["threads"])
	}
}

func TestSaveConfigModifyThenRead(t *testing.T) {
	svc := newTestConfigService(t)
	dir := t.TempDir()
	writeConfig(t, dir, "# a comment\nsync_dir = \"/od\"\nthreads = 4\n")
	resp, err := svc.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.SaveConfig(context.Background(), dir, SaveConfigRequest{
		BaseSHA256: resp.VersionMeta.SHA256,
		Desired:    map[string]string{"threads": "12"},
	})
	if err != nil {
		t.Fatal(err)
	}
	// 重新读取验证生效
	resp2, err := svc.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.FileConfig["threads"] != "12" {
		t.Fatalf("threads after save = %q", resp2.FileConfig["threads"])
	}
	// 注释保留
	raw, _ := os.ReadFile(filepath.Join(dir, "config"))
	if !strings.Contains(string(raw), "# a comment") {
		t.Fatal("comment lost after save")
	}
}

func TestSaveConfigConflict(t *testing.T) {
	svc := newTestConfigService(t)
	dir := t.TempDir()
	writeConfig(t, dir, "threads = 4\n")
	resp, err := svc.ReadConfig(dir)
	if err != nil {
		t.Fatal(err)
	}
	// 模拟外部修改
	f, _ := os.OpenFile(filepath.Join(dir, "config"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("verbose = \"true\"\n")
	f.Close()

	_, err = svc.SaveConfig(context.Background(), dir, SaveConfigRequest{
		BaseSHA256: resp.VersionMeta.SHA256,
		Desired:    map[string]string{"threads": "8"},
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}

func TestSaveConfigCreatesNewFile(t *testing.T) {
	svc := newTestConfigService(t)
	dir := filepath.Join(t.TempDir(), "fresh")
	_, err := svc.SaveConfig(context.Background(), dir, SaveConfigRequest{
		Desired: map[string]string{"sync_dir": "/od", "threads": "4"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "config")); err != nil {
		t.Fatalf("config not created: %v", err)
	}
}

func TestSaveConfigSchemaError(t *testing.T) {
	svc := newTestConfigService(t)
	dir := t.TempDir()
	writeConfig(t, dir, "threads = 4\n")
	_, err := svc.SaveConfig(context.Background(), dir, SaveConfigRequest{
		Desired: map[string]string{"threads": "999"},
	})
	if err == nil {
		t.Fatal("expected schema validation error")
	}
}
