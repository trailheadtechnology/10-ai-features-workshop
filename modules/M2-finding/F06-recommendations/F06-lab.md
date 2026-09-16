# Lab 06: Recommendations

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `data/trails.json` (30 trails including the three targets), `data/trail-embeddings.json` (their `nomic-embed-text` vectors by id, so you can skip the embed call), `data/gear-reviews.jsonl` (300 reviews across 25 products, for the gear stretch goal).
- **How:** turn each trail description into a vector once with the embedding model. Then score every other trail by cosine against the target trail's own vector, sort high to low, leave the target out, and keep the top 5. `http/ollama.http` has the embed calls: 1 one description, 2 a three-description batch, 3 a gear review blob. The model only gives you vectors. The ranking is code you write.
- **Model:** `nomic-embed-text`, local. No chat model, no key.

Each step below is one thing to make the program do. Each code track's `starter/` picks trails at random. Edit it until it does all six steps. Look at `complete/` when you get stuck. Your track's walkthrough gives the run command (`dotnet run`, `uv run main.py`, or `npm run starter`) and the embeddings client it uses; the steps here only say what to send and what comes back. On the HTTP track, run the numbered request in `http/ollama.http` when a step names one, and write the ranking in any language you like. That file is three plain `POST` bodies against the local `nomic-embed-text` model, one per `###` block, with a comment above each saying what it is for; the code tracks send the same bodies through a client library.

### Step 0: Run the starter and look at the random list

Run `starter/` without changing it. It loads the catalog, finds Avalanche Lake Trail (`trail-0117`), and prints five other trails picked at random under "you might also like".

**Check:** five trails with no connection to Avalanche Lake. Run it again and the five change. That is the recommendation feature most apps ship today, and the rest of the lab replaces it.

### Step 1: Load the trails

1. Open `../../data/trails.json`. It is one JSON array of 30 objects. Each object has the fields `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (an array of strings), and `description`. These 30 are lifted verbatim from the workshop's full 200-trail catalog, feature 10's `data/trails.json`, and were chosen to hold the three target trails plus enough real neighbors and enough noise to make ranking interesting. It is a different 30 from feature 04's slice (the two share 7 trails), so feature 04's cached vectors will not cover it. The catalog is synthetic: real park names, invented trails and descriptions.
2. Parse the file into a list of trail objects. The starter already does this.
3. Keep the target lookup the starter already has. Take the command-line argument as the query. If there is no argument, use `trail-0117`. Pick the first trail whose `id` equals the query or whose `name` contains it, ignoring case.

**Check:** the list has 30 entries. The first `id` is `trail-0003`. Running with no argument prints `You liked: Avalanche Lake Trail (Glacier National Park)`.

### Step 2: Get a vector for every trail, once

1. Build a dictionary from each trail's `id` to its `description`. The description is the only text you embed. No other field goes in.
2. Embed all 30 descriptions in one batch, in id order, through your track's embeddings client with the model `nomic-embed-text`. Request 2 in `http/ollama.http` is this same call with three descriptions instead of 30.
3. The response has one vector per input, in the same order you sent them. Store each vector in a dictionary keyed by that trail's `id`.
4. Write the dictionary to `embeddings.json` next to your program. At the top of the program, check for that file. If it exists and has a key for every trail id, load it and skip the call. This is the file `complete/` writes on its first run (the copies checked in there are its output); delete it to force a fresh embed.

Shortcut if you do not want to call the model: load `../../data/trail-embeddings.json` instead. It is the same dictionary, already computed with the same model: one key per trail id, each holding the 768-float vector of that trail's `description` and nothing else. It was produced by running exactly the request 2 call in `http/ollama.http` over all 30 descriptions (the values are `complete/embeddings.json` rounded to six decimals), and it is the run that `expected-output.md`'s scores come from.

**Check:** 30 keys, each holding 768 numbers. `vectors["trail-0117"]` starts with 0.0091, 0.0802, -0.1689. The second run does not call Ollama at all.

### Step 3: Bring in cosine and prove it works

1. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`. Loop once over the 768 positions. Add up the dot product and each vector's squared length as you go. Then divide.
2. Compute the cosine between the vectors for `trail-0117` and `trail-0086`. Then compute it between `trail-0117` and `trail-0041`. Print both.

On the HTTP track, request 2 embeds exactly these three trails. Its inputs, in order `trail-0117`, `trail-0086`, `trail-0041`:

```text
After a footbridge crossing near the mouth of Avalanche Gorge, where the creek churns through sculpted blue-green stone, the trail climbs at a friendly grade through hemlock and cedar. It ends at a lake walled by two-thousand-foot cliffs streaked with ribbon waterfalls fed by Sperry Glacier. Arrive early; the small trailhead lot fills before nine all summer.
```

```text
Dropping first to the St. Mary River, the trail runs a long valley floor of avalanche paths and cow parsnip before rising to a lake cupped beneath Gunsight Pass. A suspension bridge at the outlet sways with every step, to some hikers' delight and others' dread. Grizzlies frequent the brushy sections; make noise where sightlines shrink.
```

```text
A sandy wash curls beneath the crosshatched face of Checkerboard Mesa on the park's quieter east side. The grade is barely perceptible, making this a fine leg-stretcher between viewpoints. Bighorn sheep pick their way across the slickrock above with some regularity.
```

**Check:** `trail-0117` vs `trail-0086` is 0.7849. `trail-0117` vs `trail-0041` is 0.6117. Two Glacier lake hikes score high and a Zion desert wash scores low. Equal numbers, or anything above 0.99, means you compared a vector to itself.

