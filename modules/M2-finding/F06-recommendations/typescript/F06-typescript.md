<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 06: Recommendations (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F06-dotnet.md), [Python](../python/F06-python.md). Lab overview: [F06-lab.md](../F06-lab.md).*

**The User Problem:** A hiker just finished Avalanche Lake Trail and loved it. Trailhead Guides says nothing. The obvious next screen ("you'd probably like these three trails") never got built, because everyone assumes recommendations require a data-science team, a ratings matrix, and six months. Meanwhile the gear store has the same gap: someone who bought the Cascade 65 gets shown a random carousel instead of the products that actually go with it.

*A Challenge lab. Do it if you finished [Module 2](../../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `data/trails.json` (30 trails including the three targets), `data/trail-embeddings.json` (their `nomic-embed-text` vectors by id, so you can skip the embed call), `data/gear-reviews.jsonl` (300 reviews across 25 products, for the gear stretch goal).
- **How:** turn each trail description into a vector once with the embedding model. Score every other trail by cosine against the target trail's own vector, sort high to low, leave the target out, keep the top 5. The model only gives you vectors. The ranking is code you write. If you skipped feature 04, use the precomputed vectors in step 2.
- **Model:** `nomic-embed-text`, local. No chat model, no key.

## The Concept

The classical answer to recommendations is collaborative filtering over user-behavior data, and it's real work with a real cold-start problem: it can't say anything about a new trail nobody has rated. This feature shows the shortcut that gets you most of the value: content-based recommendations from the embeddings you already have. If feature 04 gave every trail a position in meaning-space, then "trails similar to the one you just loved" is nothing more than nearest neighbors of that trail's vector. You don't train anything or build a ratings matrix, new items have no cold-start problem, and the infrastructure already exists.

That's the deliberate narrative beat of this feature: one embedding investment keeps paying. Search (04), recommendations (06), and anomaly detection (08) are three features from one piece of infrastructure. The honest caveat gets said out loud too. Content similarity recommends things that are alike, and it will never discover that people who hike waterfalls also buy headlamps. When behavior data accumulates, collaborative filtering complements this; it doesn't replace it on day one.

Every step below is one thing to make the program do. The starter picks trails at random. Edit it until it does all six steps. Look at `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both reading from [`data/`](../data/). `starter/index.ts` is the "you might also like" box, picking five trails at random. `complete/index.ts` is the finished demo as shown on stage: feature 04's embedding code over this feature's own 30-trail slice (vectors cached to `embeddings.json`), "more like this" as nearest neighbors of one item's vector, and `--gear` for the same trick over product reviews.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. Edit `starter/index.ts` in place (or copy it first). Run everything from the `typescript/` folder, where `package.json` lives: `npm install` once, then `npm run starter`. The flags below are the ones `complete/` supports, so add the same argument parsing or hard-code the value:

```bash
npm run complete                          # "more like this" for Avalanche Lake Trail
npm run complete -- trail-0008            # any trail id works
npm run complete -- Trail of the Cedars   # so does any name (or part of one)
npm run complete -- --gear Cascade 65     # the same trick on gear, from review text
```

Real output for all four commands, including the neighbors that are obviously wrong, is in [`expected-output.md`](../expected-output.md).

### Step 0: Run the starter and look at the random list

**Do:** run `starter/` without changing it. It loads the catalog, finds Avalanche Lake Trail (`trail-0117`), and prints five other trails picked at random under "you might also like".

```bash
npm run starter
```

**Check:** the five trails have no connection to Avalanche Lake, and a second run gives you a different five. The starter never looks at the trail you liked; it just draws five of the other 29. A hiker who liked a moderate 4.6-mile lake walk is as likely to be handed a hard desert route as another lake.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails.json`: 30 objects with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description`.
2. Review how the existing code from the starter project parses the file into a list of trail objects.

   It happens near the top of `starter/index.ts`. `DATA` is the `data/` folder two levels up from `index.ts`, `JSON.parse` turns the file's text into an array, and the `Trail` type names the fields so you read one with `t.name`:

   ```typescript
   const DATA = resolve(import.meta.dirname, "../../data");
   type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };
   const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails.json"), "utf8"));
   ```

   For the Check, put two temporary logs right below the `const trails` line and delete them once the numbers match:

   ```typescript
   // Hint: two throwaway logs
   console.log(trails.length);   // how many trails
   console.log(trails[0].id);    // id of the first one
   ```

3. Review how it looks up the target: it takes the command-line argument as the query, defaults to `trail-0117`, and picks the first trail whose `id` equals the query or whose `name` contains it (case-insensitive).

   It happens right below the `trails` line. `process.argv.slice(2)` holds the words after `npm run starter --`, and `find` returns the first trail that matches, or `undefined`:

   ```typescript
   const query = process.argv.slice(2).join(" ") || "trail-0117";
   const target = trails.find((t) => t.id.toLowerCase() === query.toLowerCase() || t.name.toLowerCase().includes(query.toLowerCase()));
   if (!target) throw new Error(`No trail matches '${query}'.`);
   ```

   The Check's line comes from this `console.log`, a few lines further down:

   ```typescript
   console.log(`You liked: ${target.name} (${target.park})`);
   ```

**Why:** these 30 are lifted from the workshop's full 200-trail catalog (feature 10's `data/trails.json`), chosen to hold the three target trails plus enough real neighbors and enough noise to make ranking interesting. It's a different 30 from feature 04's slice (7 trails overlap), so feature 04's cached vectors won't cover it.

**Check:** the list has 30 entries. The first `id` is `trail-0003`. Running with no argument prints `You liked: Avalanche Lake Trail (Glacier National Park)`.

### Step 2: Get a vector for every trail, once

**Do:**
1. Build a dictionary from each trail's `id` to its `description`. The description is the only text you embed.

   In TypeScript a dictionary with string keys is a plain object, typed `Record<string, string>`. `trails.map(...)` turns each trail into an `[id, description]` pair, and `Object.fromEntries` turns the list of pairs into the object. You pass it straight to the helper in item 4, so there is nothing to keep yet. To see what it makes, try it as a throwaway log right below the `const trails` line:

   ```typescript
   // Hint: one throwaway log, delete it after
   console.log(Object.fromEntries(trails.map((t) => [t.id, t.description]))["trail-0117"]);
   ```

2. Embed all 30 descriptions in **one batch**, in id order. The embedding client and call are the same as feature 04.

   Option A, embed live (option B, loading the precomputed file instead, is at the end of item 4). The client is the `openai` package pointed at Ollama. Replace the starter's first import line, `import { readFileSync } from "node:fs";`, with this one, which adds the two file functions the cache needs:

   ```typescript
   import { existsSync, readFileSync, writeFileSync } from "node:fs";
   ```

   Then add this import below `import { resolve } from "node:path";`:

   ```typescript
   import OpenAI from "openai";
   ```

   Then add these lines right below the `const DATA = ...` line:

   ```typescript
   const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
   const EMBED_MODEL = "nomic-embed-text";
   const HERE = import.meta.dirname;
   ```

   The call sends a list of strings in one request and gets back one result per string, in the same order. `keys.map((k) => texts[k])` turns the list of ids into the list of their descriptions, and `await` waits for Ollama to answer. This line lives inside the helper in item 4, so do not type it on its own:

   ```typescript
   const response = await client.embeddings.create({ model: EMBED_MODEL, input: keys.map((k) => texts[k]) });
   ```

3. Store each returned vector in a dictionary keyed by trail `id`.

   `response.data` is an array with one entry per input, and each entry's `.embedding` is that text's array of numbers. `keys.map((k, i) => ...)` walks the ids with their position `i`, pairing the first id with the first vector, and `Object.fromEntries` turns the pairs into an object. This line also lives inside the helper in item 4:

   ```typescript
   const vectors = Object.fromEntries(keys.map((k, i) => [k, response.data[i].embedding]));
   ```

4. Write the dictionary to `embeddings.json` next to your program. At startup, check for that file and skip the call if it exists and has a key for every trail id. Shape it as a plain id-to-float-array dictionary so the same load code reads either your cache or `../../data/trail-embeddings.json`.

   Here is the whole helper. If `embeddings.json` exists and holds every id, it returns what is in the file. Otherwise it makes the call from item 2, builds the dictionary from item 3, and writes it to the file with `JSON.stringify`. It is `async` because it waits on Ollama, so it returns a `Promise` and every caller puts `await` in front. Put it below the `type Trail = ...` line and above `const trails = ...`, with a blank line on each side:

   ```typescript
   async function embedWithCache(cacheName: string, texts: Record<string, string>): Promise<Record<string, number[]>> {
     const cachePath = resolve(HERE, cacheName);
     if (existsSync(cachePath)) {
       const cached: Record<string, number[]> = JSON.parse(readFileSync(cachePath, "utf8"));
       if (Object.keys(texts).every((k) => k in cached)) return cached;
     }
     const keys = Object.keys(texts);
     const response = await client.embeddings.create({ model: EMBED_MODEL, input: keys.map((k) => texts[k]) });
     const vectors = Object.fromEntries(keys.map((k, i) => [k, response.data[i].embedding]));
     writeFileSync(cachePath, JSON.stringify(vectors));
     return vectors;
   }
   ```

   Then add this line right after `const trails = ...`. It builds item 1's dictionary and passes it in. An `await` outside any function (top-level `await`) is allowed here, because `package.json` says `"type": "module"` and `tsx` runs the file as a module:

   ```typescript
   const vectors = await embedWithCache("embeddings.json", Object.fromEntries(trails.map((t) => [t.id, t.description])));
   ```

   `HERE` is the folder `index.ts` sits in, so `embeddings.json` lands in `starter/`, even though you run from `typescript/`. Delete it to re-embed.

   Option B, precomputed. Same dictionary shape, no model call. Use this one line instead of all of option A (no new imports, client, or helper), right after `const trails = ...`:

   ```typescript
   const vectors: Record<string, number[]> = JSON.parse(readFileSync(resolve(DATA, "trail-embeddings.json"), "utf8"));
   ```

   For the Check, put temporary logs right after the `vectors` line, fill in `trail-0117`, and delete them once the numbers match. Then run a second time: the helper returns from the file without calling Ollama, so the run is instant and `ls -l starter/embeddings.json` shows the same time as after the first run:

   ```typescript
   // Hint: three throwaway logs
   console.log(Object.keys(vectors).length);     // how many keys
   console.log(vectors["trail-XXXX"].length);    // numbers per vector
   console.log(vectors["trail-XXXX"].slice(0, 3)); // first three numbers
   ```

   ```bash
   npm run starter
   ```

**Why:** shortcut if you'd rather not call the model: load `../../data/trail-embeddings.json` instead. It is the same dictionary, already computed with the same model, and the run that `expected-output.md`'s scores come from.

**Check:** 30 keys, each holding 768 numbers. `vectors["trail-0117"]` starts with `0.0091, 0.0802, -0.1689`. The second run does not call Ollama at all.

### Step 3: Bring in cosine and prove it works

**Do:**
1. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`.

   Add the function below `embedWithCache` (or below the `type Trail = ...` line if you used option B) and above `const trails = ...`, with a blank line on each side. The loop walks both arrays position by position, adding up the products for the dot product and the squares for each length:

   ```typescript
   function cosine(a: number[], b: number[]): number {
     let dot = 0, magA = 0, magB = 0;
     for (let i = 0; i < a.length; i++) {
       dot += a[i] * b[i];
       magA += a[i] * a[i];
       magB += b[i] * b[i];
     }
     return dot / (Math.sqrt(magA) * Math.sqrt(magB));
   }
   ```

