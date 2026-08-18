# TypeScript Demo for 06 Recommendations

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: the "you might also like" box, picking five trails at random. Which is the current feature.
- `complete/index.ts`: the finished demo as shown on stage. Same trail vectors as feature 04 (cached to `embeddings.json`), "more like this" as nearest neighbors of one item's vector, and `--gear` for the same trick over product reviews, where the top hit for the Cascade 65 is the Cascade 40.

Setup once (`npm install` in this directory), then:

```bash
npm run complete                          # "more like this" for Avalanche Lake Trail
npm run complete -- trail-0008            # any trail id works
npm run complete -- Trail of the Cedars   # so does any name (or part of one)
npm run complete -- --gear Cascade 65     # the same trick on gear, from review text
```

Real output for all four commands, including the neighbors that are obviously wrong, is in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the TypeScript equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F06-lab.md`](../F06-lab.md), done in TypeScript: start from `starter/index.ts` and end where `complete/index.ts` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run with `npm run starter` from the `typescript/` directory; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Random Carousel

This is the recommendation feature most apps ship: five trails picked at random under "you might also like". Run it twice for the same trail and get two different lists.

Run:

```bash
npm run starter
```

Check: Nothing about the five relates to Avalanche Lake Trail.

### Step 2: Get a Vector for Every Trail (lab step 1, first half)

If you did feature 04, this is the same code and the same model. If you did not, `../data/trail-embeddings.json` has the 30 vectors precomputed (keyed by trail id, embedded from `description`) and you can load them instead of calling the model at all.

```typescript
// Option A: embed live (same as feature 04)
const embeddings = await embed(trails.map((t) => t.description));
const vectors = Object.fromEntries(trails.map((t, i) => [t.id, embeddings[i]]));
// Option B: precomputed
const vectors: Record<string, number[]> = JSON.parse(readFileSync(resolve(DATA, "trail-embeddings.json"), "utf8"));
```

Check: Whichever way, `vectors["trail-0117"]` is 768 floats. If you loaded the precomputed file, check its top-level shape first; it may wrap the vectors in an object.

### Step 3: Rank Every Other Trail by Similarity to the Target (lab step 1, second half)

"More like this" is feature 04's search with the query vector replaced by the target trail's own vector. Skip the target itself, take five.

```typescript
const hits = trails
  .filter((t) => t.id !== target.id)
  .map((t) => ({ trail: t, score: cosine(vectors[target.id], vectors[t.id]) }))
  .sort((a, b) => b.score - a.score).slice(0, 5);
for (const { trail, score } of hits) console.log(`  ${score.toFixed(4)}  ${trail.name} (${trail.park}, ${trail.difficulty}; ${trail.features.join(", ")})`);
```

Run:

```bash
npm run starter
```

Check: Gunsight Lake Approach at 0.7849 on top for Avalanche Lake Trail. Read the difficulty column: the target is a moderate family walk and most neighbors are hard. Difficulty is not in the description text, so the embedding cannot see it.

### Step 4: Do the Other Two Targets and Judge Whether You Would Ship Them (lab steps 2 and 3)

Run the other targets from `../F06-lab.md`, compare against the acceptable sets in `../expected-output.md` (there is more than one right answer), and then read your own output as a product owner. One target in this slice has no real neighbors at all; a shipping product should show nothing rather than five weak guesses.

Run:

```bash
npm run starter -- trail-0008
npm run starter -- Trail of the Cedars
```

Check: Substantial overlap with the acceptable sets, and a sentence from you on whether you would ship each list. Stretch: average two trails' vectors and rank against the average, or filter to the same park or an easier difficulty before ranking. `complete/` also has `--gear`, where the top hit for the Cascade 65 pack is the Cascade 40 pack: substitutes, not complements.
