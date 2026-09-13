// Package operation 定义异步工作单元（Operation）的聚合与状态机。
package operation

import (
	"errors"
	"fmt"
	"time"
)

// OperationStatus 是操作生命周期的全部 9 种枚举状态。
type OperationStatus string

const (
	OpPending         OperationStatus = "pending"
	OpRunning         OperationStatus = "running"
	OpWaitingForInput OperationStatus = "waiting_for_input"
	OpCancelRequested OperationStatus = "cancel_requested"
	OpSuccess         OperationStatus = "success"
	OpFailed          OperationStatus = "failed"
	OpCanceled        OperationStatus = "canceled"
	OpTimedOut        OperationStatus = "timed_out"
	OpUnknown         OperationStatus = "unknown"
)

// OperationType 枚举可执行的操作种类。
type OperationType string

const (
	TypeAuth          OperationType = "auth"
	TypeSync          OperationType = "sync"
	TypeDryRun        OperationType = "dry-run"
	TypeResync        OperationType = "resync"
	TypeValidate      OperationType = "validate"
	TypeTreeDiscovery OperationType = "tree-discovery"
)

// Operation 是一个耗时工作单元的聚合根。
type Operation struct {
	ID        string          `json:"id"`
	ProfileID string          `json:"profile_id"`
	Type      OperationType   `json:"type"`
	Status    OperationStatus `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	StartedAt *time.Time      `json:"started_at,omitempty"`
	EndedAt   *time.Time      `json:"ended_at,omitempty"`
	Logs      []string        `json:"logs,omitempty"`
}

// ErrIllegalTransition 表示状态机拒绝了一次非法转换。
var ErrIllegalTransition = errors.New("operation: illegal status transition")

// terminal 判断是否为终态。
func terminal(s OperationStatus) bool {
	switch s {
	case OpSuccess, OpFailed, OpCanceled, OpTimedOut:
		return true
	}
	return false
}

// validTransitions 定义状态机允许的转换规则。
var validTransitions = map[OperationStatus][]OperationStatus{
	OpPending:         {OpRunning, OpFailed},
	OpRunning:         {OpWaitingForInput, OpCancelRequested, OpSuccess, OpFailed, OpTimedOut},
	OpWaitingForInput: {OpRunning, OpTimedOut},
	OpCancelRequested: {OpCanceled},
	OpUnknown:         {OpRunning, OpFailed},
}

// Transition 在状态机守卫下迁移操作状态。
// 终态不可再转换；非法转换返回 ErrIllegalTransition。
func (o *Operation) Transition(newStatus OperationStatus) error {
	if terminal(o.Status) {
		return fmt.Errorf("%w: from terminal state %s", ErrIllegalTransition, o.Status)
	}
	for _, allowed := range validTransitions[o.Status] {
		if allowed == newStatus {
			now := time.Now()
			if o.Status == OpPending && newStatus == OpRunning {
				o.StartedAt = &now
			}
			if terminal(newStatus) {
				o.EndedAt = &now
			}
			o.Status = newStatus
			return nil
		}
	}
	return fmt.Errorf("%w: %s -> %s", ErrIllegalTransition, o.Status, newStatus)
}

// AppendLog 追加一条操作日志记录。
func (o *Operation) AppendLog(line string) {
	o.Logs = append(o.Logs, line)
}
