package application

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// ErrConfDirNotExist 表示创建的 confdir 不存在或不可访问。
var ErrConfDirNotExist = errors.New("application: confdir does not exist or is not a directory")

// CreateProfileRequest 是创建 Profile 的入参。
type CreateProfileRequest struct {
	ID            string              `json:"id"`
	DisplayName   string              `json:"display_name"`
	ConfDir       string              `json:"confdir"`
	RuntimeType   profile.RuntimeType `json:"runtime_type"`
	RuntimeTarget string              `json:"runtime_target"`
}

// UpdateProfileRequest 是更新 Profile 元数据的入参（部分字段）。
type UpdateProfileRequest struct {
	DisplayName   *string `json:"display_name,omitempty"`
	RuntimeTarget *string `json:"runtime_target,omitempty"`
}

// ProfileService 编排 Profile 生命周期。
type ProfileService struct {
	Repo profile.ProfileRepository
	CLI  *onedrive.CLIExecutor
}

// NewProfileService 创建 Profile 服务。
func NewProfileService(repo profile.ProfileRepository, cli *onedrive.CLIExecutor) *ProfileService {
	return &ProfileService{Repo: repo, CLI: cli}
}

// detectDefaultSystemdTarget 探测系统是否存在 onedrive.service 单元。
func (s *ProfileService) detectDefaultSystemdTarget() string {
	cmd := exec.Command("systemctl", "--user", "list-unit-files", "onedrive.service")
	if out, err := cmd.Output(); err == nil && strings.Contains(string(out), "onedrive.service") {
		return "onedrive.service"
	}
	return "onedrive@default.service"
}

// AutoDiscoverProfiles 扫描宿主机配置目录与 systemd 服务，自动注册现有或默认 Profile。
func (s *ProfileService) AutoDiscoverProfiles() ([]profile.Profile, error) {
	home, _ := os.UserHomeDir()
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" && home != "" {
		configDir = filepath.Join(home, ".config")
	}
	if configDir == "" {
		return s.Repo.List()
	}

	defaultTarget := s.detectDefaultSystemdTarget()

	// 1. 优先探测默认目录 ~/.config/onedrive
	defaultDir := filepath.Join(configDir, "onedrive")
	if fi, err := os.Stat(defaultDir); err == nil && fi.IsDir() {
		if _, err := s.Repo.Get("default"); err != nil {
			p := profile.Profile{
				ID:            "default",
				DisplayName:   "个人 OneDrive",
				ConfDir:       defaultDir,
				RuntimeType:   profile.RuntimeSystemd,
				RuntimeTarget: defaultTarget,
			}
			_ = s.Repo.Save(p)
		}
	}

	// 2. 扫描多账号目录 ~/.config/onedrive-*
	if entries, err := os.ReadDir(configDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), "onedrive-") {
				suffix := strings.TrimPrefix(entry.Name(), "onedrive-")
				if suffix == "" {
					continue
				}
				if _, err := s.Repo.Get(suffix); err != nil {
					p := profile.Profile{
						ID:            suffix,
						DisplayName:   fmt.Sprintf("OneDrive (%s)", suffix),
						ConfDir:       filepath.Join(configDir, entry.Name()),
						RuntimeType:   profile.RuntimeSystemd,
						RuntimeTarget: fmt.Sprintf("onedrive@%s.service", suffix),
					}
					_ = s.Repo.Save(p)
				}
			}
		}
	}

	// 3. 若宿主机全新、未找到任何既有目录，自动预置标准 default 账户，实现零配置开箱即用
	list, err := s.Repo.List()
	if err == nil && len(list) == 0 {
		_ = os.MkdirAll(defaultDir, 0o755)
		p := profile.Profile{
			ID:            "default",
			DisplayName:   "个人 OneDrive",
			ConfDir:       defaultDir,
			RuntimeType:   profile.RuntimeSystemd,
			RuntimeTarget: defaultTarget,
		}
		_ = s.Repo.Save(p)
	}

	return s.Repo.List()
}

// ListProfiles 返回全部 Profile；若当前无任何 Profile 则自动扫描并初始化。
func (s *ProfileService) ListProfiles() ([]profile.Profile, error) {
	profiles, err := s.Repo.List()
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return s.AutoDiscoverProfiles()
	}
	return profiles, nil
}

// CreateProfile 创建 Profile，校验 confdir 与运行时类型。
func (s *ProfileService) CreateProfile(req CreateProfileRequest) (*profile.Profile, error) {
	p := profile.Profile{
		ID:            req.ID,
		DisplayName:   req.DisplayName,
		ConfDir:       filepath.Clean(req.ConfDir),
		RuntimeType:   req.RuntimeType,
		RuntimeTarget: req.RuntimeTarget,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	// 校验 confdir 存在性（允许不存在但可创建的场景由调用方决定；这里要求父目录存在）
	if _, err := os.Stat(p.ConfDir); err != nil {
		if os.IsNotExist(err) {
			// 尝试创建
			if err := os.MkdirAll(p.ConfDir, 0o755); err != nil {
				return nil, ErrConfDirNotExist
			}
		} else {
			return nil, err
		}
	}
	if p.DisplayName == "" {
		p.DisplayName = p.ID
	}
	if err := s.Repo.Save(p); err != nil {
		return nil, err
	}
	return &p, nil
}

// GetProfile 获取单个 Profile。
func (s *ProfileService) GetProfile(id string) (*profile.Profile, error) {
	return s.Repo.Get(id)
}

// UpdateProfile 更新 Profile 元数据。
func (s *ProfileService) UpdateProfile(id string, req UpdateProfileRequest) (*profile.Profile, error) {
	p, err := s.Repo.Get(id)
	if err != nil {
		return nil, err
	}
	if req.DisplayName != nil {
		p.DisplayName = *req.DisplayName
	}
	if req.RuntimeTarget != nil {
		p.RuntimeTarget = *req.RuntimeTarget
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	if err := s.Repo.Save(*p); err != nil {
		return nil, err
	}
	return p, nil
}

// DeleteProfile 删除 Profile 元数据（绝不级联删除磁盘文件）。
func (s *ProfileService) DeleteProfile(id string) error {
	if _, err := s.Repo.Get(id); err != nil {
		return fmt.Errorf("application: profile %q not found", id)
	}
	return s.Repo.Delete(id)
}
