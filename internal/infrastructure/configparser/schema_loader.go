package configparser

import (
	"encoding/json"
	"fmt"

	"github.com/yeecean/oneweb/internal/domain/config"
	"github.com/yeecean/oneweb/schema"
)

// LoadSemanticSchema 加载指定版本的语义 Schema。
func LoadSemanticSchema(version string) (*config.ConfigSchema, error) {
	data, err := schema.FS.ReadFile("onedrive/semantic/v" + version + ".json")
	if err != nil {
		return nil, fmt.Errorf("configparser: load semantic schema v%s: %w", version, err)
	}
	var s config.ConfigSchema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("configparser: unmarshal semantic schema: %w", err)
	}
	return &s, nil
}

// LoadUISchema 加载指定版本的 UI 元数据。
func LoadUISchema(version string) (*UISchema, error) {
	data, err := schema.FS.ReadFile("onedrive/ui/v" + version + ".json")
	if err != nil {
		return nil, fmt.Errorf("configparser: load ui schema v%s: %w", version, err)
	}
	var ui UISchema
	if err := json.Unmarshal(data, &ui); err != nil {
		return nil, fmt.Errorf("configparser: unmarshal ui schema: %w", err)
	}
	return &ui, nil
}
