package configparser

import (
	"path"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/sync"
)

// Matcher 按照官方 Allow-list 语义评估路径是否被包含。
// 规则：默认全排除；自上而下应用规则；最后匹配的规则胜出；
// 包含规则 → 包含该路径；排除规则 (!) → 排除该路径。
type Matcher struct {
	rules []sync.SyncRule
}

// NewMatcher 从规则集构建匹配器。
func NewMatcher(set *sync.SyncRuleSet) *Matcher {
	var rules []sync.SyncRule
	for _, r := range set.Rules {
		if r.Type == sync.RuleComment || r.Type == sync.RuleBlank {
			continue
		}
		rules = append(rules, r)
	}
	return &Matcher{rules: rules}
}

// IsIncluded 判断给定路径（相对 OneDrive 根，如 "Documents/Report.pdf"）是否被包含。
func (m *Matcher) IsIncluded(relPath string) bool {
	relPath = strings.TrimPrefix(relPath, "/")
	included := false
	for _, r := range m.rules {
		pattern := strings.TrimPrefix(r.Pattern, "/")
		if matchPattern(pattern, r.IsRooted, relPath) {
			included = r.Type == sync.RuleInclude
		}
	}
	return included
}

// matchPattern 判断单条规则是否匹配路径。
// rooted 表示规则是否以 / 开头（仅匹配根目录层级）。
// 规则以 / 结尾或为目录路径时，匹配该目录及子内容。
func matchPattern(pattern string, rooted bool, relPath string) bool {
	// 记录是否为目录型规则（原模式以 / 结尾）
	dirRule := strings.HasSuffix(pattern, "/")
	pattern = strings.TrimSuffix(pattern, "/")
	if pattern == "" {
		return false
	}
	if strings.Contains(pattern, "**") {
		return matchDoubleStar(pattern, relPath)
	}
	// 单段路径规则：无前导 / 时递归匹配任意深度；有前导 / 时仅匹配根层级
	if !strings.Contains(pattern, "/") {
		if rooted {
			// 仅匹配根路径首段
			first := strings.SplitN(relPath, "/", 2)[0]
			ok, err := path.Match(pattern, first)
			return err == nil && ok
		}
		return matchBareName(pattern, relPath)
	}
	// 多段规则
	matched, err := path.Match(pattern, relPath)
	if err != nil {
		return false
	}
	if matched {
		return true
	}
	// 若规则匹配路径的某个前缀目录（路径是其子内容），也命中
	if strings.HasPrefix(relPath, pattern+"/") {
		return true
	}
	if dirRule && strings.HasPrefix(relPath, pattern) {
		return true
	}
	return false
}

// matchDoubleStar 处理包含 ** 的模式（跨任意层），基于路径段匹配。
func matchDoubleStar(pattern, relPath string) bool {
	patSegs := splitPath(pattern)
	pathSegs := splitPath(relPath)
	return matchSegments(patSegs, pathSegs)
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

// matchSegments 段级通配匹配：支持 *（段内任意）与 **（任意多层）。
func matchSegments(pat []string, s []string) bool {
	return matchSegs(pat, 0, s, 0)
}

func matchSegs(pat []string, pi int, s []string, si int) bool {
	if pi == len(pat) {
		return si == len(s)
	}
	seg := pat[pi]
	if seg == "**" {
		// ** 可匹配 0 或多个路径段
		if matchSegs(pat, pi+1, s, si) {
			return true
		}
		if si < len(s) {
			return matchSegs(pat, pi, s, si+1)
		}
		return false
	}
	if si >= len(s) {
		return false
	}
	if !matchSegment(seg, s[si]) {
		return false
	}
	return matchSegs(pat, pi+1, s, si+1)
}

// matchSegment 单段匹配，支持 * 通配符（段内任意字符）。
func matchSegment(pattern, s string) bool {
	ok, err := path.Match(pattern, s)
	return err == nil && ok
}

// matchBareName 无斜杠规则在整个路径任意深度递归匹配该名称。
func matchBareName(pattern, relPath string) bool {
	for _, seg := range strings.Split(relPath, "/") {
		ok, err := path.Match(pattern, seg)
		if err == nil && ok {
			return true
		}
	}
	return false
}
