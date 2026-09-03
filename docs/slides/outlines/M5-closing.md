# Deck 5: Closing (about 30 minutes)

The decision framework, pitching AI features to leadership, and Q&A. This is the deck the ten leadership cards were building toward all day. Bullets are cues, not sentences; the words live in Notes. Slide markers (`[section]`, `[static]`, `[thanks]`) are explained in `build.py`; unmarked slides reveal item by item as you advance.

---

## [deck-title] Closing: The Thing You Take Back to Work

Notes: Ten features, ten cards, one table: docs/decision-framework.md.

## Back to the First Slide

- ❌ "Where can we add AI?"
- ✅ "What problems can AI best solve for our users?"

Notes: Every feature today started with a stuck user, and the AI showed up as the answer rather than the premise.

## The Framework, All Ten Rows

Icon (document): F01 Summarize
Icon (table): F02 Extract
Icon (gauge): F03 Sentiment
Icon (search): F04 Search
Icon (question): F05 RAG
Icon (route): F06 Recommend
Icon (branch): F07 Route
Icon (scatter): F08 Detect
Icon (shield-check): F09 Approve
Icon (robot): F10 Agent

Notes: One cell per advance: name each feature as it lands, a few seconds apiece — the day, replayed in ten beats.
The full table lives in decision-framework.md: # · Feature · When · Think · Difficulty, the same three fields as every card today. (Paste the real table over this slide, or demo it from the repo.)
Don't read it. Let it sit on screen while you explain how to use it, which is the next section. It's dense on purpose; it's a reference, and they have it in the repo.

## What the Day Measured, Not Assumed

- Sentiment: 7/10 vs 10/10 — only on the hard slice
- Chunking: 75% → 97% · a 10× model bought 3 points
- Classifier: 2/2 emergencies · drafter: 0/3
- Local agent: all four tools < 1 run in 5
- Every number: an `expected-output.md`

Notes: A free local model matched the frontier model on ordinary reviews and lost only on the sarcastic ones. Better chunking, not a bigger model, moved RAG's flagship question. The classifier caught both emergencies every run; the drafting model failed the emergency every run. The runs are checked in behind every number.
This is the credibility slide. "We ran it and here is what happened" is a stronger pitch than "the vendor says."

## [static] What We Didn't Do Today, on Purpose

- No fine-tuning
- No vector database product
- No agent framework
- Reach for them when a measured problem asks

Notes: Nothing today needed fine-tuning. Cosine similarity in a loop carried three features. The agent loop fit on screen.
Optional slide; cut if short on time. It heads off the "but shouldn't we be using X" question.

---

## [section] Your Move

Notes: The recap is actions, not topics. Four moves, each one slide, each something they can do Monday.

## Mark What Your Users Hate

- The "when to reach for this" column
- Mark the rows your users would recognize
- Row 0: list the ten things they hate doing

Notes: Go down the column and mark every row that describes a problem your users actually have. The Row 0 exercise makes the features pick themselves.
This is the workshop's framing question turned into a Monday-morning exercise. It happens in a meeting room, not a codebase.

## Sort by Cost, Cheapest First

- The cheapest marked row = your first AI feature
- Rows 1–8: mostly free local models

Notes: Sort your marked rows by the cost column; the top of that list is the first feature, and the one-liner column is how you pitch it.
The local-model finding is worth repeating here; it changes the budget conversation before it starts.

## Lead with the Human in the Loop

- Row 9 de-risks the rest
- Say where the human sits · what gets logged

Notes: Leading with the human-in-the-loop design is usually what gets the other rows approved.

## Pitch One Feature, Not a Strategy

- The user's problem, in their words
- The row · the cost · local vs paid
- How you'll measure it
- Ask for one feature

Notes: Lead with how often the problem happens. Say what it costs in days or weeks and what it costs to run; say what runs local and what needs a paid model, and why; say the labeled sample and the error you're tuning never to make.
This is the shape of the conversation the framework enables: a small, measurable, de-risked ask. Reference feature 03's recipe: labeled sample, run both, count disagreements, price the errors. It generalizes.

## [static] Everything Is in the Repo

- github.com/trailheadtechnology/10-ai-features-workshop
- Every lab · every `.http` · every expected output
- `starter/` and `complete/` in .NET, Python, TypeScript
- The corpus is yours
- `SETUP.md` still works tomorrow

## [thanks] Thanks! Questions?

Notes: Leave the framework slide up again during Q&A if it's more useful than this one. The free offer (tinyurl.com/th-offer) is on this slide; point at it once.
