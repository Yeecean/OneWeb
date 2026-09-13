package systemd

import (
	"context"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

// DetectLinger 探测指定用户是否开启 linger。
func DetectLinger(username string) (bool, error) {
	ctx := context.Background()
	out, err := exec.CommandContext(ctx, "loginctl", "show-user", username, "--property=Linger").Output()
	if err != nil {
		return false, err
	}
	// 输出形如: Linger=yes / Linger=no
	parts := strings.SplitN(strings.TrimSpace(string(out)), "=", 2)
	if len(parts) != 2 {
		return false, nil
	}
	return strings.EqualFold(strings.TrimSpace(parts[1]), "yes"), nil
}

// currentUser 返回当前用户名。
func currentUser() string {
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return os.Getenv("USER")
}
