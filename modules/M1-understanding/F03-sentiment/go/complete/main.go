// Finished demo, as shown on stage:
//
//	go run .            both sets, both models, table + accuracy + disagreements
//	go run . --easy     easy set only (demo steps 3 and 4)
//	go run . --hard     hard set only (demo step 5)
//
// The big model is Azure OpenAI gpt-4.1 when the room key is pasted in below.
// When it isn't, llama3.2 on Ollama stands in so the whole comparison runs
// offline. Either way, the swap is the one block building `big` below;
// nothing downstream changes.
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

type refLabel struct {
	Set   string `json:"set"`
	Label string `json:"label"`
}

type result struct {
	Review    review
	Set       string
	Reference string
	Small     string
	Big       string
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	args := os.Args[1:]
	sets := []string{"easy", "hard"}
	for _, a := range args {
		if a == "--easy" {
			sets = []string{"easy"}
		} else if a == "--hard" {
			sets = []string{"hard"}
		}
	}

	// Model 1: the small local model. Free, private, 2GB.
	small := ollamaTarget("phi3")

	// Model 2: the big model, or its local stand-in.
	endpoint := "https://trailhead-ai-workshop.openai.azure.com"
	key := "<KEY FROM INSTRUCTOR>" // paste the room key between the quotes
	deployment := "gpt-4.1"

	var big target
	var bigName string
	if !strings.HasPrefix(key, "<") {
		big = azureTarget(endpoint, deployment, "2024-10-21", key)
		bigName = fmt.Sprintf("azure:%s", deployment)
	} else {
		fmt.Println("no room key pasted in; using llama3.2 on Ollama as the big-model stand-in.")
		fmt.Println()
		big = ollamaTarget("llama3.2")
		bigName = "llama3.2"
	}

	labels, err := loadLabels(filepath.Join(data, "reference-labels.json"))
	if err != nil {
		log.Fatal(err)
	}

	var results []result
	for _, set := range sets {
		fmt.Printf("── %s set ──\n", set)
		fmt.Printf("%-9s %-10s %-10s %-10s\n", "id", "reference", "phi3", bigName)

		reviews, err := readReviews(filepath.Join(data, set+".jsonl"))
		if err != nil {
			log.Fatal(err)
		}
		for _, rv := range reviews {
			reference := labels[rv.ID].Label
			s, err := classify(small, rv.Text)
			if err != nil {
				log.Fatal(err)
			}
			b, err := classify(big, rv.Text)
			if err != nil {
				log.Fatal(err)
			}
			results = append(results, result{Review: rv, Set: set, Reference: reference, Small: s, Big: b})
			flag := ""
			if s != b {
				flag = "  <- disagree"
			}
			fmt.Printf("%-9s %-10s %-10s %-10s%s\n", rv.ID, reference, s, b, flag)
		}
		fmt.Println()
	}

	fmt.Println("── accuracy vs. reference labels ──")
	for _, set := range sets {
		var batch []result
		for _, r := range results {
			if r.Set == set {
				batch = append(batch, r)
			}
		}
		smallOk, bigOk := 0, 0
		for _, r := range batch {
			if r.Small == r.Reference {
				smallOk++
			}
			if r.Big == r.Reference {
				bigOk++
			}
		}
		fmt.Printf("%-5s  phi3 %d/%d   %s %d/%d\n", set, smallOk, len(batch), bigName, bigOk, len(batch))
	}
	fmt.Println()

	var disagreements []result
	for _, r := range results {
		if r.Small != r.Big {
			disagreements = append(disagreements, r)
		}
	}
	fmt.Printf("── disagreements (%d of %d) ──\n", len(disagreements), len(results))
	for _, d := range disagreements {
		verdict := "both wrong"
		if d.Big == d.Reference {
			verdict = fmt.Sprintf("%s right", bigName)
		} else if d.Small == d.Reference {
			verdict = "phi3 right"
		}
		fmt.Printf("%s [%s] ref=%s phi3=%s %s=%s  (%s)\n", d.Review.ID, d.Set, d.Reference, d.Small, bigName, d.Big, verdict)
		fmt.Printf("  %q\n", truncate(d.Review.Text, 100))
	}
	if len(disagreements) == 0 {
		fmt.Println("(none this run)")
	}
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return strings.TrimRight(string(r[:max]), " \t\n") + "..."
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

func loadLabels(path string) (map[string]refLabel, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var labels map[string]refLabel
	if err := json.Unmarshal(raw, &labels); err != nil {
		return nil, err
	}
	return labels, nil
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

func azureTarget(endpoint, deployment, apiVersion, key string) target {
	return target{
		url:     fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s", endpoint, deployment, apiVersion),
		model:   deployment,
		headers: map[string]string{"api-key": key},
	}
}

// Same function as the starter: one prompt, one word back, any target.
func classify(t target, text string) (string, error) {
	// Both models get this exact prompt, and it is byte-identical to the one in
	// F03-lab.md, line breaks included. Reflowing these
	// four lines into one costs phi3 measured accuracy on both sets while leaving
	// llama3.2 unchanged, so varying the prompt shape and the model in the same
	// run measures nothing. See ../../expected-output.md.
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
