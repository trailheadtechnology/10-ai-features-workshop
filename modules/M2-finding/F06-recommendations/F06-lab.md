# Lab 06: Recommendations

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `data/trails.json` (30 trails including the three targets), `data/trail-embeddings.json` (their `nomic-embed-text` vectors by id; or reuse your feature 04 vectors), `data/gear-reviews.jsonl` (300 reviews across 25 products, for the gear request).
- **How:** Ollama's embed endpoint, only for fresh vectors. `http/ollama.http`: 1 one description, 2 a three-description batch, 3 a gear review blob. Ranking is yours: cosine against every other trail, descending, target excluded, top 5. .NET `complete/`: `dotnet run -- <trail id or name>`.
- **Model:** `nomic-embed-text`, local. No chat model, no key.

### Step 1: Rank the neighbors of trail-0117

Send request 2 or load `data/trail-embeddings.json`, then rank every other trail by cosine against `trail-0117` and print the top 5 with scores (`complete/`: `dotnet run`); request 2's inputs, in order `trail-0117`, `trail-0086`, `trail-0041`:

```text
After a footbridge crossing near the mouth of Avalanche Gorge, where the creek churns through sculpted blue-green stone, the trail climbs at a friendly grade through hemlock and cedar. It ends at a lake walled by two-thousand-foot cliffs streaked with ribbon waterfalls fed by Sperry Glacier. Arrive early; the small trailhead lot fills before nine all summer.
```

```text
Dropping first to the St. Mary River, the trail runs a long valley floor of avalanche paths and cow parsnip before rising to a lake cupped beneath Gunsight Pass. A suspension bridge at the outlet sways with every step, to some hikers' delight and others' dread. Grizzlies frequent the brushy sections; make noise where sightlines shrink.
```

```text
A sandy wash curls beneath the crosshatched face of Checkerboard Mesa on the park's quieter east side. The grade is barely perceptible, making this a fine leg-stretcher between viewpoints. Bighorn sheep pick their way across the slickrock above with some regularity.
```

**Check:** `trail-0117` vs `trail-0086` is 0.7849, vs `trail-0041` is 0.6117, and at least three of your top 5 for `trail-0117` are in {`trail-0086`, `trail-0168`, `trail-0100`, `trail-0091`, `trail-0080`, `trail-0186`, `trail-0196`, `trail-0141`}. Equal numbers or anything above 0.99 means you compared a vector to itself.

### Step 2: Rank the neighbors of trail-0003 and trail-0008

`dotnet run -- trail-0003` and `dotnet run -- trail-0008`, or the same ranking with those ids.

**Check:** for `trail-0003`, Grotto Falls (`trail-0027`) is #1 at 0.7858 and at least two of {`trail-0068`, `trail-0131`, `trail-0039`, `trail-0070`, `trail-0150`} appear; for `trail-0008`, any five scores in the 0.71 to 0.75 band pass. The failure on `trail-0008` is not noticing the band is low.

### Step 3: Read the three lists and the Cascade 65 gear list, then decide whether to ship

No code. Read the `trail-0117`, `trail-0003`, and `trail-0008` lists as a hiker would, then run request 3, or `dotnet run -- --gear Cascade 65`.

**Check:** you named Alum Cave (`trail-0010`) at #4 for `trail-0117` (Smokies, no lake, rated hard for a moderate family hike), Coalpits Wash (`trail-0048`) at #4 for `trail-0003` (desert, no trees), and the Cascade 65's #1 as the same pack in a smaller size with the Summit Bear Canister absent. Shipping any list as-is fails; a floor (say 0.74), a metadata filter, or fewer results passes.

### Stretch goal: average trail-0117 with trail-0003, or filter trail-0117 by difficulty

Average the vectors for `trail-0117` and `trail-0003` and rank against the average, or re-run `trail-0117` keeping only trails rated easy or moderate; neither has a flag in `complete/`. **Check:** the average puts Carlon Falls first at 0.7947 with every score up, a property of averaging, not a better result; the filter puts Alum Cave first at 0.7642; it removes what a family cannot do and cannot invent good results.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F06-http.md`](http/F06-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F06-dotnet.md`](dotnet/F06-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F06-python.md`](python/F06-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F06-typescript.md`](typescript/F06-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/trails.json`: a 30-trail slice of the full catalog (feature 10's `data/trails.json`), same shape as the full file (id, name, park, distance, elevation, difficulty, features, description). It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting.
- `data/trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text`, keyed by trail id (768 floats each). Reuse your feature 04 vectors if you have them; this file is here so you can start cold.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
