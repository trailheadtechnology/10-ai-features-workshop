"""Finished demo, matching the demo script in docs/slides/outlines:
  uv run main.py dog-friendly waterfall hike, not too steep
  uv run main.py somewhere quiet to take my kids
Embeds every trail description once, embeds the query, ranks by cosine
similarity, prints the top 5.
"""

import json
import math
import sys
import time
from pathlib import Path

from openai import OpenAI

# Ollama serves embeddings on the same OpenAI-compatible endpoint as chat.
client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
EMBED_MODEL = "nomic-embed-text"
DATA = Path(__file__).resolve().parents[2] / "data"

query = " ".join(sys.argv[1:]) or "dog-friendly waterfall hike, not too steep"
trails = json.loads((DATA / "trails-slice.json").read_text())


def embed(texts: list[str]) -> list[list[float]]:
    response = client.embeddings.create(model=EMBED_MODEL, input=texts)
    return [d.embedding for d in response.data]


# Embedding the catalog takes seconds, so the vectors are cached next to this
# script. The cache is keyed by trail id and nothing else: if a description
# changes, or the embedding model changes, delete embeddings.json. Otherwise
# every later query is ranked against vectors for text that no longer exists.
cache_path = Path(__file__).with_name("embeddings.json")
if cache_path.exists():
    vectors = json.loads(cache_path.read_text())
    print(f"Loaded {len(vectors)} cached vectors from embeddings.json")
else:
    started = time.perf_counter()
    embeddings = embed([t["description"] for t in trails])
    vectors = {t["id"]: e for t, e in zip(trails, embeddings)}
    cache_path.write_text(json.dumps(vectors))
    print(f"Embedded {len(vectors)} trail descriptions in {(time.perf_counter() - started) * 1000:.0f} ms")

# The query has to go through the same model that produced the cached vectors.
# Vectors from two different models are not comparable, and cosine similarity
# will still return confident-looking numbers if you mix them.
query_vector = embed([query])[0]


def cosine_similarity(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))


# This ranks on topic, not on suitability. An embedding cannot tell "great for
# kids" from "dangerous for kids" or "easy" from "never uses the word easy but
# is a cliff", so a top result can be about the right subject and still be the
# worst possible recommendation. Read the absolute scores as well: a top 5
# bunched together at a low score means nothing in the catalog is a real match
# and the order is mostly noise. Anything you can express as a filter over the
# metadata you already have (difficulty, features) is cheaper and more reliable
# than hoping the vector carries it.
results = sorted(((cosine_similarity(query_vector, vectors[t["id"]]), t) for t in trails), key=lambda r: -r[0])[:5]

print(f'\nSemantic search: "{query}"\n')
for score, trail in results:
    print(f"{score:.4f}  {trail['id']}  {trail['name']} ({trail['difficulty']}, {trail['distance_mi']} mi)  [{', '.join(trail['features'])}]")
