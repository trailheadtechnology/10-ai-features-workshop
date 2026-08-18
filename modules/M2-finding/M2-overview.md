# Module 2: Finding

**Surfacing the right thing.** About 90 minutes.

The three features in this module are built on one idea. Embeddings turn text into vectors, and distance between vectors means similarity of meaning; once you have that, search is ranking a catalog against a query, RAG is retrieving passages and handing them to a model to answer from, and recommendations are ranking a catalog against an item instead of a query. It is the same math three times over, and it is the point in the day where the work stops being about prompts.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Recommended** | [04 Semantic Search](F04-semantic-search/F04-spec.md) | Finds trails by meaning, not keywords | Ollama `nomic-embed-text` |
| Challenge | [05 RAG](F05-rag/F05-spec.md) | Answers questions from park regulations, with citations | Local embeddings + generation |
| Challenge | [06 Recommendations](F06-recommendations/F06-spec.md) | "You liked this trail, try these" | Ollama `nomic-embed-text` |

## How the Hands-On Works

The first 30 minutes are mine: the theme, and a demo of all three features. The remaining 60 are yours to build.

**There are two honest routes through this module.** Do 04 first if embeddings are new to you, which is why it is the Recommended one: it is where the concept lands and the other two assume it. But if RAG is the reason you came to this workshop, **go straight to 05** and come back to 04 afterwards. Feature 05 ships precomputed chunk vectors, so it does not require finishing 04 first. Nobody should leave this room having never built the thing they came for.

Feature 06 is the shortest of the three if you have already done 04, because it is the same code with the query swapped for an item.

## What Each Lab Costs You

- **04 Semantic Search** is the concept lab: embed 30 trails, embed a query, and rank by cosine similarity. The keyword baseline is provided so you can see what you are beating.
- **05 RAG** is the biggest lab in this module and the most asked-for feature in the industry, covering retrieval, a grounded prompt, citations, and a refusal path.
- **06 Recommendations** pays off fastest and produces the most surprising results.

## The Thread to Watch

Embeddings capture what a text is *about*, and nothing about whether it is suitable, safe, current, or complete. Each feature here shows that gap differently.

Search returns a cliff-edge trail as the top hit for "somewhere quiet to take my kids", because the description talks about children. Recommendations answer "you bought the 65 litre pack" with the 40 litre pack, the one product that buyer will never need. RAG answers correctly only because the chunks were cut so that a rule and its exception stay together; cut them apart and the same pipeline confidently tells a visitor to light a fire where fires are banned year round.

In all three cases the fix is metadata, filters, thresholds, and attention to how you split your documents, not a bigger model. Feature 05 has the measurements to prove it: better chunking moved its flagship question from 75 to 97 percent correct, and a model ten times the size bought only the last three points.

## The Leadership Beats

Collected at the debrief, becoming rows 4 through 6 of the [decision framework](../../docs/decision-framework.md).
