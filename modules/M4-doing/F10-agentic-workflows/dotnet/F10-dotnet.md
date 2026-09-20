<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 10: Agentic Workflows (.NET)

*You are on the .NET track. Other tracks: [Python](../python/F10-python.md), [TypeScript](../typescript/F10-typescript.md). Lab overview: [F10-lab.md](../F10-lab.md).*

**The User Problem:** A Trailhead Guides user types: "Plan me a 3-day trip in Glacier for mid-September." Fulfilling that today means the user does everything themselves: check the weather forecast, search for trails, cross-reference current conditions, check campsite availability, and file a permit request in a separate system. That is five tools and one afternoon, and the app helped with exactly one step. The user didn't want search results; they wanted the trip planned.

*Everyone attempts this one; it is the only feature in [Module 4](../../M4-overview.md) and the hands-on period runs about 50 minutes. The lab ships a transcript of a complete successful run, so when your own agent goes sideways you have a known-good reference to compare against rather than guessing.*

- **Goal:** run one tool-calling round-trip by hand, then extend a working agent with a new tool.
- **Input:** `data/tool-definitions.json`, the five tools (`search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`); `data/trails.json`, the trail catalog; `data/condition-reports.jsonl`, the hiker reports; `data/mock-apis/weather.json`, `campsites.json`, `permits.json`, canned results; `reference-transcript.md`, a complete run to compare against.
- **How:** chat completions with a `tools` array, against Azure OpenAI.
- **Model:** `gpt-5.5` on Azure. Every track reads `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (`gpt-5.5`), else falls back to `llama3.2`, much weaker here.

## The Concept

An agent is an LLM given tools and a goal, running in a loop: the model decides which tool to call, your code executes it, the result goes back to the model, and it continues until the goal is met. Tool calling is the mechanism underneath, where you describe your functions (name, parameters, what they do) and the model responds with "call `check_weather` with `park=Glacier, dates=Sept 14-16`" instead of prose. Your code stays in charge of actually doing things; the model only ever chooses.

This is the capstone because it composes the day. The agent searches trails semantically (04), grounds itself in current condition reports (08's data), works with structured tool results (02's lesson in reverse), and pauses for human confirmation before filing the permit (09's policy, applied). It's also the feature where model choice stops being negotiable. A dropped or malformed tool call doesn't degrade the experience, it breaks the loop, so this feature runs on Azure OpenAI rather than gambling the finale on a local model. And because an agent acts instead of answering, guardrails are part of the design: a step budget so it can't loop forever, and a confirmation gate before anything irreversible.

Every step below is one thing to make the program do. The `starter/` is a plain chat call with no tools. Edit it until it does every step. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, the deployment name the feature uses, and the key handed out in the room) and the program uses Azure OpenAI. Leave them unset and it prints a note and falls back to Ollama `llama3.2`, which is how the transcript in [`reference-transcript.md`](../reference-transcript.md) was captured.

Two console projects, both named `TripAgent`, both built on Microsoft.Extensions.AI, matching the demo from the module's slides:

- `starter/`: the demo's starting point. One `IChatClient`, one trip request, no tools.
- `complete/`: the finished demo. Five tools, function invocation, a confirmation gate before the permit, a step budget, and a printed line for every tool call as it happens.

Any non-flag arguments are the request. Both projects read data through `const string DataDir = "../../data";`, which is relative to the project folder under `dotnet run`, so run them from `starter/` and `complete/`. From `complete/`:

```bash
cd complete
dotnet run -- --yes                  # the capstone request, permit gate auto-approved
dotnet run                           # same, but it stops and asks before filing the permit
dotnet run -- "Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail (trail-0117)."
cd ../starter && dotnet run          # the no-tools version of the same sentence
```

### About llama3.2, and Why the Capstone Runs on Azure

`llama3.2` supports tool calling and it is genuinely useful for the offline path, but at 3B parameters it is not reliable enough to stake a live finale on. Observed across roughly two dozen offline runs of `complete`:

- The most common failure is stopping early. It makes one or two tool calls, then starts writing the itinerary and invents the trails it never looked up. Before any scaffolding was added, fewer than one run in five reached all four planning tools.
- It calls tools with junk arguments: `trailId` missing entirely, or `"null"`, or `"[insert trail IDs here]"`. A missing required argument used to crash the process outright.
- It sometimes emits a tool call as plain text in the message body (`{"name": "SearchTrails", "parameters": {...}}`) instead of as a tool call, so nothing executes.
- It occasionally reads the washout reports and schedules the closed trail anyway. Getting one clean run of the Avalanche Lake request took three tries, and that was after the instruction about closures was sharpened.

Two pieces of scaffolding in `complete/Program.cs` exist purely because of that, and both are labelled in the source. First, every tool parameter has a default and returns a helpful error instead of throwing, so a malformed call cannot kill the run. Second, a nudge loop: when the response comes back and a required tool has not been called yet, the app says which ones are missing and lets the loop continue, up to three times, plus one more turn asking for the itinerary if the model finished its calls and then went quiet. The `[nudge]` lines in the reference transcript are the app talking, not the model.

Even with all of that, the run captured in [`reference-transcript.md`](../reference-transcript.md) was the good one out of a batch, and several later runs still fell apart. The scaffolding is there for the local model, not for the demo. Pointed at Foundry, the same code sequenced the tools in the prescribed order on every run tried while building this (weather, search, a conditions check on every candidate, campsites, then the permit only when a backcountry site was involved), with zero `[nudge]` lines, and on the closed-trail request it checked `trail-0117` first, read the closure, and planned a partial hike to the creek with the closure stated plainly. The Python and TypeScript ports behaved the same way. That held on both of the workshop's deployments, `gpt-4.1` in the first soak and `gpt-5.5` since; the feature runs on `gpt-5.5`, and the run counts in [`expected-output.md`](../expected-output.md) are from it. If you see `[nudge]` lines against Azure, something changed. This is the concrete reason the feature card says model choice stops being negotiable here. A dropped or malformed tool call breaks the loop instead of gently degrading the answer.

### Step 0: Run the starter and read the plan with no tools

**Do:** run `starter/` as it is. It sends this request as one user message, no tools, no loop:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

```bash
dotnet run
```

The starter already does this. It picks the request (your command-line words, or the default sentence) and sends it in one call with no tools:

```csharp
var request = args.Length > 0
    ? string.Join(" ", args)
    : "Plan me a 3-day trip in Glacier National Park for September 14-16.";
