package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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

// ListProfiles 返回全部 Profile。
func (s *ProfileService) ListProfiles() ([]profile.Profile, error) {
	return s.Repo.List()
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
