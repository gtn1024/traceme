//go:build darwin

package screenshot

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func Capture() (string, error) {
	tmpDir := os.TempDir()
	path := filepath.Join(tmpDir, "traceme_capture.png")

	cmd := exec.Command("screencapture", "-x", path)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("screencapture: %w", err)
	}

	return path, nil
}
