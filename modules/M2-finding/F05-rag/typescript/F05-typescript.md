# TypeScript Demo for 05 RAG

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: one client, one question, no context: the confident wrong answer from step 1 of the demo.
- `complete/index.ts`: the finished demo as shown on stage. Hybrid retrieval over `../data/chunks.jsonl` (normalized cosine blended with a BM25-lite lexical score, alpha visible), the score table with both signals and the margin, the grounded prompt with the pinned date in the refusal clause, citation validation with one retry and then stripping, and generation on Azure OpenAI or the local model.

Setup once (`npm install` in this directory), then:

```bash
npm run complete                                       # the Sperry Chalet question, grounded
npm run complete -- "Is the Avalanche Lake Trail open right now?"
npm run complete -- --no-context                       # step 1: the confident wrong answer
npm run complete -- --alpha 1.0 --top-k 8 --retrieval-only   # pure cosine: the wrong-park neighbors
npm run complete -- --model qwen3:32b                  # a bigger local model, if you have the memory
```

Retrieval always runs locally on `nomic-embed-text`; the first run embeds 250 chunks (about 40 seconds) and caches them to `embeddings.json` next to the script. Every number in the retrieval table matches the .NET demo to four decimals, and the measured story behind the chunking, the alpha, the date rule, and the 32B comparison is in [`../expected-output.md`](../expected-output.md) and [`../dotnet/F05-dotnet.md`](../dotnet/F05-dotnet.md).

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, the deployment name the feature uses, and the key handed out in the room) and the generation call switches to Azure OpenAI through the SDK's `AzureOpenAI` client; leave them unset and it runs against Ollama.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the TypeScript equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F05-lab.md`](../F05-lab.md), done in TypeScript: start from `starter/index.ts` and end where `complete/index.ts` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run with `npm run starter` from the `typescript/` directory; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Confident Wrong Answer

No retrieval, no context. Ask the Sperry Chalet question and read the answer, then open `../data/park-docs/glacier-backcountry-camping-guide.md` Section 4.2 and read the actual rule. Everything that follows exists because of this gap.

Run:

```bash
npm run starter
```

Check: Fluent, specific, and wrong about a fire regulation.

### Step 2: Embed the Chunks and Retrieve the Top 3 by Cosine (lab step 1)

Load `../data/chunks.jsonl` (250 chunks with `chunk_id`, `source`, `text`), embed them with `nomic-embed-text` in batches of 32, cache the vectors (about 40 seconds the first time), embed the question, and print the top 3 with scores. If retrieval does not find the right material here, no prompt later can save you.

```typescript
const chunks: Chunk[] = readFileSync(resolve(DATA, "chunks.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
const index: Record<string, number[]> = {};
for (let start = 0; start < chunks.length; start += 32) {
  const batch = chunks.slice(start, start + 32);
  const vectors = await embed(batch.map((c) => c.text));
  batch.forEach((c, i) => { index[c.chunk_id] = vectors[i]; });
}
const [questionVector] = await embed([question]);
const scored = chunks
  .map((c) => ({ chunk: c, score: cosineSimilarity(questionVector, index[c.chunk_id]) }))
  .sort((a, b) => b.score - a.score);
for (const { chunk, score } of scored.slice(0, 3)) console.log(`${score.toFixed(4)}  ${chunk.chunk_id}`);
```

Check: `glacier-backcountry-camping-guide:04.2` at rank 1 with cosine 0.7422, and the margin over rank 2 is small. Print the scores; the margin is the story.

### Step 3: Blend in a Lexical Score so "Sperry" Counts (lab step 2)

Cosine alone puts Acadia and Yosemite campfire sections in the top 8, because the embedder collapses "campfire regulations" from five parks onto nearly the same point. Add a BM25-lite score: tokenize, weight each query word by how few chunks contain it, min-max both signals to 0..1, and combine with an alpha. The full tokenizer, stop-word list, and BM25 formula are in `complete/`; the shape is below. The commands below use the flags `complete/` has; in your own copy, change the `alpha` and top-k variables by hand.

