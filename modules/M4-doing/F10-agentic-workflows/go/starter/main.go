// The trip request goes to a plain chat completion, with no tools and no loop.
// Nothing here can reach the trail catalog, the weather feed, or the condition
// reports, so the itinerary that comes back is fluent, generic, and books
// nothing. That gap is the point of this program; the fix lives in ../complete.
//
// Run: go run . [your own trip request]
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type target struct {
	url     string
	model   string
	headers map[string]string
}

func ollamaTarget(model string) target {
	return target{
		url:     "http://localhost:11434/v1/chat/completions",
		model:   model,
		headers: map[string]string{"Authorization": "Bearer ollama"}, // placeholder: Ollama needs no key
	}
}

func azureTarget(endpoint, deployment, apiVersion, key string) target {
	return target{
		url:     fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", endpoint, deployment, apiVersion),
		model:   deployment,
		headers: map[string]string{"api-key": key},
	}
}

func createChatTarget() target {
	endpoint := "https://trailhead-ai-workshop.openai.azure.com"
	key := "<KEY FROM INSTRUCTOR>" // paste the room key between the quotes
	deployment := "gpt-5.5"
	if !strings.HasPrefix(key, "<") {
		return azureTarget(endpoint, deployment, "2024-10-21", key)
	}
	fmt.Println("[note] no room key pasted in; falling back to Ollama llama3.2.")
	return ollamaTarget("llama3.2")
}

func main() {
	request := strings.Join(os.Args[1:], " ")
	if request == "" {
		request = "Plan me a 3-day trip in Glacier National Park for September 14-16."
	}

	fmt.Printf("Request: %s\n", request)
	fmt.Println(strings.Repeat("-", 60))

	t := createChatTarget()
	prompt := fmt.Sprintf("You are the trip planner for Trailhead Guides, a hiking app.\n\n%s", request)
	reply, err := chat(t, prompt)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(reply)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("Note: zero tool calls were made. No weather was checked, no")
	fmt.Println("conditions were read, nothing was booked. Fluent and useless.")
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

func chat(t target, prompt string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:    t.model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

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
		return "", fmt.Errorf("%s returned %s: %s", t.url, resp.Status, respBody)
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
