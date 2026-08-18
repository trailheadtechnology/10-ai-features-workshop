# TypeScript Demo for 04 Semantic Search

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: today's baseline. Keyword search, no AI at all: lowercase the query, keep words of three letters or more, count whole-word hits in each trail's name and description.
- `complete/index.ts`: the finished demo as shown on stage. `nomic-embed-text` through the SDK's embeddings call, the 30 trail descriptions embedded once and cached to `embeddings.json` next to the script, cosine similarity in one visible function, top 5 with scores.

Setup once (`npm install` in this directory), then:

```bash
npm run starter                                  # keyword baseline
npm run complete                                 # dog-friendly waterfall hike, not too steep
npm run complete -- somewhere quiet to take my kids
npm run complete -- an easy hike to a great view
```

Run `starter` first, get junk, then run `complete` on the same words. Delete `embeddings.json` to re-embed live and show the timing. The recorded rankings, including Taft Point at 0.4876 for the kids query, are in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the TypeScript equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F04-lab.md`](../F04-lab.md), done in TypeScript: start from `starter/index.ts` and end where `complete/index.ts` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run with `npm run starter` from the `typescript/` directory; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Keyword Baseline

The starter is today's search box: lowercase, split into words, count whole-word hits. Run the demo query and then the kids query. This is what you are beating.

Run:

```bash
npm run starter
npm run starter -- somewhere quiet to take my kids
```

Check: Junk for the first, one trail ("kids") for the second. Compare the keyword blocks in `../expected-output.md`.

### Step 2: Embed the 30 Descriptions Once (lab step 1)

Replace the keyword scoring with an embedding client and embed every trail's description. `nomic-embed-text` returns a 768-float vector per text; keep them in a dictionary keyed by trail id. Time it: it is seconds, and it happens once.

```typescript
const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
async function embed(texts: string[]): Promise<number[][]> {
  return (await client.embeddings.create({ model: "nomic-embed-text", input: texts })).data.map((d) => d.embedding);
}
const embeddings = await embed(trails.map((t) => t.description));
const vectors = Object.fromEntries(trails.map((t, i) => [t.id, embeddings[i]]));
```

Check: 30 vectors of 768 floats. Print one and look at it: it is just numbers. `complete/` caches them to `embeddings.json`; keep the cache keyed by id and delete it if the text or the model changes.

### Step 3: Embed the Query, Write Cosine Similarity, Print the Top 5 (lab step 2)

The query goes through the same model as the catalog (vectors from two models are not comparable, and cosine will still return confident numbers if you mix them). Cosine similarity fits in one visible function.

```typescript
const [queryVector] = await embed([query]);

function cosineSimilarity(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) { dot += a[i] * b[i]; magA += a[i] * a[i]; magB += b[i] * b[i]; }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}

const results = trails
  .map((t) => ({ trail: t, score: cosineSimilarity(queryVector, vectors[t.id]) }))
  .sort((a, b) => b.score - a.score).slice(0, 5);
for (const { trail, score } of results) console.log(`${score.toFixed(4)}  ${trail.id}  ${trail.name} (${trail.difficulty}, ${trail.distance_mi} mi)`);
```

Run:

```bash
npm run starter
```

Check: The gentle shaded waterfall trails at the top for the demo query, scores around 0.77. Compare `../expected-output.md`.

### Step 4: Run the Other Two Queries and Read the Scores, Not Just the Order (lab step 3)

The three test queries and their expected top hits are in `../data/queries.json`. The kids query is the one to sit with: the top hit is Taft Point, a cliff edge, at 0.4876. Perfect topical match, terrible advice, and the score is a third lower than query one's.

Run:

```bash
npm run starter -- somewhere quiet to take my kids
npm run starter -- an easy hike to a great view
```

Check: The expected trail is in your top 3 for each query. Stretch: filter on the metadata you already have (`features` contains `dog-friendly`, `distance_mi < 6`) before ranking, or blend the keyword count into the score. Either is a few lines, and either fixes Taft Point in a way no better model would.
