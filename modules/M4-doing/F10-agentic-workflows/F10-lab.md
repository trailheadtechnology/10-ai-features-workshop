# Lab 10: Agentic Workflows

*Everyone attempts this one; it is the only feature in [Module 4](../M4-overview.md) and the hands-on period runs about 50 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand, then extend a working agent with a new tool.
- **Input:** `data/tool-definitions.json`, the five tools (`search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`); `data/trails.json`, the trail catalog; `data/condition-reports.jsonl`, the hiker reports; `data/mock-apis/weather.json`, `campsites.json`, `permits.json`, canned results; `reference-transcript.md`, a complete run to compare against.
- **How:** POST to Azure OpenAI chat completions. `http/azure.http` holds requests 1a, 1b, 1c, 2, 3, and the stretch, one per turn, with every tool result you hand back; you are the loop.
- **Model:** `gpt-5.5` on Azure, endpoint filled in; paste the room key over `<KEY FROM INSTRUCTOR>`. Code tracks read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT` (`gpt-5.5`), else fall back to `llama3.2`, much weaker here.

### Step 1: Requests 1a to 1c, the two-tool round-trip by hand

Send 1a in `http/azure.http`, whose `tools` array holds these two definitions from `data/tool-definitions.json`:

```json
{
  "type": "function",
  "function": {
    "name": "search_trails",
    "description": "Search the trail catalog. Returns matching trails with id, name, park, distance, elevation, difficulty, and features. Call this before recommending any trail; trail ids from here are what the other tools expect.",
    "parameters": {
      "type": "object",
      "properties": {
        "park": {
          "type": "string",
          "description": "Park name, for example 'Glacier National Park'. Partial names like 'Glacier' also match."
        },
        "features": {
          "type": "array",
          "items": { "type": "string" },
          "description": "Optional keywords matched against each trail's features and its name, for example ['lake', 'waterfall'] or ['Avalanche Lake']."
        },
        "max_difficulty": {
          "type": "string",
          "enum": ["easy", "moderate", "hard"],
          "description": "Optional ceiling on difficulty. 'moderate' returns easy and moderate trails."
        }
      },
      "required": ["park"]
    }
  }
},
{
  "type": "function",
  "function": {
    "name": "check_campsites",
    "description": "Check campground availability for a park. Returns each campground with open sites per date, whether it is frontcountry or backcountry, amenities, and notes such as seasonal closures.",
    "parameters": {
      "type": "object",
      "properties": {
        "park": {
          "type": "string",
          "description": "Park name, for example 'Glacier National Park'."
        }
      },
      "required": ["park"]
    }
  }
}
```

System:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, campgrounds, or availability. Every trail name and campground in your answer must have come back from a tool call in this conversation. Call search_trails and check_campsites before writing any part of the itinerary. If you have not yet called both, your next move is a tool call, not prose.
```

User:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

Copy the 1a assistant message into 1b with its `id` over `CALL-ID`, send, then repeat in 1c with the second call (`CALL-ID-2`), swapping the results if the model asked for `check_campsites` first. **Check:** 1a returns `finish_reason: "tool_calls"` naming `search_trails` or `check_campsites`; 1c returns `finish_reason: "stop"` and an itinerary naming only trails from 1b and campgrounds from 1c. An itinerary at 1a means the model never saw your tools; any other name means a tool result did not reach it.

### Step 2: Request 2, add get_weather

In request 2, write your own `get_weather` definition (name, description, a `park` parameter) over `WRITE-GET-WEATHER-HERE` before reading the reference in `data/tool-definitions.json`:

```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Get the multi-day forecast and advisories for a park. Returns one entry per date with high, low, conditions, chance of precipitation, and wind.",
    "parameters": {
      "type": "object",
      "properties": {
        "park": {
          "type": "string",
          "description": "Park name, for example 'Glacier National Park'."
        }
      },
      "required": ["park"]
    }
  }
}
```

System:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, and campground in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails. 3. check_campsites for where to stay each night. Then write the itinerary: one section per day with trail, campsite, and how the forecast shaped the choice (put harder or more exposed hiking on the drier days).
```

User:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather.
```

