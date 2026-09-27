// Finished demo, as shown on stage:
//
//	go run .                        extract both data/ reports, then validate
//	go run . path1.md [path2.md]    extract any report(s) instead
//
// The schema is the tripFactsSchema value below, sent as an OpenAI-compatible
// json_schema response format; Ollama returns JSON that already matches it.
// Nullable fields plus their "null if not stated" descriptions are what keep
// the sparse report honest. The validator underneath is what catches the
// times they don't.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	ollamaURL = "http://localhost:11434/v1/chat/completions"
	model     = "llama3.2"
)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	var reportPaths []string
	if len(os.Args) > 1 {
		reportPaths = os.Args[1:]
	} else {
		reportPaths = []string{filepath.Join(data, "tr-0007.md"), filepath.Join(data, "tr-0011.md")}
	}

	for _, reportPath := range reportPaths {
		raw, err := os.ReadFile(reportPath)
		if err != nil {
			log.Fatal(err)
		}
		report := stripFrontMatter(string(raw))

		// Prose in, populated struct out. Nothing here strips a markdown fence or a
		// preamble, because the json_schema response format makes the shape the
		// model's problem.
		prompt := fmt.Sprintf(`Extract the trail facts from this trip report.
Use null for any field the report does not state, and empty arrays
when nothing applies. Do not guess.

%s`, report)

		facts, err := extractTripFacts(prompt)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("== %s ==\n\n", filepath.Base(reportPath))
		fmt.Println("-- what the model gave us --")
		show(facts)

		// The schema guarantees the JSON parses, not that it is true. Every scalar
		// goes through a rule here, and anything that fails is coerced to null before
		// it could reach a database. Do not shortcut this to store `facts` directly.
		verdicts := validate(facts, report)

		fmt.Println("\n-- what the validator says --")
		for _, v := range verdicts {
			mark := "PASS  "
			if !v.Passed {
				mark = "REJECT"
			}
			shown := "null"
			if v.Value != nil {
				shown = *v.Value
			}
			fmt.Printf("  %s  %-18s %s\n", mark, v.Field, shown)
			if !v.Passed {
				fmt.Printf("          reason: %s\n", v.Reason)
			} else if v.Normalized != "" && (v.Value == nil || v.Normalized != *v.Value) {
				fmt.Printf("          normalized to: %s\n", v.Normalized)
			}
		}

		rejected := 0
		for _, v := range verdicts {
			if !v.Passed {
				rejected++
			}
		}
		fmt.Println()
		if rejected == 0 {
			fmt.Println("-- what we would store (nothing rejected this run) --")
		} else {
			plural := "fields"
			if rejected == 1 {
				plural = "field"
			}
			fmt.Printf("-- what we would store (%d %s coerced to null) --\n", rejected, plural)
		}
		// Rejected fields become null. Storing nothing beats storing a plausible lie:
		// null is a gap a human can fill, 0 is a number nobody will ever question.
		show(clean(facts, verdicts))
		fmt.Println()
	}
}

func stripFrontMatter(markdown string) string {
	parts := strings.SplitN(markdown, "---", 3)
	if len(parts) == 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(markdown)
}

// The schema does the prompting. Every scalar is nullable and every description
// says when to use null, which is the half of the hallucination fix that a plea
// in the prompt cannot do; the validator below is the other half. Making a field
// required-without-null here forces the model to invent a value for it.
type TripFacts struct {
	TrailName       *string  `json:"trail_name"`
	Park            *string  `json:"park"`
	DateHiked       *string  `json:"date_hiked"`
	DistanceMi      *float64 `json:"distance_mi"`
	ElevationGainFt *float64 `json:"elevation_gain_ft"`
	Wildlife        []string `json:"wildlife"`
	Conditions      []string `json:"conditions"`
	Hazards         []string `json:"hazards"`
}

var tripFactsSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"trail_name": map[string]any{
			"type":        []string{"string", "null"},
			"description": "The name of the trail hiked. null if the report never names the trail.",
		},
		"park": map[string]any{
			"type":        []string{"string", "null"},
			"description": "The park the trail is in. null if the report never names the park.",
		},
		"date_hiked": map[string]any{
			"type":        []string{"string", "null"},
			"description": "The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.",
		},
		"distance_mi": map[string]any{
			"type":        []string{"number", "null"},
			"description": "Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.",
		},
		"elevation_gain_ft": map[string]any{
			"type":        []string{"number", "null"},
			"description": "Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.",
		},
		"wildlife": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Animals the author actually saw on this hike. Empty array if none are mentioned.",
		},
		"conditions": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none.",
		},
		"hazards": map[string]any{
			"type":        "array",
			"items":       map[string]any{"type": "string"},
			"description": "Hazards or closures the report mentions. Empty array if none.",
		},
	},
	"required":             []string{"trail_name", "park", "date_hiked", "distance_mi", "elevation_gain_ft", "wildlife", "conditions", "hazards"},
	"additionalProperties": false,
}

