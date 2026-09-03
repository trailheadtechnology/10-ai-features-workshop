"""Demo starting point: the naive approach. Ask for JSON in the prompt, get
back whatever the model feels like: prose preamble, markdown fences, drifting
field names. This is what the parsed, typed response replaces.
Run: uv run main.py [path-to-trip-report.md]
"""

import sys
from pathlib import Path

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
DATA = Path(__file__).resolve().parents[2] / "data"


def strip_front_matter(markdown: str) -> str:
    parts = markdown.split("---", 2)
    return parts[2].strip() if len(parts) == 3 else markdown.strip()


report_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DATA / "tr-0007.md"
report = strip_front_matter(report_path.read_text())

response = client.chat.completions.create(
    model="llama3.2",
    messages=[{"role": "user", "content": f"Extract the details of this trip report as JSON.\n\n{report}"}],
)
print(response.choices[0].message.content)
