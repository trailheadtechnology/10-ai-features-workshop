# Deck 4: Module 4, Doing (about 60 minutes)

One feature, and it uses everything from the rest of the day. Same demo-centric rhythm, single pass: divider, the two concept diagrams (an agent loop is genuinely complex), DEMO, the lab everyone does, the leadership card. Everything J. says lives in Notes; the demo script lives under "What to Watch".

Runsheet (lengths, not clock times):
- 10 min instructor: the two diagram slides and the demo are the bulk; the demo scrolling its tool loop is most of it.
- 50 min build: everyone does 10; there is no menu. Foundry key and the `gpt-5.5` deployment name on the board. Sticky check at about 20 min.
- Cut if behind: the break-it-on-purpose re-run with full campsites (step 4); say the leadership card in one sentence.
- Last 5 min: debrief, then hand off to the closing deck.

---

## [deck-title] Module 4: Doing

Notes: The capstone. "Plan me a 3-day trip in Glacier" becomes real tool calls. Everyone attempts this one.

## This Module

Icon (robot): F10 Agentic Workflows

Notes: The capstone. Everyone builds it, no menu, and it uses F02, F04, F08, and F09.
The user problem, spoken: "Plan me a 3-day trip in Glacier for mid-September." Today the user checks the forecast, searches trails, cross-references conditions, checks campsites, and files a permit in a separate system — five tools, one afternoon, and the app helped with exactly one step. They didn't want search results; they wanted the trip planned.
Why it's the capstone: it searches trails semantically (04), grounds itself in 08's condition reports, consumes structured tool results (02 in reverse), and pauses for a human before the permit (09). And it fails every way the first nine fail, at once, compounding.

## [define] A model that **picks which tools to call**, in a loop, until the job is done.

Icon (robot): F10

Notes: The textbook version: an agent is given a goal and tools; it chooses an action, your code executes it, it observes the result, and it repeats until the goal is met or a budget runs out. Reason, act, observe, repeat.

## F10 · How the Loop Works

Flow: Trip request + tool descriptions -> Model: call a tool, or answer -> Your code runs the tool (weather, trails, conditions, campsites) -> Result appended to the conversation -> Model again, until it answers or hits the step budget -> Itinerary + a full trace

- The loop is your code
- The descriptions are the model's manual
- Guardrails are boxes on this line

Notes: An agent is a model given tools and a goal, running in a loop. The model only ever returns "call this with these arguments" or a final answer; your code stays in charge of doing. Build the loop box by box, then trace one iteration with your finger: model, tool, result back in, model. The loop runs seven or eight times for this request.
The guardrails, spoken while pointing: a step budget of 12 so it cannot loop forever, a confirmation gate before `request_permit`, tools that return errors instead of throwing, and every call/args/result printed as the loop runs. Be precise: the demo prints the trace, it does not persist one — shipping this means durable storage next to 09's approval record, and that pair is what you show your security team.

## RAG Fetches Before, an Agent Asks During

Flow (05 RAG): Question -> Your code retrieves context -> Model answers once -> Done
Flow (10 Agent): Request -> Model asks for what it needs -> Your code fetches it -> Model asks again, or answers -> Done, after N turns

- RAG: you pick the context, once · one call
- Agent: the model picks, repeatedly · N× the cost
- Can you name the context up front? Then RAG.

Notes: One row per advance. This is the "why not just RAG" slide. The trip request needs weather before it can pick a trail and conditions before it can book, and nobody knows that order up front. That is the whole case for the loop.

## [demo] **DEMO** · F10 Agentic Workflows

## [static] F10 · Agentic Workflows · Demo: What to Watch

- ❌ Plain completion: lovely, generic, books nothing
- Five C# methods · descriptions = the manual
- Weather → trails → conditions → around the washout
- It must *discover*: Sept 16 rain · the bridge
- Campsites full → it adapts
- The permit waits for a human yes

