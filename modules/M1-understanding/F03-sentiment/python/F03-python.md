<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 03: Sentiment (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F03-dotnet.md), [TypeScript](../typescript/F03-typescript.md). Lab overview: [F03-lab.md](../F03-lab.md).*

**The User Problem:** Trailhead Guides sells gear, and the Cascade 65 backpack has 300 reviews. The product team asks a simple question: are people happy with it, and what are they mad about? Star ratings lie: "4 stars, but the hip belt broke on day two" is not a happy customer. Someone would have to read all 300, and every new product adds to the pile. The team doesn't need eloquent analysis, just a reliable happy/unhappy/mixed signal at scale.

The user in this feature is the product team, not the hiker, and that's deliberate. Stakeholders are users too. Their version of the problem is that nobody has time to read every review, so a defect surfaces only when returns spike. Track the same happy/unhappy signal weekly and it surfaces months earlier, as a chart sliding downhill. One product in this corpus has exactly that kind of problem buried in its reviews, and the demo gets to find it.

*A Challenge lab. Do it if you finished [Module 1](../../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20. All three are hand-picked from feature 06's `data/gear-reviews.jsonl`; no script builds them.
- **How:** send one prompt per review through your track's chat client. Keep the one-word label that comes back. Compare it with the hand label. Same prompt bytes everywhere, temperature 0.
- **Model:** `phi3` is the small model. The big model is `gpt-4.1` on Azure. Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` to use it. With no key, `llama3.2` on Ollama fills in for the big model, and the whole lab runs offline.

## The Concept

Sentiment analysis is classification applied to text, and it's this workshop's vehicle for the most useful model-selection lesson of the day: you don't always need the big model. A small local model (`phi3`, about 2GB, free, private) labels straightforward reviews just as well as a frontier cloud model. At 300 reviews per product across a whole catalog, per-token pricing versus free-on-your-hardware is a real budget line.

The comparison cuts both ways, though. Feed both models the corpus's hard cases (sarcasm like "Great bag, if you enjoy shoulder pain", mixed feelings, ratings that contradict the text) and accuracy drops for everyone. How far it drops for each model is the thing you measure, and there is no verdict here for either side. The decision is measurable: run both on a labeled sample, count the disagreements, look at what the errors cost you, then choose. Most teams never run that experiment; you'll run it before lunch.

Each step below adds one thing to the program. The `starter/` already classifies one review on `phi3`. Edit it until it does all six steps. Look at `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Swapping providers later is a different constructor and nothing else.

- `starter/main.py`: one client, one `classify` function, one review (`gr-0007`, the sarcastic two-star). Prints the review and what `phi3` says.
- `complete/main.py`: the finished demo. Both review sets through both models with the byte-identical four-line prompt, a table with disagreements flagged, accuracy per set, and the disagreement list with a verdict on who was right.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [SETUP.md](../../../../SETUP.md)) is the one install for all ten features. Run the starter from `starter/`. From `complete/`:

```bash
uv run main.py            # both sets, both models
uv run main.py --easy     # easy set only
uv run main.py --hard     # hard set only
```

### Step 0: Run the starter on the sarcastic review

**Do:** run `starter/` as it is. It loads `gr-0007` from `data/hard.jsonl`, prints the product, star rating, reviewer, and text, then sends this prompt to `phi3` and prints the label:

```text
Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: Absolutely love it when the mesh blew out at the pinky toe inside two weeks the second day of a trip. Five-star experience, truly, if the stars are measuring my personal growth through adversity. Rating it what it deserves.
```

The starter already does this. It picks the id from the command line (or `gr-0007`), reads both data files into one list of dicts, finds the matching review, prints it, and classifies it:

```python
wanted = sys.argv[1] if len(sys.argv) > 1 else "gr-0007"
reviews = [json.loads(line) for name in ("easy.jsonl", "hard.jsonl") for line in (DATA / name).read_text().splitlines() if line.strip()]
review = next(r for r in reviews if r["id"] == wanted)

print(f"{review['product']} ({review['rating']} stars), reviewed by {review['reviewer']}")
print(review["text"])
print()
```

```python
print(f"phi3 says: {classify(client, 'phi3', review['text'])}")
```

`sys.argv[1]` is the first word after `main.py` on the command line. Each line of a `.jsonl` file is one JSON object, and `json.loads` turns it into a dict, so `review["text"]` is the review text. Run it from the `starter/` folder:

