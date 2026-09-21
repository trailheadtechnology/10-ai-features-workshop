<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 04: Semantic Search (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F04-dotnet.md), [TypeScript](../typescript/F04-typescript.md). Lab overview: [F04-lab.md](../F04-lab.md).*

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

Two scripts, both reading from [`data/`](../data/). `starter/main.py` is the keyword search. `complete/main.py` is the finished demo as shown on stage: `nomic-embed-text` through the SDK's embeddings call, the 30 descriptions embedded once and cached, cosine similarity in one visible function, top 5 with scores. The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else.

The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [SETUP.md](../../../../SETUP.md)), and `uv run` finds it from any folder. No venv to activate. Run from the `starter/` folder; the query is the positional arguments joined with spaces, and `complete/` takes the same arguments:

```bash
uv run main.py                                   # dog-friendly waterfall hike, not too steep
uv run main.py somewhere quiet to take my kids
uv run main.py an easy hike to a great view
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
uv run main.py
uv run main.py somewhere quiet to take my kids
```

The starter already does the word splitting with this line in `starter/main.py`. `re.findall` pulls out runs of letters, `if len(w) >= 3` keeps words of three letters or more, and `dict.fromkeys` drops repeats while keeping their order. The `for t in trails:` loop right below it counts the hits.

```python
tokens = list(dict.fromkeys(w for w in re.findall(r"[a-z]+", query.lower()) if len(w) >= 3))
```

**Check:** the first query puts `trail-0007` Upper Yosemite Falls Trail first, a hard 7.2-mile climb that matched on `waterfall` and `too`. The second query returns exactly one trail, `trail-0004` Ocean Path, because its description happens to contain the word `kids`. Both are in [`expected-output.md`](../expected-output.md)'s keyword blocks. The rest of the lab replaces that scoring with embeddings.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails-slice.json` (that path is relative to your track's `starter/` folder): one JSON array of 30 objects, each with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description` (a three- or four-sentence blurb).
2. Review how the existing code from the starter project parses it into a list of trail objects.

   `Path(__file__).resolve().parents[2]` is the folder two levels above the one `main.py` sits in, so `DATA` points at `../../data/` no matter which folder you run from. `json.loads` turns the file's text into a Python list of dictionaries, one per trail. It happens near the top of `main.py`. Keep both lines:

   ```python
   DATA = Path(__file__).resolve().parents[2] / "data"
   ```

   ```python
   trails = json.loads((DATA / "trails-slice.json").read_text())
   ```

**Why:** the 30 were hand-picked from feature 10's 200-trail catalog to make keyword matching fail on purpose: four dog-friendly waterfall trails whose descriptions never say so in those words (`trail-0011`, `trail-0027`, `trail-0055`, `trail-0068`), plus three keyword traps. `trail-0074` Easy Creek Trail is a hard 2,610-foot climb named after homesteader Elias Easy, `trail-0187` Dog Lake Trail prohibits pets, and `trail-0058` Panorama Cliffs Bypass contains the word "steep" while describing how it avoids the steep sections.

**Check:** the list has 30 entries. The first `id` is `trail-0003`, Trail of the Cedars.

`len(trails)` is 30 and `trails[0]["id"]` is `trail-0003`. To see it, put a temporary line right below the `trails = ...` line:

```python
# Hint: print the count and the first id, then delete this line
print(f"{len(trails)} trails, first is {trails[0]['id']}")
```

### Step 2: Embed the 30 descriptions once and cache the vectors

**Do:**
1. Collect the 30 `description` strings, in list order.

   First, in `starter/main.py`, comment out the keyword scoring: every line from `tokens = ...` down to the last `print(...)` at the bottom of the file. Select those lines and press Cmd+/ (Ctrl+/ on Windows) in VS Code, or put `#` in front of each. If you plan to try the keyword-blend stretch goal, comment it out rather than deleting it, because that goal reuses it. Keep the docstring, the imports (including `import re`), `DATA`, `query`, and `trails`. All the new code in this lab goes at the bottom of the file, below the commented-out lines, unless an item says otherwise.

   You do not need a separate loop to collect the strings. This list comprehension, used in item 2, builds the list of descriptions in list order:

   ```python
   [t["description"] for t in trails]
   ```

