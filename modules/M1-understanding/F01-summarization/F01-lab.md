# Lab 01: Summarization

*This is the Recommended lab for [Module 1](../M1-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** two trip reports provided in `data/`, one with a buried hazard warning, drawn from the forty in `data/trip-reports/`.
- **How:** POST to Ollama's chat endpoint (`llama3.2`). `http/ollama.http` has the exact request; port it to your language or run it as-is.
- **Steps:**
  1. Send report #1 with the naive prompt ("summarize this") and read the book report you get back.
  2. Rewrite the prompt to demand exactly 3 bullets covering conditions, hazards, and crowding, and nothing else. Run it a few times, not once: report #1 has no closure in it, and a prompt that demands a hazard bullet will happily invent one.
  3. Run report #2 through your improved prompt. Success check: your 3 bullets surface the buried hazard (compare `expected-output.md`), and report #1 still comes back with nothing closed.
- **Stretch goal:** make the summary audience-switchable. The same report, summarized for a hiker and then for a park ranger who cares about maintenance issues rather than scenery.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F01-http.md`](http/F01-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F01-dotnet.md`](dotnet/F01-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F01-python.md`](python/F01-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F01-typescript.md`](typescript/F01-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `data/tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