```bash
uv run main.py
```

**Why:** the line breaks in that prompt are load-bearing. Reflowing the same words onto one line, changing nothing but the newlines, drops `phi3` from 9/10 to 7/10 on the easy set and from 7/10 to 4/10 on the hard set. Keep it wrapped exactly as it is here.

The prompt lives inside the starter's `classify` function. The starter already does this; the `f` before the triple quotes makes `{text}` the place where the review goes:

```python
    prompt = f"""Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: {text}"""
```

**Check:** `phi3 says: negative`. The review has two stars and says "five-star experience, truly". That is sarcasm, and the small model gets it right. Pass a different id (`gr-0034`, for instance) to classify a different review.

```bash
uv run main.py gr-0034
```

### Step 1: Load the reviews and the reference labels

**Do:**
1. Review how the existing code from the starter project opens `../../data/easy.jsonl` (relative to `starter/`): 10 reviews where text and stars agree, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`).

   `DATA` is a `Path`, and `/` joins folder and file names, so `DATA / "easy.jsonl"` is the easy file:

   ```python
   DATA = Path(__file__).resolve().parents[2] / "data"
   ```

2. Read it one line at a time, skip blanks, parse each line into a list.

   The starter's `reviews = ...` line already does this for both files at once. For the easy file alone, put this below the `DATA = ...` line. It is a list comprehension: `json.loads(line)` for every line that is not blank:

   ```python
   # Hint: one dict per non-blank line of the easy file
   easy = [json.loads(line) for line in (DATA / "easy.jsonl").read_text().splitlines() if line.strip()]
   ```

   `complete/` skips the list: its loop in step 2 reads the file line by line directly, so you can delete `easy` once the check below passes.

3. Open `../../data/reference-labels.json`: one object keyed by review `id`, each value a `set` (`easy`/`hard`), a `label`, and on hard cases a `rationale`.
4. Parse it into a dictionary keyed by `id`. You only need `label` in code.

`DATA` already points at `../../data/`. Put this below your `easy = ...` line (it is the same in `complete/main.py`). `json.loads` turns the whole file into a dict of dicts, so `labels["gr-0002"]["label"]` gives you one label:

```python
labels = json.loads((DATA / "reference-labels.json").read_text())
```

To see the check below, print the counts once and then delete the line:

```python
# Hint: print what you loaded
print(len(easy), easy[0]["id"], len(labels), labels["gr-0002"]["label"])
```

**Why:** `gr-0004`'s label changed from `positive` to `mixed`. `gpt-4.1` called it `mixed` on every soak-test run, and a reread agreed that a two-star review which praises the product is split. `expected-output.md` tells that story.

**Check:** the easy list has 10 reviews, first `id` is `gr-0002`. The label dictionary has 20 keys; `gr-0002` maps to `negative`.

### Step 2: Classify the easy set on `phi3` and score it

**Do:**
1. Review how the existing code from the starter project classifies a review: one review text in, one prompt out, `phi3`, temperature 0.

   `client` and `model` are passed in, and `temperature=0` is set on every call:

   ```python
   def classify(client: OpenAI, model: str, text: str) -> str:
   ```

   ```python
       response = client.chat.completions.create(model=model, messages=[{"role": "user", "content": prompt}], temperature=0)
   ```

2. Review how it lowercases the reply, looks for `positive`/`negative`/`mixed`, and keeps whichever shows up first (or the trimmed reply if none do).

   It happens at the end of `classify`. `found` is a list of `(position, label)` pairs for the labels that appear in the reply, and `min` picks the pair with the smallest position:

   ```python
       raw = (response.choices[0].message.content or "").lower()
       # Small models sometimes wrap the label in a sentence; keep the first label mentioned.
       found = [(raw.index(l), l) for l in ("positive", "negative", "mixed") if l in raw]
       return min(found)[1] if found else raw.strip()
   ```

3. Loop over the easy list, classify each review's `text`, look up the reference label by `id`.
4. Print one row per review: `id`, reference label, `phi3` label.
5. Count matches; print `phi3 N/10`.

`client` is the starter's Ollama client. Delete the starter's single-review lines (from `wanted = ...` down to the first `print()`, and the last line, `print(f"phi3 says: ...")`) and your step 1 lines (`easy = ...`, `labels = ...`, and the print). Keep `classify`, and put this at the bottom of the file, below it, so the function exists before the loop calls it. It covers steps 1 and 2 together. `:<9` pads a value to 9 characters so the columns line up, and `correct += label == reference` adds 1 when they match (`True` counts as 1):

```python
labels = json.loads((DATA / "reference-labels.json").read_text())
correct = total = 0
for line in (DATA / "easy.jsonl").read_text().splitlines():
    if not line.strip():
        continue
    review = json.loads(line)
    reference = labels[review["id"]]["label"]
    label = classify(client, "phi3", review["text"])
    print(f"{review['id']:<9} {reference:<10} {label:<10}")
    total += 1
    correct += label == reference
