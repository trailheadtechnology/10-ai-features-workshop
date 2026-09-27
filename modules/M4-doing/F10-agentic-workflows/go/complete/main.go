// Five tools over the workshop's mock APIs, wired into a hand-written tool-calling
// loop. Every tool call prints as it happens, the permit step waits for a human
// yes, and a step budget bounds the loop.
//
//	go run .                                       the capstone request
//	go run . Plan me a trip on Avalanche Lake Trail in September
//	go run . --yes <request>                       auto-approve the permit gate
//
// Model: Azure OpenAI gpt-5.5 when the room key is pasted in below; otherwise
// Ollama llama3.2, which is much weaker at sequencing five tools. See F10-go.md
// before judging a local run.
//
// There is no agent framework here on purpose. The loop is the same one the
// lab's F10-lab.md walks by hand: send the messages with the tools array, read
// the tool calls out of the reply, run them, append the results, repeat.
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
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const maxIterations = 12

var (
	dataDir            string
	autoApprovePermits bool
	calledTools        = map[string]bool{}
	lastResultIDs      []string
	stdin              = bufio.NewScanner(os.Stdin)
)

// ---------------------------------------------------------------------------
// Chat target: Azure OpenAI if a room key is pasted in, Ollama otherwise. Same
// target{url, model, headers} shape as every other Go track in this workshop.
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// Chat wire types, extended with tool-calling: a tool_calls array on assistant
// messages, a tool_call_id on tool-role replies, and a tools/tool_choice pair
// on the request. Content is a pointer because an assistant message that only
// made tool calls carries a null content field, not an empty string.
// ---------------------------------------------------------------------------

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    *string    `json:"content"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type tool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters"`
	} `json:"function"`
}

