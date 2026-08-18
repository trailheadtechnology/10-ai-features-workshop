# Python Demo for 03 Sentiment

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: one client, one `classify` function, one review (`gr-0007`, the sarcastic two-star). Prints the review and what `phi3` says.
- `complete/main.py`: the finished demo as shown on stage. Both review sets through both models with the byte-identical four-line prompt, a table with disagreements flagged, accuracy per set against `reference-labels.json`, and the disagreement list with a verdict on who was right.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py            # both sets, both models
python main.py --easy     # easy set only (demo steps 3 and 4)
python main.py --hard     # hard set only (demo step 5)
```

Without the Azure variables, `llama3.2` stands in for the big model so the whole comparison runs offline; that pairing ties (9/10 easy, 7/10 hard for both); against the workshop's `gpt-4.1` deployment the frontier model scores 10/10 on both sets, and both results are in [`../expected-output.md`](../expected-output.md). Keep the prompt's line breaks exactly as they are: reflowing it onto one line costs `phi3` measured accuracy.

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, the deployment name the feature uses, and the key handed out in the room) and the big model switches to Azure OpenAI through the SDK's `AzureOpenAI` client; leave them unset and it runs against Ollama.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F03-lab.md`](../F03-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter on the Sarcastic Review

One client, one `classify` function, one review: `gr-0007`, two stars, "five-star experience, truly". The prompt is the whole feature and it is byte-identical to `../ollama.http`; keep its line breaks, because reflowing it onto one line costs `phi3` measured accuracy.

Run:

```bash
python main.py
```

Check: `phi3 says: negative`. Try `gr-0034` or any other id from `easy.jsonl` / `hard.jsonl`.

### Step 2: Loop the Easy Set and Score It Against the Reference Labels (lab step 1)

Replace the single review with a loop over `easy.jsonl`, look each id up in `reference-labels.json`, and count matches.

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

Run:

```bash
python main.py
```

Check: 9/10 on the easy set in the recorded runs. Yours may differ by one.

### Step 3: Add the Second Model and Run the Hard Set Through Both (lab step 2)

Build a second client: Azure OpenAI if you have the room key in `AZURE_OPENAI_KEY` (with `AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com` and `AZURE_OPENAI_DEPLOYMENT=gpt-4.1`), otherwise `llama3.2` on the same Ollama as a stand-in. Nothing in `classify` changes; that is the provider-swap point of the whole module. Then run `hard.jsonl` through both.

```python
big = (AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment)
# or, offline: (ollama, "llama3.2")
small = classify(*small_target, review["text"])
big_label = classify(*big, review["text"])
```

Run:

```bash
python main.py
```

Check: Two columns of labels for the hard set. Recorded: 7/10 for `phi3`, 10/10 for `gpt-4.1` on Azure, and 7/10 for the `llama3.2` stand-in. The frontier model earns its price on this slice; the local stand-in would have told you otherwise.

### Step 4: Print the Disagreement List and Call Each One (lab step 3, the success check)

Every review where the two models differ, with the reference label and a verdict on who was right. This list is the actual deliverable of the feature: it is what tells you which slice of your traffic needs the expensive model.

```python
for d in (r for r in results if r.small != r.big):
    verdict = "big right" if d.big == d.reference else "phi3 right" if d.small == d.reference else "both wrong"
    print(f"{d.review['id']} [{d.set}] ref={d.reference} phi3={d.small} big={d.big}  ({verdict})")
```

Check: Your version of the two tables in `../expected-output.md`: accuracy per set per model, and the disagreements with your call on each. Stretch: change the label to `{overall, aspects: {comfort, durability, price}}` with structured output and see which model can go deeper.
