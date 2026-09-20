<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 07: Classification & Routing (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F07-dotnet.md), [TypeScript](../typescript/F07-typescript.md). Lab overview: [F07-lab.md](../F07-lab.md).*

**The User Problem:** Every message to Trailhead Guides lands in one inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, and, occasionally, someone reporting an actual emergency. A ranger triages the pile by hand once or twice a day. The permit request waits behind the granola questions, the complaint goes to the wrong person twice, and the emergency sits unread for four hours. The users' real problem is that their message goes into a hole, and how fast it comes out depends on luck.

*This is the Recommended lab for [Module 3](../../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** sort visitor messages into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and send each one to the right queue.
- **Input:** `data/inquiries-slice.jsonl` (20 messages: `id`, `channel`, `received`, `text`; emergencies `inq-0013` and `inq-0041`, ambiguous `inq-0035`), `data/reference-labels.json` (the correct category for each id, and the queue for each category), `data/inquiries.jsonl` (the full 100-message inbox the 20 came from; the demo scrolls it, the lab does not read it).
- **How:** send each message to `llama3.2` with the taxonomy prompt. Use a JSON schema so the model can only answer with one of the seven labels. Loop all 20 at temperature 0. Print the routing table, then score the results against the reference labels.
- **Model:** `llama3.2`, local. No key.

## The Concept

This is classification again (feature 03 was the warm-up), but now the label has consequences: it decides where the message goes and how fast. An LLM makes a solid zero-shot classifier. You describe the categories in plain language and it labels messages with no training data, which is exactly the situation most teams are in on day one.

Two design decisions carry the feature. The first is that the taxonomy is the product: category names and one-sentence descriptions in the prompt are where accuracy lives, and when the model misfiles something, you usually fix the description rather than the model. The second is that errors are not symmetric, since misrouting a complaint costs a day of annoyance and misrouting an emergency is a headline. So the system needs an "unsure" route to a human, and it should be tuned so the expensive class never slips through, even at the cost of extra false alarms. Recall on the class that matters, not overall accuracy, is the number to watch. A local model does this fine, and at inbox volume, free matters.

Every step below is one thing to make the program do. The starter classifies one message and prints whatever text the model sends back. Edit it until it does all six steps. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Swapping the provider later is a different constructor and nothing else. `starter/main.py` classifies a single inquiry and prints the free-text label. `complete/main.py` classifies the whole slice through structured output into a Python `Enum`, prints emergencies first, and scores accuracy and emergency recall.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [SETUP.md](../../../../SETUP.md)) is the one install for all ten features. `uv run` finds it from any folder, so there is no venv to activate. From `complete/` or `starter/` (`starter/main.py` takes no flags, at most the one positional id its header comment names):

```bash
uv run main.py             # all 20, scored
uv run main.py inq-0013    # (starter) one inquiry
```

`complete/` takes no arguments: keep the starter's positional id through step 2 and drop it when step 3 loops the slice.

### Step 0: Run the starter and read the free-text label

**Do:** run `starter/` as it is. It reads `inq-0005` from the slice, builds the taxonomy prompt, puts the message after `Message:`, and sends it to `llama3.2` as one user message. Then run it again with `inq-0035` as the argument, and again with `inq-0013`.

From `starter/`:

```bash
uv run main.py
uv run main.py inq-0035
uv run main.py inq-0013
```

The starter already does this. It picks the id from the command line (`sys.argv[1]`, or `inq-0005` when you give none), reads the slice, and keeps the line whose `id` matches:

```python
wanted = sys.argv[1] if len(sys.argv) > 1 else "inq-0005"
inquiry = next(json.loads(l) for l in (DATA / "inquiries-slice.jsonl").read_text().splitlines() if l.strip() and json.loads(l)["id"] == wanted)
```

The starter already does this too. The prompt is one big `f"""` string that ends with the message text, and the call sends it as one user message and prints whatever text comes back:

```python
Message:
{inquiry['text']}"""

response = client.chat.completions.create(model="llama3.2", messages=[{"role": "user", "content": prompt}])
print(f"{inquiry['id']}: {response.choices[0].message.content}")
```

