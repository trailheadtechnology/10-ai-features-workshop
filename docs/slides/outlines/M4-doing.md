# Deck 4: Module 4, Doing (about 60 minutes)

One feature, double the time, and it uses everything from the rest of the day.

This file is the instructor script for the module. The bullets under each slide are what is on screen; the `Notes:` text and the "Demo script" steps under each "Demo: What to Watch" slide are what you say and click. The specs in `modules/` are written for attendees and no longer carry the demo steps.

Runsheet (lengths, not clock times):
- 10 min instructor: slides 1 to 11. The two diagram slides (6 and 7) and the demo (slide 8) are the bulk of it, so keep the tool loop on screen and let it scroll; slides 9 to 11 are one sentence each if the demo ran long.
- 50 min build: everyone does 10. Foundry key and the `gpt-5.5` deployment name on the board. Sticky check at about 20 min.
- Cut if behind: the break-it-on-purpose re-run with full campsites (step 4); leave the leadership card as one sentence.
- Last 5 min: debrief, then hand off to the closing deck.

---

## 1. Module 4: Doing

- The capstone
- 10 Agentic Workflows: "Plan me a 3-day trip in Glacier" becomes real tool calls
- Everyone attempts this one; there is no menu

## 2. 10 · The User Problem

- "Plan me a 3-day trip in Glacier for mid-September"
- Today the user does it all: check the forecast, search trails, cross-reference conditions, check campsites, file a permit in a separate system
- Five tools, one afternoon, and the app helped with exactly one step
- They didn't want search results; they wanted the trip planned

## 3. 10 · The Concept

- An agent is a model given tools and a goal, running in a loop
- The model decides which tool to call; your code executes it; the result goes back; repeat until the goal is met
- Tool calling underneath: describe your functions, and the model answers "call `check_weather` with `park=Glacier`" instead of prose
- Your code stays in charge of doing anything; the model only ever chooses

## 4. 10 · Why This Is the Capstone

