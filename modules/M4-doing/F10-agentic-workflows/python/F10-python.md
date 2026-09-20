<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 10: Agentic Workflows (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F10-dotnet.md), [TypeScript](../typescript/F10-typescript.md). Lab overview: [F10-lab.md](../F10-lab.md).*

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

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), or at Azure through the SDK's `AzureOpenAI` client when the variables are set:

- `starter/main.py`: the trip request as a plain chat completion, no tools, no loop. It takes no flags, only the optional request text.
- `complete/main.py`: the finished demo as shown on stage. Five tools as ordinary functions over `data/`, their definitions loaded from `data/tool-definitions.json` so the model sees exactly the definitions this lab prints, a hand-written tool-calling loop with a step budget of 12, the permit gate that waits for a human yes, and the nudge logic for a model that stops early.

There is no agent framework here on purpose: the loop is about thirty lines, and reading it is the fastest way to see what frameworks hide. No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [`SETUP.md`](../../../../SETUP.md)) is the one install for all ten features. `uv run` finds it from any folder. Run the starter from `starter/` with `uv run main.py`. From `complete/`:

```bash
uv run main.py                                       # the capstone request
uv run main.py Plan me a trip on Avalanche Lake Trail in September
uv run main.py --yes <request>                       # auto-approve the permit gate
```

Without the Azure variables it runs on `llama3.2`, which is much weaker at sequencing five tools; the `[nudge]` lines are the app compensating, and [`dotnet/F10-dotnet.md`](../dotnet/F10-dotnet.md) has the measured failure counts before judging a local run.

Set the three variables in the same terminal you run from (macOS or Linux shown; the values come from the room):

```bash
# Hint: fill in the key and deployment, then run uv run main.py in this same terminal
export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
export AZURE_OPENAI_KEY=<key handed out in the room>
export AZURE_OPENAI_DEPLOYMENT=gpt-5.5
```

The starter already reads them. `os.environ.get` returns the variable's value, or `None` when it is not set, and the `if` picks Azure only when all three are there:

```python
def create_chat_client() -> tuple[OpenAI, str]:
    endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
    key = os.environ.get("AZURE_OPENAI_KEY")
    deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
    if endpoint and key and deployment:
        return AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment
    print("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.")
    return OpenAI(base_url="http://localhost:11434/v1", api_key="ollama"), "llama3.2"
```

### Step 0: Run the starter and read the plan with no tools

**Do:** run `starter/` as it is. It sends this request as one user message, no tools, no loop:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

```bash
uv run main.py
```

The starter already does this. It picks the request (your command-line words joined with spaces, or the default sentence when there are none) and sends it in one call with no tools:

```python
request = " ".join(sys.argv[1:]) or "Plan me a 3-day trip in Glacier National Park for September 14-16."
```

```python
response = client.chat.completions.create(
    model=model,
    messages=[{"role": "user", "content": f"You are the trip planner for Trailhead Guides, a hiking app.\n\n{request}"}],
)

print(response.choices[0].message.content)
```

**Check:** a fluent three-day plan with zero tool calls. The model is working from memory, so the trails and campgrounds it names may not exist, and nothing in the plan reflects the weather for September 14-16 or whether those trails are open.

### Step 1: Write the first two tools as ordinary functions

**Do:** each tool is a plain function that reads a file under `data/` and returns a JSON string. From `starter/` that folder is two levels up (`../../data`); keep the prefix in one constant, and resolve it against the source file's own location rather than the working directory, so the run command's folder does not matter. At the top of each, print `[tool] ` plus the tool name and its arguments. That line is how you watch the loop run.

Add these two imports at the top of `main.py`, directly below the starter's `import sys`. `json` turns JSON text into Python lists and dicts and back; `Path` builds file paths:

```python
import json
from pathlib import Path
```

Below the starter's `create_chat_client` function and above its `client, model = create_chat_client()` line, add the data folder and two shared variables. `Path(__file__).resolve()` is this file's full path; `.parents[0]` is `starter/`, `.parents[1]` is `python/`, `.parents[2]` is the feature folder, and `/ "data"` adds the folder name. `called` is a set, a collection like a list that holds each name only once. `called` and `last_result_ids` are used by the stretch goals; declaring them now costs nothing:

```python
DATA = Path(__file__).resolve().parents[2] / "data"
called: set[str] = set()
last_result_ids: list[str] = []
```

