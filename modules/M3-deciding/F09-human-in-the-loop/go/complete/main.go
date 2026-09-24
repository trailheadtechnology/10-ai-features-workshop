// Review queue: the model drafts, a human decides, and every decision is logged
// to decisions.jsonl.
//
//	go run .                              review the queue: [a]pprove / [e]dit / [r]eject / [s]kip
//	go run . --policy                     print the routing policy table and exit
//	go run . --auto-approve-dry-run       non-interactive run for testing and CI
//	go run . ../../data/inquiries.jsonl   any queue file works
//
// SAFETY INVARIANT, load-bearing, do not weaken:
// emergencies never reach the model. The policy table below routes them to
// human-only and the loop skips the API call entirely, in code, before any
// request is built. The system prompt also tells the model to escalate instead
// of drafting, but that instruction is a request and this lane is a guarantee.
// A model that ignores the instruction and writes a warm, fluent, confident
// reply to someone reporting an overdue hiker is not a hypothetical; it is the
// documented behavior of the model this demo ships with. Anyone editing this
// file must keep the emergency path free of model calls. Adding a "just draft
// it and let the reviewer catch it" shortcut here puts a reassuring lie in front
// of a person who needed a dispatcher.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
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

type decisionRecord struct {
	At           string  `json:"at"`
	InquiryID    string  `json:"inquiryId"`
	Category     string  `json:"category"`
	Lane         string  `json:"lane"`
	Decision     string  `json:"decision"`
	Reviewer     string  `json:"reviewer"`
	Draft        *string `json:"draft"`
	Final        *string `json:"final"`
	EditDistance int     `json:"editDistance"`
}

// The lane is chosen by what a wrong answer costs, not by how good the model is
// at the category. Everything reversible can be drafted; emergency is
// irreversible and stays human-only. laneFor defaults an unknown category to
// human-only, so a category added upstream fails closed rather than quietly
// acquiring a draft lane.
var policy = []struct{ category, lane string }{
	{"trail-condition", "draft-for-approval"},
	{"permit", "draft-for-approval"},
	{"complaint", "draft-for-approval"},
	{"general", "draft-for-approval"},
	{"lost-and-found", "draft-for-approval"},
	{"emergency", "human-only"},
}

func laneFor(category string) string {
	for _, p := range policy {
		if p.category == category {
			return p.lane
		}
	}
	return "human-only"
}

