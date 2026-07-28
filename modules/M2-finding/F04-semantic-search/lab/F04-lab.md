# Lab assets for 04 Semantic Search

Everything the lab spec in [../F04-spec.md](../F04-spec.md) references:

- `ollama.http`: three ready-to-run requests against Ollama (`nomic-embed-text`): one sentence, a batch of two, and the demo query. This endpoint is the only network call the lab makes; the ranking is arithmetic you write yourself.
- `trails-slice.json`: 30 trails lifted verbatim from [`data/trails.json`](../../../../data/trails.json). Four of them are dog-friendly waterfall trails with gentle grades and not one phrases it that way, and three are keyword traps: "Easy Creek Trail" is a hard 2,610-foot climb named after a homesteader, "Dog Lake Trail" bans dogs, and Panorama Cliffs Bypass matches the word "steep" while describing how it avoids the steep parts.
- `queries.json`: the three test queries, why each one is in the set, and the success check for each.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
