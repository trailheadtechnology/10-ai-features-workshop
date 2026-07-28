# .NET demo for 04 Semantic Search

Two console projects, both named `Search`, both reading the 30-trail slice in [`../lab/trails-slice.json`](../lab/trails-slice.json):

- `starter/`: the demo's starting point, and today's baseline. Keyword substring search, no AI at all: lowercase the query, keep words of three letters or more, count whole-word hits in each trail's name and description. It is 40 lines, and it is genuinely how most search boxes work.
- `complete/`: the finished demo as shown on stage. Microsoft.Extensions.AI's `IEmbeddingGenerator` over OllamaSharp, `nomic-embed-text`, cosine similarity in one visible method, top 5 with scores.

Both take the query as arguments and default to the demo query. From either directory:

```bash
dotnet run                                       # dog-friendly waterfall hike, not too steep
dotnet run -- somewhere quiet to take my kids
dotnet run -- an easy hike to a great view
```

Run `starter` first, get junk, then run `complete` on the same words. The 30 descriptions embed in under two seconds on a laptop, and `complete` caches the vectors to `embeddings.json` next to the binary, so only the first run pays even that. Delete that file to re-embed live and show the timing. Real output from both projects is recorded in [`../lab/expected-output.md`](../lab/expected-output.md).

**Starting point:** J.'s existing talk demo [trailheadtechnology/dotnet-semantic-search](https://github.com/trailheadtechnology/dotnet-semantic-search) ("Warm and Fuzzy: Semantic Search in .NET") already covers this feature's stack: Microsoft.Extensions.AI, Ollama, and vectorization. The build-out here adapts that code to the Trailhead Guides trail-description corpus in [`data/`](../../../../data/) instead of starting from scratch.
