package onedrive

import (
	"regexp"
	"strings"
)

// authURLPattern 匹配 onedrive 输出的微软授权 URL。
var authURLPattern = regexp.MustCompile(`https://login\.microsoftonline\.com[^\s"']+`)

// deviceCodePattern 匹配设备代码（DEVICE CODE）输出。
var deviceCodePattern = regexp.MustCompile(`code:\s*([A-Z0-9]{6,})`)

// successPattern 匹配认证成功标志。
var successPattern = regexp.MustCompile(`(?i)(successfully authenticated|auth.*success|application.*authenticated)`)

// failurePattern 匹配认证失败标志。
var failurePattern = regexp.MustCompile(`(?i)(authorization.*failed|auth.*failed|error.*token|invalid.*grant|access denied)`)

// AuthParser 提供 onedrive CLI 认证输出的解析能力。
type AuthParser struct{}

// ParseAuthURL 从一行输出中提取 Microsoft 登录 URL。
func (p *AuthParser) ParseAuthURL(line string) (string, bool) {
	m := authURLPattern.FindString(line)
	if m == "" {
		return "", false
	}
	return m, true
}

// ParseDeviceCode 从一行输出中提取设备代码。
func (p *AuthParser) ParseDeviceCode(line string) (string, bool) {
	m := deviceCodePattern.FindStringSubmatch(line)
	if len(m) < 2 {
		return "", false
	}
	return m[1], true
}

// IsAuthSuccess 判断是否为认证成功标志。
func (p *AuthParser) IsAuthSuccess(line string) bool {
	return successPattern.MatchString(line)
}

// IsAuthFailure 判断是否为认证失败标志。
func (p *AuthParser) IsAuthFailure(line string) bool {
	return failurePattern.MatchString(line)
}

// IsDeviceAuthRequest 判断是否为设备码认证提示。
func (p *AuthParser) IsDeviceAuthRequest(line string) bool {
	return strings.Contains(strings.ToLower(line), "device code") ||
		strings.Contains(strings.ToLower(line), "device_code")
}