2. Compute the cosine between the vectors for `trail-0117` and `trail-0086`, then between `trail-0117` and `trail-0041`. Print both.

   Put the two logs right after the `const vectors = ...` line. `.toFixed(4)` prints four decimals:

   ```typescript
   console.log(cosine(vectors["trail-0117"], vectors["trail-0086"]).toFixed(4));
   console.log(cosine(vectors["trail-0117"], vectors["trail-0041"]).toFixed(4));
   ```

   Run it, then delete the two logs once you have seen the numbers:

   ```bash
   npm run starter
   ```

**Check:** `trail-0117` vs `trail-0086` is `0.7849`. `trail-0117` vs `trail-0041` is `0.6117`. Two Glacier lake hikes score high and a Zion desert wash scores low. Equal numbers, or anything above 0.99, means you compared a vector to itself.

### Step 4: Rank the neighbors of trail-0117

**Do:**
1. Take the target trail's vector.

   `target` is a `Trail` object, so its id is `target.id` and its vector is `vectors[target.id]`. That expression goes straight into item 2's code, so there is no line to add yet.

2. Loop over every trail except the target, computing the cosine between the target's vector and each.

   Array methods chain, each one taking the previous result. `.filter(...)` keeps the trails whose test returns `true`, so `t.id !== target.id` skips the target. `.map(...)` turns each remaining trail into an object holding the trail and its score. Try it as a throwaway log right below the `if (!target) throw ...` line (after `target` is found, since the log uses it), then delete it:

   ```typescript
   // Hint: 30 trails minus the target
   console.log(trails.filter((t) => t.id !== target.id).map((t) => ({ trail: t, score: cosine(vectors[target.id], vectors[t.id]) })).length); // should print 29
   ```

