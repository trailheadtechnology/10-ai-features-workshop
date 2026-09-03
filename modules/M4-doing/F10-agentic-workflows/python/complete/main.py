"""Five tools over the workshop's mock APIs, wired into a hand-written tool-calling
loop. Every tool call prints as it happens, the permit step waits for a human
yes, and a step budget bounds the loop.

  python main.py                                       the capstone request
  python main.py Plan me a trip on Avalanche Lake Trail in September
  python main.py --yes <request>                       auto-approve the permit gate

Model: Azure OpenAI when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY /
AZURE_OPENAI_DEPLOYMENT are set; otherwise Ollama llama3.2, which is much
weaker at sequencing five tools. See ../F10-python.md before judging a local run.

There is no agent framework here on purpose. The loop is the same one the lab's
http/azure.http walks by hand: send the messages with the tools array, read the tool
calls out of the reply, run them, append the results, repeat.
"""

import json
import os
import sys
from pathlib import Path

from openai import AzureOpenAI, OpenAI

DATA = Path(__file__).resolve().parents[2] / "data"


def create_chat_client() -> tuple[OpenAI, str]:
    endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
    key = os.environ.get("AZURE_OPENAI_KEY")
    deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
    if endpoint and key and deployment:
        return AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment
    print("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.")
    return OpenAI(base_url="http://localhost:11434/v1", api_key="ollama"), "llama3.2"


# ---------------------------------------------------------------------------
# The five tools: ordinary Python functions over the workshop's fixture files.
# The descriptions in TOOLS below are the model's only documentation for each
# tool and parameter, so rewording them changes which tools get called and with
# what arguments. Treat that prose as behavior, not commentary. Every function
# prints itself on entry so the loop is visible while it runs.
# ---------------------------------------------------------------------------
auto_approve_permits = False
called: set[str] = set()
last_result_ids: list[str] = []


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


def park_entry(table: dict, park: str):
    first = park.split(" ")[0].lower()
    for name, value in table.items():
        if name.startswith("_"):
            continue
        if park.lower() in name.lower() or name.lower() in park.lower() or first in name.lower():
            return value
    return None


def search_trails(park: str = "Glacier National Park", features: list[str] | None = None, max_difficulty: str | None = None) -> str:
    narrate("search_trails", {"park": park, "features": features, "max_difficulty": max_difficulty})
    # Small models sometimes send a string where the schema says array.
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
        # Keywords match features OR the trail name. Without the name match a
        # request for "Avalanche Lake Trail" can never find trail-0117: it sits past
        # the 8-result cut and "Avalanche" is not a feature tag, so the washed-out
        # bridge lesson never fires (6 of 10 gpt-5.5 runs in the 2026-09-03 soak).
        if features and not any(f.lower() in t["name"].lower() or f.lower() in x.lower() for f in features for x in t["features"]):
            continue
        found.append({k: t[k] for k in ("id", "name", "park", "distance_mi", "elevation_ft", "difficulty", "features")})
        if len(found) == 8:
            break
    last_result_ids[:] = [t["id"] for t in found]
    return result(found)


def get_weather(park: str = "Glacier National Park") -> str:
    narrate("get_weather", {"park": park})
    entry = park_entry(load("mock-apis/weather.json"), park)
    return result(entry if entry is not None else {"error": f"No forecast available for '{park}'."})


def get_trail_conditions(trail_id: str | None = None) -> str:
    narrate("get_trail_conditions", {"trail_id": trail_id})
    # Every tool parameter here has a default and every failure returns an error
    # string instead of raising. A model that supplies a missing or malformed
    # argument gets a correctable message back naming the valid ids, rather than
    # crashing the process mid-loop.
    if not trail_id or trail_id.strip() in ("", "null", "string"):
        candidates = ", ".join(last_result_ids) if last_result_ids else "call search_trails first"
        return result({"error": f"trailId is required. Call this tool again with one of these ids: {candidates}."})
    # A model may pass the trail name where an id is expected, so resolve names
    # too instead of returning nothing found.
    if not trail_id.lower().startswith("trail-"):
        by_name = next((t for t in load("trails.json") if trail_id.lower() in t["name"].lower()), None)
        if by_name:
            trail_id = by_name["id"]
    reports = [json.loads(l) for l in (DATA / "condition-reports.jsonl").read_text().splitlines() if l.strip()]
    mine = sorted((r for r in reports if r["trail_id"].lower() == trail_id.lower()), key=lambda r: r["date"], reverse=True)[:4]
    if not mine:
        return result({"error": f"No condition reports found for '{trail_id}'."})
    return result([{"date": r["date"], "report": r["text"]} for r in mine])


def check_campsites(park: str = "Glacier National Park") -> str:
    narrate("check_campsites", {"park": park})
    entry = park_entry(load("mock-apis/campsites.json"), park)
    return result(entry if entry is not None else {"error": f"No campsite data for '{park}'."})


def request_permit(park: str = "Glacier National Park", zone: str = "Lake McDonald / Sperry", dates: str = "unspecified", group_size: int = 2) -> str:
    narrate("request_permit", {"park": park, "zone": zone, "dates": dates, "group_size": group_size})
    # Filing a permit is the one irreversible action in this agent, so it never
    # runs on the model's say-so; a human confirms first. --yes bypasses the
    # prompt and exists for demo runs only.
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


