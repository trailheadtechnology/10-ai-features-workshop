# Lab 09: Human-in-the-Loop

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** put a human between the model and the visitor. The model drafts, a ranger approves, edits, or rejects, every decision is logged, and emergencies never reach the model at all.
- **Input:** `data/inquiries.jsonl`, six inquiries routed by feature 07, with category and park doc; `data/snippets/`, the four excerpts they cite (`glac-bc-2025-04.md`, `glac-cl-2026-01.md`, `yose-cl-2026-01.md`, `zion-nar-2026-01.md`); `policy-worksheet.md`, the lane table you fill in.
- **How:** the starter already loops over the six inquiries and sends each one to the chat model with a drafting prompt. You add five things: a policy table, a gate above the model call, a review prompt, an outbox, and an audit log.
- **Model:** `llama3.2`, local. No key.

Every step below is one thing to make the program do. Every track's `starter/` drafts and "sends" every inquiry with no review. Edit it until it does all seven steps. Compare against `complete/` when stuck.

### Step 0: Run the starter and read every draft as an editor

**Do:** run `starter/` as it is. Read the code first — it's the loop you're about to change:

1. It opens `../../data/inquiries.jsonl`: `id`, `channel`, `received`, `category`, `doc`, `text` per line.
2. For each inquiry, if `doc` is not empty and `data/snippets/<doc>` exists, it reads and trims that file as the excerpt. Otherwise the excerpt is the literal text `(none on file for this message)`.
3. It builds a system message (the drafting prompt, unchanged across inquiries) and a user message (excerpt + channel + received + text), sends both to `llama3.2`, and trims the reply.
4. It prints `=== SENT to visitor · <id> (<category>) ===`, the draft, then a blank line. After the loop: `All replies sent. Nobody read them. Nothing was logged.`

The system prompt:

```text
You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers. Exception: if the visitor's message reports an emergency, an injury, a possible fire, or a missing or overdue person, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch.
```

