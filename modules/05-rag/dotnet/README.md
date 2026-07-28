# .NET demo for 05 RAG

Two console projects, both built on Microsoft.Extensions.AI:

- `starter/`: the demo's starting point. One `IChatClient`, one question, no context. It produces the confident wrong answer that puts Sperry Chalet in the San Gabriel Mountains of California.
- `complete/`: the finished demo as shown on stage. Embeds `../../lab/chunks.jsonl` with `nomic-embed-text`, retrieves top-k with hybrid scoring, generates a cited answer, and validates the citations before printing them.

Retrieval always runs locally. Generation goes to Azure OpenAI when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are all set, and falls back to local `llama3.2` when they are not, printing which one it picked. That fallback is the whole of step 7 in the demo outline: one client construction changes, the prompt and the retrieval do not.

From `starter/`:

```bash
dotnet run                                  # the Sperry question, no context, no receipts
dotnet run -- "Is the Avalanche Lake Trail open right now?"
```

From `complete/`:

```bash
dotnet run                                                    # the Sperry question, grounded and cited
dotnet run -- --alpha 1.0                                     # pure cosine: the 0.0004 margin
dotnet run -- --retrieval-only "What are the campfire rules at Sperry Chalet?"
dotnet run -- --retrieval-only --alpha 1.0 --top-k 10 "What are the campfire rules at Sperry Chalet?"
dotnet run -- "What is the maximum group size on a Glacier backcountry permit?"
dotnet run -- "Are there EV charging stations in Glacier National Park?"   # expect a refusal
dotnet run -- --no-context                                    # step 1 of the demo, inside the finished app
dotnet run -- --top-k 8                                       # more context, worse answers
```

The first run embeds all 241 chunks (roughly 40 seconds) and caches the vectors to `complete/embeddings.json`. Every run after that is instant. Delete that file if you change `chunks.jsonl`.

## Hybrid retrieval

Cosine similarity alone put the correct Sperry chunk at rank 1 by 0.0004, with an Acadia campfire section right behind it. The embedder was matching the phrase "campfire regulations" across five different parks' documents; the word "Sperry" barely moved the needle. Rephrase the question slightly and the correct chunk fell to rank 7. That is not a demo you want to run live.

So `complete/` scores every chunk twice:

- **semantic**, the cosine similarity from module 04, unchanged;
- **lexical**, a BM25-lite score over the chunk text. Query terms are lowercased, split on non-alphanumerics, stripped of question filler ("can", "have", "what"), and singularized crudely so "campfires" in a document matches "campfire" in a question. Each surviving term is weighted by inverse document frequency, so "sperry" (8 of 241 chunks) counts far more than "campfire" (32) or "park" (241).

Both signals are rescaled to 0..1 across the whole corpus for that one question, then combined as `alpha * semantic + (1 - alpha) * lexical`. Default alpha is 0.6. `--alpha 1.0` is the old pure-cosine behavior, which is how you show the before and after without switching branches.

The tool prints the query terms with their IDF, then a table of combined, semantic, and lexical scores for the top hits, then the margin over rank 2. Everything on that table is a number you can point at while explaining why a chunk won.

Alpha 0.6 was picked by measuring, not by taste. Lower weights on the semantic half start pulling in chunks that share rare words by accident: at alpha 0.3, the EV charging question retrieves the bear advisory paragraph about what to do when a grizzly charges, because "charging" appears in exactly one chunk of the corpus and IDF has no idea what it means. Real numbers for every question at both settings are in [../lab/expected-output.md](../lab/expected-output.md).

## Citation validation

`llama3.2` does not reliably copy a chunk_id. It writes `[glacier-backcountry-camping-guide:04.2]`, fusing the id with the section number inside it. It writes `[glacier-bear-safety-advisory:02]` when only `:03` was retrieved. At higher top-k it has written `[glacier-backcountry-campground-regulations:04]`, an id welded out of two real document names that names no document at all. All three look like receipts and none of them are.

After generation, `complete/` pulls every bracketed token containing a colon out of the answer, splits comma-separated lists, and checks each id against the set of chunk_ids actually put in the context.

**On an invalid citation: flag loudly, retry once, then strip.**

```
!! CITATION CHECK FAILED: [glacier-bear-safety-advisory:02] not in the retrieved set
!! retrying once with the valid chunk_ids spelled out
```

The retry appends the exact list of legal ids to the prompt and asks for a rewrite. If the second answer is still wrong, the offending ids are replaced with `[invalid-citation-removed]` and a warning says the answer is unverified where the citation was removed. Every run ends with a `[citations: N valid (...), M invalid]` line, so a clean run is visibly clean rather than silently assumed to be.

Why retry rather than strip immediately: a fabricated citation is usually a copying error, not a reasoning error. The underlying answer is typically correct and grounded, and the model fixes the id when you hand it the list. Stripping first would throw away a good answer's receipts. Why strip rather than retry twice: a second retry costs another round trip on stage and, in the runs measured here, the first retry always fixed it. Why strip rather than fail the whole answer: an answer with one removed citation and a visible warning is more useful to a reader than an error page, and the warning is what a reviewer needs to see.

A tempting alternative is to repair near-misses automatically, truncating `:04.2` back to `:04` when the prefix is a real retrieved id. That would make the demo look tidier and would hide exactly the behavior worth showing, so this build does not do it. Mention it as the production answer if someone asks.

Both projects run fully offline against Ollama. See [../lab/expected-output.md](../lab/expected-output.md) for what the answers and the scores actually look like.