```

```csharp
var response = await client.GetResponseAsync(
    $"""
    You are the trip planner for Trailhead Guides, a hiking app.

    {request}
    """);

Console.WriteLine(response.Text);
```

**Check:** a fluent three-day plan with zero tool calls. The model is working from memory, so the trails and campgrounds it names may not exist, and nothing in the plan reflects the weather for September 14-16 or whether those trails are open.

### Step 1: Write the first two tools as ordinary functions

**Do:** each tool is a plain function that reads a file under `data/` and returns a JSON string. Make them static methods on a `Trailhead` class with `const string DataDir = "../../data";` at the top. That path is relative to the project folder, which is the working directory under `dotnet run`, so run from `starter/` rather than launching the built binary. At the top of each, print `[tool] ` plus the tool name and its arguments. That line is how you watch the loop run. The model will see each tool under its C# method name (`SearchTrails`, not `search_trails`), so print the snake_case name by hand; the `Narrate` helper below does it.

Add these to the `using` lines at the top of `Program.cs`. `[Description]` comes from the first, `JsonSerializer` from the second, `JsonNode` from the third:

```csharp
using System.ComponentModel;
using System.Text.Json;
using System.Text.Json.Nodes;
```

The class goes at the very bottom of `Program.cs`, below the closing `}` of `CreateChatClient`. C# requires the top-level statements (the code that runs) to come first and any types to come after them. `static` means you call its members through the class name, as `Trailhead.SearchTrails`, without creating an object. `AutoApprovePermits`, `Called`, and `LastResultIds` are used by the stretch goals; declaring them now costs nothing.

```csharp
static class Trailhead
{
    // Relative to the project folder, not the build output, so run with
    // `dotnet run` from complete/ rather than launching the binary directly.
    const string DataDir = "../../data";
    public static bool AutoApprovePermits;
    public static readonly HashSet<string> Called = [];
    public static readonly List<string> LastResultIds = [];

