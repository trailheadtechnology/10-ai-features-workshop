<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 04: Semantic Search (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F04-dotnet.md), [Python](../python/F04-python.md). Lab overview: [F04-lab.md](../F04-lab.md).*

**The User Problem:** A user types "dog-friendly waterfall hike, not too steep" into Trailhead Guides. The catalog has at least a dozen perfect matches, but keyword search returns almost nothing, because no trail description contains the phrase "not too steep." One trail says "a gentle grade shaded by cedars." Another says "easy elevation, good for families." The user gets three bad results, assumes the app has no good trails, and goes back to asking strangers on Reddit.

*This is the Recommended lab for [Module 2](../../M2-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** rank trails by similarity to a natural-language query, and beat keyword search on three queries it fails.
- **Input:** `data/trails-slice.json`, 30 trails with three keyword traps; `data/queries.json`, the three test queries and their checks.
- **How:** two embedding calls to `nomic-embed-text` through your track's embeddings client, one for the catalog and one for the query. Cosine similarity and ranking are code you write.
- **Model:** `nomic-embed-text` on Ollama at `http://localhost:11434`. No key. 768 floats per input.

The big idea is to turn every trail description into a list of numbers (a vector), turn the search query into the same kind of list, and rank trails by how close the two lists are.

## The Concept

Embeddings turn text into vectors, points in a high-dimensional space where distance means similarity of meaning. "Gentle grade" and "not too steep" land close together in that space even though they share no words. Semantic search is the whole trick applied to a catalog: embed every trail description once, embed the user's query at search time, and rank by distance. The math at the center is cosine similarity, which is a few lines of code in any language.

Two things make this feature land. First, it runs entirely locally: `nomic-embed-text` is a small, free embedding model, and 200 trail descriptions embed in seconds on a laptop. Second, the search box stays a search box. Users don't have to learn anything new; the same input just starts understanding what they meant. This is also the foundation feature for the rest of the day, since RAG (05), recommendations (06), and anomaly detection (08) all reuse the idea, and 06 and 08 reuse the actual infrastructure.

Every step below is one thing to make the program do. The `starter/` is a plain keyword search with no AI in it. Edit it until it does all four steps. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both reading from [`data/`](../data/). `starter/index.ts` is the keyword search. `complete/index.ts` is the finished demo as shown on stage: `nomic-embed-text` through the SDK's embeddings call, the 30 descriptions embedded once and cached, cosine similarity in one visible function, top 5 with scores. The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

Run everything from the `typescript/` folder, where `package.json` lives. Setup once with `npm install`, then pass the query after `--`, joined with spaces:

```bash
npm run starter                                  # keyword baseline
npm run complete                                 # dog-friendly waterfall hike, not too steep
npm run complete -- somewhere quiet to take my kids
npm run complete -- an easy hike to a great view
```

`complete/` has no flags; the stretch goals are yours to write.

### Step 0: Run the starter and watch keyword search fail

**Do:** run `starter/` as it is with the first two queries below. The third one comes back in step 4, once there is something better to compare it against. It lowercases the query, keeps words of three letters or more, and counts how many appear in each trail's name and description.

```text
dog-friendly waterfall hike, not too steep
```

```text
somewhere quiet to take my kids
```

```bash
npm run starter
npm run starter -- somewhere quiet to take my kids
```

The starter already does the word splitting with this line in `starter/index.ts`. `.match(/[a-z]+/g)` pulls out runs of letters (`?? []` stands in an empty array when there are none), `.filter((w) => w.length >= 3)` keeps words of three letters or more, and `[...new Set(...)]` drops repeats while keeping their order. The `const results = trails` statement right below it counts the hits.

```typescript
const tokens = [...new Set((query.toLowerCase().match(/[a-z]+/g) ?? []).filter((w) => w.length >= 3))];
```

**Check:** the first query puts `trail-0007` Upper Yosemite Falls Trail first, a hard 7.2-mile climb that matched on `waterfall` and `too`. The second query returns exactly one trail, `trail-0004` Ocean Path, because its description happens to contain the word `kids`. Both are in [`expected-output.md`](../expected-output.md)'s keyword blocks. The rest of the lab replaces that scoring with embeddings.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails-slice.json` (that path is relative to your track's `starter/` folder): one JSON array of 30 objects, each with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description` (a three- or four-sentence blurb).
2. Review how the existing code from the starter project parses it into a list of trail objects.

   `import.meta.dirname` is the folder `index.ts` sits in, so `resolve(import.meta.dirname, "../../data")` points at `../../data/` no matter which folder you run from. `readFileSync(..., "utf8")` reads the file as text, and `JSON.parse` turns that text into an array of objects, one per trail. The `Trail` type tells TypeScript which fields each object has. It happens near the top of `index.ts`. Keep all three lines:

   ```typescript
   const DATA = resolve(import.meta.dirname, "../../data");
   ```

   ```typescript
   type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };
   ```

   ```typescript
   const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails-slice.json"), "utf8"));
   ```

**Why:** the 30 were hand-picked from feature 10's 200-trail catalog to make keyword matching fail on purpose: four dog-friendly waterfall trails whose descriptions never say so in those words (`trail-0011`, `trail-0027`, `trail-0055`, `trail-0068`), plus three keyword traps. `trail-0074` Easy Creek Trail is a hard 2,610-foot climb named after homesteader Elias Easy, `trail-0187` Dog Lake Trail prohibits pets, and `trail-0058` Panorama Cliffs Bypass contains the word "steep" while describing how it avoids the steep sections.

**Check:** the list has 30 entries. The first `id` is `trail-0003`, Trail of the Cedars.

`trails.length` is 30 and `trails[0].id` is `trail-0003`. To see it, put a temporary line right below the `const trails: Trail[] = ...` line:

```typescript
// Hint: print the count and the first id, then delete this line
console.log(`${trails.length} trails, first is ${trails[0].id}`);
```

```bash
npm run starter
```

### Step 2: Embed the 30 descriptions once and cache the vectors

**Do:**
1. Collect the 30 `description` strings, in list order.

   First, in `starter/index.ts`, comment out the keyword scoring: every line from `const tokens = ...` down to the last `}` at the bottom of the file. Select those lines and press Cmd+/ (Ctrl+/ on Windows) in VS Code, or put `//` in front of each. If you plan to try the keyword-blend stretch goal, comment it out rather than deleting it, because that goal reuses it. Keep the comment block at the top, the two `import` lines, `DATA`, the `Trail` type, `query`, and `trails`. All the new code in this lab goes at the bottom of the file, below the commented-out lines, unless an item says otherwise.

   You do not need a separate loop to collect the strings. `.map` builds a new array with one entry per trail, so this expression, used in item 2, is the array of descriptions in list order:

   ```typescript
   trails.map((t) => t.description)
   ```

