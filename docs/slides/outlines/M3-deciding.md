# Deck 3: Module 3, Deciding (about 90 minutes)

Demo-centric rhythm: name the three features and the thread (errors are not symmetric), then per feature — divider, DEMO, the lab, the leadership card. Concept slides only where a picture earns it: the Three Ways to Decide comparison, 08's centroid/threshold pipeline, and 09's human-gate flow. Everything J. says lives in Notes; demo scripts live under each "What to Watch" slide.

Runsheet (lengths, not clock times):
- 30 min instructor: opener 4 min (This Module + the Three Ways to Decide diagram), then per feature — spoken problem setup over the section slide, demo, leadership card. Demo budget: 07 ≈ 8, 08 ≈ 9, 09 ≈ 9.
- 60 min build: the Hands-On slide up; 07 is Recommended, 08 and 09 are Challenge. Sticky check at about 20 min.
- Cut if behind: the "unsure" route in demo 07 (step 5), the `--raw` first pass in demo 08 (step 3, go straight to the prefixed run), the audit log in demo 09 (step 4), then say a card in one sentence.
- Last 5 min: debrief.

---

## [deck-title] Module 3: Deciding

Notes: The first two modules produced answers for a user. This one produces decisions for the people running the product — a ranger, an ops lead, a support manager.

## This Module

Icon (branch): F07 Route
Icon (scatter): F08 Detect
Icon (shield-check): F09 Approve

Notes: The user is usually a colleague; the rhythm again: three demos, then your lab hour.
Which queue does this go in, which of these 500 reports deserves attention, what is the software allowed to send without a human reading it first.
The thread to watch: errors are not symmetric. A misrouted complaint costs somebody a day; a misrouted emergency is a headline. Measure recall on the class that matters, give ambiguity an `unsure` lane to a person, and remember: a prompt instruction is a request, a policy lane is a guarantee. Feature 09 makes the sharpest version of that by failing, live, every time — set it up now.

## Three Ways to Decide

Flow (07 Classify): Inquiry -> Model picks one label from your list -> Queue
Flow (08 Detect): Reports -> Embed -> Distance from "normal" -> Threshold rule -> Alert
Flow (09 Approve): Inquiry -> Model drafts a reply -> Human approves, edits, or rejects -> Send + log

- A label · a "this is unusual" · a signature
- 08 never asks the model to decide — only for numbers
- 09 wraps the other two

Notes: One row per advance. Only 07 and 09 ask a model to decide anything; 08 decides with arithmetic. The routing policy in 09 says which decisions a model may make alone. Say which box in each row is the decision.

---

## [section] F07 · Classification & Routing

Notes: The user problem, spoken, 60 seconds: every message to Trailhead lands in one inbox — permits, conditions, complaints, lost-and-found, and occasionally an actual emergency. A ranger triages by hand once or twice a day; the permit waits behind the granola questions, and the emergency sits unread for four hours. The user's real problem: their message goes into a hole.
No concept slide: it's classification again (03 was the warm-up), now with consequences. Zero-shot, taxonomy as plain-language descriptions, and the demo shows the whole thing — including fixing a misroute by editing a sentence, not the model.

## [define] Give each message a **label**, and let the label pick the **queue**.

Icon (branch): F07

Notes: The textbook version: text classification assigns one of a fixed set of categories to a piece of text; routing is what the label triggers. Zero-shot means the categories are described in plain language, with no training data.

## [demo] **DEMO** · F07 Classification & Routing

## [static] F07 · Classification & Routing · Demo: What to Watch

- The room finds the emergency in the raw inbox
- One method · plain-language taxonomy · one label
- The pile → tidy queues, in seconds
- `inq-0030`: fixed by editing a sentence
- `inq-0035` → `unsure`, correctly
- 📊 17/20 · 2/2 emergencies · every run

Notes: The unsure route is deliberately narrow, for messages two queues must both act on, and its description forbids sending anything dangerous to it. Constrain the output: one label from the list, reject anything off-list.
Close on asymmetric cost: what this system is tuned never to miss, and what noise level that tolerance buys.

Demo script (~8 min; step 5 is the first cut):
1. Scroll the raw inbox, `data/inquiries.jsonl`, mixed and unlabeled, and let the room spot the emergency buried in it.
2. Starter: one classify method with the taxonomy written as plain-language category descriptions in the prompt, returning exactly one label.
3. Run the inbox through it and print each message under its routed queue. The payoff is that the pile becomes a set of tidy queues in seconds, with the emergency at the top of its own.
4. Show a misclassification (inq-0030 is planted). Fix it by editing the category description, not the code, and re-run to show the fix landing.
5. Walk the "unsure" route in the taxonomy: it is deliberately narrow, for messages that two queues must both act on. Show inq-0035 landing there, which is correct behavior, and point out that the description forbids sending anything dangerous to it.
6. Close on asymmetric error costs: what this system is tuned never to miss, and what noise level that tolerance buys.

