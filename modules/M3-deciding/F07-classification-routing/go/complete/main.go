// Classifies every inquiry in ../../data/inquiries-slice.jsonl and scores the
// result against ../../data/reference-labels.json.
// Run: go run .
//
// The category comes back through a JSON-schema-constrained response format,
// so the model can only return one of the seven labels the routing table
// already knows how to handle.
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

type inquiry struct {
	ID       string `json:"id"`
	Channel  string `json:"channel"`
	Received string `json:"received"`
	Text     string `json:"text"`
}

type referenceLabels struct {
	Routing map[string]string `json:"routing"`
	Labels  map[string]string `json:"labels"`
}

var categories = []string{"permit", "conditions", "complaint", "lost-and-found", "emergency", "general", "unsure"}

var triageSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"category": map[string]any{
			"type": "string",
			"enum": categories,
		},
	},
	"required":             []string{"category"},
	"additionalProperties": false,
}

type result struct {
	inquiry  inquiry
	category string
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	inquiries, err := readInquiries(filepath.Join(data, "inquiries-slice.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	reference, err := loadReferenceLabels(filepath.Join(data, "reference-labels.json"))
	if err != nil {
		log.Fatal(err)
	}

	var results []result
	for _, inq := range inquiries {
		category, err := classify(inq.Text)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result{inquiry: inq, category: category})
		fmt.Print(".")
	}
	fmt.Println()
	fmt.Println()

	// Emergencies print before the routing table and are sorted to the top of it.
	// A person scanning this output under time pressure must not have to read past
	// the first screen to find one.
	var emergencies []result
	for _, r := range results {
		if r.category == "emergency" {
			emergencies = append(emergencies, r)
		}
	}
	if len(emergencies) > 0 {
		fmt.Println("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!")
		for _, e := range emergencies {
			fmt.Printf("!!! %s  %s\n", e.inquiry.ID, clip(e.inquiry.Text, 70))
		}
		fmt.Println()
	}

	// Stable partition: emergencies first, original order preserved within each group.
	sorted := partitionEmergenciesFirst(results)

	fmt.Printf("%-10s %-15s routed to\n", "id", "category")
	fmt.Println(strings.Repeat("-", 62))
	for _, r := range sorted {
		fmt.Printf("%-10s %-15s %s\n", r.inquiry.ID, r.category, reference.Routing[r.category])
	}

	// Two scores, and they are not equally important. Overall accuracy is the
	// headline number; recall on the emergency class is the one that decides
	// whether this taxonomy is safe to ship. A missed emergency is a person waiting
	// in a queue nobody is watching, and no amount of accuracy elsewhere offsets it.
	correct := 0
	for _, r := range results {
		if r.category == reference.Labels[r.inquiry.ID] {
			correct++
		}
	}
	var emergencyIDs []string
	for id, label := range reference.Labels {
		if label == "emergency" {
			emergencyIDs = append(emergencyIDs, id)
		}
	}
	caught := 0
	for _, e := range emergencies {
		for _, id := range emergencyIDs {
			if e.inquiry.ID == id {
				caught++
				break
			}
		}
	}

	fmt.Println()
	fmt.Printf("Accuracy vs reference labels: %d/%d\n", correct, len(results))
	verdict := "(MISSED ONE; this fails, whatever the accuracy says)"
	if caught == len(emergencyIDs) {
		verdict = "(all caught; the metric that matters)"
	}
	fmt.Printf("Emergency recall: %d/%d %s\n", caught, len(emergencyIDs), verdict)
	for _, r := range results {
		if r.category != reference.Labels[r.inquiry.ID] {
			fmt.Printf("  miss: %s got %s, reference says %s\n", r.inquiry.ID, r.category, reference.Labels[r.inquiry.ID])
		}
	}
}

// partitionEmergenciesFirst keeps emergencies at the top, preserving the
// original classification order within each group (a stable partition).
func partitionEmergenciesFirst(results []result) []result {
	var out []result
	for _, r := range results {
		if r.category == "emergency" {
			out = append(out, r)
		}
	}
	for _, r := range results {
		if r.category != "emergency" {
			out = append(out, r)
		}
	}
	return out
}

func clip(text string, max int) string {
	if len(text) <= max {
		return text
	}
	return text[:max] + "..."
}

func readInquiries(path string) ([]inquiry, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inquiries []inquiry
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var i inquiry
		if err := json.Unmarshal([]byte(line), &i); err != nil {
			return nil, err
		}
		inquiries = append(inquiries, i)
	}
	return inquiries, nil
}

