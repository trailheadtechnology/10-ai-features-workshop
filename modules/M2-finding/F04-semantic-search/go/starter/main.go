// Demo starting point: keyword search over the trail slice. Split the query into
// words and count how many appear, as whole words, in each trail's name and
// description.
// Run: go run . [query]          (defaults to the demo query)
//
// Matching on words alone has no access to meaning. A description saying the
// trail avoids the steep section still matches "steep", and a trail whose text
// reads "leashed dogs are welcome" never matches "dog-friendly". That gap is
// what the complete project replaces with embeddings.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
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

var wordPattern = regexp.MustCompile(`[a-z]+`)

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

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

	tokens := tokenize(query)

	type hit struct {
		trail trail
		hits  []string
	}
	var results []hit
	for _, t := range trails {
		haystack := strings.ToLower(t.Name + " " + t.Description)
		var hits []string
		for _, w := range tokens {
			if regexp.MustCompile(`\b` + w + `\b`).MatchString(haystack) {
				hits = append(hits, w)
			}
		}
		if len(hits) > 0 {
			results = append(results, hit{trail: t, hits: hits})
		}
	}
	sort.SliceStable(results, func(i, j int) bool {
		if len(results[i].hits) != len(results[j].hits) {
			return len(results[i].hits) > len(results[j].hits)
		}
		return results[i].trail.ID < results[j].trail.ID
	})
	if len(results) > 5 {
		results = results[:5]
	}

	fmt.Printf("Keyword search: \"%s\"\n", query)
	fmt.Printf("Query words: %s\n\n", strings.Join(tokens, ", "))

	if len(results) == 0 {
		fmt.Println("No results. Not one trail contains those words.")
		return
	}

	for _, r := range results {
		fmt.Printf("%d word(s) [%s]  %s  %s (%s, %g mi)\n",
			len(r.hits), strings.Join(r.hits, ", "), r.trail.ID, r.trail.Name, r.trail.Difficulty, r.trail.DistanceMi)
	}
}

// tokenize lowercases the query, keeps runs of a-z of 3+ letters, and
// deduplicates while preserving first-seen order.
func tokenize(query string) []string {
	matches := wordPattern.FindAllString(strings.ToLower(query), -1)
	seen := make(map[string]bool)
	var tokens []string
	for _, w := range matches {
		if len(w) >= 3 && !seen[w] {
			seen[w] = true
			tokens = append(tokens, w)
		}
	}
	return tokens
}
