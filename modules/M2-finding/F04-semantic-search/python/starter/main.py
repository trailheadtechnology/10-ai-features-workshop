"""Demo starting point: keyword search over the trail slice. Split the query into
words and count how many appear, as whole words, in each trail's name and
description.
Run: uv run main.py <query>          (defaults to the demo query)

Matching on words alone has no access to meaning. A description saying the
trail avoids the steep section still matches "steep", and a trail whose text
reads "leashed dogs are welcome" never matches "dog-friendly". That gap is
what the complete script replaces with embeddings.
"""

import json
import re
import sys
from pathlib import Path

DATA = Path(__file__).resolve().parents[2] / "data"

query = " ".join(sys.argv[1:]) or "dog-friendly waterfall hike, not too steep"
trails = json.loads((DATA / "trails-slice.json").read_text())

tokens = list(dict.fromkeys(w for w in re.findall(r"[a-z]+", query.lower()) if len(w) >= 3))

results = []
for t in trails:
    haystack = f"{t['name']} {t['description']}".lower()
    hits = [w for w in tokens if re.search(rf"\b{w}\b", haystack)]
    if hits:
        results.append((t, hits))
results.sort(key=lambda r: (-len(r[1]), r[0]["id"]))
results = results[:5]

print(f'Keyword search: "{query}"')
print(f"Query words: {', '.join(tokens)}\n")

if not results:
    print("No results. Not one trail contains those words.")
    raise SystemExit

for trail, hits in results:
    print(f"{len(hits)} word(s) [{', '.join(hits)}]  {trail['id']}  {trail['name']} ({trail['difficulty']}, {trail['distance_mi']} mi)")
