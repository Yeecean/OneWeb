package application

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/profile"
)

type memoryProfileRepo struct {
	profiles map[string]profile.Profile
}

func (m *memoryProfileRepo) List() ([]profile.Profile, error) {
	var list []profile.Profile
	for _, p := range m.profiles {
		list = append(list, p)
	}
	return list, nil
}

func (m *memoryProfileRepo) Get(id string) (*profile.Profile, error) {
	p, ok := m.profiles[id]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &p, nil
}

func (m *memoryProfileRepo) Save(p profile.Profile) error {
	m.profiles[p.ID] = p
	return nil
}

func (m *memoryProfileRepo) Delete(id string) error {
	delete(m.profiles, id)
	return nil
}

func TestProfileServiceAutoDiscover(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// 创建模拟 ~/.config/onedrive
	defaultDir := filepath.Join(tmpDir, "onedrive")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// 创建模拟 ~/.config/onedrive-work
	workDir := filepath.Join(tmpDir, "onedrive-work")
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		t.Fatal(err)
	}

	repo := &memoryProfileRepo{profiles: make(map[string]profile.Profile)}
	svc := NewProfileService(repo, nil)

	profiles, err := svc.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 discovered profiles, got %d", len(profiles))
	}

	def, err := svc.GetProfile("default")
	if err != nil {
		t.Fatalf("default profile not found: %v", err)
	}
	if def.ConfDir != defaultDir {
		t.Errorf("expected confdir %s, got %s", defaultDir, def.ConfDir)
	}

	work, err := svc.GetProfile("work")
	if err != nil {
		t.Fatalf("work profile not found: %v", err)
	}
	if work.ConfDir != workDir {
		t.Errorf("expected confdir %s, got %s", workDir, work.ConfDir)
	}
}

func TestProfileServiceAutoProvisionFreshMachine(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// 不创建任何 onedrive 目录，模拟全新机器
	repo := &memoryProfileRepo{profiles: make(map[string]profile.Profile)}
	svc := NewProfileService(repo, nil)

	profiles, err := svc.ListProfiles()
	if err != nil {
		t.Fatalf("ListProfiles failed: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 auto-provisioned profile, got %d", len(profiles))
	}

	if profiles[0].ID != "default" {
		t.Errorf("expected id 'default', got '%s'", profiles[0].ID)
	}
}
