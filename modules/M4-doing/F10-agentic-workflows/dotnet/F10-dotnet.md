# .NET demo for 10 Agentic Workflows

Two console projects, both named `TripAgent`, both built on Microsoft.Extensions.AI, matching the demo outline in [../F10-spec.md](../F10-spec.md):

- `starter/`: the demo's starting point. One `IChatClient`, one trip request, no tools. It produces a confident itinerary full of trails and lodges the model half-remembers, books nothing, and knows nothing about the washed-out bridge. Fluent and useless, which is the reason the feature exists.
- `complete/`: the finished demo. Five tools, function invocation, a confirmation gate before the permit, a step budget, and a printed line for every tool call as it happens.

## Model

Both read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT`. When all three are set they use Azure OpenAI, which is what runs on stage. When they are not, they print a note and fall back to Ollama `llama3.2` at `http://localhost:11434`, which is how they were developed and how the transcripts in `../lab/` were captured.

## Running

```bash
cd complete
dotnet run -- --yes                  # the capstone request, permit gate auto-approved
dotnet run                           # same, but it stops and asks before filing the permit
dotnet run -- "Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail (trail-0117)."
cd ../starter && dotnet run          # the no-tools version of the same sentence
```

Any non-flag arguments are the request. Both projects resolve the data files relative to their own folder, so run them from `complete/` and `starter/`.

## The five tools

Ordinary static C# methods on the `Trailhead` class, described with `[Description]`, registered with `AIFunctionFactory.Create`, handed to `ChatOptions.Tools`, and looped by `UseFunctionInvocation`. Each method prints itself on entry, which is what makes the loop watchable.

| Tool | Reads | Why it matters |
| --- | --- | --- |
| `search_trails(park, features?, maxDifficulty?)` | `data/trails.json` | candidate trails, and the ids the other tools need |
| `get_weather(park)` | `data/mock-apis/weather.json` | the September 16 rain day that should reshape the plan |
| `get_trail_conditions(trailId)` | `data/condition-reports.jsonl` | the four newest hiker reports; this is where the `trail-0117` washout lives |
| `check_campsites(park)` | `data/mock-apis/campsites.json` | per-date availability, including the full backcountry sites |
| `request_permit(park, zone, dates, groupSize)` | `data/mock-apis/permits.json` | the irreversible action, so it sits behind the human gate |

Both guardrails are visible in `complete/Program.cs`: `MaximumIterationsPerRequest = 12` is the step budget, and `request_permit` prints its summary and waits for a yes before it does anything (`--yes` skips the prompt for demo runs).

## About llama3.2, and why the capstone runs on Azure

`llama3.2` supports tool calling and it is genuinely useful for the offline path, but at 3B parameters it is not reliable enough to stake a live finale on. Observed across roughly two dozen offline runs of `complete`:

- The most common failure is stopping early. It makes one or two tool calls, then starts writing the itinerary and invents the trails it never looked up. Before any scaffolding was added, fewer than one run in five reached all four planning tools.
- It calls tools with junk arguments: `trailId` missing entirely, or `"null"`, or `"[insert trail IDs here]"`. A missing required argument used to crash the process outright.
- It sometimes emits a tool call as plain text in the message body (`{"name": "SearchTrails", "parameters": {...}}`) instead of as a tool call, so nothing executes.
- It occasionally reads the washout reports and schedules the closed trail anyway. Getting one clean run of the Avalanche Lake request took three tries, and that was after the instruction about closures was sharpened.

Two pieces of scaffolding in `complete/Program.cs` exist purely because of that, and both are labelled in the source. First, every tool parameter has a default and returns a helpful error instead of throwing, so a malformed call cannot kill the run. Second, a nudge loop: when the response comes back and a required tool has not been called yet, the app says which ones are missing and lets the loop continue, up to three times, plus one more turn asking for the itinerary if the model finished its calls and then went quiet. The `[nudge]` lines in the reference transcript are the app talking, not the model.

Even with all of that, the run captured in `../lab/reference-transcript.md` was the good one out of a batch, and several later runs still fell apart. The scaffolding is there for the local model, not for the demo: a frontier model on Azure sequences these five tools from a single request, and the `[nudge]` lines should not appear at all. Check that when you first point the app at your deployment. This is the concrete reason the feature card says model choice stops being negotiable here. A dropped or malformed tool call breaks the loop instead of gently degrading the answer.