## F07 · Classification & Routing · Leadership Card

- **When:** A shared inbox or queue where a human sorts before anyone acts.
- **Think:** "Every message reaches the right person in seconds, and the urgent ones stop waiting in line."

Difficulty: easy

Notes: Support, sales leads, HR, incoming documents. Ongoing work is refining the taxonomy as traffic reveals edge cases. Row 7.
If asked about cost: days; no training data to start; local models make per-message cost roughly zero.

---

## [section] F08 · Anomaly Detection

Notes: The user problem, spoken, 60 seconds: about 500 trail-condition reports a season across 200 trails, and almost all of them say "muddy in spots, otherwise fine." Then in one week three hikers report a washed-out bridge on the same trail and a fourth mentions aggressive bear activity two trails over — and nobody notices. The park hears about the bridge from a one-star review a month later.

## [define] Spot the item in a stream that **doesn't look like the rest**.

Icon (scatter): F08

Notes: The textbook version: anomaly detection identifies items that deviate significantly from the majority of the data. Here the "majority" is a centroid of embedded reports and deviation is cosine distance.

## F08 · How It Works

Flow: A trail's condition reports -> nomic-embed-text, with the classification: prefix -> Average the vectors: the centroid of "normal" -> Distance of each new report from it -> Beyond threshold, and 2 or more in 14 days -> Alert

- The model only makes vectors
- The alert rule is the feature: 1 is noise · 2 in 14 days is a washout
- 04's code + one task prefix

Notes: Barely an AI feature: embeddings plus arithmetic. Every box after the embedding is subtraction and a rule; "bridge washed out" sits farther from the mud cluster than mud reports sit from each other.
Say "some AI features are mostly arithmetic wearing an AI badge" while pointing at the last three boxes.

## [demo] **DEMO** · F08 Anomaly Detection

## [static] F08 · Anomaly Detection · Demo: What to Watch

- The stream is boring on purpose
- The centroid: the center of normal
- ❌ `--raw`: washouts lost mid-list
- ✅ `classification:` prefix: rank 11 → 2
- Alert rule: 1 alert · 3 real reports · 0 false
- The bear trail: cleaner still
- Model calls: embeddings only

Notes: Embedding models have usage contracts; reading the model card is engineering work. When 8 of 40 reports describe the same washout they drag the centroid toward themselves; that is why the alert rule beats the ranking, and why the stretch goal builds the centroid from reports before the window. Accuracy here is a property of your corpus, not your code.

Demo script (~9 min; step 3 is the first cut — go straight to the prefixed run):
1. Scroll the condition-report stream, which is boring on purpose, and ask the room to find the problem. Nobody can, and that is the situation the park is in.
2. Embed one trail's reports with the feature 04 embedding code, average the vectors into a centroid, and put the idea on one slide: the center of normal.
3. Run it with `dotnet run -- --raw` first and print every report's distance, sorted. It underwhelms, because the washout reports scatter through the middle of the list, and it is worth sitting in that for a second, since this is what the technique actually does out of the box.
4. Add the model's task prefix (`classification:`) and re-run with plain `dotnet run`. A washout report jumps to rank 2 and the mud reports settle at the bottom. The lesson: embedding models have usage contracts, and reading the model card is engineering work, not homework.
5. Now the alert rule, which is where the feature actually lives: distance beyond threshold, plus two or more flagged reports within a two-week window. One alert fires on this trail, three genuine washout reports in it, nothing false. Show it also catching the bear-activity spike on the other trail, where the signal is even cleaner because that trail's routine chatter is more uniform. Accuracy here is a property of your corpus, not your code.
6. Count the model calls: embeddings only, and everything after them was subtraction. Some AI features are mostly arithmetic wearing an AI badge.

## F08 · Anomaly Detection · Leadership Card

- **When:** A stream of routine text where the rare exception is expensive to miss.
- **Think:** "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

Difficulty: medium

Notes: Tickets, logs, safety reports, transaction notes, review streams.
Pairs with 07: classification handles the categories you knew to define; anomaly detection catches the things you didn't. Row 8.
If asked about cost: days to build, cheap to run forever; embeddings are the only model cost.

---

## [section] F09 · Human-in-the-Loop