print(f"phi3 {correct}/{total}")
```

**Why:** the prompt asks for one word, but a small model sometimes wraps the label in a sentence. Lowercasing the reply and keeping the first of `positive`, `negative`, or `mixed` that appears scores those replies as the label they contain, so the number you get measures the model's judgment and not its phrasing.

**Check:** `gr-0003` comes back `positive`, `gr-0002` `negative`. Recorded `phi3` score 9/10, missing `gr-0074`. A 7/10 with every miss `mixed` means you reflowed the prompt onto one line.

### Step 3: Add the big model and classify every review twice

**Do:**
1. Read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT`.

   The endpoint is `https://trailhead-ai-workshop.openai.azure.com`, the deployment is `gpt-4.1`, and the key is handed out in the room.

   Set them in the terminal you run `uv run` from (they last until you close it; skip this to use the `llama3.2` fallback):

   ```bash
   export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
   export AZURE_OPENAI_KEY=<KEY FROM INSTRUCTOR>
   export AZURE_OPENAI_DEPLOYMENT=gpt-4.1
   ```

   In Python, `os.environ.get("NAME")` reads one of them and returns `None` when it is not set. The code under item 2 does the reading.

2. Build a second chat client from those three variables, falling back to `llama3.2` when any is missing; name it `azure:<deployment>` or `llama3.2` for printing.

   The starter imports only `OpenAI`. The Azure client is a second class in the same `openai` package (already in the root `pyproject.toml`, nothing to install), so change the import and add `os` and `dataclass`. At the top of `main.py`, put `import os` below `import json`, put `from dataclasses import dataclass` below `import sys`, and replace `from openai import OpenAI` with the last line:

   ```python
   import os
   from dataclasses import dataclass

   from openai import AzureOpenAI, OpenAI
   ```

   Rename the starter's `client` to `ollama`, then build the second client from the three env vars. The block below replaces the starter's `client = OpenAI(...)` line, so your step 2 call `classify(client, "phi3", review["text"])` no longer works; item 5 replaces it. `small` and `big` are each a pair (a tuple) of client and model name. `AzureOpenAI` takes `azure_endpoint`, `api_key`, and `api_version` (a REST API date, not a model version); the deployment name is what you pass as `model` on each call, which is why the pair below carries it. This also covers item 3:

   ```python
   ollama = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
   small = (ollama, "phi3")

   endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
   key = os.environ.get("AZURE_OPENAI_KEY")
   deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
   if endpoint and key and deployment:
       big = (AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment)
       big_name = f"azure:{deployment}"
   else:
       print("AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.\n")
       big = (ollama, "llama3.2")
       big_name = "llama3.2"
   ```

3. When falling back, print `AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.`
4. Change nothing in the body of classify. The second model is just a different client; some tracks pass that client in as a parameter, so its signature line may change.

   `classify(*small, text)` unpacks the pair into the starter's `(client, model, text)` signature; `complete/` instead changes the signature to take the pair, and the body is identical either way.

5. In the step 2 loop, classify each review with both `phi3` and the big model; store a record with the review, set name, reference label, small label, big label.
6. Add a fourth column for the big model's label; append `  <- disagree` when the two differ.

   Add a record type below `classify` and above the loop. `@dataclass` writes the constructor for you, so `Result(review, "easy", reference, s, b)` fills the five fields in order and `r.small` reads one back:

   ```python
   @dataclass
   class Result:
       review: dict
       set: str
       reference: str
       small: str
       big: str
   ```

   Then grow the step 2 loop to a four-column table. Replace the step 2 `correct = total = 0` line with these two lines:

   ```python
   results: list[Result] = []
   print(f"{'id':<9} {'reference':<10} {'phi3':<10} {big_name:<10}")
   ```

   Inside the loop, replace the four lines from `label = classify(...)` down to `correct += label == reference` with the five lines below (indented 4 spaces, like `reference = ...`). Delete the `print(f"phi3 {correct}/{total}")` line after the loop; step 4 brings scoring back:

   ```python
   s = classify(*small, review["text"])
   b = classify(*big, review["text"])
   results.append(Result(review, "easy", reference, s, b))
   flag = "  <- disagree" if s != b else ""
   print(f"{review['id']:<9} {reference:<10} {s:<10} {b:<10}{flag}")
   ```

