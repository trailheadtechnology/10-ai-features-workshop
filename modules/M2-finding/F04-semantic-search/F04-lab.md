# Lab 04: Semantic Search

*This is the Recommended lab for [Module 2](../M2-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** rank trails by semantic similarity to a natural-language query.
- **Input:** `data/` provides a 30-trail slice of the Trailhead catalog (the full 200-trail `trails.json` lives in feature 10's `data/`) and three test queries with expected top hits.
- **How:** POST to Ollama's embeddings endpoint (`nomic-embed-text`). `http/ollama.http` has the exact request. You implement cosine similarity yourself; it's a dozen lines, and `expected-output.md` includes pseudocode if you want it.
- **Steps:**
  1. Embed the 30 trail descriptions and cache the vectors in memory.
  2. Embed query #1, compute similarity against every trail, and print the top 5.
  3. Repeat for the other queries. Success check: the expected trail appears in your top 3 for each query.
- **Stretch goal:** combine semantic score with metadata filters (only `dog-friendly`, only under 6 miles), or blend keyword and semantic scores into one ranking.

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
