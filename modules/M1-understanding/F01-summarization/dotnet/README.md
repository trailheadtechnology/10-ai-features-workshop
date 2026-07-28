# .NET demo for 01 Summarization

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one naive prompt ("Summarize this trip report."), which produces the book report.
- `complete/`: the finished demo as shown on stage.

Both run against Ollama (`llama3.2`), matching the demo outline in [../README.md](../README.md). From `complete/`:

```bash
dotnet run                                  # naive prompt on tr-0004 (the book report)
dotnet run -- --briefing                    # 3-bullet hiker briefing; the washout leads
dotnet run -- --headline                    # one-line status for a trail card UI
dotnet run -- --briefing --audience ranger  # stretch goal: same report, different audience
dotnet run -- ../../lab/tr-0001.md          # any report path works
```
