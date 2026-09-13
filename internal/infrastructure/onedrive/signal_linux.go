package onedrive

import (
	"syscall"
)

// sigTerm 返回适合当前平台的 SIGTERM。
func sigTerm() syscall.Signal {
	return syscall.SIGTERM
}
