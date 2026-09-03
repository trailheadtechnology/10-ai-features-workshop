# Lab 01: Summarization

*This is the Recommended lab for [Module 1](../M1-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** `data/tr-0001.md`, a clean report (mud and crowds, no closure), and `data/tr-0004.md`, the report with the washed-out footbridge buried mid-text. The other 38 in `data/` are the rest of the Avalanche Lake corpus.
- **How:** POST to Ollama's chat endpoint. `http/ollama.http` holds all three requests with the reports inlined; the code tracks read the same files from `data/`.
- **Model:** `llama3.2`, local. No key.

### Step 1: The naive prompt on `tr-0001.md`

Request 1 in `http/ollama.http`, or your track's starter, on `data/tr-0001.md`. The prompt, one line above the report:

```text
Summarize this trip report.
```

**Check:** a faithful, useless book report about gear. Right model, wrong instruction.

### Step 2: The briefing prompt on `tr-0001.md`

Request 2, or paste this over the starter's prompt, still on `data/tr-0001.md`. Keep the line breaks; reflowing changes `llama3.2`'s behavior.

```text
You are helping a hiker planning to hike this trail within the next week.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.
```

**Check:** three bullets for `tr-0001.md`: no gear, mud and crowds in, and "no closures or hazards reported." A bullet that closes the trail because of the bear is the failure the last two prompt lines exist to stop; delete them and run again to see it.

### Step 3: The briefing prompt on `tr-0004.md`

Request 3: the step 2 prompt on `data/tr-0004.md`. **Check:** the washed-out footbridge is the first bullet, and `tr-0001.md` still comes back with nothing closed.

### Stretch goal: the same report for a different reader

Replace the first line of the step 2 prompt with:

```text
You are helping a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery.
```

Run `data/tr-0004.md` through both. **Check:** the ranger version leads with where the bridge went out and the barricade; the hiker version keeps the crowding. Identical output means the audience line is not reaching the prompt. (.NET `complete/`: `--audience ranger`.)

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F01-http.md`](http/F01-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F01-dotnet.md`](dotnet/F01-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F01-python.md`](python/F01-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F01-typescript.md`](typescript/F01-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `data/tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `data/tr-*.md`: all 40 Avalanche Lake trip reports; `tr-0001` and `tr-0004` are the lab's two named inputs
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
