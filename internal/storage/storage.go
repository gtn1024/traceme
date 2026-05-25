package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gtn1024/traceme/internal/llm"
)

type Record struct {
	TS        string   `json:"ts"`
	Activity  string   `json:"activity"`
	App       string   `json:"app"`
	Project   string   `json:"project"`
	Summary   string   `json:"summary"`
	Topics    []string `json:"topics"`
	Model     string   `json:"model"`
	LatencyMs int64    `json:"latency_ms"`
}

func NewRecord(activity *llm.Activity, model string, latencyMs int64) Record {
	return Record{
		TS:        time.Now().Format(time.RFC3339),
		Activity:  activity.Activity,
		App:       activity.App,
		Project:   activity.Project,
		Summary:   activity.Summary,
		Topics:    activity.Topics,
		Model:     model,
		LatencyMs: latencyMs,
	}
}

func (r Record) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func Append(logsDir string, record Record) error {
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("ensure logs dir: %w", err)
	}

	filename := time.Now().Format("2006-01-02") + ".jsonl"
	path := filepath.Join(logsDir, filename)

	line, err := record.Marshal()
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	_, err = f.Write(append(line, '\n'))
	return err
}
