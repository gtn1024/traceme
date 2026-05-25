//go:build windows

package autostart

import (
	"fmt"
	"os"
	"os/exec"
)

const taskName = "traceme"

func binaryPath() (string, error) {
	if bin, err := exec.LookPath("traceme.exe"); err == nil {
		return bin, nil
	}
	if exe, err := os.Executable(); err == nil {
		return exe, nil
	}
	return "", fmt.Errorf("cannot determine traceme binary path")
}

func Install() error {
	bin, err := binaryPath()
	if err != nil {
		return err
	}

	cmd := exec.Command("schtasks", "/Create",
		"/SC", "ONLOGON",
		"/TN", taskName,
		"/TR", fmt.Sprintf("\"%s\" run", bin),
		"/RL", "LIMITED",
		"/F",
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks create: %w\n%s", err, string(output))
	}

	fmt.Println("traceme installed as scheduled task")
	return nil
}

func Uninstall() error {
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	if err := cmd.Run(); err != nil {
		fmt.Println("traceme service not installed")
		return nil
	}

	cmd = exec.Command("schtasks", "/Delete", "/TN", taskName, "/F")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("schtasks delete: %w\n%s", err, string(output))
	}

	fmt.Println("traceme service uninstalled")
	return nil
}

func Restart() error {
	cmd := exec.Command("schtasks", "/Query", "/TN", taskName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("service not installed (run 'traceme install' first)")
	}

	if err := Uninstall(); err != nil {
		return err
	}
	return Install()
}
