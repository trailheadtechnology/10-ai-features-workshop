# Lab 10: Agentic Workflows

*Everyone attempts this one; it is the only feature in [Module 4](../M4-overview.md) and the hands-on period runs about 50 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand, then extend a working agent with a new tool.
- **Input:** `data/tool-definitions.json`, the five tools (`search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`); `data/trails.json`, the trail catalog; `data/condition-reports.jsonl`, the hiker reports; `data/mock-apis/weather.json`, `campsites.json`, `permits.json`, canned results; `reference-transcript.md`, a complete run to compare against.
- **How:** chat completions with a `tools` array, against Azure OpenAI. In a code track your program runs the loop. In the HTTP track `http/azure.http` holds requests 1a, 1b, 1c, 2, 3, and the stretch, one per turn, with every tool result you hand back; you are the loop.
- **Model:** `gpt-5.5` on Azure. In the HTTP track the endpoint is filled in; paste the room key over `<KEY FROM INSTRUCTOR>`. Every code track reads `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (`gpt-5.5`), else falls back to `llama3.2`, much weaker here.

Every step below is one thing to make the program do. Each code track's `starter/` is a plain chat call with no tools. Edit it until it does every step. Compare against `complete/` when you get stuck. Your track's walkthrough (linked under Pick a Track below) shows how its chat client declares tools and reads tool calls back; the steps here stay in words that hold on every track. HTTP-track readers run the numbered requests in `http/azure.http` where a step names one. In the HTTP track you play the loop yourself: the tool results sit in the file as comments, and you paste them back in by hand.

### Step 0: Run the starter and read the plan with no tools

Run `starter/` as it is. It sends this request to the model as one user message. There are no tools and no loop:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

**Check:** a fluent three-day plan with zero tool calls. It names trails and campgrounds the model half-remembers. It checks no weather, reads no conditions, and books nothing. The rest of the lab closes that gap.

### Step 1: Write the first two tools as ordinary functions

Each tool is a plain function. It reads a file under `data/` and returns a JSON string. At the top of each one, print `[tool] ` plus the tool name and its arguments. That line is how you watch the loop run.

1. Write `search_trails(park, features, max_difficulty)`. Open `data/trails.json`, an array of trail objects with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features`, and `description`. This is the workshop's whole trail catalog: 200 trails across six parks (Acadia, Glacier, Great Smoky Mountains, Rocky Mountain, Yosemite, Zion), 45 of them in Glacier, `trail-0001` Sentinel Dome Loop through `trail-0200`. The 30-trail files in features 04 and 06 are copied record for record from this one. It ships with the workshop; no script builds it. The planted record is `trail-0117`, Avalanche Lake Trail, whose description mentions the footbridge crossing at the mouth of the gorge and says nothing about it being gone. Only the condition reports do.
2. Keep the trails whose `park` contains the `park` argument, ignoring case.
3. If `max_difficulty` was given, drop trails harder than it. The order is `easy`, then `moderate`, then `hard`.
4. If `features` was given, keep a trail only when at least one keyword appears in the trail's `name` or in one of its `features` tags, ignoring case. Do not skip the name match. Avalanche Lake Trail is the 27th Glacier trail in the catalog, and none of its feature tags contain "Avalanche". Without the name match, search can never reach it.
5. Stop after 8 matches. Return them as a JSON array without the `description` field.
6. Write `check_campsites(park)`. Open `data/mock-apis/campsites.json`, an object keyed by park name. Skip the `_comment` key. The file is a hand-written fixture standing in for a reservation system: three parks, and under each an array of campgrounds with `campground`, `type` (frontcountry or backcountry), `sites_available` keyed by date from 2026-09-13 to 2026-09-17, `amenities`, and a `note`. Glacier has four campgrounds. The planted fact, named in the file's own `_comment`, is that Sperry Chalet Area Sites shows 0 open sites on the 13th, 14th, and 15th, so a plan that sleeps there those nights ignored the tool.
7. Find the key that matches the park: the key contains the argument, or the argument contains the key, or the key contains the first word of the argument, ignoring case.
8. Return that entry as a JSON string. If nothing matches, return `{"error": "No campsite data for '<park>'."}`.

HTTP track: you do not write these functions. Their outputs are already pasted into requests 1b and 1c as the `content` of each `tool` message. Read them, then go to step 2. `http/azure.http` is six hand-written requests to the `gpt-5.5` deployment, one per turn of the loop (1a, 1b, 1c, 2, 3, and the stretch), each a full chat-completions body with the `messages` so far and the `tools` array. The tool results sitting in its comments are the real outputs of the workshop's tools over `data/`, so what you paste back is what a code track prints.

