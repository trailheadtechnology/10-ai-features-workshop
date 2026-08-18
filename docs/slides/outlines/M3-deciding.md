# Deck 3: Module 3, Deciding (about 90 minutes)

The first two modules produced answers for a user. This one produces decisions for the people running the product.

This file is the instructor script for the module. The bullets under each slide are what is on screen; the `Notes:` text and the "Demo script" steps under each "Demo: What to Watch" slide are what you say and click. The specs in `modules/` are written for attendees and no longer carry the demo steps.

Runsheet (lengths, not clock times):
- 30 min instructor: slides 1 to 3, then for each feature the problem, concept, demo (about 8 min each), and leadership card. Feature 07 first, then 08, then 09.
- 60 min build: slide "Hands-On" up; attendees pick 07 (recommended), 08, or 09. Sticky check at about 20 min.
- Cut if behind: the "unsure" route in demo 07 (step 5), the `--raw` first pass in demo 08 (step 3, go straight to the prefixed run), and the audit log in demo 09 (step 4).
- Last 5 min: debrief slide.

---

## 1. Module 3: Deciding

- Triage and judgment
- 07 Classification & Routing (Recommended) · 08 Anomaly Detection · 09 Human-in-the-Loop
- The user here is usually a colleague: a ranger, an ops lead, a support manager

Notes: Which queue does this go in, which of these 500 reports deserves attention, what is the software allowed to send without a human reading it first.

## 2. The Thread to Watch

- Errors here are not symmetric
- Misrouting a complaint costs somebody a day. Misrouting the overdue-hiker message is a headline.
- Measure recall on the class that matters, not overall accuracy
- Add an `unsure` route so an ambiguous message goes to a person instead of confidently into the wrong queue
- A prompt instruction is a request. A policy lane is a guarantee.

Notes: Feature 09 makes the sharpest version of this by failing, live, every time. Set it up now.

---

## 3. 07 · Classification & Routing: The User Problem

- Every message to Trailhead lands in one inbox: permits, conditions, complaints, lost-and-found, and occasionally an actual emergency
- A ranger triages by hand once or twice a day
- The permit waits behind the granola questions; the emergency sits unread for four hours
- The user's real problem: their message goes into a hole

## 4. 07 · The Concept

- Classification again (03 was the warm-up), but now the label has consequences
- Zero-shot: describe the categories in plain language, no training data, which is where every team is on day one
- The taxonomy is the product. When the model misfiles something, fix the description, not the model.
- Tune so the expensive class never slips through, even at the cost of extra false alarms

## 5. 07 · Demo: What to Watch

- Scroll the raw inbox; let the room find the emergency
- One classify method, taxonomy as plain-language descriptions, exactly one label
- The pile becomes tidy queues in seconds; the emergency is at the top of its own
- Planted misclassification (`inq-0030`): fix it by editing the category description, re-run
- The `unsure` route is deliberately narrow, for messages two queues must both act on. `inq-0035` lands there, correctly, and the description forbids sending anything dangerous to it.
- Measured: 17/20 overall, 2/2 emergencies, every run

Notes: Close on asymmetric cost: what this system is tuned never to miss, and what noise level that tolerance buys.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Scroll the raw inbox, `data/inquiries.jsonl`, mixed and unlabeled, and let the room spot the emergency buried in it.
2. Starter: one classify method with the taxonomy written as plain-language category descriptions in the prompt, returning exactly one label.
3. Run the inbox through it and print each message under its routed queue. The payoff is that the pile becomes a set of tidy queues in seconds, with the emergency at the top of its own.
4. Show a misclassification (inq-0030 is planted). Fix it by editing the category description, not the code, and re-run to show the fix landing.
5. Walk the "unsure" route in the taxonomy: it is deliberately narrow, for messages that two queues must both act on. Show inq-0035 landing there, which is correct behavior, and point out that the description forbids sending anything dangerous to it.
6. Close on asymmetric error costs: what this system is tuned never to miss, and what noise level that tolerance buys.

## 6. 07 · Leadership Card

- When: any shared inbox, ticket queue, or intake form where a human sorts before anyone acts. Support, sales leads, HR, incoming documents.
- Cost: days. No training data to start; ongoing work is refining the taxonomy as traffic reveals edge cases. Local models make per-message cost roughly zero.
- "Every message reaches the right person in seconds, and the urgent ones stop waiting in line."

---

## 7. 08 · Anomaly Detection: The User Problem

- ~500 trail-condition reports a season across 200 trails. Almost all say "muddy in spots, otherwise fine."
- Then in one week three hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over
- Nobody notices. The park hears about the bridge from a one-star review a month later.

## 8. 08 · The Concept

- Barely an AI feature: embeddings plus arithmetic
- Embed a trail's reports; the routine ones cluster. Average them: a centroid, the center of "normal."
- A report's distance from the centroid is its anomaly score. "Bridge washed out" sits farther from the mud cluster than mud reports sit from each other.
- Cosine distance and a threshold do *most* of the job

Notes: "Most" is doing real work in that sentence, and the demo is honest about it.

## 9. 08 · Demo: What to Watch

- Scroll the report stream, which is boring on purpose. Nobody can find the problem.
- Centroid on one slide: the center of normal
- `--raw` first: the washout reports scatter through the middle of the list. Let that sit for a second.
- Add the model's task prefix (`classification:`): a washout report jumps from rank 11 to rank 2, and the mud reports settle to the bottom
- The alert rule: distance beyond threshold, plus two flagged reports within 14 days. One alert fires, with three genuine washout reports in it and no false positives.
- Second trail: the bear spike, cleaner because that trail's routine chatter is more uniform
- Count the model calls: embeddings only. Everything else was subtraction.

