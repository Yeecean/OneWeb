package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yeecean/oneweb/internal/domain/profile"
)

// ProfileStore 基于 ~/.config/oneweb/profiles.json 实现 ProfileRepository。
type ProfileStore struct {
	// Path 默认为 $XDG_CONFIG_HOME/oneweb/profiles.json。
	Path string
}

// DefaultProfileStore 使用用户配置目录创建存储。
func DefaultProfileStore() *ProfileStore {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, _ := os.UserHomeDir()
		dir = filepath.Join(home, ".config")
	}
	return &ProfileStore{Path: filepath.Join(dir, "oneweb", "profiles.json")}
}

// List 列出全部 Profile。
func (s *ProfileStore) List() ([]profile.Profile, error) {
	profiles, err := s.read()
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

// Get 获取指定 Profile；不存在返回 os.ErrNotExist。
func (s *ProfileStore) Get(id string) (*profile.Profile, error) {
	profiles, err := s.read()
	if err != nil {
		return nil, err
	}
	for i := range profiles {
		if profiles[i].ID == id {
			p := profiles[i]
			return &p, nil
		}
	}
	return nil, os.ErrNotExist
}

// Save 保存 Profile（新增或覆盖同 ID）。
func (s *ProfileStore) Save(p profile.Profile) error {
	profiles, err := s.read()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	replaced := false
	for i := range profiles {
		if profiles[i].ID == p.ID {
			profiles[i] = p
			replaced = true
			break
		}
	}
	if !replaced {
		profiles = append(profiles, p)
	}
	return s.write(profiles)
}

// Delete 删除 Profile 元数据（绝不级联删除磁盘文件）。
func (s *ProfileStore) Delete(id string) error {
	profiles, err := s.read()
	if err != nil {
		return err
	}
	kept := profiles[:0]
	found := false
	for i := range profiles {
		if profiles[i].ID != id {
			kept = append(kept, profiles[i])
		} else {
			found = true
		}
	}
	if !found {
		return os.ErrNotExist
	}
	return s.write(kept)
}

func (s *ProfileStore) read() ([]profile.Profile, error) {
	b, err := os.ReadFile(s.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return []profile.Profile{}, nil
		}
		return nil, err
	}
	var profiles []profile.Profile
	if err := json.Unmarshal(b, &profiles); err != nil {
		return nil, fmt.Errorf("filesystem: parse profiles: %w", err)
	}
	return profiles, nil
}

func (s *ProfileStore) write(profiles []profile.Profile) error {
	b, err := json.MarshalIndent(profiles, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	return AtomicWriteFile(s.Path, b, 0o600)
}
