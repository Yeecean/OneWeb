package configparser

import (
	"strings"

	"github.com/yeecean/oneweb/internal/domain/config"
)

// Render 将 ConfigDocument 重新序列化回纯文本。
// 未修改节点使用原始文本（RawText）保证逐字节保真；
// 修改节点根据语义重建行文本。
func Render(doc *config.ConfigDocument) string {
	var sb strings.Builder
	for i := range doc.Nodes {
		n := &doc.Nodes[i]
		if n.Modified {
			sb.WriteString(renderModified(n))
		} else {
			sb.WriteString(n.RawText)
		}
		sb.WriteString("\n")
	}
	if !doc.EndsWithNewline {
		out := sb.String()
		if strings.HasSuffix(out, "\n") {
			out = strings.TrimSuffix(out, "\n")
		}
		return out
	}
	return sb.String()
}

// renderModified 为被编辑的节点重建行文本。
func renderModified(n *config.ConfigNode) string {
	switch n.Type {
	case config.NodeKeyValue:
		var sb strings.Builder
		sb.WriteString(n.Key)
		sb.WriteString(" = ")
		sb.WriteString(quoteValue(n.Value))
		if n.InlineComment != "" {
			sb.WriteString(" # ")
			sb.WriteString(n.InlineComment)
		}
		return sb.String()
	case config.NodeDisabledKeyValue:
		return "# " + n.Key + " = " + quoteValue(n.Value)
	case config.NodeComment:
		return "# " + n.Value
	case config.NodeBlankLine:
		return ""
	default:
		return n.RawText
	}
}

// quoteValue 始终为值加引号。
// 依据 onedrive 官方解析规则：所有配置值（含数字、布尔）必须带引号，
// 否则会被判定为 "Malformed config line"。
func quoteValue(v string) string {
	return `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
}
