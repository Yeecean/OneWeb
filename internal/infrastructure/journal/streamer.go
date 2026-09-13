// Package journal 实现 journalctl 日志流式读取。
package journal

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/yeecean/oneweb/internal/domain/runtime"
)

// JournalEntry 表示一条 journalctl JSON 输出。
// PRIORITY 在不同 journal 版本中可能是数字或字符串，使用灵活类型解码。
type JournalEntry struct {
	Message    string  `json:"MESSAGE"`
	Priority   FlexInt `json:"PRIORITY"`
	RealtimeTS string  `json:"__REALTIME_TIMESTAMP"`
}

// FlexInt 兼容 JSON 数字与字符串的数字类型。
type FlexInt int

// UnmarshalJSON 兼容两种编码形式。
func (f *FlexInt) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)
	n, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	*f = FlexInt(n)
	return nil
}

var priorityLevel = map[int]string{
	0: "emerg", 1: "alert", 2: "crit", 3: "err",
	4: "warning", 5: "notice", 6: "info", 7: "debug",
}

// StreamLogs 通过 journalctl --user-unit --follow --output=json 流式读取日志。
// ctx 取消时优雅终止子进程并关闭 channel。
func StreamLogs(ctx context.Context, unit string, opts runtime.LogOptions) (<-chan runtime.LogEntry, error) {
	args := []string{
		"--user-unit=" + unit,
		"--output=json",
		"--no-pager",
	}
	if opts.Since != nil {
		args = append(args, "--since="+opts.Since.Format(time.RFC3339))
	}
	if !opts.Follow {
		args = append(args, "--no-tail")
	}
	if opts.Lines > 0 {
		args = append(args, "-n", strconv.Itoa(opts.Lines))
	}
	if opts.Follow {
		args = append(args, "-f")
	}

	cmd := exec.CommandContext(ctx, "journalctl", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("journal: stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("journal: start: %w", err)
	}

	ch := make(chan runtime.LogEntry, 128)
	go func() {
		defer close(ch)
		defer func() {
			_ = cmd.Wait()
		}()
		scanner := bufio.NewScanner(stdout)
		scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
		for scanner.Scan() {
			var je JournalEntry
			if err := json.Unmarshal(scanner.Bytes(), &je); err != nil {
				// 非 JSON 行按原始文本输出
				ch <- runtime.LogEntry{
					Timestamp: time.Now(),
					Stream:    runtime.StreamStdout,
					Message:   scanner.Text(),
				}
				continue
			}
			ts := time.Now()
			if micros, err := strconv.ParseInt(je.RealtimeTS, 10, 64); err == nil {
				ts = time.UnixMicro(micros)
			}
			level := priorityLevel[int(je.Priority)]
			if level == "" {
				level = "info"
			}
			stream := runtime.StreamStdout
			if level == "err" || level == "crit" || level == "alert" || level == "emerg" {
				stream = runtime.StreamStderr
			}
			ch <- runtime.LogEntry{
				Timestamp: ts,
				Stream:    stream,
				Level:     level,
				Message:   strings.TrimSpace(je.Message),
			}
		}
	}()
	return ch, nil
}
