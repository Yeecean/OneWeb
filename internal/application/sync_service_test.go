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

func TestScanLocalTree(t *testing.T) {
	svc := NewSyncService()
	confdir := t.TempDir()
	syncDir := filepath.Join(confdir, "data")

	// 构造目录结构:
	// data/
	//   Documents/
	//     Projects/
	//       Code/
	//   Pictures/
	//   .hidden/        (应跳过)
	for _, d := range []string{
		"Documents/Projects/Code",
		"Pictures",
		".hidden",
	} {
		if err := os.MkdirAll(filepath.Join(syncDir, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// 写 config 指向 syncDir
	if err := os.WriteFile(filepath.Join(confdir, "config"),
		[]byte("sync_dir = \""+syncDir+"\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	nodes, err := svc.ScanLocalTree(confdir, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 top-level dirs (hidden skipped), got %d: %+v", len(nodes), nodes)
	}
	// 找到 Documents
	var docs *DirNode
	for i := range nodes {
		if nodes[i].Name == "Documents" {
			docs = &nodes[i]
		}
	}
	if docs == nil {
		t.Fatal("Documents not found")
	}
	if docs.Path != "/Documents" {
		t.Fatalf("Documents path = %q, want /Documents", docs.Path)
	}
	if len(docs.Children) != 1 || docs.Children[0].Path != "/Documents/Projects" {
		t.Fatalf("unexpected children: %+v", docs.Children)
	}
}

func TestScanLocalTreeDepthLimit(t *testing.T) {
	svc := NewSyncService()
	confdir := t.TempDir()
	syncDir := filepath.Join(confdir, "data")
	if err := os.MkdirAll(filepath.Join(syncDir, "A/B/C/D"), 0o755); err != nil {
		t.Fatal(err)
	}
	os.WriteFile(filepath.Join(confdir, "config"), []byte("sync_dir = \""+syncDir+"\"\n"), 0o644)

	nodes, err := svc.ScanLocalTree(confdir, 2)
	if err != nil {
		t.Fatal(err)
	}
	// A -> B -> (depth 2 停止，无 C)
	if len(nodes) != 1 || nodes[0].Name != "A" {
		t.Fatalf("unexpected top: %+v", nodes)
	}
	if len(nodes[0].Children) != 1 || nodes[0].Children[0].Name != "B" {
		t.Fatalf("unexpected children at depth 1: %+v", nodes[0].Children)
	}
	if len(nodes[0].Children[0].Children) != 0 {
		t.Fatalf("depth limit exceeded: %+v", nodes[0].Children[0].Children)
	}
}

func TestScanLocalTreeMissingSyncDir(t *testing.T) {
	svc := NewSyncService()
	confdir := t.TempDir()
	// config 指向不存在的目录
	os.WriteFile(filepath.Join(confdir, "config"), []byte("sync_dir = \"/nonexistent/xyz\"\n"), 0o644)
	nodes, err := svc.ScanLocalTree(confdir, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) != 0 {
		t.Fatalf("expected empty tree for missing dir, got %+v", nodes)
	}
}