3. Sort by score, highest first. Keep the top 5.

   `.sort(...)` takes a compare function that gets two items `a` and `b`. A positive result puts `b` first, so `b.score - a.score` sorts highest score first. `.slice(0, 5)` keeps the first five. Delete these two lines from the bottom of the starter:

   ```typescript
   const others = trails.filter((t) => t.id !== target.id).sort(() => Math.random() - 0.5).slice(0, 5);
   for (const t of others) console.log(`  ${t.name} (${t.park}, ${t.difficulty})`);
   ```

   and put these lines in their place, at the bottom of the file. The `.filter` and `.map` lines are item 2's chain:

   ```typescript
   const hits = trails
     .filter((t) => t.id !== target.id)
     .map((t) => ({ trail: t, score: cosine(vectors[target.id], vectors[t.id]) }))
     .sort((a, b) => b.score - a.score)
     .slice(0, 5);
   ```

4. Replace the random loop's output with one line per hit: score to four decimals, then `name (park, difficulty; features joined with commas)`. `expected-output.md` shows the id instead of the features; the scores and order are what to match, not the exact line shape.

   Add the loop below the `hits` lines. Each item in `hits` is a `{ trail, score }` object, so `for (const { trail, score } of hits)` pulls both out. `.join(", ")` glues the features array into one string:

   ```typescript
   for (const { trail, score } of hits) {
     console.log(`  ${score.toFixed(4)}  ${trail.name} (${trail.park}, ${trail.difficulty}; ${trail.features.join(", ")})`);
   }
   ```

   Then change the heading line above `hits` to drop "(picked at random, which is the current feature)", so it reads:

   ```typescript
   console.log("You might also like:\n");
   ```

   Then run:

   ```bash
   npm run starter
   ```