type chatRequest struct {
	Model      string        `json:"model"`
	Messages   []chatMessage `json:"messages"`
	Tools      []tool        `json:"tools,omitempty"`
	ToolChoice string        `json:"tool_choice,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message      chatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

func chat(t target, messages []chatMessage, tools []tool) (chatMessage, error) {
	reqBody := chatRequest{Model: t.model, Messages: messages}
	if len(tools) > 0 {
		reqBody.Tools = tools
		reqBody.ToolChoice = "auto"
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return chatMessage{}, err
	}

	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return chatMessage{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return chatMessage{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return chatMessage{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return chatMessage{}, fmt.Errorf("%s returned %s: %s", t.url, resp.Status, respBody)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return chatMessage{}, err
	}
	if len(parsed.Choices) == 0 {
		return chatMessage{}, fmt.Errorf("no choices in response")
	}
	return parsed.Choices[0].Message, nil
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// The five tools: ordinary Go functions over the workshop's fixture files.
// The descriptions in data/tool-definitions.json are the model's only
// documentation for each tool and parameter; treat that prose as behavior,
// not commentary. Every function prints itself on entry so the loop is
// visible while it runs.
// ---------------------------------------------------------------------------

type trail struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Park        string   `json:"park"`
	DistanceMi  float64  `json:"distance_mi"`
	ElevationFt int      `json:"elevation_ft"`
	Difficulty  string   `json:"difficulty"`
	Features    []string `json:"features"`
	Description string   `json:"description"`
}

type conditionReport struct {
	ID      string `json:"id"`
	TrailID string `json:"trail_id"`
	Date    string `json:"date"`
	Text    string `json:"text"`
}

func loadTrails() []trail {
	raw, err := os.ReadFile(filepath.Join(dataDir, "trails.json"))
	if err != nil {
		log.Fatal(err)
	}
	var trails []trail
	if err := json.Unmarshal(raw, &trails); err != nil {
		log.Fatal(err)
	}
	return trails
}

func loadConditionReports() []conditionReport {
	raw, err := os.ReadFile(filepath.Join(dataDir, "condition-reports.jsonl"))
	if err != nil {
		log.Fatal(err)
	}
	var reports []conditionReport
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var r conditionReport
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			log.Fatal(err)
		}
		reports = append(reports, r)
	}
	return reports
}

func loadJSONMap(name string) map[string]any {
	raw, err := os.ReadFile(filepath.Join(dataDir, name))
	if err != nil {
		log.Fatal(err)
	}
	var v map[string]any
	if err := json.Unmarshal(raw, &v); err != nil {
		log.Fatal(err)
	}
	return v
}

func narrate(toolName string, args map[string]any) {
	calledTools[toolName] = true
	body, _ := json.Marshal(args)
	fmt.Printf("[tool] %s %s\n", toolName, body)
}

func toolResult(payload any) string {
	var text string
	if s, ok := payload.(string); ok {
		text = s
	} else {
		b, _ := json.Marshal(payload)
		text = string(b)
	}
	preview := text
	if len(preview) > 120 {
		preview = preview[:120] + "..."
	}
	fmt.Printf("  [result] %s\n", preview)
	return text
}

// parkEntry mirrors the Python/TS park_entry helper: match a park name against
// a fixture table's keys either direction, or by the query's first word.
func parkEntry(table map[string]any, park string) any {
	parkLower := strings.ToLower(park)
	firstWord := strings.ToLower(strings.SplitN(park, " ", 2)[0])
	for name, value := range table {
		if strings.HasPrefix(name, "_") {
			continue
		}
		nameLower := strings.ToLower(name)
		if strings.Contains(parkLower, nameLower) || strings.Contains(nameLower, parkLower) || strings.Contains(nameLower, firstWord) {
			return value
		}
	}
	return nil
}

func argString(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return def
}

func argRawString(args map[string]any, key string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func argInt(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

// argFeatures accepts either a JSON array of strings or, since small models
// sometimes send a string where the schema says array, a single string.
func argFeatures(args map[string]any) []string {
	v, ok := args["features"]
	if !ok || v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if strings.TrimSpace(val) == "" {
			return nil
		}
		return []string{val}
	case []any:
		var out []string
		for _, item := range val {
			if s, ok := item.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func anyOrNil[T comparable](v T, zero T) any {
	if v == zero {
		return nil
	}
	return v
}

func searchTrails(args map[string]any) string {
	park := argString(args, "park", "Glacier National Park")
	features := argFeatures(args)
	maxDifficulty := argString(args, "max_difficulty", "")

	narrateArgs := map[string]any{"park": park, "max_difficulty": anyOrNil(maxDifficulty, "")}
	if len(features) > 0 {
		narrateArgs["features"] = features
	} else {
		narrateArgs["features"] = nil
	}
	narrate("search_trails", narrateArgs)

	rank := map[string]int{"easy": 0, "moderate": 1}
	maxRank := 2
	if maxDifficulty != "" {
		if r, ok := rank[strings.ToLower(maxDifficulty)]; ok {
			maxRank = r
		}
	}

	found := []map[string]any{}
	for _, t := range loadTrails() {
		if !strings.Contains(strings.ToLower(t.Park), strings.ToLower(park)) {
			continue
		}
		r, ok := rank[t.Difficulty]
		if !ok {
			r = 2
		}
		if r > maxRank {
			continue
		}
		// Keywords match features OR the trail name. Without the name match a
		// request for "Avalanche Lake Trail" can never find trail-0117: it sits
		// past the 8-result cut and "Avalanche" is not a feature tag, so the
		// washed-out bridge lesson never fires.
		if len(features) > 0 {
			matched := false
			for _, f := range features {
				fl := strings.ToLower(f)
				if strings.Contains(strings.ToLower(t.Name), fl) {
					matched = true
					break
				}
				for _, feat := range t.Features {
					if strings.Contains(strings.ToLower(feat), fl) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				continue
			}
		}
		found = append(found, map[string]any{
			"id": t.ID, "name": t.Name, "park": t.Park, "distance_mi": t.DistanceMi,
			"elevation_ft": t.ElevationFt, "difficulty": t.Difficulty, "features": t.Features,
		})
		if len(found) == 8 {
			break
		}
	}

	lastResultIDs = lastResultIDs[:0]
	for _, f := range found {
		lastResultIDs = append(lastResultIDs, f["id"].(string))
	}
	return toolResult(found)
}

func getWeather(args map[string]any) string {
	park := argString(args, "park", "Glacier National Park")
	narrate("get_weather", map[string]any{"park": park})
	entry := parkEntry(loadJSONMap("mock-apis/weather.json"), park)
	if entry == nil {
		return toolResult(map[string]any{"error": fmt.Sprintf("No forecast available for '%s'.", park)})
	}
	return toolResult(entry)
}

func getTrailConditions(args map[string]any) string {
	rawTrailID := argRawString(args, "trail_id")
	narrate("get_trail_conditions", map[string]any{"trail_id": anyOrNil(rawTrailID, "")})

	trailID := strings.TrimSpace(rawTrailID)
	// Every tool parameter here has a default and every failure returns an
	// error string instead of panicking. A model that supplies a missing or
	// malformed argument gets a correctable message back naming the valid
	// ids, rather than crashing the process mid-loop.
	if trailID == "" || trailID == "null" || trailID == "string" {
		candidates := "call search_trails first"
		if len(lastResultIDs) > 0 {
			candidates = strings.Join(lastResultIDs, ", ")
		}
		return toolResult(map[string]any{"error": fmt.Sprintf("trailId is required. Call this tool again with one of these ids: %s.", candidates)})
	}
	// A model may pass the trail name where an id is expected, so resolve
	// names too instead of returning nothing found.
	if !strings.HasPrefix(strings.ToLower(trailID), "trail-") {
		for _, t := range loadTrails() {
			if strings.Contains(strings.ToLower(t.Name), strings.ToLower(trailID)) {
				trailID = t.ID
				break
			}
		}
	}

	var mine []conditionReport
	for _, r := range loadConditionReports() {
		if strings.EqualFold(r.TrailID, trailID) {
			mine = append(mine, r)
		}
	}
	sort.Slice(mine, func(i, j int) bool { return mine[i].Date > mine[j].Date })
	if len(mine) > 4 {
		mine = mine[:4]
	}
	if len(mine) == 0 {
		return toolResult(map[string]any{"error": fmt.Sprintf("No condition reports found for '%s'.", trailID)})
	}

	out := make([]map[string]any, len(mine))
	for i, r := range mine {
		out[i] = map[string]any{"date": r.Date, "report": r.Text}
	}
	return toolResult(out)
}

func checkCampsites(args map[string]any) string {
	park := argString(args, "park", "Glacier National Park")
	narrate("check_campsites", map[string]any{"park": park})
	entry := parkEntry(loadJSONMap("mock-apis/campsites.json"), park)
	if entry == nil {
		return toolResult(map[string]any{"error": fmt.Sprintf("No campsite data for '%s'.", park)})
	}
	return toolResult(entry)
}

func requestPermit(args map[string]any) string {
	park := argString(args, "park", "Glacier National Park")
	zone := argString(args, "zone", "Lake McDonald / Sperry")
	dates := argString(args, "dates", "unspecified")
	groupSize := argInt(args, "group_size", 2)
	narrate("request_permit", map[string]any{"park": park, "zone": zone, "dates": dates, "group_size": groupSize})

	// Filing a permit is the one irreversible action in this agent, so it
	// never runs on the model's say-so; a human confirms first. --yes
	// bypasses the prompt and exists for demo runs only.
	fmt.Printf("  [gate] About to file a permit request: %s, zone '%s', %s, group of %d.\n", park, zone, dates, groupSize)
	var approved bool
	if autoApprovePermits {
		fmt.Println("  [gate] --yes supplied; auto-approved.")
		approved = true
	} else {
		fmt.Print("  [gate] File it? [y/N] ")
		if stdin.Scan() {
			answer := strings.ToLower(strings.TrimSpace(stdin.Text()))
			approved = answer == "y" || answer == "yes"
		}
	}
	if !approved {
		return toolResult(map[string]any{
			"status":  "cancelled",
			"message": "The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed.",
		})
	}
	permits := loadJSONMap("mock-apis/permits.json")
	return toolResult(permits["submit_response"])
}

var toolFunctions = map[string]func(map[string]any) string{
	"search_trails":        searchTrails,
	"get_weather":          getWeather,
	"get_trail_conditions": getTrailConditions,
	"check_campsites":      checkCampsites,
	"request_permit":       requestPermit,
}

func loadTools() []tool {
	raw, err := os.ReadFile(filepath.Join(dataDir, "tool-definitions.json"))
	if err != nil {
		log.Fatal(err)
	}
	var file struct {
		Tools []tool `json:"tools"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		log.Fatal(err)
	}
	return file.Tools
}

