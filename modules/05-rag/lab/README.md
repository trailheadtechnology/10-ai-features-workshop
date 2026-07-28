# Lab assets for 05 RAG

Everything the lab spec in [../README.md](../README.md) references:

- `chunks.jsonl`: all 25 park docs from `data/park-docs/`, split by numbered section into 241 chunks, one JSON object per line: `{"chunk_id": "glacier-backcountry-camping-guide:04", "source": "...", "text": "..."}`. Chunk IDs are `<filename-slug>:<section-number>`, so `:04` is Section 4 of that document and `:00` is the document header. Every chunk carries its document title as a prefix, because a bare "Section 4.2" retrieved out of context tells the model nothing about which park it belongs to. Longest chunk is 268 words; sections over ~350 words split on paragraph boundaries and take an `a`/`b` suffix.
- `questions.json`: four test questions with an `answerable` flag and a note on where each answer lives. Three are answerable (Sperry Chalet campfires, Glacier group size limit, Avalanche Lake closure); the fourth, about EV charging stations, is not covered anywhere in the corpus.
- `ollama.http`: four ready-to-run requests against Ollama: the ungrounded question (the failure), embedding a question, embedding a batch of chunks, and the grounded prompt built from the top 3 chunks hybrid retrieval actually returns. Chunk text is inlined, so each request runs as-is or ports to your language by copying the JSON body. Request 4's prompt is byte-identical to the one `../dotnet/complete/Program.cs` builds, date line included; if you edit one, edit the other.
- `azure.http`: the same grounded prompt against Azure OpenAI, plus a multi-document question that pulls the fire rule from one document and the site status from another. Fill in ENDPOINT, DEPLOYMENT, and YOUR-KEY from the card handed out at the door.
- `expected-output.md`: real `nomic-embed-text` retrieval scores and real `llama3.2` answers for all four questions, measured both with pure cosine and with hybrid scoring, plus what citation validation caught over twelve runs.

Retrieval itself is not in this folder. It is your module 04 embedding search, pointed at `chunks.jsonl`.

Two things the finished demo adds on top of that search, both because the measured numbers demanded it:

- **Hybrid scoring.** Pure cosine ranks the correct Sperry chunk first by 0.0004 over an Acadia campfire section, and a mild rephrasing ("What are the campfire rules at Sperry Chalet?") drops it to rank 7. Blending a BM25-lite lexical score into the ranking makes the proper noun count and takes the margin to 0.1444. See `expected-output.md` step 1 and [../dotnet/README.md](../dotnet/README.md).
- **Citation validation.** Parse the `[chunk-id]` tokens out of the answer and check each one against the ids you actually put in the context. `llama3.2` invents plausible-looking ids often enough that you will see one during the lab.
- **The current date, in the prompt.** Question 3 asks whether a trail is open "right now" and the corpus answers with a dated notice: closed effective June 20, 2026, until further notice. Without a date in the prompt the model refused that question in 10 runs out of 18, with the right chunks sitting at rank 1. Retrieval cannot fix this and neither can the date on its own; see `expected-output.md` for the three experiments and the numbers.
