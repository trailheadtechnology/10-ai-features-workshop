<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 06: Recommendations (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F06-dotnet.md), [TypeScript](../typescript/F06-typescript.md). Lab overview: [F06-lab.md](../F06-lab.md).*

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

Two scripts, both reading from [`data/`](../data/). `starter/main.py` is the "you might also like" box, picking five trails at random. `complete/main.py` is the finished demo as shown on stage: feature 04's embedding code over this feature's own 30-trail slice (vectors cached to `embeddings.json`), "more like this" as nearest neighbors of one item's vector, and `--gear` for the same trick over product reviews.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [`SETUP.md`](../../../../SETUP.md)), and `uv run` finds it from any folder, with no venv to activate. Edit `starter/main.py` in place (or copy it first) and run `uv run main.py` from the `starter/` folder. `starter/main.py` takes no flags, at most the one positional argument its header comment names; the flags below are the ones `complete/` supports, so add the same argument parsing or hard-code the value. From `complete/`:

```bash
uv run main.py                        # "more like this" for Avalanche Lake Trail
uv run main.py trail-0008             # any trail id works
uv run main.py Trail of the Cedars    # so does any name (or part of one)
uv run main.py --gear Cascade 65      # the same trick on gear, from review text
```

Real output for all four commands, including the neighbors that are obviously wrong, is in [`expected-output.md`](../expected-output.md).

### Step 0: Run the starter and look at the random list

**Do:** run `starter/` without changing it. It loads the catalog, finds Avalanche Lake Trail (`trail-0117`), and prints five other trails picked at random under "you might also like".

```bash
uv run main.py
```

**Check:** the five trails have no connection to Avalanche Lake, and a second run gives you a different five. The starter never looks at the trail you liked; it just draws five of the other 29. A hiker who liked a moderate 4.6-mile lake walk is as likely to be handed a hard desert route as another lake.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails.json`: 30 objects with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description`.
2. Review how the existing code from the starter project parses the file into a list of trail objects.

   It happens near the top of `starter/main.py`. `json.loads` turns the file's text into a Python list, and each trail in it is a dict, so you read a field with `t["name"]`, not `t.name`:

   ```python
   DATA = Path(__file__).resolve().parents[2] / "data"
   trails = json.loads((DATA / "trails.json").read_text())
   ```

   For the Check, put two temporary prints right below the `trails` line and delete them once the numbers match:

   ```python
   # Hint: two throwaway prints
   print(len(trails))       # how many trails
   print(trails[0]["id"])   # id of the first one
   ```

3. Review how it looks up the target: it takes the command-line argument as the query, defaults to `trail-0117`, and picks the first trail whose `id` equals the query or whose `name` contains it (case-insensitive).

   It happens right below the `trails` line. `sys.argv[1:]` holds the words after `uv run main.py`, and `next(...)` returns the first trail that matches, or `None`:

   ```python
   query = " ".join(sys.argv[1:]) or "trail-0117"
   target = next((t for t in trails if t["id"].lower() == query.lower() or query.lower() in t["name"].lower()), None)
   if target is None:
       raise SystemExit(f"No trail matches '{query}'.")
   ```

   The Check's line comes from this `print`, a few lines further down:

   ```python
   print(f"You liked: {target['name']} ({target['park']})")
   ```

**Why:** these 30 are lifted from the workshop's full 200-trail catalog (feature 10's `data/trails.json`), chosen to hold the three target trails plus enough real neighbors and enough noise to make ranking interesting. It's a different 30 from feature 04's slice (7 trails overlap), so feature 04's cached vectors won't cover it.

**Check:** the list has 30 entries. The first `id` is `trail-0003`. Running with no argument prints `You liked: Avalanche Lake Trail (Glacier National Park)`.

### Step 2: Get a vector for every trail, once

**Do:**
1. Build a dictionary from each trail's `id` to its `description`. The description is the only text you embed.

   A dict comprehension builds a dictionary in one expression: `{key: value for item in list}`. You pass it straight to the helper in item 4, so there is nothing to keep yet. To see what it makes, try it as a throwaway print right below the `trails` line:

   ```python
   # Hint: one throwaway print, delete it after
   print({t["id"]: t["description"] for t in trails}["trail-0117"])
   ```

2. Embed all 30 descriptions in **one batch**, in id order. The embedding client and call are the same as feature 04.

   Option A, embed live (option B, loading the precomputed file instead, is at the end of item 4). The client is the `openai` package pointed at Ollama. Add this import below `from pathlib import Path`, with a blank line between them:

   ```python
   from openai import OpenAI
   ```

   Then add these lines right below the `DATA = ...` line. Keep `import random` for now, because the random loop uses it until step 4:

   ```python
   client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
   EMBED_MODEL = "nomic-embed-text"
   HERE = Path(__file__).resolve().parent
   ```

   The call sends a list of strings in one request and gets back one result per string, in the same order. `[texts[k] for k in keys]` turns the list of ids into the list of their descriptions. This line lives inside the helper in item 4, so do not type it on its own:

   ```python
   response = client.embeddings.create(model=EMBED_MODEL, input=[texts[k] for k in keys])
   ```

