// Package config 定义 OneWeb 配置引擎的领域模型。
// 配置文档在内存中抽象为节点树，确保任何未识别的内容都不会在反序列化或回写时丢失（架构红线 4）。
package config

import "strings"

// NodeType 描述一行配置在 AST 中的类型。
type NodeType int

const (
	// NodeKeyValue 普通键值对: key = "value"
	NodeKeyValue NodeType = iota
	// NodeDisabledKeyValue 被注释但具有语义的选项: # sync_dir = "..."
	NodeDisabledKeyValue
	// NodeComment 纯注释行: # This is a comment
	NodeComment
	// NodeBlankLine 空行
	NodeBlankLine
)

// String 返回节点类型的可读名称。
func (t NodeType) String() string {
	switch t {
	case NodeKeyValue:
		return "key_value"
	case NodeDisabledKeyValue:
		return "disabled_key_value"
	case NodeComment:
		return "comment"
	case NodeBlankLine:
		return "blank_line"
	default:
		return "unknown"
	}
}

// ConfigNode 表示 AST 中的单个节点。
type ConfigNode struct {
	Type          NodeType
	RawText       string
	Key           string
	Value         string
	InlineComment string
	LineNumber    int
	// Modified 标记节点是否在解析后被编辑，序列化时据此重建该行文本。
	Modified bool `json:"-"`
}

// IsSemantic 判断该节点是否为可编辑的语义节点（普通或禁用键值对）。
func (n *ConfigNode) IsSemantic() bool {
	return n.Type == NodeKeyValue || n.Type == NodeDisabledKeyValue
}

// ConfigDocument 表示一份完整配置文档。
type ConfigDocument struct {
	Nodes []ConfigNode
	// EndsWithNewline 记录原始文档末尾是否以换行结束，用于无损往返。
	EndsWithNewline bool
}

// Get 返回首个匹配指定 key 的键值节点；若为禁用键值节点同样返回。
func (d *ConfigDocument) Get(key string) *ConfigNode {
	for i := range d.Nodes {
		if d.Nodes[i].IsSemantic() && d.Nodes[i].Key == key {
			return &d.Nodes[i]
		}
	}
	return nil
}

// Set 修改已有键值对的值；不存在则追加到文档末尾。
func (d *ConfigDocument) Set(key, value string) {
	if n := d.Get(key); n != nil {
		n.Value = value
		n.Type = NodeKeyValue
		n.Modified = true
		return
	}
	d.Nodes = append(d.Nodes, ConfigNode{
		Type:       NodeKeyValue,
		Key:        key,
		Value:      value,
		LineNumber: len(d.Nodes) + 1,
		Modified:   true,
	})
	d.EndsWithNewline = true
}

// Enable 将禁用键值节点转换为普通键值节点。
func (d *ConfigDocument) Enable(key string) {
	if n := d.Get(key); n != nil && n.Type == NodeDisabledKeyValue {
		n.Type = NodeKeyValue
		n.Modified = true
	}
}

// Disable 将普通键值节点转换为禁用键值节点。
func (d *ConfigDocument) Disable(key string) {
	if n := d.Get(key); n != nil && n.Type == NodeKeyValue {
		n.Type = NodeDisabledKeyValue
		n.Modified = true
	}
}

// Remove 将键值节点转为纯注释（保留记录，不真删）。
func (d *ConfigDocument) Remove(key string) {
	for i := range d.Nodes {
		if d.Nodes[i].IsSemantic() && d.Nodes[i].Key == key {
			d.Nodes[i].Type = NodeComment
			d.Nodes[i].Key = ""
			d.Nodes[i].Value = ""
			d.Nodes[i].Modified = true
		}
	}
}

// Keys 返回所有已启用的 key 列表（不含禁用与注释）。
func (d *ConfigDocument) Keys() []string {
	var keys []string
	for i := range d.Nodes {
		if d.Nodes[i].Type == NodeKeyValue {
			keys = append(keys, d.Nodes[i].Key)
		}
	}
	return keys
}

// String 返回已启用键值对的文本摘要（调试用）。
func (d *ConfigDocument) String() string {
	var sb strings.Builder
	for i := range d.Nodes {
		if d.Nodes[i].Type == NodeKeyValue {
			sb.WriteString(d.Nodes[i].Key)
			sb.WriteString(" = ")
			sb.WriteString(d.Nodes[i].Value)
			sb.WriteString("\n")
		}
	}
	return sb.String()
}
