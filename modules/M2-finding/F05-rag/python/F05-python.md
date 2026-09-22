<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 05: RAG (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F05-dotnet.md), [TypeScript](../typescript/F05-typescript.md). Lab overview: [F05-lab.md](../F05-lab.md).*

**The User Problem:** A visitor asks Trailhead Guides: "Can I have a campfire at Sperry Chalet in September?" The answer exists, in paragraph four of a 12-page backcountry regulations document that nobody will ever read. Search alone returns the document, not the answer. A plain chatbot answers fluently and makes it up, and a confidently wrong answer about fire regulations is worse than no answer at all.

*A Challenge lab. Do it if you finished [Module 2](../../M2-overview.md)'s Recommended lab and want another, or skip it without guilt. You will have seen this feature demonstrated either way.*

- **Goal:** answer questions from the park docs with citations you can verify, and refuse when the docs are silent.
- **Input:** `data/chunks.jsonl`, 250 chunks of park regulations; `data/questions.json`, four test questions, one unanswerable.
- **How:** embed the chunks and the question with Ollama, then rank the chunks by cosine similarity using feature 04's code. Paste the top 3 into a prompt and send it to a chat model. Then check every citation the model wrote against the chunks you gave it.
- **Model:** `nomic-embed-text` for the vectors, `llama3.2` for the answer. Both run locally without a key. The stretch goals swap in `gpt-4.1` on Azure.

## The Concept

RAG bolts feature 04's retrieval onto an LLM's generation. Instead of asking the model what it knows, you retrieve the most relevant chunks of your own documents and hand them over with the question: "Answer using only this context. If the context doesn't cover it, say so." The model becomes a reading assistant for your content rather than an oracle, and it can cite which document the answer came from.

The mechanics you'll touch: chunking (splitting 25 park docs into retrievable pieces), retrieval (feature 04's embedding search, plus a lexical signal it turns out to need), and grounded prompting (context in, citation out, refusal when the context is silent).

A third mechanic is easy to leave out and expensive to leave out: **the model has to be told what "now" means.** The park corpus is written the way operational documents are actually written, in dated notices ("Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice"). Ask "is the trail open right now?" and a model with no calendar cannot connect the two, so it refuses a question its documents answer twice over. The finished demo puts the current date in the prompt next to the refusal rule, and the refusal rate on that question drops from better than half to one run in twenty. Almost every real knowledge base is a corpus of dated notices, and a RAG system that never tells the model the date will either refuse answerable questions or answer them as of an unknown date, with nothing in the output to tell you which.

Each of those mechanics has a failure mode worth showing rather than glossing. **Chunking is a correctness decision.** Splitting the park docs one chunk per numbered section is the obvious default, and it put a conditional fire rule and the absolute exception that overrides it into the same 256-word chunk. Retrieval ranked that chunk first on every phrasing of the question, and the model read the conditional, stopped, and told the visitor a campfire was fine in 4 runs out of 20. Splitting oversized sections at their own subsection boundaries took that to 0 in 60, with the retrieval scores barely moving. **Embedding search alone is weak on proper nouns.** Blending a plain keyword score into the ranking, weighted by how rare each word is in the corpus, makes "Sperry" count for something; that blend is called hybrid retrieval and it is most production RAG systems' first upgrade. **And a citation is only a string the model typed.** Small models emit chunk ids that look right and point nowhere. Checking each cited id against the set you actually retrieved is a five-line function, and without it a citation proves nothing.

The model strategy is hybrid in a second sense, on purpose. Retrieval runs on free local embeddings, and you'll try generation both ways: a local model first, then Azure OpenAI. Watching the cloud model handle a multi-document answer more cleanly is the honest version of the "when do I pay for the big model" conversation from feature 03, now applied to generation.

Every step below is one thing to make the program do. The `starter/` is the plain chatbot. Edit it until it does all seven of steps 1 through 7, and compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Switching to Azure OpenAI later is a different constructor and nothing else. `starter/main.py` has one client, one question, and no context. `complete/main.py` is the finished demo as shown on stage: hybrid retrieval over `data/chunks.jsonl` (normalized cosine blended with a BM25-lite lexical score, alpha visible), the score table with both signals and the margin, the grounded prompt with the pinned date in the refusal clause, citation validation with one retry and then stripping, and generation on Azure OpenAI or the local model.

The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [SETUP.md](../../../../SETUP.md)), and `uv run` finds it from any folder. Run from the `starter/` folder:

```bash
uv run main.py
```

`starter/main.py` takes no flags, at most the one positional argument its header comment names (a question). `complete/` adds flags for the steps and stretch goals. From `complete/`:

```bash
uv run main.py                                    # the Sperry Chalet question, grounded
uv run main.py "Is the Avalanche Lake Trail open right now?"
uv run main.py --no-context                       # step 1: the confident wrong answer
uv run main.py --alpha 1.0 --top-k 8 --retrieval-only   # pure cosine: the wrong-park neighbors
uv run main.py --model qwen3:32b                  # a bigger local model, if you have the memory
```

