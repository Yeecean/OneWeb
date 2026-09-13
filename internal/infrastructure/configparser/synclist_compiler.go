package configparser

import (
	"strings"

	"github.com/yeecean/oneweb/internal/domain/sync"
)

// CompileSyncList 将 Rule Model 序列化回 sync_list 文本。
// 注释与空行通过 RawText 原样保留。
func CompileSyncList(set *sync.SyncRuleSet) string {
	var sb strings.Builder
	for i := range set.Rules {
		r := &set.Rules[i]
		switch r.Type {
		case sync.RuleComment, sync.RuleBlank:
			sb.WriteString(r.RawText)
		case sync.RuleExclude:
			sb.WriteString("!")
			sb.WriteString(r.Pattern)
		default:
			sb.WriteString(r.Pattern)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
