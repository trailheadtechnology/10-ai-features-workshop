<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 09: Human-in-the-Loop (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F09-dotnet.md), [TypeScript](../typescript/F09-typescript.md). Lab overview: [F09-lab.md](../F09-lab.md).*

**The User Problem:** Feature 07 routed the inbox, so now a ranger stares at forty messages that all need replies. Most answers are boilerplate the ranger has typed a hundred times, and typing them eats the afternoon. The obvious move is to let the AI answer, and the obvious disaster is the AI telling a visitor that campfires are fine during a burn ban, on official park letterhead. The ranger's problem is drudgery; the park's problem is that full automation of outbound communication is how you end up apologizing publicly.

*A Challenge lab. Do it if you finished [Module 3](../../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** put a human between the model and the visitor. The model drafts, a ranger approves, edits, or rejects, every decision is logged, and emergencies never reach the model at all.
- **Input:** `data/inquiries.jsonl`, six inquiries routed by feature 07, with category and park doc; `data/snippets/`, the four excerpts they cite (`glac-bc-2025-04.md`, `glac-cl-2026-01.md`, `yose-cl-2026-01.md`, `zion-nar-2026-01.md`); `policy-worksheet.md`, the lane table you fill in.
- **How:** the starter already loops over the six inquiries and sends each one to the chat model with a drafting prompt. You add five things: a policy table, a gate above the model call, a review prompt, an outbox, and an audit log.
- **Model:** `llama3.2`, local. No key.

## The Concept

Human-in-the-loop is a product pattern, not a model feature, and it's the difference between AI features that ship and AI features that get killed in legal review. The core move: the AI drafts, the human approves, edits, or rejects, and the system remembers what happened. The user-facing risk drops to near zero while most of the typing still disappears.

The design question is where to put the human, and the answer comes from error cost and reversibility, which connects straight back to feature 07's asymmetry lesson. A sensible policy has three lanes: full automation for cheap, reversible, low-stakes replies; draft-plus-approval for the middle; human-only for the expensive and irreversible (in Trailhead Guides terms, emergencies never get an AI draft at all). Two practical details do a lot of work in real systems. Keep an audit trail of what was drafted, who approved it, and what they changed. And measure the gap between draft and final text, because how much humans edit tells you whether trust in each lane is earned, and edit patterns show you exactly where the drafts fall short.

Every step below is one thing to make the program do. The `starter/` drafts and "sends" every inquiry with no review. Edit it until it does all seven steps. Compare against `complete/` when stuck; its comments say why each piece is there.

## Running

`outbox/` and `decisions.jsonl` are run artifacts and are gitignored.

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. `starter/main.py` is deliberately unsafe: every draft goes straight out, nothing is logged, and nothing treats an emergency differently except a sentence in the system prompt that the model is free to ignore. `complete/main.py` is the finished demo as shown on stage.

The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [SETUP.md](../../../../SETUP.md)), and `uv run` finds it from any folder, with no venv to activate. Run from the `starter/` folder:

```bash
uv run main.py
```

`starter/main.py` takes no flags, at most the one positional argument its header comment names. `complete/` adds flags for the stretch goals. From `complete/`:

```bash
uv run main.py                            # review the queue interactively
uv run main.py --policy                   # print the routing policy table and exit
uv run main.py --auto-approve-dry-run     # non-interactive run for testing
```

### Step 0: Run the starter and read every draft as an editor

**Do:** run `starter/` as it is. Read the code first. It is the loop you are about to change:

1. It opens `../../data/inquiries.jsonl`: `id`, `channel`, `received`, `category`, `doc`, `text` per line.

   The starter already does this. `sys.argv` is the command-line argument list (`sys.argv[1]` is the first argument), `DATA` is the `data/` folder two levels above `main.py`, and `json.loads` turns one line of JSON into a dict, so later code reads fields as `inquiry["id"]` or `inquiry['text']`:

   ```python
   DATA = Path(__file__).resolve().parents[2] / "data"
   inquiries_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DATA / "inquiries.jsonl"
   data_dir = inquiries_path.resolve().parent
   ```

   ```python
   for line in inquiries_path.read_text().splitlines():
       if not line.strip():
           continue
       inquiry = json.loads(line)
   ```

