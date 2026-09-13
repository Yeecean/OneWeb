// Package filesystem 提供原子写入、指纹计算与 Profile 持久化等基础设施。
package filesystem

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"time"
)

// FileFingerprint 描述文件的指纹快照，用于外部漂移检测。
type FileFingerprint struct {
	Mtime  time.Time `json:"mtime"`
	Size   int64     `json:"size"`
	SHA256 string    `json:"sha256"`
}

// ComputeFingerprint 计算文件的 mtime、size 与 sha256。
func ComputeFingerprint(path string) (*FileFingerprint, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, os.ErrNotExist
		}
		return nil, fmt.Errorf("filesystem: stat %s: %w", path, err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("filesystem: open %s: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("filesystem: hash %s: %w", path, err)
	}
	return &FileFingerprint{
		Mtime:  fi.ModTime(),
		Size:   fi.Size(),
		SHA256: hex.EncodeToString(h.Sum(nil)),
	}, nil
}

// CompareFingerprint 比对磁盘当前指纹与期望指纹是否一致。
// 文件不存在返回 (false, nil)。
func CompareFingerprint(path string, expected *FileFingerprint) (bool, error) {
	actual, err := ComputeFingerprint(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	if actual.SHA256 != expected.SHA256 {
		return false, nil
	}
	return true, nil
}
