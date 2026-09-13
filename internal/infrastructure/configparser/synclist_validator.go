package configparser

import (
	"strings"

	"github.com/yeecean/oneweb/internal/domain/sync"
)

// ValidationWarning 描述一条非阻塞的同步规则告警。
type ValidationWarning struct {
	Line    int    `json:"line"`
	Rule    string `json:"rule"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// ValidateSyncList 检查规则集的性能与顺序告警。
// 返回非阻塞告警（不阻断保存，但必须展示给用户）。
func ValidateSyncList(set *sync.SyncRuleSet) []ValidationWarning {
	var warnings []ValidationWarning
	// 1. 性能告警：无前导 / 的规则将在任意深度递归匹配
	for i := range set.Rules {
		r := &set.Rules[i]
		if strings.HasPrefix(r.RawText, "#") {
			continue
		}
		if !r.IsRooted {
			warnings = append(warnings, ValidationWarning{
				Line:    r.LineNumber,
				Rule:    r.RawText,
				Level:   "performance",
				Message: "rule without leading '/' will recursively match at any depth — high CPU/IO cost",
			})
		}
	}
	// 2. 顺序告警：排除规则出现在包含规则之后（后置规则优先级更高，排除可能被覆盖）
	active := false // 是否已出现包含规则
	for i := range set.Rules {
		r := &set.Rules[i]
		if strings.HasPrefix(r.RawText, "#") {
			continue
		}
		if r.Type == sync.RuleExclude && active {
			warnings = append(warnings, ValidationWarning{
				Line:    r.LineNumber,
				Rule:    r.RawText,
				Level:   "ordering",
				Message: "exclude rule appears after an include rule — later rules take precedence",
			})
		}
		if r.Type == sync.RuleInclude {
			active = true
		}
	}
	return warnings
}
