# HTTP Walkthrough for 05 RAG

The lab steps in [`../F05-lab.md`](../F05-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: four requests against Ollama: the ungrounded question, embedding a question, embedding a batch of chunks, and the grounded prompt built from the top 3 chunks hybrid retrieval actually returns. Request 4's prompt is byte-identical to the one the `complete/` projects send.
- `azure.http`: the same grounded prompt against Azure OpenAI, plus a multi-document question. The endpoint and deployment are filled in; paste the key handed out in the room over `<KEY FROM INSTRUCTOR>`.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### First, Request 1: No Context

Send it and read the answer: confident, specific, and wrong, with Sperry Chalet in the wrong state and a guessed fire rule. Everything below exists because of this.

### Lab Step 1: Requests 2 and 3, Embed and Retrieve

Request 2 embeds the question; request 3 embeds a batch of chunks. Run request 3 over all 250 lines of `data/chunks.jsonl` in your language (32 at a time is fine), or start from `data/chunk-embeddings.json`, which has them precomputed. Cosine-rank the chunks against the question vector and print the top 3 with scores.

Check: `glacier-backcountry-camping-guide:04.2` at rank 1, cosine 0.7422, and a small margin over rank 2. Print the scores; the margin is the story.

### Lab Step 2: Add a Lexical Score

Count query words in each chunk, weight each by how few chunks contain it (IDF), min-max both signals to 0..1, and combine with an alpha (0.6 on the semantic side is what the demo uses). Re-rank at top-8 with alpha 1.0 and then 0.6 and compare which parks fill the context. The formula is in any `complete/`; the tokenizer and stop-word list are ten seconds to read.

Check: at alpha 1.0, five of eight chunks are from the wrong park; at 0.6 they are Glacier documents that name Sperry.

### Lab Step 3: Request 4, the Grounded Prompt

Request 4 is the grounded prompt with the real hybrid top 3 inlined. Read it before sending: the exact refusal string, and today's date inside the refusal clause. Send it, then send it against Azure with `azure.http` request 1.

Check: the correct answer with `[glacier-backcountry-camping-guide:04.2]` cited.

### Lab Step 4: Validate the Citations

Pull every bracketed token containing a colon out of the answer and check it against the three chunk ids you put in the context. Fail loudly on a mismatch. Send request 4 for the unanswerable question a few times and watch the model attach an invented id to its own refusal.

### Lab Steps 5 and 6: All Four Questions, Then Twenty Runs

`data/questions.json` has three answerable questions and one that is not. Build request 4 for each (retrieve, inline the top 3, ask). Then send the Sperry version twenty times and count how many open with "Yes" before saying fires are banned. Question 3 asks about "right now"; if the model refuses it, do not debug retrieval, put the date in the prompt.

Check: three correct cited answers, a refusal on the fourth, no invalid citation reaching the output unflagged. Compare [`../expected-output.md`](../expected-output.md).