Retrieval always runs locally on `nomic-embed-text`; the first run of `complete/` embeds 250 chunks, the slow part, and caches them to `embeddings.json` next to the script. Every number in the retrieval table matches the other tracks to four decimals. The flags shown for later steps are the ones `complete/` supports; in the starter, add the same argument parsing or hard-code the value.

You need flags only for the keyword-score stretch goal. To accept them in the starter, replace its `question = " ".join(sys.argv[1:]) or ...` line with this loop, copied from `complete/main.py`. `sys.argv[1:]` is everything typed after `main.py`; the loop walks it one word at a time, a flag that takes a value moves `i` forward one extra word to read it, and anything that is not a flag becomes part of the question:

```python
top_k = 3
alpha = 0.6            # weight on the semantic signal; 1.0 = cosine only
no_context = False
retrieval_only = False
question_parts: list[str] = []
args = sys.argv[1:]
i = 0
while i < len(args):
    a = args[i]
    if a == "--no-context":
        no_context = True
    elif a == "--retrieval-only":
        retrieval_only = True
    elif a == "--top-k":
        i += 1
        top_k = int(args[i])
    elif a == "--alpha":
        i += 1
        alpha = float(args[i])
    else:
        question_parts.append(a)
    i += 1
question = " ".join(question_parts) or "Can I have a campfire at Sperry Chalet in September?"
```

`--no-context` only sets a variable nothing reads yet; `retrieval_only`, `top_k`, and `alpha` are used in the keyword-score stretch goal.

### Step 0: Run the starter and read the wrong answer

**Do:** run `starter/` as it is. It sends this question to `llama3.2` with no documents attached:

```text
Can I have a campfire at Sperry Chalet in September?
```

```bash
uv run main.py
```

The starter already does this. It builds a client pointed at Ollama, takes the question from the command line (or falls back to the Sperry question), sends it, and prints the reply:

```python
client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")

question = " ".join(sys.argv[1:]) or "Can I have a campfire at Sperry Chalet in September?"

print(f"Q: {question}\n")
response = client.chat.completions.create(model="llama3.2", messages=[{"role": "user", "content": question}])
print(response.choices[0].message.content)
```

The code you add in steps 1 through 3 goes between the `print(f"Q: {question}\n")` line and the `response = ...` line, in order, each step below the one before. Step 4 replaces the last two lines.

Then open `data/park-docs/glacier-backcountry-camping-guide.md`, Section 4.2, and read the real rule.

**Why:** `data/park-docs/` is 25 markdown documents covering six parks (Acadia, Glacier, Great Smoky Mountains, Rocky Mountain, Yosemite, Zion), each laid out like a real park document with a document number, effective date, and numbered sections. The park names are real; every rule in them is fiction. The rest of the lab closes the gap between what the model guesses and what the docs actually say.

**Check:** the answer is fluent, confident, and wrong about where Sperry Chalet is; the recorded runs place it in California and name a national forest as the managing agency, when it sits in Glacier National Park, Montana. On the fire rule it guesses or hedges about fire danger ratings instead of the year-round ban Section 4.2 actually states. Run it again and the wording changes while the confidence does not.

### Step 1: Load the chunks

**Do:**
1. Open `../../data/chunks.jsonl`. Every line is one JSON object with `chunk_id`, `source`, `text`. That is where the data folder sits relative to `starter/`; your track's block below says how to point at it.
2. Read it line by line, parse each line, and keep the results in a list. Name your parsed fields exactly `chunk_id`, `source`, and `text` so the JSON keys map without extra configuration.

Resolve the path from the script's own location, not the working directory. First, the imports. They go at the very top of `main.py`, next to the starter's `import sys`. `math` and `cache_path` below are for steps 2 and 3:

```python
import json
import math
from pathlib import Path
```

Next, rename the client. Replace the starter's `client = OpenAI(...)` line with this one. Until step 4 replaces it, also change `client.chat` to `ollama.chat` in the starter's `response = ...` line, or the program stops with `NameError: name 'client' is not defined`:

```python
ollama = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
```

`ollama` is the same `openai` client the starter already builds, pointed at Ollama; `complete/` renames it `ollama` because a second client appears in the cloud stretch goal.

Then the paths and the loading, below `print(f"Q: {question}\n")`. `Path(__file__)` is this script's own file, `.parents[2]` climbs two folders up from `starter/` to the feature folder, and `/ "data"` joins a folder name onto the path. The last line is a list comprehension: for every non-blank line of the file, `json.loads` turns that one line of JSON into a dictionary, so `chunks[0]["chunk_id"]` is the first chunk's id:

```python
DATA = Path(__file__).resolve().parents[2] / "data"
chunks_path = DATA / "chunks.jsonl"
cache_path = Path(__file__).with_name("embeddings.json")
chunks = [json.loads(line) for line in chunks_path.read_text().splitlines() if line.strip()]
```

To see the Check numbers, print them once and delete the line afterward:

