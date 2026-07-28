# Module 2: Finding

**Surfacing the right thing.** 11:15 to 12:45.

Three features built on one idea. Embeddings turn text into vectors, and distance between vectors means similarity of meaning. Search ranks a catalog against a query. RAG retrieves passages and hands them to a model to answer from. Recommendations rank a catalog against an item instead of a query. Same math, three products.

This is the module where the day stops being about prompts.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Core** | [04 Semantic Search](04-semantic-search/) | Finds trails by meaning, not keywords | Ollama `nomic-embed-text` |
| Challenge | [05 RAG](05-rag/) | Answers questions from park regulations, with citations | Local embeddings + generation |
| Challenge | [06 Recommendations](06-recommendations/) | "You liked this trail, try these" | Ollama `nomic-embed-text` |

## How the hands-on works

I present all three with live demos, then you get about 45 minutes.

**Two honest routes through this module.** Do 04 first if embeddings are new to you, because it is where the concept lands and the other two assume it. But if RAG is the reason you came to this workshop, **go straight to 05** and come back to 04 afterwards. Feature 05 ships precomputed chunk vectors, so it does not require finishing 04 first. Nobody should leave this room having never built the thing they came for.

Feature 06 is the shortest of the three if you have already done 04, because it is the same code with the query swapped for an item.

## What each lab costs you

- **04 Semantic Search** is the concept lab. Embed 30 trails, embed a query, rank by cosine similarity. The keyword baseline is provided so you can see what you are beating.
- **05 RAG** is the biggest lab in this module and the most asked-for feature in the industry. Retrieval, a grounded prompt, citations, and a refusal path.
- **06 Recommendations** is the fastest payoff and the most surprising results.

## The thread to watch

Embeddings capture what text is *about*. They do not capture whether it is suitable, safe, current, or complete, and each feature here shows that differently.

Search returns a cliff-edge trail as the top hit for "somewhere quiet to take my kids", because the description talks about children. Recommendations answer "you bought the 65 litre pack" with the 40 litre pack, the one product that buyer will never need. RAG answers correctly only because the chunks were cut so that a rule and its exception stay together; cut them apart and the same pipeline confidently tells a visitor to light a fire where fires are banned year round.

The fix in all three cases is not a bigger model. It is metadata, filters, thresholds, and paying attention to how you split your documents. Feature 05 has the measurements to prove it: better chunking moved its flagship question from 75 to 97 percent correct, and a model ten times the size bought only the last three points.

## The leadership beats

Collected at the debrief, becoming rows 4 through 6 of the [decision framework](../../docs/decision-framework.md).
