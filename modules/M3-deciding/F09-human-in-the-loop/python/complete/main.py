"""Review queue: the model drafts, a human decides, and every decision is logged
to decisions.jsonl.

  python main.py                            review the queue: [a]pprove / [e]dit / [r]eject / [s]kip
  python main.py --policy                   print the routing policy table and exit
  python main.py --auto-approve-dry-run     non-interactive run for testing and CI
  python main.py ../../data/inquiries.jsonl any queue file works

SAFETY INVARIANT, load-bearing, do not weaken:
emergencies never reach the model. The policy table below routes them to
human-only and the loop skips the API call entirely, in code, before any
request is built. The system prompt also tells the model to escalate instead
of drafting, but that instruction is a request and this lane is a guarantee.
A model that ignores the instruction and writes a warm, fluent, confident
reply to someone reporting an overdue hiker is not a hypothetical; it is the
documented behavior of the model this demo ships with. Anyone editing this
file must keep the emergency path free of model calls. Adding a "just draft
it and let the reviewer catch it" shortcut here puts a reassuring lie in front
of a person who needed a dispatcher.
"""

import getpass
import json
import os
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path

from openai import OpenAI

SYSTEM_PROMPT = """You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides.
A human ranger reviews your draft before anything is sent, so write it ready to send:
friendly, plain, professional, at most two short paragraphs, signed
'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure,
state the rule and cite the source document number and section (for example
GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided;
if the excerpt does not answer the question, say a ranger will follow up with
specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
Exception: if the visitor's message reports an emergency, an injury, a possible fire,
or a missing or overdue person, do not draft a reply at all. Output exactly one line
beginning with ESCALATE: followed by a one-line reason, so the message goes straight
to dispatch."""

# The lane is chosen by what a wrong answer costs, not by how good the model is
# at the category. Everything reversible can be drafted; emergency is
# irreversible and stays human-only. Note the lookup below defaults an unknown
# category to human-only, so a category added upstream fails closed rather than
# quietly acquiring a draft lane.
POLICY = {
    "trail-condition": "draft-for-approval",
    "permit": "draft-for-approval",
    "complaint": "draft-for-approval",
    "general": "draft-for-approval",
    "lost-and-found": "draft-for-approval",
    "emergency": "human-only",
}

DATA = Path(__file__).resolve().parents[2] / "data"
HERE = Path(__file__).resolve().parent
inquiries_path = DATA / "inquiries.jsonl"
outbox_dir = HERE / "outbox"
decisions_path = HERE / "decisions.jsonl"
auto_approve = False


def print_policy() -> None:
    print("Routing policy (error cost decides the lane):")
    for category, lane in POLICY.items():
        print(f"  {category:<16} {lane}")
    print()


args = sys.argv[1:]
i = 0
while i < len(args):
    if args[i] == "--auto-approve-dry-run":
        auto_approve = True
    elif args[i] == "--outbox":
        i += 1
        outbox_dir = Path(args[i])
    elif args[i] == "--decisions":
        i += 1
        decisions_path = Path(args[i])
    elif args[i] == "--policy":
        print_policy()
        raise SystemExit
    else:
        inquiries_path = Path(args[i])
    i += 1

data_dir = inquiries_path.resolve().parent
outbox_dir.mkdir(parents=True, exist_ok=True)

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
reviewer = "auto-approve-dry-run" if auto_approve else getpass.getuser()
counts: Counter[str] = Counter()


def indent(text: str) -> str:
    return "\n".join("  | " + l.rstrip() for l in text.split("\n"))


def log(decision: dict) -> None:
    with decisions_path.open("a") as f:
        f.write(json.dumps(decision) + "\n")


def record(inquiry, lane, decision, draft, final, distance) -> dict:
    return {
        "at": datetime.now(timezone.utc).isoformat(),
        "inquiryId": inquiry["id"],
        "category": inquiry["category"],
        "lane": lane,
        "decision": decision,
        "reviewer": reviewer,
        "draft": draft,
        "final": final,
        "editDistance": distance,
    }


def read_edited(draft: str) -> str:
    print("  Type the reply you want to send. End with a single '.' on its own line.")
    print("  Press Enter on the first line to start from the draft text instead.\n")
    lines: list[str] = []
    first = True
    while True:
        try:
            line = input()
        except EOFError:
            break
        if line == ".":
            break
        if first and line == "":
            lines.append(draft)
            print("  (draft copied in; keep typing to append, '.' to finish)")
        else:
            lines.append(line)
        first = False
    edited = "\n".join(lines).strip()
    return edited or draft


