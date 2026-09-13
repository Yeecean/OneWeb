package onedrive

import (
	"context"
	"testing"
	"time"
)

func TestProcessManagerRunCompletes(t *testing.T) {
	m := &ProcessManager{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	h, err := m.Run(ctx, "echo", ProcessOptions{Args: []string{"hello"}, GracePeriod: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case line := <-h.Stdout:
		if line != "hello" {
			t.Fatalf("stdout = %q", line)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting stdout")
	}
	select {
	case res := <-h.Done:
		if res.ExitCode != 0 {
			t.Fatalf("exit code = %d", res.ExitCode)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting done")
	}
}

func TestProcessManagerStdin(t *testing.T) {
	m := &ProcessManager{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// cat 回显 stdin
	h, err := m.Run(ctx, "cat", ProcessOptions{GracePeriod: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if err := h.WriteInput("ping\n"); err != nil {
		t.Fatal(err)
	}
	select {
	case line := <-h.Stdout:
		if line != "ping" {
			t.Fatalf("stdout = %q", line)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting echo")
	}
	cancel()
}

func TestProcessManagerCancel(t *testing.T) {
	m := &ProcessManager{}
	ctx := context.Background()
	h, err := m.Run(ctx, "sleep", ProcessOptions{Args: []string{"30"}, GracePeriod: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	select {
	case res := <-h.Done:
		t.Fatalf("process should not exit before cancel, got %+v", res)
	case <-time.After(200 * time.Millisecond):
		// 正常运行
	}
	if err := h.Cancel(); err != nil {
		t.Fatal(err)
	}
	select {
	case res := <-h.Done:
		// SIGTERM 后进程退出即算取消成功（sleep 可能以 0 退出）
		if res.Err != nil && res.ExitCode == 0 {
			t.Fatalf("unexpected result: %+v", res)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("process did not exit after cancel")
	}
}