2. Embed all 30 in **one call** to `nomic-embed-text` (a single batch, not a loop). The response holds one 768-float vector per input, in the order sent.

   The client is the `openai` package's `OpenAI` class pointed at Ollama's OpenAI-compatible endpoint. The `openai` package is the one dependency in the repo root `pyproject.toml`, so there is nothing to install, and Ollama ignores the `api_key`, which just has to be non-empty. Add the import at the top of `main.py`, below `from pathlib import Path`:

   ```python
   from openai import OpenAI
   ```

   Create the client and name the model right below the imports, above the `DATA = ...` line:

   ```python
   client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
   EMBED_MODEL = "nomic-embed-text"
   ```

   `client.embeddings.create` takes the whole list as `input` in one call and returns one item per input in `response.data`, in the order sent. This helper keeps just the vectors. Put it at the bottom of the file; a function has to be defined above the code that calls it:

   ```python
   def embed(texts: list[str]) -> list[list[float]]:
       response = client.embeddings.create(model=EMBED_MODEL, input=texts)
       return [d.embedding for d in response.data]
   ```

   Then call it once with all 30 descriptions, below the function (item 4 moves this line inside an `else:`):

   ```python
   embeddings = embed([t["description"] for t in trails])
   ```

3. Walk the list and the response together and store each vector in a dictionary keyed by that trail's `id`. Store each vector as a plain array of floats, not the client's own vector type, so it serializes and indexes without surprises.

   `zip(trails, embeddings)` pairs each trail with its vector: `t` is the trail, `e` is its vector. The `{key: value for ...}` form is a dictionary comprehension, which builds the dictionary in one line. Each `d.embedding` is already a plain Python list of 768 floats, so nothing needs converting. Put this directly under the `embeddings = ...` line:

   ```python
   vectors = {t["id"]: e for t, e in zip(trails, embeddings)}
   ```

