# .NET demo for 02 Extraction

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one naive prompt ("Extract the details of this trip report as JSON."), no schema, so you get whatever shape the model feels like emitting.
- `complete/`: the finished demo as shown on stage. A C# record `TripFacts` with nullable fields and `[Description]` attributes, handed to `GetResponseAsync<TripFacts>()`. The schema is code, and there is no parsing step.

Both run against Ollama (`llama3.2`, JSON mode), matching the demo outline in [../README.md](../README.md). From `complete/`:

```bash
dotnet run                                       # both lab reports as typed rows
dotnet run -- ../../lab/tr-0011.md               # just the sparse one, for the null check
dotnet run -- ../../../../data/trip-reports/tr-0002.md   # any report path works
```

Real output from a run lives in [../lab/expected-output.md](../lab/expected-output.md), including the spots where `llama3.2` still ignores the "null if not stated" rule.
