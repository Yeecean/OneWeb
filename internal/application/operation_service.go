package application

import (
	"fmt"
	"sync"
	"time"

	"github.com/yeecean/oneweb/internal/domain"
	"github.com/yeecean/oneweb/internal/domain/operation"
	"github.com/yeecean/oneweb/internal/infrastructure/events"
	"github.com/yeecean/oneweb/internal/infrastructure/onedrive"
)

// ErrOperationNotFound 表示操作不存在。
var ErrOperationNotFound = fmt.Errorf("application: operation not found")

// ErrOperationNotWaiting 表示操作不处于等待输入状态。
var ErrOperationNotWaiting = fmt.Errorf("application: operation is not waiting for input")

// OperationService 管理异步操作生命周期。
type OperationService struct {
	EventBus   *events.EventBus
	ProcessMgr *onedrive.ProcessManager
	CLI        *onedrive.CLIExecutor

	mu        sync.Mutex
	registry  map[string]*operation.Operation
	processes map[string]*onedrive.ProcessHandle
}

// NewOperationService 创建操作服务。
func NewOperationService(eventBus *events.EventBus, cli *onedrive.CLIExecutor) *OperationService {
	return &OperationService{
		EventBus:   eventBus,
		ProcessMgr: &onedrive.ProcessManager{},
		CLI:        cli,
		registry:   map[string]*operation.Operation{},
		processes:  map[string]*onedrive.ProcessHandle{},
	}
}

// Create 创建操作（初始 Pending）。
func (s *OperationService) Create(profileID string, opType operation.OperationType) (*operation.Operation, error) {
	op := &operation.Operation{
		ID:        newOpID(),
		ProfileID: profileID,
		Type:      opType,
		Status:    operation.OpPending,
		CreatedAt: time.Now(),
	}
	s.mu.Lock()
	s.registry[op.ID] = op
	s.mu.Unlock()
	return op, nil
}

// Get 获取操作。
func (s *OperationService) Get(opID string) (*operation.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	op, ok := s.registry[opID]
	if !ok {
		return nil, ErrOperationNotFound
	}
	return op, nil
}

// List 列出指定 Profile 的操作。
func (s *OperationService) List(profileID string) ([]*operation.Operation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []*operation.Operation
	for _, op := range s.registry {
		if op.ProfileID == profileID {
			out = append(out, op)
		}
	}
	return out, nil
}

// Start 启动操作（Pending → Running），由执行器驱动。
func (s *OperationService) Start(opID string) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	if err := op.Transition(operation.OpRunning); err != nil {
		return err
	}
	s.publish(op, domain.EvtOpStarted, nil)
	return nil
}

// Cancel 请求取消操作。
func (s *OperationService) Cancel(opID string) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	s.mu.Lock()
	handle := s.processes[opID]
	s.mu.Unlock()
	if handle != nil {
		_ = handle.Cancel()
	}
	return op.Transition(operation.OpCancelRequested)
}

// SendInput 向等待输入的操作注入输入。
func (s *OperationService) SendInput(opID string, input string) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	if op.Status != operation.OpWaitingForInput {
		return ErrOperationNotWaiting
	}
	s.mu.Lock()
	handle := s.processes[opID]
	s.mu.Unlock()
	if handle == nil {
		return fmt.Errorf("application: no live process for operation %s", opID)
	}
	if err := handle.WriteInput(input + "\n"); err != nil {
		return err
	}
	return op.Transition(operation.OpRunning)
}

// Complete 标记操作成功。
func (s *OperationService) Complete(opID string, payload interface{}) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	if err := op.Transition(operation.OpSuccess); err != nil {
		return err
	}
	s.cleanup(opID)
	s.publish(op, domain.EvtOpCompleted, payload)
	return nil
}

// Fail 标记操作失败。
func (s *OperationService) Fail(opID string, payload interface{}) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	if err := op.Transition(operation.OpFailed); err != nil {
		return err
	}
	s.cleanup(opID)
	s.publish(op, domain.EvtOpFailed, payload)
	return nil
}

// TransitionToWaiting 将操作转入等待输入状态。
func (s *OperationService) TransitionToWaiting(opID string, payload interface{}) error {
	op, err := s.Get(opID)
	if err != nil {
		return err
	}
	if err := op.Transition(operation.OpWaitingForInput); err != nil {
		return err
	}
	s.publish(op, domain.EvtOpNeedInput, payload)
	return nil
}

// AttachProcess 将子进程句柄关联到操作。
func (s *OperationService) AttachProcess(opID string, handle *onedrive.ProcessHandle) {
	s.mu.Lock()
	s.processes[opID] = handle
	s.mu.Unlock()
}

// AppendLog 追加操作日志并发布日志事件。
func (s *OperationService) AppendLog(opID string, line string, stream string) {
	op, err := s.Get(opID)
	if err != nil {
		return
	}
	op.AppendLog(line)
	s.publish(op, domain.EvtOpLog, map[string]interface{}{
		"stream": stream,
		"level":  "info",
		"raw":    line,
	})
}

func (s *OperationService) publish(op *operation.Operation, et domain.EventType, payload interface{}) {
	s.EventBus.Publish(domain.DomainEvent{
		ID:        op.ID,
		Type:      et,
		ProfileID: op.ProfileID,
		Timestamp: time.Now(),
		Payload:   payload,
	})
}

func (s *OperationService) cleanup(opID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.processes, opID)
}

func newOpID() string {
	return fmt.Sprintf("op-%d-%d", time.Now().UnixNano(), time.Now().UnixNano()%1000)
}
