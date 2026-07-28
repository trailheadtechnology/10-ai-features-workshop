# .NET demo for 07 Classification & Routing

Two console projects named `Triage`, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one classify call, one inquiry. The taxonomy is plain-language category descriptions in the prompt and the model answers as free text, which is fine until it answers "Emergency." or "the category is emergency" and your `switch` falls through to the default queue.
- `complete/`: the finished demo as shown on stage. Same taxonomy, but through `GetResponseAsync<TriageResult>` with a C# enum, so the label is one of six values or the call fails. It classifies all 20 messages in `../../lab/inquiries-slice.jsonl`, prints anything it called an emergency at the top in a block you cannot scroll past, prints the routing table (id, category, queue), and scores itself against `../../lab/reference-labels.json`.

Both run against Ollama (`llama3.2`), matching the demo outline in [../README.md](../README.md).

```bash
cd starter && dotnet run              # classify inq-0005
cd starter && dotnet run inq-0041     # any id from the slice
cd complete && dotnet run             # all 20, routed and scored
```

A real run of `complete/` scored 17/20 against the reference labels with both emergencies caught and no false emergencies. The three misses were `inq-0008` and `inq-0030` (permit questions filed as `general`, which is the misclassification you fix live by widening the `permit` description) and `inq-0035`, the ambiguous permit-versus-conditions message. See [../lab/expected-output.md](../lab/expected-output.md) for the full labeling and what it means.

Emergency recall is the number to watch on stage. Accuracy moves a point or two between runs; missing an emergency is the failure the demo exists to talk about.
