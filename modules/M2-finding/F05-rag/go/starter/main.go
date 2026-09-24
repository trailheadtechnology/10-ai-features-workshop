// Demo starting point: one chat target, one question, no context.
// This is the "plain chatbot" from step 1 of the demo script. It answers
// confidently. It is also wrong: the model puts Sperry Chalet in California
// and guesses at fire rules the park wrote down years ago.
// Run: go run . ["your question"]
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

func main() {
	question := "Can I have a campfire at Sperry Chalet in September?"
	if len(os.Args) > 1 {
		question = strings.Join(os.Args[1:], " ")
	}

	fmt.Printf("Q: %s\n\n", question)
	answer, err := chat(ollamaTarget("llama3.2"), question)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(answer)
}

// target is one chat endpoint: its URL, model name, and any headers it needs.
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