**Check:** one label per run, printed after the id. `inq-0013` says `emergency`. `inq-0035` usually says `conditions`, and some runs say `unsure`, `emergency`, or `complaint`: measured over 30 runs across the three tracks, 24 `conditions`, 4 `unsure`, 1 `emergency`, 1 `complaint`. That spread on one message is the problem. Nothing in the starter stops a run from printing a label that is not one of the seven names, such as `Emergency.` or a sentence, either. The rest of the lab closes both gaps.

### Step 1: Load the 20 inquiries and the reference labels

**Do:**
1. Open `../../data/inquiries-slice.jsonl`. Every line is one JSON object with `id`, `channel`, `received`, `text`. Read it line by line, skip blanks, parse each line, keep the results in a list. The starter already resolves that `data/` folder into a constant; reuse it for both files.

   The starter resolves the data folder as `DATA` (a `Path`), and `DATA / "inquiries-slice.jsonl"` is the file inside it. This one line reads the file, splits it into lines, skips blank ones (`if l.strip()`), and parses each with `json.loads`, giving a list of dictionaries. Put it right below the `DATA = ...` line. Leave the starter's `wanted`/`inquiry` lines in place for now; step 3 removes them.

   ```python
   inquiries = [json.loads(l) for l in (DATA / "inquiries-slice.jsonl").read_text().splitlines() if l.strip()]
   ```
2. Open `../../data/reference-labels.json` and parse it: an object with `routing` (category to queue name), `labels` (id to correct category), and `notes` (why `inq-0013`/`inq-0041` are emergencies and why `inq-0035` is `unsure`). Keep the `routing` and `labels` dictionaries.

   Put this directly below the `inquiries` line. `reference["routing"]` and `reference["labels"]` are the two dictionaries; there is nothing else to unpack.

   ```python
   reference = json.loads((DATA / "reference-labels.json").read_text())
   ```

   To see the Check numbers, print them once below the two load lines (delete the line afterward):

   ```python
   # Hint: a throwaway print for the Check
   print(len(inquiries), inquiries[0]["id"], len(reference["routing"]), len(reference["labels"]))
   ```

**Why:** these 20 are drawn from the fictional Trailhead Guides 100-message inbox, chosen to mirror the inbox's mix, including both emergencies and the one ambiguous message. `unsure` was added to the taxonomy, labels, and routing table together once the ambiguous message needed a queue.

**Check:** the inquiry list has 20 entries. The first `id` is `inq-0001`. `routing` has 7 keys and `labels` has 20.

### Step 2: Pin the answer to the seven labels with structured output

**Do:**
1. Delete the line `Answer with the category name only.` from the starter's prompt. The schema below does that job now.

   In `main.py` it is this line near the end of the `prompt = f"""` string, just above `Message:`. Delete it and the blank line under it:

   ```python
   Answer with the category name only.
   ```
2. Define the category as an enum type with exactly seven allowed values, wrapped in a result type with one field, `category`. If your language cannot spell `lost-and-found` as an identifier, map that member to the exact JSON name. The generated schema must list `lost-and-found`, or three of the 20 messages can never match their reference label.

   A Python identifier cannot contain a hyphen, so `lost-and-found` is the member's value on a `str` Enum, and the schema pydantic generates lists the values, not the member names. Two imports the starter does not have: `from enum import Enum` and `from pydantic import BaseModel` (pydantic ships with the `openai` package).

   Add the two imports at the top of `main.py`. `from enum import Enum` goes below `import json`, and `from pydantic import BaseModel` goes below `from openai import OpenAI`:

   ```python
   from enum import Enum
   ```

   ```python
   from pydantic import BaseModel
   ```

   Then paste the two classes below the `DATA = ...` line, above the step 1 load lines. Python runs the file top to bottom, so a class has to be defined above any line that uses it. In `class Category(str, Enum)`, each line is `member_name = "value"`: the name on the left is what your code writes (`Category.lost_and_found`), and the string on the right is what the model sends back and what the schema lists. `class TriageResult(BaseModel)` with the one line `category: Category` declares an object with one field that must be one of those seven values:

   ```python
   class Category(str, Enum):
       permit = "permit"
       conditions = "conditions"
       complaint = "complaint"
       lost_and_found = "lost-and-found"
       emergency = "emergency"
       general = "general"
       unsure = "unsure"


   class TriageResult(BaseModel):
       category: Category
   ```