Notes: The user problem, spoken, 60 seconds: 07 routed the inbox, and now the ranger has forty messages that need replies, most of them boilerplate typed a hundred times. The obvious move is to let the AI answer; the obvious disaster is the AI telling a visitor campfires are fine during a burn ban, on official letterhead. The ranger's problem is drudgery; the park's problem is that fully automating outbound communication is how you end up apologizing publicly.
The demo opens with the failure that makes the point: asked to draft a reply to a woman whose husband is four hours overdue, the model writes a warm, reassuring, useless note — every run, even with the escalation instruction moved to the top, where it announces ESCALATE and then writes the note anyway. More prompt engineering is not the lesson; the policy lives in code.

## [define] The AI drafts; a **person approves** before anything goes out.

Icon (shield-check): F09

Notes: The textbook version (IBM): a human actively participates in the operation, supervision, or decision-making of an automated system. The machine gets the efficiency; the human keeps the judgment.

## F09 · How It Works

Flow: Routed inquiry (from 07) -> Retrieve park docs (05) -> Model drafts a reply -> Human: approve, edit, or reject -> Send, and log draft + decision + final text

- The model types · a person decides · the log proves it
- A policy table sits in front of this line
- Today's edits → tomorrow's prompt

Notes: A product pattern, not a model feature — the difference between AI features that ship and ones killed in legal review. Three lanes by error cost and reversibility: auto-send, draft-for-approval, human-only. The emergency row of the policy table skips the whole diagram; that is the design, not a limitation.

## [demo] **DEMO** · F09 Human-in-the-Loop

## [static] F09 · Human-in-the-Loop · Demo: What to Watch

- The escalation failure, live: `ESCALATE`… then the note anyway
- A routed queue awaiting replies
- One grounded draft: good, not perfect
- Approve · edit · reject-and-rewrite
- The audit log, with edit distance
- The policy table — point at the emergency row

Notes: The raw-prompt failure runs 3/3 both ways; reference runs are in `expected-output.md` if you skip showing it live. The edits collected live are tomorrow's prompt improvements; the human is the feedback signal as much as the safety check.
Keep the emergency row on screen for a beat: no AI touches it, by policy.

Demo script (~9 min; step 4 is the first cut):
0. If time allows, open with the raw-prompt escalation failure live: the overdue-husband inquiry straight into the drafting prompt, the warm useless note out. Then the version with the instruction on top: ESCALATE announced, note written anyway. That is the case for everything that follows.
1. Pick up where feature 07 left off: a routed queue of condition questions awaiting replies.
2. Generate a draft reply for one inquiry, grounded in the park docs the same way feature 05 grounded answers. Show the draft on screen; it is good without being perfect.
3. Run the approval flow in a small console UI: approve one draft untouched, edit a second before sending, reject a third and write it by hand. Three inquiries handled in the time one used to take.
4. Show the audit log the flow just produced: draft, decision, final text, edit distance per reply.
5. Put the routing policy on screen as a table: which categories auto-send, which get drafts, which stay human-only. Point at the emergency row: no AI draft, ever, by policy.
6. Close the loop: the edits collected in step 3 are tomorrow's prompt improvements. The human is the feedback signal as much as the safety check.

## F09 · Human-in-the-Loop · Leadership Card

- **When:** Anything AI-written that goes out under your name.
- **Think:** "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can."

Difficulty: medium

Notes: Support replies, quotes, claims decisions. Often what makes the other nine shippable. Row 9.
If asked about cost: drafting is trivial; the approval UI, audit trail, and policy design are ordinary product engineering, measured in weeks.

---

## [static] Lab 3: ~60 Minutes

- ⭐ Recommended: **F07 Classification & Routing**
- ⛰️ Challenge: **F08 Anomaly Detection** · **F09 Human-in-the-Loop**
- `FNN-lab.md` · `http/ollama.http` · `expected-output.md` · `data/`
- ✅ Done = 2/2 emergencies caught — 19/20 with a miss is a fail

Notes: 07 is structured output with an enum plus an honest scoring pass; the number that matters is emergency recall, not overall accuracy.
Then choose by energy as much as by interest: 08 lets you go straight to the math; 09 is mostly a decision about what your software may do unsupervised — running out of gas? Take this. It is also the one most likely to matter back at work.

## [static] Lab 3: Debrief

- How'd it go?				  Observations?			        Questions?

Notes: Ask two people what surprised them.
Prompts if quiet: whose classifier caught both emergencies? Whose model wrote the reassuring note? Rows 7–9 of the framework are done. Break, then Module 4 — everyone does it.
