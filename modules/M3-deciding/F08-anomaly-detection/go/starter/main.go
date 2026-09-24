// Starting point: ranks trail reports by distance from the trail's baseline.
// Run: go run .
//
// Makes no model calls and needs no network. The vectors in
// ../../data/embeddings-0117.json were precomputed with nomic-embed-text so this
// runs with Ollama down, which is the fallback when the room's network is not
// cooperating. Those vectors were embedded with the "classification: " task
// prefix nomic requires, so anything you add to this corpus must be embedded the
// same way or its distances will not be comparable to these.
//
// complete/ embeds live and adds the alert rule on top of this ranking.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type report struct {
	ID      string `json:"id"`
	TrailID string `json:"trail_id"`
	Date    string `json:"date"`
	Text    string `json:"text"`
}

type embeddingsFile struct {
	Embeddings map[string][]float64 `json:"embeddings"`
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	reports, err := readReports(filepath.Join(data, "reports-0117.jsonl"))
	if err != nil {
		log.Fatal(err)
	}

	raw, err := os.ReadFile(filepath.Join(data, "embeddings-0117.json"))
	if err != nil {
		log.Fatal(err)
	}
	var file embeddingsFile
	if err := json.Unmarshal(raw, &file); err != nil {
		log.Fatal(err)
	}
	vectors := make(map[string][]float64, len(file.Embeddings))
	for id, v := range file.Embeddings {
		vectors[id] = normalize(v)
	}

	// The centroid is this trail's definition of normal, and it is built from the
	// same reports it is about to judge. Anomalies that appear often enough pull the
	// centroid toward themselves and stop looking anomalous.
	dimensions := len(vectors[reports[0].ID])
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
		results[i] = scored{report: r, distance: cosineDistance(vectors[r.ID], centroid)}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].distance > results[j].distance })

	fmt.Printf("trail-0117 · %d reports · %d-dim nomic-embed-text vectors\n\n", len(reports), dimensions)
	fmt.Println("  dist    id       date        report")
	for _, s := range results {
		fmt.Printf("  %.4f  %s  %s  %s\n", s.distance, s.report.ID, s.report.Date, truncate(s.report.Text, 62))
	}
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