**Why:** "more like this" is feature 04's search with the query vector replaced by the target trail's own vector. Difficulty is not in the description text, so the embedding cannot see it.

**Check:** `trail-0086` Gunsight Lake Approach is #1 at `0.7849`, and at least three of your top 5 are in this set. Your program prints names, so both are here: `trail-0086` Gunsight Lake Approach, `trail-0168` Black Lake via Glacier Gorge, `trail-0100` Piegan Pass Trail, `trail-0091` Otokomi Lake Trail, `trail-0080` Chasm Lake Trail, `trail-0186` Snyder Lake Trail, `trail-0196` Fern Lake Trail, `trail-0141` Akaiyan Falls via Sperry Junction. Ranks 6-8 sit within 0.01 of rank 5, so a slightly different order is not a bug. If the target appears in its own list at `1.0000`, step 2 of this list is missing.

### Step 5: Rank the neighbors of trail-0003 and trail-0008

**Do:** run the program again with `trail-0003` as the argument, then again with `trail-0008`. The code does not change. Only the target vector does.

```bash
npm run starter -- trail-0003
npm run starter -- trail-0008
```

**Check:** these numbers are from the shipped vectors; embedding live gives you scores a few thousandths off and can reorder near-ties. For `trail-0003`, Grotto Falls (`trail-0027`) is #1 at `0.7858`, and at least two of these appear: `trail-0068` Carlon Falls Trail, `trail-0131` Gatlinburg Gateway Greenway, `trail-0039` Ship Harbor Nature Trail, `trail-0070` Oconaluftee River Trail, `trail-0150` Lily Lake Loop. For `trail-0008`, any five scores in the 0.71-0.75 band pass. The failure on `trail-0008` is not noticing the band is low.