func printPolicy() {
	fmt.Println("Routing policy (error cost decides the lane):")
	for _, p := range policy {
		fmt.Printf("  %-16s %s\n", p.category, p.lane)
	}
	fmt.Println()
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	here := filepath.Dir(thisFile)
	defaultInquiriesPath := filepath.Join(here, "..", "..", "data", "inquiries.jsonl")

	inquiriesPath := defaultInquiriesPath
	outboxDir := "outbox"
	decisionsPath := "decisions.jsonl"
	autoApprove := false

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--auto-approve-dry-run":
			autoApprove = true
		case "--outbox":
			i++
			outboxDir = args[i]
		case "--decisions":
			i++
			decisionsPath = args[i]
		case "--policy":
			printPolicy()
			return
		default:
			inquiriesPath = args[i]
		}
	}

	absPath, err := filepath.Abs(inquiriesPath)
	if err != nil {
		log.Fatal(err)
	}
	dataDir := filepath.Dir(absPath)
	if err := os.MkdirAll(outboxDir, 0o755); err != nil {
		log.Fatal(err)
	}

	reviewer := "auto-approve-dry-run"
	if !autoApprove {
		reviewer = currentUsername()
	}
	counts := map[string]int{}

	printPolicy()
	if autoApprove {
		fmt.Println("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.")
		fmt.Println()
	}

	inquiries, err := readInquiries(inquiriesPath)
	if err != nil {
		log.Fatal(err)
	}

	stdin := bufio.NewScanner(os.Stdin)

	for _, inq := range inquiries {
		lane := laneFor(inq.Category)

		fmt.Println(strings.Repeat("-", 72))
		fmt.Printf("%s  ·  %s  ·  %s  ·  lane: %s\n", inq.ID, inq.Category, inq.Channel, lane)
		fmt.Println(strings.Repeat("-", 72))
		fmt.Println(indent(inq.Text))
		fmt.Println()

		// THE GATE. This must stay above the API call, and the API call must stay
		// below it. A human-only message is escalated and logged without a single
		// token being spent on it, so there is no draft to leak, no reviewer fatigue
		// to survive, and no sampling luck involved. The ESCALATE handling further
		// down is a backstop for emergencies that arrive miscategorized; it is never
		// the control, because it runs after the model has already had its say.
		if lane == "human-only" {
			fmt.Println("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.")
			fmt.Println()
			logDecision(decisionsPath, decisionRecord{
				At: nowUTC(), InquiryID: inq.ID, Category: inq.Category, Lane: lane,
				Decision: "escalated", Reviewer: reviewer, Draft: nil, Final: nil, EditDistance: 0,
			})
			counts["escalated"]++
			continue
		}

		snippet := loadSnippet(dataDir, inq.Doc)
		userMessage := fmt.Sprintf(`Reference excerpt:
%s

Visitor message (%s, received %s):
%s

Draft the reply.`, snippet, inq.Channel, inq.Received, inq.Text)

		fmt.Print("  drafting...")
		reply, err := chat([]chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		})
		if err != nil {
			log.Fatal(err)
		}
		draft := strings.TrimSpace(reply)
		fmt.Print("\r  draft:      \n\n")
		fmt.Println(indent(draft))
		fmt.Println()

		// Second layer, for an emergency that reached here under the wrong category.
		// An ESCALATE prefix is a hard stop: the draft is logged for the audit trail
		// but never offered for approval, because a reviewer presented with a
		// sendable-looking reply may send it. Do not soften this into a warning.
		if strings.HasPrefix(strings.ToUpper(draft), "ESCALATE") {
			fmt.Println("  Model asked to escalate. Draft discarded, routing to a human.")
			fmt.Println()
			d := draft
			logDecision(decisionsPath, decisionRecord{
				At: nowUTC(), InquiryID: inq.ID, Category: inq.Category, Lane: lane,
				Decision: "escalated", Reviewer: reviewer, Draft: &d, Final: nil, EditDistance: 0,
			})
			counts["escalated"]++
			continue
		}

		var decision string
		var final *string

		if autoApprove {
			decision = "approved"
			d := draft
			final = &d
			fmt.Println("  [auto] approved")
			fmt.Println()
		} else {
			fmt.Print("  [a]pprove  [e]dit  [r]eject  [s]kip > ")
			key := "s"
			if stdin.Scan() {
				key = strings.ToLower(strings.TrimSpace(stdin.Text()))
			}
			fmt.Println()
			switch key {
			case "a":
				decision = "approved"
				d := draft
				final = &d
			case "e":
				decision = "edited"
				edited := readEdited(stdin, draft)
				final = &edited
			case "r":
				decision = "rejected"
			default:
				decision = "skipped"
			}
		}

		if final != nil {
			path := filepath.Join(outboxDir, inq.ID+".txt")
			if err := os.WriteFile(path, []byte(*final+"\n"), 0o644); err != nil {
				log.Fatal(err)
			}
			fmt.Printf("  -> %s, queued at %s\n\n", decision, path)
		} else {
			fmt.Printf("  -> %s, nothing queued\n\n", decision)
		}

		d := draft
		finalForLog := ""
		if final != nil {
			finalForLog = *final
		}
		logDecision(decisionsPath, decisionRecord{
			At: nowUTC(), InquiryID: inq.ID, Category: inq.Category, Lane: lane,
			Decision: decision, Reviewer: reviewer, Draft: &d, Final: final,
			EditDistance: editDistance(draft, finalForLog),
		})
		counts[decision]++
	}

	fmt.Println(strings.Repeat("=", 72))
	var keys []string
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = fmt.Sprintf("%d %s", counts[k], k)
	}
	fmt.Println("Queue done: " + strings.Join(parts, ", "))
	fmt.Printf("Audit trail: %s   ·   Outbox: %s/\n", decisionsPath, outboxDir)
}

func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func currentUsername() string {
	u, err := user.Current()
	if err != nil {
		return "unknown"
	}
	name := u.Username
	if idx := strings.LastIndex(name, `\`); idx >= 0 {
		name = name[idx+1:]
	}
	return name
}

func indent(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		lines[i] = "  | " + strings.TrimRight(l, " \t\r")
	}
	return strings.Join(lines, "\n")
}

// readEdited collects the reviewer's replacement text, terminated by a lone
// "." on its own line. Pressing Enter on the first line copies the draft in
// so the reviewer can append to it instead of retyping it.
func readEdited(stdin *bufio.Scanner, draft string) string {
	fmt.Println("  Type the reply you want to send. End with a single '.' on its own line.")
	fmt.Println("  Press Enter on the first line to start from the draft text instead.")
	fmt.Println()

	var sb strings.Builder
	first := true
	for stdin.Scan() {
		line := stdin.Text()
		if line == "." {
			break
		}
		if first && line == "" {
			sb.WriteString(draft)
			sb.WriteString("\n")
			fmt.Println("  (draft copied in; keep typing to append, '.' to finish)")
		} else {
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		first = false
	}
	edited := strings.TrimSpace(sb.String())
	if edited == "" {
		return draft
	}
	return edited
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

func logDecision(path string, d decisionRecord) {
	body, err := json.Marshal(d)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.Write(append(body, '\n')); err != nil {
		log.Fatal(err)
	}
}

// editDistance is logged on every decision so the promotion question has data
// behind it rather than a feeling. It measures how much someone typed, not
// whether they were fixing a comma or preventing a lawsuit, so it can support
// an argument for promoting a lane and must never be the only evidence for one.
// O(a*b) and unbounded by draft length; fine for a review queue, not for bulk.
func editDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	previous := make([]int, len(rb)+1)
	current := make([]int, len(rb)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		current[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			current[j] = min3(current[j-1]+1, previous[j]+1, previous[j-1]+cost)
		}
		previous, current = current, previous
	}
	return previous[len(rb)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
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