const systemPrompt = `You are the trip-planning agent for Trailhead Guides, a hiking app.
Today's date is September 11, 2026.

Plan trips using your tools; never invent trails, weather, availability,
or conditions. Every trail name, forecast, campground, and condition in
your answer must have come back from a tool call in this conversation.

Call the tools one at a time, in this order, and do not write any part of
the itinerary until all of them have been called:
1. get_weather for the park.
2. search_trails for candidate trails that fit the request.
3. get_trail_conditions for EVERY trail you intend to recommend, one call
   per trail, using the trail id returned by search_trails.
   If the newest reports for a trail mention a closure, a washout, a bridge
   that is out, or any other reason hikers are turning around, that trail is
   CLOSED. Do not schedule a day on a closed trail. Replace it with another
   trail from search_trails and state plainly, in the itinerary, that the
   original trail is closed and why.
4. check_campsites for where to stay each night.
5. request_permit once, only if a backcountry site or permit zone is involved.

If you have not yet called search_trails and get_trail_conditions, your
next move is a tool call, not prose.

Then write the final itinerary: one section per day with trail, campsite,
and how the forecast shaped the choice (put harder or more exposed hiking
on the drier days). End with the permit status.`

// ---------------------------------------------------------------------------
// The loop. Send everything so far, read the reply; if it asked for tools, run
// them and append the results; repeat until it answers in prose or the step
// budget runs out. That budget is the only bound on the loop: a model that
// keeps deciding to call one more tool has no other stopping condition.
// ---------------------------------------------------------------------------