### Step 6: Read the three lists and decide whether to ship

**Do:** no code here. Read the `trail-0117`, `trail-0003`, and `trail-0008` lists the way a hiker would, looking at `difficulty` and `park` as you go.

**Check:** you named Alum Cave (`trail-0010`) at #4 for `trail-0117` (Smokies, no lake, matched on prose texture) and that four of the five are rated hard while the target is a moderate family hike. You named Coalpits Wash (`trail-0048`) at #4 for `trail-0003` (desert, no trees). You said the `trail-0008` list has no real neighbor at all. Shipping any list as-is fails; a floor (say 0.74), a metadata filter, or fewer results passes.

### Stretch goals

Pick any of these. The gear one is already built in `complete/`. The other two are not. `expected-output.md` has the numbers for all three.

- **Recommend gear from review text.** Products have no descriptions, so a product's vector comes from its reviews instead. Open `../../data/gear-reviews.jsonl`: 300 reviews over 25 products, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`), 16 of them for the Cascade 65 Backpack. Group `text` by `product`, join each product's reviews with newlines into one string, embed those 25 strings the same way as step 2, and cache to `gear-embeddings.json` keyed by product name. Find the product whose name contains the query. Rank the other 24 by cosine, print the top 5. **Check:** Cascade 40 Daypack is #1 at `0.8039`, the same pack in a smaller size and the one product a Cascade 65 owner will never buy. The Summit Bear Canister, which three reviewers mention alongside the Cascade 65, is absent. Content similarity finds substitutes; complements need behavior data.

  `complete/` takes a `--gear` flag and embeds each product's joined review text through the same `embedWithCache` helper (cache `gear-embeddings.json`). From `typescript/`:

  ```bash
  npm run complete -- --gear Cascade 65
  ```

  To build it yourself (with option A from step 2, since the helper does the embedding), add a type for one review line right below the `type Trail = ...` line:

  ```typescript
  type Review = { id: string; product: string; rating: number; reviewer: string; text: string };
  ```

  Start a function below `cosine` and above `const trails = ...`. JSON Lines means one JSON object per line, so `.split("\n")` cuts the file into lines, `.filter((l) => l.trim())` drops blank ones, and `.map((l) => JSON.parse(l))` parses each. `texts` starts as an empty object. The `for` loop adds each review to its product's entry: if the product already has text, it appends a newline and the new review, otherwise it starts with this review. That builds the same product-to-text dictionary shape `embedWithCache` takes:

  ```typescript
  async function recommendGear(query: string): Promise<void> {
    const reviews: Review[] = readFileSync(resolve(DATA, "gear-reviews.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
    const texts: Record<string, string> = {};
    for (const r of reviews) texts[r.product] = texts[r.product] ? `${texts[r.product]}\n${r.text}` : r.text;

    const vectors = await embedWithCache("gear-embeddings.json", texts);
  ```

  Still inside the function, find the product whose name contains the query, the same `find` lookup the starter uses for trails. `Object.keys(texts)` is the list of product names:

  ```typescript
    const target = Object.keys(texts).find((p) => p.toLowerCase().includes(query.toLowerCase()));
    if (!target) throw new Error(`No product matches '${query}'.`);
  ```

  Finish the function with the ranking and close it with `}`. `Object.entries(vectors)` gives `[product, vector]` pairs, and `([p]) =>` or `([p, v]) =>` pulls the parts out of each pair:

  ```typescript
    console.log(`You bought: ${target}`);
    console.log("Goes well with:\n");
    const hits = Object.entries(vectors)
      .filter(([p]) => p !== target)
      .map(([p, v]) => ({ product: p, score: cosine(vectors[target], v) }))
      .sort((a, b) => b.score - a.score)
      .slice(0, 5);
    for (const { product, score } of hits) console.log(`  ${score.toFixed(4)}  ${product}`);
  }
  ```

  Finally, check for the flag right above `const trails = ...` (below the functions), so a gear run stops before the trail code. `process.exit(0)` ends the program:

  ```typescript
  const args = process.argv.slice(2);
  if (args[0] === "--gear") {
    await recommendGear(args.slice(1).join(" "));
    process.exit(0);
  }
  ```

  Then run it from `typescript/`:

  ```bash
  npm run starter -- --gear Cascade 65
  ```

