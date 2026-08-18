# Python Demo for 09 Human-in-the-Loop

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: deliberately unsafe. Every draft goes straight out, nothing is logged, and nothing treats an emergency differently except a sentence in the system prompt that the model is free to ignore.
- `complete/main.py`: the finished demo as shown on stage. A policy table decides the lane by error cost; emergencies never reach the model (the gate sits above the API call, in code); a review loop with approve / edit / reject / skip; every decision appended to `decisions.jsonl` with the draft, the final text, and the edit distance; approved text queued to `outbox/`.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py                            # review the queue interactively
python main.py --policy                   # print the routing policy table and exit
python main.py --auto-approve-dry-run     # non-interactive run for testing
```

The emergency row is the demo's punchline: `inq-0013` never reaches the model at all. The starter, run on the same queue, hands the overdue-hiker message to the model and gets a warm reassuring reply back; the measured 3/3 is in [`../expected-output.md`](../expected-output.md). `decisions.jsonl` and `outbox/` are written next to the script and are gitignored.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F09-lab.md`](../F09-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Read Every Draft, Especially the Sixth (lab step 1)

The starter drafts a reply to all six inquiries and prints each as sent, with no review and no log. Read them as an editor, and then read the one for `inq-0013`, the woman whose husband is four hours overdue. The system prompt tells the model to output `ESCALATE:` for emergencies. Watch what it does instead.

Run:

```bash
python main.py
```

Check: Five usable drafts you would want to touch before sending, and a warm, reassuring, useless note to the overdue hiker's wife. In the recorded runs the model ignored the escalation instruction 3/3, and even with the instruction moved to the top it wrote `ESCALATE` and then the note anyway. One run of the starter makes the case for the whole feature.

### Step 2: Fill in the Policy Worksheet Before Writing More Code (lab step 2)

Open `../policy-worksheet.md`, and for each of feature 07's categories choose auto-send, draft-for-approval, or human-only, and write one sentence of justification based on what a wrong answer costs and whether it can be undone. This is half the lab and it is judgment, not typing. The reference policy in `../expected-output.md` puts every reversible category in draft-for-approval and emergency in human-only.

Check: A completed table with a reason per row. Your lanes may differ from the reference; your justifications are what count.

### Step 3: Put the Policy in Code, Above the API Call

The gate. A lookup from category to lane, defaulting to human-only for anything unknown, and a check that runs before any request is built. A human-only message is escalated and logged without a single token being spent on it. This is the difference between a prompt instruction (a request) and a policy lane (a guarantee).

```python
POLICY = {"trail-condition": "draft-for-approval", "permit": "draft-for-approval",
          "complaint": "draft-for-approval", "general": "draft-for-approval",
          "lost-and-found": "draft-for-approval", "emergency": "human-only"}
lane = POLICY.get(inquiry["category"], "human-only")
if lane == "human-only":
    print("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.")
    continue                      # the model is never called
# ... only now build the messages and call the model
```

Run:

```bash
python main.py
```

Check: `inq-0013` prints `NO DRAFT` and no model call is made for it. Keep the `ESCALATE` prefix check after the call too, as a backstop for emergencies that arrive miscategorized; it is never the control, because it runs after the model has had its say.

### Step 4: Add the Review Loop and the Audit Trail (lab step 3)

Instead of printing "SENT", show the draft and ask: approve, edit, reject, or skip. Approved and edited text goes to an outbox; every decision, including the escalations, appends a line to `decisions.jsonl` with the draft, the final text, the reviewer, and the lane. `complete/` also records the edit distance between draft and final, which is the number that later tells you whether a lane has earned promotion.

```python
key = input("  [a]pprove  [e]dit  [r]eject  [s]kip > ").strip().lower()
final = None
if key == "a":
    decision, final = "approved", draft
elif key == "e":
    decision, final = "edited", read_edited(draft)
elif key == "r":
    decision = "rejected"
else:
    decision = "skipped"
if final is not None:
    (outbox_dir / f"{inquiry['id']}.txt").write_text(final + "\n")
with decisions_path.open("a") as f:
    f.write(json.dumps({"at": datetime.now(timezone.utc).isoformat(), "inquiryId": inquiry["id"],
                        "category": inquiry["category"], "lane": lane, "decision": decision,
                        "reviewer": getpass.getuser(), "draft": draft, "final": final,
                        "editDistance": edit_distance(draft, final or "")}) + "\n")
```

Run:

```bash
python main.py
python main.py --auto-approve-dry-run   # non-interactive, for a quick check
```

Check: Approve one, edit one, reject one, then read `decisions.jsonl`: six lines, one of them the escalation with no draft. Compare `../expected-output.md`. Stretch: compute the edit distance between draft and your edited version and sketch what threshold would earn a category promotion from draft-mode to auto-send.
