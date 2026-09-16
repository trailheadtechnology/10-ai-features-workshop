# Python Demo for 03 Sentiment

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: one client, one `classify` function, one review (`gr-0007`, the sarcastic two-star). Prints the review and what `phi3` says.
- `complete/main.py`: the finished demo. Both review sets through both models with the byte-identical four-line prompt, a table with disagreements flagged, accuracy per set, and the disagreement list with a verdict on who was right.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [`SETUP.md`](../../../../SETUP.md)) is the one install for all ten features. From `complete/`:

```bash
uv run main.py            # both sets, both models
uv run main.py --easy     # easy set only (lab step 3-4 shape)
uv run main.py --hard     # hard set only (lab step 5 shape)
```

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, deployment `gpt-4.1`, key handed out in the room) to use Azure OpenAI as the big model; leave them unset and `llama3.2` on Ollama stands in, so the whole comparison runs offline. The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`); swapping providers later is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps are in [`../F03-lab.md`](../F03-lab.md); this maps each onto `starter/main.py` → `complete/main.py`. Edit the starter in place (or copy it first); `complete/`'s comments say why each piece is there.

### Lab step 0: run the starter

```bash
uv run main.py
```

Check: `phi3 says: negative` on `gr-0007`.

### Lab steps 1-2: load reference labels, loop the easy set, score it

```python
labels = json.loads((DATA / "reference-labels.json").read_text())
correct = total = 0
for line in (DATA / "easy.jsonl").read_text().splitlines():
    review = json.loads(line)
    label = classify(client, "phi3", review["text"])
    reference = labels[review["id"]]["label"]
    print(f"{review['id']:<9} {reference:<10} {label:<10}")
    total += 1
    correct += label == reference
print(f"phi3 {correct}/{total}")
```

Check: 9/10 on the easy set in the recorded runs; yours may differ by one.

### Lab steps 3-4: second client, hard set through both models

```python
big = (AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment)
# or, offline: (ollama, "llama3.2")
small = classify(*small_target, review["text"])
big_label = classify(*big, review["text"])
```

Check: 7/10 for `phi3` on the hard set, 10/10 for `gpt-4.1`, 8/10 for the `llama3.2` stand-in.

### Lab step 5: disagreement list

```python
for d in (r for r in results if r.small != r.big):
    verdict = "big right" if d.big == d.reference else "phi3 right" if d.small == d.reference else "both wrong"
    print(f"{d.review['id']} [{d.set}] ref={d.reference} phi3={d.small} big={d.big}  ({verdict})")
```

Check: your version of the two tables in `../expected-output.md`. Stretch: change the label to `{overall, aspects: {comfort, durability, price}}` with structured output.
