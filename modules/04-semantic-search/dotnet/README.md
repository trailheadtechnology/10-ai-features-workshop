# .NET demo for 04 Semantic Search (to be built)

Two console projects will live here, both built on Microsoft.Extensions.AI:

- `starter/`: the demo's starting point, for following along live
- `complete/`: the finished demo as shown on stage

Both will run with `dotnet run` against Ollama (`nomic-embed-text`), matching the demo outline in [../README.md](../README.md).

**Starting point:** J.'s existing talk demo [trailheadtechnology/dotnet-semantic-search](https://github.com/trailheadtechnology/dotnet-semantic-search) ("Warm and Fuzzy: Semantic Search in .NET") already covers this module's stack — Microsoft.Extensions.AI, Ollama, and vectorization. The build-out here is an adaptation of that code, retargeted at the Trailhead Guides trail-description corpus in [`data/`](../../../data/) rather than a from-scratch demo.