3. Store each returned vector in a dictionary keyed by trail `id`.

   `response.data` is a list with one entry per input, and each entry's `.embedding` is that text's list of floats. `zip` walks the ids and the results side by side, pairing the first id with the first vector. This line also lives inside the helper in item 4:

   ```python
   vectors = {k: d.embedding for k, d in zip(keys, response.data)}
   ```

4. Write the dictionary to `embeddings.json` next to your program. At startup, check for that file and skip the call if it exists and has a key for every trail id. Shape it as a plain id-to-float-array dictionary so the same load code reads either your cache or `../../data/trail-embeddings.json`.

   Here is the whole helper. If `embeddings.json` exists and holds every id, it returns what is in the file. Otherwise it makes the call from item 2, builds the dictionary from item 3, and writes it to the file with `json.dumps`. Put it below the `HERE = ...` line and above `trails = ...`, with a blank line on each side (a function has to be defined above the line that calls it):

   ```python
   def embed_with_cache(cache_name: str, texts: dict[str, str]) -> dict[str, list[float]]:
       cache_path = HERE / cache_name
       if cache_path.exists():
           cached = json.loads(cache_path.read_text())
           if all(k in cached for k in texts):
               return cached
       keys = list(texts)
       response = client.embeddings.create(model=EMBED_MODEL, input=[texts[k] for k in keys])
       vectors = {k: d.embedding for k, d in zip(keys, response.data)}
       cache_path.write_text(json.dumps(vectors))
       return vectors
   ```

   Then add this line right after `trails = ...`. It builds item 1's dictionary and passes it in:

   ```python
   vectors = embed_with_cache("embeddings.json", {t["id"]: t["description"] for t in trails})
   ```

   `HERE` is the folder `main.py` sits in, so `embeddings.json` lands in `starter/`. Delete it to re-embed.

   Option B, precomputed. Same dictionary shape, no model call. Use this one line instead of all of option A (no import, client, or helper), right after `trails = ...`:

   ```python
   vectors = json.loads((DATA / "trail-embeddings.json").read_text())
   ```

   For the Check, put temporary prints right after the `vectors` line, fill in `trail-0117`, and delete them once the numbers match. Then run a second time: the helper returns from the file without calling Ollama, so the run is instant and `ls -l embeddings.json` shows the same time as after the first run:

   ```python
   # Hint: three throwaway prints
   print(len(vectors))                 # how many keys
   print(len(vectors["trail-XXXX"]))   # numbers per vector
   print(vectors["trail-XXXX"][:3])    # first three numbers
   ```

**Why:** shortcut if you'd rather not call the model: load `../../data/trail-embeddings.json` instead. It is the same dictionary, already computed with the same model, and the run that `expected-output.md`'s scores come from.

**Check:** 30 keys, each holding 768 numbers. `vectors["trail-0117"]` starts with `0.0091, 0.0802, -0.1689`. The second run does not call Ollama at all.

### Step 3: Bring in cosine and prove it works