2. Embed all 30 in **one call** to `nomic-embed-text` (a single batch, not a loop). The response holds one 768-float vector per input, in the order sent.

   The client is the `openai` package's default export, `OpenAI`, pointed at Ollama's OpenAI-compatible endpoint. `openai` is already in this folder's `package.json`, so `npm install` put it in `node_modules` and there is nothing else to install. Ollama ignores the `apiKey`, which just has to be non-empty. Add the import at the top of `index.ts`, below `import { resolve } from "node:path";`:

   ```typescript
   import OpenAI from "openai";
   ```

   Create the client and name the model right below the imports, above the `const DATA = ...` line:

   ```typescript
   const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
   const EMBED_MODEL = "nomic-embed-text";
   ```

   `client.embeddings.create` takes the whole array as `input` in one call and returns one item per input in `response.data`, in the order sent. It returns a Promise (the answer arrives later), so the function is `async` and uses `await` to wait for it. This helper keeps just the vectors; each `d.embedding` is a plain `number[]` of 768 floats. Put it at the bottom of the file:

   ```typescript
   async function embed(texts: string[]): Promise<number[][]> {
     const response = await client.embeddings.create({ model: EMBED_MODEL, input: texts });
     return response.data.map((d) => d.embedding);
   }
   ```

   Then call it once with all 30 descriptions, below the function (item 4 moves this line inside an `else`). `await` works here, outside any function, because `package.json` sets `"type": "module"` and `tsx` runs the file as a module:

   ```typescript
   const embeddings = await embed(trails.map((t) => t.description));
   ```

