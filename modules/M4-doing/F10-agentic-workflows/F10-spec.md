# 10 · Agentic workflows (capstone)

Module 4: Doing · Runs on Azure OpenAI / AI Foundry · Double-length slot (about 60 minutes)

## The user problem

A Trailhead Guides user types: "Plan me a 3-day trip in Glacier for mid-September." Fulfilling that today means the user does everything themselves: check the weather forecast, search for trails, cross-reference current conditions, check campsite availability, and file a permit request in a separate system. Five tools, one afternoon, and the app helped with exactly one step. The user didn't want search results; they wanted the trip planned.

## The concept

An agent is an LLM given tools and a goal, running in a loop: the model decides which tool to call, your code executes it, the result goes back to the model, and it continues until the goal is met. Tool calling is the mechanism underneath, where you describe your functions (name, parameters, what they do) and the model responds with "call `check_weather` with `park=Glacier, dates=Sept 14-16`" instead of prose. Your code stays in charge of actually doing things; the model only ever chooses.

This is the capstone because it composes the day. The agent searches trails semantically (04), grounds itself in current condition reports (08's data), works with structured tool results (02's lesson in reverse), and pauses for human confirmation before filing the permit (09's policy, applied). It's also the feature where model choice stops being negotiable. A dropped or malformed tool call doesn't degrade the experience, it breaks the loop, so this feature runs on Azure OpenAI rather than gambling the finale on a local model. And because an agent acts instead of answering, guardrails are part of the design: a step budget so it can't loop forever, and a confirmation gate before anything irreversible.

## Demo outline (25 min, .NET)

1. Type the trip request into a plain chat completion first. You get a lovely generic itinerary that ignores the washed-out bridge and books nothing, which is the reason this feature exists.
2. Show the five tools as ordinary C# methods over the mock APIs in `data/mock-apis/` and the corpus: `search_trails`, `get_weather`, `check_campsites`, `request_permit`, and `get_trail_conditions`. Register them with Microsoft.Extensions.AI's function-calling support; the descriptions you write are the model's only manual. `get_trail_conditions` reads the same condition reports feature 08 mined, which is how the agent can find out a bridge is gone.
3. Run the agent and narrate the loop live as each call scrolls past: weather first, then trail search, then a conditions check that discovers the feature 08 bridge washout and routes around it. The payoff is watching the model sequence tools nobody ordered it to sequence.
4. Show the full trace, then break something on purpose: mark the campsites full and re-run. The agent adapts and proposes different dates.
5. The permit step hits the feature 09 gate: the agent pauses, presents the summary, and waits for a human yes before filing. Show the step budget in the loop while you're there.
6. Close on the trace as an artifact: every tool call, its arguments, and its result printed as the loop runs, so every decision is inspectable rather than mysterious. Be precise about what this demo does and doesn't do. It prints the trace; it does not persist one. Shipping this means writing that trace to durable storage alongside the approval record from feature 09, and that pair is what you show your security team when they ask whether this thing is safe.

## Lab spec (the capstone, any language)

*Everyone attempts this one; it is the only feature in [Module 4](../M4-overview.md) and the hands-on period runs about 35 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand to feel the mechanics, then extend a working agent with a new tool.
- **Input:** `lab/` provides the tool definitions as JSON, the mock API fixtures from `data/mock-apis/`, and a transcript of a complete agent run for reference.
- **How:** `lab/azure.http` (Azure OpenAI, key handed out in the room) contains the full round-trip as sequential requests: send the request with the `tools` array, read the tool call out of the response, then send the follow-up with the tool result. In an `.http` file, you are the loop, which is the best way to understand what agent frameworks hide.
- **Steps:**
  1. Run the provided two-tool round-trip (`search_trails`, `check_campsites`) by hand and watch the model choose, receive, and continue.
  2. Add `get_weather` to the tools array (its JSON definition is your job) and re-run with a weather-dependent request. Success check: the model calls your new tool and the final itinerary reflects the forecast (compare `lab/expected-output.md`).
  3. Ask for a trip on the trail with the washed-out bridge. Success check: the itinerary avoids or flags it.
- **Stretch goal:** add the human gate. Insert a confirmation step before `request_permit` gets executed, and only pass the tool result back after an explicit yes. If you're in a language with a loop already written, wire the whole thing end to end.

## Leadership beat

- **When to reach for this:** multi-step workflows that span systems, where users state a goal and then do the clicking themselves. Trip planning, onboarding, procurement, incident response, report assembly.
- **Rough cost & effort:** the highest of the ten. Weeks to months, frontier-model API costs, and real design work on guardrails, confirmation gates, and observability. Do one of features 01-09 first; do this when the goal-shaped problem is worth it.
- **The one-liner for your CTO:** "Users state the goal in one sentence, the software does the clicking, and a human still signs off on anything that matters."

This card is row 10 of the [decision framework](../../../docs/decision-framework.md).
