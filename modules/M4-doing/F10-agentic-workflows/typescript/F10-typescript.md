# TypeScript Demo for 10 Agentic Workflows

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: the trip request as a plain chat completion, no tools, no loop: a fluent generic itinerary that checks nothing and books nothing.
- `complete/index.ts`: the finished demo as shown on stage. Five tools as ordinary functions over `../data/` (`search_trails`, `get_weather`, `get_trail_conditions`, `check_campsites`, `request_permit`), their definitions loaded from `../data/tool-definitions.json` so the model sees exactly what the `.http` lab sends, a hand-written tool-calling loop with a step budget of 12, the permit gate that waits for a human yes, and the nudge logic for a model that stops early.

Setup once (`npm install` in this directory), then:

```bash
npm run complete                                            # the capstone request
npm run complete -- Plan me a trip on Avalanche Lake Trail in September
npm run complete -- --yes <request>                         # auto-approve the permit gate
```

There is no agent framework here on purpose: the loop is the same one `../http/azure.http` walks by hand, about thirty lines, and reading it is the fastest way to see what frameworks hide. Without the Azure variables it runs on `llama3.2`, which is much weaker at sequencing five tools; the `[nudge]` lines are the app compensating, and [`../dotnet/F10-dotnet.md`](../dotnet/F10-dotnet.md) has the measured failure counts before judging a local run. The reference run is in [`../reference-transcript.md`](../reference-transcript.md).

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, the deployment name the feature uses, and the key handed out in the room) and the agent switches to Azure OpenAI through the SDK's `AzureOpenAI` client; leave them unset and it runs against Ollama.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the TypeScript equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F10-lab.md`](../F10-lab.md), done in TypeScript: start from `starter/index.ts` and end where `complete/index.ts` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run with `npm run starter` from the `typescript/` directory; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: A Plan with No Tools

A plain chat completion, no tools, no loop. The itinerary is fluent, generic, ignores the washed-out bridge, and books nothing, because nothing here can reach the catalog, the weather feed, or the condition reports.

Run:

```bash
npm run starter
```

Check: A lovely three-day plan with zero tool calls. That is the reason this feature exists.

### Step 2: Two Tools and the Loop

This is lab step 1, the round-trip that `../http/azure.http` walks by hand. Write `search_trails` and `check_campsites` as ordinary functions over `../data/trails.json` and `../data/mock-apis/campsites.json`, load their definitions from `../data/tool-definitions.json` (the two entries you need), and write the loop: send the messages with the `tools` array, read the tool calls out of the reply, run them, append the results, repeat until the reply is prose. Give the loop a step budget; it is the only thing that stops a model that keeps deciding to call one more tool.

```typescript
const TOOLS: ChatCompletionTool[] = load("tool-definitions.json").tools.filter((t: any) => ["search_trails", "check_campsites"].includes(t.function.name));
const FUNCTIONS: Record<string, (args: any) => string | Promise<string>> = { search_trails: searchTrails, check_campsites: checkCampsites };

async function runAgent(messages: ChatCompletionMessageParam[], maxIterations = 12): Promise<string> {
  for (let i = 0; i < maxIterations; i++) {
    const response = await client.chat.completions.create({ model, messages, tools: TOOLS, tool_choice: "auto" });
    const message = response.choices[0].message;
    if (!message.tool_calls?.length) return message.content ?? "";
    messages.push({ role: "assistant", content: message.content, tool_calls: message.tool_calls });
    for (const call of message.tool_calls) {
      if (call.type !== "function") continue;
      const output = await FUNCTIONS[call.function.name](JSON.parse(call.function.arguments || "{}"));
      messages.push({ role: "tool", tool_call_id: call.id, content: output });
    }
  }
  return "";
}
```

Run:

```bash
npm run starter
```

Check: Print each tool call as it happens. You should see `search_trails` and `check_campsites` fire, then an itinerary that names real trails (Trail of the Cedars, Iceberg Lake) and real campgrounds. Invented names mean a tool result did not reach the model.

