# 05 · RAG (retrieval-augmented generation)

Module 2: Finding · Runs on local embeddings (`nomic-embed-text`) + Azure OpenAI generation, with a local-generation contrast

## The user problem

A visitor asks Trailhead Guides: "Can I have a campfire at Sperry Chalet in September?" The answer exists, in paragraph four of a 12-page backcountry regulations document that nobody will ever read. Search alone returns the document, not the answer. A plain chatbot answers fluently and makes it up, and a confidently wrong answer about fire regulations is worse than no answer at all.

## The concept

RAG bolts feature 04's retrieval onto an LLM's generation. Instead of asking the model what it knows, you retrieve the most relevant chunks of your own documents and hand them over with the question: "Answer using only this context. If the context doesn't cover it, say so." The model becomes a reading assistant for your content rather than an oracle, and it can cite which document the answer came from.

The mechanics you'll touch: chunking (splitting 25 park docs into retrievable pieces), retrieval (feature 04's embedding search, plus a lexical signal it turns out to need), and grounded prompting (context in, citation out, refusal when the context is silent).

A third mechanic is easy to leave out and expensive to leave out: **the model has to be told what "now" means.** The park corpus is written the way operational documents are actually written, in dated notices ("Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice"). Ask "is the trail open right now?" and a model with no calendar cannot connect the two, so it refuses a question its documents answer twice over. The finished demo puts the current date in the prompt next to the refusal rule, and the refusal rate on that question drops from better than half to one run in twenty. Almost every real knowledge base is a corpus of dated notices, and a RAG system that never tells the model the date will either refuse answerable questions or answer them as of an unknown date, with nothing in the output to tell you which.

Each of those mechanics has a failure mode worth showing rather than glossing. **Chunking is a correctness decision, not a plumbing decision.** Splitting the park docs one chunk per numbered section is the obvious default, and it put a conditional fire rule and the absolute exception that overrides it into the same 256-word chunk. Retrieval ranked that chunk first on every phrasing of the question, and the model read the conditional, stopped, and told the visitor a campfire was fine in 4 runs out of 20. Splitting oversized sections at their own subsection boundaries took that to 0 in 60, with the retrieval scores barely moving. **Embedding search alone is weak on proper nouns.** Blending a plain keyword score into the ranking, weighted by how rare each word is in the corpus, makes "Sperry" count for something; that blend is called hybrid retrieval and it is most production RAG systems' first upgrade. **And a citation is a string the model typed, not a fact.** Small models emit chunk ids that look right and point nowhere. Checking each cited id against the set you actually retrieved is a five-line function and it is the difference between a receipt and a decoration.

The model strategy is hybrid in a second sense, on purpose. Retrieval runs on free local embeddings, and you'll try generation both ways: a local model first, then Azure OpenAI. Watching the cloud model handle a multi-document answer more cleanly is the honest version of the "when do I pay for the big model" conversation from feature 03, now applied to generation.

## Demo outline (about 12 min, .NET)

1. Ask the model the Sperry Chalet campfire question with no context. It answers confidently. Then open the actual regulation and show the answer is wrong. Let that sit for a beat; this is the whole reason RAG exists.
2. Show the chunked park docs, then open Section 4 of the Glacier backcountry guide in `data/park-docs/` and read 4.1 and 4.2 aloud in order. A conditional rule, then the absolute exception that cancels it. Ask the room which one a model answers from. This is the chunking beat and it is worth 90 seconds; the numbers are in `lab/expected-output.md` under "Chunking".
3. Reuse the embedding search from feature 04 against the chunks. Retrieval is literally the previous feature. Show the score table for the Sperry question, then run `--top-k 8 --alpha 1.0` and point at ranks 4 through 8: Acadia and Yosemite campfire sections, five of eight chunks from the wrong park.
4. Turn on hybrid retrieval. Same cosine score, plus a keyword score weighted by word rarity, combined with a visible alpha. Re-run `--top-k 8` at the default alpha: Acadia is gone and the wrong-park chunks are replaced by Glacier documents that name Sperry. The printed table shows both component scores, so you can point at "sperry" doing the work.
5. Build the grounded prompt: retrieved chunks + question + "answer only from the context, cite the source, say when you don't know."
6. Run it. The payoff: the right answer, with the regulation document cited by name.
7. Show the citation check. Parse the ids out of the answer, compare against what was retrieved, print a loud failure on a mismatch. Run the unanswerable question a couple of times until the model appends an invented chunk id to its own refusal, which it does often enough to count on.
8. Ask it something the docs don't cover ("Can I bring a drone?" if absent) and get an honest "the provided documents don't say," not a guess.
8b. Ask "Is the Avalanche Lake Trail open right now?" and show the date sitting in the prompt. Explain why it is there: without it the same question got refused in 10 runs out of 18, not because retrieval missed but because "effective June 20, 2026, until further notice" is unanswerable without a calendar. Worth 60 seconds; it is the most transferable thing in the feature.
9. Swap generation from the local model to Azure OpenAI with the one-line Microsoft.Extensions.AI client change, re-run a multi-document question, and compare answer quality side by side.

