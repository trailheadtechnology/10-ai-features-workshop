# Lab assets for 10 Agentic Workflows

Everything the lab spec in [../F10-spec.md](../F10-spec.md) references:

- `azure.http`: the tool-calling round-trip against Azure OpenAI, one request per turn, with you playing the loop. Step 1 is fully written out (three requests: the model asks for `search_trails`, you hand back the result, it asks for `check_campsites`, you hand that back, you get an itinerary). Step 2 leaves a marked hole where your `get_weather` definition goes. Step 3 has all five tools and the closed-trail request. Every tool result you paste is the real output of the .NET tools over the workshop data, so it matches what the demo saw. There is no `ollama.http` for this feature: the loop mechanics are identical, but the lab runs on the cloud model on purpose (see the spec for why).
- `tool-definitions.json`: the five tools as JSON-schema function definitions, ready to paste into the `tools` array of an OpenAI-compatible request: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The descriptions are the model's only manual, so read them before you write your own.
- `reference-transcript.md`: a real, unedited run of `../dotnet/complete` against `llama3.2` for "Plan me a 3-day trip in Glacier National Park for September 14-16". Every tool call in order with its arguments, the results (truncated as printed, then in full for the first few), and the itinerary that came out. A second run at the end asks for the closed trail and shows the agent routing around it.
- `expected-output.md`: the success checks for all three lab steps plus the stretch goal, with the failure modes each one is meant to catch.

The tools read the workshop's own data: `data/trails.json`, `data/condition-reports.jsonl`, and the fixtures in `data/mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` surfaces it.

If you'd rather run the loop from your own language than from an `.http` file, `tool-definitions.json` gives you the request bodies and `../dotnet/complete/Program.cs` shows the loop; the tool results in `azure.http` are still the ones to hand back.
