// Demo starting point: the "you might also like" box, filled with five
// trails picked at random. No model, no vectors — just the baseline the
// complete project replaces with embeddings.
// Run: go run . [trail-id-or-name]     (defaults to trail-0117, Avalanche Lake)
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
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

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	data := filepath.Join(filepath.Dir(thisFile), "..", "..", "data")

	raw, err := os.ReadFile(filepath.Join(data, "trails.json"))
	if err != nil {
		log.Fatal(err)
	}
	var trails []trail
	if err := json.Unmarshal(raw, &trails); err != nil {
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
	fmt.Println("You might also like (picked at random, which is the current feature):")
	fmt.Println()

	var others []trail
	for _, t := range trails {
		if t.ID != target.ID {
			others = append(others, t)
		}
	}
	rand.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	if len(others) > 5 {
		others = others[:5]
	}
	for _, t := range others {
		fmt.Printf("  %s (%s, %s)\n", t.Name, t.Park, t.Difficulty)
	}
}

// findTrail matches by exact id (case-insensitive) or by name containing the
// query (case-insensitive).
func findTrail(trails []trail, query string) (*trail, error) {
	q := strings.ToLower(query)
	for i := range trails {
		if strings.ToLower(trails[i].ID) == q || strings.Contains(strings.ToLower(trails[i].Name), q) {
			return &trails[i], nil
		}
	}
	return nil, fmt.Errorf("no trail matches %q", query)
}
