// Package schema 内嵌 OneDrive 配置语义与 UI 元数据 JSON，供后端加载。
package schema

import "embed"

// FS 暴露内嵌的 schema JSON 文件（onedrive/semantic 与 onedrive/ui）。
//
//go:embed onedrive
var FS embed.FS
