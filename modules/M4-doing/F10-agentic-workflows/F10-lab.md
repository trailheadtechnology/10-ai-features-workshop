# Lab 10: Agentic Workflows

*Everyone attempts this one; it is the only feature in [Module 4](../M4-overview.md) and the hands-on period runs about 50 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand to feel the mechanics, then extend a working agent with a new tool.
- **Input:** `data/` provides the tool definitions as JSON, the mock API fixtures in `data/mock-apis/`, and a transcript of a complete agent run for reference.
- **How:** `http/azure.http` (Azure OpenAI, key handed out in the room) contains the full round-trip as sequential requests: send the request with the `tools` array, read the tool call out of the response, then send the follow-up with the tool result. In an `.http` file, you are the loop, which is the best way to understand what agent frameworks hide.
- **Steps:**
  1. Run the provided two-tool round-trip (`search_trails`, `check_campsites`) by hand and watch the model choose, receive, and continue.
  2. Add `get_weather` to the tools array (its JSON definition is your job) and re-run with a weather-dependent request. Success check: the model calls your new tool and the final itinerary reflects the forecast (compare `expected-output.md`).
  3. Ask for a trip on the trail with the washed-out bridge. Success check: the itinerary avoids or flags it.
- **Stretch goal:** add the human gate. Insert a confirmation step before `request_permit` gets executed, and only pass the tool result back after an explicit yes. If you're in a language with a loop already written, wire the whole thing end to end.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F10-http.md`](http/F10-http.md) | the requests in `http/ollama.http` and `http/azure.http`, or a port of them in your language |
| .NET | [`dotnet/F10-dotnet.md`](dotnet/F10-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F10-python.md`](python/F10-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F10-typescript.md`](typescript/F10-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tool-definitions.json`: the five tools as JSON-schema function definitions, ready to paste into the `tools` array of an OpenAI-compatible request: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The descriptions are the model's only manual, so read them before you write your own.
- `reference-transcript.md`: a real, unedited run of `dotnet/complete` against `llama3.2` for "Plan me a 3-day trip in Glacier National Park for September 14-16". Every tool call in order with its arguments, the results (truncated as printed, then in full for the first few), and the itinerary that came out. A second run at the end asks for the closed trail and shows the agent routing around it.
- `expected-output.md`: the success checks for all three lab steps plus the stretch goal, with the failure modes each one is meant to catch.

The tools read the data in this folder: `trails.json` (the full 200-trail catalog that features 04 and 06 use a slice of), `condition-reports.jsonl` (the full stream feature 08 mined), and the fixtures in `mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` surfaces it.

One detail of `search_trails` matters more than it looks. Its `features` keywords match the trail's name as well as its feature tags. The tool returns at most eight trails, and Avalanche Lake is the 27th Glacier trail in the catalog with no feature tag containing "Avalanche". Without the name match, a request for "Avalanche Lake Trail" has no way to reach `trail-0117` through search. In a soak test against `gpt-5.5` with a tag-only search, the agent gave up on the trail in six runs out of ten. It said so honestly and planned around it, but the washed-out bridge never came up, so the lab step had nothing to teach. The model did what it could with the tool it had. That is a useful thing to remember when your own agent misbehaves, because the tool is often the cheaper place to look first.

If you'd rather run the loop from your own language than from an `.http` file, `data/tool-definitions.json` gives you the request bodies and the loop is written out three ways: `dotnet/complete/Program.cs`, `python/complete/main.py`, and `typescript/complete/index.ts` (the last two are about thirty lines with no framework, which makes them the easiest to read). The tool results in `http/azure.http` are still the ones to hand back.
