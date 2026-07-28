# Lab assets for 06 Recommendations

Everything the lab spec in [../F06-spec.md](../F06-spec.md) references:

- `trails.json`: a 30-trail slice of `../../data/trails.json`, same shape as the full file (id, name, park, distance, elevation, difficulty, features, description). It holds the three target trails plus enough neighbors and enough noise to make the ranking interesting.
- `trail-embeddings.json`: all 30 vectors, precomputed with `nomic-embed-text`, keyed by trail id (768 floats each). Reuse your feature 04 vectors if you have them; this file is here so you can start cold.
- `ollama.http`: the embeddings requests, needed only if you want fresh vectors. One description, a batch of three for a cosine sanity check, and the gear variant that embeds a product's reviews.
- `expected-output.md`: real top-5 neighbor lists with scores for all three targets, the acceptable sets to grade yourself against, the gear result, and an honest accounting of which recommendations are bad and why.

The three targets: `trail-0117` Avalanche Lake Trail (the demo's trail), `trail-0003` Trail of the Cedars (the easy end of the catalog), and `trail-0008` Highline Trail (where the whole approach struggles).
