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
07: which queue does this message go in.
~
08: which of these 500 reports deserves attention.
~
09: what is the software allowed to send without a human reading it first.
The thread to watch: errors are not symmetric. A misrouted complaint costs somebody a day; a misrouted emergency is a headline. Measure recall on the class that matters, give ambiguity an `unsure` lane to a person, and remember: a prompt instruction is a request, a policy lane is a guarantee. Feature 09 makes the sharpest version of that by failing, live, every time. Set it up now.

## Three Ways to Decide

Flow (07 Classify): Inquiry -> Model picks one label from your list -> Queue
Flow (08 Detect): Reports -> Embed -> Distance from "normal" -> Threshold rule -> Alert
Flow (09 Approve): Inquiry -> Model drafts a reply -> Human approves, edits, or rejects -> Send + log

- A label · a "this is unusual" · a signature
- 08 never asks the model to decide — only for numbers
- 09 wraps the other two

Notes: One row per advance. Say which box in each row is the decision.
07 classify: the model picks one label from your list, and the label picks the queue. The decision is the model's.
~
08 detect: embed, measure distance from "normal," apply a threshold rule. The decision is arithmetic.
~
09 approve: the model drafts, a human approves, edits, or rejects, then send and log. The decision is a person's.
~
A label, a "this is unusual," a signature. Three different kinds of decision.
~
Only 07 and 09 ask a model to decide anything; 08 asks it only for numbers.
~
09 wraps the other two: the routing policy says which decisions a model may make alone.

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

Notes: ~8 min · `cd modules/M3-deciding/F07-classification-routing/dotnet`
Files: `inquiries.jsonl` = the full inbox, 100 · `inquiries-slice.jsonl` = the 20 the demo scores · `reference-labels.json` = hand labels
Args: an inquiry id; starter default inq-0005, complete with none = all 20
1. Open `data/inquiries.jsonl`, scroll. Let the room find the emergency.
2. Before: `cd starter && dotnet run`. One label, free text. Nothing stops a label that does not exist.
3. After: `cd ../complete && dotnet run`. All 20 routed and scored, enum-constrained. The emergency at the top of its own queue.
4. inq-0030 is wrong. Edit its category description in `Program.cs`, not the code. `dotnet run` again.
5. `dotnet run -- inq-0035`: `unsure`, correctly. The description forbids anything dangerous landing here.
6. 17/20, 2/2 emergencies. Which number matters.
Cut: 5.

## F07 · Classification & Routing · Leadership Card

- **When:** A shared inbox or queue where a human sorts before anyone acts.
- **Think:** "Every message reaches the right person in seconds, and the urgent ones stop waiting in line."

Difficulty: easy

Notes: Support, sales leads, HR, incoming documents. Ongoing work is refining the taxonomy as traffic reveals edge cases.
~
The line for your boss. Row 7.
~
Easy. If asked about cost: days; no training data to start; local models make per-message cost roughly zero.

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

Notes: Barely an AI feature: embeddings plus arithmetic. Start with one trail's condition reports.
~
nomic-embed-text, with the classification: prefix. The prefix is in the model card, and it matters.
~
Average the vectors. That point is the centroid of "normal."
~
Measure each new report's distance from it. "Bridge washed out" sits farther from the mud cluster than mud reports sit from each other.
~
Beyond the threshold, and two or more in fourteen days.
~
Alert. Every box after the embedding is subtraction and a rule.
~
The model only makes vectors. It never decides anything.
~
The alert rule is the feature: one flagged report is noise, two in fourteen days is a washout.
~
04's code plus one task prefix. Say "some AI features are mostly arithmetic wearing an AI badge" while pointing at the last three boxes.

## [demo] **DEMO** · F08 Anomaly Detection

## [static] F08 · Anomaly Detection · Demo: What to Watch

- The stream is boring on purpose
- The centroid: the center of normal
- ❌ `--raw`: washouts lost mid-list
- ✅ `classification:` prefix: rank 11 → 2
- Alert rule: 1 alert · 3 real reports · 0 false
- The bear trail: cleaner still
- Model calls: embeddings only

