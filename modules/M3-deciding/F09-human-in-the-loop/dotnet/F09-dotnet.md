# .NET demo for 09 Human-in-the-Loop

Two console projects, both named `Ranger`, both built on Microsoft.Extensions.AI over OllamaSharp. Both read the routed queue in [`../lab/inquiries.jsonl`](../lab/inquiries.jsonl) and use the same drafting prompt, so the only thing that differs is where the human sits.

- `starter/`: the demo's starting point, and the anti-pattern. It drafts a reply to every inquiry and prints it as sent. No review, no record, and the emergency goes out with everything else. Run it once and read what it mailed to the woman whose husband is overdue on the Highline.
- `complete/`: the finished demo. A review loop, a policy table, an outbox, and an audit trail.

Both run against Ollama (`llama3.2`), matching the demo outline in [../F09-spec.md](../F09-spec.md).

## starter

```bash
cd starter
dotnet run                              # drafts and "sends" all six, unreviewed
dotnet run -- ../../lab/inquiries.jsonl # any queue file works
```

## complete

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

The emergency row is the demo's punchline. `inq-0013` never reaches the model at all: its category maps to the `human-only` lane, and the loop skips the API call in code and logs an escalation. The prompt also asks the model to emit `ESCALATE:` instead of drafting, and the app honors that prefix if it sees it, but that is a backstop and not the control. `lab/expected-output.md` shows what happened when the prompt was the only thing standing between an overdue-hiker voicemail and a reassuring auto-reply.

`outbox/` and `decisions.jsonl` are run artifacts and are gitignored. Delete them between takes for a clean demo.
