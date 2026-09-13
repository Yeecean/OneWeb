package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// AtomicWriteFile 以 POSIX 原子替换方式写入文件：
// 1. 同目录创建临时文件 config.tmp.{random}
// 2. 写入全部内容
// 3. fsync(fd) 刷盘文件数据
// 4. Close(fd)
// 5. os.Rename(tmp, target) 原子替换
// 6. 打开父目录 → fsync(dirfd) 刷盘目录元数据
// 7. 错误时自动清理临时文件
func AtomicWriteFile(targetPath string, content []byte, perm os.FileMode) error {
	dir := filepath.Dir(targetPath)
	tmp, err := os.CreateTemp(dir, filepath.Base(targetPath)+".tmp.*")
	if err != nil {
		return fmt.Errorf("filesystem: create temp: %w", err)
	}
	tmpPath := tmp.Name()

	cleanup := func() {
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("filesystem: write temp: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("filesystem: chmod temp: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		cleanup()
		return fmt.Errorf("filesystem: fsync temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("filesystem: close temp: %w", err)
	}
	if err := os.Rename(tmpPath, targetPath); err != nil {
		cleanup()
		return fmt.Errorf("filesystem: rename: %w", err)
	}
	// fsync 父目录元数据
	d, err := os.Open(dir)
	if err == nil {
		_ = d.Sync()
		_ = d.Close()
	}
	return nil
}

// FsyncDir 对目录执行 fsync 以持久化目录项元数据。
func FsyncDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	if err := d.Sync(); err != nil {
		return fmt.Errorf("filesystem: fsync dir %s: %w", dir, err)
	}
	return nil
}

// FsyncFile 对已打开文件执行 fsync。
func FsyncFile(f *os.File) error {
	return syscall.Fsync(int(f.Fd()))
}
