// Starting point, and deliberately unsafe. Every draft goes straight out with no
// human between the model and the visitor, and nothing is logged, so there is no
// record of what was sent or any way to find out later.
// Run: go run . [path-to-inquiries.jsonl]
//
// Nothing here treats an emergency differently. The system prompt asks the model
// to escalate instead of drafting, and the model is free to ignore that and
// often does, which is why complete/ moves the decision out of the prompt and
// into a policy check in code. Do not use this shape on a real inbox.
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

const systemPrompt = `You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides.
A human ranger reviews your draft before anything is sent, so write it ready to send:
friendly, plain, professional, at most two short paragraphs, signed
'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure,
state the rule and cite the source document number and section (for example
GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided;
if the excerpt does not answer the question, say a ranger will follow up with
specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
Exception: if the visitor's message reports an emergency, an injury, a possible fire,
or a missing or overdue person, do not draft a reply at all. Output exactly one line
beginning with ESCALATE: followed by a one-line reason, so the message goes straight
to dispatch.`

type inquiry struct {
	ID       string `json:"id"`
	Channel  string `json:"channel"`
	Received string `json:"received"`
	Category string `json:"category"`
	Doc      string `json:"doc"`
	Text     string `json:"text"`
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	here := filepath.Dir(thisFile)
	defaultPath := filepath.Join(here, "..", "..", "data", "inquiries.jsonl")

	inquiriesPath := defaultPath
	if len(os.Args) > 1 {
		inquiriesPath = os.Args[1]
	}
	absPath, err := filepath.Abs(inquiriesPath)
	if err != nil {
		log.Fatal(err)
	}
	dataDir := filepath.Dir(absPath)

	inquiries, err := readInquiries(inquiriesPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, inq := range inquiries {
		snippet := loadSnippet(dataDir, inq.Doc)
		userMessage := fmt.Sprintf(`Reference excerpt:
%s

Visitor message (%s, received %s):
%s

Draft the reply.`, snippet, inq.Channel, inq.Received, inq.Text)

		draft, err := chat([]chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		})
		if err != nil {
			log.Fatal(err)
		}

		// No review step and no audit record. In a real deployment this line is the
		// send call, and by the time anyone reads the output the mail has gone.
		fmt.Printf("=== SENT to visitor · %s (%s) ===\n", inq.ID, inq.Category)
		fmt.Println(strings.TrimSpace(draft))
		fmt.Println()
	}

	fmt.Println("All replies sent. Nobody read them. Nothing was logged.")
}

func loadSnippet(dataDir, doc string) string {
	if doc == "" {
		return "(none on file for this message)"
	}
	raw, err := os.ReadFile(filepath.Join(dataDir, "snippets", doc))
	if err != nil {
		return "(none on file for this message)"
	}
	return strings.TrimSpace(string(raw))
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

func chat(messages []chatMessage) (string, error) {
	body, err := json.Marshal(chatRequest{Model: "llama3.2", Messages: messages})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost:11434/v1/chat/completions", bytes.NewReader(body))
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
