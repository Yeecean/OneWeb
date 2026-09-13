package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/profile"
)

func newTestStore(t *testing.T) *ProfileStore {
	t.Helper()
	return &ProfileStore{Path: filepath.Join(t.TempDir(), "profiles.json")}
}

func TestProfileStoreCRUD(t *testing.T) {
	store := newTestStore(t)
	p := profile.Profile{ID: "default", DisplayName: "Personal", ConfDir: "/home/u/.config/onedrive", RuntimeType: profile.RuntimeSystemd, RuntimeTarget: "onedrive@default.service"}

	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}
	got, err := store.Get("default")
	if err != nil {
		t.Fatal(err)
	}
	if got.ConfDir != p.ConfDir {
		t.Fatalf("confdir = %q", got.ConfDir)
	}
	// 更新
	p.DisplayName = "Renamed"
	if err := store.Save(p); err != nil {
		t.Fatal(err)
	}
	list, _ := store.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(list))
	}
	if list[0].DisplayName != "Renamed" {
		t.Fatalf("display = %q", list[0].DisplayName)
	}
	// 删除
	if err := store.Delete("default"); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Get("default"); !os.IsNotExist(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestProfileStoreFirstRunCreatesEmpty(t *testing.T) {
	store := newTestStore(t)
	list, err := store.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d", len(list))
	}
}

func TestProfileStoreDeleteMissing(t *testing.T) {
	store := newTestStore(t)
	if err := store.Delete("nope"); !os.IsNotExist(err) {
		t.Fatalf("expected not found, got %v", err)
	}
}
