# Lab 06: Recommendations

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `data/trails.json` (30 trails including the three targets), `data/trail-embeddings.json` (their `nomic-embed-text` vectors by id, so you can skip the embed call), `data/gear-reviews.jsonl` (300 reviews across 25 products, for the gear stretch goal).
- **How:** turn each trail description into a vector once with the embedding model. Score every other trail by cosine against the target trail's own vector, sort high to low, leave the target out, keep the top 5. The model only gives you vectors — the ranking is code you write.
- **Model:** `nomic-embed-text`, local. No chat model, no key.

Every step below is one thing to make the program do. Every track's `starter/` picks trails at random. Edit it until it does all six steps. Look at `complete/` when you get stuck.

### Step 0: Run the starter and look at the random list

**Do:** run `starter/` without changing it. It loads the catalog, finds Avalanche Lake Trail (`trail-0117`), and prints five other trails picked at random under "you might also like".

**Check:** five trails with no connection to Avalanche Lake. Run it again and the five change. That is the recommendation feature most apps ship today, and the rest of the lab replaces it.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails.json`: 30 objects with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description`.
2. Parse the file into a list of trail objects (the starter already does this).
3. Keep the target lookup the starter already has: take the command-line argument as the query, default to `trail-0117`, pick the first trail whose `id` equals the query or whose `name` contains it (case-insensitive).

**Why:** these 30 are lifted from the workshop's full 200-trail catalog (feature 10's `data/trails.json`), chosen to hold the three target trails plus enough real neighbors and enough noise to make ranking interesting. It's a different 30 from feature 04's slice (7 trails overlap), so feature 04's cached vectors won't cover it.

**Check:** the list has 30 entries. The first `id` is `trail-0003`. Running with no argument prints `You liked: Avalanche Lake Trail (Glacier National Park)`.

### Step 2: Get a vector for every trail, once

**Do:**
1. Build a dictionary from each trail's `id` to its `description` — the description is the only text you embed.
2. Embed all 30 descriptions in **one batch**, in id order.
3. Store each returned vector in a dictionary keyed by trail `id`.
4. Write the dictionary to `embeddings.json` next to your program. At startup, check for that file and skip the call if it exists and has a key for every trail id.

**Why:** shortcut if you'd rather not call the model: load `../../data/trail-embeddings.json` instead — the same dictionary, already computed with the same model, and the run that `expected-output.md`'s scores come from.

**Check:** 30 keys, each holding 768 numbers. `vectors["trail-0117"]` starts with `0.0091, 0.0802, -0.1689`. The second run does not call Ollama at all.

### Step 3: Bring in cosine and prove it works

**Do:**
1. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`.
2. Compute the cosine between the vectors for `trail-0117` and `trail-0086`, then between `trail-0117` and `trail-0041`. Print both.

**Check:** `trail-0117` vs `trail-0086` is `0.7849`. `trail-0117` vs `trail-0041` is `0.6117`. Two Glacier lake hikes score high and a Zion desert wash scores low. Equal numbers, or anything above 0.99, means you compared a vector to itself.

### Step 4: Rank the neighbors of trail-0117

**Do:**
1. Take the target trail's vector.
2. Loop over every trail except the target, computing the cosine between the target's vector and each.
3. Sort by score, highest first. Keep the top 5.
4. Replace the random loop's output with the top 5: score to four decimals, `name`, `park`, `difficulty`, `features` joined with commas.

**Check:** `trail-0086` Gunsight Lake Approach is #1 at `0.7849`, and at least three of your top 5 are in {`trail-0086`, `trail-0168`, `trail-0100`, `trail-0091`, `trail-0080`, `trail-0186`, `trail-0196`, `trail-0141`}. Ranks 6-8 sit within 0.01 of rank 5, so a slightly different order is not a bug. If the target appears in its own list at `1.0000`, step 2 of this list is missing.

### Step 5: Rank the neighbors of trail-0003 and trail-0008

**Do:** run the program again with `trail-0003` as the argument, then again with `trail-0008`. The code does not change — only the target vector does.

**Check:** for `trail-0003`, Grotto Falls (`trail-0027`) is #1 at `0.7858`, and at least two of {`trail-0068`, `trail-0131`, `trail-0039`, `trail-0070`, `trail-0150`} appear. For `trail-0008`, any five scores in the 0.71-0.75 band pass. The failure on `trail-0008` is not noticing the band is low.

### Step 6: Read the three lists and decide whether to ship

**Do:** no code here — read the `trail-0117`, `trail-0003`, and `trail-0008` lists the way a hiker would, looking at `difficulty` and `park` as you go.

**Check:** you named Alum Cave (`trail-0010`) at #4 for `trail-0117` (Smokies, no lake, matched on prose texture) and that four of the five are rated hard while the target is a moderate family hike. You named Coalpits Wash (`trail-0048`) at #4 for `trail-0003` (desert, no trees). You said the `trail-0008` list has no real neighbor at all. Shipping any list as-is fails; a floor (say 0.74), a metadata filter, or fewer results passes.

### Stretch goals

Pick any of these. The gear one is already built in `complete/`. The other two are not. `expected-output.md` has the numbers for all three.

- **Recommend gear from review text.** Products have no descriptions, so a product's vector comes from its reviews instead. Open `../../data/gear-reviews.jsonl`: 300 reviews over 25 products, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`), 16 of them for the Cascade 65 Backpack. Group `text` by `product`, join each product's reviews with newlines into one string, embed those 25 strings the same way as step 2, and cache to `gear-embeddings.json` keyed by product name. Find the product whose name contains the query. Rank the other 24 by cosine, print the top 5. Every track's `complete/` takes a `--gear` flag: `dotnet run -- --gear Cascade 65`, `uv run main.py --gear Cascade 65`, or `npm run complete -- --gear Cascade 65`. **Check:** Cascade 40 Daypack is #1 at `0.8039` — the same pack in a smaller size, the one product a Cascade 65 owner will never buy. The Summit Bear Canister, which three reviewers mention alongside the Cascade 65, is absent. Content similarity finds substitutes; complements need behavior data.
- **Average two trails.** Add the vectors for `trail-0117` and `trail-0003` position by position, divide each position by 2, rank every other trail against that average, and leave both source trails out. **Check:** Carlon Falls (`trail-0068`) is #1 at `0.7947` and every score is higher than before. That rise is a property of averaging vectors, not a better result.
- **Filter by difficulty.** Run `trail-0117` again, this time dropping every trail whose `difficulty` is not `easy` or `moderate` before you sort. **Check:** Alum Cave (`trail-0010`) is #1 at `0.7642` and Fern Lake (`trail-0196`) is #2 at `0.7508`. The filter removes what a family cannot do; it cannot invent good results, because this slice has almost no easy lake hikes.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F06-dotnet.md`](dotnet/F06-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F06-python.md`](python/F06-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F06-typescript.md`](typescript/F06-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/trails.json`: a 30-trail slice of the full catalog (feature 10's `data/trails.json`), same shape as the full file. It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting. Not the same 30 as feature 04's slice.
- `data/trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text`, keyed by trail id (768 floats each). Here so you can start cold without calling the model.
- `data/gear-reviews.jsonl`: 300 reviews across 25 products, one JSON object per line, from the shared Trailhead Guides corpus, for the gear stretch goal.
- `complete/embeddings.json` and `complete/gear-embeddings.json` (inside each track's `complete/`): vector caches the finished program writes on its first run. Delete them to re-embed.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
