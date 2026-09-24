// Finished demo, as shown on stage:
//
//	go run .                                the naive prompt (the book report)
//	go run . --briefing                     3-bullet hiker briefing
//	go run . --headline                     one-line trail status for a card UI
//	go run . --briefing --audience ranger
//	Any non-flag argument is a path to a different trip report.
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

	reportPath := filepath.Join(data, "tr-0004.md")
	mode := "naive"
	audience := "hiker"

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--briefing":
			mode = "briefing"
		case "--headline":
			mode = "headline"
		case "--audience":
			i++
			audience = args[i]
		default:
			reportPath = args[i]
		}
	}

	raw, err := os.ReadFile(reportPath)
	if err != nil {
		log.Fatal(err)
	}
	report := stripFrontMatter(string(raw))

	audienceFocus := "a hiker planning to hike this trail within the next week"
	if audience == "ranger" {
		audienceFocus = "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
	}

	var prompt string
	switch mode {
	// Deliberately the weak prompt. It is kept so the two prompts can be run
	// back to back against the same report; nothing about it should be fixed.
	case "naive":
		prompt = fmt.Sprintf("Summarize this trip report.\n\n%s", report)

	// The last two lines are load-bearing, not politeness. The bullets require a
	// hazards slot and require any hazard to come first, so on a report with no
	// hazard the model will promote the nearest noun (a bear, a creek, the word
	// "avalanche" in the trail name) into a closure. Giving it a legal way to
	// report nothing is what stops that. Measurements in ../../expected-output.md.
	case "briefing":
		prompt = fmt.Sprintf(`You are helping %s.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.

%s`, audienceFocus, report)

	// Same client, same report, same call. Only the instruction changes to fit a
	// different UI slot, so no new infrastructure is needed for a new surface.
	case "headline":
		prompt = fmt.Sprintf(`From the trip report below, write ONE line of at most 12 words,
suitable for a status badge on a trail card in an app.
Lead with the most important condition or closure. No preamble.

%s`, report)

	default:
		log.Fatalf("unknown mode %s", mode)
	}

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