**Check:** `search_trails` with `{"park":"Glacier National Park"}` returns 8 trails, from `trail-0003` Trail of the Cedars to `trail-0037` Bowman Lake Shoreline Trail. `check_campsites` with the same park returns 4 campgrounds, and Sperry Chalet Area Sites shows 0 open sites on September 14 and 15. Both match the `content` strings in requests 1b and 1c of `http/azure.http`.

### Step 2: Write the loop and run the two-tool round-trip

1. Declare the two tools `search_trails` and `check_campsites` with the definitions in `data/tool-definitions.json` (your walkthrough shows how your track does it). That file is one object with a `_comment` and a `tools` array holding all five tools in the JSON-schema function shape that OpenAI-compatible chat APIs accept as the `tools` value: each entry has a `type` of `function` and a `function` with `name`, `description`, and `parameters`. The descriptions were written for this lab and are the model's only manual for each tool. The two definitions are:

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

2. Make a dictionary that maps each tool name to its function.
3. Start the `messages` list with a `system` message and a `user` message. The system prompt is:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, campgrounds, or availability. Every trail name and campground in your answer must have come back from a tool call in this conversation. Call search_trails and check_campsites before writing any part of the itinerary. If you have not yet called both, your next move is a tool call, not prose.
```

The user message is:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

4. Loop at most 12 times. That cap is the step budget. Nothing else stops a model that keeps asking for one more tool.
5. Each time through, send `messages` and the tool definitions through your track's chat client, with tool choice left automatic so the model decides. Read back the assistant message.
6. If the message carries no tool calls, print its text and stop. That text is the itinerary.
7. Otherwise, append the assistant message to `messages` exactly as it came back, tool calls included. Each tool call carries an `id`, the tool's name, and its arguments as a JSON string.
8. For each tool call, parse the arguments string into a dictionary. Look up the function by the tool's name. Call it with those arguments.
9. Append the result as a `tool` message tied to that call's `id`, with the string the function returned as its content. Then go back to 5.

HTTP track: send request 1a. Copy the assistant message out of the response into request 1b. Put its `id` over `CALL-ID`. Send 1b. Do the same in 1c with the second call over `CALL-ID-2`. If the model asked for `check_campsites` first, swap the two results.

**Check:** in the HTTP track, 1a returns `finish_reason: "tool_calls"` naming `search_trails` or `check_campsites`; 1c returns `finish_reason: "stop"` and an itinerary naming only trails from 1b and campgrounds from 1c. In a code track, one `[tool]` line prints for each of the two tools, then the itinerary. An itinerary at 1a means the model never saw your tools; any other name means a tool result did not reach it.

### Step 3: Add get_weather and plan around the rain day

1. Write `get_weather(park)`. Build it the same way as `check_campsites` in step 1, but read `data/mock-apis/weather.json`. If nothing matches, return `{"error": "No forecast available for '<park>'."}`. That file is a hand-written fixture standing in for a weather service, keyed by park like the campsites file. Each park has a `forecast` array, one object per date with `date`, `high_f`, `low_f`, `conditions`, `precip_chance`, and `wind_mph`, plus an `advisories` array of strings. Glacier covers 2026-09-13 to 2026-09-18; the planted day is the 16th, rain showers at 49/33 with a 0.7 chance of precipitation and 18 mph wind, after two sunny days and one partly cloudy one.
2. Write the `get_weather` tool definition yourself: a name, a description, and one `park` parameter. In the HTTP track it goes over `WRITE-GET-WEATHER-HERE` in request 2. Write yours first. Only then read the reference in `data/tool-definitions.json`:

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

3. Declare `get_weather` alongside the first two tools and add the function to the dictionary.
4. Replace the system prompt with this one:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, and campground in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails. 3. check_campsites for where to stay each night. Then write the itinerary: one section per day with trail, campsite, and how the forecast shaped the choice (put harder or more exposed hiking on the drier days).
```

5. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather.
```

HTTP track: send request 2. When the model calls `get_weather`, hand back the forecast from the comment under request 2. Use the matching `tool_call_id`. Then hand back the 1b and 1c results as the model asks for them.

**Check:** in the request 2 itinerary the hard hike lands on the 14th or 15th and the 16th, the rain day in `data/mock-apis/weather.json`, gets something short or sheltered, with a sentence saying why. Failing looks like the same three trails plus "expect rain on the 16th"; if the model never calls `get_weather`, fix your description first.

### Step 4: Add get_trail_conditions and ask for the washed-out bridge on trail-0117

1. Write `get_trail_conditions(trail_id)`. If `trail_id` is missing or blank, do not throw. Return `{"error": "trailId is required. Call search_trails first and use one of its ids."}` instead.
2. Open `data/condition-reports.jsonl`. Read it line by line. Parse each non-empty line as JSON. The fields are `id`, `trail_id`, `date`, and `text`. The file is the workshop's whole stream of hiker-submitted condition reports: 500 lines, `cr-0001` to `cr-0500`, across 122 trails, dated 2025-05-03 to 2026-07-25, most of them mud, ice, bugs, wildflowers, parking, and blowdown. It ships with the workshop and no script builds it. Feature 08's `reports-0117.jsonl` and `reports-0042.jsonl` are the `trail-0117` and `trail-0042` lines of this file, copied out in stream order. Forty reports belong to `trail-0117`, and eight of them are the planted washout: five from 2026-06-18 to 2026-06-24 announcing that the footbridge over the gorge is gone, and three July follow-ups saying it is still out.
3. Keep the reports whose `trail_id` equals the argument, ignoring case.
4. Sort them by `date`, newest first. Keep the first 4.
5. If none matched, return `{"error": "No condition reports found for '<trail_id>'."}`. Otherwise return a JSON array of objects with `date` and `report`. The `report` value is the line's `text`.
6. Declare `get_trail_conditions` with the definition below from `data/tool-definitions.json`. Add the function to the dictionary:

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
}
```

7. Replace the system prompt with this one:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, campground, and condition in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails that fit the request. 3. get_trail_conditions for EVERY trail you intend to recommend, one call per trail, using the trail id returned by search_trails. If the newest reports for a trail mention a closure, a washout, a bridge that is out, or any other reason hikers are turning around, that trail is CLOSED. Do not schedule a day on a closed trail. Replace it with another trail from search_trails and state plainly, in the itinerary, that the original trail is closed and why. 4. check_campsites for where to stay each night. 5. request_permit once, only if a backcountry site or permit zone is involved. If you have not yet called search_trails and get_trail_conditions, your next move is a tool call, not prose. Then write the final itinerary: one section per day with trail, campsite, and how the forecast shaped the choice. End with the permit status.
```

8. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).
```

The catalog entry for `trail-0117` says nothing about the bridge. Only the condition reports do.

HTTP track: send request 3. When the model searches, hand back the `["waterfall"]` search result from the comment under request 3. When it asks about `trail-0117`, hand back the conditions from the same comment. Hand back the weather and campsites from before. For any other trail id, hand back `{"error": "No condition reports found for 'trail-XXXX'."}`. Request 3 also carries `request_permit`. If the model calls it, hand back the cancelled result from the stretch goal. In a code track, the answer key in `complete/` takes the request as its argument, so you can run the same request against it: `dotnet run -- "<request>"`, `uv run main.py "<request>"`, or `npm run complete -- "<request>"`, with `Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).` as the request.

**Check:** the model calls `get_trail_conditions` with `trail-0117` and the itinerary drops or flags the trail because the bridge is out, traceable to the tool result. Failures, worst last: never calling the tool and scheduling `trail-0117`; reading "the bridge is OUT" and scheduling it anyway; dropping it with an invented reason without calling the tool.

### Stretch goals

Pick any. Each one is already built in `complete/`.

