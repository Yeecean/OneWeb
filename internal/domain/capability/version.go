// Package capability 定义客户端版本能力规约。
package capability

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ClientVersion 表示 onedrive 客户端版本。
type ClientVersion struct {
	Major int
	Minor int
	Patch int
}

// versionPattern 匹配 "onedrive v2.5.11" / "v2.5.11" / "2.5.11" 等格式。
var versionPattern = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// ParseVersion 从版本字符串中解析出 ClientVersion。
func ParseVersion(versionString string) (ClientVersion, error) {
	m := versionPattern.FindStringSubmatch(versionString)
	if m == nil {
		return ClientVersion{}, fmt.Errorf("capability: no semver found in %q", versionString)
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return ClientVersion{Major: major, Minor: minor, Patch: patch}, nil
}

func (v ClientVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// AtLeast 判断当前版本是否不低于给定版本。
func (v ClientVersion) AtLeast(o ClientVersion) bool {
	if v.Major != o.Major {
		return v.Major > o.Major
	}
	if v.Minor != o.Minor {
		return v.Minor > o.Minor
	}
	return v.Patch >= o.Patch
}

// FeatureSupport 描述一个版本下可用的功能矩阵。
type FeatureSupport struct {
	ConfigEditor  bool   `json:"config_editor"`
	SyncList      bool   `json:"sync_list"`
	DeviceAuth    bool   `json:"device_auth"`
	RuntimeCtrl   bool   `json:"runtime_ctrl"`
	AuthAssistant bool   `json:"auth_assistant"`
	SchemaVersion string `json:"schema_version"`
	Recommended   bool   `json:"recommended"`
}

// Evaluate 根据版本区间返回功能矩阵。
func Evaluate(version ClientVersion) FeatureSupport {
	s := FeatureSupport{
		ConfigEditor:  true,
		RuntimeCtrl:   true,
		AuthAssistant: true,
		Recommended:   true,
	}
	switch {
	case version.Major < 2, version.Major == 2 && version.Minor < 4:
		s.ConfigEditor = false
		s.RuntimeCtrl = false
		s.AuthAssistant = false
		s.Recommended = false
		s.SchemaVersion = "legacy"
		return s
	case version.Minor == 4:
		if version.Patch < 25 {
			s.Recommended = false
			s.ConfigEditor = false
		}
		s.SchemaVersion = "2.4"
		return s
	default:
		s.SchemaVersion = "2.5"
		s.SyncList = true
		s.DeviceAuth = true
		return s
	}
}

// CapabilitiesOf 组合解析与评估。
func CapabilitiesOf(versionString string) (FeatureSupport, error) {
	versionString = strings.TrimSpace(versionString)
	v, err := ParseVersion(versionString)
	if err != nil {
		return FeatureSupport{}, err
	}
	return Evaluate(v), nil
}