3. Walk the list and the response together and store each vector in a dictionary keyed by that trail's `id`. Store each vector as a plain array of floats, not the client's own vector type, so it serializes and indexes without surprises.

   `Record<string, number[]>` is the TypeScript type for an object whose keys are strings and whose values are arrays of numbers, which is the dictionary. `let` (not `const`) lets item 4 assign it in two different places. In `trails.map((t, i) => [t.id, embeddings[i]])`, `i` is the trail's position, so `embeddings[i]` is its vector, and `Object.fromEntries` turns those `[key, value]` pairs into the object. Nothing needs converting before `JSON.stringify`. Put these two lines directly under the `const embeddings = ...` line:

   ```typescript
   let vectors: Record<string, number[]>;
   vectors = Object.fromEntries(trails.map((t, i) => [t.id, embeddings[i]]));
   ```

4. Write the dictionary to `embeddings.json` (your track's block below says which folder it lands in). That is the file to delete when you want to re-embed. At program start, load it and skip the embed call if it exists. Wrap the embed call in a timer and print how many vectors were embedded and how many milliseconds it took, or how many were loaded from the cache.

   `resolve(import.meta.dirname, "embeddings.json")` is a file next to `index.ts`, so the cache lands at `starter/embeddings.json`. `existsSync` says whether a file is there, `readFileSync` reads it, and `writeFileSync` writes it. `performance.now()` returns a clock reading in milliseconds; subtract two readings and `Math.round` the difference. Replace the starter's first import line, `import { readFileSync } from "node:fs";`, with this one (importing `readFileSync` twice is an error):

   ```typescript
   import { existsSync, readFileSync, writeFileSync } from "node:fs";
   ```

   Replace your `const embeddings = ...`, `let vectors ...`, and `vectors = ...` lines (keep the `async function embed` above them) with this block. The `else` branch holds the embed call and the `vectors = ...` line, with the timer around them and the file write after. `JSON.stringify` turns the object into text and `JSON.parse` turns it back:

   ```typescript
   const cachePath = resolve(import.meta.dirname, "embeddings.json");
   let vectors: Record<string, number[]>;
   if (existsSync(cachePath)) {
     vectors = JSON.parse(readFileSync(cachePath, "utf8"));
     console.log(`Loaded ${Object.keys(vectors).length} cached vectors from embeddings.json`);
   } else {
     const started = performance.now();
     const embeddings = await embed(trails.map((t) => t.description));
     vectors = Object.fromEntries(trails.map((t, i) => [t.id, embeddings[i]]));
     writeFileSync(cachePath, JSON.stringify(vectors));
     console.log(`Embedded ${trails.length} trail descriptions in ${Math.round(performance.now() - started)} ms`);
   }
   ```

**Why:** `embeddings.json` is a cache your program creates, not a shipped data file. It is gitignored so the first run always embeds live. It is keyed only by `id`, so delete it whenever a description or the model changes, or every later query ranks against vectors for text that no longer exists.

**Check:** 30 keys, each holding 768 floats. The first run is the slower one, since the model is embedding; the second is instant and prints that it loaded 30 cached vectors. Fewer than 30 vectors, or one not 768 long, means the batch did not go through.

`Object.keys(vectors).length` is 30 and `vectors["trail-0003"].length` is 768. Print one and look at it: it is just numbers. The first run prints `Embedded 30 trail descriptions in ... ms`; the second run prints `Loaded 30 cached vectors from embeddings.json`. Delete `starter/embeddings.json` if the text or the model changes. A temporary check, placed below the `if`/`else` block:

```typescript
// Hint: count, length, and the first few numbers of one vector
console.log(`${Object.keys(vectors).length} vectors, ${vectors["trail-0003"].length} floats each`);
console.log(vectors["trail-0003"].slice(0, 8));
```

Run it twice:

```bash
npm run starter
npm run starter
```

### Step 3: Embed query 1, write cosine similarity, print the top 5

**Do:**
1. Review how the existing code from the starter project takes the query from the command-line arguments, joined with spaces. It defaults to this when there are none:

   ```text
   dog-friendly waterfall hike, not too steep
   ```

   It happens near the top of `index.ts`. `process.argv.slice(2)` holds the words typed after `npm run starter --`; when there are none, the joined string is empty and `||` falls back to the default:

   ```typescript
   const query = process.argv.slice(2).join(" ") || "dog-friendly waterfall hike, not too steep";
   ```

2. Embed the query as a single string, through the same client and same model. Keep the one vector that comes back.

   The query goes through the same `embed` function. `embed` takes an array and returns an array, so pass `[query]` (a one-item array); `const [queryVector] = ...` keeps the first (and only) vector. Put this at the bottom of the file, below the cache `if`/`else` block:

   ```typescript
   const [queryVector] = await embed([query]);
   ```

3. Write `cosine_similarity(a, b)`: loop `i` 0 to 767, accumulate `dot += a[i]*b[i]`, `magA += a[i]*a[i]`, `magB += b[i]*b[i]`, then return `dot / (sqrt(magA) * sqrt(magB))`.

   In TypeScript the function is `cosineSimilarity`, the index loop above as written. `Math.sqrt` is the square root, and `a.length` is 768. Put it at the bottom of the file, below the `const [queryVector] = ...` line:

   ```typescript
   function cosineSimilarity(a: number[], b: number[]): number {
     let dot = 0, magA = 0, magB = 0;
     for (let i = 0; i < a.length; i++) {
       dot += a[i] * b[i];
       magA += a[i] * a[i];
       magB += b[i] * b[i];
     }
     return dot / (Math.sqrt(magA) * Math.sqrt(magB));
   }
   ```

4. For every trail, compute the cosine between the query vector and that trail's stored vector.

   `.map` turns each trail into an object holding the trail and its score. The parentheses around `{ ... }` tell the arrow function to return that object. Put this below the `cosineSimilarity` function; item 5 finishes the statement:

   ```typescript
   const results = trails
     .map((t) => ({ trail: t, score: cosineSimilarity(queryVector, vectors[t.id]) }))
   ```

5. Sort by score, highest first, and keep the top 5.

   `.sort` takes a comparator, a function given two items `a` and `b`; returning `b.score - a.score` (positive when `b` scores higher) puts the highest first. `.slice(0, 5)` keeps the first 5. Add these two lines directly under the `.map(...)` line. The `;` ends the statement:

   ```typescript
     .sort((a, b) => b.score - a.score)
     .slice(0, 5);
   ```

6. Print each of the 5 on one line: score to four decimals, `id`, `name`, `difficulty`, `distance_mi`, and the `features` list. The semantic blocks in [`expected-output.md`](../expected-output.md) show the layout.

   `for (const { trail, score } of results)` unpacks each object into two variables. `score.toFixed(4)` formats the score to four decimals, and `trail.features.join(", ")` turns the features array into one string. Put this at the bottom of the file, below the `results` statement:

   ```typescript
   console.log(`\nSemantic search: "${query}"\n`);
   for (const { trail, score } of results) {
     console.log(`${score.toFixed(4)}  ${trail.id}  ${trail.name} (${trail.difficulty}, ${trail.distance_mi} mi)  [${trail.features.join(", ")}]`);
   }
   ```

Run:

```bash
npm run starter
```

**Why:** the query has to go through the same model as the descriptions. Vectors from two different models are not comparable, and cosine similarity will still return confident-looking numbers if you mix them. The result is a number near 1 when two texts mean similar things and near 0 when they do not.

**Check:** at least two of `trail-0068`, `trail-0055`, `trail-0027`, `trail-0011` in your top 3; recorded first is `trail-0068` Carlon Falls at `0.7733`. `trail-0007` Upper Yosemite Falls first means you are still counting keyword hits, the step 0 baseline.

### Step 4: Run the other two queries and read the scores

**Do:** run the program again with each of these:

```text
somewhere quiet to take my kids
```

```text
an easy hike to a great view
```

```bash
npm run starter -- somewhere quiet to take my kids
npm run starter -- an easy hike to a great view
```

**Why:** these come from `data/queries.json`, which holds all three test queries with `id`, `query`, `why` (what that query proves), and `success_check`. Look at the scores themselves, not just the order. When the whole top 5 is low and close together, nothing in the catalog is a strong match, and the order is mostly noise.

**Check:** query 2 puts Taft Point (`trail-0020`) first at `0.4876` (correct, and still a cliff edge), with short family-friendly trails filling the rest of the top 5, all between 0.43 and 0.49; query 3 puts `trail-0058` Panorama Cliffs Bypass in the top 3 (`0.6481`, first when I ran it), top score `0.77` on query 1. `trail-0074` Easy Creek Trail in query 3's top 5 means you are matching words, not meaning.

### Stretch goals

Pick either. Each uses information you already have, trail metadata or the keyword hits, instead of hoping the vector carries it. The embedding fixed recall (finding the right trails); these fix precision (dropping the wrong ones).

- **Filter before you rank.** Between step 3's item 3 and item 4, drop trails whose `difficulty` is `hard`. For query 1, also drop trails whose `features` list does not contain `dog-friendly`. Then rank whatever is left. **Check:** query 1 returns only dog-friendly trails; query 3 loses Beehive Loop (`trail-0017`) and Chimney Tops (`trail-0005`) from the top 5. Either still there means the filter ran after the top 5 was taken.

  Add a `.filter(...)` on `trails` before the `.map(...)` that scores them, so the filter runs before `.slice(0, 5)`. `.filter` keeps only the items for which the function returns `true`, and `t.features.includes(...)` checks whether the features array holds a string. The query 1 text is the default string on the `const query = ...` line. The hint replaces the first two lines of the step 3 `results` statement below `cosineSimilarity` (`const results = trails` and its `.map(...)` line), not the commented-out keyword one; keep the `.sort(...)` and `.slice(0, 5);` lines under it.

  ```typescript
  // Hint: filter first, then score and rank what is left
  const isQuery1 = query === "<query 1 text>";
  const results = trails
    .filter((t) => t.difficulty !== "<difficulty to drop>")
    .filter((t) => !isQuery1 || t.features.includes("<required feature>"))
    .map((t) => ({ trail: t, score: cosineSimilarity(queryVector, vectors[t.id]) }))
  ```

- **Blend in the keyword score.** Keep the keyword-hit count from step 0 (the keyword code you commented out in step 2) alongside the cosine score. Rescale both to 0..1 with min-max: subtract the lowest score, then divide by the gap between the highest and lowest. Rank on a weighted sum whose two weights add up to 1, and print both scores on each row. Run query 3 at 0.7 cosine and 0.3 keyword first, then raise the cosine weight. **Check:** every row shows a cosine score and a keyword score. At 0.7 cosine, Easy Creek Trail (`trail-0074`) enters query 3's top 5 on its keyword count alone, which means the keyword weight is too high. Measured on this slice, it stays out once the cosine weight reaches about 0.92. On 30 trails the keywords mostly add noise; Feature 05 measures the same blend on a bigger corpus, where they help.

  Uncomment the `const tokens = ...` line from step 0 and leave the rest of the keyword code commented out. It sits above your new code, so `tokens` is declared before anything uses it. Then replace the step 3 `results` statement below `cosineSimilarity` (not the commented-out keyword one) and the printing lines below it with the hint. It counts hits per trail with the same `RegExp` test the starter used, min-max rescales both scores, sorts on the weighted sum, and prints both scores. `Math.min(...array)` spreads the array into separate arguments, so it returns the smallest number in it. The two weights are numbers you pick that add up to 1:

  ```typescript
  // Hint: two scores per trail, each min-max rescaled to 0..1, ranked on a weighted sum
  const scored = trails.map((t) => {
    const haystack = `${t.name} ${t.description}`.toLowerCase();
    const hits = tokens.filter((w) => new RegExp(`\\b${w}\\b`).test(haystack)).length;
    return { trail: t, cosine: cosineSimilarity(queryVector, vectors[t.id]), hits };
  });

  function rescale(x: number, lo: number, hi: number): number {
    return hi === lo ? 0 : (x - lo) / (hi - lo);
  }

  const cosLo = Math.min(...scored.map((r) => r.cosine)), cosHi = Math.max(...scored.map((r) => r.cosine));
  const hitLo = Math.min(...scored.map((r) => r.hits)), hitHi = Math.max(...scored.map((r) => r.hits));
  const blend = (r: { cosine: number; hits: number }) => <cosine weight> * rescale(r.cosine, cosLo, cosHi) + <keyword weight> * rescale(r.hits, hitLo, hitHi);
  const results = scored.sort((a, b) => blend(b) - blend(a)).slice(0, 5);
  console.log(`\nBlended search: "${query}"\n`);
  for (const { trail, cosine, hits } of results) {
    console.log(`${cosine.toFixed(4)} cos  ${rescale(hits, hitLo, hitHi).toFixed(2)} kw  ${trail.id}  ${trail.name}`);
  }
  ```

## What Is in This Folder

- `data/trails-slice.json`: 30 trails lifted verbatim from feature 10's 200-trail [`trails.json`](../../../M4-doing/F10-agentic-workflows/data/trails.json), chosen so the catalog has four right answers for query 1 that never use the query's words, and three keyword traps. Step 1 names them.
- `data/queries.json`: the three test queries, why each one is in the set, and the success check for each. Step 4 describes the record shape.
- `embeddings.json`: not in the folder until you run the program. Step 2 explains the cache.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
