# What passing looks like

Labels vary a little run to run; the checks below are what has to be true. Everything here came from an actual `llama3.2` run of the taxonomy in `ollama.http` over all 20 messages in `inquiries-slice.jsonl`, at temperature 0, scored against `reference-labels.json`.

## The scoreboard from that run

| metric | result |
|---|---|
| overall accuracy | 18/20 (90%) |
| emergency recall | 2/2 |
| false emergencies | 0 |
| ambiguous message (inq-0035) | `unsure`, which is the reference label |

**Emergency recall is the metric that matters.** Both `inq-0013` (overdue husband on the Highline) and `inq-0041` (injured ankle mid-descent on the Beehive) came back `emergency`, with no false emergencies anywhere else in the slice. Adding the `unsure` route did not soften either one: the taxonomy tells the model to settle the danger question first and never answer `unsure` for a message where someone might be hurt. A run that scores 19/20 but files one of those two under `conditions`, or parks one in `unsure` for a human to find later, fails this lab. Overall accuracy is the number you report to your manager; emergency recall is the number that keeps someone alive.

## The full labeling

```
inq-0001  permit           OK
inq-0002  conditions       OK
inq-0005  conditions       OK
inq-0006  lost-and-found   OK
inq-0008  permit           OK
inq-0010  complaint        OK
inq-0012  lost-and-found   OK
inq-0013  emergency        OK
inq-0015  permit           OK
inq-0016  complaint        OK
inq-0020  lost-and-found   OK
inq-0021  conditions       OK
inq-0025  conditions       OK
inq-0026  permit           OK
inq-0029  general          OK
inq-0030  general          MISS (reference: permit)
inq-0035  unsure           OK
inq-0037  conditions       OK
inq-0041  emergency        OK
inq-0051  conditions       MISS (reference: general)
```

## The two misses, and what your prompt should do about them

- **inq-0030** (wedding photographer asking whether a special use permit is required) landed in `general`. The `permit` description talks about reserving, changing, and paying for a permit you already want, so the model reads "questions *about* whether a permit is required" as trip planning. This is the planted misclassification from the demo, and it is fixable in the taxonomy, not the code: extend the `permit` description to cover questions about permit availability, eligibility, and rules, not just transactions on an existing reservation. Rewording that one line is the point of step 2. Watch `inq-0008` (half dome lottery) while you do it: it lands in `permit` correctly with the shipped taxonomy, and a clumsy rewrite can knock it back out.
- **inq-0051** (campfires allowed near Sperry Chalet?) landed in `conditions`. Defensible: it is a question about what is allowed up there right now. The clean fix is to sharpen `conditions` to mean the physical passability of a trail, and let `general` own rules and regulations.

## What it did with the ambiguous message

`inq-0035` is Priya, holding a backcountry permit for a night at Avalanche Lake, asking whether her itinerary still works with the bridge out and whether she can swap the night or get a refund. The model called it `unsure`, which is the reference label and the right answer.

The point is not to argue the model into `permit` or into `conditions`. Either one is half right and half useless: the trail info desk cannot refund anything, and the permits office does not decide whether a washed-out bridge is crossable. Two queues have to act, so no single queue owns the message, and the honest output is to hand it to a human. If your run puts inq-0035 confidently in `conditions`, Priya waits until Thursday and then leaves without an answer.

`unsure` only earns its keep if it stays narrow. The taxonomy defines it as the two-queue case and says outright that it is not a catch-all for anything hard, because an `unsure` bucket that fills up with ordinary permit questions is just the original unsorted inbox with extra steps. Watch that queue when you edit descriptions: exactly one message should be in it.

## The .NET `complete/` app on the same taxonomy

`dotnet/complete` runs the same category descriptions through a JSON-schema enum instead of a hand-copied HTTP request. It scored **17/20 with emergency recall 2/2**, and it also put inq-0035 in `unsure`. Its extra miss is `inq-0001` (backcountry permit booking, website session keeps expiring), which it routed to `unsure` rather than `permit`. That is one over-cautious call on the safe side of the ledger, and it is the difference between two request shapes rather than a different prompt. It is also a fair warning: the `unsure` route needs watching, and the fix when it drifts is a sharper description of the queue that should have owned the message.

## Success checks

1. Both emergencies labeled `emergency`. Non-negotiable. Neither one in `unsure`.
2. No routine message labeled `emergency` more than occasionally. A few false alarms are the price of check 1; a triage queue that cries wolf on lost sunglasses gets ignored.
3. inq-0035 in `unsure`.
4. At most one or two other messages in `unsure`. If a third of the slice lands there, the description has become a dumping ground and you have rebuilt the unsorted inbox.
5. Overall accuracy somewhere in the 17 to 19 out of 20 range after your taxonomy edits. If you are at 20/20, check whether you overfit the descriptions to these exact 20 messages.

## Stretch goal

Add a `priority` field alongside `category`, or a confidence score with anything under threshold routed to `unsure`. Watch what happens to inq-0006 (lost daypack with a child's inhaler in it): it is genuinely `lost-and-found`, but it is not a normal lost-and-found. Priority is a second axis, and conflating it with category is how the inhaler ends up behind the wedding ring.