3. Send the message through your chat client with that result type as the structured-output schema, temperature 0.

   The typed call is `client.chat.completions.parse` with `response_format=TriageResult`, `temperature=0` is a keyword argument on that same call, and the parsed model is `response.choices[0].message.parsed`. Keep the starter's single-id lookup and swap the call. `model_dump_json()` prints the wire form the step 2 Check describes (pydantic writes it without the space, `{"category":"conditions"}`).

   Replace the starter's last two lines (`response = client.chat.completions.create(...)` and the `print` under it) with this. `result.category` is a `Category` member, and `.value` is its string, such as `lost-and-found`:

   ```python
   response = client.chat.completions.parse(
       model="llama3.2",
       messages=[{"role": "user", "content": prompt}],
       response_format=TriageResult,
       temperature=0,
   )
   result = response.choices[0].message.parsed
   print(f"{inquiry['id']}: {result.category.value}")
   print(result.model_dump_json())
   ```

4. Keep the prompt text:

```text
You are the triage system for the Trailhead Guides shared inbox. Classify the visitor message into exactly one category.

- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, including billing problems and missing confirmations for a permit application.
- conditions: asking whether a trail, road, or area is open, safe, or passable right now: snow, water levels, washouts, wildlife activity, closures.
- complaint: unhappy about a park facility, service, or staff member and wants it acknowledged or fixed.
- lost-and-found: reporting a lost or found physical item.
- emergency: a person may be hurt, missing, or in danger right now and needs immediate human attention.
- general: anything else: park rules, fees, trip planning, questions that fit none of the above.
- unsure: two different queues both have to act before this message can be resolved, so no single queue owns it. The case that qualifies: the sender asks about trail conditions AND asks someone to change, refund, or cancel a booking. Trail info cannot issue a refund, and the permits office does not decide whether a trail is passable, so a human reads this queue and splits the work. Also use unsure when the message fits none of the categories above.

Decide in this order. First, if anyone might be hurt, missing, or in danger, answer emergency and stop; never answer unsure for those, even when the message also mentions permits, conditions, or a lost item. Second, if one queue can resolve the whole message on its own, answer that queue; a booking or reservation problem with nothing else attached is permit, not unsure. Third, only if two queues must both act, answer unsure. Unsure is not a catch-all for anything hard.

Message:
```

The starter already does this: its `prompt = f"""` string holds that exact text, and it ends by putting the message after `Message:`. After item 1 the only change is the deleted line, so the end of the string still reads:

```python
Message:
{inquiry['text']}"""
```

This is the schema your client generates from that type:

```json
{
  "type": "object",
  "properties": {
    "category": {
      "type": "string",
      "enum": ["permit", "conditions", "complaint", "lost-and-found", "emergency", "general", "unsure"]
    }
  },
  "required": ["category"]
}
```

5. Run the program on `inq-0005`, then `inq-0041`, then `inq-0035`. The message text for `inq-0041`:

```text
My dad slipped on scree on the Beehive descent, we're just past the ladder section. His ankle is swollen bad, can't put weight on it. We have water and I have 2 bars of signal. 2 adults 1 teen. How do we get help up here? Submitting this because 911 kept dropping.
```

And for `inq-0035`:

```text
Hi, I have a backcountry permit that includes a night at the Avalanche Lake area on June 24 (conf #GL-2026-07733). With the bridge out, is my itinerary even doable, and if not, will you let me swap that night for a different site without penalty, or refund it? I need to know before we leave Thursday. Thanks, Priya
```

From `starter/`:

```bash
uv run main.py inq-0005
uv run main.py inq-0041
uv run main.py inq-0035
```

**Why:** `inq-0035` now lands in `unsure`, while step 0's free-text version called it `conditions`. The enum made `unsure` a real choice for the model.

**Check:** the parsed result's `category` is `conditions` for `inq-0005`, `emergency` for `inq-0041`, and `unsure` for `inq-0035`. Serialize the result back to JSON to see the wire form, `{"category": "conditions"}`. The value is always one of the seven strings.