- Searches trails semantically (04)
- Grounds itself in condition reports (08's data)
- Works with structured tool results (02 in reverse)
- Pauses for human confirmation before filing the permit (09's policy, applied)
- And it fails in every way the first nine fail, at once, with the failures compounding

## 5. 10 · Guardrails Are Part of the Design

- A step budget so it cannot loop forever (12 iterations in the demo)
- A confirmation gate before anything irreversible (`request_permit` waits for a human yes)
- Tools that return errors instead of throwing, so a malformed call cannot kill the run
- Every tool call, its arguments, and its result printed as the loop runs

Notes: Be precise: the demo prints the trace, it does not persist one. Shipping this means writing that trace to durable storage next to the approval record from 09. That pair is what you show your security team.

## 6. 10 · How the Loop Works

Flow: Trip request + tool descriptions -> Model: call a tool, or answer -> Your code runs the tool (weather, trails, conditions, campsites) -> Result appended to the conversation -> Model again, until it answers or hits the step budget -> Itinerary + a full trace

- The loop is in your code. The model only ever returns "call this with these arguments" or a final answer.
- Tools are ordinary methods; their descriptions are the model's manual
- Guardrails are boxes on this line: a step budget on the loop, an approval gate before `request_permit`

Notes: Trace one iteration with your finger: model, tool, result back in, model. Then say the loop runs seven or eight times for this request.

## 7. RAG Fetches Before, an Agent Asks During

Flow (05 RAG): Question -> Your code retrieves context -> Model answers once -> Done
Flow (10 Agent): Request -> Model asks for what it needs -> Your code fetches it -> Model asks again, or answers -> Done, after N turns

- RAG: you decide what the model reads, once, up front. Cheap, predictable, one call.
- Agent: the model decides what to fetch, in what order, and when to stop. Flexible, and N times the cost.
- Pick the left row whenever you can name the context in advance. Pick the right row when you can't.

Notes: This is the "why not just RAG" slide. The trip request needs weather before it can pick a trail and conditions before it can book, and nobody knows that order up front. That is the whole case for the loop.

## 8. 10 · Demo: What to Watch

- Plain chat completion first: a lovely generic itinerary that ignores the bridge and books nothing
- The five tools as ordinary C# methods over the mock APIs: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The descriptions are the model's only manual.
- Run the agent and narrate the loop: weather, then trails, then a conditions check that discovers the washout and routes around it
- Two facts the agent has to discover, not be told: September 16 is the rain day (49/33, 70% precip, 18 mph); Avalanche Lake's bridge is out, and only `get_trail_conditions` says so
- Break something: mark the campsites full, re-run, watch it adapt
- The permit hits the 09 gate: summary on screen, wait for a yes

Notes: This is most of the 10-minute instructor block, so let the tool loop scroll and narrate the calls as they land. Run it once against `gpt-5.5` on Foundry; the measured behavior (7 to 8 calls, zero nudges) is in `dotnet/F10-dotnet.md`.

Demo script (the demo outline that used to live in the spec, sized for the 10-minute block: step 4 is the first thing to skip if you are running long):
1. Type the trip request into a plain chat completion first. You get a lovely generic itinerary that ignores the washed-out bridge and books nothing, which is the reason this feature exists.
2. Show the five tools as ordinary C# methods over the mock APIs in `data/mock-apis/` and the corpus: `search_trails`, `get_weather`, `check_campsites`, `request_permit`, and `get_trail_conditions`. Register them with Microsoft.Extensions.AI's function-calling support; the descriptions you write are the model's only manual. `get_trail_conditions` reads the same condition reports feature 08 mined, which is how the agent can find out a bridge is gone.
3. Run the agent and narrate the loop live as each call scrolls past: weather first, then trail search, then a conditions check that discovers the feature 08 bridge washout and routes around it. The payoff is watching the model sequence tools nobody ordered it to sequence.
4. Show the full trace, then break something on purpose: mark the campsites full and re-run. The agent adapts and proposes different dates.
5. The permit step hits the feature 09 gate: the agent pauses, presents the summary, and waits for a human yes before filing. Show the step budget in the loop while you're there.
6. Close on the trace as an artifact: every tool call, its arguments, and its result printed as the loop runs, so every decision is inspectable rather than mysterious. Be precise about what this demo does and doesn't do: it prints the trace, and it does not persist one. Shipping this means writing that trace to durable storage alongside the approval record from feature 09, and that pair is what you show your security team when they ask whether this thing is safe.

## 9. 10 · Reliability Is the Whole Engineering Problem

- Across roughly two dozen local runs while building this, `llama3.2`:
  - called a tool with the literal argument `[insert trail IDs here]`
  - announced it had called every tool, then wrote nothing
  - emitted tool calls as plain text, so none executed
  - read the reports saying the bridge was out and scheduled the trail anyway
- Before scaffolding: fewer than one run in five reached all four planning tools
- A dropped or malformed tool call doesn't degrade the answer; it breaks the loop

Notes: These counts are in `dotnet/F10-dotnet.md` as measured fact. They are the honest answer to "should we ship an agent?" Yes: with a step budget, validated arguments, tools that return errors, a persisted trace, and a human approving anything irreversible. Which is to say, with feature 09.

## 10. 10 · Why This One Pays for a Frontier Model

- The only feature today where the workshop pays for a cloud model rather than running local
- The reason is the numbers on the previous slide
- On Azure the loop should sequence all five tools from a single request, with no nudges

Notes: Measured against the workshop's Microsoft Foundry deployment (gpt-4.1): every run sequenced weather, search, conditions on every candidate, campsites, then the permit only when needed, with zero nudges, in all three languages; the closed-trail request checked trail-0117 first and planned around the closure. Say that plainly; the contrast with the local counts on the previous slide is the argument.

## 11. 10 · Leadership Card

- When: multi-step workflows spanning systems where users state a goal and then do the clicking themselves. Trip planning, onboarding, procurement, incident response, report assembly.
- Cost: the highest of the ten. Weeks to months, frontier-model API costs, real design work on guardrails, confirmation gates, observability. Do one of 01 to 09 first.
- "Users state the goal in one sentence, the software does the clicking, and a human still signs off on anything that matters."

Notes: Row 10, and the one your board has already asked about.

## 12. Hands-On: 50 Minutes

- `http/azure.http`: you are the loop. Send the request with the tools array, read the tool call out of the response, send the follow-up with the tool result.
- Step 1: the two-tool round-trip (`search_trails`, `check_campsites`) is written out for you
- Step 2: write the `get_weather` definition yourself; the itinerary should do something about the 16th
- Step 3: ask for a trip on `trail-0117`; the itinerary should drop or flag it, for a reason that came from the tool
- Stretch: the human gate before `request_permit`
- `reference-transcript.md` is a complete known-good run to compare against when yours goes sideways. It will.

Notes: If you get an itinerary instead of a tool call on step 1, the model never saw your tools. Three failure modes on step 3, in ascending order of embarrassment: never calls conditions; calls it, reads "bridge is OUT", schedules anyway; drops the trail but invents a reason without calling. The third guessed right, which is worse.

## 13. Debrief

- Whose agent checked the bridge? Whose scheduled it anyway?
- Row 10 is done. All ten rows are done.
- Closing next: the decision framework, and how to pitch this.
