# Lab 05: RAG

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** answer questions from the park docs with verified citations, and refuse when the docs are silent.
- **Input:** `data/chunks.jsonl` (250 chunks from `data/park-docs/`, each with `chunk_id` and `source`), `data/chunk-embeddings.json` (their `nomic-embed-text` vectors), `data/questions.json` (four questions, one unanswerable); `data/build-chunks.py` made the chunks, not needed.
- **How:** Ollama's embed endpoint for vectors, chat endpoint for answers. `http/ollama.http`: 1 no-context baseline, 2 question embedding, 3 chunk batch embedding, 4 grounded prompt with the top-3 chunks inlined. `http/azure.http`: 1 the grounded prompt on the cloud model, 2 a multi-document question. Retrieval is code (feature 04's cosine search plus step 2); the .NET `complete/` runs it all.
- **Model:** `nomic-embed-text` for retrieval; `gpt-4.1` on Azure for generation, `llama3.2` as local fallback. The room key replaces `<KEY FROM INSTRUCTOR>` in `http/azure.http`; code tracks use Azure when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are set.

### Step 1: Embed the chunks and retrieve the top 3 for question 1

Request 2 embeds question 1, request 3 embeds all 250 chunk texts from `chunks.jsonl` (or load `data/chunk-embeddings.json`); rank every chunk by cosine (`complete/`: `dotnet run -- --retrieval-only --alpha 1.0`). Request 2's input:

```text
Can I have a campfire at Sperry Chalet in September?
```

**Check:** `glacier-backcountry-camping-guide:04.2` at rank 1 for question 1, cosine 0.7422, rank 2 at 0.6913. Anything else at rank 1 means stop; generation cannot fix retrieval.

### Step 2: Add a lexical score and blend it for question 1

No prompt. Score each chunk on the query terms it contains, weighted by how few chunks contain each, min-max both scores to 0..1, and blend `alpha * semantic + (1 - alpha) * lexical` at alpha 0.6 (`complete/Program.cs`: `Tokenize`, `idf`, `lexical`, `MinMax`). Run `dotnet run -- --retrieval-only --top-k 8`, then the same with `--alpha 1.0` added, and compare the two rankings.

**Check:** question 1's rank-1 margin grows from 0.1630 to 0.2321 with the same top 3, and ranks 4 through 8 turn from Acadia and Yosemite to Glacier. If the margin does not grow, check the normalization.

### Step 3: Build the grounded prompt and answer question 1

Request 4 in `http/ollama.http` (`llama3.2`), request 1 in `http/azure.http` (`gpt-4.1`, needs the key), or `dotnet run` in `complete/`; request 4 is these rules, then `Context:`, the three question 1 chunks as `chunk_id:`, `source:`, and text, then `Question: Can I have a campfire at Sperry Chalet in September?` and `Answer:`.

```text
You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "The provided documents don't say."
  A question about "right now" is answered from the context, not refused: today is September 23, 2026,
  and a notice that is in effect "until further notice" is still in effect right now.
```

Azure request 1's system message (the user message is the context, then the `Question:` line):

```text
You are a park information assistant. Answer the visitor's question using ONLY the context the user provides. Base every statement on that context and do not use outside knowledge. Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02]. If, and only if, none of the context is relevant to the question, reply exactly: "The provided documents don't say." Today is September 23, 2026; a notice that is in effect "until further notice" is still in effect right now, so a question about the present is answered from the context rather than refused.
```

**Check:** question 1 gets a no, a real chunk_id in brackets, and nothing from Yosemite or the frontcountry; an answer that opens "Yes" and then says fires are prohibited still passes. The failure is `You can have a campfire at Sperry Chalet in September, but only pressurized-gas stoves are permitted...` with no citation.

### Step 4: Validate the citations in the question 1 answer

No prompt. Pull every `[...]` token containing a colon out of the answer, split comma-separated lists, and check each id against the context's chunk_ids; on a miss, print it, retry once with the valid ids spelled out, then strip anything still wrong and label it `invalid-citation-removed` (`complete/Program.cs`: `Citations`, `InvalidCitations`).

**Check:** an invented id on any question prints `!! CITATION CHECK FAILED: [glacier-bear-safety-advisory:02] not in the retrieved set`. Any bad citation reaching the output unflagged is the failure.

### Step 5: Run all four questions in data/questions.json

Run each question through step 3's prompt with its own top-3 chunks, `dotnet run -- "<question>"` in `complete/`:

```text
Can I have a campfire at Sperry Chalet in September?
What is the maximum group size on a Glacier backcountry permit?
Is the Avalanche Lake Trail open right now?
Are there EV charging stations in Glacier National Park?
```

Then request 2 in `http/azure.http`, a fifth question that needs two documents:

```text
Can I have a campfire at Sperry Chalet in September?
What is the maximum group size on a Glacier backcountry permit?
Is the Avalanche Lake Trail open right now?
Are there EV charging stations in Glacier National Park?
```

**Check:** question 2: eight, citing `[glacier-backcountry-permit-regulations:04]`; question 3: closed effective June 20, 2026, citing `glacier-seasonal-closures-2026:04.1` or `glacier-visitor-faq:02`; question 4: `The provided documents don't say.`, no charger claimed. If question 3 refuses, check the date sits inside the refusal clause as in step 3; `No, fuel is not available anywhere within the Park` on question 4 is a miss.

### Step 6: Run question 1 twenty times

From `complete/`, loop question 1 and read every answer:

```bash
for i in $(seq 20); do dotnet run 2>/dev/null | grep -A3 "^Q:" | tail -2; done
```

**Check:** every run says no with a citation. The failure, `campfires are permitted at Sperry Chalet area (site code SPE) when the posted fire danger rating is below Very High` or a September date borrowed from another park's chunk, only shows up in a loop.

### Stretch goal: a real evaluation loop over ten new questions

Write ten more questions, each with the chunk_id that should win, then sweep `--alpha` from 0 to 1 and record recall@3 and the mean rank-1 margin at each setting. **Check:** a table, not one question, and whether the alpha that wins on question 1 wins on the other ten; the question 1 rephrasings table in `expected-output.md` seeds the first four rows.

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