```python
# Hint:
print(f"{len(chunks)} chunks, first is {chunks[0]['chunk_id']}")
```

**Why:** the chunks are the 25 park docs cut up by `data/build-chunks.py`: one chunk per numbered section, except a section over 200 words is split at its own subsection numbers, and consecutive subsections are packed together until a chunk clears a 50-word floor (five sections were long enough to split, turning 241 sections into 250 chunks). Every chunk opens with the document's title in square brackets and ends with a pointer to the next unit, marked `(continues in ...)`, so a rule is never retrieved without the exception that follows it. `chunk_id` is the file name, a colon, and the section number (`:00` is the header block, `:04.2` is Section 4.2, `:04.3-5` is subsections 4.3-4.5 packed together).

**Check:** the list has 250 entries. The first `chunk_id` is `acadia-campfire-and-campground-regulations:00`.

### Step 2: Embed every chunk once and cache the vectors

**Do:**
1. Loop over the list in batches of 32 chunks.
2. For each batch, embed the batch's `text` values in one call to `nomic-embed-text`.
3. Store each returned vector in a dictionary keyed by that chunk's `chunk_id`.
4. When the loop finishes, write the dictionary to `embeddings.json` next to your program; check for that file at startup and skip the loop if it exists.

The embedding call is the same `openai` client, pointed at Ollama. You pass a list of strings as `input`, and each returned `data[i].embedding` is the vector (a list of numbers) for `texts[i]`, in the same order. Wrap it in a function; Python must see a `def` before the line that calls it, so put this directly below the step 1 lines:

```python
def embed(texts: list[str]) -> list[list[float]]:
    return [d.embedding for d in ollama.embeddings.create(model="nomic-embed-text", input=texts).data]
```

Then the cache and the batch loop, directly below the function. `range(0, len(chunks), 32)` counts 0, 32, 64, and so on; `chunks[start:start + 32]` slices out the next 32 (the last batch is shorter); `[c["text"] for c in batch]` pulls out just their texts; and `zip` pairs each chunk in the batch with its vector, in order. `json.dumps` turns the dictionary into text for the file, and `json.loads` turns it back:

```python
if cache_path.exists():
    index = json.loads(cache_path.read_text())
else:
    print(f"[embedding {len(chunks)} chunks, one-time; caching to {cache_path.name}]")
    index = {}
    for start in range(0, len(chunks), 32):
        batch = chunks[start:start + 32]
        for chunk, vector in zip(batch, embed([c["text"] for c in batch])):
            index[chunk["chunk_id"]] = vector
    cache_path.write_text(json.dumps(index))
```

For the Check, print the size once:

```python
# Hint:
print(f"{len(index)} keys, {len(index[chunks[0]['chunk_id']])} numbers each")
```

For the shortcut in the Why below, change only the path in step 1's `cache_path = ...` line; the `if cache_path.exists():` branch then loads the shipped file and the loop never runs:

```python
# Hint:
cache_path = DATA / "chunk-embeddings.json"
```

**Why:** re-embedding 250 chunks on every run wastes the slowest part of the program. `embeddings.json` is your program's own cache. Shortcut if you want to reach the RAG part faster: load `../../data/chunk-embeddings.json` instead. It is the same dictionary, already computed and shipped with the workshop.

**Check:** 250 keys, each holding 768 numbers. The first run is the slow one and the second is instant, which is the whole point of the cache. How slow depends on your machine (mine has taken anywhere from a few seconds to about 40).

### Step 3: Embed the question and rank the chunks

**Do:**
1. Embed question 1 the same way, as a single string. Keep the one vector that comes back.
2. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`.
3. For every chunk, compute the cosine between the question vector and that chunk's vector.
4. Sort by score, highest first, keep the top 3.
5. Print each of the 3 as score and `chunk_id`.

Feature 04's cosine similarity, as a function. `zip(a, b)` walks the two vectors side by side, so `dot` adds up each pair multiplied together, and each `math.sqrt(sum(...))` is one vector's length. Put it directly below the step 2 lines:

```python
def cosine_similarity(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))
```

Then the ranking, directly below the function. `question` is the variable the starter already builds (shown at step 0), and `embed([question])[0]` sends a list of one string and keeps its one vector. The `sorted(...)` line builds a `(score, chunk)` pair for every chunk and sorts the pairs by score; `key=lambda x: -x[0]` sorts on the negative score, which puts the highest first. `scored[:3]` is the first three pairs:

```python
question_vector = embed([question])[0]
scored = sorted(((cosine_similarity(question_vector, index[c["chunk_id"]]), c) for c in chunks), key=lambda x: -x[0])
for score, chunk in scored[:3]:
    print(f"{score:.4f}  {chunk['chunk_id']}")
