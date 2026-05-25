package launchd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const label = "com.traceme"

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", label+".plist")
}

func binaryPath() (string, error) {
	if bin, err := exec.LookPath("traceme"); err == nil {
		return bin, nil
	}
	if exe, err := os.Executable(); err == nil {
		return exe, nil
	}
	return "", fmt.Errorf("cannot determine traceme binary path")
}

func generatePlist() (string, error) {
	bin, err := binaryPath()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>%s/.traceme/traceme.err.log</string>
    <key>StandardOutPath</key>
    <string>%s/.traceme/traceme.out.log</string>
</dict>
</plist>`, label, bin, homeDir(), homeDir()), nil
}

func homeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

func Install() error {
	pl, err := generatePlist()
	if err != nil {
		return err
	}

	path := plistPath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create LaunchAgents dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(pl), 0644); err != nil {
		return fmt.Errorf("write plist: %w", err)
	}

	cmd := exec.Command("launchctl", "load", path)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("launchctl load: %w", err)
	}

	fmt.Println("traceme installed as launchd service")
	return nil
}

func Uninstall() error {
	path := plistPath()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("traceme service not installed")
		return nil
	}

	cmd := exec.Command("launchctl", "unload", path)
	_ = cmd.Run()

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("remove plist: %w", err)
	}

	fmt.Println("traceme service uninstalled")
	return nil
}

func Restart() error {
	path := plistPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("service not installed (run 'traceme install' first)")
	}

	exec.Command("launchctl", "unload", path).Run()
	return Install()
}
