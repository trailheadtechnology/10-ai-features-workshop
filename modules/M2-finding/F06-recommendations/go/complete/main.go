// Finished demo: feature 04's search with the query vector replaced by an
// item's own vector. Embed once (cached to embeddings.json next to this
// file), rank every other item by cosine similarity to the target, print
// the top 5.
// Run:
//
//	go run .                        more like this, for Avalanche Lake Trail
//	go run . trail-0008              any trail id from ../../data/trails.json
//	go run . Trail of the Cedars     or any name, or part of one
//	go run . --gear Cascade 65       the same trick on gear, from review text
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

type review struct {
	ID       string `json:"id"`
	Product  string `json:"product"`
	Rating   int    `json:"rating"`
	Reviewer string `json:"reviewer"`
	Text     string `json:"text"`
}

var here string

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	here = filepath.Dir(thisFile)
	data := filepath.Join(here, "..", "..", "data")

	if len(os.Args) > 1 && os.Args[1] == "--gear" {
		if err := recommendGear(data, strings.Join(os.Args[2:], " ")); err != nil {
			log.Fatal(err)
		}
		return
	}

	raw, err := os.ReadFile(filepath.Join(data, "trails.json"))
	if err != nil {
		log.Fatal(err)
	}
	var trails []trail
	if err := json.Unmarshal(raw, &trails); err != nil {
		log.Fatal(err)
	}

	ids := make([]string, len(trails))
	texts := make([]string, len(trails))
	for i, t := range trails {
		ids[i] = t.ID
		texts[i] = t.Description
	}
	vectors, err := embedWithCache(filepath.Join(here, "embeddings.json"), ids, texts)
	if err != nil {
		log.Fatal(err)
	}

	query := "trail-0117"
	if len(os.Args) > 1 {
		query = strings.Join(os.Args[1:], " ")
	}
	target, err := findTrail(trails, query)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("You liked: %s (%s)\n", target.Name, target.Park)
	fmt.Println("You might also like:")
	fmt.Println()

	type scored struct {
		trail trail
		score float64
	}
	var hits []scored
	for _, t := range trails {
		if t.ID != target.ID {
			hits = append(hits, scored{trail: t, score: cosineSimilarity(vectors[target.ID], vectors[t.ID])})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > 5 {
		hits = hits[:5]
	}
	for _, h := range hits {
		fmt.Printf("  %.4f  %s (%s, %s; %s)\n", h.score, h.trail.Name, h.trail.Park, h.trail.Difficulty, strings.Join(h.trail.Features, ", "))
	}
}

func findTrail(trails []trail, query string) (*trail, error) {
	q := strings.ToLower(query)
	for i := range trails {
		if strings.ToLower(trails[i].ID) == q || strings.Contains(strings.ToLower(trails[i].Name), q) {
			return &trails[i], nil
		}
	}
	return nil, fmt.Errorf("no trail matches %q", query)
}

// recommendGear is the stretch goal: a product has no description, so its
// vector comes from every one of its reviews joined into one string.
func recommendGear(data, query string) error {
	raw, err := os.ReadFile(filepath.Join(data, "gear-reviews.jsonl"))
	if err != nil {
		return err
	}

	var order []string
	grouped := make(map[string][]string)
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var r review
		if err := json.Unmarshal([]byte(line), &r); err != nil {
			return err
		}
		if _, ok := grouped[r.Product]; !ok {
			order = append(order, r.Product)
		}
		grouped[r.Product] = append(grouped[r.Product], r.Text)
	}

	texts := make([]string, len(order))
	for i, p := range order {
		texts[i] = strings.Join(grouped[p], "\n")
	}
	vectors, err := embedWithCache(filepath.Join(here, "gear-embeddings.json"), order, texts)
	if err != nil {
		return err
	}

	q := strings.ToLower(query)
	target := ""
	for _, p := range order {
		if strings.Contains(strings.ToLower(p), q) {
			target = p
			break
		}
	}
	if target == "" {
		return fmt.Errorf("no product matches %q", query)
	}

	fmt.Printf("You bought: %s\n", target)
	fmt.Println("Goes well with:")
	fmt.Println()

	type scored struct {
		product string
		score   float64
	}
	var hits []scored
	for _, p := range order {
		if p != target {
			hits = append(hits, scored{product: p, score: cosineSimilarity(vectors[target], vectors[p])})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	if len(hits) > 5 {
		hits = hits[:5]
	}
	for _, h := range hits {
		fmt.Printf("  %.4f  %s\n", h.score, h.product)
	}
	return nil
}

// embedWithCache treats the cache at cachePath as valid when it holds a
// vector for every id in ids, regardless of whether the underlying text
// changed. Delete the cache file to force a re-embed.
func embedWithCache(cachePath string, ids, texts []string) (map[string][]float64, error) {
	if raw, err := os.ReadFile(cachePath); err == nil {
		var cached map[string][]float64
		if err := json.Unmarshal(raw, &cached); err != nil {
			return nil, err
		}
		complete := true
		for _, id := range ids {
			if _, ok := cached[id]; !ok {
				complete = false
				break
			}
		}
		if complete {
			return cached, nil
		}
	}

	embeddings, err := embed(texts)
	if err != nil {
		return nil, err
	}
	vectors := make(map[string][]float64, len(ids))
	for i, id := range ids {
		vectors[id] = embeddings[i]
	}

	out, err := json.Marshal(vectors)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, out, 0o644); err != nil {
		return nil, err
	}
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