**Why:** the swap is one extra client, built from `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT`, plus one extra `classify` call per review. The prompt, the body of `classify`, and the scoring all stay as they were, so whatever differs between the two label columns came from the model.

**Check:** the easy table has four columns, and `gr-0074` is flagged `<- disagree`: reference `positive`, `phi3` says `mixed`, and the big-model column (headed `azure:gpt-4.1`) says `positive`. On the `llama3.2` stand-in the row is still flagged, big label `negative`.

### Step 4: Run the hard set too and print accuracy per set

**Do:**
1. Put the step 3 loop inside an outer loop over `easy` then `hard`, reading `../../data/{set}.jsonl` each time, with a heading line per set.
2. After both tables, loop over the two set names again; count matches for small and big labels against the reference.
3. Print one line per set: name, `phi3 N/10`, big model's name with `N/10`.

Use `name` instead of the literal `"easy"` in the `Result`. Replace everything from `results: list[Result] = []` to the end of the file with the block below: `results` moves above the outer loop, your step 3 loop moves inside it (one indent deeper), and `f"{name}.jsonl"` picks the file per set. The accuracy loop goes right after it. `[r for r in results if r.set == name]` keeps only that set's records, and `sum(1 for ...)` counts the ones that match:

```python
# Hint: your step 3 loop, one indent deeper inside a loop over the set names
sets = ["easy", "hard"]

results: list[Result] = []

for name in sets:
    print(f"── {name} set ──")
    print(f"{'id':<9} {'reference':<10} {'phi3':<10} {big_name:<10}")
    for line in (DATA / f"{name}.jsonl").read_text().splitlines():
        if not line.strip():
            continue
        review = json.loads(line)
        reference = labels[review["id"]]["label"]
        s = classify(*small, review["text"])
        b = classify(*big, review["text"])
        results.append(Result(review, name, reference, s, b))
        flag = "  <- disagree" if s != b else ""
        print(f"{review['id']:<9} {reference:<10} {s:<10} {b:<10}{flag}")
    print()

print("── accuracy vs. reference labels ──")
for name in sets:
    batch = [r for r in results if r.set == name]
    small_ok = sum(1 for r in batch if r.small == r.reference)
    big_ok = sum(1 for r in batch if r.big == r.reference)
    print(f"{name:<5}  phi3 {small_ok}/{len(batch)}   {big_name} {big_ok}/{len(batch)}")
print()
```

**Check:** `gr-0034` comes back `negative` from both models. On `gr-0013`, `phi3` says `mixed`, big model says `negative` (the reference label). Your numbers will not be mine. When I ran it: `phi3` 9/10 easy and 7/10 hard, `gpt-4.1` 10/10 on both, `llama3.2` stand-in 9/10 and 8/10. What has to hold is the shape, the big model ahead of `phi3` on the hard set, not the digits.

### Step 5: Print the disagreement list

**Do:**
1. Take only records where small and big labels differ.
2. Print a heading: `disagreements (N of 20)`.
3. For each: verdict is the big model's name + ` right` if it matches the reference, `phi3 right` if the small label matches instead, else `both wrong`.
4. Print one line per disagreement: `id`, set in brackets, `ref=`, `phi3=`, big model `=`, verdict in parens; then on its own indented line the review text in quotes, cut to 100 characters with `...` if longer.
5. If empty, print `(none this run)`.

Put these lines at the very end of the file, below the accuracy loop's final `print()`. The `verdict` line is a chained conditional expression: it reads left to right like an if / elif / else. `text[:100]` is the first 100 characters:

```python
disagreements = [r for r in results if r.small != r.big]
print(f"── disagreements ({len(disagreements)} of {len(results)}) ──")
for d in disagreements:
    verdict = f"{big_name} right" if d.big == d.reference else "phi3 right" if d.small == d.reference else "both wrong"
    print(f"{d.review['id']} [{d.set}] ref={d.reference} phi3={d.small} {big_name}={d.big}  ({verdict})")
    text = d.review["text"]
    print(f'  "{text if len(text) <= 100 else text[:100].rstrip() + "..."}"')
if not disagreements:
    print("(none this run)")
```

