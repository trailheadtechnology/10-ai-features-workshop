# 04 · Semantic Search

Module 2: Finding · Runs on Ollama embeddings (`nomic-embed-text`), fully local

## The User Problem

A user types "dog-friendly waterfall hike, not too steep" into Trailhead Guides. The catalog has at least a dozen perfect matches, but keyword search returns almost nothing, because no trail description contains the phrase "not too steep." One trail says "a gentle grade shaded by cedars." Another says "easy elevation, good for families." The user gets three bad results, assumes the app has no good trails, and goes back to asking strangers on Reddit.

## The Concept

Embeddings turn text into vectors, points in a high-dimensional space where distance means similarity of meaning. "Gentle grade" and "not too steep" land close together in that space even though they share no words. Semantic search is the whole trick applied to a catalog: embed every trail description once, embed the user's query at search time, and rank by distance. The math at the center is cosine similarity, which is a few lines of code in any language.

Two things make this feature land. First, it runs entirely locally: `nomic-embed-text` is a small, free embedding model, and 200 trail descriptions embed in seconds on a laptop. Second, the search box stays a search box. Users don't have to learn anything new; the same input just starts understanding what they meant. This is also the foundation feature for the rest of the day, since RAG (05), recommendations (06), and anomaly detection (08) all reuse the idea, and 06 and 08 reuse the actual infrastructure.

## The Lab

The hands-on lab is [F04-lab.md](F04-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is the Recommended lab for its module: start here unless you have a reason not to.

## Leadership Beat

- **When to reach for this:** any search box users complain about, and any catalog, knowledge base, or product list where people describe what they want instead of naming it.
- **Rough cost & effort:** days to a prototype on an existing catalog. Embeddings are cheap to free, and small local models handle them well. The work is in evaluation and tuning, not infrastructure.
- **The one-liner for your CTO:** "Our search finds what users mean, not just what they type."

This card is row 4 of the [decision framework](../../../docs/decision-framework.md).
