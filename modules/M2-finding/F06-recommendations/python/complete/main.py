"""Finished demo, matching the demo script in docs/slides/outlines:
  uv run main.py                        "more like this" for Avalanche Lake Trail
  uv run main.py trail-0008             any trail id works
  uv run main.py Trail of the Cedars    so does any name (or part of one)
  uv run main.py --gear Cascade 65      the same trick on gear, from review text

Vectors are cached in embeddings.json / gear-embeddings.json next to this
script. The cache is only checked for missing keys, so an edited description
or a different embedding model leaves the stale vectors in place. Delete the
cache file whenever the source text or the model changes.
"""

import json
import math
import sys
from collections import defaultdict
from pathlib import Path

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
EMBED_MODEL = "nomic-embed-text"
DATA = Path(__file__).resolve().parents[2] / "data"
HERE = Path(__file__).resolve().parent


def cosine(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))


# Embed each text once and cache the vectors to disk. Note the staleness trap:
# the cache is accepted whenever it holds every key, so changed text under an
# existing key keeps its old vector. Delete the file to force a re-embed.
def embed_with_cache(cache_name: str, texts: dict[str, str]) -> dict[str, list[float]]:
    cache_path = HERE / cache_name
    if cache_path.exists():
        cached = json.loads(cache_path.read_text())
        if all(k in cached for k in texts):
            return cached
    keys = list(texts)
    response = client.embeddings.create(model=EMBED_MODEL, input=[texts[k] for k in keys])
    vectors = {k: d.embedding for k, d in zip(keys, response.data)}
    cache_path.write_text(json.dumps(vectors))
    return vectors


# Products have no descriptions, so each vector comes from that product's reviews
# concatenated: "similar" here means "reviewers describe them the same way".
#
# Content similarity finds substitutes, not complements. The nearest neighbor to
# a backpack is usually another size of the same backpack, which is the one item
# its owner will never buy. Complements come from behavior data (what people buy
# or mention together), and no embedding of the product text can supply it.
def recommend_gear(query: str) -> None:
    review_text: dict[str, list[str]] = defaultdict(list)
    for line in (DATA / "gear-reviews.jsonl").read_text().splitlines():
        if line.strip():
            r = json.loads(line)
            review_text[r["product"]].append(r["text"])
    texts = {product: "\n".join(reviews) for product, reviews in review_text.items()}

    vectors = embed_with_cache("gear-embeddings.json", texts)

    target = next((p for p in texts if query.lower() in p.lower()), None)
    if target is None:
        raise SystemExit(f"No product matches '{query}'.")

    print(f"You bought: {target}")
    print("Goes well with:\n")
    hits = sorted(((cosine(vectors[target], v), p) for p, v in vectors.items() if p != target), reverse=True)[:5]
    for score, product in hits:
        print(f"  {score:.4f}  {product}")


args = sys.argv[1:]
if args and args[0] == "--gear":
    recommend_gear(" ".join(args[1:]))
    raise SystemExit

# Same catalog and same embedding model as feature 04. Recommendations need no
# new model and no new data, only the vectors search already produced.
trails = json.loads((DATA / "trails.json").read_text())
vectors = embed_with_cache("embeddings.json", {t["id"]: t["description"] for t in trails})

query = " ".join(args) or "trail-0117"
target = next((t for t in trails if t["id"].lower() == query.lower() or query.lower() in t["name"].lower()), None)
if target is None:
    raise SystemExit(f"No trail matches '{query}'.")

# "More like this" is the search from feature 04 with the query vector replaced
# by an item's own vector. That means it ranks on what the descriptions talk
# about, so whatever the prose leaves out is invisible: difficulty and distance
# live in structured fields and are never mentioned in the text, and a moderate
# family hike will cheerfully return a list of hard all-day climbs. If those
# fields matter to the user, filter or re-rank on them after the similarity pass.
print(f"You liked: {target['name']} ({target['park']})")
print("You might also like:\n")

hits = sorted(((cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"]), key=lambda h: -h[0])[:5]
for score, trail in hits:
    print(f"  {score:.4f}  {trail['name']} ({trail['park']}, {trail['difficulty']}; {', '.join(trail['features'])})")
