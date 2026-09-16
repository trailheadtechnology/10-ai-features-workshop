# Lab 05: RAG

*A Challenge lab. Do it if you finished [Module 2](../M2-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** answer questions from the park docs with citations you can verify, and refuse when the docs are silent.
- **Input:** `data/chunks.jsonl`, 250 chunks of park regulations; `data/questions.json`, four test questions, one unanswerable.
- **How:** embed the chunks and the question with Ollama, then rank the chunks by cosine similarity using feature 04's code. Paste the top 3 into a prompt and send it to a chat model. Then check every citation the model wrote against the chunks you gave it.
- **Model:** `nomic-embed-text` for the vectors, `llama3.2` for the answer. Both local, no key. The stretch goals swap in `gpt-4.1` on Azure.

Every step below is one thing to make the program do. Each code track's `starter/` is the plain chatbot. Edit it until it does all seven of steps 1 through 7, and compare against `complete/` when you get stuck. If you are on the HTTP track, run the numbered request in `http/ollama.http` wherever a step names one. That file holds four numbered requests against Ollama, plain JSON bodies you can send as they are or copy into any HTTP client; request 4 already has the top 3 chunks pasted into the prompt. `http/azure.http` holds the same grounded prompt against Azure OpenAI plus one multi-document question, with the endpoint and deployment filled in and a placeholder for the key handed out in the room.

### Step 0: Run the starter and read the wrong answer

Run `starter/` as it is. It sends question 1 to `llama3.2` with no documents attached (request 1 in `http/ollama.http`):

```text
Can I have a campfire at Sperry Chalet in September?
```

**Check:** a fluent, confident answer that puts Sperry Chalet in California, or invents a fire rule. Then open `data/park-docs/glacier-backcountry-camping-guide.md`, Section 4.2, and read the real rule. The rest of the lab closes that gap.

`data/park-docs/` is the corpus: 25 markdown documents covering six parks (Acadia, Glacier, Great Smoky Mountains, Rocky Mountain, Yosemite, Zion), each a regulation, a guide, a seasonal closure notice, or a visitor FAQ, laid out like a real park document with a document number, an effective date, and numbered sections. They were written for this workshop. The park names are real and every rule in them is fiction, so do not plan a trip from them.

### Step 1: Load the chunks

1. Open `../../data/chunks.jsonl`. Every line is one JSON object with three string fields: `chunk_id`, `source`, `text`.
2. Read the file line by line, parse each line, and keep the results in a list.

Those chunks are the 25 park docs cut up by `data/build-chunks.py`: one chunk per numbered section of each document, except that a section over 200 words is split at its own subsection numbers and consecutive subsections are packed together until a chunk clears a 50-word floor. Five sections were long enough to split, which is how 241 sections became 250 chunks. Every chunk opens with the document's title in square brackets and ends with the first sentence of the next unit in the document, marked `(continues in ...)`, so a rule is never retrieved without a pointer to the exception that follows it. `chunk_id` is the file name without `.md`, a colon, and the section number: `:00` is the header block before Section 1, `:04.2` is Section 4.2, and `:04.3-5` is subsections 4.3 through 4.5 packed into one chunk. `source` is the file name. The script's docstring says why each of those choices exists and what breaks when you change the numbers.

**Check:** the list has 250 entries. The first `chunk_id` is `acadia-campfire-and-campground-regulations:00`.

### Step 2: Embed every chunk once and cache the vectors

1. Loop over the list in batches of 32 chunks.
2. For each batch, embed the batch's `text` values in one call to `nomic-embed-text` through your track's embedding client (request 3 in `http/ollama.http` is one such call with two chunks).
3. The response has one vector per input, in the same order you sent them. Store each vector in a dictionary keyed by that chunk's `chunk_id`.
4. When the loop finishes, write the dictionary to `embeddings.json` next to your program. At the top of the program, check whether that file exists; if it does, load it and skip the loop. That file is your program's own cache, the `chunk_id`-to-vector dictionary as JSON, and `complete/` on every track keeps one next to the program for the same reason.

Shortcut if you want to get to the RAG part faster: skip this step and load `../../data/chunk-embeddings.json` instead. It is the same dictionary, already computed and shipped with the workshop: one key per `chunk_id`, each holding the 768-number vector `nomic-embed-text` produced for that chunk's `text`.

**Check:** 250 keys, each holding 768 numbers. The first run takes about 40 seconds and prints nothing wrong. The second run is instant.

### Step 3: Embed the question and rank the chunks

1. Embed question 1 the same way, as a single string (request 2 in `http/ollama.http`). Keep the one vector that comes back.
2. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`. It scores how closely two vectors point the same way, from 0 (unrelated) to 1 (identical).
3. For every chunk, compute the cosine between the question vector and that chunk's vector.
4. Sort the chunks by that score, highest first, and keep the top 3.
5. Print each of the 3 as score and `chunk_id`.

**Check:** rank 1 is `glacier-backcountry-camping-guide:04.2` at cosine 0.7422, rank 2 is `glacier-frontcountry-campground-regulations:04.1-2` at 0.6913, rank 3 is `yosemite-campfire-regulations:03` at 0.6812. Anything else at rank 1 means stop here; no prompt later can fix retrieval.

### Step 4: Build the grounded prompt and get an answer

1. Turn the 3 chunks into a context block. For each chunk, write three lines: `chunk_id: ` plus the id, `source: ` plus the source, then the text. Separate the three chunks with a blank line.
2. Build this prompt, filling in the context block and the question:

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

Context:
<the context block>

Question: <the question>

Answer:
```

3. Send the whole prompt as a single user message to `llama3.2` through your track's chat client, the same call the starter already makes (request 4 in `http/ollama.http` is this prompt with the context already pasted in).
4. Print the answer.

Two details of that prompt are deliberate. The date sits inside the refusal rule, not in its own rule, and the refusal is one exact sentence rather than a general instruction. Both choices were measured; the numbers are in `expected-output.md` under "Telling the Model What Day It Is".

**Check:** the answer says no, wood fires are banned year-round at Sperry Chalet, and cites `[glacier-backcountry-camping-guide:04.2]`. An answer that opens "Yes" and then says fires are prohibited still passes. The failure is `You can have a campfire at Sperry Chalet in September, but only pressurized-gas stoves are permitted...` with no citation.

### Step 5: Check the citations against the chunks you sent

1. Make a set of the 3 `chunk_id` values you put in the context.
2. Find every `[...]` in the answer whose contents include a colon. Split each one on commas, trim the pieces, and keep the pieces that still contain a colon. Those are the citations the model wrote, including the comma-separated lists it sometimes writes inside one pair of brackets.
3. Compare each citation against your set. Any citation not in the set is invalid. For each invalid one, print `!! CITATION CHECK FAILED: [the id] not in the retrieved set`.
4. After the answer, print one summary line: `[citations: N valid (the ids), M invalid]`.

**Check:** on question 1 the summary says 1 valid, 0 invalid. Run the EV charging question from step 6 a few times and you will see the model attach an invented id to its own refusal; the check has to catch it. A made-up id reaching the output unflagged is the failure.

### Step 6: Run all four questions

Run the program once per question in `data/questions.json`. That file is a list of four objects, each with an `id`, the `question` text, an `answerable` flag, and `answer_lives_in`, the document and section the answer sits in (or `nowhere` for question 4). Your program does not need to read it; every track's starter already takes the question as a command-line argument, so pass each one that way:

```text
Can I have a campfire at Sperry Chalet in September?
```

```text
What is the maximum group size on a Glacier backcountry permit?
```

```text
Is the Avalanche Lake Trail open right now?
```

```text
Are there EV charging stations in Glacier National Park?
```

**Check:** question 2 says eight and cites `[glacier-backcountry-permit-regulations:04]`. Question 3 says closed effective June 20, 2026, citing `glacier-seasonal-closures-2026:04.1` or `glacier-visitor-faq:02`. Question 4 replies `The provided documents don't say.` and claims no charger. If question 3 refuses, the date line from step 4 is missing or moved. `No, fuel is not available anywhere within the Park` on question 4 is a miss: the model answered the question next door.

### Step 7: Run question 1 twenty times

If the program is wrong one run in five, a single run will usually look fine. You have to run it enough times to see the bad one. From the folder you have been running the program from:

```bash
for i in $(seq 20); do <run your program> 2>/dev/null | tail -3; done
```

`<run your program>` is whatever you have been typing: `dotnet run` in `dotnet/starter/`, `uv run main.py` in `python/starter/`, `npm run starter` in `typescript/`. The last three lines of each run are the answer, a blank line, and the `[citations: ...]` summary from step 5.

**Check:** every run says no and cites a real id. Count how many open with "Yes" before getting to the ban; the measured rate for `llama3.2` is about 40 percent, and it is not a retrieval bug. The failure, `campfires are permitted at Sperry Chalet area (site code SPE) when the posted fire danger rating is below Very High`, is what this loop exists to catch.

### Stretch goals

Pick any. Each one is already built in every track's `complete/`, and the measurements that justify it are in [`expected-output.md`](expected-output.md).

- **Add a keyword score.** Cosine alone treats "campfire regulations" from five parks as nearly the same thing and barely notices the word "Sperry". Fix that with a second score: count the question's words in each chunk, weight each word by how few chunks contain it (a rare word like "Sperry" counts for more than a common one like "campfire"), rescale both scores to 0..1, and blend `0.6 * cosine + 0.4 * keyword`. Print the top 8 both ways. **Check:** the top 3 do not change, the rank-1 margin grows from 0.1630 to 0.2321, Acadia disappears from ranks 4 through 8, and ranks 4 through 6 become Glacier documents that name Sperry Chalet.
- **Repair a bad citation.** When step 5 finds an invalid id, send the prompt again with an extra line that lists the 3 legal ids and asks for a rewrite. If the second answer is still wrong, replace the bad id with `invalid-citation-removed`. One exception: when the answer is the refusal sentence with a citation attached, just delete the citation and skip the retry. **Check:** no invalid id ever reaches the printed answer, and the refusal string comes back word for word.
- **Point generation at the cloud.** Build the chat client from `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` when they are set, and fall back to `llama3.2` when they are not. Retrieval stays local either way. Then ask a question that needs three documents at once (request 2 in `http/azure.http` on the HTTP track; on a code track, pass it as the argument):

  ```text
  I want to bring nine friends to camp at Sperry Chalet in September and cook over a fire. What do I need to know?
  ```

  **Check:** `gpt-4.1` says the group is over the limit of eight, says no wood fires at Sperry, and cites all three chunks without mixing them up.
- **Build an evaluation loop.** Write ten more questions, each with the chunk_id that should win, then sweep the blend weight from 0 to 1 and record recall@3 at each setting (recall@3 is the share of questions whose correct chunk landed in the top 3). **Check:** a table, and an answer to whether the weight that wins on question 1 wins on the other ten. The rephrasings table in `expected-output.md` seeds the first four rows.

## Pick a Track

Every track does the same steps against the same data and checks its results against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough shows how the steps above look in that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F05-http.md`](http/F05-http.md) | the requests in `http/ollama.http` and `http/azure.http`, or a port of them in your language |
| .NET | [`dotnet/F05-dotnet.md`](dotnet/F05-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F05-python.md`](python/F05-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F05-typescript.md`](typescript/F05-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/park-docs/`: the corpus, 25 fictional park documents across six parks, described at step 0.
- `data/chunks.jsonl`: those documents cut into 250 chunks, one per line, each with `chunk_id`, `source`, and `text`. The chunking rule is spelled out at step 1.
- `data/chunk-embeddings.json`: the 250 `nomic-embed-text` vectors from step 2, keyed by `chunk_id`, in case you want to skip the 40 seconds.
- `data/questions.json`: the four test questions, with an `answerable` flag and where each answer lives.
- `data/build-chunks.py`: the script that made `chunks.jsonl`. Not needed for the lab. Run it with a different word ceiling or floor (`python3 build-chunks.py out.jsonl 400 0`, for example) to change the chunking and see what breaks; the outcomes are already measured.
- `expected-output.md`: real retrieval scores and real answers for all four questions, plus the measurements behind every choice above: chunk size, the keyword blend, citation checking, and why the date is in the prompt.