Notes: Embedding models have usage contracts; reading the model card is engineering work. When 8 of 40 reports describe the same washout they drag the centroid toward themselves; that is why the alert rule beats the ranking, and why the stretch goal builds the centroid from reports before the window. Accuracy here is a property of your corpus, not your code.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Scroll the condition-report stream, which is boring on purpose, and ask the room to find the problem. Nobody can, and that is the situation the park is in.
2. Embed one trail's reports with the feature 04 embedding code, average the vectors into a centroid, and put the idea on one slide: the center of normal.
3. Run it with `--raw` first and print every report's distance, sorted. It underwhelms, because the washout reports scatter through the middle of the list, and it is worth sitting in that for a second, since this is what the technique actually does out of the box.
4. Add the model's task prefix (`classification:`) and re-run. A washout report jumps to rank 2 and the mud reports settle at the bottom. The lesson: embedding models have usage contracts, and reading the model card is engineering work, not homework.
5. Now the alert rule, which is where the feature actually lives: distance beyond threshold, plus two or more flagged reports within a two-week window. One alert fires on this trail, three genuine washout reports in it, nothing false. Show it also catching the bear-activity spike on the other trail, where the signal is even cleaner because that trail's routine chatter is more uniform. Accuracy here is a property of your corpus, not your code.
6. Count the model calls: embeddings only, and everything after them was subtraction. Some AI features are mostly arithmetic wearing an AI badge.

## 10. 08 · Leadership Card

- When: any stream of routine text where the rare exception is expensive to miss. Tickets, logs, safety reports, transaction notes, review streams.
- Cost: days, and cheap to run forever. Embeddings are the only model cost.
- "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

Notes: Pairs with 07: classification handles the categories you knew to define; anomaly detection catches the things you didn't.

---

## 11. 09 · Human-in-the-Loop: The User Problem

- 07 routed the inbox; now the ranger has forty messages that need replies. Most are boilerplate typed a hundred times.
- The obvious move: let the AI answer
- The obvious disaster: the AI tells a visitor campfires are fine during a burn ban, on official letterhead
- The ranger's problem is drudgery. The park's problem is that fully automating outbound communication is how you end up apologizing publicly.

## 12. 09 · The Concept

- A product pattern, not a model feature. The difference between AI features that ship and ones killed in legal review.
- AI drafts; a human approves, edits, or rejects; the system remembers what happened
- Three lanes by error cost and reversibility: auto-send · draft-for-approval · human-only
- Emergencies never get an AI draft at all
- Audit trail, and measure the draft-to-final gap: how much people edit tells you whether trust in each lane is earned

## 13. 09 · The Failure That Makes the Point

- Asked to draft a reply to a woman whose husband is four hours overdue, the model ignores its escalation instruction and writes a warm, reassuring, useless note. Every run.
- Move the instruction to the top: it announces `ESCALATE`, then writes the note anyway. Every run.
- More prompt engineering is not the lesson
- The finished demo refuses to send emergencies to the model at all. The policy lives in code instead of in the prompt.

Notes: This is the sharpest slide in the module. Show the raw prompt fail live if time allows; the reference runs are in `expected-output.md`, 3/3 both ways.

## 14. 09 · Demo: What to Watch

- A routed queue of condition questions awaiting replies
- Draft one reply, grounded in park docs the way 05 grounded answers. It is good without being perfect.
- Approve one untouched, edit a second, reject a third and write it by hand
- The audit log: draft, decision, final text, edit distance per reply
- The policy table on screen; point at the emergency row
- The edits from step 3 are tomorrow's prompt improvements. The human is the feedback signal as much as the safety check.

Notes: Keep the emergency row on screen for a beat: no AI touches it, and that is a design decision, not a limitation.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Pick up where feature 07 left off: a routed queue of condition questions awaiting replies.
2. Generate a draft reply for one inquiry, grounded in the park docs the same way feature 05 grounded answers. Show the draft on screen; it is good without being perfect.
3. Run the approval flow in a small console UI: approve one draft untouched, edit a second before sending, reject a third and write it by hand. Three inquiries handled in the time one used to take.
4. Show the audit log the flow just produced: draft, decision, final text, edit distance per reply.
5. Put the routing policy on screen as a table: which categories auto-send, which get drafts, which stay human-only. Point at the emergency row: no AI draft, ever, by policy.
6. Close the loop: the edits collected in step 3 are tomorrow's prompt improvements. The human is the feedback signal as much as the safety check.

## 15. 09 · Leadership Card

- When: any customer-facing or high-stakes generation. Support replies, quotes, claims decisions, anything sent under your name.
- Cost: drafting is trivial; the approval UI, audit trail, and policy design are ordinary product engineering, measured in weeks. Often what makes the other nine shippable.
- "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can."

---

## 16. Hands-On: 60 Minutes

- **Start with 07 Classification** (Recommended): structured output with an enum plus an honest scoring pass. The number that matters is whether both emergencies were caught. Missing one fails the lab at 19/20.
- Then choose by energy as much as by interest:
  - **08 Anomaly Detection**: the most code. Ships precomputed vectors so you can go straight to the math.
  - **09 Human-in-the-Loop**: the least code; mostly a policy worksheet and a decision about what your software may do unsupervised. Running out of gas? Take this. It is also the one most likely to matter back at work.

## 17. Debrief

- Whose classifier caught both emergencies? Whose model wrote the reassuring note?
- Rows 7 to 9 of the decision framework are done
- Break, then Module 4, and everyone does it.
