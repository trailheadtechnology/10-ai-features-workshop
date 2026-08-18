# Deck 5: Closing (about 30 minutes)

The decision framework, pitching AI features to leadership, and Q&A. This is the deck the ten leadership cards were building toward all day.

---

## 1. Closing: The Thing You Take Back to Work

- Ten features, ten cards, one table
- `docs/decision-framework.md`

## 2. Back to the First Slide

- The wrong question: "Where can we add AI?"
- The better one: "What problems can AI best solve for our users?"
- Every feature today started with a stuck user, and the AI showed up as the answer rather than the premise

## 3. The Framework, All Ten Rows

*(one slide, the table from `decision-framework.md`: # · Feature · When to reach for this · Cost & effort · The one-liner)*

- 1 Summarization · 2 Extraction · 3 Sentiment
- 4 Semantic search · 5 RAG · 6 Recommendations
- 7 Classification & routing · 8 Anomaly detection · 9 Human-in-the-loop
- 10 Agentic workflows

Notes: Don't read it. Let it sit on screen while you explain how to use it (next slide). It's dense on purpose; it's a reference, and they have it in the repo.

## 4. How to Use It Monday Morning

- Go down the "when to reach for this" column. Mark every row that describes a problem your users actually have.
- Sort your marks by the cost column, cheapest first.
- The top of that list is your first AI feature.
- The one-liner column is how you pitch it.

## 5. Two Lessons That Belong in Every Pitch

- Rows 1 through 8 mostly run on free local models. The cost conversation is smaller than leadership expects.
- Row 9 is the trust pattern that de-risks everything else. Leading with the human-in-the-loop design is usually what gets the other rows approved.

## 6. What the Day Measured, Not Assumed

- On sentiment, a free local model matched the frontier model on ordinary reviews and lost to it 7/10 vs 10/10 on the sarcastic ones
- Chunking moved RAG's flagship question from 75% to 97% correct; a model ten times the size bought the last three points
- The classifier caught both emergencies every run; the drafting model failed the emergency every run
- The local agent reached all four planning tools in fewer than one run in five
- Every number is in a `expected-output.md`, with the runs behind it

Notes: This is the credibility slide. "We ran it and here is what happened" is a stronger pitch than "the vendor says."

## 7. Pitching an AI Feature to Leadership

- Lead with the user's problem, in their words, and how often it happens
- Name the row. Say what it costs in days or weeks, and what it costs to run.
- Say what runs local and what needs a paid model, and why
- Say where the human sits (row 9), and what gets logged
- Say how you'll measure it: the labeled sample, the error you're tuning never to make
- Ask for one feature rather than a strategy

Notes: This is the shape of the conversation the framework enables. A small, measurable, de-risked ask. Reference feature 03's recipe: labeled sample, run both, count disagreements, price the errors. It generalizes.

## 8. What We Didn't Do Today, on Purpose

- No fine-tuning. Nothing today needed it.
- No vector database product. Cosine similarity in a loop carried three features.
- No agent framework. The loop was on screen.
- Reach for those when a measured problem calls for them, not before

Notes: Optional slide; cut if short on time. It heads off the "but shouldn't we be using X" question.

## 9. Everything Is in the Repo

- github.com/trailheadtechnology/10-ai-features-workshop
- Every lab, every `.http` file, every expected output, and every starter and complete project in .NET, Python, and TypeScript
- The corpus is yours to reuse
- `SETUP.md` still works after today

## 10. Questions

- [contact details / how to reach J. Tower]

Notes: Leave the framework slide up again during Q&A if it's more useful than this one.