FUNCTIONS = {
    "search_trails": search_trails,
    "get_weather": get_weather,
    "get_trail_conditions": get_trail_conditions,
    "check_campsites": check_campsites,
    "request_permit": request_permit,
}

# The same five definitions the lab's data/tool-definitions.json ships, so what
# the model sees here is exactly what it sees from the .http file.
TOOLS = load("tool-definitions.json")["tools"]

SYSTEM_PROMPT = """You are the trip-planning agent for Trailhead Guides, a hiking app.
Today's date is September 11, 2026.

Plan trips using your tools; never invent trails, weather, availability,
or conditions. Every trail name, forecast, campground, and condition in
your answer must have come back from a tool call in this conversation.

Call the tools one at a time, in this order, and do not write any part of
the itinerary until all of them have been called:
1. get_weather for the park.
2. search_trails for candidate trails that fit the request.
3. get_trail_conditions for EVERY trail you intend to recommend, one call
   per trail, using the trail id returned by search_trails.
   If the newest reports for a trail mention a closure, a washout, a bridge
   that is out, or any other reason hikers are turning around, that trail is
   CLOSED. Do not schedule a day on a closed trail. Replace it with another
   trail from search_trails and state plainly, in the itinerary, that the
   original trail is closed and why.
4. check_campsites for where to stay each night.
5. request_permit once, only if a backcountry site or permit zone is involved.

If you have not yet called search_trails and get_trail_conditions, your
next move is a tool call, not prose.

Then write the final itinerary: one section per day with trail, campsite,
and how the forecast shaped the choice (put harder or more exposed hiking
on the drier days). End with the permit status."""


# ---------------------------------------------------------------------------
# The loop. Send everything so far, read the reply; if it asked for tools, run
# them and append the results; repeat until it answers in prose or the step
# budget runs out. That budget is the only bound on the loop: a model that keeps
# deciding to call one more tool has no other stopping condition.
# ---------------------------------------------------------------------------
MAX_ITERATIONS = 12


def run_agent(client: OpenAI, model: str, messages: list[dict]) -> str:
    for _ in range(MAX_ITERATIONS):
        response = client.chat.completions.create(model=model, messages=messages, tools=TOOLS, tool_choice="auto")
        message = response.choices[0].message
        if not message.tool_calls:
            messages.append({"role": "assistant", "content": message.content or ""})
            return message.content or ""
        # The assistant turn that made the calls has to be replayed before the
        # tool results, in the shape the API returned it.
        messages.append({
            "role": "assistant",
            "content": message.content,
            "tool_calls": [{"id": c.id, "type": "function", "function": {"name": c.function.name, "arguments": c.function.arguments}} for c in message.tool_calls],
        })
        for call in message.tool_calls:
            fn = FUNCTIONS.get(call.function.name)
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
                    # A malformed argument set gets a correctable message back,
                    # not a crash.
                    output = json.dumps({"error": f"Bad arguments for {call.function.name}: {e}"})
            messages.append({"role": "tool", "tool_call_id": call.id, "content": output})
    print(f"[budget] stopped after {MAX_ITERATIONS} iterations")
    return ""


args = sys.argv[1:]
auto_approve_permits = "--yes" in args
request = " ".join(a for a in args if a != "--yes") or "Plan me a 3-day trip in Glacier National Park for September 14-16."

client, model = create_chat_client()
messages: list[dict] = [{"role": "system", "content": SYSTEM_PROMPT}, {"role": "user", "content": request}]

print(f"Request: {request}")
print("=" * 60)

# A model can stop mid-plan believing it is done and start writing the itinerary
# from tools it never called. When a required tool is still uncalled, name it and
# let the loop continue. Capped at three nudges so a stuck model cannot spin here.
# A frontier model should need none of these; [nudge] lines mean the model is
# underpowered for the task, not that the app is broken.
REQUIRED = ["get_weather", "search_trails", "get_trail_conditions", "check_campsites"]
answer = run_agent(client, model, messages)

for _ in range(3):
    missing = [t for t in REQUIRED if t not in called]
    if not missing:
        break
    print(f"[nudge] still missing: {', '.join(missing)}")
    hint = f" Use one of these trail ids: {', '.join(last_result_ids)}." if "get_trail_conditions" in missing and last_result_ids else ""
    messages.append({"role": "user", "content": f"You have not called these tools yet: {', '.join(missing)}. Call the next one now with real arguments.{hint} Do not write the itinerary yet."})
    answer = run_agent(client, model, messages)

# The mirror failure: the model announces it has finished calling tools and then
# stops without ever writing the plan. One turn asking for it directly.
if "day" not in answer.lower():
    print("[nudge] tools are done but no itinerary was written; asking for it.")
    messages.append({"role": "user", "content": "Every tool you need has been called. Write the final itinerary now, using only what the tools returned. Do not call any more tools."})
    answer = run_agent(client, model, messages)

print("=" * 60)
print(answer)