### Step 3: Add Get_weather and Ask for a Trip on the Rain Day (lab step 2)

Write the function over `../data/mock-apis/weather.json`, add its definition to the tools array (write the JSON schema yourself before copying it from `tool-definitions.json`; the description is the model's only manual), and ask for September 14 to 16. The 16th is a rain day: 49/33, 70 percent, 18 mph.

```typescript
function getWeather({ park = "Glacier National Park" }: { park?: string }): string {
  const entry = parkEntry(load("mock-apis/weather.json"), park);
  return result(entry ?? { error: `No forecast available for '${park}'.` });
}
// then add the get_weather definition to TOOLS and the function to FUNCTIONS
```

Run:

```bash
npm run starter
```

Check: `get_weather` is called and the forecast shapes the plan rather than decorating it: the hardest day lands on the 14th or 15th and the 16th gets something short or sheltered, with a sentence saying why. Compare the sample in `../expected-output.md`. Failing looks like the same three trails plus a line reading "expect rain on the 16th".

### Step 4: Add Get_trail_conditions and Ask for the Closed Trail (lab step 3)

The function reads `../data/condition-reports.jsonl` and returns the newest four reports for a trail id; make it return an error string, not throw, when the id is missing or malformed, and let it resolve a trail name too. Add it, then ask for a trip that includes Avalanche Lake Trail (`trail-0117`). The catalog says nothing about the bridge; only this tool does.

```typescript
function getTrailConditions({ trail_id = null }: { trail_id?: string | null }): string {
  if (!trail_id || ["", "null", "string"].includes(trail_id.trim()))
    return result({ error: "trailId is required. Call search_trails first and use one of its ids." });
  const reports: Report[] = readFileSync(resolve(DATA, "condition-reports.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
  const mine = reports.filter((r) => r.trail_id.toLowerCase() === trail_id.toLowerCase()).sort((a, b) => b.date.localeCompare(a.date)).slice(0, 4);
  if (mine.length === 0) return result({ error: `No condition reports found for '${trail_id}'.` });
  return result(mine.map((r) => ({ date: r.date, report: r.text })));
}
```

Run:

```bash
npm run starter -- Plan me a 2-day trip in Glacier National Park for September 14-15 that includes the Avalanche Lake Trail \(trail-0117\).
```

Check: The model calls `get_trail_conditions` with `trail-0117`, gets the four bridge-is-out reports, and the itinerary drops or flags the trail for that reason. Three failure modes, in ascending order of embarrassment: it never calls the tool; it calls it, reads "bridge is OUT", and schedules the trail anyway (`llama3.2` does this; tighten the closure instruction in the system prompt, and if it persists that is your argument for a stronger model); it drops the trail but invents a reason without calling. The system prompt in `complete/` has the closure rule that fixed the second case.

### Step 5: Stretch: The Human Gate, and the Two Nudges (lab stretch goal)

`request_permit` is the one irreversible action, so it never runs on the model's say-so: print the summary, wait for a human yes, and only then return the tool result. Declining returns something the model can act on rather than silence. `complete/` also has two nudges for a local model that stops early: a check that every required tool was called (up to three re-prompts) and one turn asking for the itinerary if the tools are done and no plan was written. A frontier model should need neither; `[nudge]` lines mean the model is underpowered for the task.

```typescript
console.log(`  [gate] About to file a permit request: ${park}, zone '${zone}', ${dates}, group of ${group_size}.`);
const answer = (await rl.question("  [gate] File it? [y/N] ")).trim().toLowerCase();
if (answer !== "y" && answer !== "yes")
  return result({ status: "cancelled", message: "The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed." });
return result(load("mock-apis/permits.json").submit_response);
```

Check: `request_permit` does not execute on the model's say-so; declining ends with an itinerary that says no permit was filed. Then compare your trace with `../reference-transcript.md`, and be precise about what this does and does not do: it prints the trace, it does not persist one. Writing that trace to durable storage next to feature 09's decisions log is what you would show a security team.
