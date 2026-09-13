package application

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncServiceGetMissing(t *testing.T) {
	svc := NewSyncService()
	dir := filepath.Join(t.TempDir(), "nope")
	resp, err := svc.GetSyncList(dir)
	if err != nil {
		t.Fatal(err)
	}
	if resp.FileExists {
		t.Fatal("FileExists should be false")
	}
	if len(resp.Rules) != 0 {
		t.Fatalf("rules = %v", resp.Rules)
	}
}

func TestSyncServiceSaveAndGet(t *testing.T) {
	svc := NewSyncService()
	dir := t.TempDir()
	source := "/Documents\n!/Documents/Temp\n"
	resp, err := svc.SaveSyncList(dir, SaveSyncListRequest{Source: source})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.ResyncRequired {
		t.Fatal("ResyncRequired should be true after save")
	}
	if len(resp.Rules) != 2 {
		t.Fatalf("rules = %d, want 2", len(resp.Rules))
	}

	got, err := svc.GetSyncList(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got.Source != source {
		t.Fatalf("source mismatch: %q", got.Source)
	}
	if got.VersionMeta == nil {
		t.Fatal("version meta missing")
	}
}

func TestSyncServiceConflict(t *testing.T) {
	svc := NewSyncService()
	dir := t.TempDir()
	if _, err := svc.SaveSyncList(dir, SaveSyncListRequest{Source: "/Documents\n"}); err != nil {
		t.Fatal(err)
	}
	// 外部修改
	f, _ := os.OpenFile(filepath.Join(dir, "sync_list"), os.O_APPEND|os.O_WRONLY, 0o644)
	f.WriteString("/Pictures\n")
	f.Close()

	_, err := svc.SaveSyncList(dir, SaveSyncListRequest{
		Source:     "/New\n",
		BaseSHA256: "stale-sha",
	})
	if err == nil || !strings.Contains(err.Error(), "conflict") {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestSyncServicePerformanceWarning(t *testing.T) {
	svc := NewSyncService()
	dir := t.TempDir()
	resp, err := svc.SaveSyncList(dir, SaveSyncListRequest{Source: "Documents\n"})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, w := range resp.Warnings {
		if w.Level == "performance" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected performance warning for bare rule")
	}
}
