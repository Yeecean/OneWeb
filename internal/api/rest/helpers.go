// Package rest 实现 OneWeb 的 REST API 接入层。
package rest

import (
	"encoding/json"
	"errors"
	"net/http"
)

// ErrNotFound 表示资源不存在。
var ErrNotFound = errors.New("rest: not found")

// errorBody 统一错误响应格式。
type errorBody struct {
	Error apiError `json:"error"`
}

// apiError 描述一个 API 错误。
type apiError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// writeJSON 写 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError 写统一错误响应。
func writeError(w http.ResponseWriter, status int, code, message string, details interface{}) {
	writeJSON(w, status, errorBody{Error: apiError{
		Code:    code,
		Message: message,
		Details: details,
	}})
}

// decodeJSON 解码请求体到目标结构。
func decodeJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}
