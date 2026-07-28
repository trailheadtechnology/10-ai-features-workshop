# .NET demo for 06 Recommendations

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. It loads the trail catalog and fills the "you might also like" box with five random trails, which is roughly what that box does in most apps today and takes no model at all.
- `complete/`: the finished demo as shown on stage. It embeds the 30 trail descriptions through `IEmbeddingGenerator`, ranks every other trail by cosine similarity to the one you name, skips the trail itself, and prints the top 5 with scores.

Both run against Ollama (`nomic-embed-text`), matching the demo outline in [../F06-spec.md](../F06-spec.md). From `complete/`:

```bash
dotnet run                             # more like this, for Avalanche Lake Trail
dotnet run -- trail-0008               # any trail id from ../../lab/trails.json
dotnet run -- Trail of the Cedars      # or any name, or part of one
dotnet run -- --gear Cascade 65        # step 5: same trick over gear review text
```

Vectors land in `complete/embeddings.json` and `complete/gear-embeddings.json` on the first run, so every run after that is instant. Delete a cache file to re-embed.

`Cosine` is a dozen lines at the bottom of `complete/Program.cs`, and the recommendation itself is the LINQ query above it. That is the demo's point: this is the feature 04 search code with the query text swapped for an item's vector.

Real output for all four commands, including the neighbors that are obviously wrong, is in [../lab/expected-output.md](../lab/expected-output.md).