```

**Check:** rank 1 is `glacier-backcountry-camping-guide:04.2` at cosine `0.7422`, rank 2 is `glacier-frontcountry-campground-regulations:04.1-2` at `0.6913`, rank 3 is `yosemite-campfire-regulations:03` at `0.6812`. Anything else at rank 1 means stop here, because no prompt later can fix retrieval. `complete/` prints a blended score by default; the raw cosine is the number in parentheses, or run it with `--alpha 1.0` to see only these.

### Step 4: Build the grounded prompt and get an answer

**Do:**
1. Turn the 3 chunks into a context block: for each chunk, write `chunk_id: `, `source: `, then the text, separated by blank lines.
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

3. Send the whole prompt as a single user message to `llama3.2`, the same call the starter already makes.
4. Print the answer.

The chat call is the one the starter already makes; `complete/` wraps it in `generate()` so the same function can serve the retry in step 5 and the cloud swap in the stretch goal. Until you reach that stretch goal, `chat_client` is the Ollama client and `chat_model` is `local_model`. Put these lines directly below the step 3 lines. `response.choices[0].message.content` is the reply text, and `or ""` turns a missing reply into an empty string:

```python
local_model = "llama3.2"
chat_client, chat_model = ollama, local_model


def generate(prompt: str) -> str:
    response = chat_client.chat.completions.create(model=chat_model, messages=[{"role": "user", "content": prompt}])
    return response.choices[0].message.content or ""
```

Next, the context block (item 1), directly below `generate`. `TODAY` is a constant; production would pass `date.today()`, and the lab pins a value so the recorded outputs stay reproducible. `top` is the three `(score, chunk)` pairs from step 3; `for _, c in top` unpacks each pair and ignores the score, and `"\n\n".join(...)` glues the three pieces together with a blank line between them:

```python
TODAY = "September 23, 2026"
REFUSAL = "The provided documents don't say."
top = scored[:3]
context = "\n\n".join(f"chunk_id: {c['chunk_id']}\nsource: {c['source']}\n{c['text']}" for _, c in top)
```

Then the prompt and the call (items 2 and 3), directly below. A triple-quoted `f"""..."""` string keeps the quotes and line breaks, and each `{name}` is filled in with that variable's value. Paste it exactly, starting at the left margin, because any spaces you add in front of a line are sent to the model too:

```python
prompt = f"""You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "{REFUSAL}"
  A question about "right now" is answered from the context, not refused: today is {TODAY},
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
{context}

Question: {question}

Answer:
"""
answer = generate(prompt)
```

Finally (item 4), delete the starter's `response = ...` line and replace its `print(response.choices[0].message.content)` line with this. Step 5 moves it below the citation check:

```python
print(answer)
```

**Why:** two details of that prompt were settled by measurement, written up in `expected-output.md` under "Telling the Model What Day It Is". The date sits inside the refusal rule rather than standing as its own rule, because a date at the top of the prompt changed nothing (6 refusals in 16 runs against a baseline of 5), and a broadly worded currency rule in the Rules block fixed the trail question while dropping the Sperry campfire answer from 22 of 24 correct to 14 of 24. The refusal is one exact sentence rather than a general instruction so that your code can recognize a refusal when it sees one.

**Check:** the answer says no, wood fires are banned year-round at Sperry Chalet, and cites `[glacier-backcountry-camping-guide:04.2]`. An answer that opens "Yes" and then says fires are prohibited still passes. The failure is `You can have a campfire at Sperry Chalet in September, but only pressurized-gas stoves are permitted...` with no citation.

### Step 5: Check the citations against the chunks you sent

**Do:**
1. Make a set of the 3 `chunk_id` values in the context.
2. Find every `[...]` in the answer whose contents include a colon; split each on commas, trim, keep the pieces that still contain a colon. Those are the citations the model wrote.
3. Compare each against your set. For any not in the set, print `!! CITATION CHECK FAILED: [the id] not in the retrieved set`.
4. After the answer, print `[citations: N valid (the ids), M invalid]`, counting each distinct id once, so a repeated citation does not inflate N.

Add this import at the very top of `main.py`, next to the others:

```python
import re
```

The helper that finds the citations (item 2) goes directly below `answer = generate(prompt)`. The regex `\[([^\]]*:[^\]]*)\]` matches a `[`, then any run of characters that contains a colon, then `]`, and `m.group(1)` is the part between the brackets. `split(",")` breaks a list like `[a:01, b:02]` into pieces, and `strip()` trims the spaces:

```python
def citations(text: str) -> list[str]:
    out = []
    for m in re.finditer(r"\[([^\]]*:[^\]]*)\]", text):
        out.extend(c.strip() for c in m.group(1).split(",") if ":" in c)
    return out
```

The rest goes directly below the helper, and its `print(answer)` is step 4's line moved down, so delete the old one. `{... for _, c in top}` builds a set (item 1), `bad` keeps each cited id that is not in it (item 3), and `cited` keeps the ones that are (item 4):

```python
retrieved_ids = {c["chunk_id"] for _, c in top}
bad = [c for c in dict.fromkeys(citations(answer)) if c not in retrieved_ids]
if bad:
    print(f"!! CITATION CHECK FAILED: {', '.join(f'[{c}]' for c in bad)} not in the retrieved set")