## Lab spec (Challenge lab, any language)

*A Challenge lab. Do it if you finished [Module 2](../README.md)'s Core lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** answer natural-language questions from the park docs, with citations you have verified, and refuse cleanly when the docs are silent.
- **Input:** `lab/` provides pre-chunked park docs (with chunk IDs and source filenames) from `data/park-docs/`, plus four test questions: three answerable, one not.
- **How:** `lab/ollama.http` for embeddings and local generation, `lab/azure.http` for cloud generation. Retrieval is your feature 04 code pointed at the chunks.
- **Steps:**
  1. Embed the chunks, retrieve the top 3 for question #1, and eyeball whether retrieval found the right material. Print the scores, not just the ids: the margin is the story. If retrieval fails, generation can't save you; that's a lesson, not a bug.
  2. Add a lexical score (count query words in the chunk, weight each by how few chunks contain it) and combine it with the cosine score. Re-run question #1 and the rephrasings in `expected-output.md` at `--top-k 8`, and compare both the margins and which parks fill the rest of the context.
  3. Build the grounded prompt and generate the answer with the source cited.
  4. Validate the citations: pull the bracketed ids out of the answer, check each against the chunks you retrieved, and fail loudly on a mismatch.
  5. Run all four questions. Success check: three correct cited answers, a refusal on the fourth, and no invalid citation reaching the output unflagged (compare `lab/expected-output.md`). Question 3 asks about "right now"; if your model refuses it, do not go debugging retrieval. Put today's date in the prompt and read the measured before-and-after in `lab/expected-output.md`.
  6. Run question #1 twenty times, not once. A wrong answer that shows up one run in five is invisible in a single run and is the only defect in this feature that could hurt somebody.
- **Stretch goal:** build a real evaluation loop instead of eyeballing one question. Write ten more questions with the chunk that should win, then sweep alpha from 0 to 1 and report recall@3 and mean rank-1 margin at each setting. Defend your chosen alpha with the table rather than with the Sperry question, and see whether the setting that wins on Sperry wins on the other ten.

## Leadership beat

- **When to reach for this:** internal knowledge bases, support and policy docs, product documentation. Anywhere the org already wrote the answer down and people still ask humans.
- **Rough cost & effort:** weeks to production quality. The demo takes a day; the real work is chunking strategy, retrieval evaluation, and keeping the document set current. Ongoing model cost applies to generation only.
- **The uncomfortable one:** the worst bug in this feature was not in any model or any prompt. It was a text-splitting rule that separated a rule from its exception, it produced a confident wrong answer about fire safety once every five runs, and every retrieval metric on the dashboard said the system was working. Ask what your evaluation would have caught.
- **Where the money goes:** swapping generation from a 3B local model to a 32B one closes every remaining defect in this feature and costs about 20x the latency per answer, 0.9 seconds to 16.6. Swapping it before fixing the chunk boundary would have bought a smarter model reading the wrong context. Chunking is upstream of model choice; the measured table is in `lab/expected-output.md`.
- **The one-liner for your CTO:** "Our documents already answer these questions. This makes them answer directly, with receipts."

This card is row 5 of the [decision framework](../../../docs/decision-framework.md).