- **Average two trails.** Add the vectors for `trail-0117` and `trail-0003` position by position, divide each position by 2, rank every other trail against that average, and leave both source trails out. **Check:** Carlon Falls (`trail-0068`) is #1 at `0.7947` and every score is higher than before. That rise is a property of averaging vectors, not a better result.

  Replace step 4's `const hits = ...` lines (keep the `for` loop below them). There is no averaging function to call; `.map((x, i) => ...)` builds a new array one position at a time, where `x` is the number at position `i` of `first` and `second[i]` is the matching number:

  ```typescript
  // Hint: average two vectors, then rank against the average
  const first = vectors["trail-AAAA"];
  const second = vectors["trail-BBBB"];
  const average = first.map((x, i) => (x + second[i]) / 2);
  const hits = trails
    .filter((t) => t.id !== "trail-AAAA" && t.id !== "trail-BBBB")
    .map((t) => ({ trail: t, score: cosine(average, vectors[t.id]) }))
    .sort((a, b) => b.score - a.score)
    .slice(0, 5);
  ```

  Put step 4's lines back when you are done.

- **Filter by difficulty.** Run `trail-0117` again, this time dropping every trail whose `difficulty` is not `easy` or `moderate` before you sort. **Check:** Alum Cave (`trail-0010`) is #1 at `0.7642` and Fern Lake (`trail-0196`) is #2 at `0.7508`. The filter removes what a family cannot do; it cannot invent good results, because this slice has almost no easy lake hikes.

  Add one more `.filter` line to step 4's `hits` chain, right after the first `.filter` and before `.map`. `difficulty` holds lowercase strings (`easy`, `moderate`, `hard`), and `||` means "or":

  ```typescript
  // Hint: one extra filter line in the hits chain
    .filter((t) => t.difficulty === "VALUE1" || t.difficulty === "VALUE2")
  ```

## What Is in This Folder

- `data/trails.json`: a 30-trail slice of the full catalog (feature 10's `data/trails.json`), same shape as the full file. It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting. Not the same 30 as feature 04's slice.
- `data/trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text`, keyed by trail id (768 floats each). Here so you can start cold without calling the model.
- `data/gear-reviews.jsonl`: 300 reviews across 25 products, one JSON object per line, from the shared Trailhead Guides corpus, for the gear stretch goal.
- `complete/embeddings.json` and `complete/gear-embeddings.json` (inside each track's `complete/`): vector caches the finished program writes on its first run. Delete them to re-embed.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