### Step 4: Rank the neighbors of trail-0117

1. Take the target trail's vector from the dictionary.
2. Loop over every trail in the list except the target itself. For each one, compute the cosine between the target's vector and that trail's vector.
3. Sort the trails by that score, highest first. Keep the top 5.
4. Replace the random loop's output with the top 5. For each one, print the score to four decimals, the `name`, the `park`, the `difficulty`, and the `features` joined with commas.

**Check:** `trail-0086` Gunsight Lake Approach is #1 at 0.7849, and at least three of your top 5 are in {`trail-0086`, `trail-0168`, `trail-0100`, `trail-0091`, `trail-0080`, `trail-0186`, `trail-0196`, `trail-0141`}. Ranks 6 through 8 sit within 0.01 of rank 5, so a slightly different order is not a bug. If the target appears in its own list at 1.0000, step 2 of this list is missing.

### Step 5: Rank the neighbors of trail-0003 and trail-0008

1. Run the program again with `trail-0003` as the argument.
2. Run it again with `trail-0008` as the argument. The code does not change. Only the target vector does.

**Check:** for `trail-0003`, Grotto Falls (`trail-0027`) is #1 at 0.7858, and at least two of {`trail-0068`, `trail-0131`, `trail-0039`, `trail-0070`, `trail-0150`} appear. For `trail-0008`, any five scores in the 0.71 to 0.75 band pass. The failure on `trail-0008` is not noticing the band is low.

### Step 6: Read the three lists and decide whether to ship

No code here. Read the `trail-0117`, `trail-0003`, and `trail-0008` lists the way a hiker would. Look at the `difficulty` and `park` columns while you read.

**Check:** you named Alum Cave (`trail-0010`) at #4 for `trail-0117` (Smokies, no lake, matched on prose texture) and that four of the five are rated hard while the target is a moderate family hike. You named Coalpits Wash (`trail-0048`) at #4 for `trail-0003` (desert, no trees). You said the `trail-0008` list has no real neighbor at all. Shipping any list as-is fails; a floor (say 0.74), a metadata filter, or fewer results passes.

### Stretch goals

Pick any of these. The gear one is already built in `complete/`. The other two are not. `expected-output.md` has the numbers for all three.

- **Recommend gear from review text.** Products have no descriptions. So a product's vector comes from its reviews instead. Open `../../data/gear-reviews.jsonl`. Every line is one JSON object with the fields `id`, `product`, `rating`, `reviewer`, `text`, for example `{"id": "gr-0002", "product": "Granite Peak 4P Tent", "rating": 2, "reviewer": "Trailname_Pockets", "text": "Returned it. ..."}`. It is the gear-review set of the workshop's synthetic Trailhead Guides corpus, shipped with the workshop rather than built by any script here: 300 reviews over 25 products, 16 of them for the Cascade 65 Backpack, and the ratings do not always agree with the text, so read a few before you trust either. Read it line by line. Parse each line. Group the `text` values by `product`. Join each product's reviews with newlines into one string. Now you have 25 strings keyed by product name. Embed those 25 strings the same way as step 2, and cache them to `gear-embeddings.json` next to your program, keyed by product name, the same way step 2 caches the trail vectors (`complete/` writes this file on its first `--gear` run; request 3 in `http/ollama.http` is this call for three of the Cascade 65's 16 reviews). Find the product whose name contains the query. Rank the other 24 by cosine. Print the top 5. Every track's `complete/` takes the `--gear` flag; run whichever is yours: `dotnet run -- --gear Cascade 65`, `uv run main.py --gear Cascade 65`, or `npm run complete -- --gear Cascade 65`. **Check:** Cascade 40 Daypack is #1 at 0.8039. It is the same pack in a smaller size, the one product a Cascade 65 owner will never buy. The Summit Bear Canister, which three reviewers mention alongside the Cascade 65, is absent. Content similarity finds substitutes; complements need behavior data.
- **Average two trails.** Add the vectors for `trail-0117` and `trail-0003` position by position. Divide each position by 2. Rank every other trail against that average. Leave both source trails out. **Check:** Carlon Falls (`trail-0068`) is #1 at 0.7947 and every score is higher than before. That rise is a property of averaging vectors, not a better result.
- **Filter by difficulty.** Run `trail-0117` again. This time, drop every trail whose `difficulty` is not `easy` or `moderate` before you sort. **Check:** Alum Cave (`trail-0010`) is #1 at 0.7642 and Fern Lake (`trail-0196`) is #2 at 0.7508. The filter removes what a family cannot do. It cannot invent good results, because this slice has almost no easy lake hikes.

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

- `data/trails.json`: a 30-trail slice of the full catalog (feature 10's `data/trails.json`), same shape as the full file (id, name, park, distance, elevation, difficulty, features, description). It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting. Not the same 30 as feature 04's slice.
- `data/trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text` from the requests in `http/ollama.http`, keyed by trail id (768 floats each). This file is here so you can start cold without calling the model.
- `data/gear-reviews.jsonl`: 300 reviews across 25 products, one JSON object per line, from the shared Trailhead Guides corpus, for the gear stretch goal.
- `http/ollama.http`: the three embed requests the steps point at, runnable as-is from an editor's REST client.
- `complete/embeddings.json` and `complete/gear-embeddings.json` (inside each code track's `complete/`): vector caches the finished program writes on its first run. Delete them to re-embed.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
