# Python Demo for 06 Recommendations

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: the "you might also like" box, picking five trails at random. Which is the current feature.
- `complete/main.py`: the finished demo as shown on stage. Feature 04's embedding code over this feature's own 30-trail slice (vectors cached to `embeddings.json`), "more like this" as nearest neighbors of one item's vector, and `--gear` for the same trick over product reviews, where the top hit for the Cascade 65 is the Cascade 40.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [`SETUP.md`](../../../../SETUP.md)) is the one install for all ten features. `uv run` finds it from any folder. From `complete/`: (`starter/main.py` takes no flags, at most the one positional argument its header comment names.)

```bash
uv run main.py                        # "more like this" for Avalanche Lake Trail
uv run main.py trail-0008             # any trail id works
uv run main.py Trail of the Cedars    # so does any name (or part of one)
uv run main.py --gear Cascade 65      # the same trick on gear, from review text
```

Real output for all four commands, including the neighbors that are obviously wrong, is in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F06-lab.md`](../F06-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run `uv run main.py` from the `starter/` directory (the repo root `pyproject.toml` and one `uv sync` there cover every feature, no venv to activate); the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Random Carousel

This is the recommendation feature most apps ship: five trails picked at random under "you might also like". Run it twice for the same trail and get two different lists.

Run:

```bash
uv run main.py
```

Check: Nothing about the five relates to Avalanche Lake Trail.

### Step 2: Get a Vector for Every Trail (lab step 2)

If you did feature 04, this is the same code and the same model, but not the same trails: `../data/trails.json` is a different 30-trail slice from the same 200-trail catalog (the two share 7 trails), so feature 04's cached vectors do not cover it. Either embed live, or load `../data/trail-embeddings.json`, which holds all 30 vectors precomputed for this slice (keyed by trail id, embedded from `description`), instead of calling the model at all.

```python
# Option A: embed live (same call as feature 04)
vectors = {t["id"]: e for t, e in zip(trails, embed([t["description"] for t in trails]))}
# Option B: precomputed
vectors = json.loads((DATA / "trail-embeddings.json").read_text())
```

Check: Whichever way, `vectors["trail-0117"]` is 768 floats. If you loaded the precomputed file, it is a flat id-to-vector dictionary: each key is a trail id and each value is the 768-float array.

### Step 3: Rank Every Other Trail by Similarity to the Target (lab steps 3 and 4)

"More like this" is feature 04's search with the query vector replaced by the target trail's own vector. Skip the target itself, take five.

```python
hits = sorted(((cosine(vectors[target["id"]], vectors[t["id"]]), t) for t in trails if t["id"] != target["id"]), key=lambda h: -h[0])[:5]
for score, trail in hits:
    print(f"  {score:.4f}  {trail['name']} ({trail['park']}, {trail['difficulty']}; {', '.join(trail['features'])})")
```

Run:

```bash
uv run main.py
```

Check: Gunsight Lake Approach at 0.7849 on top for Avalanche Lake Trail. Read the difficulty column: the target is a moderate family walk and most neighbors are hard. Difficulty is not in the description text, so the embedding cannot see it.

### Step 4: Do the Other Two Targets and Judge Whether You Would Ship Them (lab steps 5 and 6)

Run the other targets from `../F06-lab.md`, compare against the acceptable sets in `../expected-output.md` (there is more than one right answer), and then read your own output as a product owner. One target in this slice has no real neighbors at all; a shipping product should show nothing rather than five weak guesses.

Run:

```bash
uv run main.py trail-0008
uv run main.py Trail of the Cedars
```

Check: Substantial overlap with the acceptable sets, and a sentence from you on whether you would ship each list. Stretch: average two trails' vectors and rank against the average, or filter to the same park or an easier difficulty before ranking. `complete/` also has `--gear`, where the top hit for the Cascade 65 pack is the Cascade 40 pack: substitutes, not complements.