Notes: ~9 min · `cd modules/M3-deciding/F08-anomaly-detection/dotnet`
Files: `reports-0117.jsonl` = 40 reports, the washout trail · `reports-0042.jsonl` = 25 reports, the bear trail · `embeddings-0117.json` = precomputed vectors, the starter runs offline on them
Flags: `--raw` = no task prefix · `--trail 0042` = the other trail · `--sigma` = threshold (default mean + 1 sd) · `--window` = days (default 14)
1. Open `data/reports-0117.jsonl`, scroll. Boring on purpose. Find the problem. Nobody can.
2. Centroid: average the vectors, the center of normal.
3. Before: `cd complete && dotnet run -- --raw`. Washouts lost mid-list.
4. After: `dotnet run`. `classification:` prefix. A washout jumps to rank 2. Read the model card.
5. Same output, the alert block: 1 alert, 3 real reports, 0 false.
6. `dotnet run -- --trail 0042`: the bear spike, cleaner still.
7. Model calls: embeddings only. The rest was subtraction.
Cut: 3, go straight to 4.

## F08 · Anomaly Detection · Leadership Card

- **When:** A stream of routine text where the rare exception is expensive to miss.
- **Think:** "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

Difficulty: medium

Notes: Tickets, logs, safety reports, transaction notes, review streams. Pairs with 07: classification handles the categories you knew to define; anomaly detection catches the things you didn't.
~
The line for your boss. Row 8.
~
Medium. If asked about cost: days to build, cheap to run forever; embeddings are the only model cost.

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

Notes: A product pattern, not a model feature: the difference between AI features that ship and ones killed in legal review. It starts with a routed inquiry from 07.
~
Retrieve the park docs, the way 05 did.
~
The model drafts a reply.
~
A human approves, edits, or rejects. This box is the feature.
~
Send, and log the draft, the decision, and the final text.
~
The model types, a person decides, the log proves it.
~
A policy table sits in front of this line. Three lanes by error cost and reversibility: auto-send, draft-for-approval, human-only. The emergency row skips the whole diagram; that is the design, not a limitation.
~
Today's edits are tomorrow's prompt. The human is the feedback signal as much as the safety check.

## [demo] **DEMO** · F09 Human-in-the-Loop

## [static] F09 · Human-in-the-Loop · Demo: What to Watch

- The escalation failure, live: `ESCALATE`… then the note anyway
- A routed queue awaiting replies
- One grounded draft: good, not perfect
- Approve · edit · reject-and-rewrite
- The audit log, with edit distance
- The policy table — point at the emergency row

Notes: ~9 min · `cd modules/M3-deciding/F09-human-in-the-loop/dotnet`
Files: `inquiries.jsonl` = a queue of 6, one is the overdue-husband emergency · `outbox/` = what got sent · `decisions.jsonl` = the audit log
Flags: `--policy` = print the routing table and exit · `--auto-approve-dry-run` = non-interactive, testing only
1. Before: `cd starter && dotnet run`. Six drafts, all sent, none reviewed. The emergency gets a warm, useless note. Every time.
2. After: `cd ../complete && dotnet run`. The queue, one at a time. Approve one. Edit one. Reject one and write it.
3. Open `decisions.jsonl`: draft, decision, final text, edit distance.
4. `dotnet run -- --policy`: the table. Point at the emergency row. No AI, by policy.
5. The edits are tomorrow's prompt.
Cut: 3.

## F09 · Human-in-the-Loop · Leadership Card

- **When:** Anything AI-written that goes out under your name.
- **Think:** "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can."

Difficulty: medium

Notes: Support replies, quotes, claims decisions. Often what makes the other nine shippable.
~
The line for your boss. Row 9.
~
Medium. If asked about cost: drafting is trivial; the approval UI, audit trail, and policy design are ordinary product engineering, measured in weeks.

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
