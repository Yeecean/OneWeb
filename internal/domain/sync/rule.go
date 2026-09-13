// Package sync 定义 SyncList 选择性同步规则的领域模型。
package sync

import "context"

// RuleType 是同步规则的语义类型。
type RuleType string

const (
	// RuleInclude 包含规则（默认白名单包含）。
	RuleInclude RuleType = "include"
	// RuleExclude 排除规则（以 ! 开头）。
	RuleExclude RuleType = "exclude"
	// RuleComment 注释行（原样保留）。
	RuleComment RuleType = "comment"
	// RuleBlank 空行（原样保留）。
	RuleBlank RuleType = "blank"
)

// SyncRule 表示单条 sync_list 行。
type SyncRule struct {
	RawText    string   `json:"raw_text"`
	Type       RuleType `json:"type"`
	Pattern    string   `json:"pattern,omitempty"`
	IsRooted   bool     `json:"is_rooted,omitempty"` // 是否以 / 开头（精确匹配根路径）
	LineNumber int      `json:"line_number"`
}

// SyncRuleSet 表示一组有序规则。
type SyncRuleSet struct {
	Rules []SyncRule `json:"rules"`
}

// TreeNode 表示 OneDrive 远程/本地目录树节点。
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Children []*TreeNode `json:"children,omitempty"`
}

// RemoteTreeProvider 提供构建目录树的能力接口。
type RemoteTreeProvider interface {
	BuildTree(ctx context.Context, confDir string) (*TreeNode, error)
}
