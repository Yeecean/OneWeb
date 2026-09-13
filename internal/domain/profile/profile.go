// Package profile 定义 OneWeb 配置轮廓的核心领域实体。
// Profile 仅描述"如何找到和运行一个 OneDrive 配置"，
// 绝不持久化 sync_dir 等同步业务参数（架构红线 3）。
package profile

import (
	"errors"
	"regexp"
)

// RuntimeType 枚举可用的运行时后端类型。
type RuntimeType string

const (
	RuntimeSystemd RuntimeType = "systemd"
	RuntimeDocker  RuntimeType = "docker"
	RuntimePodman  RuntimeType = "podman"
)

var (
	ErrEmptyID        = errors.New("profile: id must not be empty")
	ErrInvalidID      = errors.New("profile: id may only contain letters, digits, hyphen and underscore")
	ErrEmptyConfDir   = errors.New("profile: confdir must not be empty")
	ErrInvalidRuntime = errors.New("profile: invalid runtime type")

	idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
)

// Profile 表示如何定位并驱动一个独立的 OneDrive 实例。
// 注意：Profile 模型中绝对不包含 sync_dir，sync_dir 由 confdir/config 唯一裁定。
type Profile struct {
	ID            string      `json:"id"`
	DisplayName   string      `json:"display_name"`
	ConfDir       string      `json:"confdir"`
	RuntimeType   RuntimeType `json:"runtime_type"`
	RuntimeTarget string      `json:"runtime_target"`
}

// Validate 校验 Profile 的基本约束。
func (p *Profile) Validate() error {
	if p.ID == "" {
		return ErrEmptyID
	}
	if !idPattern.MatchString(p.ID) {
		return ErrInvalidID
	}
	if p.ConfDir == "" {
		return ErrEmptyConfDir
	}
	switch p.RuntimeType {
	case RuntimeSystemd, RuntimeDocker, RuntimePodman:
	default:
		return ErrInvalidRuntime
	}
	return nil
}

// ProfileRepository 定义 Profile 的持久化仓储接口。
type ProfileRepository interface {
	List() ([]Profile, error)
	Get(id string) (*Profile, error)
	Save(p Profile) error
	Delete(id string) error
}