func runAgent(t target, messages *[]chatMessage, tools []tool) string {
	for i := 0; i < maxIterations; i++ {
		msg, err := chat(t, *messages, tools)
		if err != nil {
			log.Fatal(err)
		}

		if len(msg.ToolCalls) == 0 {
			content := ""
			if msg.Content != nil {
				content = *msg.Content
			}
			*messages = append(*messages, chatMessage{Role: "assistant", Content: strPtr(content)})
			return content
		}

		// The assistant turn that made the calls has to be replayed before the
		// tool results, in the shape the API returned it.
		*messages = append(*messages, chatMessage{Role: "assistant", Content: msg.Content, ToolCalls: msg.ToolCalls})

		for _, call := range msg.ToolCalls {
			var args map[string]any
			if call.Function.Arguments != "" {
				if err := json.Unmarshal([]byte(call.Function.Arguments), &args); err != nil {
					args = map[string]any{}
				}
			} else {
				args = map[string]any{}
			}

			var output string
			fn, ok := toolFunctions[call.Function.Name]
			if !ok {
				output = mustMarshal(map[string]any{"error": fmt.Sprintf("Unknown tool '%s'.", call.Function.Name)})
			} else {
				output = safeCall(fn, call.Function.Name, args)
			}
			*messages = append(*messages, chatMessage{Role: "tool", Content: strPtr(output), ToolCallID: call.ID})
		}
	}
	fmt.Printf("[budget] stopped after %d iterations\n", maxIterations)
	return ""
}

// safeCall catches a panic from a tool function (a malformed argument set
// slipping past the defensive helpers above) and turns it into a correctable
// error message instead of crashing the loop.
func safeCall(fn func(map[string]any) string, name string, args map[string]any) (output string) {
	defer func() {
		if r := recover(); r != nil {
			output = mustMarshal(map[string]any{"error": fmt.Sprintf("Bad arguments for %s: %v", name, r)})
		}
	}()
	return fn(args)
}

func mustMarshal(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	dataDir = filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	var requestParts []string
	for _, a := range os.Args[1:] {
		if a == "--yes" {
			autoApprovePermits = true
			continue
		}
		requestParts = append(requestParts, a)
	}
	request := strings.Join(requestParts, " ")
	if request == "" {
		request = "Plan me a 3-day trip in Glacier National Park for September 14-16."
	}

	t := createChatTarget()
	tools := loadTools()

	messages := []chatMessage{
		{Role: "system", Content: strPtr(systemPrompt)},
		{Role: "user", Content: strPtr(request)},
	}

	fmt.Printf("Request: %s\n", request)
	fmt.Println(strings.Repeat("=", 60))

	// A model can stop mid-plan believing it is done and start writing the
	// itinerary from tools it never called. When a required tool is still
	// uncalled, name it and let the loop continue. Capped at three nudges so a
	// stuck model cannot spin here. A frontier model should need none of
	// these; [nudge] lines mean the model is underpowered for the task, not
	// that the app is broken.
	required := []string{"get_weather", "search_trails", "get_trail_conditions", "check_campsites"}
	answer := runAgent(t, &messages, tools)

	for i := 0; i < 3; i++ {
		var missing []string
		for _, r := range required {
			if !calledTools[r] {
				missing = append(missing, r)
			}
		}
		if len(missing) == 0 {
			break
		}
		fmt.Printf("[nudge] still missing: %s\n", strings.Join(missing, ", "))
		hint := ""
		if contains(missing, "get_trail_conditions") && len(lastResultIDs) > 0 {
			hint = fmt.Sprintf(" Use one of these trail ids: %s.", strings.Join(lastResultIDs, ", "))
		}
		nudge := fmt.Sprintf("You have not called these tools yet: %s. Call the next one now with real arguments.%s Do not write the itinerary yet.", strings.Join(missing, ", "), hint)
		messages = append(messages, chatMessage{Role: "user", Content: strPtr(nudge)})
		answer = runAgent(t, &messages, tools)
	}

	// The mirror failure: the model announces it has finished calling tools
	// and then stops without ever writing the plan. One turn asking for it
	// directly.
	if !strings.Contains(strings.ToLower(answer), "day") {
		fmt.Println("[nudge] tools are done but no itinerary was written; asking for it.")
		messages = append(messages, chatMessage{Role: "user", Content: strPtr("Every tool you need has been called. Write the final itinerary now, using only what the tools returned. Do not call any more tools.")})
		answer = runAgent(t, &messages, tools)
	}

	fmt.Println(strings.Repeat("=", 60))
	fmt.Println(answer)
}

func contains(items []string, want string) bool {
	for _, it := range items {
		if it == want {
			return true
		}
	}
	return false
}