    static readonly JsonSerializerOptions Pretty = new() { WriteIndented = false };
}
```

Every tool prints its `[tool]` line and its result preview through two small helpers. Put them inside the class, just above its closing `}`. `JsonSerializer.Serialize(args)` turns an anonymous object such as `new { park }` into `{"park":"..."}`, which is how the arguments get printed with their snake_case names:

```csharp
    static void Narrate(string tool, object args)
    {
        Called.Add(tool);
        Console.WriteLine($"[tool] {tool} {JsonSerializer.Serialize(args)}");
    }

    static string Result(string json)
    {
        var preview = json.Length > 120 ? json[..120] + "..." : json;
        Console.WriteLine($"  [result] {preview}");
        return json;
    }
```

Each tool method in items 1 and 2 also goes inside the class, above `Narrate`.

1. Write `search_trails(park, features, max_difficulty)`. Open `data/trails.json` (200 trails across six parks, 45 in Glacier). Keep trails whose `park` contains the `park` argument (case-insensitive). If `max_difficulty` was given, drop trails harder than it (`easy` < `moderate` < `hard`). If `features` was given, keep a trail only when at least one keyword appears in its `name` **or** its `features` tags. Do not skip the name match. Stop after 8 matches. Return them as a JSON array without the `description` field.

   In C# the parameters become `string park`, `string[]? features`, and `string? maxDifficulty`, each with a default so a call with a missing argument still runs. The `[Description]` text in front of the method and each parameter is what the model reads (step 2 explains). The body parses the file with `JsonNode`, filters with LINQ `Where`, stops with `Take(8)`, and uses `Select(t => new { ... })` to copy every field except `description` into a new anonymous object. The `rank` lambda turns `easy`/`moderate`/`hard` into 0/1/2 so difficulties can be compared:

   ```csharp
       [Description("Search the trail catalog. Returns matching trails with id, name, park, distance, elevation, difficulty, and features.")]
       public static string SearchTrails(
           [Description("Park name, e.g. 'Glacier National Park'. Partial names like 'Glacier' work.")] string park = "Glacier National Park",
           [Description("Optional keywords matched against each trail's features and its name, e.g. ['lake', 'waterfall'] or ['Avalanche Lake'].")] string[]? features = null,
           [Description("Optional maximum difficulty: 'easy', 'moderate', or 'hard'.")] string? maxDifficulty = null)
       {
           Narrate("search_trails", new { park, features, max_difficulty = maxDifficulty });

           var trails = JsonNode.Parse(File.ReadAllText($"{DataDir}/trails.json"))!.AsArray();
           var rank = (string d) => d switch { "easy" => 0, "moderate" => 1, _ => 2 };
           var maxRank = maxDifficulty is null ? 2 : rank(maxDifficulty.ToLowerInvariant());

           var matches = trails
               .Where(t => ((string)t!["park"]!).Contains(park, StringComparison.OrdinalIgnoreCase))
               .Where(t => rank((string)t!["difficulty"]!) <= maxRank)
               .Where(t => features is null || features.Length == 0 || features.Any(f =>
                   ((string)t!["name"]!).Contains(f, StringComparison.OrdinalIgnoreCase) ||
                   t!["features"]!.AsArray().Any(x => ((string)x!).Contains(f, StringComparison.OrdinalIgnoreCase))))
               .Take(8)
               .Select(t => new
               {
                   id = (string)t!["id"]!,
                   name = (string)t["name"]!,
                   park = (string)t["park"]!,
                   distance_mi = (double)t["distance_mi"]!,
                   elevation_ft = (int)t["elevation_ft"]!,
                   difficulty = (string)t["difficulty"]!,
                   features = t["features"]!.AsArray().Select(x => (string)x!).ToArray(),
               });

           var found = matches.ToArray();
           LastResultIds.Clear();
           LastResultIds.AddRange(found.Select(t => t.id));
           return Result(JsonSerializer.Serialize(found, Pretty));
       }
   ```

2. Write `check_campsites(park)`. Open `data/mock-apis/campsites.json` (an object keyed by park name, skip the `_comment` key). Find the key that matches the park: the key contains the argument, the argument contains the key, or the key contains the argument's first word, ignoring case. Return that entry as a JSON string, or `{"error": "No campsite data for '<park>'."}` if nothing matches.

   `all` is the whole file as a JSON object; `FirstOrDefault` walks its key/value pairs and stops at the first key that matches one of the three rules. The `_comment` key never contains a park name, so it is never picked. If nothing matches, `entry.Value` is `null` and the error string goes back instead:

   ```csharp
       [Description("Check campground availability in a park. Returns campgrounds with open sites per date, type (frontcountry or backcountry), and notes.")]
       public static string CheckCampsites(
           [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park")
       {
           Narrate("check_campsites", new { park });

           var all = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/campsites.json"))!.AsObject();
           var entry = all.FirstOrDefault(kv =>
               kv.Key.Contains(park, StringComparison.OrdinalIgnoreCase) ||
               park.Contains(kv.Key, StringComparison.OrdinalIgnoreCase) ||
               kv.Key.Contains(park.Split(' ')[0], StringComparison.OrdinalIgnoreCase));

           return Result(entry.Value is null
               ? $"{{\"error\": \"No campsite data for '{park}'.\"}}"
               : entry.Value.ToJsonString(Pretty));
       }
   ```

**Why:** the planted record is `trail-0117`, Avalanche Lake Trail, whose catalog description mentions the footbridge crossing at the gorge's mouth and says nothing about it being gone. Only the condition reports (step 4) do. Avalanche Lake Trail is the 27th Glacier trail in the catalog and none of its feature tags contain "Avalanche", so without the name match in `search_trails`, a search for it can never reach it. Similarly, `check_campsites`'s planted fact (named in the fixture's own `_comment`) is that Sperry Chalet Area Sites shows 0 open sites on the 13th, 14th, and 15th, so a plan that sleeps there those nights ignored the tool.

**Check:** `search_trails` with `{"park":"Glacier National Park"}` returns 8 trails, from `trail-0003` Trail of the Cedars to `trail-0037` Bowman Lake Shoreline Trail. `check_campsites` with the same park returns 4 campgrounds, and Sperry Chalet Area Sites shows 0 open sites on September 14 and 15.

### Step 2: Write the loop and run the two-tool round-trip

**Do:**
1. Declare the two tools. Instead of loading `data/tool-definitions.json`, put the same descriptions in `[Description]` attributes on each method and each parameter; they are the model's only manual for the tool. The model sees parameters under their C# names (`maxDifficulty`, `trailId`). These are the two definitions from the file whose wording to reuse:

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

   No new code for item 1: the `[Description]` attributes you wrote in step 1 are the declaration. The one above `public static string SearchTrails(` is the tool's `description`; the one in front of each parameter is that parameter's `description`.

2. Register each method with `AIFunctionFactory.Create` and hand the list to `ChatOptions.Tools`. `AIFunctionFactory.Create` reads the method's name, parameters, and `[Description]` text and builds the tool definition the model sees. Paste this below the starter's `var request = ...` statement (the three lines ending in `September 14-16.";`):

   ```csharp
   var options = new ChatOptions
   {
       Tools =
       [
           AIFunctionFactory.Create(Trailhead.SearchTrails),
           AIFunctionFactory.Create(Trailhead.CheckCampsites),
       ],
   };
   ```
3. Start `messages` with a system message and a user message:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, campgrounds, or availability. Every trail name and campground in your answer must have come back from a tool call in this conversation. Call search_trails and check_campsites before writing any part of the itinerary. If you have not yet called both, your next move is a tool call, not prose.
```

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

   The user message is the starter's `request` variable. Put this below the `options` block from item 2, and paste the system message above between the two `"""` lines (C# raw string quotes, so the text needs no escaping). Indent every pasted line at least as far as the closing `"""`, or it will not compile:

   ```csharp
   List<ChatMessage> messages =
   [
       new(ChatRole.System, """
           """),
       new(ChatRole.User, request),
   ];
   ```

4. Let `UseFunctionInvocation` run the loop and cap it at 12 iterations. That cap is the step budget. Nothing else stops a model that keeps asking for one more tool. Microsoft.Extensions.AI sends the tools, runs each call, appends the results, and repeats until the reply is prose, so this is the whole loop. `ChatClientBuilder` wraps the plain client from `CreateChatClient()` in one that does the loop. These lines replace the starter's `IChatClient client = CreateChatClient();`:

   ```csharp
   IChatClient inner = CreateChatClient();
   IChatClient client = new ChatClientBuilder(inner)
       // Step budget. This cap is the only bound on the tool-calling loop; a model
       // that keeps deciding to call one more tool has no other stopping condition.
       .UseFunctionInvocation(configure: c => c.MaximumIterationsPerRequest = 12)
       .Build();
   ```

   Then send the list and the tools. This one line replaces the starter's whole `var response = await client.GetResponseAsync(` call, from that line down to its closing `""");`:

   ```csharp
   var response = await client.GetResponseAsync(messages, options);
   ```

5. Print `response.Text`. That text is the itinerary. The starter already does this, on the line below the call:

   ```csharp
   Console.WriteLine(response.Text);
   ```

   The starter's two `Console.WriteLine("Note: zero tool calls ...` lines at the end are no longer true; delete them. Run it:

   ```bash
   dotnet run
   ```

**Check:** one `[tool]` line prints for each of the two tools, then the itinerary. The plan should name trails that came back from `search_trails`, such as Trail of the Cedars, Iceberg Lake Trail, or Bowman Lake Shoreline Trail. An itinerary with no `[tool]` lines means the model never saw your tools; an itinerary naming trails you cannot find in `data/trails.json` means the tool result never reached the model.

### Step 3: Add get_weather and plan around the rain day

**Do:**
1. Write `get_weather(park)` the same way as `check_campsites`, reading `data/mock-apis/weather.json`; return `{"error": "No forecast available for '<park>'."}` if nothing matches.

   Put it inside `Trailhead`, next to `CheckCampsites`. Only the file name and the error text differ from that method:

   ```csharp
       [Description("Get the multi-day weather forecast and advisories for a park.")]
       public static string GetWeather(
           [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park")
       {
           Narrate("get_weather", new { park });

           var all = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/weather.json"))!.AsObject();
           var entry = all.FirstOrDefault(kv =>
               kv.Key.Contains(park, StringComparison.OrdinalIgnoreCase) ||
               park.Contains(kv.Key, StringComparison.OrdinalIgnoreCase) ||
               kv.Key.Contains(park.Split(' ')[0], StringComparison.OrdinalIgnoreCase));

           return Result(entry.Value is null
               ? $"{{\"error\": \"No forecast available for '{park}'.\"}}"
               : entry.Value.ToJsonString(Pretty));
       }
   ```

2. Write the `get_weather` description yourself, as the `[Description]` text on the method and its `park` parameter. Only then check the reference in `data/tool-definitions.json`:

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

3. Add `AIFunctionFactory.Create(Trailhead.GetWeather)` to `ChatOptions.Tools` alongside the first two tools. It is one more line inside the `Tools = [ ... ]` list from step 2, below the `SearchTrails` line; keep the trailing comma:

   ```csharp
           AIFunctionFactory.Create(Trailhead.GetWeather),
   ```
4. Replace the system prompt:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, and campground in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails. 3. check_campsites for where to stay each night. Then write the itinerary: one section per day with trail, campsite, and how the forecast shaped the choice (put harder or more exposed hiking on the drier days).
```

   In `Program.cs` the system prompt is the text between the two `"""` lines of `new(ChatRole.System, """` in the `messages` list from step 2. Delete the old text there and paste this in its place.

5. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather.
```

```bash
dotnet run -- "Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather."
```

**Why:** `data/mock-apis/weather.json` covers Glacier 2026-09-13 to 2026-09-18; the planted day is the 16th, rain showers at 49/33 with 0.7 precipitation chance and 18 mph wind, after two sunny days and one partly cloudy one.

**Check:** the itinerary puts the hard hike on the 14th or 15th, and the 16th gets something short or sheltered with a sentence saying why. Failing looks like the same three trails plus "expect rain on the 16th"; if the model never calls `get_weather`, fix your description first. Compare the sample in [`expected-output.md`](../expected-output.md).

### Step 4: Add get_trail_conditions and ask for the washed-out bridge on trail-0117

**Do:**
1. Write `get_trail_conditions(trail_id)`. If `trail_id` is missing/blank, return `{"error": "trailId is required. Call search_trails first and use one of its ids."}` instead of throwing.

   This version already has the forgiving stretch goal in place: a blank or `"null"` id gets an error listing the ids `SearchTrails` last returned, and a trail name resolves to its id. That is why its error text says `Call this tool again with one of these ids: ...` instead of the sentence above; return the code's version, which gives the model the ids to retry with. The method goes inside `Trailhead`. It opens with the attributes, the `[tool]` line, and the missing-id check, which returns early instead of throwing:

   ```csharp
       [Description("Get the most recent hiker-submitted condition reports for a trail. Always check this before recommending a trail; reports surface closures and hazards such as washouts.")]
       public static string GetTrailConditions(
           [Description("The trail id from search_trails, e.g. 'trail-0117'. A trail name also works.")] string? trailId = null)
       {
           Narrate("get_trail_conditions", new { trail_id = trailId });

           if (string.IsNullOrWhiteSpace(trailId) || trailId is "null" or "string")
           {
               var candidates = LastResultIds.Count > 0 ? string.Join(", ", LastResultIds) : "call search_trails first";
               return Result($"{{\"error\": \"trailId is required. Call this tool again with one of these ids: {candidates}.\"}}");
           }
   ```

   Directly below it, still in the method, the name lookup: anything that does not start with `trail-` is looked up by name in `trails.json` and swapped for that trail's id:

   ```csharp
           if (!trailId.StartsWith("trail-", StringComparison.OrdinalIgnoreCase))
           {
               var trails = JsonNode.Parse(File.ReadAllText($"{DataDir}/trails.json"))!.AsArray();
               var byName = trails.FirstOrDefault(t =>
                   ((string)t!["name"]!).Contains(trailId, StringComparison.OrdinalIgnoreCase));
               if (byName is not null) trailId = (string)byName["id"]!;
           }
   ```

2. Open `data/condition-reports.jsonl` line by line, parsing each non-empty line as JSON (`id`, `trail_id`, `date`, `text`). Keep reports whose `trail_id` matches (case-insensitive). Sort by `date`, newest first. Keep the first 4.

   Next in the same method. `File.ReadLines` yields one line at a time, each is parsed as its own JSON object, and the LINQ chain filters, sorts newest first, keeps 4, and projects each report into an anonymous `{ date, report }` object:

   ```csharp
           var id = trailId;
           var reports = File.ReadLines($"{DataDir}/condition-reports.jsonl")
               .Select(l => JsonNode.Parse(l)!)
               .Where(r => string.Equals((string)r["trail_id"]!, id, StringComparison.OrdinalIgnoreCase))
               .OrderByDescending(r => (string)r["date"]!)
               .Take(4)
               .Select(r => new { date = (string)r["date"]!, report = (string)r["text"]! })
               .ToArray();
   ```

3. If none matched, return `{"error": "No condition reports found for '<trail_id>'."}`. Otherwise return a JSON array of `{date, report}` objects (`report` = the line's `text`).

   The method ends by returning one or the other, then closes:

   ```csharp
           return Result(reports.Length == 0
               ? $"{{\"error\": \"No condition reports found for '{trailId}'.\"}}"
               : JsonSerializer.Serialize(reports, Pretty));
       }
   ```

4. Declare `get_trail_conditions` by adding `AIFunctionFactory.Create(Trailhead.GetTrailConditions)` to `Tools`, one more line in the same list as step 3:

   ```csharp
           AIFunctionFactory.Create(Trailhead.GetTrailConditions),
   ```

   The reference definition, whose wording belongs in the `[Description]` text:

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

   As in step 3, paste it between the two `"""` lines of `new(ChatRole.System, """`, replacing the step 3 prompt.

6. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).
```

```bash
dotnet run -- "Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117)."
```

**Why:** the catalog entry for `trail-0117` says nothing about the bridge. Only the condition reports do. 40 reports belong to `trail-0117`, and eight are the planted washout: five from 2026-06-18 to 2026-06-24 announcing the footbridge is gone, three July follow-ups saying it is still out.

**Check:** the model calls `get_trail_conditions` with `trail-0117` and the itinerary drops or flags the trail because the bridge is out, traceable to the tool result. Failures, worst last: never calling the tool and scheduling `trail-0117`; reading "the bridge is OUT" and scheduling it anyway; dropping it with an invented reason without calling the tool. `llama3.2` does the second one; tighten the closure instruction in the system prompt, and if it persists that is your argument for a stronger model.

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

  **Why this needs a gate:** filing a permit is the one action in this agent that cannot be undone, so the function never runs just because the model asked. Print the four arguments as one line, ask `File it? [y/N]`, and read the answer. On `y`/`yes`, return the `submit_response` object from `data/mock-apis/permits.json` (a fixture: `status: "submitted"`, `confirmation_id: "TRG-2026-091482"`, a message clarifying it's a request rather than a confirmed permit, and `next_steps`). On anything else, return this string instead, so the model has something to act on:

  ```text
  {"status":"cancelled","message":"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed."}
  ```

  Add a `--yes` flag that skips the question for demo runs.

  Add the method inside `Trailhead`, next to the other tools. The static `AutoApprovePermits` from step 1 holds the flag. `Console.ReadLine()` pauses the loop until someone types an answer, and `?.` handles the case where there is no input at all:

  ```csharp
      [Description("Submit a backcountry permit request. This files a real request, so use it once, at the end, after the plan is settled.")]
      public static string RequestPermit(
          [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park",
          [Description("Permit zone, e.g. 'Lake McDonald / Sperry'.")] string zone = "Lake McDonald / Sperry",
          [Description("Trip dates, e.g. '2026-09-14 to 2026-09-16'.")] string dates = "unspecified",
          [Description("Number of people in the group.")] int groupSize = 2)
      {
          Narrate("request_permit", new { park, zone, dates, group_size = groupSize });

          // Filing a permit is the one irreversible action in this agent, so it
          // never runs on the model's say-so; a human confirms first. --yes
          // bypasses the prompt and exists for demo runs only.
          Console.WriteLine($"  [gate] About to file a permit request: {park}, zone '{zone}', {dates}, group of {groupSize}.");
          bool approved;
          if (AutoApprovePermits)
          {
              Console.WriteLine("  [gate] --yes supplied; auto-approved.");
              approved = true;
          }
          else
          {
              Console.Write("  [gate] File it? [y/N] ");
              approved = Console.ReadLine()?.Trim().ToLowerInvariant() is "y" or "yes";
          }

          if (!approved)
              return Result("{\"status\": \"cancelled\", \"message\": \"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed.\"}");

          var permits = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/permits.json"))!;
          return Result(permits["submit_response"]!.ToJsonString(Pretty));
      }
  ```

  Register it with one more line in the `Tools = [ ... ]` list:

  ```csharp
          AIFunctionFactory.Create(Trailhead.RequestPermit),
  ```

  These lines replace the starter's `var request = args.Length > 0 ...` statement (three lines) near the top of the file, so `--yes` is not sent as part of the request. `args.Where(a => a != "--yes")` keeps every word except the flag:

  ```csharp
  var autoApprove = args.Contains("--yes");
  var requestArgs = args.Where(a => a != "--yes").ToArray();
  var request = requestArgs.Length > 0
      ? string.Join(" ", requestArgs)
      : "Plan me a 3-day trip in Glacier National Park for September 14-16.";

  Trailhead.AutoApprovePermits = autoApprove;
  ```

  **Check:** `request_permit` never runs on the model's say-so, and a declined request ends with an itinerary that says no permit was filed, without retrying.

- **Nudge a model that stops early.** A small model often quits after one or two tools, then writes the plan from trails it never looked up. After `GetResponseAsync` returns, compare the set of tools called against `get_weather`, `search_trails`, `get_trail_conditions`, and `check_campsites`. Track called tool names in a static set that each tool adds itself to; `Narrate` from step 1 already does `Called.Add(tool)`. If any are missing, first append `response.Messages`, the whole previous response, so the tool calls and results stay in the history, then append a user message naming them and asking for the next call; call `GetResponseAsync` again, up to 3 times. If every tool was called but the answer has no "day" in it, append the response and one user message asking for the itinerary and run once more. Print `[nudge]` whenever this fires.

  This line goes just above step 2's `var response = await client.GetResponseAsync(messages, options);`:

  ```csharp
  string[] required = ["get_weather", "search_trails", "get_trail_conditions", "check_campsites"];
  ```

  The retry loop goes below that call and above `Console.WriteLine(response.Text);`. Each pass works out which required names are not in `Trailhead.Called`, stops if none are missing, otherwise appends the previous response and a user message and calls again. The `hint` adds the trail ids from the last search when `get_trail_conditions` is the one missing:

  ```csharp
  for (var nudge = 0; nudge < 3; nudge++)
  {
      var missing = required.Where(t => !Trailhead.Called.Contains(t)).ToArray();
      if (missing.Length == 0) break;

      Console.WriteLine($"[nudge] still missing: {string.Join(", ", missing)}");
      messages.AddRange(response.Messages);
      var hint = missing.Contains("get_trail_conditions") && Trailhead.LastResultIds.Count > 0
          ? $" Use one of these trail ids: {string.Join(", ", Trailhead.LastResultIds)}."
          : "";
      messages.Add(new ChatMessage(ChatRole.User,
          $"You have not called these tools yet: {string.Join(", ", missing)}. " +
          $"Call the next one now with real arguments.{hint} Do not write the itinerary yet."));
      response = await client.GetResponseAsync(messages, options);
  }
  ```

  Directly below the loop, still above `Console.WriteLine(response.Text);`, the one extra turn for a reply with no "day" in it:

  ```csharp
  if (!response.Text.Contains("Day", StringComparison.OrdinalIgnoreCase))
  {
      Console.WriteLine("[nudge] tools are done but no itinerary was written; asking for it.");
      messages.AddRange(response.Messages);
      messages.Add(new ChatMessage(ChatRole.User,
          "Every tool you need has been called. Write the final itinerary now, " +
          "using only what the tools returned. Do not call any more tools."));
      response = await client.GetResponseAsync(messages, options);
  }
  ```

  **Check:** against `llama3.2`, `[nudge]` lines appear and the run still reaches all four tools, as in [`reference-transcript.md`](../reference-transcript.md) (verbatim console output of one `complete/` run for the step 2 request, plus a second run asking for a 2-day trip including `trail-0117`, showing the agent routing around the closed trail). Against `gpt-5.5` on Azure, no `[nudge]` line prints at all.

- **Make the tools forgiving.** Give every tool parameter a default. Let `GetTrailConditions` accept a trail name where it expects an id, by looking the name up in `data/trails.json`. Return an error string on bad input instead of throwing, and have `SearchTrails` remember the ids it last returned (the static `LastResultIds` list, cleared and refilled on every search) so the error can list them; the step 1 and step 4 code above already does all of this. `UseFunctionInvocation` parses the `arguments` string for you, so there is no loop code to change; the defaults are what let a call with missing arguments still run. **Check:** a call with `trailId` set to `"null"` gets an error naming the valid ids back, a made-up id such as `"[insert trail IDs here]"` gets the `No condition reports found` error, and the run keeps going instead of crashing.

When you are done, compare your trace with [`reference-transcript.md`](../reference-transcript.md), and be precise about what this does and does not do: it prints the trace, it does not persist one. Writing that trace to durable storage next to feature 09's decisions log is what you would show a security team.

## What Is in This Folder

- `data/tool-definitions.json`: the five tools as JSON-schema function definitions: `search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`. The Python and TypeScript answer keys load this file directly; the .NET answer key declares the same tools from attributes on its functions. Read the descriptions before you write your own.
- `reference-transcript.md`: a real, unedited `llama3.2` run for the step 2 request, plus the closed-trail run, described under the nudge stretch goal. Use it to check the shape of your loop, not the wording.
- `expected-output.md`: the success checks for the lab steps plus the stretch goal, with the failure modes each one is meant to catch.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The tools read the data in this folder: `trails.json` (the full 200-trail catalog that features 04 and 06 use a slice of), `condition-reports.jsonl` (the full 500-report stream that feature 08's per-trail files are cut from), and the hand-written fixtures in `mock-apis/` (`weather.json`, `campsites.json`, `permits.json`). Each step above says what the file it opens looks like. None of it is generated by a script; it all ships with the workshop. Nothing here calls a real park service, and `request_permit` returns a canned confirmation id.

Two facts hide in that data, and the agent has to discover them rather than be told:

- September 16, 2026 in Glacier is a rain day: 49/33, 70 percent chance of precipitation, 18 mph wind, after two dry days. A good plan moves the hard hiking off it.
- Avalanche Lake Trail is `trail-0117`, and its condition reports have said the footbridge over the gorge is gone since June 2026. The catalog entry says nothing about it. Only `get_trail_conditions` shows it.

One detail of `search_trails` matters more than it looks. Its `features` keywords match the trail's name as well as its feature tags. The tool returns at most eight trails. Avalanche Lake is the 27th Glacier trail in the catalog, and no feature tag on it contains "Avalanche". Without the name match, a request for "Avalanche Lake Trail" has no way to reach `trail-0117` through search. In a soak test against `gpt-5.5` with a tag-only search, the agent gave up on the trail in six runs out of ten. It said so honestly and planned around it. But the washed-out bridge never came up, so the lab step had nothing to teach. The model did what it could with the tool it had. Remember that when your own agent misbehaves. The tool is often the cheaper place to look first.

The loop is written out three ways: `dotnet/complete/Program.cs`, `python/complete/main.py`, and `typescript/complete/index.ts`. Read whichever one is closest to your language; the two without a framework are about thirty lines.