print(answer)

cited = list(dict.fromkeys(c for c in citations(answer) if c in retrieved_ids))
print(f"\n[citations: {len(cited)} valid ({', '.join(cited)}), {len(bad)} invalid]")
```

`dict.fromkeys` keeps the first occurrence of each id and drops repeats, so a citation the model wrote twice counts once in both lists.

**Why:** the model sometimes writes an id that looks real but was never in the context, most often stapled to its own refusal; on the EV charging question it sometimes cites `glacier-bear-safety-advisory:02` when the chunk you sent was `:03` (I have seen that in as many as 8 runs out of 20, and as few as 1). Nothing in the answer text tells you that, and only this check does.

**Check:** on question 1 the summary says 1 valid, 0 invalid. Run the EV charging question from step 6 a few times and you'll see the model attach an invented id to its own refusal; a made-up id reaching the output unflagged is the failure.

### Step 6: Run all four questions

**Do:** run the program once per question:

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

Pass each question as the argument, for example:

```bash
uv run main.py "Is the Avalanche Lake Trail open right now?"
```

The quotes keep the question together as one argument. The starter already does this. `sys.argv[1:]` holds the words after `main.py`, and with no argument the question falls back to question 1:

```python
question = " ".join(sys.argv[1:]) or "Can I have a campfire at Sperry Chalet in September?"
```

If you added the flag loop from Running, its last line does the same job.

**Why:** these come from `data/questions.json`, four objects each with `id`, `question`, an `answerable` flag, and `answer_lives_in` (the document/section the answer sits in, or `nowhere` for question 4).

**Check:** question 2 says eight and cites `[glacier-backcountry-permit-regulations:04]`. Question 3 says closed effective June 20, 2026, citing `glacier-seasonal-closures-2026:04.1` or `glacier-visitor-faq:02`. Question 4 should refuse, usually with the exact sentence `The provided documents don't say.` The refusal is the point; wording that drifts from it still passes as long as the model does not invent a charger. If question 3 refuses, the date line from step 4 is missing or moved. `No, fuel is not available anywhere within the Park` on question 4 is a miss. The model answered the question next door.

### Step 7: Run question 1 twenty times

**Do:** from `starter/`:

```bash
for i in $(seq 20); do uv run main.py 2>/dev/null | tail -3; done
```

The last three lines of each run are the answer, a blank line, and the `[citations: ...]` summary from step 5.

**Why:** the answer to this question varies run to run, so one run tells you what the program did once, not how often it does it. Looping twenty times turns "it worked when I tried it" into a rate you can quote, which is what the Check below asks you to count.

**Check:** the answer lands on no, and the citation is a real id. Count how many open with "Yes" before getting to the ban; when I measured it on `llama3.2` the rate was about 40 percent, and it is not a retrieval bug. Expect a run here and there to cite nothing usable, too (your numbers will move around; the model is non-deterministic). What fails is a run that never reaches the ban at all. The failure to watch for: `campfires are permitted at Sperry Chalet area (site code SPE) when the posted fire danger rating is below Very High`.

### Stretch goals

Pick any. Each one is already built in `complete/`, and the measurements that justify it are in [`expected-output.md`](../expected-output.md).

