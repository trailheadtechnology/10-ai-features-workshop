"""Starting point, and deliberately unsafe. Every draft goes straight out with no
human between the model and the visitor, and nothing is logged, so there is no
record of what was sent or any way to find out later.
Run: python main.py [path-to-inquiries.jsonl]

Nothing here treats an emergency differently. The system prompt asks the model
to escalate instead of drafting, and the model is free to ignore that and
often does, which is why complete/ moves the decision out of the prompt and
into a policy check in code. Do not use this shape on a real inbox.
"""

import json
import sys
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

DATA = Path(__file__).resolve().parents[2] / "data"
inquiries_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DATA / "inquiries.jsonl"
data_dir = inquiries_path.resolve().parent

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")

for line in inquiries_path.read_text().splitlines():
    if not line.strip():
        continue
    inquiry = json.loads(line)

    snippet_path = data_dir / "snippets" / (inquiry.get("doc") or "")
    snippet = snippet_path.read_text().strip() if inquiry.get("doc") and snippet_path.exists() else "(none on file for this message)"

    draft = client.chat.completions.create(
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

    # No review step and no audit record. In a real deployment this line is the
    # send call, and by the time anyone reads the output the mail has gone.
    print(f"=== SENT to visitor · {inquiry['id']} ({inquiry['category']}) ===")
    print((draft.choices[0].message.content or "").strip())
    print()

print("All replies sent. Nobody read them. Nothing was logged.")