### Step 3: Classify all 20 in a loop

**Do:**
1. Replace the starter's single-id lookup with a loop over the step 1 list.

   Delete the starter's single-id lookup. These are the two lines to remove (the `inquiries` list from step 1 replaces them, so the program no longer takes an id argument). Also delete the `import sys` line at the top, which only they used:

   ```python
   wanted = sys.argv[1] if len(sys.argv) > 1 else "inq-0005"
   inquiry = next(json.loads(l) for l in (DATA / "inquiries-slice.jsonl").read_text().splitlines() if l.strip() and json.loads(l)["id"] == wanted)
   ```

   Also delete step 2's single `parse` call and its `result = ...` line and two prints; the loop in item 3 replaces them.
2. For each inquiry, build the step 2 prompt with that inquiry's `text` after `Message:`, send it with the same schema and temperature 0.

   Turn the starter's `prompt = f"""...` string into a function `prompt(text)` that puts `text` after `Message:` (as `complete/` does). Delete the whole `prompt = f"""` block and paste this in its place (a function only has to be defined above the loop that calls it). Only the `def` and `return` lines are indented; the prompt text itself stays at the left edge, because every space inside the `"""` string would be sent to the model. The message is now `{text}` instead of `{inquiry['text']}`, because inside the function there is no `inquiry` variable:

   ```python
   def prompt(text: str) -> str:
       return f"""You are the triage system for the Trailhead Guides shared inbox.
   Classify the visitor message into exactly one category.

   - permit: reserving, changing, canceling, or paying for a permit, pass,
     or reservation, including billing problems and missing confirmations
     for a permit application.
   - conditions: asking whether a trail, road, or area is open, safe, or
     passable right now: snow, water levels, washouts, wildlife activity,
     closures.
   - complaint: unhappy about a park facility, service, or staff member
     and wants it acknowledged or fixed.
   - lost-and-found: reporting a lost or found physical item.
   - emergency: a person may be hurt, missing, or in danger right now and
     needs immediate human attention.
   - general: anything else: park rules, fees, trip planning, questions
     that fit none of the above.
   - unsure: two different queues both have to act before this message can
     be resolved, so no single queue owns it. The case that qualifies: the
     sender asks about trail conditions AND asks someone to change, refund,
     or cancel a booking. Trail info cannot issue a refund, and the permits
     office does not decide whether a trail is passable, so a human reads
     this queue and splits the work. Also use unsure when the message fits
     none of the categories above.

   Decide in this order. First, if anyone might be hurt, missing, or in
   danger, answer emergency and stop; never answer unsure for those, even
   when the message also mentions permits, conditions, or a lost item.
   Second, if one queue can resolve the whole message on its own, answer
   that queue; a booking or reservation problem with nothing else attached
   is permit, not unsure. Third, only if two queues must both act, answer
   unsure. Unsure is not a catch-all for anything hard.

   Message:
   {text}"""
   ```

   The loop in item 3 uses `MODEL`. Put this constant right below the `client = OpenAI(...)` line:

   ```python
   MODEL = "llama3.2"
   ```
3. Read `category` from the response, store the (inquiry, category) pair in a results list.

   Put the loop at the end of the file, below the step 1 load lines. `results: list[tuple[dict, Category]] = []` makes an empty list; the part after the colon is only a type label saying each entry is a pair of (inquiry dictionary, `Category`). `results.append((inquiry, ...))` adds one pair; note the double parentheses, since the pair itself is the one argument:

   ```python
   results: list[tuple[dict, Category]] = []
   for inquiry in inquiries:
       response = client.chat.completions.parse(
           model=MODEL,
           messages=[{"role": "user", "content": prompt(inquiry["text"])}],
           response_format=TriageResult,
           temperature=0,
       )
       results.append((inquiry, response.choices[0].message.parsed.category))
       print(".", end="", flush=True)
   print("\n")
   ```
4. Print a `.` after each call to show progress, then a blank line after the loop.

   The item 3 loop already does this. `print(".", end="", flush=True)` prints a dot with no line break (`end=""`) and shows it right away (`flush=True`) instead of waiting for the line to finish. `print("\n")` after the loop, not indented, prints a newline plus its own line end, which gives the blank line.

