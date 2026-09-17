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

// DirNode 表示目录树的一个节点。
type DirNode struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"` // 相对于 sync_dir 的路径，如 /Documents
	IsDir    bool      `json:"is_dir"`
	Children []DirNode `json:"children,omitempty"`
}

// ScanLocalTree 扫描本地 sync_dir 目录树，最多 maxDepth 层。
// sync_dir 从配置文件读取，默认为 ~/OneDrive。
// 只扫描目录，跳过隐藏目录（以 . 开头），不依赖网络。
func (s *SyncService) ScanLocalTree(confdir string, maxDepth int) ([]DirNode, error) {
	if maxDepth <= 0 {
		maxDepth = 3
	}
	syncDir := s.resolveSyncDir(confdir)
	if syncDir == "" {
		return []DirNode{}, nil
	}
	if _, err := os.Stat(syncDir); err != nil {
		if os.IsNotExist(err) {
			return []DirNode{}, nil
		}
		return nil, err
	}
	nodes, err := scanDir(syncDir, syncDir, 0, maxDepth)
	if err != nil {
		return nil, err
	}
	if nodes == nil {
		nodes = []DirNode{}
	}
	return nodes, nil
}

// resolveSyncDir 读取 confdir/config 获取 sync_dir，失败则返回 ~/OneDrive。
func (s *SyncService) resolveSyncDir(confdir string) string {
	const defaultSyncDir = "~/OneDrive"
	value := defaultSyncDir

	path := filepath.Join(confdir, "config")
	if f, err := os.Open(path); err == nil {
		if doc, perr := configparser.Parse(f); perr == nil {
			if n := doc.Get("sync_dir"); n != nil && n.Value != "" {
				value = n.Value
			}
		}
		f.Close()
	}
	return expandHome(value)
}

// expandHome 将前导 ~ 展开为用户主目录。
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			if p == "~" {
				return home
			}
			return filepath.Join(home, p[2:])
		}
	}
	return p
}

// scanDir 递归扫描目录，Path 使用相对于 baseDir 的格式（以 / 开头）。
func scanDir(baseDir, currentDir string, depth, maxDepth int) ([]DirNode, error) {
	if depth >= maxDepth {
		return nil, nil
	}
	entries, err := os.ReadDir(currentDir)
	if err != nil {
		return nil, nil
	}
	var nodes []DirNode
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		// 跳过隐藏目录
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		full := filepath.Join(currentDir, e.Name())
		rel, relErr := filepath.Rel(baseDir, full)
		if relErr != nil {
			continue
		}
		relPath := "/" + filepath.ToSlash(rel)
		children, _ := scanDir(baseDir, full, depth+1, maxDepth)
		nodes = append(nodes, DirNode{
			Name:     e.Name(),
			Path:     relPath,
			IsDir:    true,
			Children: children,
		})
	}
	return nodes, nil
}
