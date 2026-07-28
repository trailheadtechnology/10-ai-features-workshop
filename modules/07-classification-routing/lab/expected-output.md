# What passing looks like

Labels vary a little run to run; the checks below are what has to be true. Everything here came from an actual `llama3.2` run of the taxonomy in `ollama.http` over all 20 messages in `inquiries-slice.jsonl`, at temperature 0, scored against `reference-labels.json`.

## The scoreboard from that run

| metric | result |
|---|---|
| overall accuracy | 16/20 (80%) |
| emergency recall | 2/2 |
| ambiguous message (inq-0035) | `conditions`, reference says `permit` |

**Emergency recall is the metric that matters.** Both `inq-0013` (overdue husband on the Highline) and `inq-0041` (injured ankle mid-descent on the Beehive) came back `emergency`, with no false emergencies anywhere else in the slice. A run that scores 19/20 but files one of those two under `conditions` fails this lab. Overall accuracy is the number you report to your manager; emergency recall is the number that keeps someone alive.

## The full labeling

Correct on the first pass, with the starter taxonomy:

```
inq-0001  permit           OK
inq-0002  conditions       OK
inq-0005  conditions       OK
inq-0006  lost-and-found   OK
inq-0008  general          MISS (reference: permit)
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
inq-0035  conditions       MISS (reference: permit)
inq-0037  conditions       OK
inq-0041  emergency        OK
inq-0051  conditions       MISS (reference: general)
```

## The four misses, and which ones your prompt should fix

- **inq-0008** ("how far in advance do half dome permits open, are there daily ones?") and **inq-0030** (wedding photographer asking whether a special use permit is required). Both landed in `general`. The starter description of `permit` talks about reserving and changing a permit you are trying to get, so the model reads "questions *about* permit rules" as trip planning. This is the planted misclassification from the demo, and it is fixable in the taxonomy: extend the `permit` description to cover questions about permit availability, eligibility, and rules, not just transactions on an existing reservation. Rewording that one line, not the code, is the point of step 2.
- **inq-0051** (campfires allowed near Sperry Chalet?) landed in `conditions`. Defensible: it is a question about what is allowed up there right now. The clean fix is to sharpen `conditions` to mean the physical passability of a trail, and let `general` own rules and regulations. The .NET `complete/` app, which uses the same taxonomy through a JSON-schema enum, labeled this one `general` and scored 17/20; that difference is sampling noise on a borderline message, not a different prompt.
- **inq-0035** is the ambiguous one and it is not really a miss.

## What it did with the ambiguous message

`inq-0035` is Priya, holding a backcountry permit for a night at Avalanche Lake, asking whether her itinerary still works with the bridge out and whether she can swap the night or get a refund. The model called it `conditions`. The reference label is `permit`, because the action she needs is a permit change, and only the permits office can grant it. Both readings are defensible, which is the whole reason the message is in the slice.

The right outcome is not to argue the model into `permit`. It is to notice that a confident single label is the wrong shape of answer here, and add the `unsure` route from the module README's step 5: when the message asks for two different queues to act, send it to a human. If your run puts inq-0035 in `unsure`, that is a pass, not a miss. If it puts inq-0035 confidently in `conditions` and that queue has no authority to refund anything, Priya waits until Thursday and then leaves without an answer.

## Success checks

1. Both emergencies labeled `emergency`. Non-negotiable.
2. No routine message labeled `emergency` more than occasionally. A few false alarms are the price of check 1; a triage queue that cries wolf on lost sunglasses gets ignored.
3. Overall accuracy somewhere in the 16 to 18 out of 20 range after your taxonomy edits. If you are at 20/20, check whether you overfit the descriptions to these exact 20 messages.
4. inq-0035 either in `unsure` or with a note explaining why the label you gave it is defensible.

## Stretch goal

Add a `priority` field alongside `category`, or a confidence score with anything under threshold routed to `unsure`. Watch what happens to inq-0006 (lost daypack with a child's inhaler in it): it is genuinely `lost-and-found`, but it is not a normal lost-and-found. Priority is a second axis, and conflating it with category is how the inhaler ends up behind the wedding ring.