**Check:** 20 dots, then 20 stored pairs, every category one of the seven strings. The run takes under a minute.

### Step 4: Print emergencies first, then the routing table

**Do:**
1. Filter results to `category == "emergency"`.

   Put this below the loop's `print("\n")`. It is a list comprehension: it walks `results`, splits each pair into `i` (the inquiry) and `c` (the category), and keeps the pair only when `c == Category.emergency`. Comparing against the Enum member directly is fine:

   ```python
   emergencies = [(i, c) for i, c in results if c == Category.emergency]
   ```
2. If any, print `!!! EMERGENCY: route to dispatch, page the duty ranger now !!!`, then one line per emergency with `!!! `, the id, and the first 70 characters of the text. Blank line after.

   `if emergencies:` is true when the list is not empty. In `for inquiry, _ in emergencies`, the `_` is a name for the category you do not use. Put this directly below the `emergencies` line:

   ```python
   if emergencies:
       print("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!")
       for inquiry, _ in emergencies:
           print(f"!!! {inquiry['id']}  {clip(inquiry['text'], 70)}")
       print()
   ```

   `clip(text, n)` is a two-line helper in `complete/` that cuts the text at `n` characters and appends `...`. Paste it below the `prompt` function, above the load lines:

   ```python
   def clip(text: str, n: int) -> str:
       return text if len(text) <= n else text[:n] + "..."
   ```
3. Print a header line (`id`, `category`, `routed to`) and a rule of 62 dashes.

   Directly below the emergency block. Inside an f-string, `{'id':<10}` prints the text `id` padded with spaces to 10 characters, left-aligned (`<`). `"-" * 62` repeats the dash 62 times:

   ```python
   print(f"{'id':<10} {'category':<15} routed to")
   print("-" * 62)
   ```
4. Sort so emergencies come first. For each pair, print the id padded to 10 characters, category padded to 15, and the queue from `routing` for that category. If your category is an enum, convert it back to its JSON name (`lost-and-found`, not `LostAndFound`) before the `routing` lookup, and reuse that helper in step 5.

   `reference["routing"]` and `reference["labels"]` are keyed by the JSON names, so every lookup and comparison uses `category.value` (the string), not the Enum member. A `str` Enum member already has that string as `.value`, so no helper is needed.

   `sorted(results, key=...)` returns a sorted copy, ordered by whatever the `key` function returns for each pair. `lambda r: r[1] != Category.emergency` takes a pair `r` and returns `False` for an emergency and `True` for everything else; `False` sorts before `True`, so emergencies come first. Put this directly below the header lines:

   ```python
   for inquiry, category in sorted(results, key=lambda r: r[1] != Category.emergency):
       print(f"{inquiry['id']:<10} {category.value:<15} {reference['routing'][category.value]}")
   ```

**Check:** two lines in the emergency block, `inq-0013` and `inq-0041`, both before the table. The table has 20 rows. `inq-0035` reads `unsure` and routes to `human-review-queue (a ranger reads it and picks the queue)`.

### Step 5: Score accuracy, then score emergency recall separately

**Do:**
1. Count results where `category == labels[id]`. That count is the accuracy numerator.

   This goes below the routing table, at the end of the file. `sum(1 for ... if ...)` adds 1 for every pair where the test is true, so it counts them. `labels[i["id"]]` is the reference category for that inquiry:

   ```python
   labels = reference["labels"]
   correct = sum(1 for i, c in results if c.value == labels[i["id"]])
   ```
2. Collect ids in `labels` valued `emergency`; count how many appear in the step 4 emergency list. That count is emergency recall.

   `labels.items()` gives (id, category) pairs, so the first line keeps the ids whose category is `emergency`. The second counts the step 4 emergencies whose id is in that list. Put both below the `correct` line:

   ```python
   emergency_ids = [k for k, v in labels.items() if v == "emergency"]
   caught = sum(1 for i, _ in emergencies if i["id"] in emergency_ids)
   ```
