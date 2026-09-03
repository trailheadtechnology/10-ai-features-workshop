"""Classifies every inquiry in ../../data/inquiries-slice.jsonl and scores the
result against ../../data/reference-labels.json.
Run: uv run main.py

The category comes back as a Python Enum through structured output, so the
model can only return a label the routing table already knows how to handle.
Adding a category here means adding a routing destination for it too, or the
lookup on the routing table will fail.
"""

import json
from enum import Enum
from pathlib import Path

from openai import OpenAI
from pydantic import BaseModel

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
MODEL = "llama3.2"
DATA = Path(__file__).resolve().parents[2] / "data"


class Category(str, Enum):
    permit = "permit"
    conditions = "conditions"
    complaint = "complaint"
    lost_and_found = "lost-and-found"
    emergency = "emergency"
    general = "general"
    unsure = "unsure"


class TriageResult(BaseModel):
    category: Category


# These descriptions are the taxonomy, and editing them changes behavior more
# than any code below. Two rules constrain any rewrite. Emergency wins over
# every other category, including messages that also mention a permit or a lost
# item, so the ordering paragraph at the end must stay. And unsure has to stay
# narrow: it means two queues must both act on one message, not that the model
# found the message hard. Widen it and it fills up with ordinary traffic, which
# is the unsorted inbox this system replaced.
def prompt(text: str) -> str:
    return f"""You are the triage system for the Trailhead Guides shared inbox.
Classify the visitor message into exactly one category.

- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, including billing problems and missing confirmations
  for a permit application.
- conditions: asking whether a trail, road, or area is open, safe, or
  passable right now: snow, water levels, washouts, wildlife activity,
  closures.
- complaint: unhappy about a park facility, service, or staff member
  and wants it acknowledged or fixed.
- lost-and-found: reporting a lost or found physical item.
- emergency: a person may be hurt, missing, or in danger right now and
  needs immediate human attention.
- general: anything else: park rules, fees, trip planning, questions
  that fit none of the above.
- unsure: two different queues both have to act before this message can
  be resolved, so no single queue owns it. The case that qualifies: the
  sender asks about trail conditions AND asks someone to change, refund,
  or cancel a booking. Trail info cannot issue a refund, and the permits
  office does not decide whether a trail is passable, so a human reads
  this queue and splits the work. Also use unsure when the message fits
  none of the categories above.

Decide in this order. First, if anyone might be hurt, missing, or in
danger, answer emergency and stop; never answer unsure for those, even
when the message also mentions permits, conditions, or a lost item.
Second, if one queue can resolve the whole message on its own, answer
that queue; a booking or reservation problem with nothing else attached
is permit, not unsure. Third, only if two queues must both act, answer
unsure. Unsure is not a catch-all for anything hard.

Message:
{text}"""


def clip(text: str, n: int) -> str:
    return text if len(text) <= n else text[:n] + "..."


inquiries = [json.loads(l) for l in (DATA / "inquiries-slice.jsonl").read_text().splitlines() if l.strip()]
reference = json.loads((DATA / "reference-labels.json").read_text())

results: list[tuple[dict, Category]] = []
for inquiry in inquiries:
    response = client.chat.completions.parse(
        model=MODEL,
        messages=[{"role": "user", "content": prompt(inquiry["text"])}],
        response_format=TriageResult,
        # Anything above 0 makes the same message land in different queues on
        # different runs, which makes a scored comparison against fixed
        # reference labels meaningless.
        temperature=0,
    )
    results.append((inquiry, response.choices[0].message.parsed.category))
    print(".", end="", flush=True)
print("\n")

# Emergencies print before the routing table and are sorted to the top of it.
# A person scanning this output under time pressure must not have to read past
# the first screen to find one.
emergencies = [(i, c) for i, c in results if c == Category.emergency]
if emergencies:
    print("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!")
    for inquiry, _ in emergencies:
        print(f"!!! {inquiry['id']}  {clip(inquiry['text'], 70)}")
    print()

print(f"{'id':<10} {'category':<15} routed to")
print("-" * 62)
for inquiry, category in sorted(results, key=lambda r: r[1] != Category.emergency):
    print(f"{inquiry['id']:<10} {category.value:<15} {reference['routing'][category.value]}")

# Two scores, and they are not equally important. Overall accuracy is the
# headline number; recall on the emergency class is the one that decides
# whether this taxonomy is safe to ship. A missed emergency is a person waiting
# in a queue nobody is watching, and no amount of accuracy elsewhere offsets it.
# If you tune the category descriptions, judge the change on emergency recall
# first and treat a drop there as a failure even when accuracy improves.
labels = reference["labels"]
correct = sum(1 for i, c in results if c.value == labels[i["id"]])
emergency_ids = [k for k, v in labels.items() if v == "emergency"]
caught = sum(1 for i, _ in emergencies if i["id"] in emergency_ids)

print()
print(f"Accuracy vs reference labels: {correct}/{len(results)}")
print(f"Emergency recall: {caught}/{len(emergency_ids)} "
      + ("(all caught; the metric that matters)" if caught == len(emergency_ids) else "(MISSED ONE; this fails, whatever the accuracy says)"))
for inquiry, category in results:
    if category.value != labels[inquiry["id"]]:
        print(f"  miss: {inquiry['id']} got {category.value}, reference says {labels[inquiry['id']]}")