**Why:** `data/inquiries.jsonl` is six of the 100 messages in feature 07's inbox, with two fields feature 07's output would supply: `category` (feature 07's `conditions` is spelled `trail-condition` here) and `doc` (the park-doc excerpt file name in `data/snippets/`, or empty for `inq-0007` and `inq-0013`). Each snippet file quotes only the sections that answer its inquiry, word for word from the full documents in feature 05's corpus, so the model sees the source section and nothing else.

**Check:** read every draft the way a ranger would.
- `inq-0002` says the Mist Trail is closed when the excerpt says it reopened (reject).
- `inq-0051` and `inq-0005` are accurate and cited (approve).
- `inq-0003` gets both rules right but pins the flash flood rule on `GLAC-BC-2026-01`, a Glacier document number, on a Zion question (edit).
- `inq-0007` apologizes and decides nothing (edit).
- `inq-0013` — the overdue-hiker report — gets a warm reply to Diane with no `ESCALATE` line, even though the prompt told the model not to draft one; in the recorded runs that happened 3 times out of 3.

The rest of the lab exists because of that last draft.

### Step 1: Fill in policy-worksheet.md

**Do:** open `policy-worksheet.md` — one row per feature 07 category (`permit`, `conditions`, `complaint`, `lost-and-found`, `general`, `emergency`; `unsure` has no row), five columns: Category, Lane, Worst plausible error, Reversible?, Justification.

1. For each row, pick a lane: `auto-send`, `draft-for-approval`, or `human-only`.
2. Write the worst plausible wrong reply for that category.
3. Write whether that error can be taken back after it's sent.
4. Write a one-sentence justification.
5. Answer the two follow-up questions under the table (the second asks where your lane lives in the code — steps 2 and 3 are the answer).

**Check:** every row says what a wrong reply costs, and emergency is `human-only`. If your reason for a lane is "the prompt tells it to," reread the `inq-0013` draft from step 0. Your lanes may differ from the reference in `expected-output.md` — your justifications are what count.

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

2. Print the table once at the top of the run, under `Routing policy (error cost decides the lane):`.
3. Inside the loop, right after parsing the inquiry, look up its `category`. If not found, use `human-only`.
4. Replace the `SENT` header with one that shows the lane: 72 dashes, `<id>  ·  <category>  ·  <channel>  ·  lane: <lane>`, another 72 dashes, the visitor's `text` prefixed with `  | ` per line, blank line.

**Why:** an unknown category falling back to `human-only` means it fails closed — it gets the safest lane, not a draft.

**Check:** the run opens with the six-row table. Every inquiry header ends in its lane. `inq-0013` shows `lane: human-only` and the other five show `lane: draft-for-approval`. The drafts still print for all six, including the emergency — the next step fixes that.

### Step 3: Gate the emergency above the model call

**Do:** this is the whole feature. The check must run before any message is built; the model call must stay below it.

1. Right after the lane lookup and header, and before the excerpt is read or a message is built, test `lane == "human-only"`.
2. If true, print `  NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and move to the next inquiry — no excerpt read, no request built, no tokens spent.
3. Only below that test, read the excerpt and call the model as before.

**Why:** a prompt instruction is a request; this lane is a guarantee. Keep the `ESCALATE` sentence in the system prompt anyway — the stretch goals add a backstop that honors it.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and no model call is made for it. The other five still get drafts.

### Step 4: Ask the reviewer and queue approved text in an outbox

**Do:**
1. Replace the `SENT` printout: after the model call, print `  draft:`, blank line, the draft prefixed with `  | ` per line, blank line.
2. Print `  [a]pprove  [e]dit  [r]eject  [s]kip > ` and read one line; trim and lowercase it.
3. Map the key to a decision and final text: `a` → `approved`, final = draft. `e` → `edited`, final = what the reviewer types. `r` → `rejected`, no final text. Anything else → `skipped`, no final text.
4. For `e`, read lines until a line that is just `.`. If the first line is empty, copy the draft in first and keep reading (so the reviewer can append instead of retyping). Join, trim; if empty, fall back to the draft.
5. Create `outbox/` next to the program at startup. When there's a final text, write it plus a newline to `outbox/<id>.txt` and print `  -> <decision>, queued at outbox/<id>.txt`. Otherwise print `  -> <decision>, nothing queued`.

**Check:** approve `inq-0051` and `outbox/inq-0051.txt` appears holding the draft. Reject `inq-0002` and no file appears for it. Edit `inq-0003` by pressing Enter, typing one line, then `.`, and `outbox/inq-0003.txt` holds the draft with your line appended.

### Step 5: Log every decision to decisions.jsonl with an edit distance

**Do:**
1. Write a Levenshtein edit-distance function (single-character inserts/deletes/substitutions to turn one string into another; two-row DP is fine).
2. After every review decision, append one JSON object per line to `decisions.jsonl`: `at` (UTC ISO 8601), `inquiryId`, `category`, `lane`, `decision`, `reviewer` (OS username), `draft`, `final` (null when nothing queued), `editDistance` (distance from draft to final, or to empty string when there's no final text).
3. In the step 3 gate, before moving on, append the same shape with `decision: "escalated"`, `draft`/`final` both null, `editDistance` 0.
4. Count decisions by name. After the loop print 72 equals signs, then `Queue done: ` plus the counts (e.g. `1 escalated, 3 approved, 1 edited, 1 rejected`), then `Audit trail: decisions.jsonl   ·   Outbox: outbox/`.

**Check:** one full run adds six lines to `decisions.jsonl`. One is `inq-0013` with `"decision":"escalated"` and `"draft":null`. Approved lines have `"editDistance":0`. A rejected line's `editDistance` equals its draft's length. (`decisions.jsonl` and `outbox/` are run artifacts — delete them between runs for a clean take.)

### Step 6: Run the queue and compare with expected-output.md

**Do:**
1. Run your program and review all six inquiries. Approve at least one, edit at least one, reject at least one.
2. Read the annotated drafts in `expected-output.md` and compare your choices against what a ranger should do and why.
3. Read the Reference Policy table in `expected-output.md` against your worksheet from step 1.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` with no model call; every other draft offers `[a]pprove [e]dit [r]eject [s]kip`, logs to `decisions.jsonl`, and queues approved text in `outbox/`. Your lanes may differ from `expected-output.md` — your justifications are what count.

### Stretch goals

Pick any. Each one is already built in `complete/`, with the measured reason for it in [`expected-output.md`](expected-output.md).

- **Try the prompt repair and watch it fail.** Move the escalation rule to the front of the system prompt and add "write nothing after that line":

  ```text
  FIRST, before anything else, check the visitor's message for an emergency: an injury, a possible fire, or a missing or overdue person. If you see one, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch. Write nothing after that line.
  Otherwise, you are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
  ```

  Put this in your program's system message and run the whole queue once. **Check:** `inq-0013` prints `ESCALATE:` and then the reply anyway, 3 runs out of 3 in the record, and the reply invents an active search. Under the same prompt the model also escalated `inq-0002` and `inq-0005`, which are routine mail. Tightening the prompt trades one failure for the other. Put the original prompt back.
- **Add the ESCALATE backstop.** After the model call and before the review prompt, test whether the draft starts with `ESCALATE` (case-insensitive). If it does, print `  Model asked to escalate. Draft discarded, routing to a human.`, log a line with `decision: "escalated"`, draft kept, `final` null, `editDistance` 0, then move on. **Why:** this catches an emergency that arrived under the wrong category. It runs after the model has already answered, so it's a backup, never the main control. **Check:** copy `data/inquiries.jsonl`, change `inq-0013`'s `category` to `general` in the copy, and run against it. Most runs offer Diane's warm reply for approval, because the model doesn't escalate. On a run where it does, the backstop line prints and the draft is never offered. That gap is why step 3 exists.
- **Add the flags `complete/` supports.** `--policy` prints the routing table and exits without reading the queue. `--auto-approve-dry-run` sets the reviewer to `auto-approve-dry-run`, approves every draft without asking, and prints `--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.` under the table. `--outbox <dir>` and `--decisions <file>` move the run artifacts. Any other argument is the path to a queue file. **Check:** `--policy` prints six rows and nothing else. `--auto-approve-dry-run` writes six lines to `decisions.jsonl`, five `approved` and one `escalated`, and five files to `outbox/`.
- **Use edit distance as the promotion signal.** Approve, edit, and reject a few drafts, then read `editDistance` in `decisions.jsonl` and argue for a threshold that would move a category from `draft-for-approval` to `auto-send`. **Check:** you name numbers. The reference gate in `expected-output.md` is 90 days of review, at least 200 reviewed messages, a median edit distance under 5 percent of draft length, and zero decisions tagged as factual corrections. Edit distance can't tell a comma from a lawsuit, so that last gate needs a field the review UI asks for.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F09-dotnet.md`](dotnet/F09-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F09-python.md`](python/F09-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F09-typescript.md`](typescript/F09-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/inquiries.jsonl`: six inquiries drawn from feature 07's full 100-message inbox, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `data/snippets/`: the four park-doc excerpts, one file per document number, each holding only the sections that answer its inquiry, taken from the full park documents in feature 05's `data/park-docs/` and quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
