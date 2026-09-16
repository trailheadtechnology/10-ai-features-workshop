# Lab 10: Agentic Workflows

*Everyone attempts this one; it is the only feature in [Module 4](../M4-overview.md) and the hands-on period runs about 50 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand, then extend a working agent with a new tool.
- **Input:** `data/tool-definitions.json`, the five tools (`search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`); `data/trails.json`, the trail catalog; `data/condition-reports.jsonl`, the hiker reports; `data/mock-apis/weather.json`, `campsites.json`, `permits.json`, canned results; `reference-transcript.md`, a complete run to compare against.
- **How:** chat completions with a `tools` array, against Azure OpenAI.
- **Model:** `gpt-5.5` on Azure. Every track reads `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (`gpt-5.5`), else falls back to `llama3.2`, much weaker here.

Every step below is one thing to make the program do. Every track's `starter/` is a plain chat call with no tools. Edit it until it does every step. Compare against `complete/` when you get stuck.

### Step 0: Run the starter and read the plan with no tools

**Do:** run `starter/` as it is. It sends this request as one user message, no tools, no loop:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

**Check:** a fluent three-day plan with zero tool calls. It names trails and campgrounds the model half-remembers. It checks no weather, reads no conditions, and books nothing. The rest of the lab closes that gap.

### Step 1: Write the first two tools as ordinary functions

**Do:** each tool is a plain function that reads a file under `data/` and returns a JSON string. At the top of each, print `[tool] ` plus the tool name and its arguments — that line is how you watch the loop run.

1. Write `search_trails(park, features, max_difficulty)`. Open `data/trails.json` (200 trails across six parks, 45 in Glacier). Keep trails whose `park` contains the `park` argument (case-insensitive). If `max_difficulty` was given, drop trails harder than it (`easy` < `moderate` < `hard`). If `features` was given, keep a trail only when at least one keyword appears in its `name` **or** its `features` tags — do not skip the name match. Stop after 8 matches. Return them as a JSON array without the `description` field.
2. Write `check_campsites(park)`. Open `data/mock-apis/campsites.json` (an object keyed by park name, skip the `_comment` key). Find the key that matches the park: the key contains the argument, the argument contains the key, or the key contains the argument's first word, ignoring case. Return that entry as a JSON string, or `{"error": "No campsite data for '<park>'."}` if nothing matches.

**Why:** the planted record is `trail-0117`, Avalanche Lake Trail, whose catalog description mentions the footbridge crossing at the gorge's mouth and says nothing about it being gone — only the condition reports (step 4) do. Avalanche Lake Trail is the 27th Glacier trail in the catalog and none of its feature tags contain "Avalanche", so without the name match in `search_trails`, a search for it can never reach it. Similarly, `check_campsites`'s planted fact (named in the fixture's own `_comment`) is that Sperry Chalet Area Sites shows 0 open sites on the 13th, 14th, and 15th, so a plan that sleeps there those nights ignored the tool.

**Check:** `search_trails` with `{"park":"Glacier National Park"}` returns 8 trails, from `trail-0003` Trail of the Cedars to `trail-0037` Bowman Lake Shoreline Trail. `check_campsites` with the same park returns 4 campgrounds, and Sperry Chalet Area Sites shows 0 open sites on September 14 and 15.

### Step 2: Write the loop and run the two-tool round-trip

**Do:**
1. Declare the two tools with the definitions in `data/tool-definitions.json` (your track's walkthrough shows the mechanism):

```json
{
  "type": "function",
  "function": {
    "name": "search_trails",
    "description": "Search the trail catalog. Returns matching trails with id, name, park, distance, elevation, difficulty, and features. Call this before recommending any trail; trail ids from here are what the other tools expect.",
    "parameters": {
      "type": "object",
      "properties": {
        "park": {"type": "string", "description": "Park name, for example 'Glacier National Park'. Partial names like 'Glacier' also match."},
        "features": {"type": "array", "items": {"type": "string"}, "description": "Optional keywords matched against each trail's features and its name, for example ['lake', 'waterfall'] or ['Avalanche Lake']."},
        "max_difficulty": {"type": "string", "enum": ["easy", "moderate", "hard"], "description": "Optional ceiling on difficulty. 'moderate' returns easy and moderate trails."}
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
      "properties": {"park": {"type": "string", "description": "Park name, for example 'Glacier National Park'."}},
      "required": ["park"]
    }
  }
}
```

2. Make a dictionary mapping each tool name to its function.
3. Start `messages` with a system message and a user message:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, campgrounds, or availability. Every trail name and campground in your answer must have come back from a tool call in this conversation. Call search_trails and check_campsites before writing any part of the itinerary. If you have not yet called both, your next move is a tool call, not prose.
```

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

4. Loop at most 12 times — that cap is the step budget; nothing else stops a model that keeps asking for one more tool.
5. Each pass, send `messages` and the tool definitions, tool choice automatic. Read back the assistant message.
6. If it carries no tool calls, print its text and stop — that text is the itinerary.
7. Otherwise, append the assistant message to `messages` exactly as it came back, tool calls included.
8. For each tool call, parse its arguments string, look up the function by name, call it.
9. Append the result as a `tool` message tied to that call's `id`. Go back to 5.

**Check:** one `[tool]` line prints for each of the two tools, then the itinerary. An itinerary with no tool lines means the model never saw your tools; a mismatched name means a tool result did not reach it.

### Step 3: Add get_weather and plan around the rain day

**Do:**
1. Write `get_weather(park)` the same way as `check_campsites`, reading `data/mock-apis/weather.json`; return `{"error": "No forecast available for '<park>'."}` if nothing matches.
2. Write the `get_weather` tool definition yourself — name, description, one `park` parameter. Only then check the reference in `data/tool-definitions.json`:

```json
{
  "type": "function",
  "function": {
    "name": "get_weather",
    "description": "Get the multi-day forecast and advisories for a park. Returns one entry per date with high, low, conditions, chance of precipitation, and wind.",
    "parameters": {
      "type": "object",
      "properties": {"park": {"type": "string", "description": "Park name, for example 'Glacier National Park'."}},
      "required": ["park"]
    }
  }
}
```

3. Declare `get_weather` alongside the first two tools and add it to the dictionary.
4. Replace the system prompt:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, and campground in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails. 3. check_campsites for where to stay each night. Then write the itinerary: one section per day with trail, campsite, and how the forecast shaped the choice (put harder or more exposed hiking on the drier days).
```

5. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather.
```

**Why:** `data/mock-apis/weather.json` covers Glacier 2026-09-13 to 2026-09-18; the planted day is the 16th, rain showers at 49/33 with 0.7 precipitation chance and 18 mph wind, after two sunny days and one partly cloudy one.

**Check:** the itinerary puts the hard hike on the 14th or 15th, and the 16th gets something short or sheltered with a sentence saying why. Failing looks like the same three trails plus "expect rain on the 16th"; if the model never calls `get_weather`, fix your description first.

### Step 4: Add get_trail_conditions and ask for the washed-out bridge on trail-0117

**Do:**
1. Write `get_trail_conditions(trail_id)`. If `trail_id` is missing/blank, return `{"error": "trailId is required. Call search_trails first and use one of its ids."}` instead of throwing.
2. Open `data/condition-reports.jsonl` line by line, parsing each non-empty line as JSON (`id`, `trail_id`, `date`, `text`). Keep reports whose `trail_id` matches (case-insensitive). Sort by `date`, newest first. Keep the first 4.
3. If none matched, return `{"error": "No condition reports found for '<trail_id>'."}`. Otherwise return a JSON array of `{date, report}` objects (`report` = the line's `text`).
4. Declare `get_trail_conditions`:

```json
{
  "type": "function",
  "function": {
    "name": "get_trail_conditions",
    "description": "Get the most recent hiker-submitted condition reports for one trail, newest first. Call this for every trail you intend to recommend. Reports are where closures, washouts, and hazards show up; the catalog entry will not mention them.",
    "parameters": {
      "type": "object",
      "properties": {"trail_id": {"type": "string", "description": "A trail id returned by search_trails, for example 'trail-0117'."}},
      "required": ["trail_id"]
    }
  }
}
```

5. Replace the system prompt:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, campground, and condition in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails that fit the request. 3. get_trail_conditions for EVERY trail you intend to recommend, one call per trail, using the trail id returned by search_trails. If the newest reports for a trail mention a closure, a washout, a bridge that is out, or any other reason hikers are turning around, that trail is CLOSED. Do not schedule a day on a closed trail. Replace it with another trail from search_trails and state plainly, in the itinerary, that the original trail is closed and why. 4. check_campsites for where to stay each night. 5. request_permit once, only if a backcountry site or permit zone is involved. If you have not yet called search_trails and get_trail_conditions, your next move is a tool call, not prose. Then write the final itinerary: one section per day with trail, campsite, and how the forecast shaped the choice. End with the permit status.
```

6. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).
```

**Why:** the catalog entry for `trail-0117` says nothing about the bridge — only the condition reports do. 40 reports belong to `trail-0117`, and eight are the planted washout: five from 2026-06-18 to 2026-06-24 announcing the footbridge is gone, three July follow-ups saying it is still out.

**Check:** the model calls `get_trail_conditions` with `trail-0117` and the itinerary drops or flags the trail because the bridge is out, traceable to the tool result. Failures, worst last: never calling the tool and scheduling `trail-0117`; reading "the bridge is OUT" and scheduling it anyway; dropping it with an invented reason without calling the tool.

### Stretch goals

Pick any. Each one is already built in `complete/`.

- **The human gate on request_permit.** Add the fifth tool, `request_permit(park, zone, dates, group_size)`:

  ```json
  {
    "type": "function",
    "function": {
      "name": "request_permit",
      "description": "Submit a backcountry permit request. This files a request on the user's behalf, so call it once, at the end, after the plan is settled and the user has confirmed.",
      "parameters": {
        "type": "object",
        "properties": {
          "park": {"type": "string", "description": "Park name, for example 'Glacier National Park'."},
          "zone": {"type": "string", "description": "Permit zone, for example 'Lake McDonald / Sperry' or 'Many Glacier'."},
          "dates": {"type": "string", "description": "Trip dates, for example '2026-09-14 to 2026-09-16'."},
          "group_size": {"type": "integer", "description": "Number of people in the group. Zones cap out at 8 in Glacier."}
        },
        "required": ["park", "zone", "dates", "group_size"]
      }
    }
  }
  ```

  **Why this needs a gate:** filing a permit is the one action in this agent that cannot be undone, so the function never runs just because the model asked. Print the four arguments as one line, ask `File it? [y/N]`, and read the answer. On `y`/`yes`, return the `submit_response` object from `data/mock-apis/permits.json` (a fixture: `status: "submitted"`, `confirmation_id: "TRG-2026-091482"`, a message clarifying it's a request rather than a confirmed permit, and `next_steps`). On anything else, return this string instead — the model needs something to act on, not silence:

  ```text
  {"status":"cancelled","message":"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed."}
  ```

  Add a `--yes` flag that skips the question for demo runs. **Check:** `request_permit` never runs on the model's say-so, and a declined request ends with an itinerary that says no permit was filed, without retrying.

- **Nudge a model that stops early.** A small model often quits after one or two tools, then writes the plan from trails it never looked up. After the loop returns, compare the set of tools called against `get_weather`, `search_trails`, `get_trail_conditions`, and `check_campsites`. If any are missing, append a user message naming them and asking for the next call; run the loop again, up to 3 times. If every tool was called but the answer has no "day" in it, append one user message asking for the itinerary and run once more. Print `[nudge]` whenever this fires. **Check:** against `llama3.2`, `[nudge]` lines appear and the run still reaches all four tools, as in `reference-transcript.md` (verbatim console output of one `complete/` run for the step 2 request, plus a second run asking for a 2-day trip including `trail-0117`, showing the agent routing around the closed trail). Against `gpt-5.5` on Azure, no `[nudge]` line prints at all.

- **Make the tools forgiving.** Give every tool parameter a default. Let `get_trail_conditions` accept a trail name where it expects an id, by looking the name up in `data/trails.json`. Treat a `features` string as a one-item array. Return an error string on bad input instead of throwing; in the loop, catch a bad `arguments` JSON string and return `{"error": "Bad arguments for <name>: <reason>"}`. **Check:** a call with `trail_id` set to `"null"` or `"[insert trail IDs here]"` gets an error naming the valid ids back, and the run keeps going instead of crashing.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F10-dotnet.md`](dotnet/F10-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F10-python.md`](python/F10-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F10-typescript.md`](typescript/F10-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tool-definitions.json`: the five tools as JSON-schema function definitions: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The Python and TypeScript answer keys load this file directly; the .NET answer key declares the same tools from attributes on its functions. Read the descriptions before you write your own.
- `reference-transcript.md`: a real, unedited `llama3.2` run for the step 2 request, plus the closed-trail run, described under the nudge stretch goal. Use it to check the shape of your loop, not the wording.
- `expected-output.md`: the success checks for the lab steps plus the stretch goal, with the failure modes each one is meant to catch.

The tools read the data in this folder: `trails.json` (the full 200-trail catalog that features 04 and 06 use a slice of), `condition-reports.jsonl` (the full 500-report stream that feature 08's per-trail files are cut from), and the hand-written fixtures in `mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Each step above says what the file it opens looks like. None of it is generated by a script; it all ships with the workshop. Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` surfaces it.

One detail of `search_trails` matters more than it looks. Its `features` keywords match the trail's name as well as its feature tags. The tool returns at most eight trails. Avalanche Lake is the 27th Glacier trail in the catalog, and no feature tag on it contains "Avalanche". Without the name match, a request for "Avalanche Lake Trail" has no way to reach `trail-0117` through search. In a soak test against `gpt-5.5` with a tag-only search, the agent gave up on the trail in six runs out of ten. It said so honestly and planned around it. But the washed-out bridge never came up, so the lab step had nothing to teach. The model did what it could with the tool it had. Remember that when your own agent misbehaves. The tool is often the cheaper place to look first.

The loop is written out three ways: `dotnet/complete/Program.cs`, `python/complete/main.py`, and `typescript/complete/index.ts`. Read whichever one is closest to your language; the two without a framework are about thirty lines.