4. Write the dictionary to `embeddings.json` (your track's block below says which folder it lands in). That is the file to delete when you want to re-embed. At program start, load it and skip the embed call if it exists. Wrap the embed call in a timer and print how many vectors were embedded and how many milliseconds it took, or how many were loaded from the cache.

   `Path(__file__).with_name("embeddings.json")` is a file next to `main.py`, so the cache lands at `starter/embeddings.json`. `time.perf_counter()` returns a clock reading in seconds; subtract two readings and multiply by 1000 for milliseconds. Add the import at the top of `main.py`, with the other `import` lines:

   ```python
   import time
   ```

   Replace your `embeddings = ...` and `vectors = ...` lines (keep the `def embed` function above them) with this block. The `else:` branch holds those same two lines, now indented four spaces, with the timer around them and the file write after. `json.dumps` turns the dictionary into text and `json.loads` turns it back:

   ```python
   cache_path = Path(__file__).with_name("embeddings.json")
   if cache_path.exists():
       vectors = json.loads(cache_path.read_text())
       print(f"Loaded {len(vectors)} cached vectors from embeddings.json")
   else:
       started = time.perf_counter()
       embeddings = embed([t["description"] for t in trails])
       vectors = {t["id"]: e for t, e in zip(trails, embeddings)}
       cache_path.write_text(json.dumps(vectors))
       print(f"Embedded {len(vectors)} trail descriptions in {(time.perf_counter() - started) * 1000:.0f} ms")
   ```

**Why:** `embeddings.json` is a cache your program creates, not a shipped data file. It is gitignored so the first run always embeds live. It is keyed only by `id`, so delete it whenever a description or the model changes, or every later query ranks against vectors for text that no longer exists.

**Check:** 30 keys, each holding 768 floats. The first run is the slower one, since the model is embedding; the second is instant and prints that it loaded 30 cached vectors. Fewer than 30 vectors, or one not 768 long, means the batch did not go through.

`len(vectors)` is 30 and `len(vectors["trail-0003"])` is 768. Print one and look at it: it is just numbers. The first run prints `Embedded 30 trail descriptions in ... ms`; the second run prints `Loaded 30 cached vectors from embeddings.json`. Delete `starter/embeddings.json` if the text or the model changes. A temporary check, placed below the `else:` block (not indented):

```python
# Hint: count, length, and the first few numbers of one vector
print(f"{len(vectors)} vectors, {len(vectors['trail-0003'])} floats each")
print(vectors["trail-0003"][:8])
```

### Step 3: Embed query 1, write cosine similarity, print the top 5

**Do:**
1. Review how the existing code from the starter project takes the query from the command-line arguments, joined with spaces. It defaults to this when there are none:

   ```text
   dog-friendly waterfall hike, not too steep
   ```

   It happens near the top of `main.py`. `sys.argv[1:]` holds the words typed after `uv run main.py`; when there are none, the joined string is empty and `or` falls back to the default:

   ```python
   query = " ".join(sys.argv[1:]) or "dog-friendly waterfall hike, not too steep"
   ```

2. Embed the query as a single string, through the same client and same model. Keep the one vector that comes back.

   The query goes through the same `embed` function. `embed` takes a list and returns a list, so pass `[query]` (a one-item list) and take `[0]` to keep the one vector. Put this at the bottom of the file, below the cache `if`/`else:` block:

   ```python
   query_vector = embed([query])[0]
   ```

3. Write `cosine_similarity(a, b)`: loop `i` 0 to 767, accumulate `dot += a[i]*b[i]`, `magA += a[i]*a[i]`, `magB += b[i]*b[i]`, then return `dot / (sqrt(magA) * sqrt(magB))`.

   The version below is the index loop above written the Python way, and the loop works too. `zip(a, b)` walks both lists side by side, `sum(x * y for ...)` adds up the products, and `math.sqrt` is the square root. Add the import at the top of `main.py`, with the other `import` lines:

   ```python
   import math
   ```

   Put the function at the bottom of the file, below the `query_vector = ...` line:

   ```python
   def cosine_similarity(a: list[float], b: list[float]) -> float:
       dot = sum(x * y for x, y in zip(a, b))
       return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))
   ```

4. For every trail, compute the cosine between the query vector and that trail's stored vector.

   This generator expression makes one `(score, trail)` pair per trail. It is not a statement on its own; item 5 wraps it in `sorted(...)`:

   ```python
   (cosine_similarity(query_vector, vectors[t["id"]]), t) for t in trails
   ```

5. Sort by score, highest first, and keep the top 5.

   `sorted` returns a new sorted list. `key=lambda r: -r[0]` sorts each pair by its score (`r[0]`), and the minus sign puts the highest first. `[:5]` keeps the first 5. Put this below the `def cosine_similarity` function, with item 4's piece inside the double parentheses:

   ```python
   results = sorted(((cosine_similarity(query_vector, vectors[t["id"]]), t) for t in trails), key=lambda r: -r[0])[:5]
   ```

6. Print each of the 5 on one line: score to four decimals, `id`, `name`, `difficulty`, `distance_mi`, and the `features` list. The semantic blocks in [`expected-output.md`](../expected-output.md) show the layout.

   `for score, trail in results:` unpacks each pair into two variables. `{score:.4f}` formats the score to four decimals, and `', '.join(...)` turns the features list into one string. Put this at the bottom of the file, below the `results = ...` line:

   ```python
   print(f'\nSemantic search: "{query}"\n')
   for score, trail in results:
       print(f"{score:.4f}  {trail['id']}  {trail['name']} ({trail['difficulty']}, {trail['distance_mi']} mi)  [{', '.join(trail['features'])}]")
   ```

Run:

```bash
uv run main.py
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
uv run main.py somewhere quiet to take my kids
uv run main.py an easy hike to a great view
```

**Why:** these come from `data/queries.json`, which holds all three test queries with `id`, `query`, `why` (what that query proves), and `success_check`. Look at the scores themselves, not just the order. When the whole top 5 is low and close together, nothing in the catalog is a strong match, and the order is mostly noise.

**Check:** query 2 puts Taft Point (`trail-0020`) first at `0.4876` (correct, and still a cliff edge), with short family-friendly trails filling the rest of the top 5, all between 0.43 and 0.49; query 3 puts `trail-0058` Panorama Cliffs Bypass in the top 3 (`0.6481`, first when I ran it), top score `0.77` on query 1. `trail-0074` Easy Creek Trail in query 3's top 5 means you are matching words, not meaning.

### Stretch goals

Pick either. Each uses information you already have, trail metadata or the keyword hits, instead of hoping the vector carries it. The embedding fixed recall (finding the right trails); these fix precision (dropping the wrong ones).

- **Filter before you rank.** Between step 3's item 3 and item 4, drop trails whose `difficulty` is `hard`. For query 1, also drop trails whose `features` list does not contain `dog-friendly`. Then rank whatever is left. **Check:** query 1 returns only dog-friendly trails; query 3 loses Beehive Loop (`trail-0017`) and Chimney Tops (`trail-0005`) from the top 5. Either still there means the filter ran after the top 5 was taken.

  Build a filtered list with a list comprehension, then rank that list instead of `trails`, so the filter runs before `[:5]`. The query 1 text is the default string on the `query = ...` line. Put the first two lines of the hint right above the step 3 `results = ...` line, and replace that line with the third (the only change is `for t in candidates` in place of `for t in trails`):

  ```python
  # Hint: filter first, then score and rank what is left
  is_query1 = query == "<query 1 text>"
  candidates = [t for t in trails if t["difficulty"] != "<difficulty to drop>" and (not is_query1 or "<required feature>" in t["features"])]
  results = sorted(((cosine_similarity(query_vector, vectors[t["id"]]), t) for t in candidates), key=lambda r: -r[0])[:5]
  ```

- **Blend in the keyword score.** Keep the keyword-hit count from step 0 (the keyword code you commented out in step 2) alongside the cosine score. Rescale both to 0..1 with min-max: subtract the lowest score, then divide by the gap between the highest and lowest. Rank on a weighted sum whose two weights add up to 1, and print both scores on each row. Run query 3 at 0.7 cosine and 0.3 keyword first, then raise the cosine weight. **Check:** every row shows a cosine score and a keyword score. At 0.7 cosine, Easy Creek Trail (`trail-0074`) enters query 3's top 5 on its keyword count alone, which means the keyword weight is too high. Measured on this slice, it stays out once the cosine weight reaches about 0.92. On 30 trails the keywords mostly add noise; Feature 05 measures the same blend on a bigger corpus, where they help.

  Uncomment the `tokens = ...` line from step 0 (it needs `import re`, which the starter already has). Leave the rest of the keyword code commented out. Then replace the step 3 `results = ...` line and the printing lines below it with the hint. It counts hits per trail with the same `re.search` test the starter used, min-max rescales both scores, sorts on the weighted sum, and prints both scores. The two weights are numbers you pick that add up to 1:

  ```python
  # Hint: two scores per trail, each min-max rescaled to 0..1, ranked on a weighted sum
  scored = []
  for t in trails:
      haystack = f"{t['name']} {t['description']}".lower()
      hits = len([w for w in tokens if re.search(rf"\b{w}\b", haystack)])
      scored.append((cosine_similarity(query_vector, vectors[t["id"]]), hits, t))

  def rescale(x, lo, hi):
      return 0.0 if hi == lo else (x - lo) / (hi - lo)

  cos_lo, cos_hi = min(r[0] for r in scored), max(r[0] for r in scored)
  hit_lo, hit_hi = min(r[1] for r in scored), max(r[1] for r in scored)
  results = sorted(scored, key=lambda r: -(<cosine weight> * rescale(r[0], cos_lo, cos_hi) + <keyword weight> * rescale(r[1], hit_lo, hit_hi)))[:5]
  print(f'\nBlended search: "{query}"\n')
  for cosine, hits, trail in results:
      print(f"{cosine:.4f} cos  {rescale(hits, hit_lo, hit_hi):.2f} kw  {trail['id']}  {trail['name']}")
  ```

## What Is in This Folder

- `data/trails-slice.json`: 30 trails lifted verbatim from feature 10's 200-trail [`trails.json`](../../../M4-doing/F10-agentic-workflows/data/trails.json), chosen so the catalog has four right answers for query 1 that never use the query's words, and three keyword traps. Step 1 names them.
- `data/queries.json`: the three test queries, why each one is in the set, and the success check for each. Step 4 describes the record shape.
- `embeddings.json`: not in the folder until you run the program. Step 2 explains the cache.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
