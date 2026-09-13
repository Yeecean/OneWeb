package onedrive

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
	"time"
)

// ProcessOptions 描述子进程启动选项。
type ProcessOptions struct {
	// Args 追加在二进制路径后的参数。
	Args []string
	// Dir 工作目录。
	Dir string
	// GracePeriod 发送 SIGTERM 后等待强制 SIGKILL 的时长。
	GracePeriod time.Duration
}

// ProcessResult 描述子进程的最终退出信息。
type ProcessResult struct {
	ExitCode int
	Err      error
}

// ProcessHandle 暴露子进程的管道与生命周期。
type ProcessHandle struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
	Stdin  io.WriteCloser
	Stdout <-chan string
	Stderr <-chan string
	Done   <-chan ProcessResult

	doneOnce sync.Once
	done     chan ProcessResult
}

// processManager 管理 onedrive CLI 子进程生命周期。
type ProcessManager struct{}

// Run 启动子进程并返回句柄。
// ctx 取消时发送 SIGTERM，宽限期后 SIGKILL。
func (m *ProcessManager) Run(ctx context.Context, binaryPath string, opts ProcessOptions) (*ProcessHandle, error) {
	if opts.GracePeriod <= 0 {
		opts.GracePeriod = 5 * time.Second
	}
	procCtx, cancel := context.WithCancel(ctx)
	cmd := exec.CommandContext(procCtx, binaryPath, opts.Args...)
	cmd.Dir = opts.Dir

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		cancel()
		return nil, err
	}

	h := &ProcessHandle{
		cmd:    cmd,
		cancel: cancel,
		Stdin:  stdin,
		done:   make(chan ProcessResult, 1),
	}
	h.Stdout = lineChannel(stdout)
	h.Stderr = lineChannel(stderr)
	h.Done = h.done

	// 进程结束
	go func() {
		err := cmd.Wait()
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else if ctx.Err() != nil {
				exitCode = -1
			}
		}
		h.finish(ProcessResult{ExitCode: exitCode, Err: err})
	}()

	go func() {
		select {
		case <-ctx.Done():
			// 取消时先 SIGTERM，宽限后 SIGKILL
			_ = cmd.Process.Signal(sigTerm())
			select {
			case <-h.Done:
				return
			case <-time.After(opts.GracePeriod):
				_ = cmd.Process.Kill()
			}
		case <-h.Done:
			return
		}
	}()

	return h, nil
}

// Cancel 主动请求子进程取消（SIGTERM）。
func (h *ProcessHandle) Cancel() error {
	if h.cancel != nil {
		h.cancel()
	}
	if h.cmd != nil && h.cmd.Process != nil {
		return h.cmd.Process.Signal(sigTerm())
	}
	return nil
}

func (h *ProcessHandle) finish(r ProcessResult) {
	h.doneOnce.Do(func() {
		h.done <- r
		close(h.done)
	})
}

// WriteInput 向子进程 stdin 写入输入。
func (h *ProcessHandle) WriteInput(input string) error {
	_, err := io.WriteString(h.Stdin, input)
	if err != nil {
		return err
	}
	return nil
}

// lineChannel 将 reader 逐行读取推送到 channel。
func lineChannel(r io.Reader) <-chan string {
	ch := make(chan string, 128)
	go func() {
		defer close(ch)
		scanner := bufio.NewScanner(r)
		scanner.Buffer(make([]byte, 64*1024), 64*1024)
		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()
	return ch
}
