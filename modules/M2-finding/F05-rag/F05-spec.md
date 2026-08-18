# 05 · RAG (Retrieval-Augmented Generation)

Module 2: Finding · Runs on local embeddings (`nomic-embed-text`) + `gpt-4.1` on Microsoft Foundry for generation, with a local-generation contrast

## The User Problem

A visitor asks Trailhead Guides: "Can I have a campfire at Sperry Chalet in September?" The answer exists, in paragraph four of a 12-page backcountry regulations document that nobody will ever read. Search alone returns the document, not the answer. A plain chatbot answers fluently and makes it up, and a confidently wrong answer about fire regulations is worse than no answer at all.

## The Concept

RAG bolts feature 04's retrieval onto an LLM's generation. Instead of asking the model what it knows, you retrieve the most relevant chunks of your own documents and hand them over with the question: "Answer using only this context. If the context doesn't cover it, say so." The model becomes a reading assistant for your content rather than an oracle, and it can cite which document the answer came from.

The mechanics you'll touch: chunking (splitting 25 park docs into retrievable pieces), retrieval (feature 04's embedding search, plus a lexical signal it turns out to need), and grounded prompting (context in, citation out, refusal when the context is silent).

A third mechanic is easy to leave out and expensive to leave out: **the model has to be told what "now" means.** The park corpus is written the way operational documents are actually written, in dated notices ("Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice"). Ask "is the trail open right now?" and a model with no calendar cannot connect the two, so it refuses a question its documents answer twice over. The finished demo puts the current date in the prompt next to the refusal rule, and the refusal rate on that question drops from better than half to one run in twenty. Almost every real knowledge base is a corpus of dated notices, and a RAG system that never tells the model the date will either refuse answerable questions or answer them as of an unknown date, with nothing in the output to tell you which.

Each of those mechanics has a failure mode worth showing rather than glossing. **Chunking is a correctness decision.** Splitting the park docs one chunk per numbered section is the obvious default, and it put a conditional fire rule and the absolute exception that overrides it into the same 256-word chunk. Retrieval ranked that chunk first on every phrasing of the question, and the model read the conditional, stopped, and told the visitor a campfire was fine in 4 runs out of 20. Splitting oversized sections at their own subsection boundaries took that to 0 in 60, with the retrieval scores barely moving. **Embedding search alone is weak on proper nouns.** Blending a plain keyword score into the ranking, weighted by how rare each word is in the corpus, makes "Sperry" count for something; that blend is called hybrid retrieval and it is most production RAG systems' first upgrade. **And a citation is only a string the model typed.** Small models emit chunk ids that look right and point nowhere. Checking each cited id against the set you actually retrieved is a five-line function, and without it a citation proves nothing.

The model strategy is hybrid in a second sense, on purpose. Retrieval runs on free local embeddings, and you'll try generation both ways: a local model first, then Azure OpenAI. Watching the cloud model handle a multi-document answer more cleanly is the honest version of the "when do I pay for the big model" conversation from feature 03, now applied to generation.

## The Lab

The hands-on lab is [F05-lab.md](F05-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** internal knowledge bases, support and policy docs, product documentation. Anywhere the org already wrote the answer down and people still ask humans.
- **Rough cost & effort:** weeks to production quality. The demo takes a day; the real work is chunking strategy, retrieval evaluation, and keeping the document set current. Ongoing model cost applies to generation only.
- **The uncomfortable one:** the worst bug in this feature was not in any model or any prompt. It was a text-splitting rule that separated a rule from its exception, it produced a confident wrong answer about fire safety once every five runs, and every retrieval metric on the dashboard said the system was working. Ask what your evaluation would have caught.
- **Where the money goes:** swapping generation from a 3B local model to a 32B one closes every remaining defect in this feature and costs about 20x the latency per answer, 0.9 seconds to 16.6. Swapping it before fixing the chunk boundary would have bought a smarter model reading the wrong context. Chunking is upstream of model choice; the measured table is in `expected-output.md`.
- **The one-liner for your CTO:** "Our documents already answer these questions. This makes them answer directly, with receipts."

This card is row 5 of the [decision framework](../../../docs/decision-framework.md).