```typescript
const idf = Object.fromEntries(queryTerms.map((t) => [t, Math.log(1 + (n - df(t) + 0.5) / (df(t) + 0.5))]));
// lexical[c] = sum over query terms of idf[t] * tf*(K1+1) / (tf + K1*(1 - B + B*len/avgLen))
const combined = alpha * semanticNorm[id] + (1 - alpha) * lexicalNorm[id];
```

Run:

```bash
npm run complete -- --top-k 8 --alpha 1.0 --retrieval-only
npm run complete -- --top-k 8 --retrieval-only
```

Check: At alpha 1.0, five of eight chunks are from the wrong park. At the default 0.6, the wrong-park chunks are replaced by Glacier documents that name Sperry and the margin over rank 2 grows from 0.16 to 0.23. Try the rephrasings listed in `../expected-output.md`.

### Step 4: Build the Grounded Prompt and Generate (lab step 3)

Context in, citations out, refusal when the context is silent. Two details are load-bearing and both are measured in `../expected-output.md`: the exact refusal string, and today's date inside the refusal clause (not in the rules block), so "is the trail open right now?" is answered from a dated notice rather than refused.

```typescript
const TODAY = "September 23, 2026";
const REFUSAL = "The provided documents don't say.";
const context = top.map((h) => `chunk_id: ${h.chunk.chunk_id}\nsource: ${h.chunk.source}\n${h.chunk.text}`).join("\n\n");
const prompt = `You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context.
- If, and only if, none of the context is relevant to the question, reply exactly: "${REFUSAL}"
  A question about "right now" is answered from the context, not refused: today is ${TODAY},
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
${context}

Question: ${question}

Answer:
`;
let answer = await generate(prompt);
```

Check: The right answer (no wood fires at Sperry, year-round) with `[glacier-backcountry-camping-guide:04.2]` cited. Ask "Is the Avalanche Lake Trail open right now?" too; without the date line it was refused in 10 runs out of 18.

### Step 5: Validate the Citations (lab step 4)

A citation is only a string the model typed. Pull every bracketed token with a colon out of the answer and check it against the ids you actually retrieved. Fail loudly on a mismatch: `complete/` retries once with the legal ids spelled out and then strips whatever is still wrong, so a bad receipt never reaches the visitor.

```typescript
function citations(text: string): string[] {
  return [...text.matchAll(/\[([^\]]*:[^\]]*)\]/g)].flatMap((m) => m[1].split(",")).map((c) => c.trim()).filter((c) => c.includes(":"));
}
const retrievedIds = new Set(top.map((h) => h.chunk.chunk_id));
const bad = [...new Set(citations(answer).filter((c) => !retrievedIds.has(c)))];
if (bad.length > 0) console.log(`!! CITATION CHECK FAILED: ${bad.join(", ")} not in the retrieved set`);
```

Check: Run the unanswerable question a few times: the model will eventually attach an invented chunk id to its own refusal, and the check catches it. Every run should end with a `[citations: N valid, M invalid]` line.

### Step 6: Run All Four Questions, Then Run Question 1 Twenty Times (lab steps 5 and 6)

`../data/questions.json` has three answerable questions and one that is not. Then loop the Sperry question: a wrong answer one run in five is invisible in a single run and is the only defect in this feature that could hurt somebody.

Run:

```bash
for i in $(seq 20); do npm run --silent complete | tail -3 | head -1; done
```

Check: Three correct cited answers, a refusal on the fourth, and no invalid citation reaching the output unflagged. Over 20 runs, count how many open with "Yes" before saying fires are banned; the measured rate for `llama3.2` is 40%, and it is 0% for the 32B model at 18x the latency (`--model qwen3:32b`, if you have the memory). Stretch: write ten more questions with the chunk that should win, sweep alpha from 0 to 1, and defend your alpha with recall@3 rather than with the Sperry question.
