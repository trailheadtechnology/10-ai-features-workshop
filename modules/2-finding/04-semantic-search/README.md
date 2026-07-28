# 04 · Semantic search

Module 2: Finding · Runs on Ollama embeddings (`nomic-embed-text`), fully local

## The user problem

A user types "dog-friendly waterfall hike, not too steep" into Trailhead Guides. The catalog has at least a dozen perfect matches, but keyword search returns almost nothing, because no trail description contains the phrase "not too steep." One trail says "a gentle grade shaded by cedars." Another says "easy elevation, good for families." The user gets three bad results, assumes the app has no good trails, and goes back to asking strangers on Reddit.

## The concept

Embeddings turn text into vectors, points in a high-dimensional space where distance means similarity of meaning. "Gentle grade" and "not too steep" land close together in that space even though they share no words. Semantic search is the whole trick applied to a catalog: embed every trail description once, embed the user's query at search time, and rank by distance. The math at the center is cosine similarity, which is a few lines of code in any language.

Two things make this feature land. First, it runs entirely locally: `nomic-embed-text` is a small, free embedding model, and 200 trail descriptions embed in seconds on a laptop. Second, the search box stays a search box. Users don't have to learn anything new; the same input just starts understanding what they meant. This is also the foundation feature for the rest of the day, since RAG (05), recommendations (06), and anomaly detection (08) all reuse the idea, and 06 and 08 reuse the actual infrastructure.

## Demo outline (13 min, .NET)

1. Run the "dog-friendly waterfall hike, not too steep" query against naive keyword search over `data/trails.json` and get junk results. That's today's baseline.
2. Show what an embedding is: embed one sentence via Microsoft.Extensions.AI's `IEmbeddingGenerator` against Ollama, and look at the raw float array on screen for a moment. It's just numbers.
3. Embed the trail catalog in a loop and time it live. The checked-in demo embeds the 30-trail slice in `lab/trails-slice.json` in under two seconds. To run all 200 instead, change the one file path near the top of `Program.cs` to `../../../../data/trails.json`. Either way the point lands: seconds, not minutes, and it happens once at startup rather than per search.
4. Write cosine similarity on screen. It fits in one visible method, which surprises people.
5. Re-run the same query semantically. The payoff: the gentle shaded waterfall trails rise to the top, with the similarity scores visible next to each hit.
6. Show one more query with zero keyword overlap ("somewhere quiet to take my kids"), which proves it isn't a fluke and then hands you the feature's best moment. The top hit is Taft Point, whose own description warns that the ground gives no second chances beside a 3,000-foot drop. It is a perfect topical match and terrible advice. Say what that means: embeddings capture what text is about, not whether it's suitable, and the fix is metadata filters and a score floor, not a better model. Note that everything ran locally, and that the scores for this query top out near 0.49 against 0.77 for query one, which is the number a score floor would key on.

## Lab spec (13 min, any language)

- **Goal:** rank trails by semantic similarity to a natural-language query.
- **Input:** `lab/` provides a slice of `data/trails.json` (about 30 trails) and three test queries with expected top hits.
- **How:** POST to Ollama's embeddings endpoint (`nomic-embed-text`). `lab/ollama.http` has the exact request. You implement cosine similarity yourself; it's a dozen lines, and `lab/expected-output.md` includes pseudocode if you want it.
- **Steps:**
  1. Embed the 30 trail descriptions and cache the vectors in memory.
  2. Embed query #1, compute similarity against every trail, and print the top 5.
  3. Repeat for the other queries. Success check: the expected trail appears in your top 3 for each query.
- **Stretch goal:** combine semantic score with metadata filters (only `dog-friendly`, only under 6 miles), or blend keyword and semantic scores into one ranking.

## Leadership beat

- **When to reach for this:** any search box users complain about, and any catalog, knowledge base, or product list where people describe what they want instead of naming it.
- **Rough cost & effort:** days to a prototype on an existing catalog. Embeddings are cheap to free, and small local models handle them well. The work is in evaluation and tuning, not infrastructure.
- **The one-liner for your CTO:** "Our search finds what users mean, not just what they type."

This card is row 4 of the [decision framework](../../../docs/decision-framework.md).
