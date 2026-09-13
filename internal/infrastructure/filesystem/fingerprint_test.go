package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFingerprintStable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("threads = 4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp1, err := ComputeFingerprint(path)
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := ComputeFingerprint(path)
	if err != nil {
		t.Fatal(err)
	}
	if fp1.SHA256 != fp2.SHA256 || fp1.Size != fp2.Size {
		t.Fatalf("fingerprints differ without modification: %+v vs %+v", fp1, fp2)
	}
	ok, err := CompareFingerprint(path, fp1)
	if err != nil || !ok {
		t.Fatalf("compare should match: ok=%v err=%v", ok, err)
	}
}

func TestFingerprintDetectsModification(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config")
	if err := os.WriteFile(path, []byte("threads = 4\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp, err := ComputeFingerprint(path)
	if err != nil {
		t.Fatal(err)
	}
	// 追加内容模拟外部修改
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("sync_dir = \"/tmp\"\n"); err != nil {
		t.Fatal(err)
	}
	f.Close()

	ok, err := CompareFingerprint(path, fp)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("fingerprint should detect external modification")
	}
}

func TestFingerprintMissingFile(t *testing.T) {
	if _, err := ComputeFingerprint("/nonexistent/never-exists"); err == nil {
		t.Fatal("expected error for missing file")
	}
	ok, err := CompareFingerprint("/nonexistent/never-exists", &FileFingerprint{})
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("missing file should not match")
	}
}
