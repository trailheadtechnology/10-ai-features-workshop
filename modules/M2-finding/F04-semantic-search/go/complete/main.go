// Finished demo: embed all 30 trail descriptions once (cached to
// embeddings.json next to this file), embed the query, rank by cosine
// similarity, print the top 5.
// Run: go run . [query]          (defaults to the demo query)
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
	"strings"
	"time"
)

const (
	ollamaEmbedURL = "http://localhost:11434/v1/embeddings"
	embedModel     = "nomic-embed-text"
)

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

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	here := filepath.Dir(thisFile)
	data := filepath.Join(here, "..", "..", "data")

	query := "dog-friendly waterfall hike, not too steep"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}

	raw, err := os.ReadFile(filepath.Join(data, "trails-slice.json"))
	if err != nil {
		log.Fatal(err)
	}
	var trails []trail
	if err := json.Unmarshal(raw, &trails); err != nil {
		log.Fatal(err)
	}

	cachePath := filepath.Join(here, "embeddings.json")
	vectors, err := loadOrEmbed(cachePath, trails)
	if err != nil {
		log.Fatal(err)
	}

	queryVectors, err := embed([]string{query})
	if err != nil {
		log.Fatal(err)
	}
	queryVector := queryVectors[0]

	type scored struct {
		trail trail
		score float64
	}
	results := make([]scored, len(trails))
	for i, t := range trails {
		results[i] = scored{trail: t, score: cosineSimilarity(queryVector, vectors[t.ID])}
	}
	sort.Slice(results, func(i, j int) bool { return results[i].score > results[j].score })
	if len(results) > 5 {
		results = results[:5]
	}

	fmt.Printf("\nSemantic search: \"%s\"\n\n", query)
	for _, r := range results {
		fmt.Printf("%.4f  %s  %s (%s, %g mi)  [%s]\n",
			r.score, r.trail.ID, r.trail.Name, r.trail.Difficulty, r.trail.DistanceMi, strings.Join(r.trail.Features, ", "))
	}
}

// loadOrEmbed returns the cached id->vector map if embeddings.json exists,
// otherwise embeds every trail's description once and writes the cache.
func loadOrEmbed(cachePath string, trails []trail) (map[string][]float64, error) {
	if raw, err := os.ReadFile(cachePath); err == nil {
		var vectors map[string][]float64
		if err := json.Unmarshal(raw, &vectors); err != nil {
			return nil, err
		}
		fmt.Printf("Loaded %d cached vectors from embeddings.json\n", len(vectors))
		return vectors, nil
	}

	started := time.Now()
	texts := make([]string, len(trails))
	for i, t := range trails {
		texts[i] = t.Description
	}
	embeddings, err := embed(texts)
	if err != nil {
		return nil, err
	}

	vectors := make(map[string][]float64, len(trails))
	for i, t := range trails {
		vectors[t.ID] = embeddings[i]
	}

	out, err := json.Marshal(vectors)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, out, 0o644); err != nil {
		return nil, err
	}
	fmt.Printf("Embedded %d trail descriptions in %d ms\n", len(vectors), time.Since(started).Milliseconds())
	return vectors, nil
}

func cosineSimilarity(a, b []float64) float64 {
	var dot, magA, magB float64
	for i := range a {
		dot += a[i] * b[i]
		magA += a[i] * a[i]
		magB += b[i] * b[i]
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
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

// embed calls Ollama's OpenAI-compatible /v1/embeddings endpoint and returns
// one vector per input text, in the same order.
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