**Check:** your version of the two tables in `expected-output.md`: recorded 4 of 20 against `gpt-4.1` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0021`) and 4 of 20 against `llama3.2` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0089`). Different numbers are fine; no tables is not.

### Stretch goals

Pick any. The first two are already built in `complete/`.

- **Add `--easy` and `--hard` flags.** If `--easy` is present, run only the easy set; if `--hard`, only the hard set; otherwise both. **Check:** `--hard` prints one table, one accuracy line, and a disagreement count out of 10 instead of 20.

  Replace your step 4 `sets = ["easy", "hard"]` line with the line from `complete/main.py`. `"--easy" in sys.argv` is `True` when that word was passed:

  ```python
  sets = ["easy"] if "--easy" in sys.argv else ["hard"] if "--hard" in sys.argv else ["easy", "hard"]
  ```

  ```bash
  uv run main.py --hard
  ```

- **Reflow the prompt onto one line and measure the damage.** Replace the four line breaks with spaces, rerun both sets on both models, then put the line breaks back. **Check:** `phi3` drops from 9/10 to 7/10 easy and from 7/10 to somewhere between 4/10 and 6/10 hard (4/10 in the recorded run, 6/10 in two later runs), every miss `mixed`. `llama3.2` scores the same either way. The small model is the one that cares about prompt shape.

  The change is inside `classify`. Join the four instruction lines into one (spaces where the line breaks were) and leave the blank line and the `Review:` line alone:

  ```python
  # Hint: same words, one line; fill in the middle sentences
      prompt = f"""Classify this gear review as exactly one word: positive, negative, or mixed. Positive means ... Reply with only the label.

  Review: {text}"""
  ```

- **Aspect-based sentiment.** Ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` instead of one word, and tell the model to reply with that JSON and nothing else. Parse the reply, and catch the parse error so one bad reply doesn't stop the run. **Check:** the run finishes without crashing, and the replies that parse leave aspects `null` when the review never mentions them. Expect many `phi3` replies not to parse: in one measured run only 7 of 20 did, and the rest added an explanation after the JSON or wrapped it in a ```` ```json ```` fence. Those are the parse errors you catch, and they are why production code uses a structured-output schema like step 1 of feature 02. A `price` sentiment on a review that never mentions price is the failure to look for.

  `complete/` does not build this one. Write a second function below `classify` with a prompt that asks for that JSON. `complete/main.py` makes no `format` or `response_format` call, so this hint parses the reply text with `json.loads` instead. In an f-string, `{{` and `}}` print a literal `{` and `}`, so only `{text}` is filled in. JSON `null` becomes Python `None`:

  ```python
  # Hint: a second classify that returns a dict parsed from the reply
  def classify_aspects(client: OpenAI, model: str, text: str) -> dict:
      prompt = f"""Classify this gear review. Reply with only JSON shaped like
  {{"overall": "positive|negative|mixed", "aspects": {{"comfort": ..., "durability": ..., "price": ...}}}}
  Use null for any aspect the review never mentions.

  Review: {text}"""
      response = client.chat.completions.create(model=model, messages=[{"role": "user", "content": prompt}], temperature=0)
      return json.loads(response.choices[0].message.content or "")
  ```

  If `json.loads` raises `json.JSONDecodeError`, the reply was not pure JSON: a small model that adds a sentence or a second object after the JSON is the "parseable JSON every time" check failing. `phi3` does this on some reviews, so catch the error and print the reply instead of letting it stop the run (`e.doc` is the text `json.loads` was given, and `repr` keeps its line breaks on one line):

  ```python
  # Hint: call it inside your loop, after the two classify calls (indent to match them)
  try:
      parsed = classify_aspects(*small, review["text"])
      print(review["id"], parsed["overall"], "price =", parsed["aspects"]["price"])
  except json.JSONDecodeError as e:
      print(review["id"], "not JSON:", repr(e.doc[:80]))
  ```

## What Is in This Folder

- `data/easy.jsonl`: 10 straightforward reviews chosen by hand from the full review set (feature 06's `data/gear-reviews.jsonl`). The text says what it means and the star rating agrees.
- `data/hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `data/reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case. One label (`gr-0004`) was revised after the soak test; `expected-output.md` records why.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The reviews keep their original ids, so any of them can be traced back to the full corpus.
