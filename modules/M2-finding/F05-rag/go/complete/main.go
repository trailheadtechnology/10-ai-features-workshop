// Finished demo, as shown on stage:
//
//	go run .                                    the Sperry Chalet question, grounded
//	go run . "Is the Avalanche Lake Trail open right now?"
//	go run . --no-context                       step 1: the confident wrong answer
//	go run . --alpha 1.0                        pure cosine, the wrong-park neighbors
//	go run . --retrieval-only                   print the score table and stop
//	go run . --top-k 8 "your question"          vary retrieval depth
//	go run . --model qwen3:32b                  a bigger model, if you have the memory
//
// Retrieval is hybrid: normalized cosine similarity blended with a BM25-lite
// lexical score, so a distinctive proper noun like "Sperry" counts for something.
// Every generated answer's [chunk_id] citations are validated against the chunks
// that were actually retrieved.
//
// Retrieval always runs locally (nomic-embed-text). Generation uses Azure OpenAI
// (gpt-4.1) when the room key is pasted in below, and falls back to the local
// model chosen just below when it is not.
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
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// Which local model generates the answer. Everything else in this file is
// identical either way: same retrieval, same chunks, same prompt.
var localModel = "llama3.2"

const (
	ollamaEmbedURL = "http://localhost:11434/v1/embeddings"
	embedModel     = "nomic-embed-text"
	embedBatchSize = 32
)

type chunk struct {
	ChunkID string `json:"chunk_id"`
	Source  string `json:"source"`
	Text    string `json:"text"`
}