3. Print `Accuracy vs reference labels: N/20` and `Emergency recall: N/2`.

   Directly below. `complete/` adds a verdict after the recall numbers; the `+` joins two strings, and `A if test else B` picks one of two texts:

   ```python
   print()
   print(f"Accuracy vs reference labels: {correct}/{len(results)}")
   print(f"Emergency recall: {caught}/{len(emergency_ids)} "
         + ("(all caught; the metric that matters)" if caught == len(emergency_ids) else "(MISSED ONE; this fails, whatever the accuracy says)"))
   ```
4. For each mismatch, print `miss: ` plus the id, the model's category, and the reference category.

   At the very end of the file:

   ```python
   for inquiry, category in results:
       if category.value != labels[inquiry["id"]]:
           print(f"  miss: {inquiry['id']} got {category.value}, reference says {labels[inquiry['id']]}")
   ```

Run from `starter/`:

```bash
uv run main.py
```

**Why:** accuracy is the headline number, but emergency recall is the number that decides whether this is safe to ship. A missed emergency is a person waiting in a queue nobody is watching.

**Check:** one recorded run scored 18/20 with two misses: `inq-0030` (wedding photographer) got `general`, reference `permit`; `inq-0051` (Sperry campfires) got `conditions`, reference `general`. Another recorded 17/20 with one more: `inq-0001` got `unsure`, reference `permit`. Either way, emergency recall is 2/2, no routine message is `emergency`, and `inq-0035` is `unsure`. An emergency anywhere but `emergency` fails, whatever the accuracy.

### Step 6: Fix the misses by editing the category descriptions, not the code

`complete/` stops at step 5 with the original descriptions. This step's edits are yours, and the Check below is the only answer key.

**Do:**
1. Extend the `permit` description to also cover whether an activity needs a permit at all:

```text
- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, and questions about whether an activity requires a permit at all, including billing problems and missing confirmations.
```

2. Narrow `conditions` to mean only whether a trail, road, or area is physically passable.
3. Let `general` own park rules and regulations.
4. Leave `Decide in this order` and the `unsure` description alone.
5. Run all 20 again and read the scoreboard.

All four edits happen inside the `prompt(text)` function's `f"""` string in `main.py`. Each description is plain text; a description may wrap onto more lines, and the wrapped lines keep their two leading spaces like the ones around them. For item 1, replace the three `- permit:` lines with the new wording. Paste only the three text lines, not the `# Hint:` line, since anything inside the `"""` string is sent to the model:

```python
# Hint: the permit lines inside prompt(), rewritten (wrap where you like)
- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, and questions about whether an activity requires a
  permit at all, including billing problems and missing confirmations.
```

Items 2 and 3 are the same kind of edit on the `- conditions:` and `- general:` lines. The wording is yours:

```python
# Hint: same shape, your words
- conditions: <only whether a trail, road, or area is physically passable>
- general: <park rules and regulations, plus what general already covers>
```

Then from `starter/`:

```bash
uv run main.py
```

This program prints the routing table, not JSON. Where the Check below says `{"category": "emergency"}` or `{"category": "unsure"}`, read the `category` column of the table for that id.

**Why:** the misses came from the wording, not the model. `permit` described only transactions on a reservation, so a question about whether a permit is needed read as trip planning, and `conditions` was loose enough to swallow a question about campfire rules. When a message lands in the wrong queue, sharpen the description of the queue that should have owned it before you touch the code.

