// Package domain 定义跨子领域的公共领域概念（领域事件）。
package domain

import "time"

// EventType 是领域事件的类型标识。
type EventType string

const (
	EvtOpStarted     EventType = "operation.started"
	EvtOpLog         EventType = "operation.log_appended"
	EvtOpNeedInput   EventType = "operation.waiting_input"
	EvtOpCompleted   EventType = "operation.completed"
	EvtOpFailed      EventType = "operation.failed"
	EvtRuntimeChange EventType = "runtime.state_changed"
)

// DomainEvent 表示一条领域事件，由领域对象触发并投递至事件总线。
type DomainEvent struct {
	ID        string      `json:"id"`
	Type      EventType   `json:"type"`
	ProfileID string      `json:"profile_id"`
	Timestamp time.Time   `json:"timestamp"`
	Payload   interface{} `json:"payload,omitempty"`
}
