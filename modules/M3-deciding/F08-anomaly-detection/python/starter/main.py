"""Starting point: ranks trail reports by distance from the trail's baseline.
Run: python main.py

Makes no model calls and needs no network. The vectors in
../../data/embeddings-0117.json were precomputed with nomic-embed-text so this
runs with Ollama down, which is the fallback when the room's network is not
cooperating. Those vectors were embedded with the "classification: " task
prefix nomic requires, so anything you add to this corpus must be embedded the
same way or its distances will not be comparable to these.

complete/ embeds live and adds the alert rule on top of this ranking.
"""

import json
import math
from pathlib import Path

DATA = Path(__file__).resolve().parents[2] / "data"


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


reports = [json.loads(l) for l in (DATA / "reports-0117.jsonl").read_text().splitlines() if l.strip()]
vectors = {k: normalize(v) for k, v in json.loads((DATA / "embeddings-0117.json").read_text())["embeddings"].items()}

# The centroid is this trail's definition of normal, and it is built from the
# same reports it is about to judge. Anomalies that appear often enough pull the
# centroid toward themselves and stop looking anomalous.
dimensions = len(next(iter(vectors.values())))
centroid = [0.0] * dimensions
for vector in vectors.values():
    for i, v in enumerate(vector):
        centroid[i] += v / len(vectors)
centroid = normalize(centroid)

scored = sorted(((cosine_distance(vectors[r["id"]], centroid), r) for r in reports), key=lambda x: -x[0])

print(f"trail-0117 · {len(reports)} reports · {dimensions}-dim nomic-embed-text vectors\n")
print("  dist    id       date        report")
for distance, report in scored:
    print(f"  {distance:.4f}  {report['id']}  {report['date']}  {truncate(report['text'], 62)}")
