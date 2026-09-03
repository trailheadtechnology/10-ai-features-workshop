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

Notes: An agent is a model given tools and a goal, running in a loop. Build the loop box by box. It starts with the trip request and the tool descriptions, sent together.
~
The model only ever returns one of two things: "call this tool with these arguments," or a final answer.
~
Your code runs the tool: weather, trails, conditions, campsites. Your code stays in charge of doing.
~
The result is appended to the conversation.
~
The model again, until it answers or hits the step budget. Trace one iteration with your finger: model, tool, result back in, model. The loop runs seven or eight times for this request.
~
Out comes an itinerary and a full trace.
~
The loop is your code. Not a framework, not the model.
~
The descriptions are the model's manual. They are the only thing it knows about your tools.
~
Guardrails are boxes on this line, spoken while pointing: a step budget of 12 so it cannot loop forever, a confirmation gate before `request_permit`, tools that return errors instead of throwing, and every call/args/result printed as the loop runs. Be precise: the demo prints the trace, it does not persist one. Shipping this means durable storage next to 09's approval record, and that pair is what you show your security team.

## RAG Fetches Before, an Agent Asks During

Flow (05 RAG): Question -> Your code retrieves context -> Model answers once -> Done
Flow (10 Agent): Request -> Model asks for what it needs -> Your code fetches it -> Model asks again, or answers -> Done, after N turns

- RAG: you pick the context, once · one call
- Agent: the model picks, repeatedly · N× the cost
- Can you name the context up front? Then RAG.

Notes: One row per advance. This is the "why not just RAG" slide.
RAG: your code retrieves the context, the model answers once, done.
~
Agent: the model asks for what it needs, your code fetches it, the model asks again or answers. Done after N turns.
~
RAG: you pick the context, once, and pay for one call.
~
Agent: the model picks, repeatedly, and you pay N times.
~
Can you name the context up front? Then RAG. The trip request needs weather before it can pick a trail and conditions before it can book, and nobody knows that order up front. That is the whole case for the loop.

## [demo] **DEMO** · F10 Agentic Workflows

## [static] F10 · Agentic Workflows · Demo: What to Watch

- ❌ Plain completion: lovely, generic, books nothing
- Five C# methods · descriptions = the manual
- Weather → trails → conditions → around the washout
- It must *discover*: Sept 16 rain · the bridge
- Campsites full → it adapts
- The permit waits for a human yes

Notes: ~7 min · Terminal 1 (Azure, `AZURE_OPENAI_DEPLOYMENT=gpt-5.5`) · `cd modules/M4-doing/F10-agentic-workflows/dotnet`
Files: `trails.json` = 200 trails · `condition-reports.jsonl` = 08's stream, the bridge is in it · `mock-apis/` = weather, campsites, permits as JSON
Flags: `--yes` = auto-approve the permit gate · any other words = your own request (default: 3 days in Glacier, Sept 14 to 16)
1. Before: `cd starter && dotnet run`. The same request, no tools. Lovely, generic, books nothing.
2. Show the five methods in `complete/Program.cs`. The descriptions are the manual.
3. After: `cd ../complete && dotnet run`. Narrate the trace. Weather, trails, conditions per trail, campsites. It stops at the permit gate; say yes.
4. `dotnet run -- "Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail (trail-0117)."`: it reads the bridge report and routes around it.
5. Point at the step budget and the trace. Printed, not persisted. Say so.
Cut: 4 if the first run already hit the bridge.

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
~
The line for your boss. On Azure the loop sequences all five tools from a single request with zero nudges, in all three languages, measured against the workshop's Foundry deployment; the contrast with the local counts is the argument. Row 10, and the one your board has already asked about.
~
Hard. If asked about cost: the highest of the ten. Weeks to months, frontier-model API costs, real design work on guardrails, confirmation gates, observability. Do one of 01 through 09 first.

---

## [static] Lab 4: Debrief

- How'd it go?				  Observations?			        Questions?

Notes: Ask two people what surprised them.
Prompts if quiet: whose agent checked the bridge, and whose scheduled it anyway? Row 10 is done — all ten rows are done. Closing next: the framework, and the pitch. Hand off to the closing deck.