**Check:** when I ran it, `inq-0051` (Sperry campfires) moved to `general` and `inq-0001` moved from `unsure` to `permit`, while `inq-0008` (Half Dome lottery) stayed `permit`. Those two moves are the ones your edit is aiming at, and they do not land every run (your results will vary; the model is non-deterministic). `inq-0030` (wedding photographer) usually stays `general` on `llama3.2`, even with the permit description above; a commercial-photography permit is a hard call for a small model, so treat it as a known miss rather than a sign your edit failed. `inq-0005` may move to `unsure`. `inq-0041` and `inq-0013` still return `{"category": "emergency"}`, and `inq-0035` still returns `{"category": "unsure"}`. At most one or two messages besides `inq-0035` sit in `unsure`. Accuracy lands between 17 and 19 out of 20; measured on `llama3.2`, the .NET and Python runs scored 18/20, missing `inq-0030` and `inq-0005`, and the TypeScript runs scored 19/20 because `inq-0005` stayed `conditions`. A run at 20/20 means check whether the descriptions now fit only these 20 messages. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves. `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goals

Pick either. Neither is built in `complete/`. The reasoning is in [`expected-output.md`](../expected-output.md) under "Stretch Goal".

- **Add a priority field.** Add `priority` to the result type next to `category`, with its own small set of allowed values. Add a line to the prompt that says what each priority value means (for example, high when someone's safety or health is at risk), or the model rates almost everything high. Print it in the routing table. Priority is a second axis. It says how fast, and category says where. Mixing the two is how a lost inhaler ends up in line behind a lost wedding ring. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority.

  Give priority its own `str` Enum, built like `Category`, and add a second field to `TriageResult`. Both go where the step 2 classes are, and `Priority` must sit above `TriageResult`, which uses it:

  ```python
  # Hint: a second small str Enum, and a second field on the result model
  class Priority(str, Enum):
      high = "<value>"
      normal = "<value>"


  class TriageResult(BaseModel):
      category: Category
      priority: Priority
  ```

  Then carry it through the loop and the table. Each pair becomes a triple, and every line that builds or splits a pair changes to match:

  ```python
  # Hint: three parts instead of two
  results: list[tuple[dict, Category, Priority]] = []
  parsed = response.choices[0].message.parsed
  results.append((inquiry, parsed.category, parsed.priority))
  for inquiry, category, priority in sorted(results, key=lambda r: r[1] != Category.emergency):
      print(f"{inquiry['id']:<10} {category.value:<15} {priority.value:<8} {reference['routing'][category.value]}")
  ```

  The other lines that split a pair need a third name too: `[(i, c) for i, c in results ...]` becomes `[(i, c, p) for i, c, p in results ...]` in `emergencies`, `for inquiry, _ in emergencies:` becomes `for inquiry, _, _ in emergencies:`, `for i, c in results` becomes `for i, c, _ in results` in `correct`, `for i, _ in emergencies` becomes `for i, _, _ in emergencies` in `caught`, and the step 5 miss loop becomes `for inquiry, category, _ in results:`.

- **Add a confidence threshold.** Add a numeric `confidence` field to the result type and schema. After the loop, change the category to `unsure` on any result whose confidence is under a threshold you pick. Run the scoreboard again. **Check:** both emergencies stay `emergency` with high confidence, `inq-0035` stays `unsure`, and the `unsure` queue does not fill up with ordinary permit questions. If a third of the slice lands in `unsure`, the threshold is too high and you have rebuilt the unsorted inbox.

  A `float` field on the model becomes a number in the generated schema:

  ```python
  # Hint: a number field next to category
  class TriageResult(BaseModel):
      category: Category
      confidence: float
  ```

  Store the confidence in the tuple the way the priority stretch stores priority. Each pair becomes a triple, and every line that builds or splits a pair changes to match (the priority stretch lists them). Printing the confidence in the routing table lets you check it; `:<5.2f` pads to 5 characters with 2 decimals:

  ```python
  # Hint: three parts instead of two
  results: list[tuple[dict, Category, float]] = []
  parsed = response.choices[0].message.parsed
  results.append((inquiry, parsed.category, parsed.confidence))
  for inquiry, category, confidence in sorted(results, key=lambda r: r[1] != Category.emergency):
      print(f"{inquiry['id']:<10} {category.value:<15} {confidence:<5.2f} {reference['routing'][category.value]}")
  ```

  Then rebuild `results` after the loop's `print("\n")` and before the `emergencies` line. The comprehension keeps each triple but swaps in `Category.unsure` when the confidence is under your threshold:

  ```python
  # Hint: after the loop, pick your own threshold
  THRESHOLD = <your number>
  results = [(i, Category.unsure if conf < THRESHOLD else c, conf) for i, c, conf in results]
  ```

## What Is in This Folder

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder, part of the workshop corpus and shared with feature 09), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix matches the full inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the hand-assigned category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus, the source the reference labels were drawn from. Not for handout.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