- **Add a keyword score.** Cosine alone treats "campfire regulations" from five parks as nearly the same thing and barely notices the word "Sperry". Fix that with a second score: count the question's words in each chunk, weight each word by how few chunks contain it (a rare word like "Sperry" counts for more than a common one like "campfire"), rescale both scores to 0..1, and blend `0.6 * cosine + 0.4 * keyword`. Print the top 8 both ways. The full tokenizer, stop-word list, and BM25 formula are in `complete/`; the shape is below.

  This stretch goal needs the `--top-k`, `--alpha`, and `--retrieval-only` flags, so add the flag loop from Running first if you have not. Then one more import at the very top of `main.py`, next to step 5's `import re`. `Counter` is a dictionary that counts things:

  ```python
  from collections import Counter
  ```

  First, a tokenizer: lowercase, split into words, drop short and filler words, and trim a plural `s`. Copy it and its stop-word list from `complete/main.py` exactly; a shorter tokenizer changes which chunks win, and the Check below will not match. `re.findall(r"[a-z0-9]+", ...)` returns every run of letters and digits as a list of words. Both go directly below step 3's `cosine_similarity` function, above `question_vector = ...`:

  ```python
  STOP_WORDS = {
      "the", "and", "for", "are", "but", "not", "you", "your", "with", "that", "this", "these",
      "those", "from", "have", "has", "had", "was", "were", "been", "being", "can", "could",
      "will", "would", "shall", "should", "may", "might", "must", "does", "did", "doing",
      "what", "when", "where", "which", "who", "whom", "why", "how", "any", "all", "some",
      "there", "here", "then", "than", "them", "they", "their", "its", "his", "her", "our",
      "get", "got", "still", "now", "right", "just", "about", "into", "onto", "over", "under",
      "out", "off", "per", "via", "one", "two", "also", "more", "most", "much", "many", "each",
      "other", "such", "only", "own", "same", "too", "very", "let", "need", "want",
  }


  def tokenize(text: str) -> list[str]:
      out = []
      for t in re.findall(r"[a-z0-9]+", text.lower()):
          if len(t) > 2 and t not in STOP_WORDS:
              out.append(t[:-1] if len(t) > 3 and t.endswith("s") and not t.endswith("ss") else t)
      return out
  ```

  The rescaling helper goes directly below `tokenize`. It maps the lowest score in a dictionary to 0 and the highest to 1; the dictionary comprehension `{k: ... for k, v in raw.items()}` builds a new dictionary with the same keys:

  ```python
  def min_max(raw: dict[str, float]) -> dict[str, float]:
      lo, hi = min(raw.values()), max(raw.values())
      span = hi - lo
      return {k: (v - lo) / span if span > 1e-9 else 0.0 for k, v in raw.items()}
  ```

  Next, turn step 3's ranking into a dictionary of cosine scores keyed by `chunk_id`, so there is something to blend with. Keep `question_vector = ...`, and replace step 3's `scored = sorted(...)` line and its two-line `for` loop with:

  ```python
  cosine = {c["chunk_id"]: cosine_similarity(question_vector, index[c["chunk_id"]]) for c in chunks}
  ```

  Then count how many chunks contain each word. These lines go directly below. `tokenized` holds each chunk's word list, `avg_length` is the average list length, and `doc_freq.update(set(terms))` adds 1 for each distinct word in a chunk, so a word repeated inside one chunk still counts that chunk once:

  ```python
  tokenized = {c["chunk_id"]: tokenize(c["text"]) for c in chunks}
  avg_length = sum(len(t) for t in tokenized.values()) / len(tokenized)
  doc_freq: Counter[str] = Counter()
  for terms in tokenized.values():
      doc_freq.update(set(terms))
  ```

  Then the rare-word weight (`idf`) for each word of the question, directly below. `K1` and `B` are the two BM25 tuning constants, and `list(dict.fromkeys(...))` drops repeated words while keeping their order:

  ```python
  K1, B = 1.2, 0.3
  n = len(chunks)
  query_terms = list(dict.fromkeys(tokenize(question)))
  idf = {t: math.log(1 + (n - doc_freq[t] + 0.5) / (doc_freq[t] + 0.5)) for t in query_terms}
  ```

  Then the keyword score for every chunk (`lexical`), directly below. `counts` says how many times each word appears in this chunk, so `tf` is the count for one question word, or `None` when the chunk does not contain it, and `continue` skips to the next word:

  ```python
  lexical: dict[str, float] = {}
  for c in chunks:
      terms = tokenized[c["chunk_id"]]
      counts = Counter(terms)
      score = 0.0
      for t in query_terms:
          tf = counts.get(t)
          if not tf:
              continue
          score += idf[t] * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * len(terms) / avg_length))
      lexical[c["chunk_id"]] = score
  ```

  Now the blend, directly below. Each chunk becomes a small dictionary carrying the chunk, both raw scores, both rescaled scores, and the blend; `sorted(..., key=lambda h: -h["combined"])` puts the highest blend first, and `top` is the first `top_k` of them:

  ```python
  semantic_norm = min_max(cosine)
  lexical_norm = min_max(lexical)

  scored = sorted(
      (
          {
              "chunk": c,
              "cosine": cosine[c["chunk_id"]],
              "semantic_norm": semantic_norm[c["chunk_id"]],
              "lexical": lexical[c["chunk_id"]],
              "lexical_norm": lexical_norm[c["chunk_id"]],
              "combined": alpha * semantic_norm[c["chunk_id"]] + (1 - alpha) * lexical_norm[c["chunk_id"]],
          }
          for c in chunks
      ),
      key=lambda h: -h["combined"],
  )
  top = scored[:top_k]
  ```

  Print the table and the rank-1 margin directly below, then stop early when `--retrieval-only` is set:

  ```python
  print(f"[retrieved top {top_k}]  combined = {alpha:.2f} * semantic + {1 - alpha:.2f} * lexical")
  print("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id")
  for r, h in enumerate(top, 1):
      print(f"  {r:>4}  {h['combined']:8.4f}   {h['semantic_norm']:5.3f} ({h['cosine']:.4f})   {h['lexical_norm']:5.3f} ({h['lexical']:5.2f})   {h['chunk']['chunk_id']}")
  margin = scored[0]["combined"] - scored[1]["combined"] if len(scored) > 1 else 0
  print(f"  margin over rank 2: {margin:.4f}\n")

  if retrieval_only:
      raise SystemExit
  ```

  Steps 4 and 5 still expect `top` to hold `(score, chunk)` pairs, and now it holds dictionaries. Delete step 4's `top = scored[:3]` line (the new `top` already has `top_k` entries), and replace step 4's `context = ...` line and step 5's `retrieved_ids = ...` line with these two, which read the chunk out of each dictionary:

  ```python
  context = "\n\n".join(f"chunk_id: {h['chunk']['chunk_id']}\nsource: {h['chunk']['source']}\n{h['chunk']['text']}" for h in top)
  retrieved_ids = {h["chunk"]["chunk_id"] for h in top}
  ```

  Compare the two in your program:

  ```bash
  uv run main.py --top-k 8 --alpha 1.0 --retrieval-only
  uv run main.py --top-k 8 --retrieval-only
  ```

  In `complete/`, compare the two with:

  ```bash
  uv run main.py --top-k 8 --alpha 1.0 --retrieval-only
  uv run main.py --top-k 8 --retrieval-only
  ```

  **Check:** the top 3 do not change, the rank-1 margin grows from 0.1630 to 0.2321, Acadia disappears from ranks 4 through 8, and ranks 4-6 become Glacier documents that name Sperry Chalet. Then try the rephrasings listed in `expected-output.md`.
