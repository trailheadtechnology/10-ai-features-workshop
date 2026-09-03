# The Decision Framework

This is the document the whole day builds toward. Each feature closes with the same three-part card: when to reach for the feature, the one-liner for your CTO, and a difficulty rating. Those cards accumulate here, so by the closing session this page is a complete answer to the question you brought to the workshop: which of these problems do my users have, and which one should we build first?

How to use it back at the office: go down the "When" column and mark every row that describes a problem your users actually have. Sort your marks by the difficulty column, easiest first. The top of that list is your first AI feature, and the "Think" column is how you pitch it.

Difficulty is the workshop's own rating from building all ten. Easy means days of work on free local models with no new infrastructure. Medium means a working pipeline in days, with real design work in schemas, thresholds, or approval flows. Hard means weeks to production quality, and in the agent's case frontier-model costs and guardrail engineering on top.

| # | Feature | When | Think | Difficulty |
|---|---|---|---|---|
| 1 | [Summarization](../modules/M1-understanding/F01-summarization/F01-spec.md) | Too much content for people to read, easy-to-miss detail buried in an avalanche of text | "Our users are drowning in text we already have. A few hours of work turns it into automatic answers." | Easy |
| 2 | [Extraction](../modules/M1-understanding/F02-extraction/F02-spec.md) | Valuable data trapped in documents nobody will re-type | "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it." | Medium |
| 3 | [Sentiment](../modules/M1-understanding/F03-sentiment/F03-spec.md) | A high-volume text stream needing a judgment call | "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun." | Easy |
| 4 | [Semantic search](../modules/M2-finding/F04-semantic-search/F04-spec.md) | A search box users complain about. They describe what they want, and it matches what things are called | "Our search finds what users mean, not just what they type." | Easy |
| 5 | [RAG](../modules/M2-finding/F05-rag/F05-spec.md) | The org already wrote the answer down, and people still ask humans | "Our documents already answer these questions. This makes them answer directly, with receipts." | Hard |
| 6 | [Recommendations](../modules/M2-finding/F06-recommendations/F06-spec.md) | Users finish one thing and the product offers no next step | "The same vectors that power our search give us 'more like this' for free." | Easy |
| 7 | [Classification & routing](../modules/M3-deciding/F07-classification-routing/F07-spec.md) | A shared inbox or queue where a human sorts before anyone acts | "Every message reaches the right person in seconds, and the urgent ones stop waiting in line." | Easy |
| 8 | [Anomaly detection](../modules/M3-deciding/F08-anomaly-detection/F08-spec.md) | A stream of routine text where the rare exception is expensive to miss | "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later." | Medium |
| 9 | [Human-in-the-loop](../modules/M3-deciding/F09-human-in-the-loop/F09-spec.md) | Anything AI-written that goes out under your name | "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can." | Medium |
| 10 | [Agentic workflows](../modules/M4-doing/F10-agentic-workflows/F10-spec.md) | Users state a goal, then do all the clicking themselves, across systems | "Users state the goal in one sentence, the software does the clicking, and a human still signs off on anything that matters." | Hard |

Two cross-cutting lessons from the day belong in any pitch built from this table. First, six of the ten rows are rated Easy and run on free local models, so the cost conversation is smaller than leadership expects. Second, row 9 is the trust pattern that de-risks everything else: leading your pitch with the human-in-the-loop design is usually what gets the other rows approved.
