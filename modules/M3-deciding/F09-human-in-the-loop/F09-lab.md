# Lab 09: Human-in-the-Loop

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** put a human between the model and the visitor. The model drafts, a ranger approves, edits, or rejects, every decision is logged, and emergencies never reach the model at all.
- **Input:** `data/inquiries.jsonl`, six inquiries routed by feature 07, with category and park doc; `data/snippets/`, the four excerpts they cite (`glac-bc-2025-04.md`, `glac-cl-2026-01.md`, `yose-cl-2026-01.md`, `zion-nar-2026-01.md`); `policy-worksheet.md`, the lane table you fill in.
- **How:** the starter already loops over the six inquiries. It sends each one to the chat model with a drafting prompt. You add five things: a policy table, a gate above the model call, a review prompt, an outbox, and an audit log. `http/ollama.http` holds five requests: three everyday drafts, the emergency `inq-0013` (4), and `inq-0013` again with the escalation rule moved first (5).
- **Model:** `llama3.2`, local. The HTTP requests set `temperature: 0.2`; the code tracks use the client default. No key.

Every step below is one thing to make the program do. Each code track's `starter/` drafts and "sends" every inquiry with no review. Edit it until it does all seven steps. Compare against `complete/` when stuck. HTTP-track readers run the numbered requests in `http/ollama.http` where a step names one, and fill in the policy worksheet. The code steps are what `complete/` adds on top of those requests.

### Step 0: Run the starter and read every draft as an editor