func show(f TripFacts) {
	fmt.Printf("  trail:      %s\n", strOrNull(f.TrailName))
	fmt.Printf("  park:       %s\n", strOrNull(f.Park))
	fmt.Printf("  date:       %s\n", strOrNull(f.DateHiked))
	fmt.Printf("  distance:   %s mi\n", numOrNull(f.DistanceMi, 1))
	fmt.Printf("  elev gain:  %s ft\n", numOrNull(f.ElevationGainFt, 0))
	fmt.Printf("  wildlife:   [%s]\n", strings.Join(f.Wildlife, ", "))
	fmt.Printf("  conditions: [%s]\n", strings.Join(f.Conditions, ", "))
	fmt.Printf("  hazards:    [%s]\n", strings.Join(f.Hazards, ", "))
}

func strOrNull(v *string) string {
	if v == nil {
		return "null"
	}
	return *v
}

func numOrNull(v *float64, decimals int) string {
	if v == nil {
		return "null"
	}
	return formatNumber(*v, decimals)
}

// formatNumber rounds to the given number of decimals and drops a trailing
// zero or dot, the Go equivalent of C#'s "0.#" / Python's ":g" formatting.
func formatNumber(v float64, decimals int) string {
	scale := math.Pow(10, float64(decimals))
	rounded := math.Round(v*scale) / scale
	return strconv.FormatFloat(rounded, 'f', -1, 64)
}

// Verdict is field, value, passed, reason (set on failure).
type Verdict struct {
	Field  string
	Value  *string
	Passed bool
	Reason string
	// Set when a field passed only after being rewritten, e.g. a date the model
	// wrote as "July 4, 2026" that we store as "2026-07-04".
	Normalized string
}

func passVerdict(field string, value *string) Verdict {
	return Verdict{Field: field, Value: value, Passed: true}
}

func passVerdictNormalized(field string, value *string, normalized string) Verdict {
	return Verdict{Field: field, Value: value, Passed: true, Normalized: normalized}
}

func failVerdict(field string, value *string, reason string) Verdict {
	return Verdict{Field: field, Value: value, Passed: false, Reason: reason}
}

func nonEmpty(field string, value *string) Verdict {
	if value == nil {
		return passVerdict(field, nil)
	}
	if strings.TrimSpace(*value) == "" {
		shown := fmt.Sprintf("%q", *value)
		return failVerdict(field, &shown, "empty or whitespace-only string; should be null")
	}
	return passVerdict(field, value)
}

// Cheap grounding check: a name whose distinctive words never appear in the
// report is a name the model supplied. Deliberately crude, because a check worth
// having costs a dozen lines rather than a research project. The boilerplate list
// is what lets "Glacier National Park" ground on a report that only says
// "Glacier"; shortening it will start rejecting correct park names.
var boilerplate = map[string]bool{
	"national": true, "park": true, "state": true, "trail": true,
	"trailhead": true, "loop": true, "canyon": true, "falls": true, "the": true,
}

func grounded(field string, value *string, sourceText string) Verdict {
	basic := nonEmpty(field, value)
	if !basic.Passed || value == nil {
		return basic
	}

	lower := strings.ToLower(sourceText)
	isDelim := func(r rune) bool { return r == ' ' || r == '-' || r == ',' || r == '\'' }
	var words []string
	for _, w := range strings.FieldsFunc(*value, isDelim) {
		if len(w) >= 4 && !boilerplate[strings.ToLower(w)] {
			words = append(words, w)
		}
	}

	var missing []string
	if len(words) == 0 {
		// Nothing distinctive to check: fall back to the whole string.
		trimmed := strings.TrimSpace(*value)
		if !strings.Contains(lower, strings.ToLower(trimmed)) {
			missing = append(missing, trimmed)
		}
	} else {
		for _, w := range words {
			if !strings.Contains(lower, strings.ToLower(w)) {
				missing = append(missing, w)
			}
		}
	}

	if len(missing) == 0 {
		return passVerdict(field, value)
	}
	quoted := make([]string, len(missing))
	for i, m := range missing {
		quoted[i] = fmt.Sprintf("%q", m)
	}
	reason := fmt.Sprintf("not grounded in the source report (no mention of %s)", strings.Join(quoted, ", "))
	return failVerdict(field, value, reason)
}

// Explicit formats only. A real date in an odd format gets normalized;
// "last month (exact date not specified)" is prose, and prose in a date
// column is a bug waiting for a reporting query.
var dateLayouts = []string{
	"2006-01-02", "2006/01/02", "01/02/2006", "1/2/2006",
	"January 2, 2006", "Jan 2, 2006", "2 January 2006", "January 2 2006",
}

