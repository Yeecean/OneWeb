package configparser

import (
	"bufio"
	"io"
	"strings"

	"github.com/yeecean/oneweb/internal/domain/sync"
)

// ParseSyncList 逐行解析 sync_list 白名单规则。
// 空行、注释行原样保留（供编译往返保真）。
func ParseSyncList(reader io.Reader) (*sync.SyncRuleSet, error) {
	set := &sync.SyncRuleSet{}
	scanner := bufio.NewScanner(reader)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		raw := line
		rule := sync.SyncRule{
			RawText:    raw,
			LineNumber: lineNumber,
		}
		if strings.TrimSpace(line) == "" {
			rule.Type = sync.RuleBlank
			set.Rules = append(set.Rules, rule)
			continue
		}
		if strings.HasPrefix(line, "#") {
			rule.Type = sync.RuleComment
			set.Rules = append(set.Rules, rule)
			continue
		}
		if strings.HasPrefix(line, "!") {
			rule.Type = sync.RuleExclude
			rule.Pattern = strings.TrimPrefix(line, "!")
			rule.IsRooted = strings.HasPrefix(rule.Pattern, "/")
			set.Rules = append(set.Rules, rule)
			continue
		}
		rule.Type = sync.RuleInclude
		rule.Pattern = line
		rule.IsRooted = strings.HasPrefix(rule.Pattern, "/")
		set.Rules = append(set.Rules, rule)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return set, nil
}
