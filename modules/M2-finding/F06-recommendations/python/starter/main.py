"""Demo starting point: the "you might also like" box, picking trails at random.
Run: python main.py [trail id or name]   (default: trail-0117, Avalanche Lake Trail)
"""

import json
import random
import sys
from pathlib import Path

DATA = Path(__file__).resolve().parents[2] / "data"
trails = json.loads((DATA / "trails.json").read_text())

query = " ".join(sys.argv[1:]) or "trail-0117"
target = next((t for t in trails if t["id"].lower() == query.lower() or query.lower() in t["name"].lower()), None)
if target is None:
    raise SystemExit(f"No trail matches '{query}'.")

print(f"You liked: {target['name']} ({target['park']})")
print("You might also like (picked at random, which is the current feature):\n")

others = [t for t in trails if t["id"] != target["id"]]
for t in random.sample(others, 5):
    print(f"  {t['name']} ({t['park']}, {t['difficulty']})")
