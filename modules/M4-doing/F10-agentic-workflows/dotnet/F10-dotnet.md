# .NET Demo for 10 Agentic Workflows

Two console projects, both named `TripAgent`, both built on Microsoft.Extensions.AI, matching the demo script in [docs/slides/outlines/M4-doing.md](../../../../docs/slides/outlines/M4-doing.md):

- `starter/`: the demo's starting point. One `IChatClient`, one trip request, no tools. It produces a confident itinerary full of trails and lodges the model half-remembers, books nothing, and knows nothing about the washed-out bridge. Fluent and useless, which is the reason the feature exists.
- `complete/`: the finished demo. Five tools, function invocation, a confirmation gate before the permit, a step budget, and a printed line for every tool call as it happens.

## Model

Both read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT`. When all three are set they use Azure OpenAI, which is what runs on stage. When they are not, they print a note and fall back to Ollama `llama3.2` at `http://localhost:11434`, which is how they were developed and how the transcripts in `../data/` were captured.

## Running

```bash
cd complete
dotnet run -- --yes                  # the capstone request, permit gate auto-approved
dotnet run                           # same, but it stops and asks before filing the permit
dotnet run -- "Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail (trail-0117)."
cd ../starter && dotnet run          # the no-tools version of the same sentence
```

Any non-flag arguments are the request. Both projects resolve the data files relative to their own folder, so run them from `complete/` and `starter/`.

## The Five Tools

Ordinary static C# methods on the `Trailhead` class, described with `[Description]`, registered with `AIFunctionFactory.Create`, handed to `ChatOptions.Tools`, and looped by `UseFunctionInvocation`. Each method prints itself on entry, which is what makes the loop watchable.

| Tool | Reads | Why it matters |
| --- | --- | --- |
| `search_trails(park, features?, maxDifficulty?)` | `data/trails.json` | candidate trails, and the ids the other tools need |
| `get_weather(park)` | `data/mock-apis/weather.json` | the September 16 rain day that should reshape the plan |
| `get_trail_conditions(trailId)` | `data/condition-reports.jsonl` | the four newest hiker reports; this is where the `trail-0117` washout lives |
| `check_campsites(park)` | `data/mock-apis/campsites.json` | per-date availability, including the full backcountry sites |
| `request_permit(park, zone, dates, groupSize)` | `data/mock-apis/permits.json` | the irreversible action, so it sits behind the human gate |

Both guardrails are visible in `complete/Program.cs`: `MaximumIterationsPerRequest = 12` is the step budget, and `request_permit` prints its summary and waits for a yes before it does anything (`--yes` skips the prompt for demo runs).

## About llama3.2, and Why the Capstone Runs on Azure

`llama3.2` supports tool calling and it is genuinely useful for the offline path, but at 3B parameters it is not reliable enough to stake a live finale on. Observed across roughly two dozen offline runs of `complete`:

- The most common failure is stopping early. It makes one or two tool calls, then starts writing the itinerary and invents the trails it never looked up. Before any scaffolding was added, fewer than one run in five reached all four planning tools.
- It calls tools with junk arguments: `trailId` missing entirely, or `"null"`, or `"[insert trail IDs here]"`. A missing required argument used to crash the process outright.
- It sometimes emits a tool call as plain text in the message body (`{"name": "SearchTrails", "parameters": {...}}`) instead of as a tool call, so nothing executes.
- It occasionally reads the washout reports and schedules the closed trail anyway. Getting one clean run of the Avalanche Lake request took three tries, and that was after the instruction about closures was sharpened.

Two pieces of scaffolding in `complete/Program.cs` exist purely because of that, and both are labelled in the source. First, every tool parameter has a default and returns a helpful error instead of throwing, so a malformed call cannot kill the run. Second, a nudge loop: when the response comes back and a required tool has not been called yet, the app says which ones are missing and lets the loop continue, up to three times, plus one more turn asking for the itinerary if the model finished its calls and then went quiet. The `[nudge]` lines in the reference transcript are the app talking, not the model.

Even with all of that, the run captured in `../reference-transcript.md` was the good one out of a batch, and several later runs still fell apart. The scaffolding is there for the local model, not for the demo. Pointed at the workshop's Microsoft Foundry deployment (`gpt-4.1`), the same code sequenced the tools in the prescribed order on every run tried while building this (weather, search, a conditions check on every candidate, campsites, then the permit only when a backcountry site was involved), with zero `[nudge]` lines, and on the closed-trail request it checked `trail-0117` first, read the closure, and planned a partial hike to the creek with the closure stated plainly. The Python and TypeScript ports behaved the same way. If you see `[nudge]` lines against Azure, something changed. This is the concrete reason the feature card says model choice stops being negotiable here. A dropped or malformed tool call breaks the loop instead of gently degrading the answer.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F10-lab.md`](../F10-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: A Plan with No Tools

A plain chat completion, no tools, no loop. The itinerary is fluent, generic, ignores the washed-out bridge, and books nothing, because nothing here can reach the catalog, the weather feed, or the condition reports.

Run:

```bash
dotnet run
```

Check: A lovely three-day plan with zero tool calls. That is the reason this feature exists.

### Step 2: Two Tools and the Loop

This is lab step 1, the round-trip that `../http/azure.http` walks by hand. Write `search_trails` and `check_campsites` as ordinary functions over `../data/trails.json` and `../data/mock-apis/campsites.json`, load their definitions from `../data/tool-definitions.json` (the two entries you need), and write the loop: send the messages with the `tools` array, read the tool calls out of the reply, run them, append the results, repeat until the reply is prose. Give the loop a step budget; it is the only thing that stops a model that keeps deciding to call one more tool.

