# .NET Demo for 09 Human-in-the-Loop

Two console projects, both named `Ranger`, both built on Microsoft.Extensions.AI over OllamaSharp. Both read the routed queue in [`../data/inquiries.jsonl`](../data/inquiries.jsonl) and use the same drafting prompt, so the only thing that differs is where the human sits.

- `starter/`: the demo's starting point, and the anti-pattern. It drafts a reply to every inquiry and prints it as sent. No review, no record, and the emergency goes out with everything else. Run it once and read what it mailed to the woman whose husband is overdue on the Highline.
- `complete/`: the finished demo. A review loop, a policy table, an outbox, and an audit trail.

Both run against Ollama (`llama3.2`), matching the demo script in [docs/slides/outlines/M3-deciding.md](../../../../docs/slides/outlines/M3-deciding.md).

## Starter

```bash
cd starter
dotnet run                              # drafts and "sends" all six, unreviewed
dotnet run -- ../../data/inquiries.jsonl # any queue file works
```

## Complete

```bash
cd complete
dotnet run                              # review the queue one message at a time
dotnet run -- --policy                  # print the routing policy table and exit
dotnet run -- --auto-approve-dry-run    # non-interactive run, for testing
dotnet run -- --outbox /tmp/out --decisions /tmp/decisions.jsonl
```

For each inquiry the app prints the visitor's message, the lane its category falls into, and the draft, then asks:

```
  [a]pprove  [e]dit  [r]eject  [s]kip >
```

Approve writes the draft to `outbox/<id>.txt`. Edit takes replacement text (press Enter on the first line to start from the draft, finish with a single `.` on its own line) and writes that instead. Reject and skip queue nothing. Every decision appends a line to `decisions.jsonl` with the draft, the final text, the reviewer, the lane, and the Levenshtein distance between draft and final, which is the number that tells you later whether a lane has earned a promotion.

The emergency row is the demo's punchline. `inq-0013` never reaches the model at all: its category maps to the `human-only` lane, and the loop skips the API call in code and logs an escalation. The prompt also asks the model to emit `ESCALATE:` instead of drafting, and the app honors that prefix if it sees it, but that is a backstop and not the control. `expected-output.md` shows what happened when the prompt was the only thing standing between an overdue-hiker voicemail and a reassuring auto-reply.

`outbox/` and `decisions.jsonl` are run artifacts and are gitignored. Delete them between takes for a clean demo.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F09-lab.md`](../F09-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Read Every Draft, Especially the Sixth (lab step 1)

The starter drafts a reply to all six inquiries and prints each as sent, with no review and no log. Read them as an editor, and then read the one for `inq-0013`, the woman whose husband is four hours overdue. The system prompt tells the model to output `ESCALATE:` for emergencies. Watch what it does instead.

Run:

```bash
dotnet run
```

Check: Five usable drafts you would want to touch before sending, and a warm, reassuring, useless note to the overdue hiker's wife. In the recorded runs the model ignored the escalation instruction 3/3, and even with the instruction moved to the top it wrote `ESCALATE` and then the note anyway. One run of the starter makes the case for the whole feature.

### Step 2: Fill in the Policy Worksheet Before Writing More Code (lab step 2)

Open `../policy-worksheet.md`, and for each of feature 07's categories choose auto-send, draft-for-approval, or human-only, and write one sentence of justification based on what a wrong answer costs and whether it can be undone. This is half the lab and it is judgment, not typing. The reference policy in `../expected-output.md` puts every reversible category in draft-for-approval and emergency in human-only.

Check: A completed table with a reason per row. Your lanes may differ from the reference; your justifications are what count.

### Step 3: Put the Policy in Code, Above the API Call

The gate. A lookup from category to lane, defaulting to human-only for anything unknown, and a check that runs before any request is built. A human-only message is escalated and logged without a single token being spent on it. This is the difference between a prompt instruction (a request) and a policy lane (a guarantee).

```csharp
var policy = new Dictionary<string, string>
{
    ["trail-condition"] = "draft-for-approval", ["permit"] = "draft-for-approval",
    ["complaint"] = "draft-for-approval", ["general"] = "draft-for-approval",
    ["lost-and-found"] = "draft-for-approval", ["emergency"] = "human-only",
};
var lane = policy.GetValueOrDefault(inquiry.Category, "human-only");
if (lane == "human-only")
{
    Console.WriteLine("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.");
    continue;                     // the model is never called
}
// ... only now build the messages and call the model
```

Run:

```bash
dotnet run
```

Check: `inq-0013` prints `NO DRAFT` and no model call is made for it. Keep the `ESCALATE` prefix check after the call too, as a backstop for emergencies that arrive miscategorized; it is never the control, because it runs after the model has had its say.

### Step 4: Add the Review Loop and the Audit Trail (lab step 3)

Instead of printing "SENT", show the draft and ask: approve, edit, reject, or skip. Approved and edited text goes to an outbox; every decision, including the escalations, appends a line to `decisions.jsonl` with the draft, the final text, the reviewer, and the lane. `complete/` also records the edit distance between draft and final, which is the number that later tells you whether a lane has earned promotion.

```csharp
Console.Write("  [a]pprove  [e]dit  [r]eject  [s]kip > ");
var key = (Console.ReadLine() ?? "s").Trim().ToLowerInvariant();
string decision; string? final = null;
switch (key)
{
    case "a": decision = "approved"; final = draft; break;
    case "e": decision = "edited"; final = ReadEdited(draft); break;
    case "r": decision = "rejected"; break;
    default: decision = "skipped"; break;
}
if (final is not null) await File.WriteAllTextAsync(Path.Combine("outbox", $"{inquiry.Id}.txt"), final);
await File.AppendAllTextAsync("decisions.jsonl",
    JsonSerializer.Serialize(new Decision(DateTimeOffset.UtcNow, inquiry.Id, inquiry.Category, lane, decision,
        Environment.UserName, draft, final, EditDistance(draft, final ?? ""))) + Environment.NewLine);
```

Run:

```bash
dotnet run
dotnet run -- --auto-approve-dry-run   # non-interactive, for a quick check
```

Check: Approve one, edit one, reject one, then read `decisions.jsonl`: six lines, one of them the escalation with no draft. Compare `../expected-output.md`. Stretch: compute the edit distance between draft and your edited version and sketch what threshold would earn a category promotion from draft-mode to auto-send.