func loadReferenceLabels(path string) (referenceLabels, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return referenceLabels{}, err
	}
	var ref referenceLabels
	if err := json.Unmarshal(raw, &ref); err != nil {
		return referenceLabels{}, err
	}
	return ref, nil
}

// These descriptions are the taxonomy, and editing them changes behavior more
// than any code below. Two rules constrain any rewrite. Emergency wins over
// every other category, including messages that also mention a permit or a lost
// item, so the ordering paragraph at the end must stay. And unsure has to stay
// narrow: it means two queues must both act on one message, not that the model
// found the message hard.
func prompt(text string) string {
	return fmt.Sprintf(`You are the triage system for the Trailhead Guides shared inbox.
Classify the visitor message into exactly one category.

- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, including billing problems and missing confirmations
  for a permit application.
- conditions: asking whether a trail, road, or area is open, safe, or
  passable right now: snow, water levels, washouts, wildlife activity,
  closures.
- complaint: unhappy about a park facility, service, or staff member
  and wants it acknowledged or fixed.
- lost-and-found: reporting a lost or found physical item.
- emergency: a person may be hurt, missing, or in danger right now and
  needs immediate human attention.
- general: anything else: park rules, fees, trip planning, questions
  that fit none of the above.
- unsure: two different queues both have to act before this message can
  be resolved, so no single queue owns it. The case that qualifies: the
  sender asks about trail conditions AND asks someone to change, refund,
  or cancel a booking. Trail info cannot issue a refund, and the permits
  office does not decide whether a trail is passable, so a human reads
  this queue and splits the work. Also use unsure when the message fits
  none of the categories above.

Decide in this order. First, if anyone might be hurt, missing, or in
danger, answer emergency and stop; never answer unsure for those, even
when the message also mentions permits, conditions, or a lost item.
Second, if one queue can resolve the whole message on its own, answer
that queue; a booking or reservation problem with nothing else attached
is permit, not unsure. Third, only if two queues must both act, answer
unsure. Unsure is not a catch-all for anything hard.

Message:
%s`, text)
}

type jsonSchemaFormat struct {
	Name   string `json:"name"`
	Strict bool   `json:"strict"`
	Schema any    `json:"schema"`
}

type responseFormat struct {
	Type       string           `json:"type"`
	JSONSchema jsonSchemaFormat `json:"json_schema"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	Temperature    *float64        `json:"temperature,omitempty"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type triageResult struct {
	Category string `json:"category"`
}

// classify sends the taxonomy prompt with a JSON-schema response format, so
// the model can only return one of the seven known category strings.
// Anything above temperature 0 makes the same message land in different
// queues on different runs, which makes a scored comparison against fixed
// reference labels meaningless.
func classify(text string) (string, error) {
	temperature := 0.0
	body, err := json.Marshal(chatRequest{
		Model:       model,
		Messages:    []chatMessage{{Role: "user", Content: prompt(text)}},
		Temperature: &temperature,
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchemaFormat{
				Name:   "triage_result",
				Strict: true,
				Schema: triageSchema,
			},
		},
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

	var tr triageResult
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &tr); err != nil {
		return "", fmt.Errorf("model did not return schema-conformant JSON: %w", err)
	}
	return tr.Category, nil
}
