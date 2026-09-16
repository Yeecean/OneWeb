package application

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/sync"
	"github.com/yeecean/oneweb/internal/infrastructure/configparser"
	"github.com/yeecean/oneweb/internal/infrastructure/filesystem"
)

// ErrSyncListConflict 表示 sync_list 已被外部修改。
var ErrSyncListConflict = errors.New("application: sync_list conflict")

// SyncListResponse 返回规则源码与解析后的模型。
type SyncListResponse struct {
	FileExists  bool                             `json:"file_exists"`
	Source      string                           `json:"source"`
	Rules       []sync.SyncRule                  `json:"rules"`
	Warnings    []configparser.ValidationWarning `json:"warnings"`
	VersionMeta *filesystem.FileFingerprint      `json:"version_meta,omitempty"`
	// ResyncRequired 修改 sync_list 后必须全量重同步。
	ResyncRequired bool `json:"resync_required"`
}

// SaveSyncListRequest 携带新规则与指纹。
type SaveSyncListRequest struct {
	Source     string `json:"source"`
	BaseSHA256 string `json:"base_sha256"`
}

// SyncService 编排 SyncList 生命周期。
type SyncService struct {
	RuntimeDir string
}

// NewSyncService 创建 SyncList 服务。
func NewSyncService() *SyncService {
	return &SyncService{RuntimeDir: os.Getenv("XDG_RUNTIME_DIR")}
}

// syncListPath 返回 confdir 下的 sync_list 路径。
func syncListPath(confdir string) string {
	return filepath.Join(confdir, "sync_list")
}

// GetSyncList 读取并解析 sync_list。
func (s *SyncService) GetSyncList(confdir string) (*SyncListResponse, error) {
	resp := &SyncListResponse{
		Source:         "",
		Rules:          []sync.SyncRule{},
		Warnings:       []configparser.ValidationWarning{},
		ResyncRequired: false,
	}
	path := syncListPath(confdir)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return resp, nil
		}
		return nil, err
	}
	resp.FileExists = true
	fp, err := filesystem.ComputeFingerprint(path)
	if err != nil {
		return nil, err
	}
	resp.VersionMeta = fp
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	resp.Source = string(b)
	set, err := configparser.ParseSyncList(strings.NewReader(resp.Source))
	if err != nil {
		return nil, err
	}
	resp.Rules = set.Rules
	warnings := configparser.ValidateSyncList(set)
	if warnings == nil {
		warnings = []configparser.ValidationWarning{}
	}
	resp.Warnings = warnings
	return resp, nil
}

// SaveSyncList 保存 sync_list 并返回 resync 提醒。
func (s *SyncService) SaveSyncList(confdir string, req SaveSyncListRequest) (*SyncListResponse, error) {
	path := syncListPath(confdir)

	if req.BaseSHA256 != "" {
		ok, err := filesystem.CompareFingerprint(path, &filesystem.FileFingerprint{SHA256: req.BaseSHA256})
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%w: sync_list has been modified externally", ErrSyncListConflict)
		}
	}

	// 解析校验
	set, err := configparser.ParseSyncList(strings.NewReader(req.Source))
	if err != nil {
		return nil, err
	}
	warnings := configparser.ValidateSyncList(set)
	if warnings == nil {
		warnings = []configparser.ValidationWarning{}
	}

	if err := os.MkdirAll(confdir, 0o755); err != nil {
		return nil, err
	}
	if err := filesystem.AtomicWriteFile(path, []byte(req.Source), 0o644); err != nil {
		return nil, err
	}
	// 修改 sync_list 后必须 resync 才能生效
	return &SyncListResponse{
		FileExists:     true,
		Source:         req.Source,
		Rules:          set.Rules,
		Warnings:       warnings,
		ResyncRequired: true,
	}, nil
}