Notes: The five tools: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The two facts it has to discover, not be told: September 16 is the rain day (49/33, 70% precip, 18 mph), and Avalanche Lake's bridge is out — only `get_trail_conditions` says so.
Run it against `gpt-5.5` on Foundry; measured behavior (7 to 8 calls, zero nudges) is in `dotnet/F10-dotnet.md`. Soak-tested 2026-09-03, 10 of 10 on both the default and the washout request. Two things the room may see: the rain day sometimes gets a moderate hike rather than an easy one (3 of 10 runs, the hard hike still lands on a dry day; say "it shaped the plan, it didn't perfect it"), and `search_trails` keywords match trail names on purpose, because Avalanche Lake sits past the 8-result cap and without the name match the agent cannot reach it. If the agent ever says it could not find the trail, the search tool is the reason, and that makes a better lesson than the bridge. The reliability story, spoken over the scrolling trace: local `llama3.2` across ~two dozen runs called a tool with the literal argument "[insert trail IDs here]", announced it had called every tool then wrote nothing, emitted tool calls as plain text so none executed, and read "bridge is OUT" and scheduled the trail anyway. Fewer than one run in five reached all four planning tools before scaffolding. That is why this is the one feature of the day that pays for a frontier model, and why the guardrails aren't optional: a malformed call breaks the loop, not the answer.

Demo script (~7 min; step 4 is the first cut):
1. Type the trip request into a plain chat completion first. You get a lovely generic itinerary that ignores the washed-out bridge and books nothing, which is the reason this feature exists.
2. Show the five tools as ordinary C# methods over the mock APIs in `data/mock-apis/` and the corpus: `search_trails`, `get_weather`, `check_campsites`, `request_permit`, and `get_trail_conditions`. Register them with Microsoft.Extensions.AI's function-calling support; the descriptions you write are the model's only manual. `get_trail_conditions` reads the same condition reports feature 08 mined, which is how the agent can find out a bridge is gone.
3. Run the agent and narrate the loop live as each call scrolls past: weather first, then trail search, then a conditions check that discovers the feature 08 bridge washout and routes around it. The payoff is watching the model sequence tools nobody ordered it to sequence.
4. Show the full trace, then break something on purpose: mark the campsites full and re-run. The agent adapts and proposes different dates.
5. The permit step hits the feature 09 gate: the agent pauses, presents the summary, and waits for a human yes before filing. Show the step budget in the loop while you're there.
6. Close on the trace as an artifact: every tool call, its arguments, and its result printed as the loop runs, so every decision is inspectable rather than mysterious. Be precise about what this demo does and doesn't do: it prints the trace, and it does not persist one. Shipping this means writing that trace to durable storage alongside the approval record from feature 09, and that pair is what you show your security team when they ask whether this thing is safe.

## [static] Lab 4: ~50 Minutes · Everyone

- `modules/M4-doing/F10-agentic-workflows/`
- `http/azure.http` — **you are the loop**
- Step 1: the two-tool round-trip, written for you
- Step 2: add `get_weather` — do something about the 16th
- Step 3: `trail-0117` — dropped, for a tool-given reason
- Stretch: the human gate before `request_permit`
- `reference-transcript.md` when it goes sideways. It will.

Notes: Send the request with the tools array, read the tool call out of the response, send the follow-up with the tool result. Step 1's round-trip is `search_trails` + `check_campsites`.
If you get an itinerary instead of a tool call on step 1, the model never saw your tools. Three failure modes on step 3, in ascending order of embarrassment: never calls conditions; calls it, reads "bridge is OUT", schedules anyway; drops the trail but invents a reason without calling. The third guessed right, which is worse.

## F10 · Agentic Workflows · Leadership Card

- **When:** Users state a goal, then do all the clicking themselves, across systems.
- **Think:** "Users state the goal in one sentence, the software does the clicking, and a human still signs off on anything that matters."

Difficulty: hard

Notes: Trip planning, onboarding, procurement, incident response, report assembly.
If asked about cost: the highest of the ten — weeks to months, frontier-model API costs, real design work on guardrails, confirmation gates, observability. Do one of 01–09 first.
On Azure the loop sequences all five tools from a single request with zero nudges, in all three languages — measured against the workshop's Foundry deployment; the contrast with the local counts is the argument. Row 10, and the one your board has already asked about.

---

## [static] Lab 4: Debrief

- How'd it go?				  Observations?			        Questions?

Notes: Ask two people what surprised them.
Prompts if quiet: whose agent checked the bridge, and whose scheduled it anyway? Row 10 is done — all ten rows are done. Closing next: the framework, and the pitch. Hand off to the closing deck.
