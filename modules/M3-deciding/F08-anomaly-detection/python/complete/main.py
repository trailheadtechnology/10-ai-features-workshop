"""Ranks trail reports by distance from the trail's own baseline and alerts when
several outliers land close together in time. Needs Ollama running with
nomic-embed-text pulled.

  uv run main.py                  trail-0117, live embeddings, distance table + cluster alerts
  uv run main.py --raw            same trail with the task prefix removed
  uv run main.py --trail 0042     the other trail in the data folder
  uv run main.py --sigma 1.5      tighter threshold
  uv run main.py --window 30      wider clustering window, in days

Embeddings are the only model calls. Everything after them is arithmetic, so
the cost of this feature is one embedding per report and nothing per query.
"""

import json
import math
import sys
from datetime import date
from pathlib import Path

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
DATA = Path(__file__).resolve().parents[2] / "data"

trail = "0117"
sigma = 1.0
window = 14
# nomic-embed-text is trained with task prefixes (search_query:, search_document:,
# clustering:, classification:) and expects one on every input. Embedding bare text
# still returns a well-formed vector, which is why --raw fails silently rather than
# throwing, but the vectors land off-distribution and the ranking degrades badly.
# Do not drop this prefix, and if you change it, change it for every input in the
# same corpus: vectors embedded under different prefixes are not comparable.
prefix = "classification: "
args = sys.argv[1:]
i = 0
while i < len(args):
    if args[i] == "--trail":
        i += 1
        trail = args[i]
    elif args[i] == "--sigma":
        i += 1
        sigma = float(args[i])
    elif args[i] == "--window":
        i += 1
        window = int(args[i])
    elif args[i] == "--raw":
        prefix = ""
    else:
        raise SystemExit(f"unknown argument: {args[i]}")
    i += 1


def normalize(vector: list[float]) -> list[float]:
    length = math.sqrt(sum(v * v for v in vector))
    return [v / length for v in vector]


# Correct only for unit-length inputs, where the dot product is already the cosine
# similarity. Every vector reaching this function has been through normalize; pass an
# unnormalized one and it returns a number that still looks plausible.
def cosine_distance(a: list[float], b: list[float]) -> float:
    return 1.0 - sum(x * y for x, y in zip(a, b))


def truncate(text: str, n: int) -> str:
    return text if len(text) <= n else text[: n - 1] + "…"


reports = [json.loads(l) for l in (DATA / f"reports-{trail}.jsonl").read_text().splitlines() if l.strip()]

response = client.embeddings.create(model="nomic-embed-text", input=[prefix + r["text"] for r in reports])
vectors = [normalize(d.embedding) for d in response.data]

# The centroid is this trail's definition of normal, and it is built from the
# same reports it is about to judge. Anomalies that appear often enough pull the
# centroid toward themselves and stop looking anomalous, so a long-running
# detector should rebuild this from a trailing window rather than the full history.
dimensions = len(vectors[0])
centroid = [0.0] * dimensions
for vector in vectors:
    for k, v in enumerate(vector):
        centroid[k] += v / len(vectors)
centroid = normalize(centroid)

scored = sorted(((cosine_distance(v, centroid), r) for v, r in zip(vectors, reports)), key=lambda x: -x[0])

# The threshold is derived from this corpus rather than hard-coded, so it travels
# to a trail with a different spread of distances. It is still a business choice,
# not a boundary the data hands you: sigma decides how much review you are willing
# to pay for, and there is no value of it that separates incidents from oddities.
mean = sum(d for d, _ in scored) / len(scored)
deviation = math.sqrt(sum((d - mean) ** 2 for d, _ in scored) / len(scored))
threshold = mean + sigma * deviation

print(f"trail-{trail} · {len(reports)} reports · nomic-embed-text ({dimensions} dims)"
      + (" · NO task prefix" if not prefix else f' · prefix "{prefix.strip()}"'))
print(f"mean distance {mean:.4f} · sd {deviation:.4f} · threshold mean+{sigma:g}sd = {threshold:.4f}\n")

print("  dist    id       date        report")
for distance, report in scored:
    print(f"{' !' if distance > threshold else '  '}{distance:.4f}  {report['id']}  {report['date']}  {truncate(report['text'], 62)}")

# Corroboration is what makes this alertable. A single report far from normal is
# usually just an unusual subject, not an incident; two or more inside the window
# mean several people independently noticed the same thing. Requiring two is what
# keeps the alert queue small enough that someone still reads it.
flagged = sorted(((d, r) for d, r in scored if d > threshold), key=lambda x: x[1]["date"])
print(f"\n{len(flagged)} of {len(scored)} reports above threshold. Clustering them within {window} days:\n")

alerts = 0
i = 0
while i < len(flagged):
    j = i + 1
    while j < len(flagged) and (date.fromisoformat(flagged[j][1]["date"]) - date.fromisoformat(flagged[j - 1][1]["date"])).days <= window:
        j += 1
    group = flagged[i:j]
    if len(group) >= 2:
        alerts += 1
        print(f"  ALERT trail-{trail}: {len(group)} anomalous reports between {group[0][1]['date']} and {group[-1][1]['date']}")
        for _, r in group:
            print(f"        {r['id']} {r['date']}  {truncate(r['text'], 70)}")
    else:
        print(f"  (ignored) {group[0][1]['id']} {group[0][1]['date']} is a lone outlier, not an event")
    i = j
print(f"\n{alerts} alert(s). Model calls: {len(reports)} embeddings, 0 chat completions.")
