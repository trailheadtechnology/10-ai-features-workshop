# 07 · Classification & routing

Module 3: Deciding · Runs on Ollama (`llama3.2`)

## The user problem

Every message to Trailhead Guides lands in one inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, and, occasionally, someone reporting an actual emergency. A ranger triages the pile by hand once or twice a day. The permit request waits behind the granola questions, the complaint goes to the wrong person twice, and the emergency sits unread for four hours. The users' real problem is that their message goes into a hole, and how fast it comes out depends on luck.

## The concept

This is classification again (feature 03 was the warm-up), but now the label has consequences: it decides where the message goes and how fast. An LLM makes a solid zero-shot classifier. You describe the categories in plain language and it labels messages with no training data, which is exactly the situation most teams are in on day one.

Two design decisions carry the feature. First, the taxonomy is the product. Category names and one-sentence descriptions in the prompt are where accuracy lives, and when the model misfiles something, you usually fix the description, not the model. Second, errors are not symmetric. Misrouting a complaint costs a day of annoyance; misrouting an emergency is a headline. So the system needs an "unsure" route to a human, and it should be tuned so the expensive class never slips through, even at the cost of extra false alarms. Recall on the class that matters, not overall accuracy, is the number to watch. A local model does this fine, and at inbox volume, free matters.

## Demo outline (about 12 min, .NET)

1. Scroll the raw inbox from `data/inquiries.jsonl`, mixed and unlabeled, and let the room spot the emergency buried in it.
2. Starter: one classify method with the taxonomy written as plain-language category descriptions in the prompt, returning exactly one label.
3. Run the inbox through it and print each message under its routed queue. The payoff: the pile becomes a set of tidy queues in seconds, and the emergency is at the top of its own.
4. Show a misclassification (inq-0030 is planted). Fix it by editing the category description, not the code, and re-run to show the fix landing.
5. Walk the "unsure" route in the taxonomy: it is deliberately narrow, for messages that two queues must both act on. Show inq-0035 landing there, which is correct behavior, and point out that the description forbids sending anything dangerous to it.
6. Close on asymmetric error costs: what this system is tuned never to miss, and what noise level that tolerance buys.

## Lab spec (Core lab, any language)

*Everyone does this one. It is the Core lab for [Module 3](../M3-overview.md), and the hands-on period runs about 45 minutes, so there is room to do it properly rather than fast.*

- **Goal:** classify visitor inquiries into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and route them.
- **Input:** `lab/` provides 20 inquiries from `data/inquiries.jsonl` (including 2 emergencies and at least one deliberately ambiguous message) plus reference labels.
- **How:** POST to Ollama's chat endpoint (`llama3.2`). `lab/ollama.http` has the request with a starter taxonomy prompt.
- **Steps:**
  1. Classify all 20 with the starter taxonomy and score against the reference labels.
  2. Find your misclassifications and fix them by rewording the category descriptions.
  3. Success check: both emergencies classified as `emergency`, and the ambiguous message in `unsure` rather than confidently wrong (see `lab/expected-output.md`). Missing an emergency fails the lab even at 19/20 accuracy; that's the lesson.
- **Stretch goal:** add a `priority` field alongside the category, or return a confidence score and route anything below a threshold to `unsure`.

## Leadership beat

- **When to reach for this:** any shared inbox, ticket queue, or intake form where a human sorts before anyone acts. Support, sales leads, HR requests, incoming documents.
- **Rough cost & effort:** days. No training data needed to start; the ongoing work is refining the taxonomy as real traffic reveals edge cases. Local models make per-message cost roughly zero.
- **The one-liner for your CTO:** "Every message reaches the right person in seconds, and the urgent ones stop waiting in line."

This card is row 7 of the [decision framework](../../../docs/decision-framework.md).
