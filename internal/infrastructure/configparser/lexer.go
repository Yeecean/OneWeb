// Package configparser 实现配置文件的词法、语法分析与序列化。
package configparser

import (
	"strings"

	"github.com/yeecean/oneweb/internal/domain/config"
)

// lexLine 将单行文本分类并提取节点内容。
// 支持 BOM、CRLF、行内注释、带引号与不带引号的值、禁用选项识别。
func lexLine(line string, lineNumber int) config.ConfigNode {
	node := config.ConfigNode{
		RawText:    line,
		LineNumber: lineNumber,
	}

	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		node.Type = config.NodeBlankLine
		return node
	}

	// 纯注释行或禁用选项行（均以 # 开头）
	if strings.HasPrefix(trimmed, "#") {
		body := strings.TrimSpace(trimmed[1:])
		if key, value, ok := splitKeyValue(body); ok {
			node.Type = config.NodeDisabledKeyValue
			node.Key = key
			node.Value = value
		} else {
			node.Type = config.NodeComment
		}
		return node
	}

	// 键值对
	body := trimmed
	// 提取行内注释：找到值部分结束后的 # 标记
	key, value, inlineComment, ok := splitKeyValueWithComment(body)
	if !ok {
		node.Type = config.NodeComment
		return node
	}
	node.Type = config.NodeKeyValue
	node.Key = key
	node.Value = value
	node.InlineComment = inlineComment
	return node
}

// splitKeyValue 尝试从文本中分离 key 与 value。
// 若文本不是合法的 key = value 形式，返回 ok=false。
func splitKeyValue(s string) (key, value string, ok bool) {
	eq := strings.Index(s, "=")
	if eq < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:eq])
	if key == "" {
		return "", "", false
	}
	rest := strings.TrimSpace(s[eq+1:])
	// 去掉行内注释
	if idx := strings.Index(rest, " #"); idx >= 0 {
		rest = strings.TrimSpace(rest[:idx])
	}
	value = unquote(strings.TrimSpace(rest))
	if value == "" {
		return key, "", true
	}
	return key, value, true
}

// splitKeyValueWithComment 分离 key、value 与行内注释。
func splitKeyValueWithComment(s string) (key, value, comment string, ok bool) {
	eq := strings.Index(s, "=")
	if eq < 0 {
		return "", "", "", false
	}
	key = strings.TrimSpace(s[:eq])
	if key == "" {
		return "", "", "", false
	}
	rest := strings.TrimSpace(s[eq+1:])

	// 处理带引号的值：引号内的 # 不是注释
	if len(rest) > 0 && (rest[0] == '"' || rest[0] == '\'') {
		quote := rest[0]
		end := -1
		for i := 1; i < len(rest); i++ {
			if rest[i] == quote && (i == 0 || rest[i-1] != '\\') {
				end = i
				break
			}
		}
		if end >= 0 {
			value = rest[:end+1]
			remainder := strings.TrimSpace(rest[end+1:])
			if strings.HasPrefix(remainder, "#") {
				comment = strings.TrimSpace(strings.TrimPrefix(remainder, "#"))
			}
			return key, unquote(value), comment, true
		}
	}

	// 无引号值：行内注释从首个 # 开始
	if idx := strings.Index(rest, " #"); idx >= 0 {
		value = strings.TrimSpace(rest[:idx])
		comment = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(rest[idx+1:]), "#"))
	} else if strings.HasPrefix(rest, "#") {
		comment = strings.TrimSpace(strings.TrimPrefix(rest, "#"))
	} else {
		value = rest
	}
	value = unquote(strings.TrimSpace(value))
	return key, value, comment, true
}

// unquote 移除包裹字符串的引号（若存在）。
func unquote(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