- **Repair a bad citation.** When step 5 finds an invalid id, send the prompt again with an extra line that lists the 3 legal ids and asks for a rewrite. If the second answer is still wrong, replace the bad id with `invalid-citation-removed`. One exception: when the answer is the refusal sentence with a citation attached, just delete the citation and skip the retry. **Check:** no invalid id ever reaches the printed answer, and the refusal string comes back word for word.

  The check has to run twice now, once on the first answer and once on the retry, so give it a helper. Put it directly below step 5's `citations` function:

  ```python
  def invalid_citations(text: str, valid: set[str]) -> list[str]:
      return list(dict.fromkeys(c for c in citations(text) if c not in valid))
  ```

  The repair replaces step 5's two-line `if bad:` statement and sits above `print(answer)`. First the exception: a refusal with a citation attached. `REFUSAL in answer` is true when the refusal sentence appears anywhere in the answer, and the fix is to keep only the sentence:

  ```python
  if bad and REFUSAL in answer:
      print(f"!! CITATION CHECK FAILED: {', '.join(f'[{c}]' for c in bad)} not in the retrieved set")
      print("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed\n")
      answer = REFUSAL
      bad = []
  ```

  Then the retry, directly below, starting at the same indentation as that `if`. `prompt + f"""..."""` appends the extra instruction to the original prompt, and `chr(10).join(...)` puts each legal id on its own line (`chr(10)` is a line break, which an f-string cannot contain as `\n` inside `{}` before Python 3.12). If the second answer is still wrong, `answer.replace` swaps each bad id for the marker:

  ```python
  elif bad:
      print(f"!! CITATION CHECK FAILED: {', '.join(f'[{c}]' for c in bad)} not in the retrieved set")
      print("!! retrying once with the valid chunk_ids spelled out\n")
      retry_prompt = prompt + f"""

  Your previous answer cited {', '.join(f'[{c}]' for c in bad)}, which is not a real chunk_id.
  The only chunk_ids you may cite are, exactly:
  {chr(10).join('  ' + cid for cid in retrieved_ids)}
  Rewrite the answer using only those.
  """
      answer = generate(retry_prompt)
      bad = invalid_citations(answer, retrieved_ids)
      if bad:
          print(f"!! STILL INVALID after retry: {', '.join(f'[{c}]' for c in bad)}")
          print("!! stripping them; the answer below is unverified where the citation was removed\n")
          for c in bad:
              answer = answer.replace(c, "invalid-citation-removed")
  ```

  The four lines inside the triple-quoted string start at the left margin of your file, not indented like the code around them; indentation there would be sent to the model.
- **Point generation at the cloud.** Do the keyword-score stretch goal first: the Check below assumes the blended ranking. Build the chat client from the endpoint `https://trailhead-ai-workshop.openai.azure.com` and the deployment `gpt-4.1`, both written straight into the code, plus the key you paste over `<KEY FROM INSTRUCTOR>`. Fall back to `llama3.2` while the placeholder is still there. The key is handed out in the room. Retrieval stays local either way.

  No extra package: `AzureOpenAI` ships in the same `openai` package the starter imports, and it exposes the same `chat.completions.create` call, so `generate()` does not change. Only the client and model name it uses do. Retrieval keeps using `ollama`. Both imports go at the very top of `main.py`; the second one replaces the starter's `from openai import OpenAI` line:

  ```python
  import os

  from openai import AzureOpenAI, OpenAI
  ```

  The rest replaces the `chat_client, chat_model = ollama, local_model` line from step 4. Keep `local_model = "llama3.2"` directly above it, because the fallback uses it. the placeholder still starts with `<` until you paste over it, so the `if` picks Azure only once a real key is there:

  ```python
  endpoint = "https://trailhead-ai-workshop.openai.azure.com"
  key = "<KEY FROM INSTRUCTOR>"  # paste the room key between the quotes
  deployment = "gpt-4.1"
  if not key.startswith("<"):
      chat_client, chat_model = AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment
      print(f"[generation: Azure OpenAI, deployment '{deployment}']")
  else:
      chat_client, chat_model = ollama, local_model
      print(f"[generation: no room key pasted in, falling back to local {local_model}]")
  ```

  Paste the key from the room over `<KEY FROM INSTRUCTOR>`, between the quotes.

  With the key pasted in, the `[generation: ...]` line printed above the answer names the deployment.

  Then ask a question whose answer is spread across three chunks from two documents:

  ```text
  Our party of nine wants to camp at Sperry. Do we need a backcountry permit, can we all go on one permit, and can we have a wood fire there?
  ```

  ```bash
  uv run main.py "Our party of nine wants to camp at Sperry. Do we need a backcountry permit, can we all go on one permit, and can we have a wood fire there?"
  ```

  **Check:** retrieval returns `glacier-backcountry-camping-guide:04.2`, `glacier-backcountry-permit-regulations:04`, and `glacier-backcountry-camping-guide:03`. `gpt-4.1` says a permit is required, says nine is over the limit of eight so the group must split into separate permits, and says wood fires are banned at Sperry year-round with stoves only. It cites `glacier-backcountry-permit-regulations:04` and `glacier-backcountry-camping-guide:04.2`, with 0 invalid citations; some runs also cite `glacier-backcountry-camping-guide:03`. Measured over 5 runs: all 5 matched, and 2 of 5 cited all three chunks.