func validDate(field string, value *string) Verdict {
	basic := nonEmpty(field, value)
	if !basic.Passed || value == nil {
		return basic
	}

	trimmed := strings.TrimSpace(*value)
	var parsed time.Time
	var ok bool
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, trimmed); err == nil {
			parsed, ok = t, true
			break
		}
	}
	if !ok {
		return failVerdict(field, value, "does not parse as a date; no parser can store this")
	}
	if parsed.Year() < 1900 || parsed.Year() > 2100 {
		return failVerdict(field, value, fmt.Sprintf("parses, but the year %d is not plausible", parsed.Year()))
	}

	// Parseable but off-format gets normalized on the way to storage.
	canonical := parsed.Format("2006-01-02")
	return passVerdictNormalized(field, value, canonical)
}

func inRange(field string, value *float64, max float64, unit string) Verdict {
	if value == nil {
		return passVerdict(field, nil)
	}
	shown := fmt.Sprintf("%s %s", formatNumber(*value, 1), unit)
	if *value == 0 {
		return failVerdict(field, &shown, "0 is not a measurement; the report gave no figure, so this should be null")
	}
	if *value < 0 {
		return failVerdict(field, &shown, "negative value is impossible")
	}
	if *value > max {
		return failVerdict(field, &shown, fmt.Sprintf("implausible: over %s %s for a single day hike", formatNumber(max, 0), unit))
	}
	return passVerdict(field, &shown)
}

// The rejection rules. Ordinary code, no model involved. Each rule covers output
// llama3.2 has actually returned for these two reports, which is why none of them
// look defensive until you see the run that needs them (../../expected-output.md).
func validate(f TripFacts, sourceText string) []Verdict {
	return []Verdict{
		// A name the source text never contains does not get to be a fact, no
		// matter how plausible it reads.
		grounded("trail_name", f.TrailName, sourceText),
		grounded("park", f.Park, sourceText),
		validDate("date_hiked", f.DateHiked),
		// 0 is the dangerous near-miss: it is a value, so a pipeline stores it and
		// nothing downstream ever questions it. The honest answer is null.
		inRange("distance_mi", f.DistanceMi, 100, "mi"),
		inRange("elevation_gain_ft", f.ElevationGainFt, 20000, "ft"),
	}
}

func cleanList(items []string) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		if strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func clean(f TripFacts, verdicts []Verdict) TripFacts {
	byField := make(map[string]Verdict, len(verdicts))
	for _, v := range verdicts {
		byField[v.Field] = v
	}
	ok := func(field string) bool { return byField[field].Passed }

	result := TripFacts{
		// The arrays get no rule of their own: substring matching cannot ground a
		// free-text condition the way it grounds a name, so they are only cleaned.
		Wildlife:   cleanList(f.Wildlife),
		Conditions: cleanList(f.Conditions),
		Hazards:    cleanList(f.Hazards),
	}
	if ok("trail_name") && f.TrailName != nil {
		trimmed := strings.TrimSpace(*f.TrailName)
		result.TrailName = &trimmed
	}
	if ok("park") && f.Park != nil {
		trimmed := strings.TrimSpace(*f.Park)
		result.Park = &trimmed
	}
	if ok("date_hiked") {
		if norm := byField["date_hiked"].Normalized; norm != "" {
			n := norm
			result.DateHiked = &n
		} else if f.DateHiked != nil {
			trimmed := strings.TrimSpace(*f.DateHiked)
			if trimmed != "" {
				result.DateHiked = &trimmed
			}
		}
	}
	if ok("distance_mi") {
		result.DistanceMi = f.DistanceMi
	}
	if ok("elevation_gain_ft") {
		result.ElevationGainFt = f.ElevationGainFt
	}
	return result
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
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

type chatRequest struct {
	Model          string          `json:"model"`
	Messages       []chatMessage   `json:"messages"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func extractTripFacts(prompt string) (TripFacts, error) {
	body, err := json.Marshal(chatRequest{
		Model:    model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
		ResponseFormat: &responseFormat{
			Type: "json_schema",
			JSONSchema: jsonSchemaFormat{
				Name:   "trip_facts",
				Strict: true,
				Schema: tripFactsSchema,
			},
		},
	})
	if err != nil {
		return TripFacts{}, err
	}

	req, err := http.NewRequest(http.MethodPost, ollamaURL, bytes.NewReader(body))
	if err != nil {
		return TripFacts{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ollama") // placeholder: Ollama needs no key

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TripFacts{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return TripFacts{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return TripFacts{}, fmt.Errorf("ollama returned %s: %s", resp.Status, respBody)
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return TripFacts{}, err
	}
	if len(parsed.Choices) == 0 {
		return TripFacts{}, fmt.Errorf("no choices in response")
	}

	var facts TripFacts
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &facts); err != nil {
		return TripFacts{}, fmt.Errorf("model did not return schema-conformant JSON: %w", err)
	}
	return facts, nil
}
