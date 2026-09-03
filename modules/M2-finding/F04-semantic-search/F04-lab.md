# Lab 04: Semantic Search

*This is the Recommended lab for [Module 2](../M2-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** rank trails by similarity to a natural-language query, and beat keyword search on three queries it fails.
- **Input:** `data/trails-slice.json`, 30 trails with three keyword traps; `data/queries.json`, the three test queries and their checks.
- **How:** POST to Ollama's embeddings endpoint, `http/ollama.http` (requests 1 to 3). Cosine similarity and ranking are code you write.
- **Model:** `nomic-embed-text`, local. No key. 768 floats per input.

### Step 1: Embed the 30 descriptions and cache them

Request 1 in `http/ollama.http` embeds one string, request 2 a batch. Request 1's input:

```text
a gentle grade shaded by cedars
```

Request 2 as written:

```json
{
  "model": "nomic-embed-text",
  "input": [
    "Because it starts just outside the park boundary in national forest, leashed dogs can come along to this year-round waterfall on the Tuolumne's south fork. The path rolls gently through pine and incense cedar with one rooty stretch near the end. The pool below the falls warms enough for swimming by late July.",
    "Do not let the name fool you; it honors homesteader Elias Easy, not the terrain. The route gains over 2,600 feet in under four miles, much of it on loose, ankle-rolling rubble beside the creek's headwall. Search and rescue answers more calls here than on any comparable trail in the district."
  ]
}
```

Embed all 30 `description` strings from `data/trails-slice.json` in one call and keep the vectors keyed by trail id (`dotnet/complete/` caches them in `embeddings.json` next to the binary).

**Check:** one vector per input, in order, 768 floats each; request 1 starts `[0.0611, -0.0001, -0.2007, -0.0624, -0.022, -0.0063, 0.0232, -0.0023]`. Fewer than 30 vectors, or one not 768 long, means the batch did not go through.

### Step 2: Embed query 1 and rank

Request 3, or `dotnet run -- dog-friendly waterfall hike, not too steep` from `dotnet/complete/`. Request 3's input:

```text
dog-friendly waterfall hike, not too steep
```

Write cosine similarity (pseudocode in `expected-output.md`; `CosineSimilarity` in `dotnet/complete/Program.cs`), score every trail against the query, and print the top 5.

**Check:** at least two of `trail-0068`, `trail-0055`, `trail-0027`, `trail-0011` in your top 3; recorded first is `trail-0068` Carlon Falls at `0.7733`. `trail-0007` Upper Yosemite Falls first means you are counting keyword hits, the `dotnet/starter` baseline.

### Step 3: The other two queries

Change request 3's `input`, or the `dotnet run` arguments, to each of these:

```text
somewhere quiet to take my kids
```

```text
an easy hike to a great view
```

**Check:** query 2 puts Taft Point (`trail-0020`) first at `0.4876` (correct, and still a cliff edge); query 3 puts `trail-0058` Panorama Cliffs Bypass at or near the top (`0.6481`). `trail-0074` Easy Creek Trail in query 3's top 5 means you are matching words, not meaning.

### Stretch goal: filter before you rank, or blend scores

Drop `difficulty: hard` trails before ranking, require `dog-friendly` in `features` for query 1, or add the keyword-hit count from `dotnet/starter` into the score.

**Check:** query 1 returns only dog-friendly trails; query 3 loses Beehive Loop (`trail-0017`) and Chimney Tops (`trail-0005`) from the top 5. Either still there means the filter ran after the top 5 was taken.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F04-http.md`](http/F04-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F04-dotnet.md`](dotnet/F04-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F04-python.md`](python/F04-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F04-typescript.md`](typescript/F04-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/trails-slice.json`: 30 trails lifted verbatim from the full 200-trail catalog, [`trails.json`](../../M4-doing/F10-agentic-workflows/data/trails.json) in feature 10's lab. Four of them are dog-friendly waterfall trails with gentle grades and not one phrases it that way, and three are keyword traps: "Easy Creek Trail" is a hard 2,610-foot climb named after a homesteader, "Dog Lake Trail" bans dogs, and Panorama Cliffs Bypass matches the word "steep" while describing how it avoids the steep parts.
- `data/queries.json`: the three test queries, why each one is in the set, and the success check for each.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