- **Build an evaluation loop.** Write ten more questions, each with the chunk_id that should win, then sweep the blend weight from 0 to 1 and record recall@3 at each setting (the share of questions whose correct chunk landed in the top 3). **Check:** a table, and an answer to whether the weight that wins on question 1 wins on the other ten. The rephrasings table in `expected-output.md` seeds the first four rows.

  This builds on the keyword-score stretch goal; do that one first. Move the ranking into a function that takes the question and the blend weight and returns the 3 winning ids. It reuses `tokenized`, `avg_length`, `doc_freq`, `n`, `K1`, and `B`, which do not depend on the question, and redoes everything that does. Fill in the `lexical` loop body from the stretch goal, using `q_terms` and `q_idf` in place of `query_terms` and `idf`:

  ```python
  # Hint:
  def top3_ids(q, a):
      q_vector = embed([q])[0]
      q_cosine = {c["chunk_id"]: cosine_similarity(q_vector, index[c["chunk_id"]]) for c in chunks}
      q_terms = list(dict.fromkeys(tokenize(q)))
      q_idf = {t: math.log(1 + (n - doc_freq[t] + 0.5) / (doc_freq[t] + 0.5)) for t in q_terms}
      q_lexical = {}
      for c in chunks:
          terms = tokenized[c["chunk_id"]]
          counts = Counter(terms)
          score = 0.0
          # <the "for t in query_terms:" loop from the stretch goal, with q_terms and q_idf>
          q_lexical[c["chunk_id"]] = score
      sem, lex = min_max(q_cosine), min_max(q_lexical)
      ranked = sorted(chunks, key=lambda c: -(a * sem[c["chunk_id"]] + (1 - a) * lex[c["chunk_id"]]))
      return [c["chunk_id"] for c in ranked[:3]]
  ```

  Then the table. Put the function and this loop directly above the stretch goal's `if retrieval_only:` line, and run with `--retrieval-only` so no answer is generated. `range(11)` counts 0 to 10, so `a` goes 0.0, 0.1, ... 1.0, and `sum(1 for ... if ...)` counts the questions whose expected id is in the top 3:

  ```python
  # Hint:
  eval_set = [
      ("Can I have a campfire at Sperry Chalet in September?", "glacier-backcountry-camping-guide:04.2"),
      # ten more ("question", "chunk_id that should win") rows
  ]
  for step in range(11):
      a = step / 10
      hits = sum(1 for q, expected_id in eval_set if expected_id in top3_ids(q, a))
      print(f"alpha {a:.1f}  recall@3 {hits / len(eval_set):.2f}")
  ```

  Each row embeds its question 11 times, once per weight, so with ten rows expect the table to take a minute or two.

## What Is in This Folder

- `data/park-docs/`: the corpus, 25 fictional park documents across six parks, described at step 0.
- `data/chunks.jsonl`: those documents cut into 250 chunks, one per line, each with `chunk_id`, `source`, and `text`. The chunking rule is spelled out at step 1.
- `data/chunk-embeddings.json`: the 250 `nomic-embed-text` vectors from step 2, keyed by `chunk_id`, in case you want to skip the embedding wait.
- `data/questions.json`: the four test questions, with an `answerable` flag and where each answer lives.
- `data/build-chunks.py`: the script that made `chunks.jsonl`. Not needed for the lab. Run it with a different word ceiling or floor (`python3 build-chunks.py out.jsonl 400 0`, for example) to change the chunking and see what breaks; the outcomes are already measured.
- `expected-output.md`: real retrieval scores and real answers for all four questions, plus the measurements behind every choice above: chunk size, the keyword blend, citation checking, and why the date is in the prompt.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
