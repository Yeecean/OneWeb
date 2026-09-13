package onedrive

import (
	"context"

	"github.com/yeecean/oneweb/internal/domain/capability"
)

// DetectCapabilities 探测 onedrive 二进制能力矩阵。
func DetectCapabilities(ctx context.Context, cli *CLIExecutor) (*capability.FeatureSupport, error) {
	version, err := cli.Version(ctx)
	if err != nil {
		return nil, err
	}
	fs, err := capability.CapabilitiesOf(version)
	if err != nil {
		return nil, err
	}
	return &fs, nil
}
