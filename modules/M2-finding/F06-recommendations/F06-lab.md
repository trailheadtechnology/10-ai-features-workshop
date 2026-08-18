# Lab 06: Recommendations

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `data/` provides the same trail slice as feature 04 and three target trails. If you finished feature 04's lab, reuse your vectors; if not, `data/` includes precomputed embeddings so you can start here.
- **How:** Ollama embeddings via `http/ollama.http` (only if computing fresh), then your own cosine-similarity ranking from feature 04.
- **Steps:**
  1. For target trail #1, rank all other trails by similarity and print the top 5.
  2. Repeat for the other two targets. Success check: your top 5 substantially overlaps the acceptable sets in `expected-output.md` (there is more than one right answer, and the file says which neighbors are defensible).
  3. Read your own results and ask whether you'd ship them. That judgment call is the actual skill.
- **Stretch goal:** recommend for a user who liked two trails by averaging their vectors, or filter recommendations by metadata (only same park, only easier difficulty).

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