Directly below them, three small helpers every tool uses. `load` reads a file under `data/` into Python lists and dicts. `narrate` records the tool name in `called` and prints the `[tool]` line. `result` turns the answer into a JSON string (`json.dumps`), prints the first 120 characters, and returns the string:

```python
def load(name: str):
    return json.loads((DATA / name).read_text())


def narrate(tool: str, args: dict) -> None:
    called.add(tool)
    print(f"[tool] {tool} {json.dumps(args)}")


def result(payload) -> str:
    text = payload if isinstance(payload, str) else json.dumps(payload)
    preview = text if len(text) <= 120 else text[:120] + "..."
    print(f"  [result] {preview}")
    return text
```

A function has to be defined above the line that runs it, so every function in this lab goes in this area, above `client, model = create_chat_client()`.

1. Write `search_trails(park, features, max_difficulty)`. Open `data/trails.json` (200 trails across six parks, 45 in Glacier). Keep trails whose `park` contains the `park` argument (case-insensitive). If `max_difficulty` was given, drop trails harder than it (`easy` < `moderate` < `hard`). If `features` was given, keep a trail only when at least one keyword appears in its `name` **or** its `features` tags. Do not skip the name match. Stop after 8 matches. Return them as a JSON array without the `description` field.

   Put it below `result`. Each parameter has a default, so a call that leaves one out still runs. `rank` turns `easy`/`moderate` into 0/1, and `rank.get(name, 2)` gives 2 for anything else (`hard`), so difficulties can be compared as numbers. The long `if features and not any(...)` line skips a trail unless some keyword is inside its name or inside one of its feature tags. `{k: t[k] for k in (...)}` copies every listed field into a new dict, which leaves `description` out. `last_result_ids[:] = ...` refills the shared list in place:

   ```python
   def search_trails(park: str = "Glacier National Park", features: list[str] | None = None, max_difficulty: str | None = None) -> str:
       narrate("search_trails", {"park": park, "features": features, "max_difficulty": max_difficulty})
       if isinstance(features, str):
           features = [features] if features.strip() else None
       rank = {"easy": 0, "moderate": 1}
       max_rank = 2 if not max_difficulty else rank.get(max_difficulty.lower(), 2)
       found = []
       for t in load("trails.json"):
           if park.lower() not in t["park"].lower():
               continue
           if rank.get(t["difficulty"], 2) > max_rank:
               continue
           if features and not any(f.lower() in t["name"].lower() or f.lower() in x.lower() for f in features for x in t["features"]):
               continue
           found.append({k: t[k] for k in ("id", "name", "park", "distance_mi", "elevation_ft", "difficulty", "features")})
           if len(found) == 8:
               break
       last_result_ids[:] = [t["id"] for t in found]
       return result(found)
   ```

2. Write `check_campsites(park)`. Open `data/mock-apis/campsites.json` (an object keyed by park name, skip the `_comment` key). Find the key that matches the park: the key contains the argument, the argument contains the key, or the key contains the argument's first word, ignoring case. Return that entry as a JSON string, or `{"error": "No campsite data for '<park>'."}` if nothing matches.

   Add `park_entry` below `search_trails`. It walks the object's keys and values with `.items()`, skips names that start with `_` (the `_comment` key), and returns the value of the first key that matches by one of the three rules, or `None` if none does. Step 3's `get_weather` reuses it:

   ```python
   def park_entry(table: dict, park: str):
       first = park.split(" ")[0].lower()
       for name, value in table.items():
           if name.startswith("_"):
               continue
           if park.lower() in name.lower() or name.lower() in park.lower() or first in name.lower():
               return value
       return None
   ```

   Then `check_campsites` directly below it. `entry if entry is not None else {...}` picks the match, or the error dict when there was none:

   ```python
   def check_campsites(park: str = "Glacier National Park") -> str:
       narrate("check_campsites", {"park": park})
       entry = park_entry(load("mock-apis/campsites.json"), park)
       return result(entry if entry is not None else {"error": f"No campsite data for '{park}'."})
   ```

