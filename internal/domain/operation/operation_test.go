package operation

import (
	"errors"
	"testing"
)

func TestLegalTransitions(t *testing.T) {
	tests := []struct {
		name  string
		from  OperationStatus
		to    OperationStatus
		valid bool
	}{
		{"pending->running", OpPending, OpRunning, true},
		{"pending->failed", OpPending, OpFailed, true},
		{"running->waiting", OpRunning, OpWaitingForInput, true},
		{"running->cancel_requested", OpRunning, OpCancelRequested, true},
		{"running->success", OpRunning, OpSuccess, true},
		{"running->failed", OpRunning, OpFailed, true},
		{"running->timed_out", OpRunning, OpTimedOut, true},
		{"waiting->running", OpWaitingForInput, OpRunning, true},
		{"waiting->timed_out", OpWaitingForInput, OpTimedOut, true},
		{"cancel_requested->canceled", OpCancelRequested, OpCanceled, true},
		{"unknown->running", OpUnknown, OpRunning, true},
		{"unknown->failed", OpUnknown, OpFailed, true},
		{"pending->success", OpPending, OpSuccess, false},
		{"running->pending", OpRunning, OpPending, false},
		{"waiting->cancel_requested", OpWaitingForInput, OpCancelRequested, false},
		{"success->running", OpSuccess, OpRunning, false},
		{"canceled->running", OpCanceled, OpRunning, false},
		{"timed_out->success", OpTimedOut, OpSuccess, false},
		{"failed->running", OpFailed, OpRunning, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := &Operation{Status: tt.from}
			err := op.Transition(tt.to)
			if tt.valid && err != nil {
				t.Fatalf("transition %s->%s should be valid, got err %v", tt.from, tt.to, err)
			}
			if !tt.valid && !errors.Is(err, ErrIllegalTransition) {
				t.Fatalf("transition %s->%s should be illegal, got err %v", tt.from, tt.to, err)
			}
		})
	}
}

func TestTerminalStatesImmutable(t *testing.T) {
	for _, terminal := range []OperationStatus{OpSuccess, OpFailed, OpCanceled, OpTimedOut} {
		op := &Operation{Status: terminal}
		if err := op.Transition(OpRunning); !errors.Is(err, ErrIllegalTransition) {
			t.Fatalf("terminal %s must be immutable, got %v", terminal, err)
		}
	}
}

func TestStartedAtSetOnFirstRun(t *testing.T) {
	op := &Operation{Status: OpPending}
	_ = op.Transition(OpRunning)
	if op.StartedAt == nil {
		t.Fatal("StartedAt must be set when entering running")
	}
}

func TestEndedAtSetOnTerminal(t *testing.T) {
	op := &Operation{Status: OpPending}
	_ = op.Transition(OpRunning)
	_ = op.Transition(OpSuccess)
	if op.EndedAt == nil {
		t.Fatal("EndedAt must be set when entering terminal state")
	}
}