**Do:**
1. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`.

   Add this import at the top, next to `import json`:

   ```python
   import math
   ```

   Then add the function below `embed_with_cache` (or below the `DATA = ...` line if you used option B) and above `trails = ...`, with a blank line on each side. `zip(a, b)` pairs position 0 of `a` with position 0 of `b`, and so on, and `sum(...)` adds up the products:

   ```python
   def cosine(a: list[float], b: list[float]) -> float:
       dot = sum(x * y for x, y in zip(a, b))
       return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))
   ```

2. Compute the cosine between the vectors for `trail-0117` and `trail-0086`, then between `trail-0117` and `trail-0041`. Print both.

   Put the two prints right after the `vectors = ...` line. `:.4f` prints four decimals:

   ```python
   print(f"{cosine(vectors['trail-0117'], vectors['trail-0086']):.4f}")
   print(f"{cosine(vectors['trail-0117'], vectors['trail-0041']):.4f}")
   ```

   Delete the two prints once you have seen them.

**Check:** `trail-0117` vs `trail-0086` is `0.7849`. `trail-0117` vs `trail-0041` is `0.6117`. Two Glacier lake hikes score high and a Zion desert wash scores low. Equal numbers, or anything above 0.99, means you compared a vector to itself.

### Step 4: Rank the neighbors of trail-0117

**Do:**
1. Take the target trail's vector.

   `target` is a dict, so its id is `target["id"]` and its vector is `vectors[target["id"]]`. That expression goes straight into item 2's code, so there is no line to add yet.

2. Loop over every trail except the target, computing the cosine between the target's vector and each.

   Python can build a list in one line: `[value for t in trails if condition]`. Here each value is a `(score, trail)` pair, and the `if` skips the target. Try it as a throwaway print right below the `raise SystemExit(...)` line (after `target` is found, since the print uses it), then delete it:

   ```python
   # Hint: 30 trails minus the target
   print(len([(cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"]]))  # should print 29
   ```

3. Sort by score, highest first. Keep the top 5.

   `sorted(...)` takes those same pairs (written with round brackets instead of square ones, which works the same inside a call). `key=lambda h: -h[0]` tells it to sort each pair `h` by its score `h[0]`, and the minus sign puts the highest score first. `[:5]` keeps the first five. Delete these lines from the starter:

   ```python
   others = [t for t in trails if t["id"] != target["id"]]
   for t in random.sample(others, 5):
       print(f"  {t['name']} ({t['park']}, {t['difficulty']})")
   ```

   and put this line in their place, at the bottom of the file:

   ```python
   hits = sorted(((cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"]), key=lambda h: -h[0])[:5]
   ```

4. Replace the random loop's output with one line per hit: score to four decimals, then `name (park, difficulty; features joined with commas)`. `expected-output.md` shows the id instead of the features; the scores and order are what to match, not the exact line shape.

   Add the loop below the `hits` line. Each item in `hits` is a `(score, trail)` pair, so `for score, trail in hits` unpacks both. `', '.join(...)` glues the features list into one string:

   ```python
   for score, trail in hits:
       print(f"  {score:.4f}  {trail['name']} ({trail['park']}, {trail['difficulty']}; {', '.join(trail['features'])})")
   ```

   Then change the heading line above it to drop "(picked at random, which is the current feature)", so it reads:

   ```python
   print("You might also like:\n")
   ```

   Nothing uses `random` any more, so delete `import random` from the top. Then run:

   ```bash
   uv run main.py
   ```

**Why:** "more like this" is feature 04's search with the query vector replaced by the target trail's own vector. Difficulty is not in the description text, so the embedding cannot see it.

**Check:** `trail-0086` Gunsight Lake Approach is #1 at `0.7849`, and at least three of your top 5 are in {`trail-0086`, `trail-0168`, `trail-0100`, `trail-0091`, `trail-0080`, `trail-0186`, `trail-0196`, `trail-0141`}. Ranks 6-8 sit within 0.01 of rank 5, so a slightly different order is not a bug. If the target appears in its own list at `1.0000`, step 2 of this list is missing.

### Step 5: Rank the neighbors of trail-0003 and trail-0008

**Do:** run the program again with `trail-0003` as the argument, then again with `trail-0008`. The code does not change. Only the target vector does.

```bash
uv run main.py trail-0003
uv run main.py trail-0008
```

**Check:** these numbers are from the shipped vectors; embedding live gives you scores a few thousandths off and can reorder near-ties. For `trail-0003`, Grotto Falls (`trail-0027`) is #1 at `0.7858`, and at least two of {`trail-0068`, `trail-0131`, `trail-0039`, `trail-0070`, `trail-0150`} appear. For `trail-0008`, any five scores in the 0.71-0.75 band pass. The failure on `trail-0008` is not noticing the band is low.

### Step 6: Read the three lists and decide whether to ship

**Do:** no code here. Read the `trail-0117`, `trail-0003`, and `trail-0008` lists the way a hiker would, looking at `difficulty` and `park` as you go.

**Check:** you named Alum Cave (`trail-0010`) at #4 for `trail-0117` (Smokies, no lake, matched on prose texture) and that four of the five are rated hard while the target is a moderate family hike. You named Coalpits Wash (`trail-0048`) at #4 for `trail-0003` (desert, no trees). You said the `trail-0008` list has no real neighbor at all. Shipping any list as-is fails; a floor (say 0.74), a metadata filter, or fewer results passes.

### Stretch goals

Pick any of these. The gear one is already built in `complete/`. The other two are not. `expected-output.md` has the numbers for all three.

- **Recommend gear from review text.** Products have no descriptions, so a product's vector comes from its reviews instead. Open `../../data/gear-reviews.jsonl`: 300 reviews over 25 products, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`), 16 of them for the Cascade 65 Backpack. Group `text` by `product`, join each product's reviews with newlines into one string, embed those 25 strings the same way as step 2, and cache to `gear-embeddings.json` keyed by product name. Find the product whose name contains the query. Rank the other 24 by cosine, print the top 5. **Check:** Cascade 40 Daypack is #1 at `0.8039`, the same pack in a smaller size and the one product a Cascade 65 owner will never buy. The Summit Bear Canister, which three reviewers mention alongside the Cascade 65, is absent. Content similarity finds substitutes; complements need behavior data.

  `complete/` takes a `--gear` flag and embeds each product's joined review text through the same `embed_with_cache` helper (cache `gear-embeddings.json`). From `complete/`:

  ```bash
  uv run main.py --gear Cascade 65
  ```

  To build it yourself (with option A from step 2, since the helper does the embedding), add this import at the top, below `import sys`:

  ```python
  from collections import defaultdict
  ```

  `defaultdict(list)` is a dictionary that starts every new key with an empty list, so you can `.append` to a product without checking whether it is there yet. Start a function below `cosine` and above `trails = ...`. JSON Lines means one JSON object per line, so split the file into lines and `json.loads` each one (the `if` skips blank lines). The last line joins each product's reviews with newlines into one string, the same product-to-text dictionary shape `embed_with_cache` takes:

  ```python
  def recommend_gear(query: str) -> None:
      review_text: dict[str, list[str]] = defaultdict(list)
      for line in (DATA / "gear-reviews.jsonl").read_text().splitlines():
          if line.strip():
              r = json.loads(line)
              review_text[r["product"]].append(r["text"])
      texts = {product: "\n".join(reviews) for product, reviews in review_text.items()}

      vectors = embed_with_cache("gear-embeddings.json", texts)
  ```

  Still inside the function, find the product whose name contains the query, the same `next(...)` lookup the starter uses for trails:

  ```python
      target = next((p for p in texts if query.lower() in p.lower()), None)
      if target is None:
          raise SystemExit(f"No product matches '{query}'.")
  ```

  Finish the function with the ranking. `vectors.items()` gives `(product, vector)` pairs. The pairs start with the score, so `reverse=True` sorts highest first without a `key`:

  ```python
      print(f"You bought: {target}")
      print("Goes well with:\n")
      hits = sorted(((cosine(vectors[target], v), p) for p, v in vectors.items() if p != target), reverse=True)[:5]
      for score, product in hits:
          print(f"  {score:.4f}  {product}")
  ```

  Finally, check for the flag right above `trails = ...` (below the functions), so a gear run stops before the trail code. `raise SystemExit` ends the program:

  ```python
  args = sys.argv[1:]
  if args and args[0] == "--gear":
      recommend_gear(" ".join(args[1:]))
      raise SystemExit
  ```

- **Average two trails.** Add the vectors for `trail-0117` and `trail-0003` position by position, divide each position by 2, rank every other trail against that average, and leave both source trails out. **Check:** Carlon Falls (`trail-0068`) is #1 at `0.7947` and every score is higher than before. That rise is a property of averaging vectors, not a better result.

  Replace step 4's `hits` line with this. There is no averaging function to call; the list comprehension builds a new list one position at a time, and `zip` pairs position 0 of `a` with position 0 of `b`:

  ```python
  # Hint: average two vectors, then rank against the average
  a = vectors["trail-AAAA"]
  b = vectors["trail-BBBB"]
  average = [(x + y) / 2 for x, y in zip(a, b)]
  hits = sorted(((cosine(average, vectors[t["id"]]), t) for t in trails if t["id"] not in ("trail-AAAA", "trail-BBBB")), key=lambda h: -h[0])[:5]
  ```

  Put step 4's line back when you are done.

- **Filter by difficulty.** Run `trail-0117` again, this time dropping every trail whose `difficulty` is not `easy` or `moderate` before you sort. **Check:** Alum Cave (`trail-0010`) is #1 at `0.7642` and Fern Lake (`trail-0196`) is #2 at `0.7508`. The filter removes what a family cannot do; it cannot invent good results, because this slice has almost no easy lake hikes.

  Add a second condition to the `if` inside step 4's `hits` line, joined with `and`. `difficulty` holds lowercase strings (`easy`, `moderate`, `hard`), and `in (...)` checks for either value:

  ```python
  # Hint: step 4's hits line with one extra condition
  hits = sorted(((cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"] and t["difficulty"] in ("VALUE1", "VALUE2")), key=lambda h: -h[0])[:5]
  ```

## What Is in This Folder

- `data/trails.json`: a 30-trail slice of the full catalog (feature 10's `data/trails.json`), same shape as the full file. It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting. Not the same 30 as feature 04's slice.
- `data/trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text`, keyed by trail id (768 floats each). Here so you can start cold without calling the model.
- `data/gear-reviews.jsonl`: 300 reviews across 25 products, one JSON object per line, from the shared Trailhead Guides corpus, for the gear stretch goal.
- `complete/embeddings.json` and `complete/gear-embeddings.json` (inside each track's `complete/`): vector caches the finished program writes on its first run. Delete them to re-embed.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