func main() {
	_, thisFile, _, _ := runtime.Caller(0)
	here := filepath.Dir(thisFile)
	chunksPath := filepath.Join(here, "..", "..", "data", "chunks.jsonl")
	cachePath := filepath.Join(here, "embeddings.json")

	topK := 3
	alpha := 0.6 // weight on the semantic signal; 1.0 = cosine only
	noContext := false
	retrievalOnly := false
	var questionParts []string

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--no-context":
			noContext = true
		case "--retrieval-only":
			retrievalOnly = true
		case "--top-k":
			i++
			topK, _ = strconv.Atoi(args[i])
		case "--alpha":
			i++
			alpha, _ = strconv.ParseFloat(args[i], 64)
		case "--model":
			i++
			localModel = args[i]
		default:
			questionParts = append(questionParts, args[i])
		}
	}
	question := strings.Join(questionParts, " ")
	if question == "" {
		question = "Can I have a campfire at Sperry Chalet in September?"
	}

	// Generation: Azure OpenAI if configured, local llama3.2 (or --model) otherwise.
	endpoint := "https://trailhead-ai-workshop.openai.azure.com"
	key := "<KEY FROM INSTRUCTOR>" // paste the room key between the quotes
	deployment := "gpt-4.1"

	var chatTarget target
	if !strings.HasPrefix(key, "<") {
		chatTarget = azureTarget(endpoint, deployment, "2024-10-21", key)
		fmt.Printf("[generation: Azure OpenAI, deployment '%s']\n", deployment)
	} else {
		chatTarget = ollamaTarget(localModel)
		fmt.Printf("[generation: no room key pasted in, falling back to local %s]\n", localModel)
	}

	fmt.Printf("Q: %s\n\n", question)

	if noContext {
		// Step 1 of the demo: no retrieval, no context, pure model memory.
		answer, err := chat(chatTarget, question)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(answer)
		return
	}

	chunks, err := loadChunks(chunksPath)
	if err != nil {
		log.Fatal(err)
	}

	// Embed the chunks with local nomic-embed-text, caching to embeddings.json so
	// only the first run pays the embedding time.
	index, err := embedChunksWithCache(cachePath, chunks)
	if err != nil {
		log.Fatal(err)
	}

	// Signal 1: semantic. Feature 04's cosine search, verbatim, over the chunk index.
	questionVectors, err := embed([]string{question})
	if err != nil {
		log.Fatal(err)
	}
	questionVector := questionVectors[0]
	cosine := make(map[string]float64, len(chunks))
	for _, c := range chunks {
		cosine[c.ChunkID] = cosineSimilarity(questionVector, index[c.ChunkID])
	}

	// Signal 2: lexical. BM25-lite over the chunk text. The embedder collapses
	// "campfire regulations" from five different parks onto nearly the same point;
	// the word "Sperry" appears in exactly one document, and IDF makes that count.
	tokenized := make(map[string][]string, len(chunks))
	totalLength := 0
	for _, c := range chunks {
		terms := tokenize(c.Text)
		tokenized[c.ChunkID] = terms
		totalLength += len(terms)
	}
	avgLength := float64(totalLength) / float64(len(chunks))

	docFreq := make(map[string]int)
	for _, terms := range tokenized {
		for term := range distinctSet(terms) {
			docFreq[term]++
		}
	}

	const k1, b = 1.2, 0.3
	n := float64(len(chunks))
	queryTerms := distinct(tokenize(question))
	// IDF: a term in 1 of 250 chunks is worth far more than one in 200 of them.
	idf := make(map[string]float64, len(queryTerms))
	for _, t := range queryTerms {
		df := float64(docFreq[t])
		idf[t] = math.Log(1 + (n-df+0.5)/(df+0.5))
	}

	lexical := make(map[string]float64, len(chunks))
	for _, c := range chunks {
		terms := tokenized[c.ChunkID]
		counts := make(map[string]float64)
		for _, t := range terms {
			counts[t]++
		}
		score := 0.0
		for _, t := range queryTerms {
			tf, ok := counts[t]
			if !ok {
				continue
			}
			score += idf[t] * (tf * (k1 + 1)) / (tf + k1*(1-b+b*float64(len(terms))/avgLength))
		}
		lexical[c.ChunkID] = score
	}

	// Rescale both signals to 0..1 across the corpus for this question, so alpha
	// means what it looks like it means and the two numbers are comparable on screen.
	semanticNorm := minMax(cosine)
	lexicalNorm := minMax(lexical)

	type hit struct {
		chunk        chunk
		cosine       float64
		semanticNorm float64
		lexical      float64
		lexicalNorm  float64
		combined     float64
	}
	scored := make([]hit, len(chunks))
	for i, c := range chunks {
		scored[i] = hit{
			chunk:        c,
			cosine:       cosine[c.ChunkID],
			semanticNorm: semanticNorm[c.ChunkID],
			lexical:      lexical[c.ChunkID],
			lexicalNorm:  lexicalNorm[c.ChunkID],
			combined:     alpha*semanticNorm[c.ChunkID] + (1-alpha)*lexicalNorm[c.ChunkID],
		}
	}
	sort.Slice(scored, func(i, j int) bool { return scored[i].combined > scored[j].combined })
	top := scored
	if len(top) > topK {
		top = top[:topK]
	}

	sortedTerms := append([]string{}, queryTerms...)
	sort.Slice(sortedTerms, func(i, j int) bool { return idf[sortedTerms[i]] > idf[sortedTerms[j]] })
	if len(sortedTerms) > 5 {
		sortedTerms = sortedTerms[:5]
	}
	distinctiveParts := make([]string, len(sortedTerms))
	for i, t := range sortedTerms {
		distinctiveParts[i] = fmt.Sprintf("%s(%.2f/%d chunks)", t, idf[t], docFreq[t])
	}
	fmt.Printf("[query terms by IDF]  %s\n", strings.Join(distinctiveParts, "  "))
	fmt.Printf("[retrieved top %d]  combined = %.2f * semantic + %.2f * lexical\n", topK, alpha, 1-alpha)
	fmt.Println("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id")
	for r, h := range top {
		fmt.Printf("  %4d  %8.4f   %5.3f (%.4f)   %5.3f (%5.2f)   %s\n",
			r+1, h.combined, h.semanticNorm, h.cosine, h.lexicalNorm, h.lexical, h.chunk.ChunkID)
	}
	margin := 0.0
	if len(scored) > 1 {
		margin = scored[0].combined - scored[1].combined
	}
	fmt.Printf("  margin over rank 2: %.4f\n\n", margin)

	if retrievalOnly {
		return
	}

	// Grounded prompt: context in, citations out, refusal when the context is silent.
	//
	// The corpus is full of dated notices ("CLOSED effective June 20, 2026, until further
	// notice"). A model with no idea what day it is treats "is the trail open right now?" as
	// a question its documents cannot speak to, and refuses. So we tell it the date.
	//
	// The date belongs in the refusal clause specifically, not in the Rules block above.
	// Stated broadly, it invites the model to apply effective-date reasoning to rules that
	// have no expiry, and it starts deciding a year-round fire ban has lapsed.
	//
	// Production passes today's date here; this demo pins a date so the recorded outputs in
	// ../../expected-output.md stay reproducible.
	const today = "September 23, 2026"
	const refusal = "The provided documents don't say."

	retrievedIds := make(map[string]bool, len(top))
	for _, h := range top {
		retrievedIds[h.chunk.ChunkID] = true
	}
	contextParts := make([]string, len(top))
	for i, h := range top {
		contextParts[i] = fmt.Sprintf("chunk_id: %s\nsource: %s\n%s", h.chunk.ChunkID, h.chunk.Source, h.chunk.Text)
	}
	context := strings.Join(contextParts, "\n\n")

	prompt := fmt.Sprintf(`You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "%s"
  A question about "right now" is answered from the context, not refused: today is %s,
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
%s

Question: %s

Answer:
`, refusal, today, context, question)

	answer, err := chat(chatTarget, prompt)
	if err != nil {
		log.Fatal(err)
	}
	bad := invalidCitations(answer, retrievedIds)

	// Citation validation. An invalid citation is a product defect, not a style
	// issue: it is a receipt pointing at a document nobody retrieved, and sometimes
	// at a document that does not exist. Retry once with the valid ids spelled out,
	// then strip whatever is still wrong so a bad receipt never reaches the visitor.
	if len(bad) > 0 && strings.Contains(answer, refusal) {
		// A refusal that cites a source is already correct; by definition it has no sources,
		// so the only defect is the citation itself. Do not send this back to the model.
		fmt.Printf("!! CITATION CHECK FAILED: %s not in the retrieved set\n", bracketed(bad))
		fmt.Println("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed")
		fmt.Println()
		answer = refusal
		bad = nil
	} else if len(bad) > 0 {
		fmt.Printf("!! CITATION CHECK FAILED: %s not in the retrieved set\n", bracketed(bad))
		fmt.Println("!! retrying once with the valid chunk_ids spelled out")
		fmt.Println()

		var ids []string
		for id := range retrievedIds {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		retryPrompt := prompt + fmt.Sprintf("\n\nYour previous answer cited %s, which is not a real chunk_id.\nThe only chunk_ids you may cite are, exactly:\n  %s\nRewrite the answer using only those.",
			bracketed(bad), strings.Join(ids, "\n  "))
		answer, err = chat(chatTarget, retryPrompt)
		if err != nil {
			log.Fatal(err)
		}
		bad = invalidCitations(answer, retrievedIds)

		if len(bad) > 0 {
			fmt.Printf("!! STILL INVALID after retry: %s\n", bracketed(bad))
			fmt.Println("!! stripping them; the answer below is unverified where the citation was removed")
			fmt.Println()
			for _, c := range bad {
				answer = strings.ReplaceAll(answer, c, "invalid-citation-removed")
			}
		}
	}

	fmt.Println(answer)

	var cited []string
	seen := make(map[string]bool)
	for _, c := range citations(answer) {
		if retrievedIds[c] && !seen[c] {
			seen[c] = true
			cited = append(cited, c)
		}
	}
	fmt.Printf("\n[citations: %d valid (%s), %d invalid]\n", len(cited), strings.Join(cited, ", "), len(bad))
}