Run the loop as in step 1, handing back the forecast from the comment under request 2 (matching `tool_call_id`) when the model calls `get_weather`, then the 1b and 1c results. **Check:** in the request 2 itinerary the hard hike lands on the 14th or 15th and the 16th, the rain day in `data/mock-apis/weather.json`, gets something short or sheltered, with a sentence saying why. Failing looks like the same three trails plus "expect rain on the 16th"; if the model never calls `get_weather`, fix your description first.

### Step 3: Request 3, the washed-out bridge on trail-0117

Request 3 carries all five tools; the two new ones, from `data/tool-definitions.json`:

```json
{
  "type": "function",
  "function": {
    "name": "get_trail_conditions",
    "description": "Get the most recent hiker-submitted condition reports for one trail, newest first. Call this for every trail you intend to recommend. Reports are where closures, washouts, and hazards show up; the catalog entry will not mention them.",
    "parameters": {
      "type": "object",
      "properties": {
        "trail_id": {
          "type": "string",
          "description": "A trail id returned by search_trails, for example 'trail-0117'."
        }
      },
      "required": ["trail_id"]
    }
  }
},
{
  "type": "function",
  "function": {
    "name": "request_permit",
    "description": "Submit a backcountry permit request. This files a request on the user's behalf, so call it once, at the end, after the plan is settled and the user has confirmed.",
    "parameters": {
      "type": "object",
      "properties": {
        "park": {
          "type": "string",
          "description": "Park name, for example 'Glacier National Park'."
        },
        "zone": {
          "type": "string",
          "description": "Permit zone, for example 'Lake McDonald / Sperry' or 'Many Glacier'."
        },
        "dates": {
          "type": "string",
          "description": "Trip dates, for example '2026-09-14 to 2026-09-16'."
        },
        "group_size": {
          "type": "integer",
          "description": "Number of people in the group. Zones cap out at 8 in Glacier."
        }
      },
      "required": ["park", "zone", "dates", "group_size"]
    }
  }
}
```

System:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, campground, and condition in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails that fit the request. 3. get_trail_conditions for EVERY trail you intend to recommend, one call per trail, using the trail id returned by search_trails. If the newest reports for a trail mention a closure, a washout, a bridge that is out, or any other reason hikers are turning around, that trail is CLOSED. Do not schedule a day on a closed trail. Replace it with another trail from search_trails and state plainly, in the itinerary, that the original trail is closed and why. 4. check_campsites for where to stay each night. 5. request_permit once, only if a backcountry site or permit zone is involved. If you have not yet called search_trails and get_trail_conditions, your next move is a tool call, not prose. Then write the final itinerary: one section per day with trail, campsite, and how the forecast shaped the choice. End with the permit status.
```

User:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).
```

Run the loop, handing back the `["waterfall"]` search result and `trail-0117` conditions from the comments under request 3, the weather and campsites from before, and `{"error": "No condition reports found for 'trail-XXXX'."}` for any other trail id (.NET: `dotnet run -- Plan me a trip on Avalanche Lake Trail in September` in `dotnet/complete/`). **Check:** the model calls `get_trail_conditions` with `trail-0117` and the itinerary drops or flags the trail because the bridge is out, traceable to the tool result. Failures, worst last: never calling the tool and scheduling `trail-0117`; reading "the bridge is OUT" and scheduling it anyway; dropping it with an invented reason without calling the tool.

### Stretch goal: the human gate on request_permit

When the model calls `request_permit`, stop, read its arguments (`park`, `zone`, `dates`, `group_size`) as the summary a human approves, and send the stretch request in `http/azure.http` only after a yes; if you decline, send this as the tool content instead:

```text
{"status":"cancelled","message":"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed."}
```

In a code track, `dotnet/complete` asks `File it? [y/N]` inside `RequestPermit`; `--yes` skips it. **Check:** `request_permit` never runs on the model's say-so, and a declined stretch request ends with an itinerary that says no permit was filed, without retrying. Decide your step budget up front (`dotnet/complete` caps at 12 iterations).

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
