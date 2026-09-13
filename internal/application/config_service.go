// Package application 实现应用服务层，协调领域与基础设施。
package application

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/config"
	"github.com/yeecean/oneweb/internal/infrastructure/configparser"
	"github.com/yeecean/oneweb/internal/infrastructure/filesystem"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// ErrConflict 表示配置文件已被外部修改（409 Conflict）。
var ErrConflict = errors.New("application: config conflict")

// ConfigResponse 返回三重配置对比复合结构。
type ConfigResponse struct {
	FileExists      bool                        `json:"file_exists"`
	FileConfig      map[string]string           `json:"file_config"`
	EffectiveConfig map[string]string           `json:"effective_config"`
	Defaults        map[string]string           `json:"defaults"`
	Schema          *config.ConfigSchema        `json:"schema"`
	UI              *configparser.UISchema      `json:"ui,omitempty"`
	VersionMeta     *filesystem.FileFingerprint `json:"version_meta,omitempty"`
}

// SaveConfigRequest 携带目标配置与指纹校验信息。
type SaveConfigRequest struct {
	// Desired 为期望的键值对（增量更新）。
	Desired map[string]string `json:"desired"`
	// BaseSHA256 为前端读取时的文件指纹，用于并发冲突检测。
	BaseSHA256 string `json:"base_sha256"`
	// RemoveKeys 需要移除（转为注释）的键。
	RemoveKeys []string `json:"remove_keys,omitempty"`
	// ValidateDisplay / ValidateDryRun 控制沙箱校验级别。
	ValidateDisplay bool `json:"validate_display,omitempty"`
	ValidateDryRun  bool `json:"validate_dry_run,omitempty"`
}

// ConfigService 组装配置引擎各组件。
type ConfigService struct {
	Schema     *config.ConfigSchema
	CLI        *onedrive.CLIExecutor
	RuntimeDir string
}

// NewConfigService 创建配置服务并加载 v2.5 语义 Schema。
func NewConfigService(cli *onedrive.CLIExecutor) (*ConfigService, error) {
	schema, err := configparser.LoadSemanticSchema("2.5")
	if err != nil {
		return nil, err
	}
	return &ConfigService{
		Schema:     schema,
		CLI:        cli,
		RuntimeDir: os.Getenv("XDG_RUNTIME_DIR"),
	}, nil
}

// configPath 返回 confdir 下的 config 文件路径。
func configPath(confdir string) string {
	return filepath.Join(confdir, "config")
}

// ReadConfig 读取并解析配置文件，返回复合配置结构。
func (s *ConfigService) ReadConfig(confdir string) (*ConfigResponse, error) {
	resp := &ConfigResponse{
		FileConfig:      map[string]string{},
		EffectiveConfig: map[string]string{},
		Defaults:        map[string]string{},
		Schema:          s.Schema,
	}
	// 填充默认值
	for _, opt := range s.Schema.Options {
		if opt.Default != nil {
			resp.Defaults[opt.Key] = fmt.Sprintf("%v", opt.Default)
		}
	}

	path := configPath(confdir)
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			resp.EffectiveConfig = resp.Defaults
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

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	doc, err := configparser.Parse(f)
	if err != nil {
		return nil, err
	}
	for i := range doc.Nodes {
		n := &doc.Nodes[i]
		if n.Type == config.NodeKeyValue {
			resp.FileConfig[n.Key] = n.Value
		}
	}
	// effective = defaults 覆盖 file_config
	resp.EffectiveConfig = map[string]string{}
	for k, v := range resp.Defaults {
		resp.EffectiveConfig[k] = v
	}
	for k, v := range resp.FileConfig {
		resp.EffectiveConfig[k] = v
	}
	return resp, nil
}

// SaveConfig 执行完整保存流水线（指纹比对 → 修改 → Schema 校验 → 沙箱校验 → 原子写入）。
func (s *ConfigService) SaveConfig(ctx context.Context, confdir string, req SaveConfigRequest) ([]config.ValidationError, error) {
	path := configPath(confdir)

	// 1. 指纹比对
	if req.BaseSHA256 != "" {
		ok, err := filesystem.CompareFingerprint(path, &filesystem.FileFingerprint{SHA256: req.BaseSHA256})
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("%w: config file has been modified externally", ErrConflict)
		}
	}

	// 2. 解析并应用修改
	var doc *config.ConfigDocument
	fileExists := false
	if _, err := os.Stat(path); err == nil {
		fileExists = true
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		doc, err = configparser.Parse(f)
		f.Close()
		if err != nil {
			return nil, err
		}
	} else {
		doc = &config.ConfigDocument{}
	}

	for k, v := range req.Desired {
		doc.Set(k, v)
	}
	for _, k := range req.RemoveKeys {
		doc.Remove(k)
	}

	// 3. Level 1 Schema 校验
	results := configparser.ValidateAgainstSchema(doc, s.Schema)
	if configparser.HasErrors(results) {
		return results, fmt.Errorf("application: schema validation failed")
	}

	// 4. 可选沙箱校验
	if (req.ValidateDisplay || req.ValidateDryRun) && s.CLI != nil {
		rendered := configparser.Render(doc)
		if _, err := s.CLI.ValidateInSandbox(ctx, []byte(rendered), s.RuntimeDir, req.ValidateDisplay, req.ValidateDryRun); err != nil {
			return results, fmt.Errorf("application: sandbox validation failed: %w", err)
		}
	}

	// 5. 渲染并原子写入
	if !fileExists {
		if err := os.MkdirAll(confdir, 0o755); err != nil {
			return nil, err
		}
	}
	rendered := configparser.Render(doc)
	if err := filesystem.AtomicWriteFile(path, []byte(rendered), 0o644); err != nil {
		return results, err
	}
	return results, nil
}

// ValidateConfig 仅执行校验，不落盘。
func (s *ConfigService) ValidateConfig(ctx context.Context, confdir string, content string) ([]config.ValidationError, error) {
	doc, err := configparser.Parse(strings.NewReader(content))
	if err != nil {
		return nil, err
	}
	results := configparser.ValidateAgainstSchema(doc, s.Schema)
	if configparser.HasErrors(results) {
		return results, fmt.Errorf("application: schema validation failed")
	}
	if s.CLI != nil {
		if _, err := s.CLI.ValidateInSandbox(ctx, []byte(content), s.RuntimeDir, true, true); err != nil {
			return results, fmt.Errorf("application: sandbox validation failed: %w", err)
		}
	}
	return results, nil
}