func loadChunks(path string) ([]chunk, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var chunks []chunk
	for _, line := range strings.Split(string(raw), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var c chunk
		if err := json.Unmarshal([]byte(line), &c); err != nil {
			return nil, err
		}
		chunks = append(chunks, c)
	}
	return chunks, nil
}

// embedChunksWithCache embeds every chunk in batches of 32 and caches the
// result to cachePath, or loads it straight from the cache if present.
func embedChunksWithCache(cachePath string, chunks []chunk) (map[string][]float64, error) {
	if raw, err := os.ReadFile(cachePath); err == nil {
		var index map[string][]float64
		if err := json.Unmarshal(raw, &index); err != nil {
			return nil, err
		}
		return index, nil
	}

	fmt.Printf("[embedding %d chunks, one-time; caching to %s]\n", len(chunks), cachePath)
	index := make(map[string][]float64, len(chunks))
	for i := 0; i < len(chunks); i += embedBatchSize {
		end := i + embedBatchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		batch := chunks[i:end]
		texts := make([]string, len(batch))
		for j, c := range batch {
			texts[j] = c.Text
		}
		embeddings, err := embed(texts)
		if err != nil {
			return nil, err
		}
		for j, c := range batch {
			index[c.ChunkID] = embeddings[j]
		}
	}

	out, err := json.Marshal(index)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(cachePath, out, 0o644); err != nil {
		return nil, err
	}
	return index, nil
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

func minMax(raw map[string]float64) map[string]float64 {
	min, max := math.Inf(1), math.Inf(-1)
	for _, v := range raw {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	out := make(map[string]float64, len(raw))
	rng := max - min
	for k, v := range raw {
		if rng > 1e-9 {
			out[k] = (v - min) / rng
		} else {
			out[k] = 0.0
		}
	}
	return out
}

// Question filler. Without this, "can" and "have" carry as much weight as
// "Sperry" simply because no park document says "can I have".
var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true, "not": true, "you": true, "your": true,
	"with": true, "that": true, "this": true, "these": true, "those": true, "from": true, "have": true,
	"has": true, "had": true, "was": true, "were": true, "been": true, "being": true, "can": true, "could": true,
	"will": true, "would": true, "shall": true, "should": true, "may": true, "might": true, "must": true,
	"does": true, "did": true, "doing": true, "what": true, "when": true, "where": true, "which": true,
	"who": true, "whom": true, "why": true, "how": true, "any": true, "all": true, "some": true, "there": true,
	"here": true, "then": true, "than": true, "them": true, "they": true, "their": true, "its": true, "his": true,
	"her": true, "our": true, "get": true, "got": true, "still": true, "now": true, "right": true, "just": true,
	"about": true, "into": true, "onto": true, "over": true, "under": true, "out": true, "off": true, "per": true,
	"via": true, "one": true, "two": true, "also": true, "more": true, "most": true, "much": true, "many": true,
	"each": true, "other": true, "such": true, "only": true, "own": true, "same": true, "too": true, "very": true,
	"let": true, "need": true, "want": true,
}

var tokenPattern = regexp.MustCompile(`[a-z0-9]+`)

// tokenize lowercases, splits on non-alphanumerics, drops question filler,
// and knocks the plural off so "campfires" in a document matches "campfire"
// in a question. Crude on purpose: an attendee can read all of it in ten seconds.
func tokenize(text string) []string {
	matches := tokenPattern.FindAllString(strings.ToLower(text), -1)
	var out []string
	for _, t := range matches {
		if len(t) <= 2 || stopWords[t] {
			continue
		}
		if len(t) > 3 && strings.HasSuffix(t, "s") && !strings.HasSuffix(t, "ss") {
			t = t[:len(t)-1]
		}
		out = append(out, t)
	}
	return out
}

func distinct(tokens []string) []string {
	seen := make(map[string]bool, len(tokens))
	var out []string
	for _, t := range tokens {
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

func distinctSet(tokens []string) map[string]bool {
	set := make(map[string]bool, len(tokens))
	for _, t := range tokens {
		set[t] = true
	}
	return set
}

// Any bracketed token containing a colon is a citation attempt, including the
// comma-separated lists llama3.2 sometimes writes. Being generous about what
// counts as an attempt is the point: we want to catch the near-misses.
var citationPattern = regexp.MustCompile(`\[([^\]]*:[^\]]*)\]`)

func citations(text string) []string {
	var out []string
	for _, m := range citationPattern.FindAllStringSubmatch(text, -1) {
		for _, part := range strings.Split(m[1], ",") {
			part = strings.TrimSpace(part)
			if strings.Contains(part, ":") {
				out = append(out, part)
			}
		}
	}
	return out
}

func invalidCitations(text string, valid map[string]bool) []string {
	seen := make(map[string]bool)
	var out []string
	for _, c := range citations(text) {
		if !valid[c] && !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}

func bracketed(ids []string) string {
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "[" + id + "]"
	}
	return strings.Join(parts, ", ")
}

// target is one chat endpoint: its URL, model name, and any headers it needs.
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

func chat(t target, prompt string) (string, error) {
	body, err := json.Marshal(chatRequest{
		Model:    t.model,
		Messages: []chatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, t.url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}

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
		return "", fmt.Errorf("%s returned %s: %s", t.url, resp.Status, respBody)
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
