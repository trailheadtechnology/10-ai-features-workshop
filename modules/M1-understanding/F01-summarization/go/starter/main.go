// Demo starting point: one HTTP call, one naive prompt.
// Run: go run . [path-to-trip-report.md]
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	ollamaURL = "http://localhost:11434/v1/chat/completions"
	model     = "llama3.2"
)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	reportPath := filepath.Join(data, "tr-0001.md")
	if len(os.Args) > 1 {
		reportPath = os.Args[1]
	}

	raw, err := os.ReadFile(reportPath)
	if err != nil {
		log.Fatal(err)
	}
	report := stripFrontMatter(string(raw))

	prompt := fmt.Sprintf("Summarize this trip report.\n\n%s", report)

	reply, err := chat(prompt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(reply)
}

func stripFrontMatter(markdown string) string {
	parts := strings.SplitN(markdown, "---", 3)
	if len(parts) == 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(markdown)
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func chat(prompt string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, ollamaURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ollama") // placeholder: Ollama needs no key

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned %s: %s", resp.Status, respBody)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return parsed.Choices[0].Message.Content, nil
}
