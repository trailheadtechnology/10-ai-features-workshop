# Lab 05: RAG

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** answer natural-language questions from the park docs, with citations you have verified, and refuse cleanly when the docs are silent.
- **Input:** `data/` provides pre-chunked park docs (with chunk IDs and source filenames) from `data/park-docs/`, plus four test questions: three answerable, one not.
- **How:** `http/ollama.http` for embeddings and local generation, `http/azure.http` for cloud generation. Retrieval is your feature 04 code pointed at the chunks.
- **Steps:**
  1. Embed the chunks, retrieve the top 3 for question #1, and eyeball whether retrieval found the right material. Print the scores, not just the ids: the margin is the story. If retrieval fails, generation can't save you; that's a lesson, not a bug.
  2. Add a lexical score (count query words in the chunk, weight each by how few chunks contain it) and combine it with the cosine score. Re-run question #1 and the rephrasings in `expected-output.md` at `--top-k 8`, and compare both the margins and which parks fill the rest of the context.
  3. Build the grounded prompt and generate the answer with the source cited.
  4. Validate the citations: pull the bracketed ids out of the answer, check each against the chunks you retrieved, and fail loudly on a mismatch.
  5. Run all four questions. Success check: three correct cited answers, a refusal on the fourth, and no invalid citation reaching the output unflagged (compare `expected-output.md`). Question 3 asks about "right now"; if your model refuses it, do not go debugging retrieval. Put today's date in the prompt and read the measured before-and-after in `expected-output.md`.
  6. Run question #1 twenty times, not once. A wrong answer that shows up one run in five is invisible in a single run and is the only defect in this feature that could hurt somebody.
- **Stretch goal:** build a real evaluation loop instead of eyeballing one question. Write ten more questions with the chunk that should win, then sweep alpha from 0 to 1 and report recall@3 and mean rank-1 margin at each setting. Defend your chosen alpha with the table rather than with the Sperry question, and see whether the setting that wins on Sperry wins on the other ten.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F05-http.md`](http/F05-http.md) | the requests in `http/ollama.http` and `http/azure.http`, or a port of them in your language |
| .NET | [`dotnet/F05-dotnet.md`](dotnet/F05-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F05-python.md`](python/F05-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F05-typescript.md`](typescript/F05-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/chunks.jsonl`: all 25 park docs from `park-docs/`, split into 250 chunks, one JSON object per line: `{"chunk_id": "glacier-backcountry-camping-guide:04.2", "source": "...", "text": "..."}`. The unit is the numbered section, so `:04` is Section 4 of that document and `:00` is the document header. A section over 200 words is split further at its own subsection boundaries, and those chunks take the subsection number: `:04.2` is Section 4.2, and `:04.3-5` is 4.3 through 4.5 packed together because none of them clears the 50-word floor on its own. Five sections in the corpus were long enough to split. Every chunk carries its document title and section heading as a prefix, because a bare "Section 4.2" retrieved out of context tells the model nothing about which park it belongs to, and every chunk ends with the opening sentence of whatever comes next in the document, marked `(continues in 4.3)`, so a conditional rule can never be retrieved without a pointer to the exception that overrides it. Longest chunk is 230 words, well under the 350-word ceiling. The reason for all of this is measured in `expected-output.md` under "Chunking": one 256-word section held a conditional fire rule and its absolute exception, and the model answered from whichever it read first.
- `data/questions.json`: four test questions with an `answerable` flag and a note on where each answer lives. Three are answerable (Sperry Chalet campfires, Glacier group size limit, Avalanche Lake closure); the fourth, about EV charging stations, is not covered anywhere in the corpus.
- `expected-output.md`: real `nomic-embed-text` retrieval scores and real `llama3.2` answers for all four questions, measured both with pure cosine and with hybrid scoring, plus what citation validation caught, plus the before-and-after numbers for the chunking change, a rejected prompt experiment with the numbers that rejected it, and a `qwen3:32b` comparison answering how much of the remaining gap is the pipeline and how much is the model.

- `data/build-chunks.py`: the script that produced `chunks.jsonl` from `park-docs/`. You do not need it for the lab, since the chunks ship ready to use. It is here so the chunking decision is inspectable and changeable, which matters because that decision is what broke this feature. Run `python3 build-chunks.py` with no arguments and it rewrites `chunks.jsonl` identically. Pass different numbers (`python3 build-chunks.py chunks.jsonl 400 0`) and the strategy changes: the docstring suggests two experiments whose outcomes are already measured in `expected-output.md`, so you can check your prediction against what actually happened.

- `data/chunk-embeddings.json`: all 250 chunk vectors, precomputed with `nomic-embed-text` (768 floats each, keyed by `chunk_id`, rounded to 6 places). This file exists so you can start here. If you did feature 04 first, ignore it and use your own vectors. If RAG is the reason you came and you would rather build it now and come back to search later, load this instead and skip straight to the part this feature is actually about: retrieving the right passages and making a model answer from them without inventing anything.

Retrieval is the one piece this folder does not walk you through, because it is feature 04's cosine search pointed at `chunks.jsonl`. Either bring your own from 04 or load `chunk-embeddings.json` and rank by cosine similarity; both routes reach the same place.

Four things the finished demo adds on top of that search, all four because the measured numbers demanded it:

- **Chunk granularity.** The first version of `chunks.jsonl` split each document by numbered section. Section 4 of the Glacier backcountry guide is 256 words and holds both a conditional rule ("fires permitted only when fire danger is below Very High") and the absolute exception that overrides it ("wood fires are prohibited year-round at all campsites in the Sperry Chalet area"). Handed both in one chunk, `llama3.2` answered from whichever it read first and told the visitor yes in 4 runs out of 20. Splitting oversized sections at their own subsection boundaries took that to 0 in 60. See `expected-output.md`, "Chunking".
- **Hybrid scoring.** Blending a BM25-lite lexical score into the ranking makes a proper noun like "Sperry" count. On the Sperry question it takes the rank-1 margin from 0.1630 to 0.2321. The top 3 is the same either way now that the chunks are the right size, so run this one at `--top-k 8`, where pure cosine fills ranks 4 through 8 with Acadia and Yosemite campfire sections and hybrid fills them with Glacier ones. See `expected-output.md` step 1 and [dotnet/F05-dotnet.md](dotnet/F05-dotnet.md).
- **Citation validation.** Parse the `[chunk-id]` tokens out of the answer and check each one against the ids you actually put in the context. `llama3.2` invents plausible-looking ids often enough that you will see one during the lab.
- **The current date, in the prompt.** Question 3 asks whether a trail is open "right now" and the corpus answers with a dated notice: closed effective June 20, 2026, until further notice. Without a date in the prompt the model refused that question in 10 runs out of 18, with the right chunks sitting at rank 1. Retrieval cannot fix this and neither can the date on its own; see `expected-output.md` for the three experiments and the numbers.
