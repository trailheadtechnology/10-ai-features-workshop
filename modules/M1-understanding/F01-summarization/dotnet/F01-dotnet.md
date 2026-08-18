# .NET Demo for 01 Summarization

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one naive prompt ("Summarize this trip report."), which produces the book report.
- `complete/`: the finished demo as shown on stage.

Both run against Ollama (`llama3.2`), matching the demo script in [docs/slides/outlines/M1-understanding.md](../../../../docs/slides/outlines/M1-understanding.md). From `complete/`:

```bash
dotnet run                                  # naive prompt on tr-0004 (the book report)
dotnet run -- --briefing                    # 3-bullet hiker briefing; the washout leads
dotnet run -- --headline                    # one-line status for a trail card UI
dotnet run -- --briefing --audience ranger  # stretch goal: same report, different audience
dotnet run -- ../../data/tr-0001.md          # any report path works
```

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F01-lab.md`](../F01-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter As-Is and Read the Book Report

The starter is spec step 1: the naive prompt on the clean report, `tr-0001.md`. Run it twice. It is faithful, generic, and useless, and that is the baseline you are improving on.

Run:

```bash
dotnet run
```

Check: A paragraph or two about the author's gear and their day. Nothing a hiker planning Saturday could act on.

### Step 2: Rewrite the Prompt Into the 3-bullet Briefing (lab step 2)

Replace the naive prompt with one that demands exactly three bullets (conditions, hazards or closures, crowding) and nothing else. Start with just that, and run it four or five times on the clean report before adding anything: a prompt that requires a hazard bullet will invent one from a bear sighting or the word "avalanche" in the trail name. When you see that happen, add the last three lines below. They give the model a legal way to report nothing, and they are the only reason the finished prompt is trustworthy.

```csharp
var prompt = $"""
    You are helping a hiker planning to hike this trail within the next week.
    From the trip report below, produce exactly 3 bullets covering:
    current trail conditions, hazards or closures, and crowding.
    Ignore gear talk, personal stories, and scenery.
    Report only what the trip report states. Do not turn a wildlife sighting into a
    hazard or a closure, and write "no closures or hazards reported" when it says none.
    If the report does state a closure or hazard, it must appear in the first bullet.

    {report}
    """;
var response = await client.GetResponseAsync(prompt);
```

Run:

```bash
dotnet run   # several times
```

Check: Three bullets, and on `tr-0001.md` the hazards bullet says nothing is closed, every run. The measured invention rate without the last three lines is in `../expected-output.md`.

### Step 3: Run the Buried-Hazard Report Through the Same Prompt (lab step 3)

The prompt stays the same and the report changes: `tr-0004.md` mentions the washed-out footbridge in passing, halfway down; a good briefing puts it first.

Run:

```bash
dotnet run -- ../../data/tr-0004.md
```

Check: The first bullet is the closure. Compare the sample in `../expected-output.md`. If the bridge is missing or lands third, tighten the "must appear in the first bullet" line.

### Step 4: Stretch: Switch the Audience

Same call, same report, one variable swapped into the first line of the prompt. This is what `--audience ranger` does in `complete/`.

```csharp
var audienceFocus = audience == "ranger"
    ? "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
    : "a hiker planning to hike this trail within the next week";
// then: You are helping {audienceFocus}.
```

Check: The ranger version drops the crowding chatter and leads with the bridge as a maintenance item. `complete/` also has the one-line `--headline` shape: same call, a 12-word instruction, a different UI slot.