Run `starter/` as it is (`dotnet run`, `uv run main.py`, or `npm run starter`; your track's walkthrough says which folder to run it from). Read the code first. It is the loop you are about to change:

1. It opens `../../data/inquiries.jsonl`. Every line is one JSON object with six string fields: `id`, `channel`, `received`, `category`, `doc`, `text`. These are six of the 100 messages in feature 07's `data/inquiries.jsonl`, with the same `id`, `channel`, `received`, and `text`, plus two fields feature 07's output would supply: `category`, the label feature 07 assigns (feature 07's `conditions` is spelled `trail-condition` here, and its seventh label, `unsure`, does not appear because none of these six is ambiguous), and `doc`, the file name of the park-doc excerpt in `data/snippets/` that answers the message, or an empty string when there is none (`inq-0007`, the complaint, and `inq-0013`, the emergency). No script builds this file; it ships with the workshop as the routed queue.
2. It reads the file line by line, skips blank lines, and parses each line into those fields.
3. For each inquiry, if `doc` is not empty and `data/snippets/<doc>` exists, it reads that file and trims it. Otherwise the excerpt is the literal text `(none on file for this message)`. The four snippet files are named after the park document numbers they quote (`glac-bc-2025-04.md` is GLAC-BC-2025-04, and so on). Each one is a heading line naming the document, then only the numbered sections that answer the inquiry pointing at it, copied word for word from the full documents in feature 05's corpus at `modules/M2-finding/F05-rag/data/park-docs/` (for example `glacier-backcountry-camping-guide.md`). No script produces them; they ship with the workshop, so the model sees the source section and nothing else.
4. It builds two chat messages. The system message is this prompt, the same for every inquiry:

```text
You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers. Exception: if the visitor's message reports an emergency, an injury, a possible fire, or a missing or overdue person, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch.
```

5. The user message is this template, with the excerpt, `channel`, `received`, and `text` filled in:

```text
Reference excerpt:
<the excerpt>

Visitor message (<channel>, received <received>):
<text>

Draft the reply.
```

6. It sends both messages to `llama3.2` through your track's chat client and trims the reply.
7. It prints `=== SENT to visitor · <id> (<category>) ===`, then the draft, then a blank line. After the loop it prints `All replies sent. Nobody read them. Nothing was logged.`

HTTP-track readers: `http/ollama.http` is five chat requests to the local model, each carrying the same system prompt as the code and the excerpt and visitor message already pasted into the user message, at `temperature: 0.2`. It exists so the track with no code can send the same calls the loop above makes. Requests 1 to 4 are four of these six calls. Send them in order. Request 1, `inq-0002`, user:

```text
Reference excerpt:
Excerpt from YOSE-CL-2026-01, Yosemite National Park, Seasonal Closures and Access Notice, 2026 (last revised June 15, 2026):

Section 4.1: Mist Trail (Vernal Fall corridor): the winter route closure was lifted April 10, 2026. The trail closes each winter when ice accumulates on the steps; the John Muir Trail serves as the winter route.

Visitor message (web-form, received 2026-06-01T09:47:00Z):
is the mist trail open yet?? going to yosemite june 12

Draft the reply.
```

Request 2, `inq-0051`, user:

```text
Reference excerpt:
Excerpt from GLAC-BC-2025-04, Glacier National Park, Backcountry Camping Guide (revised January 15, 2026):

Section 4.1: Where campfires are authorized, they are permitted only in Park-installed metal fire rings at backcountry campgrounds specifically listed as "fires permitted" on the current backcountry map, and only when the posted fire danger rating is below Very High.

Section 4.2: A number of backcountry campgrounds are designated no-wood-fire sites due to elevation, fuel scarcity, or resource sensitivity. These designations do not vary with season or fire danger rating. In particular, wood fires are prohibited year-round at all campsites in the Sperry Chalet area (site code SPE), regardless of season or posted fire danger; only pressurized-gas stoves are permitted for cooking at these sites. This prohibition applies equally in September and during the shoulder seasons, and is not lifted when fire danger is Low.

Visitor message (email, received 2026-06-19T18:12:00Z):
Hi, we're staying overnight near Sperry Chalet in early September. Are campfires allowed up there or is it stoves only? I've gotten different answers from two different Facebook groups and would rather hear it from the source. Thanks!

Draft the reply.
```

Request 3, `inq-0005`, user:

```text
Reference excerpt:
Excerpt from GLAC-CL-2026-01, Glacier National Park, Seasonal Closures and Trail Status, 2026 Season (last revised June 24, 2026):

Section 4.1: Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice. The footbridge over Avalanche Creek approximately 0.4 miles above the Trail of the Cedars junction washed out during high runoff in mid-June 2026. The trail is closed from the Trail of the Cedars junction to Avalanche Lake in both directions. The Trail of the Cedars loop itself remains open. Engineering assessment of the bridge abutments is underway; a replacement schedule has not been established. Visitors holding backcountry itineraries that transit this segment should contact the Backcountry Office for re-routing.

Section 6.4: Day access to Avalanche Lake is not possible until the bridge is replaced.

Visitor message (email, received 2026-06-18T08:30:00Z):
Good morning, we heard from another hiker that the bridge on the Avalanche Lake Trail washed out last week. Is that true? We have a family trip planned for June 27 and my mother uses trekking poles, she cannot ford a creek. Is there an alternate route to the lake or should we pick a different hike?

Draft the reply.
```

Request 4, `inq-0013`, the emergency, user:

```text
Reference excerpt:
(none on file for this message)

Visitor message (voicemail-transcript, received 2026-06-21T21:47:00Z):
hi um im calling because my husband went out this morning to do the highline trail in glacier he said hed be back by six and its almost ten now and his phone goes straight to voicemail he always calls when hes running late always im sure theres an explanation but i dont know who else to call his name is robert ferris hes 61 wearing a green jacket please call me back this is his wife diane at four oh six five five five oh one one eight

Draft the reply.
```

Now read every draft the way a ranger would. Each one is going out under your name.

**Check:** `inq-0002` says the Mist Trail is closed when the excerpt says it reopened (reject). `inq-0051` and `inq-0005` are accurate and cited (approve). `inq-0003` gets both rules right but pins the flash flood rule on `GLAC-BC-2026-01`, a Glacier document number, on a Zion question (edit). `inq-0007` apologizes and decides nothing (edit). `inq-0013` gets a warm reply to Diane with no `ESCALATE` line, even though the prompt told the model not to draft one; in the recorded runs that happened 3 times out of 3. The rest of the lab exists because of that last draft.

### Step 1: Fill in policy-worksheet.md

No code in this step. Open `policy-worksheet.md`. It is a blank Markdown table with one row per feature 07 category (`permit`, `conditions`, `complaint`, `lost-and-found`, `general`, `emergency`; feature 07's `unsure` has no row) and five columns: Category, Lane, Worst plausible error, Reversible?, Justification. Two follow-up questions sit under the table. The worksheet's `conditions` row is the code's `trail-condition` category.

1. For each row, pick a lane: `auto-send`, `draft-for-approval`, or `human-only`.
2. Write the worst plausible wrong reply for that category.
3. Write whether that error can be taken back after it is sent.
4. Write a one-sentence justification.
5. Answer the two follow-up questions under the table. The second one asks where your lane lives in the code. Steps 2 and 3 are the answer.

**Check:** every row says what a wrong reply costs, and emergency is `human-only`. If your reason for a lane is "the prompt tells it to", reread the `inq-0013` draft from step 0. Your lanes may differ from the reference in `expected-output.md`; your justifications are what count.

### Step 2: Add the policy table and print the lane

1. Above the loop, make a dictionary from category to lane with these six entries:

```text
trail-condition   draft-for-approval
permit            draft-for-approval
complaint         draft-for-approval
general           draft-for-approval
lost-and-found    draft-for-approval
emergency         human-only
```

2. Print the table once at the top of the run, under the heading `Routing policy (error cost decides the lane):`, one line per category.
3. Inside the loop, right after parsing the inquiry, look up its `category` in the dictionary. If the category is not in the dictionary, use `human-only`. That means an unknown category fails closed: it gets the safest lane, not a draft.
4. Replace the `SENT` header with a header that shows the lane: a line of 72 dashes, then `<id>  ·  <category>  ·  <channel>  ·  lane: <lane>`, then another line of 72 dashes, then the visitor's `text` with every line prefixed by `  | `, then a blank line.

**Check:** the run opens with the six-row table. Every inquiry header ends in its lane. `inq-0013` shows `lane: human-only` and the other five show `lane: draft-for-approval`. The drafts still print for all six, including the emergency. The next step fixes that.

### Step 3: Gate the emergency above the model call

This is the whole feature. The check must run before any message is built. The model call must stay below it.

1. Right after the lane lookup and header, and before the excerpt is read or a message is built, test `lane == "human-only"`.
2. If it is, print `  NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and `continue` to the next inquiry. Nothing else happens for it. No excerpt is read. No request is built. No tokens are spent.
3. Only below that test, read the excerpt and call the model as before.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and no model call is made for it. The other five still get drafts. A prompt instruction is a request; this lane is a guarantee. Keep the `ESCALATE` sentence in the system prompt anyway; the stretch goals add a backstop that honors it.

### Step 4: Ask the reviewer and queue approved text in an outbox

1. Replace the `SENT` printout. After the model call, print `  draft:`, a blank line, the draft with every line prefixed by `  | `, and a blank line.
2. Print the prompt `  [a]pprove  [e]dit  [r]eject  [s]kip > ` and read one line from the keyboard. Trim it and lowercase it.
3. Map the key to a decision and a final text. `a` means decision `approved` and final is the draft. `e` means decision `edited` and final is what the reviewer types. `r` means decision `rejected` with no final text. Anything else means decision `skipped` with no final text.
4. For `e`, read lines from the keyboard until a line that is just `.`. If the first line is empty, copy the draft in and keep reading. That lets the reviewer add to the draft instead of retyping it. Join the lines and trim. If the result is empty, fall back to the draft.
5. Create an `outbox/` folder next to the program at startup. When there is a final text, write it plus a newline to `outbox/<id>.txt` and print `  -> <decision>, queued at outbox/<id>.txt`. Otherwise print `  -> <decision>, nothing queued`.

**Check:** approve `inq-0051` and `outbox/inq-0051.txt` appears holding the draft. Reject `inq-0002` and no file appears for it. Edit `inq-0003` by pressing Enter, typing one line, then `.`, and `outbox/inq-0003.txt` holds the draft with your line appended.

### Step 5: Log every decision to decisions.jsonl with an edit distance

1. Write a Levenshtein edit distance function. It takes two strings. It returns how many single-character inserts, deletes, and substitutions turn one into the other. Use the two-row dynamic programming version. The first row starts as `0..len(b)`. Each new cell is the smallest of three values: the cell to the left plus 1, the cell above plus 1, and the diagonal cell plus 0 if the two characters match or plus 1 if they do not.
2. After every review decision, append one JSON object as one line to `decisions.jsonl` next to the program. It has these nine fields: `at` (UTC timestamp, ISO 8601), `inquiryId`, `category`, `lane`, `decision`, `reviewer` (the OS username), `draft`, `final` (null when nothing was queued), and `editDistance` (distance from the draft to the final text, or to the empty string when there is no final text).
3. In the step 3 gate, before `continue`, append the same shape with `decision` set to `escalated`, `draft` and `final` both null, and `editDistance` 0.
4. Count decisions by name as you go. After the loop print a line of 72 equals signs, then `Queue done: ` followed by the counts, for example `1 escalated, 3 approved, 1 edited, 1 rejected`, then `Audit trail: decisions.jsonl   ·   Outbox: outbox/`.

**Check:** one full run adds six lines to `decisions.jsonl`. One of them is `inq-0013` with `"decision":"escalated"` and `"draft":null`. Approved lines have `"editDistance":0`. A rejected line's `editDistance` equals the length of its draft. The file and `outbox/` are run artifacts; delete them between runs for a clean take.

### Step 6: Run the queue and compare with expected-output.md

1. Run your program and review all six inquiries. Approve at least one, edit at least one, and reject at least one.
2. Read the annotated drafts in `expected-output.md`. For each of the five everyday drafts, compare what you chose with what the annotation says a ranger should do and why.
3. Read the Reference Policy table in `expected-output.md` and compare it with your worksheet from step 1.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` with no model call; every other draft offers `[a]pprove [e]dit [r]eject [s]kip`, logs to `decisions.jsonl`, and queues approved text in `outbox/`. Your lanes may differ from `expected-output.md`; your justifications are what count.

### Stretch goals

Pick any. Each one is already built in `complete/` or in `http/ollama.http`, with the measured reason for it in [`expected-output.md`](expected-output.md).

- **Try the prompt repair and watch it fail.** Send request 5 in `http/ollama.http`. It is the `inq-0013` user message byte for byte, with the escalation rule moved to the front of the system prompt and "write nothing after that line" added:

  ```text
  FIRST, before anything else, check the visitor's message for an emergency: an injury, a possible fire, or a missing or overdue person. If you see one, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch. Write nothing after that line.
  Otherwise, you are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
  ```

  Then put this prompt in your program's system message and run the whole queue once. **Check:** request 5 prints `ESCALATE:` for `inq-0013` and then the reply anyway, 3 runs out of 3 in the record, and the reply invents an active search. Under the same prompt the model also escalated `inq-0002` and `inq-0005`, which are routine mail. Tightening the prompt trades one failure for the other. Put the original prompt back.
- **Add the ESCALATE backstop.** After the model call and before the review prompt, test whether the draft starts with `ESCALATE`, ignoring case. If it does, print `  Model asked to escalate. Draft discarded, routing to a human.`, log a line with `decision` set to `escalated`, the draft kept, `final` null, and `editDistance` 0, then `continue`. This catches an emergency that arrived under the wrong category. It runs after the model has already answered, so it is a backup and never the main control. **Check:** copy `data/inquiries.jsonl`, change the `category` of `inq-0013` in the copy to `general`, and run against the copy. Most runs offer Diane's warm reply for approval, because the model does not escalate. On a run where it does, the backstop line prints and the draft is never offered. That gap is why step 3 exists.
- **Add the flags `complete/` supports.** `--policy` prints the routing table and exits without reading the queue. `--auto-approve-dry-run` sets the reviewer to `auto-approve-dry-run`, approves every draft without asking, and prints `--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.` under the table. `--outbox <dir>` and `--decisions <file>` move the run artifacts. Any other argument is the path to a queue file. **Check:** `--policy` prints six rows and nothing else. `--auto-approve-dry-run` writes six lines to `decisions.jsonl`, five `approved` and one `escalated`, and five files to `outbox/`.
- **Use edit distance as the promotion signal.** Approve, edit, and reject a few drafts. Then read `editDistance` in `decisions.jsonl` and argue for a threshold that would move a category from `draft-for-approval` to `auto-send`. **Check:** you name numbers. The reference gate in `expected-output.md` is 90 days of review, at least 200 reviewed messages, a median edit distance under 5 percent of draft length, and zero decisions tagged as factual corrections. Edit distance cannot tell a comma from a lawsuit, so that last gate needs a field the review UI asks for.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F09-http.md`](http/F09-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F09-dotnet.md`](dotnet/F09-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F09-python.md`](python/F09-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F09-typescript.md`](typescript/F09-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/inquiries.jsonl`: six inquiries drawn from feature 07's full 100-message inbox, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `data/snippets/`: the four park-doc excerpts, one file per document number, each holding only the sections that answer its inquiry, taken from the full park documents in feature 05's `data/park-docs/` and quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
