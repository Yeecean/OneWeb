package application

import (
	"testing"
	"time"

	"github.com/yeecean/oneweb/internal/domain/operation"
	"github.com/yeecean/oneweb/internal/infrastructure/events"
)

func TestOperationServiceLifecycle(t *testing.T) {
	bus := events.NewEventBus(16)
	svc := NewOperationService(bus, nil)

	op, err := svc.Create("p1", operation.TypeDryRun)
	if err != nil {
		t.Fatal(err)
	}
	if op.Status != operation.OpPending {
		t.Fatalf("status = %s", op.Status)
	}
	if err := svc.Start(op.ID); err != nil {
		t.Fatal(err)
	}
	if err := svc.TransitionToWaiting(op.ID, map[string]interface{}{"auth_url": "https://login.microsoftonline.com"}); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(op.ID)
	if got.Status != operation.OpWaitingForInput {
		t.Fatalf("status = %s", got.Status)
	}
	// 等待输入 → 回到 Running
	if err := svc.Start(op.ID); err != nil {
		t.Fatalf("return to running: %v", err)
	}
	if err := svc.Complete(op.ID, nil); err != nil {
		t.Fatal(err)
	}
	if got, _ := svc.Get(op.ID); got.Status != operation.OpSuccess {
		t.Fatalf("status = %s", got.Status)
	}
}

func TestOperationServiceListByProfile(t *testing.T) {
	bus := events.NewEventBus(16)
	svc := NewOperationService(bus, nil)

	svc.Create("p1", operation.TypeDryRun)
	svc.Create("p1", operation.TypeAuth)
	svc.Create("p2", operation.TypeSync)

	ops, _ := svc.List("p1")
	if len(ops) != 2 {
		t.Fatalf("p1 ops = %d, want 2", len(ops))
	}
	ops2, _ := svc.List("p2")
	if len(ops2) != 1 {
		t.Fatalf("p2 ops = %d, want 1", len(ops2))
	}
}

func TestOperationServiceEvents(t *testing.T) {
	bus := events.NewEventBus(16)
	svc := NewOperationService(bus, nil)

	ch := bus.Subscribe("operation.log_appended")
	defer bus.Unsubscribe(ch)

	op, _ := svc.Create("p1", operation.TypeAuth)
	_ = svc.Start(op.ID)
	svc.AppendLog(op.ID, "hello world", "stdout")

	select {
	case ev := <-ch:
		if ev.ProfileID != "p1" {
			t.Fatalf("event profile = %s", ev.ProfileID)
		}
	case <-time.After(time.Second):
		t.Fatal("no log event received")
	}
}