2. For each inquiry, if `doc` is not empty and `data/snippets/<doc>` exists, it reads and trims that file as the excerpt. Otherwise the excerpt is the literal text `(none on file for this message)`.

   The starter already does this, inside the loop. `inquiry.get("doc")` returns the `doc` value, or `None` when it is missing, and `x if condition else y` picks one of two values:

   ```python
       snippet_path = data_dir / "snippets" / (inquiry.get("doc") or "")
       snippet = snippet_path.read_text().strip() if inquiry.get("doc") and snippet_path.exists() else "(none on file for this message)"
   ```

3. It builds a system message (the drafting prompt, unchanged across inquiries) and a user message (excerpt + channel + received + text), sends both to `llama3.2`, and trims the reply.

   The starter already does this. `client` is created once above the loop. `client.chat.completions.create` takes the list of messages and returns a response object; the reply text is `draft.choices[0].message.content`, which step 0 item 4 prints:

   ```python
   client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
   ```

   ```python
       draft = client.chat.completions.create(
           model="llama3.2",
           messages=[
               {"role": "system", "content": SYSTEM_PROMPT},
               {"role": "user", "content": f"""Reference excerpt:
   {snippet}

   Visitor message ({inquiry['channel']}, received {inquiry['received']}):
   {inquiry['text']}

   Draft the reply."""},
           ],
       )
   ```

4. It prints `=== SENT to visitor · <id> (<category>) ===`, the draft, then a blank line. After the loop: `All replies sent. Nobody read them. Nothing was logged.`

   The starter already does this. The first three lines are the end of the loop body; the last line is after the loop, with no indent:

   ```python
       print(f"=== SENT to visitor · {inquiry['id']} ({inquiry['category']}) ===")
       print((draft.choices[0].message.content or "").strip())
       print()
   ```

   ```python
   print("All replies sent. Nobody read them. Nothing was logged.")
   ```

The system prompt:

```text
You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers. Exception: if the visitor's message reports an emergency, an injury, a possible fire, or a missing or overdue person, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch.
```

```bash
uv run main.py
```