- **The human gate on request_permit.** Add the fifth tool, `request_permit(park, zone, dates, group_size)`, declared with this definition from `data/tool-definitions.json`:

  ```json
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

  Filing a permit is the one action in this agent that cannot be undone. So the function never runs just because the model asked. Print the four arguments as one line. Ask `File it? [y/N]` and read the answer from the console. On `y` or `yes`, return the `submit_response` object from `data/mock-apis/permits.json`. That file is a hand-written fixture standing in for a permit office. Its `availability` object lists permit zones per park (two backcountry zones for Glacier, `Lake McDonald / Sperry` and `Many Glacier`, each with a status per date, a `group_size_limit` of 8, and a `method`), and its `submit_response` is the one canned success payload the tool ever returns: `status` `submitted`, `confirmation_id` `TRG-2026-091482`, a message saying it is a request rather than a confirmed permit, and `next_steps`. Nothing real is filed. On anything else, return this string instead. The model needs something it can act on, not silence:

  ```text
  {"status":"cancelled","message":"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed."}
  ```

  Add a `--yes` flag that skips the question for demo runs. In the HTTP track, the stretch request in `http/azure.http` is the "yes" branch. Read the arguments the model chose. Send the request only after you have said yes out loud. If you decline, send the cancelled string as the tool content instead. **Check:** `request_permit` never runs on the model's say-so, and a declined request ends with an itinerary that says no permit was filed, without retrying. Every track's `complete/` asks `File it? [y/N]` inside its permit function, and `--yes` skips it.

- **Nudge a model that stops early.** A small model often quits after one or two tools. Then it writes the plan from trails it never looked up. After the loop returns, compare the set of tools that were called against `get_weather`, `search_trails`, `get_trail_conditions`, and `check_campsites`. If any are missing, append a user message that names them and asks for the next call. Run the loop again. Do this up to 3 times. If every tool was called but the answer has no "day" in it, append one user message asking for the itinerary and run the loop once more. Print `[nudge]` when either one fires. **Check:** against `llama3.2`, `[nudge]` lines appear and the run still reaches all four tools, as in `reference-transcript.md`. That file is the console output of one `complete/` run (the .NET build) against local `llama3.2` for the step 2 request, captured verbatim: every `[tool]` line in order, the truncated results as printed and then the full JSON for the first three, the `[nudge]` lines, and the itinerary. A second run at the end asks for a 2-day trip that includes `trail-0117` and shows the agent routing around the closed trail. Against `gpt-5.5` on Azure, no `[nudge]` line prints at all.

- **Make the tools forgiving.** Give every tool parameter a default. Let `get_trail_conditions` accept a trail name where it expects an id, by looking the name up in `data/trails.json`. Treat a `features` string as a one-item array. Return an error string on bad input instead of throwing. In the loop, catch a bad `arguments` JSON string and return `{"error": "Bad arguments for <name>: <reason>"}` as the tool result. **Check:** a call with `trail_id` set to `"null"` or `"[insert trail IDs here]"` gets an error naming the valid ids back, and the run keeps going instead of crashing.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F10-http.md`](http/F10-http.md) | the requests in `http/azure.http`, or a port of them in your language |
| .NET | [`dotnet/F10-dotnet.md`](dotnet/F10-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F10-python.md`](python/F10-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F10-typescript.md`](typescript/F10-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tool-definitions.json`: the five tools as JSON-schema function definitions: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The Python and TypeScript answer keys load this file directly; the .NET answer key declares the same tools from attributes on its functions. Read the descriptions before you write your own.
- `reference-transcript.md`: a real, unedited `llama3.2` run for the step 2 request, plus the closed-trail run, described under the nudge stretch goal. Use it to check the shape of your loop, not the wording.
- `expected-output.md`: the success checks for the lab steps plus the stretch goal, with the failure modes each one is meant to catch.

The tools read the data in this folder: `trails.json` (the full 200-trail catalog that features 04 and 06 use a slice of), `condition-reports.jsonl` (the full 500-report stream that feature 08's per-trail files are cut from), and the hand-written fixtures in `mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Each step above says what the file it opens looks like. None of it is generated by a script; it all ships with the workshop. Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` surfaces it.

One detail of `search_trails` matters more than it looks. Its `features` keywords match the trail's name as well as its feature tags. The tool returns at most eight trails. Avalanche Lake is the 27th Glacier trail in the catalog, and no feature tag on it contains "Avalanche". Without the name match, a request for "Avalanche Lake Trail" has no way to reach `trail-0117` through search. In a soak test against `gpt-5.5` with a tag-only search, the agent gave up on the trail in six runs out of ten. It said so honestly and planned around it. But the washed-out bridge never came up, so the lab step had nothing to teach. The model did what it could with the tool it had. Remember that when your own agent misbehaves. The tool is often the cheaper place to look first.

If you would rather run the loop from your own language than from an `.http` file, `data/tool-definitions.json` gives you the `tools` value for the request bodies. The loop is written out three ways: `dotnet/complete/Program.cs`, `python/complete/main.py`, and `typescript/complete/index.ts`. Read whichever one is closest to your language; the two without a framework are about thirty lines. The tool results in `http/azure.http` are still the ones to hand back.
