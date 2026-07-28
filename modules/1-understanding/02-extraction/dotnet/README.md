# .NET demo for 02 Extraction

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one naive prompt ("Extract the details of this trip report as JSON."), no schema, so you get whatever shape the model feels like emitting.
- `complete/`: the finished demo as shown on stage. A C# record `TripFacts` with nullable fields and `[Description]` attributes, handed to `GetResponseAsync<TripFacts>()`. The schema is code, and there is no parsing step. Then the validator, which is the half of the demo that ships.

Each report prints three blocks: what the model gave us, what the validator says field by field, and what we would actually store. The rules are plain C# in `Program.cs`, no model involved:

- `date_hiked` must parse against an explicit format list or be `null`. "July 4, 2026" passes and is normalized to `2026-07-04`; "last month" is rejected.
- `distance_mi` and `elevation_gain_ft` reject `0`, negatives, and the absurd (over 100 mi, over 20,000 ft). Zero is the near-miss worth pausing on: a pipeline stores it without complaint.
- Empty and whitespace-only strings are rejected.
- `trail_name` and `park` get a grounding check against the source text, which is what catches an invented trail name.

Anything rejected becomes `null` in the stored object. Null is a gap someone can fill later; a plausible wrong number is not.

Both run against Ollama (`llama3.2`, JSON mode), matching the demo outline in [../README.md](../README.md). From `complete/`:

```bash
dotnet run                                       # both lab reports, extracted and validated
dotnet run -- ../../lab/tr-0011.md               # just the sparse one, for the null check
dotnet run -- ../../../../data/trip-reports/tr-0002.md   # any report path works
```

Run the sparse report three or four times on stage. The output moves, and that is the demo: some runs come back clean, and some hand you `0` or "last month" and the validator says so out loud. Real output from four consecutive runs, rejections included, lives in [../lab/expected-output.md](../lab/expected-output.md).