# Logged on every decision so the promotion question has data behind it rather
# than a feeling. It measures how much someone typed, not whether they were
# fixing a comma or preventing a lawsuit, so it can support an argument for
# promoting a lane and must never be the only evidence for one.
# O(a*b) and unbounded by draft length; fine for a review queue, not for bulk.
def edit_distance(a: str, b: str) -> int:
    previous = list(range(len(b) + 1))
    for i, ca in enumerate(a, 1):
        current = [i] + [0] * len(b)
        for j, cb in enumerate(b, 1):
            cost = 0 if ca == cb else 1
            current[j] = min(current[j - 1] + 1, previous[j] + 1, previous[j - 1] + cost)
        previous = current
    return previous[len(b)]


print_policy()
if auto_approve:
    print("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.\n")

for line in inquiries_path.read_text().splitlines():
    if not line.strip():
        continue
    inquiry = json.loads(line)
    lane = POLICY.get(inquiry["category"], "human-only")

    print("-" * 72)
    print(f"{inquiry['id']}  ·  {inquiry['category']}  ·  {inquiry['channel']}  ·  lane: {lane}")
    print("-" * 72)
    print(indent(inquiry["text"]))
    print()

    # THE GATE. This must stay above the API call, and the API call must stay
    # below it. A human-only message is escalated and logged without a single
    # token being spent on it, so there is no draft to leak, no reviewer fatigue
    # to survive, and no sampling luck involved. The ESCALATE handling further
    # down is a backstop for emergencies that arrive miscategorized; it is never
    # the control, because it runs after the model has already had its say.
    if lane == "human-only":
        print("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.\n")
        log(record(inquiry, lane, "escalated", None, None, 0))
        counts["escalated"] += 1
        continue

    snippet_path = data_dir / "snippets" / (inquiry.get("doc") or "")
    snippet = snippet_path.read_text().strip() if inquiry.get("doc") and snippet_path.exists() else "(none on file for this message)"

    print("  drafting...", end="", flush=True)
    response = client.chat.completions.create(
        model="llama3.2",
        messages=[
            {"role": "system", "content": SYSTEM_PROMPT},
            {"role": "user", "content": f"""Reference excerpt:
{snippet}

Visitor message ({inquiry['channel']}, received {inquiry['received']}):
{inquiry['text']}

Draft the reply."""},
        ],
    )
    draft = (response.choices[0].message.content or "").strip()
    print("\r  draft:      \n")
    print(indent(draft))
    print()

    # Second layer, for an emergency that reached here under the wrong category.
    # An ESCALATE prefix is a hard stop: the draft is logged for the audit trail
    # but never offered for approval, because a reviewer presented with a
    # sendable-looking reply may send it. Do not soften this into a warning.
    if draft.upper().startswith("ESCALATE"):
        print("  Model asked to escalate. Draft discarded, routing to a human.\n")
        log(record(inquiry, lane, "escalated", draft, None, 0))
        counts["escalated"] += 1
        continue

    final: str | None = None
    if auto_approve:
        decision = "approved"
        final = draft
        print("  [auto] approved\n")
    else:
        try:
            key = input("  [a]pprove  [e]dit  [r]eject  [s]kip > ").strip().lower()
        except EOFError:
            key = "s"
        print()
        if key == "a":
            decision, final = "approved", draft
        elif key == "e":
            decision, final = "edited", read_edited(draft)
        elif key == "r":
            decision = "rejected"
        else:
            decision = "skipped"

    if final is not None:
        path = outbox_dir / f"{inquiry['id']}.txt"
        path.write_text(final + "\n")
        print(f"  -> {decision}, queued at {os.path.relpath(path)}\n")
    else:
        print(f"  -> {decision}, nothing queued\n")

    log(record(inquiry, lane, decision, draft, final, edit_distance(draft, final or "")))
    counts[decision] += 1

print("=" * 72)
print("Queue done: " + ", ".join(f"{v} {k}" for k, v in counts.items()))
print(f"Audit trail: {os.path.relpath(decisions_path)}   ·   Outbox: {os.path.relpath(outbox_dir)}/")
