package config

import (
	"fmt"
	"strconv"
	"strings"
)

// ValueType 定义配置值的语义类型。
type ValueType string

const (
	TypeBool   ValueType = "bool"
	TypeInt    ValueType = "int"
	TypeString ValueType = "string"
	TypePath   ValueType = "path"
	TypeEnum   ValueType = "enum"
	TypeFloat  ValueType = "float"
)

// Constraints 描述数值范围与枚举允许值。
type Constraints struct {
	Min           *int     `json:"min,omitempty"`
	Max           *int     `json:"max,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
}

// OptionSchema 描述单个配置项的语义定义。
type OptionSchema struct {
	Key         string       `json:"key"`
	Type        ValueType    `json:"type"`
	Default     interface{}  `json:"default,omitempty"`
	Description string       `json:"description,omitempty"`
	MinVersion  string       `json:"min_version,omitempty"`
	Constraints *Constraints `json:"constraints,omitempty"`
	Deprecated  bool         `json:"deprecated,omitempty"`
	Group       string       `json:"group,omitempty"`
}

// ConfigSchema 描述一组配置项。
type ConfigSchema struct {
	SchemaVersion    string         `json:"schema_version"`
	MinClientVersion string         `json:"min_client_version"`
	Options          []OptionSchema `json:"options"`
}

// GetOption 返回指定 key 的 Schema 定义。
func (s *ConfigSchema) GetOption(key string) *OptionSchema {
	for i := range s.Options {
		if s.Options[i].Key == key {
			return &s.Options[i]
		}
	}
	return nil
}

// IsKnown 判断该 key 是否为已注册配置项。
func (s *ConfigSchema) IsKnown(key string) bool {
	return s.GetOption(key) != nil
}

// ValidationError 表示一次校验失败。
type ValidationError struct {
	Key     string `json:"key"`
	Level   string `json:"level"` // "error" | "warning"
	Message string `json:"message"`
}

func (v ValidationError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", v.Level, v.Key, v.Message)
}

// Validate 根据 Schema 校验单个键值对。
// 未知选项返回 warning 而非 error（前向兼容策略）。
func (s *ConfigSchema) Validate(key, value string) []ValidationError {
	opt := s.GetOption(key)
	if opt == nil {
		return []ValidationError{{
			Key: key, Level: "warning",
			Message: "unknown option, preserved as-is (forward compatibility)",
		}}
	}
	if opt.Deprecated {
		return []ValidationError{{
			Key: key, Level: "warning", Message: "option is deprecated",
		}}
	}
	value = strings.Trim(value, `"`)

	switch opt.Type {
	case TypeBool:
		if value != "true" && value != "false" {
			return []ValidationError{{Key: key, Level: "error", Message: "expected bool (true|false)"}}
		}
	case TypeInt:
		n, err := strconv.Atoi(value)
		if err != nil {
			return []ValidationError{{Key: key, Level: "error", Message: "expected int"}}
		}
		if c := opt.Constraints; c != nil {
			if c.Min != nil && n < *c.Min {
				return []ValidationError{{Key: key, Level: "error", Message: fmt.Sprintf("below min %d", *c.Min)}}
			}
			if c.Max != nil && n > *c.Max {
				return []ValidationError{{Key: key, Level: "error", Message: fmt.Sprintf("above max %d", *c.Max)}}
			}
		}
	case TypeEnum:
		if c := opt.Constraints; c != nil && len(c.AllowedValues) > 0 {
			ok := false
			for _, a := range c.AllowedValues {
				if a == value {
					ok = true
					break
				}
			}
			if !ok {
				return []ValidationError{{Key: key, Level: "error", Message: fmt.Sprintf("not in allowed values %v", c.AllowedValues)}}
			}
		}
	}
	return nil
}
