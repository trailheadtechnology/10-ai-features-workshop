"""Finished demo, matching the demo script in docs/slides/outlines:
  python main.py                              the naive prompt (the book report)
  python main.py --briefing                   3-bullet hiker briefing
  python main.py --headline                   one-line trail status for a card UI
  python main.py --briefing --audience ranger
  Any non-flag argument is a path to a different trip report.
"""

import sys
from pathlib import Path

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
MODEL = "llama3.2"
DATA = Path(__file__).resolve().parents[2] / "data"


def strip_front_matter(markdown: str) -> str:
    parts = markdown.split("---", 2)
    return parts[2].strip() if len(parts) == 3 else markdown.strip()


report_path = DATA / "tr-0004.md"
mode = "naive"
audience = "hiker"
args = sys.argv[1:]
i = 0
while i < len(args):
    if args[i] == "--briefing":
        mode = "briefing"
    elif args[i] == "--headline":
        mode = "headline"
    elif args[i] == "--audience":
        i += 1
        audience = args[i]
    else:
        report_path = Path(args[i])
    i += 1

report = strip_front_matter(report_path.read_text())

audience_focus = (
    "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
    if audience == "ranger"
    else "a hiker planning to hike this trail within the next week"
)

if mode == "naive":
    # Deliberately the weak prompt. It is kept so the two prompts can be run
    # back to back against the same report; nothing about it should be fixed.
    prompt = f"Summarize this trip report.\n\n{report}"
elif mode == "briefing":
    # The last two lines are load-bearing, not politeness. The bullets require a
    # hazards slot and require any hazard to come first, so on a report with no
    # hazard the model will promote the nearest noun (a bear, a creek, the word
    # "avalanche" in the trail name) into a closure. Giving it a legal way to
    # report nothing is what stops that. Measurements in ../../expected-output.md.
    prompt = f"""You are helping {audience_focus}.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.

{report}"""
elif mode == "headline":
    # Same client, same report, same call. Only the instruction changes to fit a
    # different UI slot, so no new infrastructure is needed for a new surface.
    prompt = f"""From the trip report below, write ONE line of at most 12 words,
suitable for a status badge on a trail card in an app.
Lead with the most important condition or closure. No preamble.

{report}"""
else:
    raise SystemExit(f"unknown mode {mode}")

response = client.chat.completions.create(model=MODEL, messages=[{"role": "user", "content": prompt}])
print(response.choices[0].message.content)
