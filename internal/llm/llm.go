package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gtn1024/traceme/internal/config"
)

const systemPrompt = `You are a local activity logger.

You will receive a screenshot of the user's computer screen.

Describe what the user appears to be doing.

Rules:
- Output JSON only.
- Be concise.
- Do not give advice.
- Do not include long code snippets.
- Do not include secrets, tokens, private messages, or sensitive content.
- If the screen contains private content (chat, email, passwords, finance), minimize details in summary.
- Summarize what the user appears to be doing, not what they should do.

Schema:
{
  "activity": "coding | reading | debugging | writing | watching_lecture | browsing | chatting | unknown",
  "app": string | null,
  "project": string | null,
  "summary": string,
  "topics": string[]
}`

const userMessage = "Analyze this screenshot and produce one JSON object following the schema."

type Activity struct {
	Activity string   `json:"activity"`
	App      string   `json:"app"`
	Project  string   `json:"project"`
	Summary  string   `json:"summary"`
	Topics   []string `json:"topics"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"`
}

type textContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type imageContent struct {
	Type     string    `json:"type"`
	ImageURL imageURL  `json:"image_url"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func Analyze(cfg *config.Config, imageBase64 string) (*Activity, int64, error) {
	reqBody := chatRequest{
		Model: cfg.Model.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: []interface{}{
				textContent{Type: "text", Text: userMessage},
				imageContent{Type: "image_url", ImageURL: imageURL{URL: imageBase64}},
			}},
		},
		Temperature: 0.1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal request: %w", err)
	}

	url := cfg.Model.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.Model.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.Model.APIKey)
	}

	client := &http.Client{Timeout: time.Duration(cfg.Model.TimeoutSeconds) * time.Second}
	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return nil, latency, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, latency, fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, latency, fmt.Errorf("decode response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, latency, fmt.Errorf("empty response from model")
	}

	content := chatResp.Choices[0].Message.Content
	var activity Activity
	if err := json.Unmarshal([]byte(extractJSON(content)), &activity); err != nil {
		return nil, latency, fmt.Errorf("parse activity JSON: %w", err)
	}

	return &activity, latency, nil
}

func extractJSON(s string) string {
	start := -1
	end := -1
	for i := 0; i < len(s); i++ {
		if s[i] == '{' && start == -1 {
			start = i
		}
		if s[i] == '}' {
			end = i
		}
	}
	if start >= 0 && end > start {
		return s[start : end+1]
	}
	return s
}