**Why:** `data/inquiries.jsonl` is six of the 100 messages in feature 07's inbox, with two fields feature 07's output would supply: `category` (feature 07's `conditions` is spelled `trail-condition` here) and `doc` (the park-doc excerpt file name in `data/snippets/`, or empty for `inq-0007` and `inq-0013`). Each snippet file quotes only the sections that answer its inquiry, word for word from the full documents in feature 05's corpus, so the model sees the source section and nothing else.

**Check:** read every draft the way a ranger would. The drafts move around run to run, so treat the list below as the kind of thing to look for rather than a script. When I ran it:
- `inq-0002` said the Mist Trail is closed when the excerpt says it reopened (reject).
- `inq-0051` and `inq-0005` were accurate and cited (approve).
- `inq-0003` got both rules right but pinned the flash flood rule on `GLAC-BC-2026-01`, a Glacier document number, on a Zion question (edit).
- `inq-0007` apologized and decided nothing (edit).
- `inq-0013`, the overdue-hiker report, gets a warm reply to Diane with no `ESCALATE` line, even though the prompt told the model not to draft one. That is the one to watch for, and it is not an unlucky run: in the recorded runs it happened 3 times out of 3. If you get an `ESCALATE` line on your first try, run it again.

The rest of the lab exists because of that last draft.

### Step 1: Fill in policy-worksheet.md

**Do:** open `policy-worksheet.md`. It sits in the feature folder, one level above your track: `modules/M3-deciding/F09-human-in-the-loop/policy-worksheet.md`. It has one row per feature 07 category (`permit`, `trail-condition`, `complaint`, `lost-and-found`, `general`, `emergency`; `unsure` has no row), five columns: Category, Lane, Worst plausible error, Reversible?, Justification.

1. For each row, pick a lane: `auto-send`, `draft-for-approval`, or `human-only`.
2. Write the worst plausible wrong reply for that category.
3. Write whether that error can be taken back after it's sent.
4. Write a one-sentence justification.
5. Answer the two follow-up questions under the table (the second asks where your lane lives in the code; steps 2 and 3 are the answer).

**Check:** every row says what a wrong reply costs, and emergency is `human-only`. If your reason for a lane is "the prompt tells it to," reread the `inq-0013` draft from step 0. Your lanes may differ from the reference in `expected-output.md`. Your justifications are what count.

### Step 2: Add the policy table and print the lane

**Do:**
1. Above the loop, build a dictionary from category to lane:

```text
trail-condition   draft-for-approval
permit            draft-for-approval
complaint         draft-for-approval
general           draft-for-approval
lost-and-found    draft-for-approval
emergency         human-only
```

A dict maps each category (the key) to its lane (the value). Put it right below the closing `"""` of `SYSTEM_PROMPT`, above the `DATA = ...` line. This is exactly what `complete/` has:

```python
POLICY = {
    "trail-condition": "draft-for-approval",
    "permit": "draft-for-approval",
    "complaint": "draft-for-approval",
    "general": "draft-for-approval",
    "lost-and-found": "draft-for-approval",
    "emergency": "human-only",
}
```

2. Print the table once at the top of the run, under `Routing policy (error cost decides the lane):`, one row per entry, two-space indent, category left-aligned in a 16-character column, then the lane, then a blank line after the table.

   Write the printout as a function so the `--policy` stretch goal can reuse it. Put it directly below the `POLICY` dict, with two blank lines around it. `POLICY.items()` walks the dict and gives each key and value as `category, lane`, and in `{category:<16}` the `<16` pads the value to 16 characters, left-aligned:

   ```python
   def print_policy() -> None:
       print("Routing policy (error cost decides the lane):")
       for category, lane in POLICY.items():
           print(f"  {category:<16} {lane}")
       print()
   ```

   Then call it once, with no indent, on its own line directly above `for line in inquiries_path.read_text().splitlines():`:

   ```python
   print_policy()
   ```

3. Inside the loop, right after parsing the inquiry, look up its `category`. If not found, use `human-only`.

   `POLICY.get(key, default)` returns the lane when the category is in the dict and the second argument when it is not, which is the fail-closed lookup. Add it on the line directly below `inquiry = json.loads(line)`, at the same indent:

   ```python
       lane = POLICY.get(inquiry["category"], "human-only")
   ```

4. Print a per-inquiry header right after the lane lookup, before the excerpt is read (the old `SENT` line after the model call goes away in step 4): 72 dashes, `<id>  ·  <category>  ·  <channel>  ·  lane: <lane>`, another 72 dashes, the visitor's `text` prefixed with `  | ` per line, blank line.

   `"-" * 72` builds a string of 72 dashes. Add these lines directly below the `lane = ...` line, above `snippet_path = ...`:

   ```python
       print("-" * 72)
       print(f"{inquiry['id']}  ·  {inquiry['category']}  ·  {inquiry['channel']}  ·  lane: {lane}")
       print("-" * 72)
       print(indent(inquiry["text"]))
       print()
   ```

   `indent` is one more helper function. It splits the text on newlines, puts `  | ` in front of each line, and joins the lines back together with `"\n".join(...)`. Put it directly below `print_policy`, above the `DATA = ...` line, with two blank lines around it:

   ```python
   def indent(text: str) -> str:
       return "\n".join("  | " + l.rstrip() for l in text.split("\n"))
   ```

```bash
uv run main.py
```

**Why:** feature 07 can return `unsure`, and the worksheet has no row for it. Falling back to `human-only` means a category you did not plan for goes to a ranger rather than out as a draft.

**Check:** the run opens with the six-row table. Every inquiry header ends in its lane. `inq-0013` shows `lane: human-only` and the other five show `lane: draft-for-approval`. The drafts still print for all six, including the emergency. The next step fixes that.

### Step 3: Gate the emergency above the model call

**Do:** this is the whole feature. The check must run before any message is built; the model call must stay below it.

1. Right after the lane lookup and header, and before the excerpt is read or a message is built, test `lane == "human-only"`.
2. If true, print `  NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and move to the next inquiry without reading the excerpt, building a request, or spending tokens.
3. Only below that test, read the excerpt and call the model as before.

Add this block directly below the header's final `print()`, above the `snippet_path = ...` line, at the same 4-space indent. `continue` jumps straight to the next inquiry, so nothing below it in the loop runs: no excerpt, no message, no `client.chat.completions.create` call. The `\n` at the end of the text prints the blank line after it.

```python
    if lane == "human-only":
        print("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.\n")
        continue
```

Step 5 adds the log line inside this block, above `continue`.

```bash
uv run main.py
```

**Why:** the system prompt already told the model not to draft a reply for a message like `inq-0013`, and in the recorded runs it drafted one anyway, 3 times out of 3. A prompt instruction is a request; this lane is a guarantee. The test sits above the model call, so the emergency never depends on the model cooperating. Leave the `ESCALATE` sentence in the system prompt where it is.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and the program makes no model call for it. The other five still get drafts.

### Step 4: Ask the reviewer and queue approved text in an outbox

**Do:**
1. Replace the `SENT` printout: after the model call, print `  draft:`, blank line, the draft prefixed with `  | ` per line, blank line.

   First, in the model call from step 0, rename the result from `draft` to `response`, so the name `draft` is free for the trimmed string. Change the first line of the call to:

   ```python
       response = client.chat.completions.create(
   ```

   Then delete the starter's three `SENT` lines after the call:

   ```python
       print(f"=== SENT to visitor · {inquiry['id']} ({inquiry['category']}) ===")
       print((draft.choices[0].message.content or "").strip())
       print()
   ```

   In their place, directly below the `)` that ends the model call, show the draft. `indent` is the helper from step 2:

   ```python
       draft = (response.choices[0].message.content or "").strip()
       print("\r  draft:      \n")
       print(indent(draft))
       print()
   ```

   `complete/` also prints `  drafting...` on the line above `response = ...`, and the `\r` in the draft line moves the cursor back to overwrite it. `end=""` keeps the cursor on the same line. Optional:

   ```python
       print("  drafting...", end="", flush=True)
   ```

2. Print `  [a]pprove  [e]dit  [r]eject  [s]kip > ` and read one line; trim and lowercase it.

   `input(prompt)` prints the prompt without a newline, waits for Enter, and returns the line. When there is no more input (a piped run, or Ctrl-D) it raises `EOFError` instead, and `try`/`except` turns that into a skip. Add below the draft display:

   ```python
       try:
           key = input("  [a]pprove  [e]dit  [r]eject  [s]kip > ").strip().lower()
       except EOFError:
           key = "s"
       print()
   ```

3. Map the key to a decision and final text: `a` → `approved`, final = draft. `e` → `edited`, final = what the reviewer types. `r` → `rejected`, no final text. Anything else → `skipped`, no final text.

   `final` starts as `None`, which is how "no final text" is stored; `str | None` says it may hold either. An `if`/`elif`/`else` chain picks the branch that matches `key`, and `else` catches everything else. `decision, final = "approved", draft` sets both names in one line. First add the `final` line directly above the `try:` from item 2, at the same indent (the auto-approve stretch goal needs it above the prompt):

   ```python
       final: str | None = None
   ```

   Then add the chain directly below the `print()` that follows the `input` lines:

   ```python
       if key == "a":
           decision, final = "approved", draft
       elif key == "e":
           decision, final = "edited", read_edited(draft)
       elif key == "r":
           decision = "rejected"
       else:
           decision = "skipped"
   ```

   `read_edited` does not exist yet. Item 4 writes it.

4. For `e`, read lines until a line that is just `.`. If the first line is empty, copy the draft in first and keep reading (so the reviewer can append instead of retyping). Join, trim; if empty, fall back to the draft.

   `read_edited` is a helper function, so it goes above the loop with `print_policy` and `indent` (directly below `indent`, with two blank lines around it). It collects lines in a list inside a `while True:` loop that only `break` ends; `EOFError` ends it too, as in item 2. `"\n".join(lines)` glues the lines back together, and `edited or draft` returns the draft when `edited` is empty:

   ```python
   def read_edited(draft: str) -> str:
       print("  Type the reply you want to send. End with a single '.' on its own line.")
       print("  Press Enter on the first line to start from the draft text instead.\n")
       lines: list[str] = []
       first = True
       while True:
           try:
               line = input()
           except EOFError:
               break
           if line == ".":
               break
           if first and line == "":
               lines.append(draft)
               print("  (draft copied in; keep typing to append, '.' to finish)")
           else:
               lines.append(line)
           first = False
       edited = "\n".join(lines).strip()
       return edited or draft
   ```

5. At startup create `outbox/` in `starter/`, the folder holding the source file you are editing, and put `decisions.jsonl` beside it. Anchor those two paths so they land there under your track's run command, never next to a compiled binary. When there's a final text, write it plus a newline to `outbox/<id>.txt` and print `  -> <decision>, queued at <path>`, where `<path>` is that file relative to the folder you ran from (`outbox/<id>.txt` when you run from `starter/`). Otherwise print `  -> <decision>, nothing queued`.

   `uv run main.py` keeps your shell's folder as the working directory, but anchor the two paths to the script's own folder anyway, so they land in `starter/` no matter where you run from. `Path(__file__).resolve().parent` is the folder holding `main.py`, and `/` joins a folder and a name. Add `import os` at the top of the file, below `import json`:

   ```python
   import os
   ```

   Declare both paths directly below the `inquiries_path = ...` line, above `data_dir = ...` (`decisions_path` is used in step 5):

   ```python
   HERE = Path(__file__).resolve().parent
   outbox_dir = HERE / "outbox"
   decisions_path = HERE / "decisions.jsonl"
   ```

   Create the folder once, directly below the `data_dir = ...` line. `parents=True, exist_ok=True` means "also create missing parent folders, and do not fail if it already exists":

   ```python
   outbox_dir.mkdir(parents=True, exist_ok=True)
   ```

   Then, inside the loop below the `if`/`elif`/`else` chain, write the file only when there is final text. `outbox_dir / f"{inquiry['id']}.txt"` is `outbox/inq-0051.txt` inside `starter/`, `write_text(final + "\n")` adds the trailing newline, and `os.path.relpath(path)` prints the path relative to the folder you ran from:

   ```python
       if final is not None:
           path = outbox_dir / f"{inquiry['id']}.txt"
           path.write_text(final + "\n")
           print(f"  -> {decision}, queued at {os.path.relpath(path)}\n")
       else:
           print(f"  -> {decision}, nothing queued\n")
   ```

   That `if`/`else` is the last thing in the loop body for now. Step 5 adds the log line below it.

```bash
uv run main.py
```

**Check:** approve `inq-0051` and `outbox/inq-0051.txt` appears holding the draft. Reject `inq-0002` and no file appears for it. Edit `inq-0003` by pressing Enter, typing one line, then `.`, and `outbox/inq-0003.txt` holds the draft with your line appended.

The file is `starter/outbox/inq-0051.txt`.

### Step 5: Log every decision to decisions.jsonl with an edit distance

**Do:**
1. Write a Levenshtein edit-distance function (single-character inserts/deletes/substitutions to turn one string into another; two-row DP is fine).

   Another helper function for above the loop, directly below `read_edited`. `previous` and `current` are the two rows: `previous[j]` holds the distance from the first `i - 1` characters of `a` to the first `j` characters of `b`. `list(range(len(b) + 1))` is the list `[0, 1, 2, ...]`, and `enumerate(a, 1)` walks the characters of `a` while counting from 1. Each cell takes the cheapest of insert, delete, or substitute, and `previous = current` makes the row just filled the one the next pass reads:

   ```python
   def edit_distance(a: str, b: str) -> int:
       previous = list(range(len(b) + 1))
       for i, ca in enumerate(a, 1):
           current = [i] + [0] * len(b)
           for j, cb in enumerate(b, 1):
               cost = 0 if ca == cb else 1
               current[j] = min(current[j - 1] + 1, previous[j] + 1, previous[j - 1] + cost)
           previous = current
       return previous[len(b)]
   ```

2. After every review decision, append one JSON object per line to `decisions.jsonl`: `at` (UTC ISO 8601), `inquiryId`, `category`, `lane`, `decision`, `reviewer` (OS username), `draft`, `final` (null when nothing queued), `editDistance` (distance from draft to final, or to empty string when there's no final text). Match the key names exactly.

   Add two imports at the top of the file, with the other imports. `getpass.getuser()` is the OS username and `datetime.now(timezone.utc).isoformat()` is the UTC timestamp:

   ```python
   import getpass
   ```

   ```python
   from datetime import datetime, timezone
   ```

   Set the reviewer above the loop, directly below `client = OpenAI(...)`. `complete/` picks the name from the `--auto-approve-dry-run` stretch flag, so `auto_approve` stays `False` until you add that flag. Put the `auto_approve` line directly below `decisions_path = ...`:

   ```python
   auto_approve = False
   ```

   ```python
   reviewer = "auto-approve-dry-run" if auto_approve else getpass.getuser()
   ```

   Two helper functions, above the loop with the others (directly below `indent` is fine). `record` builds a dict with the lab's exact key names. `log` opens the file in append mode (`"a"`, which creates it on first use) and writes the dict as one line: `json.dumps` turns the dict into JSON text and writes Python `None` as JSON `null`:

   ```python
   def log(decision: dict) -> None:
       with decisions_path.open("a") as f:
           f.write(json.dumps(decision) + "\n")


   def record(inquiry, lane, decision, draft, final, distance) -> dict:
       return {
           "at": datetime.now(timezone.utc).isoformat(),
           "inquiryId": inquiry["id"],
           "category": inquiry["category"],
           "lane": lane,
           "decision": decision,
           "reviewer": reviewer,
           "draft": draft,
           "final": final,
           "editDistance": distance,
       }
   ```

   Inside the loop, directly below the outbox `if`/`else` from step 4, log the decision. `final or ""` measures against an empty string when `final` is `None`:

   ```python
       log(record(inquiry, lane, decision, draft, final, edit_distance(draft, final or "")))
   ```

3. In the step 3 gate, before moving on, append the same shape with `decision: "escalated"`, `draft`/`final` both null, `editDistance` 0.

   Inside the step 3 `if lane == "human-only":` block, between the `NO DRAFT` line and `continue`, at the same 8-space indent:

   ```python
           log(record(inquiry, lane, "escalated", None, None, 0))
   ```

4. Count decisions by name. After the loop print 72 equals signs, then `Queue done: ` plus the counts (e.g. `1 escalated, 3 approved, 1 edited, 1 rejected`), then `Audit trail: decisions.jsonl   ·   Outbox: outbox/` (both paths relative to the folder you ran from, as in step 4).

   A `Counter` is a dict that starts every missing key at 0, so `counts["approved"] += 1` works the first time too. Import it at the top of the file:

   ```python
   from collections import Counter
   ```

   Declare it above the loop, directly below the `reviewer = ...` line:

   ```python
   counts: Counter[str] = Counter()
   ```

   Add one to a count right below each `log(...)` call. In the gate, above `continue`:

   ```python
           counts["escalated"] += 1
   ```

   At the bottom of the loop, below the other `log(...)`:

   ```python
       counts[decision] += 1
   ```

   Replace the starter's closing `All replies sent` line (no indent, after the loop) with the summary. `counts.items()` gives each name and count, `f"{v} {k}"` turns each pair into text like `3 approved`, and `", ".join(...)` puts `, ` between them:

   ```python
   print("=" * 72)
   print("Queue done: " + ", ".join(f"{v} {k}" for k, v in counts.items()))
   print(f"Audit trail: {os.path.relpath(decisions_path)}   ·   Outbox: {os.path.relpath(outbox_dir)}/")
   ```

```bash
uv run main.py
```

**Check:** one full run adds six lines to `decisions.jsonl`. One is `inq-0013` with `"decision":"escalated"` and `"draft":null`. Approved lines have `"editDistance":0`. A rejected line's `editDistance` equals its draft's length. (`decisions.jsonl` and `outbox/` are run artifacts. Delete them between runs for a clean take.)

`json.dumps` puts a space after each colon, so in your file those read `"decision": "escalated"`, `"draft": null`, and `"editDistance": 0`.

### Step 6: Run the queue and compare with expected-output.md

**Do:**
1. Run your program and review all six inquiries. Approve at least one, edit at least one, reject at least one.
2. Read the annotated drafts in `expected-output.md` and compare your choices against what a ranger should do and why.
3. Read the Reference Policy table in `expected-output.md` against your worksheet from step 1.

```bash
uv run main.py
```

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` with no model call; every other draft offers `[a]pprove [e]dit [r]eject [s]kip`, logs to `decisions.jsonl`, and queues approved text in `outbox/`. Your lanes may differ from `expected-output.md`. Your justifications are what count.

### Stretch goals

Pick any. Each one is already built in `complete/`, with the measured reason for it in [`expected-output.md`](../expected-output.md).

- **Try the prompt repair and watch it fail.** Move the escalation rule to the front of the system prompt and add "write nothing after that line":

  ```text
  FIRST, before anything else, check the visitor's message for an emergency: an injury, a possible fire, or a missing or overdue person. If you see one, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch. Write nothing after that line.
  Otherwise, you are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
  ```

  Put this in your program's system message. Your step 3 gate stops `inq-0013` before the model call, so on the real queue it prints `NO DRAFT` and the model never sees it. To watch what the prompt does with an emergency, make a copy of `data/inquiries.jsonl` named `data/inquiries-miscategorized.jsonl`, change `inq-0013`'s `category` from `emergency` to `general` in the copy, and run the whole queue from the copy once:

  ```bash
  uv run main.py ../../data/inquiries-miscategorized.jsonl
  ```

  **Check:** on the copy, `inq-0013` usually prints `ESCALATE:` and then the reply anyway (3 runs out of 3 in the record), and the reply invents an active search; some runs skip the `ESCALATE:` line and just invent the search. Under the same prompt the model also escalates a routine message or two, such as `inq-0002`, `inq-0003`, or `inq-0005`; which ones changes from run to run. Tightening the prompt trades one failure for the other. Put the original prompt back.

  The system message is the `SYSTEM_PROMPT` string at the top of main.py. A `"""` string keeps every line between the opening and closing quotes, so replace only the text inside them, and keep the two lines of the new prompt on two lines. Keep the original somewhere (a copy of the file is fine) so you can put it back:

  ```python
  # Hint:
  SYSTEM_PROMPT = """FIRST, before anything else, ... Write nothing after that line.
  Otherwise, you are drafting a reply ... Never invent dates, fees, policies, or phone numbers."""
  ```

- **Add the ESCALATE backstop.** After the model call and before the review prompt, test whether the draft starts with `ESCALATE` (case-insensitive). If it does, print `  Model asked to escalate. Draft discarded, routing to a human.`, log a line with `decision: "escalated"`, draft kept, `final` null, `editDistance` 0, then move on. **Why:** this catches an emergency that arrived under the wrong category. It runs after the model has already answered, so it's a backup, never the main control. **Check:** run against `data/inquiries-miscategorized.jsonl`, the copy from the previous stretch goal with `inq-0013` labeled `general` (make it now if you skipped that goal). Most runs offer Diane's warm reply for approval, because the model doesn't escalate. On a run where it does, the backstop line prints and the reviewer never sees the draft. That gap is why step 3 exists.

  The test goes directly below the draft display from step 4 (below the `print(indent(draft))` and `print()` lines) and above `final: str | None = None`. It looks like the step 3 gate, with the draft passed to `record` instead of `None`. `draft.upper()` makes the test case-insensitive:

  ```python
      if draft.upper().startswith("ESCALATE"):
          print("  Model asked to escalate. Draft discarded, routing to a human.\n")
          log(record(inquiry, lane, "escalated", draft, None, 0))
          counts["escalated"] += 1
          continue
  ```

  To run against the copy, from `starter/` (the starter still reads its first argument as the queue file):

  ```bash
  uv run main.py ../../data/inquiries-miscategorized.jsonl
  ```

- **Add the flags `complete/` supports.** `--policy` prints the routing table and exits without reading the queue. `--auto-approve-dry-run` sets the reviewer to `auto-approve-dry-run`, approves every draft without asking, and prints `--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.` under the table. `--outbox <dir>` and `--decisions <file>` move the run artifacts. Any other argument is the path to a queue file. **Check:** `--policy` prints six rows and nothing else. `--auto-approve-dry-run` writes six lines to `decisions.jsonl`, five `approved` and one `escalated` (or four and two, if you added the backstop and `llama3.2` wrote a spurious `ESCALATE` on a routine message), and one file in `outbox/` for each approved draft.

  `complete/main.py` shows the argument loop: a `while i < len(args)` over `sys.argv[1:]`. First change the starter's `inquiries_path` line so it no longer reads `sys.argv[1]`:

  ```python
  inquiries_path = DATA / "inquiries.jsonl"
  ```

  The loop has to come after the four names it sets (and `HERE`) exist and before `data_dir = ...` and `outbox_dir.mkdir(...)`, which use them. Check that `inquiries_path`, `HERE`, `outbox_dir`, `decisions_path`, and `auto_approve` sit together above `data_dir = ...` (move any that do not), then put the loop directly below them. `args[i]` is the current argument, `i += 1` inside a branch moves on to the value after the flag, and `raise SystemExit` ends the program:

  ```python
  args = sys.argv[1:]
  i = 0
  while i < len(args):
      if args[i] == "--auto-approve-dry-run":
          auto_approve = True
      elif args[i] == "--outbox":
          i += 1
          outbox_dir = Path(args[i])
      elif args[i] == "--decisions":
          i += 1
          decisions_path = Path(args[i])
      elif args[i] == "--policy":
          print_policy()
          raise SystemExit
      else:
          inquiries_path = Path(args[i])
      i += 1
  ```

  For `--auto-approve-dry-run`, print the warning directly below the `print_policy()` call above the queue loop:

  ```python
  if auto_approve:
      print("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.\n")
  ```

  Then wrap step 4's prompt and `if`/`elif`/`else` chain in an `if`/`else`, so the reviewer is only asked when the flag is off. Everything that was there moves one level (4 spaces) to the right under `else:`, and `final: str | None = None` stays above the new `if`:

  ```python
      if auto_approve:
          decision = "approved"
          final = draft
          print("  [auto] approved\n")
      else:
          try:
              key = input("  [a]pprove  [e]dit  [r]eject  [s]kip > ").strip().lower()
          except EOFError:
              key = "s"
          print()
          if key == "a":
              decision, final = "approved", draft
          elif key == "e":
              decision, final = "edited", read_edited(draft)
          elif key == "r":
              decision = "rejected"
          else:
              decision = "skipped"
  ```

  Run it with `uv run main.py --policy` or `uv run main.py --auto-approve-dry-run`.

- **Use edit distance as the promotion signal.** Approve, edit, and reject a few drafts, then read `editDistance` in `decisions.jsonl` and argue for a threshold that would move a category from `draft-for-approval` to `auto-send`. **Check:** you name numbers. The reference gate in `expected-output.md` is 90 days of review, at least 200 reviewed messages, a median edit distance under 5 percent of draft length, and zero decisions tagged as factual corrections. Edit distance can't tell a comma from a lawsuit, so that last gate needs a field the review UI asks for.

## What Is in This Folder

- `data/inquiries.jsonl`: six inquiries drawn from feature 07's full 100-message inbox, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `data/snippets/`: the four park-doc excerpts, one file per document number, each holding only the sections that answer its inquiry, taken from the full park documents in feature 05's `data/park-docs/` and quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
