package application

import (
	"context"
	"testing"

	"github.com/yeecean/oneweb/internal/domain/profile"
	"github.com/yeecean/oneweb/internal/domain/runtime"
)

type mockRuntimeBackend struct {
	started   bool
	stopped   bool
	restarted bool
	state     runtime.RuntimeState
}

func (m *mockRuntimeBackend) Detect(ctx context.Context) (runtime.Capability, error) {
	return runtime.Capability{Available: true}, nil
}
func (m *mockRuntimeBackend) Start(ctx context.Context, p profile.Profile) error {
	m.started = true
	m.state = runtime.StateRunning
	return nil
}
func (m *mockRuntimeBackend) Stop(ctx context.Context, p profile.Profile) error {
	m.stopped = true
	m.state = runtime.StateStopped
	return nil
}
func (m *mockRuntimeBackend) Restart(ctx context.Context, p profile.Profile) error {
	m.restarted = true
	m.state = runtime.StateRunning
	return nil
}
func (m *mockRuntimeBackend) Status(ctx context.Context, p profile.Profile) (runtime.RuntimeStatus, error) {
	return runtime.RuntimeStatus{State: m.state}, nil
}
func (m *mockRuntimeBackend) Logs(ctx context.Context, p profile.Profile, opts runtime.LogOptions) (<-chan runtime.LogEntry, error) {
	ch := make(chan runtime.LogEntry)
	close(ch)
	return ch, nil
}

func TestRuntimeServiceControlActions(t *testing.T) {
	mock := &mockRuntimeBackend{state: runtime.StateStopped}
	svc := NewRuntimeService(map[profile.RuntimeType]runtime.RuntimeBackend{
		profile.RuntimeSystemd: mock,
	})
	p := &profile.Profile{
		ID:          "default",
		RuntimeType: profile.RuntimeSystemd,
	}

	// 1. start
	if err := svc.ControlRuntime(context.Background(), p, "start"); err != nil {
		t.Fatalf("unexpected error on start: %v", err)
	}
	if !mock.started {
		t.Fatalf("expected started to be true")
	}

	// 2. sync when running -> should trigger Restart
	mock.restarted = false
	mock.state = runtime.StateRunning
	if err := svc.ControlRuntime(context.Background(), p, "sync"); err != nil {
		t.Fatalf("unexpected error on sync: %v", err)
	}
	if !mock.restarted {
		t.Fatalf("expected restarted to be true when sync is called in running state")
	}

	// 3. sync when stopped -> should trigger Start
	mock.started = false
	mock.restarted = false
	mock.state = runtime.StateStopped
	if err := svc.ControlRuntime(context.Background(), p, "sync"); err != nil {
		t.Fatalf("unexpected error on sync when stopped: %v", err)
	}
	if !mock.started {
		t.Fatalf("expected started to be true when sync is called in stopped state")
	}

	// 4. stop
	if err := svc.ControlRuntime(context.Background(), p, "stop"); err != nil {
		t.Fatalf("unexpected error on stop: %v", err)
	}
	if !mock.stopped {
		t.Fatalf("expected stopped to be true")
	}

	// 5. invalid action
	if err := svc.ControlRuntime(context.Background(), p, "invalid"); err != ErrInvalidAction {
		t.Fatalf("expected ErrInvalidAction, got %v", err)
	}
}
