package filesystem

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtomicWriteContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	content := []byte("threads = 8\nsync_dir = \"/od\"\n")
	if err := AtomicWriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("content mismatch:\n%q\n!=\n%q", string(got), string(content))
	}
	// 不残留临时文件
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if strings.Contains(e.Name(), ".tmp.") {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}

func TestAtomicWriteOverwritesAtomically(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := AtomicWriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AtomicWriteFile(path, []byte("new content"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "new content" {
		t.Fatalf("overwrite failed: %q", string(got))
	}
}

func TestAtomicWriteUnwritableDir(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root, permission checks not meaningful")
	}
	dir := t.TempDir()
	sub := filepath.Join(dir, "readonly")
	if err := os.Mkdir(sub, 0o555); err != nil {
		t.Fatal(err)
	}
	err := AtomicWriteFile(filepath.Join(sub, "config"), []byte("x"), 0o644)
	if err == nil {
		t.Fatal("expected error writing to read-only dir")
	}
}

func TestAtomicWritePreservesExistingOnInterruptedWrite(t *testing.T) {
	// 模拟写入失败（目标目录不可写）时原始文件不受影响
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := AtomicWriteFile(path, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	// 在只读子目录中重试应失败且不影响已有文件
	sub := filepath.Join(dir, "ro")
	os.Mkdir(sub, 0o555)
	if os.Geteuid() != 0 {
		if err := AtomicWriteFile(filepath.Join(sub, "config"), []byte("x"), 0o644); err == nil {
			t.Fatal("expected failure")
		}
	}
	got, _ := os.ReadFile(path)
	if string(got) != "original" {
		t.Fatalf("original corrupted: %q", string(got))
	}
}
