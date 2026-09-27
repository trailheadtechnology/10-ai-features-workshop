// Demo starting point: one chat target, one classify function, one review.
// Run: go run . [review-id]
// Ids come from ../../data/easy.jsonl and ../../data/hard.jsonl. The default is
// gr-0007, a hard-set review whose sarcasm ("five-star experience, truly") points
// the opposite way from its two-star rating.
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

type review struct {
	ID       string `json:"id"`
	Product  string `json:"product"`
	Rating   int    `json:"rating"`
	Reviewer string `json:"reviewer"`
	Text     string `json:"text"`
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	wanted := "gr-0007"
	if len(os.Args) > 1 {
		wanted = os.Args[1]
	}

	reviews, err := readReviews(filepath.Join(data, "easy.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	hard, err := readReviews(filepath.Join(data, "hard.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	reviews = append(reviews, hard...)

	var found *review
	for i := range reviews {
		if reviews[i].ID == wanted {
			found = &reviews[i]
			break
		}
	}
	if found == nil {
		log.Fatalf("no review with id %s", wanted)
	}

	fmt.Printf("%s (%d stars), reviewed by %s\n", found.Product, found.Rating, found.Reviewer)
	fmt.Println(found.Text)
	fmt.Println()

	label, err := classify(ollamaTarget("phi3"), found.Text)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("phi3 says: %s\n", label)
}

func readReviews(path string) ([]review, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reviews []review
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var r review
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, nil
}

// target is one chat endpoint: its URL, model name, and any headers it needs.
// classify below works with any target, which is what makes the model swappable.
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

// The whole feature is this function. The prompt carries it; the model is
// swappable because everything upstream only sees a target and some text.
func classify(t target, text string) (string, error) {
	// Keep this prompt byte-identical to the one in F03-lab.md, line breaks
	// included. Reflowing these four lines into one costs phi3 measured accuracy
	// on both sets, so a comparison run against a reflowed prompt is not
	// comparing models. See ../../expected-output.md.
	prompt := fmt.Sprintf(`Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: %s`, text)

	raw, err := chat(t, prompt)
	if err != nil {
		return "", err
	}
	raw = strings.ToLower(raw)

	// Small models sometimes wrap the label in a sentence; keep the first label mentioned.
	labels := []string{"positive", "negative", "mixed"}
	bestAt, bestLabel := -1, ""
	for _, l := range labels {
		if at := strings.Index(raw, l); at >= 0 && (bestAt == -1 || at < bestAt) {
			bestAt, bestLabel = at, l
		}
	}
	if bestLabel != "" {
		return bestLabel, nil
	}
	return strings.TrimSpace(raw), nil
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func chat(t target, prompt string) (string, error) {
	temperature := 0.0
	body, err := json.Marshal(chatRequest{
		Model:       t.model,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: &temperature,
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
