# Lab assets for 05 RAG

Everything the lab spec in [../README.md](../README.md) references:

- `chunks.jsonl`: all 25 park docs from `data/park-docs/`, split by numbered section into 241 chunks, one JSON object per line: `{"chunk_id": "glacier-backcountry-camping-guide:04", "source": "...", "text": "..."}`. Chunk IDs are `<filename-slug>:<section-number>`, so `:04` is Section 4 of that document and `:00` is the document header. Every chunk carries its document title as a prefix, because a bare "Section 4.2" retrieved out of context tells the model nothing about which park it belongs to. Longest chunk is 268 words; sections over ~350 words split on paragraph boundaries and take an `a`/`b` suffix.
- `questions.json`: four test questions with an `answerable` flag and a note on where each answer lives. Three are answerable (Sperry Chalet campfires, Glacier group size limit, Avalanche Lake closure); the fourth, about EV charging stations, is not covered anywhere in the corpus.
- `ollama.http`: four ready-to-run requests against Ollama: the ungrounded question (the failure), embedding a question, embedding a batch of chunks, and the grounded prompt built from the top 3 chunks retrieval actually returned. Chunk text is inlined, so each request runs as-is or ports to your language by copying the JSON body.
- `azure.http`: the same grounded prompt against Azure OpenAI, plus a multi-document question that pulls the fire rule from one document and the site status from another. Fill in ENDPOINT, DEPLOYMENT, and YOUR-KEY from the card handed out at the door.
- `expected-output.md`: real `nomic-embed-text` retrieval scores and real `llama3.2` answers for all four questions, including the retrieval near-miss on question 1 and the prompt wording that broke question 3.

Retrieval itself is not in this folder. It is your module 04 embedding search, pointed at `chunks.jsonl`.
