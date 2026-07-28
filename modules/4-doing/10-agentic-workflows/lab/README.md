# Lab assets for 10 Agentic Workflows

Everything the lab spec in [../README.md](../README.md) references:

- `tool-definitions.json`: the five tools as JSON-schema function definitions, ready to paste into the `tools` array of an OpenAI-compatible request: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The descriptions are the model's only manual, so read them before you write your own.
- `reference-transcript.md`: a real, unedited run of `../dotnet/complete` against `llama3.2` for "Plan me a 3-day trip in Glacier National Park for September 14-16". Every tool call in order with its arguments, the results (truncated as printed, then in full for the first few), and the itinerary that came out. A second run at the end asks for the closed trail and shows the agent routing around it.
- `expected-output.md`: the success checks for all three lab steps plus the stretch goal, with the failure modes each one is meant to catch.

The tools read the workshop's own data: `data/trails.json`, `data/condition-reports.jsonl`, and the fixtures in `data/mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` surfaces it.

An `azure.http` file is not in this folder yet. Until it lands, build the round-trip in your own HTTP client using `tool-definitions.json` for the request bodies, or read the loop in `../dotnet/complete/Program.cs`.
