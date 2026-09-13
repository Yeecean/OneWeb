package configparser

import (
	"strings"

	"github.com/yeecean/oneweb/internal/domain/config"
)

// ValidateAgainstSchema 对文档中所有语义节点执行 Schema 校验。
// 返回全部校验结果（error 与 warning 混合），调用方按 Level 区分。
func ValidateAgainstSchema(doc *config.ConfigDocument, schema *config.ConfigSchema) []config.ValidationError {
	var results []config.ValidationError
	for i := range doc.Nodes {
		n := &doc.Nodes[i]
		if n.Type == config.NodeKeyValue {
			results = append(results, schema.Validate(n.Key, n.Value)...)
		}
	}
	return results
}

// HasErrors 判断校验结果中是否存在 error 级别。
func HasErrors(results []config.ValidationError) bool {
	for _, r := range results {
		if r.Level == "error" {
			return true
		}
	}
	return false
}

// String 汇总校验结果。
func Summarize(results []config.ValidationError) string {
	var sb strings.Builder
	for _, r := range results {
		sb.WriteString(r.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}
