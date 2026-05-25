//go:build windows

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

	ps := `$AddType -AssemblyName System.Windows.Forms; $screen = [System.Windows.Forms.Screen]::PrimaryScreen.Bounds; $bitmap = New-Object System.Drawing.Bitmap($screen.Width, $screen.Height); $graphics = [System.Drawing.Graphics]::FromImage($bitmap); $graphics.CopyFromScreen($screen.Location, [System.Drawing.Point]::Empty, $screen.Size); $bitmap.Save('` + path + `'); $graphics.Dispose(); $bitmap.Dispose()`

	cmd := exec.Command("powershell", "-NoProfile", "-Command", ps)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("powershell screenshot: %w", err)
	}

	return path, nil
}
