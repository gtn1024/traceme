package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gtn1024/traceme/internal/config"
	"github.com/gtn1024/traceme/internal/launchd"
	"github.com/gtn1024/traceme/internal/llm"
	"github.com/gtn1024/traceme/internal/screenshot"
	"github.com/gtn1024/traceme/internal/storage"
	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "traceme",
	Short: "Local activity logger",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize config and log directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.DefaultConfig()
		if err := os.MkdirAll(cfg.LogsDir(), 0755); err != nil {
			return fmt.Errorf("create logs dir: %w", err)
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("Initialized at %s\n", cfg.Storage.Root)
		return nil
	},
}

var captureCmd = &cobra.Command{
	Use:   "capture",
	Short: "Take one screenshot and log activity",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("load config (run 'traceme init' first): %w", err)
		}
		return capture(cfg)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run continuously at configured interval",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("load config (run 'traceme init' first): %w", err)
		}

		interval := time.Duration(cfg.IntervalSeconds) * time.Second
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		fmt.Fprintf(os.Stderr, "traceme: running every %s\n", interval)
		for {
			if err := capture(cfg); err != nil {
				fmt.Fprintf(os.Stderr, "traceme: %v\n", err)
			}
			select {
			case <-sigCh:
				fmt.Fprintf(os.Stderr, "traceme: stopped\n")
				return nil
			case <-time.After(interval):
			}
		}
	},
}

func capture(cfg *config.Config) error {
	path, err := screenshot.Capture()
	if err != nil {
		return fmt.Errorf("screenshot: %w", err)
	}
	defer screenshot.Cleanup(path)

	b64, err := screenshot.ReadBase64(path)
	if err != nil {
		return fmt.Errorf("read screenshot: %w", err)
	}

	activity, latency, err := llm.Analyze(cfg, b64)
	if err != nil {
		return fmt.Errorf("llm: %w", err)
	}

	record := storage.NewRecord(activity, cfg.Model.Model, latency)
	if err := storage.Append(cfg.LogsDir(), record); err != nil {
		return fmt.Errorf("write log: %w", err)
	}

	fmt.Printf("%s  %s  %s\n", record.Activity, record.App, record.Project)
	fmt.Printf("  %s\n", record.Summary)
	return nil
}

var todayCmd = &cobra.Command{
	Use:   "today",
	Short: "Show today's activity log",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("load config (run 'traceme init' first): %w", err)
		}

		filename := time.Now().Format("2006-01-02") + ".jsonl"
		path := cfg.LogsDir() + "/" + filename

		f, err := os.Open(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No activity recorded today.")
				return nil
			}
			return err
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			var r storage.Record
			if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
				continue
			}
			t, _ := time.Parse(time.RFC3339, r.TS)
			ts := t.Format("15:04")
			app := r.App
			if app == "" {
				app = "-"
			}
			proj := r.Project
			if proj == "" {
				proj = "-"
			}
			fmt.Printf("%s  %s  %s  %s\n", ts, r.Activity, app, proj)
			fmt.Printf("  %s\n", r.Summary)
		}
		return scanner.Err()
	},
}

var dailyPromptCmd = &cobra.Command{
	Use:   "daily-prompt",
	Short: "Output a daily summary prompt for LLM",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadDefault()
		if err != nil {
			return fmt.Errorf("load config (run 'traceme init' first): %w", err)
		}

		filename := time.Now().Format("2006-01-02") + ".jsonl"
		path := cfg.LogsDir() + "/" + filename

		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Println("No activity recorded today.")
				return nil
			}
			return err
		}

		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		var activities []string
		for _, line := range lines {
			var r storage.Record
			if err := json.Unmarshal([]byte(line), &r); err != nil {
				continue
			}
			t, _ := time.Parse(time.RFC3339, r.TS)
			parts := []string{t.Format("15:04"), r.Activity}
			if r.App != "" {
				parts = append(parts, r.App)
			}
			if r.Project != "" {
				parts = append(parts, r.Project)
			}
			activities = append(activities, fmt.Sprintf("[%s]: %s", strings.Join(parts, " "), r.Summary))
		}

		fmt.Printf("Below is my activity log for today (%s). Please generate a daily summary:\n\n%s\n", time.Now().Format("2006-01-02"), strings.Join(activities, "\n"))
		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install traceme as a launchd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchd.Install()
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the launchd service",
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchd.Uninstall()
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the launchd service (use after updating)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return launchd.Restart()
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(todayCmd)
	rootCmd.AddCommand(dailyPromptCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("traceme " + version)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