```csharp
// Microsoft.Extensions.AI runs the loop for you; the budget is the one setting to keep.
IChatClient client = new ChatClientBuilder(inner)
    .UseFunctionInvocation(configure: c => c.MaximumIterationsPerRequest = 12)
    .Build();
var options = new ChatOptions
{
    Tools = [AIFunctionFactory.Create(Trailhead.SearchTrails), AIFunctionFactory.Create(Trailhead.CheckCampsites)],
};
var response = await client.GetResponseAsync(messages, options);
// [Description] on each method and parameter is the model's only manual for the tool.
```

Run:

```bash
dotnet run
```

Check: Print each tool call as it happens. You should see `search_trails` and `check_campsites` fire, then an itinerary that names real trails (Trail of the Cedars, Iceberg Lake) and real campgrounds. Invented names mean a tool result did not reach the model.

### Step 3: Add Get_weather and Ask for a Trip on the Rain Day (lab step 2)

Write the function over `../data/mock-apis/weather.json`, add its definition to the tools array (write the JSON schema yourself before copying it from `tool-definitions.json`; the description is the model's only manual), and ask for September 14 to 16. The 16th is a rain day: 49/33, 70 percent, 18 mph.

```csharp
[Description("Get the multi-day weather forecast and advisories for a park.")]
public static string GetWeather([Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park")
{
    var all = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/weather.json"))!.AsObject();
    var entry = all.FirstOrDefault(kv => kv.Key.Contains(park.Split(' ')[0], StringComparison.OrdinalIgnoreCase));
    return entry.Value?.ToJsonString() ?? $"{{\"error\": \"No forecast available for '{park}'.\"}}";
}
// then add AIFunctionFactory.Create(Trailhead.GetWeather) to Tools
```

Run:

```bash
dotnet run
```

Check: `get_weather` is called and the forecast shapes the plan rather than decorating it: the hardest day lands on the 14th or 15th and the 16th gets something short or sheltered, with a sentence saying why. Compare the sample in `../expected-output.md`. Failing looks like the same three trails plus a line reading "expect rain on the 16th".

### Step 4: Add Get_trail_conditions and Ask for the Closed Trail (lab step 3)

The function reads `../data/condition-reports.jsonl` and returns the newest four reports for a trail id; make it return an error string, not throw, when the id is missing or malformed, and let it resolve a trail name too. Add it, then ask for a trip that includes Avalanche Lake Trail (`trail-0117`). The catalog says nothing about the bridge; only this tool does.

```csharp
[Description("Get the most recent hiker-submitted condition reports for a trail. Always check this before recommending a trail; reports surface closures and hazards such as washouts.")]
public static string GetTrailConditions([Description("The trail id from search_trails, e.g. 'trail-0117'.")] string? trailId = null)
{
    if (string.IsNullOrWhiteSpace(trailId))
        return "{\"error\": \"trailId is required. Call search_trails first and use one of its ids.\"}";
    var reports = File.ReadLines($"{DataDir}/condition-reports.jsonl").Select(l => JsonNode.Parse(l)!)
        .Where(r => string.Equals((string)r["trail_id"]!, trailId, StringComparison.OrdinalIgnoreCase))
        .OrderByDescending(r => (string)r["date"]!).Take(4)
        .Select(r => new { date = (string)r["date"]!, report = (string)r["text"]! }).ToArray();
    return reports.Length == 0 ? $"{{\"error\": \"No condition reports found for '{trailId}'.\"}}" : JsonSerializer.Serialize(reports);
}
```

Run:

```bash
dotnet run -- Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail (trail-0117).
```

Check: The model calls `get_trail_conditions` with `trail-0117`, gets the four bridge-is-out reports, and the itinerary drops or flags the trail for that reason. Three failure modes, in ascending order of embarrassment: it never calls the tool; it calls it, reads "bridge is OUT", and schedules the trail anyway (`llama3.2` does this; tighten the closure instruction in the system prompt, and if it persists that is your argument for a stronger model); it drops the trail but invents a reason without calling. The system prompt in `complete/` has the closure rule that fixed the second case.

### Step 5: Stretch: The Human Gate, and the Two Nudges (lab stretch goal)

`request_permit` is the one irreversible action, so it never runs on the model's say-so: print the summary, wait for a human yes, and only then return the tool result. Declining returns something the model can act on rather than silence. `complete/` also has two nudges for a local model that stops early: a check that every required tool was called (up to three re-prompts) and one turn asking for the itinerary if the tools are done and no plan was written. A frontier model should need neither; `[nudge]` lines mean the model is underpowered for the task.

```csharp
Console.WriteLine($"  [gate] About to file a permit request: {park}, zone '{zone}', {dates}, group of {groupSize}.");
Console.Write("  [gate] File it? [y/N] ");
var approved = Console.ReadLine()?.Trim().ToLowerInvariant() is "y" or "yes";
if (!approved)
    return "{\"status\": \"cancelled\", \"message\": \"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed.\"}";
return permits["submit_response"]!.ToJsonString();
```

Check: `request_permit` does not execute on the model's say-so; declining ends with an itinerary that says no permit was filed. Then compare your trace with `../reference-transcript.md`, and be precise about what this does and does not do: it prints the trace, it does not persist one. Writing that trace to durable storage next to feature 09's decisions log is what you would show a security team.
