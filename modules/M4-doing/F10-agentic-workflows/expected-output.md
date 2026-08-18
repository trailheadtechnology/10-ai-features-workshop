# What Passing Looks Like

Wording varies every run, and with an agent so does the order of the calls. The checks below are what has to be true. The samples come from real `llama3.2` runs of `dotnet/complete`; the full runs are in [reference-transcript.md](reference-transcript.md).

## Step 1: The Two-Tool Round-Trip

You send the request with `search_trails` and `check_campsites` in the `tools` array. The model answers with a tool call, not prose.

**Check:** the response has `finish_reason: "tool_calls"` and a `tool_calls` entry naming one of your two tools with a JSON `arguments` string, for example:

```json
{"name": "search_trails", "arguments": "{\"park\":\"Glacier National Park\"}"}
```

If you get an itinerary instead, the model never saw your tools. Look at where `tools` sits in the body.

**Check:** after you append the assistant message and a `role: "tool"` message carrying the result (with the matching `tool_call_id`), the follow-up response uses the trails you handed back. Real trail names such as Trail of the Cedars, Iceberg Lake Trail, or Bowman Lake Shoreline Trail should appear. Invented names mean the tool result did not reach the model.

## Step 2: Add Get_weather

Write the `get_weather` definition yourself, add it to the array, and ask for a trip on September 14-16.

**Check:** the model calls `get_weather`, and the forecast shapes the plan rather than decorating it. The Glacier fixture puts rain showers on September 16 (49/33, 70 percent chance, 18 mph wind) after two dry days, so a passing itinerary does something about the 16th. In the reference run:

> **Day 3 (September 16)** Trail: **Swiftcurrent Nature Trail**, 2.5 miles, easy. Forecast: Rain showers with a high temperature of 49°F and low of 33°F.
> Note: We replaced Lake McDonald Backcountry Site with Swiftcurrent Nature Trail due to the rain forecast.

and on day 1:

> Note: Due to expected rain showers on September 16, we chose to schedule this easier trail for the first day.

Passing looks like the hardest or most exposed day landing on the 14th or 15th, and the 16th getting something short, low, or sheltered, with a sentence saying why. Failing looks like the same three trails you would have gotten with no weather tool at all, plus a decorative line reading "expect rain on the 16th".

## Step 3: The Washed-Out Bridge

Ask for a trip that includes Avalanche Lake Trail (`trail-0117`).

**Check:** the model calls `get_trail_conditions` with `trail-0117` and receives reports like these, which are the newest four on file:

> "Trail is lovely as far as the creek, but with the bridge gone that is where the trip ends. Sign at the trailhead now says closed beyond mile 2." (2026-07-22)
> "Bridge is still out at the gorge. Rangers say no repair timeline yet, so plan on an out-and-back to the crossing." (2026-07-05)

**Check:** the itinerary avoids the trail or flags it explicitly. From the reference run:

> Note: The original requested trail, Avalanche Lake Trail (trail-0117), is closed due to the washed-out bridge at mile 2 and is no longer passable.

Three failure modes to watch for, in ascending order of embarrassment:

1. The model never calls `get_trail_conditions` and cheerfully schedules the trail. Your tool description is not telling it that conditions are where closures live.
2. It calls the tool, reads "the bridge is OUT", and schedules the trail anyway. Small models do this; llama3.2 did it on one of the runs behind this file. Tighten the instruction that a closure means the trail is out of the plan, and if it still happens, that is your argument for a stronger model.
3. It drops the trail but invents a reason ("closed for maintenance") without ever calling the tool. It guessed right, which is worse than guessing wrong, because it will guess again tomorrow.

## Stretch Goal: The Human Gate

**Check:** `request_permit` does not execute on the model's say-so. Your code prints the summary, waits for input, and only then runs the tool and returns the result. Declining should send back something the model can act on rather than silence, and the final answer should say no permit was filed.

**Check:** a step budget exists. `dotnet/complete` caps function-invocation iterations per request; in an `.http` file you are the budget, so decide up front how many round-trips you will do before you stop.
