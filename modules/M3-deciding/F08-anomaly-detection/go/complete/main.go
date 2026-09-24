// Ranks trail reports by distance from the trail's own baseline and alerts when
// several outliers land close together in time. Needs Ollama running with
// nomic-embed-text pulled.
//
//	go run .                    trail-0117, live embeddings, distance table + cluster alerts
//	go run . --raw               same trail with the task prefix removed
//	go run . --trail 0042        the other trail in the data folder
//	go run . --sigma 1.5         tighter threshold
//	go run . --window 30         wider clustering window, in days
//
// Embeddings are the only model calls. Everything after them is arithmetic, so
// the cost of this feature is one embedding per report and nothing per query.
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
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	ollamaEmbedURL = "http://localhost:11434/v1/embeddings"
	embedModel     = "nomic-embed-text"
)

type report struct {
	ID      string `json:"id"`
	TrailID string `json:"trail_id"`
	Date    string `json:"date"`
	Text    string `json:"text"`
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	trail := "0117"
	sigma := 1.0
	window := 14
	// nomic-embed-text is trained with task prefixes (search_query:, search_document:,
	// clustering:, classification:) and expects one on every input. Embedding bare text
	// still returns a well-formed vector, which is why --raw fails silently rather than
	// throwing, but the vectors land off-distribution and the ranking degrades badly.
	prefix := "classification: "

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--trail":
			i++
			trail = args[i]
		case "--sigma":
			i++
			sigma, _ = strconv.ParseFloat(args[i], 64)
		case "--window":
			i++
			window, _ = strconv.Atoi(args[i])
		case "--raw":
			prefix = ""
		default:
			log.Fatalf("unknown argument: %s", args[i])
		}
	}

	reports, err := readReports(filepath.Join(data, fmt.Sprintf("reports-%s.jsonl", trail)))
	if err != nil {
		log.Fatal(err)
	}

	texts := make([]string, len(reports))
	for i, r := range reports {
		texts[i] = prefix + r.Text
	}
	embeddings, err := embed(texts)
	if err != nil {
		log.Fatal(err)
	}
	vectors := make([][]float64, len(embeddings))
	for i, e := range embeddings {
		vectors[i] = normalize(e)
	}

	// The centroid is this trail's definition of normal, and it is built from the
	// same reports it is about to judge. Anomalies that appear often enough pull the
	// centroid toward themselves and stop looking anomalous, so a long-running
	// detector should rebuild this from a trailing window rather than the full history.
	dimensions := len(vectors[0])
	centroid := make([]float64, dimensions)
	for _, v := range vectors {
		for i := range v {
			centroid[i] += v[i] / float64(len(vectors))
		}
	}
	centroid = normalize(centroid)

	type scored struct {
		report   report
		distance float64
	}
	results := make([]scored, len(reports))
	for i, r := range reports {
		results[i] = scored{report: r, distance: cosineDistance(vectors[i], centroid)}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].distance > results[j].distance })

	// The threshold is derived from this corpus rather than hard-coded, so it travels
	// to a trail with a different spread of distances. It is still a business choice,
	// not a boundary the data hands you: sigma decides how much review you are willing
	// to pay for, and there is no value of it that separates incidents from oddities.
	var sum float64
	for _, s := range results {
		sum += s.distance
	}
	mean := sum / float64(len(results))
	var sumSqDiff float64
	for _, s := range results {
		d := s.distance - mean
		sumSqDiff += d * d
	}
	deviation := math.Sqrt(sumSqDiff / float64(len(results)))
	threshold := mean + sigma*deviation

	prefixNote := fmt.Sprintf(" · prefix %q", strings.TrimSpace(prefix))
	if prefix == "" {
		prefixNote = " · NO task prefix"
	}
	fmt.Printf("trail-%s · %d reports · nomic-embed-text (%d dims)%s\n", trail, len(reports), dimensions, prefixNote)
	fmt.Printf("mean distance %.4f · sd %.4f · threshold mean+%ssd = %.4f\n\n", mean, deviation, strconv.FormatFloat(sigma, 'f', -1, 64), threshold)

	fmt.Println("  dist    id       date        report")
	for _, s := range results {
		marker := "  "
		if s.distance > threshold {
			marker = " !"
		}
		fmt.Printf("%s%.4f  %s  %s  %s\n", marker, s.distance, s.report.ID, s.report.Date, truncate(s.report.Text, 62))
	}

	// Corroboration is what makes this alertable. A single report far from normal is
	// usually just an unusual subject, not an incident; two or more inside the window
	// mean several people independently noticed the same thing. Requiring two is what
	// keeps the alert queue small enough that someone still reads it.
	var flagged []scored
	for _, s := range results {
		if s.distance > threshold {
			flagged = append(flagged, s)
		}
	}
	sort.Slice(flagged, func(i, j int) bool { return flagged[i].report.Date < flagged[j].report.Date })

	fmt.Printf("\n%d of %d reports above threshold. Clustering them within %d days:\n\n", len(flagged), len(results), window)

	alerts := 0
	for i := 0; i < len(flagged); {
		j := i + 1
		for j < len(flagged) && daysBetween(flagged[j-1].report.Date, flagged[j].report.Date) <= window {
			j++
		}
		group := flagged[i:j]
		if len(group) >= 2 {
			alerts++
			fmt.Printf("  ALERT trail-%s: %d anomalous reports between %s and %s\n",
				trail, len(group), group[0].report.Date, group[len(group)-1].report.Date)
			for _, s := range group {
				fmt.Printf("        %s %s  %s\n", s.report.ID, s.report.Date, truncate(s.report.Text, 70))
			}
		} else {
			fmt.Printf("  (ignored) %s %s is a lone outlier, not an event\n", group[0].report.ID, group[0].report.Date)
		}
		i = j
	}
	fmt.Printf("\n%d alert(s). Model calls: %d embeddings, 0 chat completions.\n", alerts, len(reports))
}

func daysBetween(a, b string) int {
	ta, err := time.Parse("2006-01-02", a)
	if err != nil {
		log.Fatal(err)
	}
	tb, err := time.Parse("2006-01-02", b)
	if err != nil {
		log.Fatal(err)
	}
	return int(tb.Sub(ta).Hours() / 24)
}

func readReports(path string) ([]report, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var reports []report
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var r report
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return nil, err
		}
		reports = append(reports, r)
	}
	return reports, nil
}

func normalize(v []float64) []float64 {
	var sumSq float64
	for _, x := range v {
		sumSq += x * x
	}
	length := math.Sqrt(sumSq)
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = x / length
	}
	return out
}

// cosineDistance is correct only for unit-length inputs, where the dot
// product is already the cosine similarity. Every vector reaching this
// function has been through normalize; pass an unnormalized one and it
// returns a number that still looks plausible.
func cosineDistance(a, b []float64) float64 {
	var dot float64
	for i := range a {
		dot += a[i] * b[i]
	}
	return 1 - dot
}

func truncate(text string, max int) string {
	r := []rune(text)
	if len(r) <= max {
		return text
	}
	return string(r[:max-1]) + "…"
}

type embedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type embedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
}

func embed(texts []string) ([][]float64, error) {
	body, err := json.Marshal(embedRequest{Model: embedModel, Input: texts})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, ollamaEmbedURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer ollama") // placeholder: Ollama needs no key

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama returned %s: %s", resp.Status, respBody)
	}

	var parsed embedResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}
	if len(parsed.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(parsed.Data))
	}

	vectors := make([][]float64, len(parsed.Data))
	for i, d := range parsed.Data {
		vectors[i] = d.Embedding
	}
	return vectors, nil
}
