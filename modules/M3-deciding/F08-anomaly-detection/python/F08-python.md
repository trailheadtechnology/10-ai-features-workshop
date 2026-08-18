# Python Demo for 08 Anomaly Detection

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: no model, no network. Loads the precomputed vectors from `../data/embeddings-0117.json`, averages them into a centroid, ranks every report by cosine distance from it. Runs with Ollama down.
- `complete/main.py`: the finished demo as shown on stage. Embeds live with the `classification:` task prefix, derives the threshold from the corpus (mean plus sigma standard deviations), and applies the alert rule: two or more flagged reports within a 14-day window. One alert fires, three genuine washout reports in it.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py                  # trail-0117, live embeddings, distance table + cluster alerts
python main.py --raw            # same trail with the task prefix removed
python main.py --trail 0042     # the bear-activity trail
python main.py --sigma 1.5      # tighter threshold
python main.py --window 30      # wider clustering window, in days
```

Embeddings are the only model calls; everything after them is arithmetic. The recorded ranking and the single alert are in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F08-lab.md`](../F08-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: Centroid Distance from Precomputed Vectors

Lab steps 1 and 2 are already in the starter, and it needs no model: it loads the 40 vectors in `../data/embeddings-0117.json` (embedded with the `classification:` prefix), averages them into a centroid, and ranks every report by cosine distance from it. Read the top of the list.

Run:

```bash
python main.py
```

Check: Washout reports rise toward the top, but not cleanly: routine reports about parking or wildflowers are mixed in. Compare the ranking in `../expected-output.md`. That is what the technique does out of the box.

### Step 2: Embed Live, First Without the Task Prefix (lab step 3, first half)

Replace the precomputed file with a live embedding call, and deliberately embed the bare text. `nomic-embed-text` expects a task prefix on every input; without one it still returns a well-formed vector, so nothing throws, but the vectors land off-distribution and the ranking degrades. Watch where the first washout report lands.

```python
client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
prefix = ""                            # step 2: bare text. step 3: "classification: "
response = client.embeddings.create(model="nomic-embed-text", input=[prefix + r["text"] for r in reports])
vectors = [normalize(d.embedding) for d in response.data]
```

Run:

```bash
python main.py
```

Check: The first washout report sits around rank 11 (`complete/ --raw` reproduces this). Then set the prefix to `"classification: "` and re-run: it jumps to rank 2 and the mud reports settle to the bottom. Reading the model card is engineering work.

### Step 3: Derive a Threshold and Add the Alert Rule (lab step 3, second half, and where the feature actually lives)

A threshold from the corpus (mean plus one standard deviation) flags outliers; the alert rule requires two or more flagged reports within 14 days of each other. One outlier is a rambling hiker; several in a week that also sit near each other is an event.

```python
mean = sum(d for d, _ in scored) / len(scored)
sd = math.sqrt(sum((d - mean) ** 2 for d, _ in scored) / len(scored))
threshold = mean + sigma * sd

flagged = sorted(((d, r) for d, r in scored if d > threshold), key=lambda x: x[1]["date"])
i = 0
while i < len(flagged):
    j = i + 1
    while j < len(flagged) and (date.fromisoformat(flagged[j][1]["date"]) - date.fromisoformat(flagged[j - 1][1]["date"])).days <= window:
        j += 1
    group = flagged[i:j]
    if len(group) >= 2:
        print(f"  ALERT trail-{trail}: {len(group)} anomalous reports between {group[0][1]['date']} and {group[-1][1]['date']}")
    else:
        print(f"  (ignored) {group[0][1]['id']} is a lone outlier, not an event")
    i = j
```

Run:

```bash
python main.py
```

Check: Exactly one alert on trail-0117, three genuine washout reports in it (`cr-0429`, `cr-0436`, `cr-0464`), lone outliers ignored. Count the model calls: 40 embeddings, zero chat completions. Stretch: build the centroid only from reports before the washout window and watch all eight washout reports reach the top 10; or run `--trail 0042` and catch the bear-activity spike, which separates more sharply.
