# 05 · RAG (retrieval-augmented generation)

Block 2 (Finding) · Runs on local embeddings (`nomic-embed-text`) + Azure OpenAI generation, with a local-generation contrast

## The user problem

A visitor asks Trailhead Guides: "Can I have a campfire at Sperry Chalet in September?" The answer exists, in paragraph four of a 12-page backcountry regulations document that nobody will ever read. Search alone returns the document, not the answer. A plain chatbot answers fluently and makes it up, and a confidently wrong answer about fire regulations is worse than no answer at all.

## The concept

RAG bolts module 04's retrieval onto an LLM's generation. Instead of asking the model what it knows, you retrieve the most relevant chunks of your own documents and hand them over with the question: "Answer using only this context. If the context doesn't cover it, say so." The model becomes a reading assistant for your content rather than an oracle, and it can cite which document the answer came from.

The mechanics you'll touch: chunking (splitting 25 park docs into retrievable pieces), retrieval (module 04's embedding search, verbatim), and grounded prompting (context in, citation out, refusal when the context is silent). The model strategy is hybrid on purpose. Retrieval runs on free local embeddings, and you'll try generation both ways: a local model first, then Azure OpenAI. Watching the cloud model handle a multi-document answer more cleanly is the honest version of the "when do I pay for the big model" conversation from module 03, now applied to generation.

## Demo outline (13 min, .NET)

1. Ask the model the Sperry Chalet campfire question with no context. It answers confidently. Then open the actual regulation and show the answer is wrong. Let that sit for a beat; this is the whole reason RAG exists.
2. Show the chunked park docs (chunking already done in the starter, with one slide-worthy comment on chunk size).
3. Reuse the embedding search from module 04 against the chunks. Retrieval is literally the previous module.
4. Build the grounded prompt: retrieved chunks + question + "answer only from the context, cite the source, say when you don't know."
5. Run it. The payoff: the right answer, with the regulation document cited by name.
6. Ask it something the docs don't cover ("Can I bring a drone?" if absent) and get an honest "the provided documents don't say," not a guess.
7. Swap generation from the local model to Azure OpenAI with the one-line Microsoft.Extensions.AI client change, re-run a multi-document question, and compare answer quality side by side.

## Lab spec (13 min, any language)

- **Goal:** answer natural-language questions from the park docs, with citations, and refuse cleanly when the docs are silent.
- **Input:** `lab/` provides pre-chunked park docs (with chunk IDs and source filenames) from `data/park-docs/`, plus four test questions: three answerable, one not.
- **How:** `lab/ollama.http` for embeddings and local generation, `lab/azure.http` for cloud generation. Retrieval is your module 04 code pointed at the chunks.
- **Steps:**
  1. Embed the chunks, retrieve the top 3 for question #1, and eyeball whether retrieval found the right material. If retrieval fails, generation can't save you; that's a lesson, not a bug.
  2. Build the grounded prompt and generate the answer with the source cited.
  3. Run all four questions. Success check: three correct cited answers, and a refusal on the fourth (compare `lab/expected-output.md`).
- **Stretch goal:** vary how many chunks you retrieve (1, 3, 8) and watch answer quality change, or make the answer quote the exact regulation sentence it relied on.

## Leadership beat

- **When to reach for this:** internal knowledge bases, support and policy docs, product documentation. Anywhere the org already wrote the answer down and people still ask humans.
- **Rough cost & effort:** weeks to production quality. The demo takes a day; the real work is chunking strategy, retrieval evaluation, and keeping the document set current. Ongoing model cost applies to generation only.
- **The one-liner for your CTO:** "Our documents already answer these questions. This makes them answer directly, with receipts."

This card is row 5 of the [decision framework](../../docs/decision-framework.md).
