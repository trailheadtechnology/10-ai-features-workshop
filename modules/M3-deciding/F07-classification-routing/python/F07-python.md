# Python Demo for 07 Classification & Routing

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: classify a single inquiry and print the free-text label, with nothing stopping the model from returning a label that does not exist.
- `complete/main.py`: the finished demo as shown on stage. Every inquiry in the slice classified through structured output into a Python `Enum`, so the model can only return a label the routing table knows; emergencies printed first; accuracy against the reference labels and, separately, emergency recall, which is the number that matters.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [`SETUP.md`](../../../../SETUP.md)) is the one install for all ten features. `uv run` finds it from any folder. From `complete/`: (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
uv run main.py             # all 20, scored
uv run main.py inq-0013    # (starter) one inquiry
```

The taxonomy lives in the prompt string, and editing those descriptions changes behavior more than any code. Temperature is pinned at 0 so a scored run means something. The recorded 17/20 with 2/2 emergencies is in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F07-lab.md`](../F07-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Try a Few Ids

One inquiry, free-text label. Try `inq-0035`, the ambiguous one, and watch the answer vary between `conditions` and `permit`; try `inq-0013` and check that it says `emergency`. Nothing stops the model returning "Emergency." or a sentence.

Run:

```bash
uv run main.py
uv run main.py inq-0035
uv run main.py inq-0013
```

Check: A label per run, and at least one that is not exactly one of the seven category names.

### Step 2: Make the Label an Enum Through Structured Output, and Loop All 20 (lab step 1)

The category becomes a type with seven values, so the model can only return a label the routing table knows how to handle. Keep the taxonomy prompt from the starter, pin temperature at 0 (anything above it makes a scored comparison against fixed labels meaningless), and loop the slice.

```python
class Category(str, Enum):
    permit = "permit"; conditions = "conditions"; complaint = "complaint"
    lost_and_found = "lost-and-found"; emergency = "emergency"; general = "general"; unsure = "unsure"

class TriageResult(BaseModel):
    category: Category

for inquiry in inquiries:
    response = client.chat.completions.parse(
        model="llama3.2", messages=[{"role": "user", "content": prompt(inquiry["text"])}],
        response_format=TriageResult, temperature=0)
    results.append((inquiry, response.choices[0].message.parsed.category))
```

Check: Twenty labels, every one of them one of the seven strings. Note that `inq-0035` now lands in `unsure`: the enum made it a live option rather than a paragraph the model skims past.

### Step 3: Score Against the Reference Labels, and Score Emergency Recall Separately (lab step 1, the scoring pass)

`../data/reference-labels.json` has `labels` (id to category) and `routing` (category to queue). Two numbers: overall accuracy, and recall on the emergency class. They are not equally important. Missing an emergency fails the lab at 19/20.

```python
labels = reference["labels"]
correct = sum(1 for i, c in results if c.value == labels[i["id"]])
emergency_ids = [k for k, v in labels.items() if v == "emergency"]
caught = sum(1 for i, c in results if c == Category.emergency and i["id"] in emergency_ids)
print(f"Accuracy vs reference labels: {correct}/{len(results)}")
print(f"Emergency recall: {caught}/{len(emergency_ids)}")
```

Run:

```bash
uv run main.py
```

Check: Recorded: 17/20 and 2/2. Print each miss with what the model said and what the reference says; that list drives the next step.

### Step 4: Fix a Miss by Editing a Description, Not the Code (lab step 2)

`inq-0030` (a wedding photographer asking whether a special-use permit is required) lands in `general` because the `permit` description talks about reserving and paying. Widen the description by a clause and re-run. Two rules constrain any edit: the ordering paragraph that makes emergency win stays, and `unsure` stays narrow.

```python
- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, and questions about whether an activity requires a
  permit at all, including billing problems and missing confirmations.
```

Run:

```bash
uv run main.py
```

Check: Lab step 3, the success check: both emergencies classified `emergency`, `inq-0035` in `unsure`, and your accuracy at or above where it started. Judge every taxonomy edit on emergency recall first. Stretch: add a `priority` field to the result type, or a confidence score routed to `unsure` below a threshold.