To see the Check before there is a loop, call both functions once. Put these lines directly below `check_campsites`, run `uv run main.py` (the starter's chat call still runs after them), then delete them before step 2:

```python
# Hint: temporary test lines; delete them before step 2
trails = json.loads(search_trails("Glacier National Park"))
print(len(trails), trails[0]["id"], trails[0]["name"], trails[-1]["id"], trails[-1]["name"])
print(check_campsites("Glacier National Park"))
```

```bash
uv run main.py
```

**Why:** the planted record is `trail-0117`, Avalanche Lake Trail, whose catalog description mentions the footbridge crossing at the gorge's mouth and says nothing about it being gone. Only the condition reports (step 4) do. Avalanche Lake Trail is the 27th Glacier trail in the catalog and none of its feature tags contain "Avalanche", so without the name match in `search_trails`, a search for it can never reach it. Similarly, `check_campsites`'s planted fact (named in the fixture's own `_comment`) is that Sperry Chalet Area Sites shows 0 open sites on the 13th, 14th, and 15th, so a plan that sleeps there those nights ignored the tool.

**Check:** `search_trails` with `{"park":"Glacier National Park"}` returns 8 trails, from `trail-0003` Trail of the Cedars to `trail-0037` Bowman Lake Shoreline Trail. `check_campsites` with the same park returns 4 campgrounds, and Sperry Chalet Area Sites shows 0 open sites on September 14 and 15.

### Step 2: Write the loop and run the two-tool round-trip

**Do:**
1. Declare the two tools by loading the definitions below from `data/tool-definitions.json`. The file is an object whose `tools` key holds all five definitions; at this step keep only the two you have functions for:

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

   Put this below `check_campsites` (after deleting step 1's test lines). `load("tool-definitions.json")["tools"]` is the list of all five definitions; `[t for t in ... if ...]` builds a new list with only the ones whose name is in the parentheses:

   ```python
   # Hint: complete/ loads all five; this keeps the two you have so far
   TOOLS = [t for t in load("tool-definitions.json")["tools"] if t["function"]["name"] in ("search_trails", "check_campsites")]
   ```

2. Make a dictionary mapping each tool name to its function.

   Directly below `TOOLS`. Each key is the name the model asks for; each value is the function itself, written without `()` so nothing runs yet:

   ```python
   FUNCTIONS = {
       "search_trails": search_trails,
       "check_campsites": check_campsites,
   }
   ```

3. Start `messages` with a system message and a user message:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, campgrounds, or availability. Every trail name and campground in your answer must have come back from a tool call in this conversation. Call search_trails and check_campsites before writing any part of the itinerary. If you have not yet called both, your next move is a tool call, not prose.
```

```text
Plan me a 3-day trip in Glacier National Park for September 14-16.
```

   Below `FUNCTIONS`, keep the system message in a constant. Triple quotes let the text contain quotes and apostrophes as they are:

   ```python
   # Hint: paste the system message above between the triple quotes
   SYSTEM_PROMPT = """<system message>"""
   ```

   The user message is the starter's `request` variable, which the starter already sets:

   ```python
   request = " ".join(sys.argv[1:]) or "Plan me a 3-day trip in Glacier National Park for September 14-16."
   ```

   At the bottom of the file, delete the starter's `response = client.chat.completions.create(` call (four lines, down to its closing `)`) and the four `print` lines below it. Keep the `print(f"Request: {request}")` and `print("-" * 60)` lines above it. In their place, start the list. Each message is a dict with a `role` and its `content`:

   ```python
   messages: list[dict] = [{"role": "system", "content": SYSTEM_PROMPT}, {"role": "user", "content": request}]
   ```

4. Loop at most 12 times. That cap is the step budget. Nothing else stops a model that keeps asking for one more tool.

   Below `SYSTEM_PROMPT`, the step budget and the outline of the loop function. `for _ in range(MAX_ITERATIONS):` runs its body at most 12 times (`_` is a name for a counter you do not use). Items 5 to 9 replace the `pass` line, so they sit inside the `for`, indented 8 spaces. The last two lines only run if all 12 passes end without a prose answer:

   ```python
   # Hint: items 5 to 9 replace the pass line
   MAX_ITERATIONS = 12


   def run_agent(client: OpenAI, model: str, messages: list[dict]) -> str:
       for _ in range(MAX_ITERATIONS):
           pass
       print(f"[budget] stopped after {MAX_ITERATIONS} iterations")
       return ""
   ```

5. Each pass, send `messages` and the tool definitions, tool choice automatic. Read back the assistant message.

   Replace `pass` with these two lines. `tools=TOOLS` sends the definitions and `tool_choice="auto"` lets the model choose between calling a tool and answering. The reply holds a list of choices; the first one's `message` is the assistant message:

   ```python
           response = client.chat.completions.create(model=model, messages=messages, tools=TOOLS, tool_choice="auto")
           message = response.choices[0].message
   ```

6. If it carries no tool calls, print its text and stop. That text is the itinerary.

   Next, still inside the `for`. `message.tool_calls` is empty (`None`) when the model answered in prose. The prose goes on the end of `messages` too, so the history is complete if a stretch goal asks a follow-up, and `return` hands the text back to whoever called `run_agent`; `or ""` covers an empty reply:

   ```python
           if not message.tool_calls:
               messages.append({"role": "assistant", "content": message.content or ""})
               return message.content or ""
   ```

   Printing happens where the function is called. At the bottom of the file, below the `messages` line from item 3:

   ```python
   answer = run_agent(client, model, messages)
   ```

   ```python
   print(answer)
   ```

7. Otherwise, append the assistant message to `messages` exactly as it came back, tool calls included.

   Next, still inside the `for`, below the `if` block. `message` is an object from the `openai` package, so this rebuilds it as a plain dict with the same fields; `[{...} for c in message.tool_calls]` makes one dict per tool call:

   ```python
           messages.append({
               "role": "assistant",
               "content": message.content,
               "tool_calls": [{"id": c.id, "type": "function", "function": {"name": c.function.name, "arguments": c.function.arguments}} for c in message.tool_calls],
           })
   ```

8. For each tool call, parse its arguments string, look up the function by name, call it.

   Next, still inside the `for`. `FUNCTIONS.get(name)` returns the function, or `None` for a name that is not in the dictionary. `call.function.arguments` is a JSON string such as `{"park": "Glacier National Park"}`; `json.loads` turns it into a dict, and `fn(**args)` passes each key as a keyword argument, so this runs `search_trails(park="Glacier National Park")`:

   ```python
           for call in message.tool_calls:
               fn = FUNCTIONS.get(call.function.name)
               args = json.loads(call.function.arguments or "{}")
               if fn is None:
                   output = json.dumps({"error": f"Unknown tool '{call.function.name}'."})
               else:
                   output = fn(**args)
   ```

9. Append the result as a `tool` message tied to that call's `id` (the message's `tool_call_id` field). Go back to 5.

   The last line inside `for call in message.tool_calls:`, indented 12 spaces. `tool_call_id` tells the model which of its calls this result answers. After it, the outer `for` starts its next pass, which is item 5 again:

   ```python
               messages.append({"role": "tool", "tool_call_id": call.id, "content": output})
   ```

The finished order in `main.py`, top to bottom: imports, `create_chat_client`, `DATA` and the helpers, the tool functions, `TOOLS`, `FUNCTIONS`, `SYSTEM_PROMPT`, `run_agent`, then the starter's `client, model = ...` and `request = ...` lines, the two `print` lines, `messages`, `answer`, and `print(answer)`. Run it:

```bash
uv run main.py
```

**Check:** one `[tool]` line prints for each of the two tools, then the itinerary. That count belongs to this step only: step 3 adds a third tool and step 4 a fourth, so once you move on you will see more `[tool]` lines and cannot re-run this check from the finished program. The plan should name trails that came back from `search_trails`, such as Trail of the Cedars, Iceberg Lake Trail, or Bowman Lake Shoreline Trail. An itinerary with no `[tool]` lines means the model never saw your tools; an itinerary naming trails you cannot find in `data/trails.json` means the tool result never reached the model.

### Step 3: Add get_weather and plan around the rain day

**Do:**
1. Write `get_weather(park)` the same way as `check_campsites`, reading `data/mock-apis/weather.json`; return `{"error": "No forecast available for '<park>'."}` if nothing matches.

   Put it below `check_campsites`, above `TOOLS`. Only the file name and the error text differ from `check_campsites`:

   ```python
   def get_weather(park: str = "Glacier National Park") -> str:
       narrate("get_weather", {"park": park})
       entry = park_entry(load("mock-apis/weather.json"), park)
       return result(entry if entry is not None else {"error": f"No forecast available for '{park}'."})
   ```

2. Write the `get_weather` tool definition yourself, with a name, a description, and one `park` parameter. Only then check the reference in `data/tool-definitions.json`:

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

In Python a definition is a dict with the same shape as that JSON. Draft yours anywhere in `main.py` (below `TOOLS` is fine) before you look; the model reads the file's version, which item 3 loads, so this draft is only for comparing wording:

```python
# Hint: fill in the two descriptions in your own words
my_weather_tool = {
    "type": "function",
    "function": {
        "name": "get_weather",
        "description": "<what the tool returns and when to call it>",
        "parameters": {
            "type": "object",
            "properties": {"park": {"type": "string", "description": "<what to pass as park>"}},
            "required": ["park"],
        },
    },
}
```

3. Declare `get_weather` alongside the first two tools and add it to the dictionary.

   Add the name to the parentheses at the end of your `TOOLS` line:

   ```python
   # Hint: the new name goes after "check_campsites", with a comma
   TOOLS = [t for t in load("tool-definitions.json")["tools"] if t["function"]["name"] in ("search_trails", "check_campsites", "<new tool name>")]
   ```

   And one more line inside `FUNCTIONS`, below `"search_trails": search_trails,`:

   ```python
       "get_weather": get_weather,
   ```

4. Replace the system prompt:

   Delete the old text between the triple quotes of `SYSTEM_PROMPT = """` and paste this in its place:

```text
You are the trip-planning agent for Trailhead Guides, a hiking app. Today's date is September 11, 2026. Plan trips using your tools; never invent trails, weather, availability, or conditions. Every trail name, forecast, and campground in your answer must have come back from a tool call in this conversation. Call the tools one at a time, in this order, and do not write any part of the itinerary until all of them have been called: 1. get_weather for the park. 2. search_trails for candidate trails. 3. check_campsites for where to stay each night. Then write the itinerary: one section per day with trail, campsite, and how the forecast shaped the choice (put harder or more exposed hiking on the drier days).
```

5. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather.
```

```bash
uv run main.py "Plan me a 3-day trip in Glacier National Park for September 14-16. I want one big hike and I don't want to be caught out in the weather."
```

**Why:** `data/mock-apis/weather.json` covers Glacier 2026-09-13 to 2026-09-18; the planted day is the 16th, rain showers at 49/33 with 0.7 precipitation chance and 18 mph wind, after two sunny days and one partly cloudy one.

**Check:** the itinerary puts the hard hike on the 14th or 15th, and the 16th gets something short or sheltered with a sentence saying why. Failing looks like the same three trails plus "expect rain on the 16th"; if the model never calls `get_weather`, fix your description first. Compare the sample in [`expected-output.md`](../expected-output.md).

### Step 4: Add get_trail_conditions and ask for the washed-out bridge on trail-0117

**Do:**
1. Write `get_trail_conditions(trail_id)`. If `trail_id` is missing/blank, return `{"error": "trailId is required. Call search_trails first and use one of its ids."}` instead of throwing.

   This version already has the forgiving stretch goal in place: a blank or `"null"` id gets an error listing the ids `search_trails` last returned, and a trail name resolves to its id. That is why its error text says `Call this tool again with one of these ids: ...` instead of the sentence above; return the code's version, which gives the model the ids to retry with. Put the function below `get_weather`, above `TOOLS`. It opens with the `[tool]` line and the missing-id check, then the name lookup: `next((t for t in ... if ...), None)` gives the first trail whose name contains the text, or `None`:

   ```python
   def get_trail_conditions(trail_id: str | None = None) -> str:
       narrate("get_trail_conditions", {"trail_id": trail_id})
       if not trail_id or trail_id.strip() in ("", "null", "string"):
           candidates = ", ".join(last_result_ids) if last_result_ids else "call search_trails first"
           return result({"error": f"trailId is required. Call this tool again with one of these ids: {candidates}."})
       if not trail_id.lower().startswith("trail-"):
           by_name = next((t for t in load("trails.json") if trail_id.lower() in t["name"].lower()), None)
           if by_name:
               trail_id = by_name["id"]
   ```

2. Open `data/condition-reports.jsonl` line by line, parsing each non-empty line as JSON (`id`, `trail_id`, `date`, `text`). Keep reports whose `trail_id` matches (case-insensitive). Sort by `date`, newest first. Keep the first 4.

   Next in the same function. `splitlines()` gives one string per line, and the list keeps each non-empty line parsed with `json.loads`. `sorted(..., key=lambda r: r["date"], reverse=True)` orders the matching reports by their `date` field, newest first, and `[:4]` keeps the first four:

   ```python
       reports = [json.loads(l) for l in (DATA / "condition-reports.jsonl").read_text().splitlines() if l.strip()]
       mine = sorted((r for r in reports if r["trail_id"].lower() == trail_id.lower()), key=lambda r: r["date"], reverse=True)[:4]
   ```

3. If none matched, return `{"error": "No condition reports found for '<trail_id>'."}`. Otherwise return a JSON array of `{date, report}` objects (`report` = the line's `text`).

   The function ends by returning one or the other. An empty list counts as false, so `if not mine` means nothing matched:

   ```python
       if not mine:
           return result({"error": f"No condition reports found for '{trail_id}'."})
       return result([{"date": r["date"], "report": r["text"]} for r in mine])
   ```

4. Declare `get_trail_conditions`:

   As in step 3, add the name to the parentheses at the end of `TOOLS`:

   ```python
   # Hint: add the new name after "get_weather", with a comma
   TOOLS = [t for t in load("tool-definitions.json")["tools"] if t["function"]["name"] in ("search_trails", "check_campsites", "get_weather", "<new tool name>")]
   ```

   And one more line inside `FUNCTIONS`:

   ```python
       "get_trail_conditions": get_trail_conditions,
   ```

   The reference definition it loads:

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

   As in step 3, paste it between the triple quotes of `SYSTEM_PROMPT = """`, replacing the step 3 prompt.

6. Send this user message and run the loop again:

```text
Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail (trail-0117).
```

```bash
uv run main.py Plan me a 3-day trip in Glacier National Park for September 14-16 that includes the Avalanche Lake Trail \(trail-0117\).
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

  Below `last_result_ids: list[str] = []` near the top, add the flag. It starts as `False`; the request parsing below sets it:

  ```python
  auto_approve_permits = False
  ```

  Put the function below `get_trail_conditions`, above `TOOLS`. `input()` prints the question and waits for a typed line. `.strip().lower() in ("y", "yes")` is `True` only for those two answers. `EOFError` is what `input()` raises when there is no input left to read (for example, when nothing was piped in), and that counts as a no:

  ```python
  def request_permit(park: str = "Glacier National Park", zone: str = "Lake McDonald / Sperry", dates: str = "unspecified", group_size: int = 2) -> str:
      narrate("request_permit", {"park": park, "zone": zone, "dates": dates, "group_size": group_size})
      print(f"  [gate] About to file a permit request: {park}, zone '{zone}', {dates}, group of {group_size}.")
      if auto_approve_permits:
          print("  [gate] --yes supplied; auto-approved.")
          approved = True
      else:
          try:
              approved = input("  [gate] File it? [y/N] ").strip().lower() in ("y", "yes")
          except EOFError:
              approved = False
      if not approved:
          return result({"status": "cancelled", "message": "The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed."})
      return result(load("mock-apis/permits.json")["submit_response"])
  ```

  All five tools are declared now, so `TOOLS` no longer needs a filter. Replace your whole `TOOLS = [...]` line with:

  ```python
  TOOLS = load("tool-definitions.json")["tools"]
  ```

  And add the last line inside `FUNCTIONS`:

  ```python
      "request_permit": request_permit,
  ```

  These three lines replace the starter's `request = " ".join(sys.argv[1:]) or ...` line, so `--yes` is not sent as part of the request. `sys.argv[1:]` is the list of words typed after `main.py`. `"--yes" in args` is `True` when the flag is one of them, and the join skips it:

  ```python
  args = sys.argv[1:]
  auto_approve_permits = "--yes" in args
  request = " ".join(a for a in args if a != "--yes") or "Plan me a 3-day trip in Glacier National Park for September 14-16."
  ```

  The model only files a permit when the plan uses a backcountry site, so ask for one. Run it once answering the question yourself (try `n`, then `y`), and once with the flag:

  ```bash
  uv run main.py Plan me a 3-day backcountry trip in Glacier National Park for September 14-16 for 2 people, camping at backcountry sites.
  uv run main.py --yes Plan me a 3-day backcountry trip in Glacier National Park for September 14-16 for 2 people, camping at backcountry sites.
  ```

  **Check:** `request_permit` never runs on the model's say-so, and a declined request ends with an itinerary that says no permit was filed, without retrying.

- **Nudge a model that stops early.** A small model often quits after one or two tools, then writes the plan from trails it never looked up. After the loop returns, compare the set of tools called against `get_weather`, `search_trails`, `get_trail_conditions`, and `check_campsites`. Track called tool names in a set that each tool adds itself to; `narrate` from step 1 already does. If any are missing, first make sure everything the previous call returned is in your message list: the loop already appended the tool turns, and it appends the final prose reply as an assistant message too. Then append a user message naming the missing tools and asking for the next call; run the loop again, up to 3 times. If every tool was called but the answer has no "day" in it, append one user message asking for the itinerary and run once more. Print `[nudge]` whenever this fires.

  At the bottom of the file, the list of required names goes directly above the `answer = run_agent(client, model, messages)` line from step 2 (shown again here so you can see where it goes):

  ```python
  REQUIRED = ["get_weather", "search_trails", "get_trail_conditions", "check_campsites"]
  answer = run_agent(client, model, messages)
  ```

  Below that line and above `print(answer)`, the retry loop. `[t for t in REQUIRED if t not in called]` lists the required names that are not in the `called` set yet; `break` leaves the loop when that list is empty. `hint` adds the trail ids from the last search when `get_trail_conditions` is one of the missing tools, and is an empty string otherwise:

  ```python
  for _ in range(3):
      missing = [t for t in REQUIRED if t not in called]
      if not missing:
          break
      print(f"[nudge] still missing: {', '.join(missing)}")
      hint = f" Use one of these trail ids: {', '.join(last_result_ids)}." if "get_trail_conditions" in missing and last_result_ids else ""
      messages.append({"role": "user", "content": f"You have not called these tools yet: {', '.join(missing)}. Call the next one now with real arguments.{hint} Do not write the itinerary yet."})
      answer = run_agent(client, model, messages)
  ```

  Directly below the loop, still above `print(answer)`, the one extra turn for an answer with no "day" in it:

  ```python
  if "day" not in answer.lower():
      print("[nudge] tools are done but no itinerary was written; asking for it.")
      messages.append({"role": "user", "content": "Every tool you need has been called. Write the final itinerary now, using only what the tools returned. Do not call any more tools."})
      answer = run_agent(client, model, messages)
  ```

  **Check:** against `llama3.2`, `[nudge]` lines appear and the run still reaches all four tools, as in [`reference-transcript.md`](../reference-transcript.md) (verbatim console output of one `complete/` run for the step 2 request, plus a second run asking for a 2-day trip including `trail-0117`, showing the agent routing around the closed trail). Against `gpt-5.5` on Azure, no `[nudge]` line prints at all.

- **Make the tools forgiving.** Give every tool parameter a default. Let `get_trail_conditions` accept a trail name where it expects an id, by looking the name up in `data/trails.json`. Treat a `features` string as a one-item array. Return an error string on bad input instead of throwing, and have `search_trails` remember the ids it last returned so the error can list them; the step 1 and step 4 code above already does all of this through `last_result_ids`. In the loop, treat an unparseable `arguments` string as an empty argument set so the defaults apply, and if the call still fails return `{"error": "Bad arguments for <name>: <reason>"}` as the tool result. The loop in `complete/` wraps both places a bad call can blow up, so a wrong keyword name comes back as an error result instead of a crash. In `run_agent`, these lines replace step 2's `args = json.loads(...)` line and the four-line `if fn is None:` / `else:` block below it. `try:` runs the indented lines, and if one raises the named error, the `except` lines run instead of the program crashing:

  ```python
              try:
                  args = json.loads(call.function.arguments or "{}")
              except json.JSONDecodeError:
                  args = {}
              if fn is None:
                  output = json.dumps({"error": f"Unknown tool '{call.function.name}'."})
              else:
                  try:
                      output = fn(**args)
                  except TypeError as e:
                      output = json.dumps({"error": f"Bad arguments for {call.function.name}: {e}"})
  ```

  **Check:** a call with `trail_id` set to `"null"` gets an error naming the valid ids back, a made-up id such as `"[insert trail IDs here]"` gets the `No condition reports found` error, and the run keeps going instead of crashing.

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
