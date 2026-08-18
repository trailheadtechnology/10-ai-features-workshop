# Python Demo for 05 RAG

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: one client, one question, no context: the confident wrong answer from step 1 of the demo.
- `complete/main.py`: the finished demo as shown on stage. Hybrid retrieval over `../data/chunks.jsonl` (normalized cosine blended with a BM25-lite lexical score, alpha visible), the score table with both signals and the margin, the grounded prompt with the pinned date in the refusal clause, citation validation with one retry and then stripping, and generation on Azure OpenAI or the local model.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py                                    # the Sperry Chalet question, grounded
python main.py "Is the Avalanche Lake Trail open right now?"
python main.py --no-context                       # step 1: the confident wrong answer
python main.py --alpha 1.0 --top-k 8 --retrieval-only   # pure cosine: the wrong-park neighbors
python main.py --model qwen3:32b                  # a bigger local model, if you have the memory
```

Retrieval always runs locally on `nomic-embed-text`; the first run embeds 250 chunks (about 40 seconds) and caches them to `embeddings.json` next to the script. Every number in the retrieval table matches the .NET demo to four decimals, and the measured story behind the chunking, the alpha, the date rule, and the 32B comparison is in [`../expected-output.md`](../expected-output.md) and [`../dotnet/F05-dotnet.md`](../dotnet/F05-dotnet.md).

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, the deployment name the feature uses, and the key handed out in the room) and the generation call switches to Azure OpenAI through the SDK's `AzureOpenAI` client; leave them unset and it runs against Ollama.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F05-lab.md`](../F05-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Confident Wrong Answer

No retrieval, no context. Ask the Sperry Chalet question and read the answer, then open `../data/park-docs/glacier-backcountry-camping-guide.md` Section 4.2 and read the actual rule. Everything that follows exists because of this gap.

Run:

```bash
python main.py
```

Check: Fluent, specific, and wrong about a fire regulation.

### Step 2: Embed the Chunks and Retrieve the Top 3 by Cosine (lab step 1)

Load `../data/chunks.jsonl` (250 chunks with `chunk_id`, `source`, `text`), embed them with `nomic-embed-text` in batches of 32, cache the vectors (about 40 seconds the first time), embed the question, and print the top 3 with scores. If retrieval does not find the right material here, no prompt later can save you.

```python
chunks = [json.loads(l) for l in (DATA / "chunks.jsonl").read_text().splitlines() if l.strip()]
index = {}
for start in range(0, len(chunks), 32):
    batch = chunks[start:start + 32]
    for chunk, vector in zip(batch, embed([c["text"] for c in batch])):
        index[chunk["chunk_id"]] = vector
question_vector = embed([question])[0]
scored = sorted(((cosine_similarity(question_vector, index[c["chunk_id"]]), c) for c in chunks), key=lambda x: -x[0])
for score, chunk in scored[:3]:
    print(f"{score:.4f}  {chunk['chunk_id']}")
```

Check: `glacier-backcountry-camping-guide:04.2` at rank 1 with cosine 0.7422, and the margin over rank 2 is small. Print the scores; the margin is the story.

### Step 3: Blend in a Lexical Score so "Sperry" Counts (lab step 2)

Cosine alone puts Acadia and Yosemite campfire sections in the top 8, because the embedder collapses "campfire regulations" from five parks onto nearly the same point. Add a BM25-lite score: tokenize, weight each query word by how few chunks contain it, min-max both signals to 0..1, and combine with an alpha. The full tokenizer, stop-word list, and BM25 formula are in `complete/`; the shape is below. The commands below use the flags `complete/` has; in your own copy, change the `alpha` and top-k variables by hand.

```python
idf = {t: math.log(1 + (n - doc_freq[t] + 0.5) / (doc_freq[t] + 0.5)) for t in query_terms}
# lexical[c] = sum over query terms of idf[t] * tf*(K1+1) / (tf + K1*(1 - B + B*len/avg_len))
combined = alpha * semantic_norm[cid] + (1 - alpha) * lexical_norm[cid]
```

Run:

```bash
python main.py --top-k 8 --alpha 1.0 --retrieval-only
python main.py --top-k 8 --retrieval-only
```

Check: At alpha 1.0, five of eight chunks are from the wrong park. At the default 0.6, the wrong-park chunks are replaced by Glacier documents that name Sperry and the margin over rank 2 grows from 0.16 to 0.23. Try the rephrasings listed in `../expected-output.md`.

### Step 4: Build the Grounded Prompt and Generate (lab step 3)

Context in, citations out, refusal when the context is silent. Two details are load-bearing and both are measured in `../expected-output.md`: the exact refusal string, and today's date inside the refusal clause (not in the rules block), so "is the trail open right now?" is answered from a dated notice rather than refused.

```python
TODAY = "September 23, 2026"
REFUSAL = "The provided documents don't say."
context = "\n\n".join(f"chunk_id: {c['chunk_id']}\nsource: {c['source']}\n{c['text']}" for _, c in top)
prompt = f"""You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context.
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

Check: The right answer (no wood fires at Sperry, year-round) with `[glacier-backcountry-camping-guide:04.2]` cited. Ask "Is the Avalanche Lake Trail open right now?" too; without the date line it was refused in 10 runs out of 18.

### Step 5: Validate the Citations (lab step 4)

A citation is only a string the model typed. Pull every bracketed token with a colon out of the answer and check it against the ids you actually retrieved. Fail loudly on a mismatch: `complete/` retries once with the legal ids spelled out and then strips whatever is still wrong, so a bad receipt never reaches the visitor.

```python
def citations(text):
    out = []
    for m in re.finditer(r"\[([^\]]*:[^\]]*)\]", text):
        out.extend(c.strip() for c in m.group(1).split(",") if ":" in c)
    return out

retrieved_ids = {c["chunk_id"] for _, c in top}
bad = [c for c in dict.fromkeys(citations(answer)) if c not in retrieved_ids]
if bad:
    print(f"!! CITATION CHECK FAILED: {', '.join(bad)} not in the retrieved set")
```

Check: Run the unanswerable question a few times: the model will eventually attach an invented chunk id to its own refusal, and the check catches it. Every run should end with a `[citations: N valid, M invalid]` line.

### Step 6: Run All Four Questions, Then Run Question 1 Twenty Times (lab steps 5 and 6)

`../data/questions.json` has three answerable questions and one that is not. Then loop the Sperry question: a wrong answer one run in five is invisible in a single run and is the only defect in this feature that could hurt somebody.

Run:

```bash
for i in $(seq 20); do python main.py | tail -3 | head -1; done
```

Check: Three correct cited answers, a refusal on the fourth, and no invalid citation reaching the output unflagged. Over 20 runs, count how many open with "Yes" before saying fires are banned; the measured rate for `llama3.2` is 40%, and it is 0% for the 32B model at 18x the latency (`--model qwen3:32b`, if you have the memory). Stretch: write ten more questions with the chunk that should win, sweep alpha from 0 to 1, and defend your alpha with recall@3 rather than with the Sperry question.
