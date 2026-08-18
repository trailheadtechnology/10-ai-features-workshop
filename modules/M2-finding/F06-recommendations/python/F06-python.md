# Python Demo for 06 Recommendations

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: the "you might also like" box, picking five trails at random. Which is the current feature.
- `complete/main.py`: the finished demo as shown on stage. Same trail vectors as feature 04 (cached to `embeddings.json`), "more like this" as nearest neighbors of one item's vector, and `--gear` for the same trick over product reviews, where the top hit for the Cascade 65 is the Cascade 40.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py                        # "more like this" for Avalanche Lake Trail
python main.py trail-0008             # any trail id works
python main.py Trail of the Cedars    # so does any name (or part of one)
python main.py --gear Cascade 65      # the same trick on gear, from review text
```

Real output for all four commands, including the neighbors that are obviously wrong, is in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F06-lab.md`](../F06-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Random Carousel

This is the recommendation feature most apps ship: five trails picked at random under "you might also like". Run it twice for the same trail and get two different lists.

Run:

```bash
python main.py
```

Check: Nothing about the five relates to Avalanche Lake Trail.

### Step 2: Get a Vector for Every Trail (lab step 1, first half)

If you did feature 04, this is the same code and the same model. If you did not, `../data/trail-embeddings.json` has the 30 vectors precomputed (keyed by trail id, embedded from `description`) and you can load them instead of calling the model at all.

```python
# Option A: embed live (same as feature 04)
vectors = {t["id"]: e for t, e in zip(trails, embed([t["description"] for t in trails]))}
# Option B: precomputed
vectors = json.loads((DATA / "trail-embeddings.json").read_text())
```

Check: Whichever way, `vectors["trail-0117"]` is 768 floats. If you loaded the precomputed file, check its top-level shape first; it may wrap the vectors in an object.

### Step 3: Rank Every Other Trail by Similarity to the Target (lab step 1, second half)

"More like this" is feature 04's search with the query vector replaced by the target trail's own vector. Skip the target itself, take five.

```python
hits = sorted(((cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"]), key=lambda h: -h[0])[:5]
for score, trail in hits:
    print(f"  {score:.4f}  {trail['name']} ({trail['park']}, {trail['difficulty']}; {', '.join(trail['features'])})")
```

Run:

```bash
python main.py
```

Check: Gunsight Lake Approach at 0.7849 on top for Avalanche Lake Trail. Read the difficulty column: the target is a moderate family walk and most neighbors are hard. Difficulty is not in the description text, so the embedding cannot see it.

### Step 4: Do the Other Two Targets and Judge Whether You Would Ship Them (lab steps 2 and 3)

Run the other targets from `../F06-lab.md`, compare against the acceptable sets in `../expected-output.md` (there is more than one right answer), and then read your own output as a product owner. One target in this slice has no real neighbors at all; a shipping product should show nothing rather than five weak guesses.

Run:

```bash
python main.py trail-0008
python main.py Trail of the Cedars
```

Check: Substantial overlap with the acceptable sets, and a sentence from you on whether you would ship each list. Stretch: average two trails' vectors and rank against the average, or filter to the same park or an easier difficulty before ranking. `complete/` also has `--gear`, where the top hit for the Cascade 65 pack is the Cascade 40 pack: substitutes, not complements.
