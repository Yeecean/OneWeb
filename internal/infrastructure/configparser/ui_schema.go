package configparser

// UISchema 描述配置项的界面呈现元数据。
type UISchema struct {
	SchemaVersion string              `json:"schema_version"`
	Groups        []UIGroup           `json:"groups"`
	Widgets       map[string]UIWidget `json:"widgets"`
}

// UIGroup 是配置分组元数据。
type UIGroup struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Order int    `json:"order"`
}

// UIWidget 描述单个配置项的界面组件。
type UIWidget struct {
	Widget   string `json:"widget"`
	Unit     string `json:"unit,omitempty"`
	Advanced bool   `json:"advanced"`
}
