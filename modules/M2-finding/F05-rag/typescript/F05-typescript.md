<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 05: RAG (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F05-dotnet.md), [Python](../python/F05-python.md). Lab overview: [F05-lab.md](../F05-lab.md).*

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

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Switching to Azure OpenAI later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. `starter/index.ts` has one client, one question, and no context. `complete/index.ts` is the finished demo as shown on stage: hybrid retrieval over `data/chunks.jsonl` (normalized cosine blended with a BM25-lite lexical score, alpha visible), the score table with both signals and the margin, the grounded prompt with the pinned date in the refusal clause, citation validation with one retry and then stripping, and generation on Azure OpenAI or the local model.

Run everything from the `typescript/` folder, where `package.json` lives:

```bash
npm install                                            # once
npm run starter
```

`complete/` adds flags for the steps and stretch goals:

```bash
npm run complete                                       # the Sperry Chalet question, grounded
npm run complete -- "Is the Avalanche Lake Trail open right now?"
npm run complete -- --no-context                       # step 1: the confident wrong answer
npm run complete -- --alpha 1.0 --top-k 8 --retrieval-only   # pure cosine: the wrong-park neighbors
npm run complete -- --model qwen3:32b                  # a bigger local model, if you have the memory
```

Retrieval always runs locally on `nomic-embed-text`; the first run of `complete/` embeds 250 chunks (about 40 seconds) and caches them to `embeddings.json` next to the script. Every number in the retrieval table matches the other tracks to four decimals. The flags shown for later steps are the ones `complete/` supports; in the starter, add the same argument parsing or hard-code the value.

You need flags only for the keyword-score stretch goal. To accept them in the starter, replace its `const question = process.argv.slice(2).join(" ") || ...` line with this loop, copied from `complete/index.ts`. `process.argv.slice(2)` is everything typed after `npm run starter --`, one word per entry; the loop walks it one word at a time, a flag that takes a value reads the next word with `args[++i]` (which also moves `i` past it), and anything that is not a flag becomes part of the question:

```typescript
let topK = 3;
let alpha = 0.6;          // weight on the semantic signal; 1.0 = cosine only
let noContext = false;
let retrievalOnly = false;
const questionParts: string[] = [];
const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case "--no-context": noContext = true; break;
    case "--retrieval-only": retrievalOnly = true; break;
    case "--top-k": topK = parseInt(args[++i], 10); break;
    case "--alpha": alpha = parseFloat(args[++i]); break;
    default: questionParts.push(args[i]); break;
  }
}
const question = questionParts.join(" ") || "Can I have a campfire at Sperry Chalet in September?";
```

`noContext` only sets a variable nothing reads yet; `retrievalOnly`, `topK`, and `alpha` are used in the keyword-score stretch goal.

### Step 0: Run the starter and read the wrong answer

**Do:** run `starter/` as it is. It sends this question to `llama3.2` with no documents attached:

```text
Can I have a campfire at Sperry Chalet in September?
```

```bash
npm run starter
```

The starter already does this. It builds a client pointed at Ollama, takes the question from the command line (or falls back to the Sperry question), sends it, and prints the reply. `await` works at the top level of the file, outside any function, because `package.json` sets `"type": "module"`:

```typescript
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });

const question = process.argv.slice(2).join(" ") || "Can I have a campfire at Sperry Chalet in September?";

console.log(`Q: ${question}\n`);
const response = await client.chat.completions.create({ model: "llama3.2", messages: [{ role: "user", content: question }] });
console.log(response.choices[0].message.content);
```

The code you add in steps 1 through 3 goes between the ``console.log(`Q: ${question}\n`);`` line and the `const response = ...` line, in order, each step below the one before. Step 4 replaces the last two lines.

Then open `data/park-docs/glacier-backcountry-camping-guide.md`, Section 4.2, and read the real rule.

**Why:** `data/park-docs/` is 25 markdown documents covering six parks (Acadia, Glacier, Great Smoky Mountains, Rocky Mountain, Yosemite, Zion), each laid out like a real park document with a document number, effective date, and numbered sections. The park names are real; every rule in them is fiction. The rest of the lab closes the gap between what the model guesses and what the docs actually say.

**Check:** the answer is fluent, confident, and wrong about where Sperry Chalet is; the recorded runs place it in California and name a national forest as the managing agency, when it sits in Glacier National Park, Montana. On the fire rule it guesses or hedges about fire danger ratings instead of the year-round ban Section 4.2 actually states. Run it again and the wording changes while the confidence does not.

### Step 1: Load the chunks

**Do:**
1. Open `../../data/chunks.jsonl`. Every line is one JSON object with `chunk_id`, `source`, `text`. That is where the data folder sits relative to `starter/`; your track's block below says how to point at it.
2. Read it line by line, parse each line, and keep the results in a list. Name your parsed fields exactly `chunk_id`, `source`, and `text` so the JSON keys map without extra configuration.

`npm run starter` runs from `typescript/`, not from `starter/`, so a path relative to the working directory misses. Resolve the paths from the script's own folder instead. First, the imports. They go at the very top of `index.ts`, directly below the starter's `import OpenAI from "openai";`. `existsSync` and `writeFileSync` are for step 2:

```typescript
import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
```

Next, rename the client. Replace the starter's `const client = new OpenAI(...)` line with this one. Until step 4 replaces it, also change `client.chat` to `ollama.chat` in the starter's `const response = ...` line, or the program stops with `ReferenceError: client is not defined`:

```typescript
const ollama = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
```

`ollama` is the same `openai` client the starter already builds, pointed at Ollama; `complete/` renames it `ollama` because a second client appears in the cloud stretch goal.

Then the paths and the loading, below ``console.log(`Q: ${question}\n`);``. `import.meta.dirname` is the folder this file sits in (`typescript/starter/`), and `resolve(folder, "../../data")` climbs two folders up to the feature folder and into `data`. `cachePath` is for step 2. `type Chunk` describes one parsed line. The last line reads the whole file as text, `.split("\n")` cuts it into lines, `.filter((l) => l.trim())` drops blank lines, and `.map((l) => JSON.parse(l))` turns each line of JSON into an object, so `chunks[0].chunk_id` is the first chunk's id:

```typescript
const DATA = resolve(import.meta.dirname, "../../data");
const chunksPath = resolve(DATA, "chunks.jsonl");
const cachePath = resolve(import.meta.dirname, "embeddings.json");
type Chunk = { chunk_id: string; source: string; text: string };
const chunks: Chunk[] = readFileSync(chunksPath, "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
```

To see the Check numbers, print them once and delete the line afterward:

```typescript
// Hint:
console.log(`${chunks.length} chunks, first is ${chunks[0].chunk_id}`);
```

**Why:** the chunks are the 25 park docs cut up by `data/build-chunks.py`: one chunk per numbered section, except a section over 200 words is split at its own subsection numbers, and consecutive subsections are packed together until a chunk clears a 50-word floor (five sections were long enough to split, turning 241 sections into 250 chunks). Every chunk opens with the document's title in square brackets and ends with a pointer to the next unit, marked `(continues in ...)`, so a rule is never retrieved without the exception that follows it. `chunk_id` is the file name, a colon, and the section number (`:00` is the header block, `:04.2` is Section 4.2, `:04.3-5` is subsections 4.3-4.5 packed together).

**Check:** the list has 250 entries. The first `chunk_id` is `acadia-campfire-and-campground-regulations:00`.

### Step 2: Embed every chunk once and cache the vectors

**Do:**
1. Loop over the list in batches of 32 chunks.
2. For each batch, embed the batch's `text` values in one call to `nomic-embed-text`.
3. Store each returned vector in a dictionary keyed by that chunk's `chunk_id`.
4. When the loop finishes, write the dictionary to `embeddings.json` next to your program; check for that file at startup and skip the loop if it exists.

The embedding call is the same `openai` client, pointed at Ollama. You pass an array of strings as `input`, and each returned `data[i].embedding` is the vector (an array of numbers) for `texts[i]`, in the same order. Wrap it in a function; `async` means it can `await` the call, and `Promise<number[][]>` means callers `await` it to get an array of vectors. Put it directly below the step 1 lines:

```typescript
async function embed(texts: string[]): Promise<number[][]> {
  const response = await ollama.embeddings.create({ model: "nomic-embed-text", input: texts });
  return response.data.map((d) => d.embedding);
}
```

Then the cache and the batch loop, directly below the function. `Record<string, number[]>` is an object whose keys are chunk ids and whose values are vectors. The `for` loop counts `start` as 0, 32, 64, and so on; `chunks.slice(start, start + 32)` cuts out the next 32 (the last batch is shorter); `batch.map((c) => c.text)` pulls out just their texts; and `batch.forEach((c, i) => ...)` pairs each chunk with the vector at the same position. `JSON.stringify` turns the object into text for the file, and `JSON.parse` turns it back:

```typescript
let index: Record<string, number[]>;
if (existsSync(cachePath)) {
  index = JSON.parse(readFileSync(cachePath, "utf8"));
} else {
  console.log(`[embedding ${chunks.length} chunks, one-time; caching to embeddings.json]`);
  index = {};
  for (let start = 0; start < chunks.length; start += 32) {
    const batch = chunks.slice(start, start + 32);
    const vectors = await embed(batch.map((c) => c.text));
    batch.forEach((c, i) => { index[c.chunk_id] = vectors[i]; });
  }
  writeFileSync(cachePath, JSON.stringify(index));
}
```

For the Check, print the size once:

```typescript
// Hint:
console.log(`${Object.keys(index).length} keys, ${index[chunks[0].chunk_id].length} numbers each`);
```

For the shortcut in the Why below, change only the path in step 1's `const cachePath = ...` line; the `if (existsSync(cachePath))` branch then loads the shipped file and the loop never runs:

```typescript
// Hint:
const cachePath = resolve(DATA, "chunk-embeddings.json");
```

**Why:** re-embedding 250 chunks every run wastes ~40 seconds. `embeddings.json` is your program's own cache. Shortcut if you want to reach the RAG part faster: load `../../data/chunk-embeddings.json` instead. It is the same dictionary, already computed and shipped with the workshop.

**Check:** 250 keys, each holding 768 numbers. The first run takes about 40 seconds; the second run is instant.

### Step 3: Embed the question and rank the chunks

**Do:**
1. Embed question 1 the same way, as a single string. Keep the one vector that comes back.
2. Bring in cosine similarity from feature 04: `dot(a, b) / (length(a) * length(b))`.
3. For every chunk, compute the cosine between the question vector and that chunk's vector.
4. Sort by score, highest first, keep the top 3.
5. Print each of the 3 as score and `chunk_id`.

Feature 04's cosine similarity, as a function. The loop walks the two vectors side by side: `dot` adds up each pair multiplied together, and `magA` and `magB` add up each vector's squares, so `Math.sqrt(magA)` is one vector's length. Put it directly below the step 2 lines:

```typescript
function cosineSimilarity(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}
```

Then the ranking, directly below the function. `question` is the variable the starter already builds (shown at step 0), and `const [questionVector] = await embed([question]);` sends an array of one string and keeps its one vector. `.map` turns every chunk into a `{ chunk, score }` object, and `.sort((a, b) => b.score - a.score)` puts the highest score first. `scored.slice(0, 3)` is the first three, and `const { chunk, score } of ...` pulls both fields out of each one. `toFixed(4)` prints four decimals:

```typescript
const [questionVector] = await embed([question]);
const scored = chunks
  .map((c) => ({ chunk: c, score: cosineSimilarity(questionVector, index[c.chunk_id]) }))
  .sort((a, b) => b.score - a.score);
for (const { chunk, score } of scored.slice(0, 3)) console.log(`${score.toFixed(4)}  ${chunk.chunk_id}`);
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

The chat call is the one the starter already makes; `complete/` wraps it in `generate()` so the same function can serve the retry in step 5 and the cloud swap in the stretch goal. Until you reach that stretch goal, `chatClient` is the Ollama client and `chatModel` is `localModel`. Put these lines directly below the step 3 lines. `response.choices[0].message.content` is the reply text, and `?? ""` turns a missing reply into an empty string:

```typescript
let localModel = "llama3.2";
let chatClient: OpenAI = ollama;
let chatModel: string = localModel;

async function generate(prompt: string): Promise<string> {
  const response = await chatClient.chat.completions.create({ model: chatModel, messages: [{ role: "user", content: prompt }] });
  return response.choices[0].message.content ?? "";
}
```

Next, the context block (item 1), directly below `generate`. `TODAY` is a constant; production would pass `new Date()`, and the lab pins a value so the recorded outputs stay reproducible. `top` is the three `{ chunk, score }` objects from step 3; `.map((h) => ...)` turns each into its piece of text, and `.join("\n\n")` glues the three pieces together with a blank line between them. Inside backticks, each `${name}` is filled in with that variable's value:

```typescript
const TODAY = "September 23, 2026";
const REFUSAL = "The provided documents don't say.";
const top = scored.slice(0, 3);
const context = top.map((h) => `chunk_id: ${h.chunk.chunk_id}\nsource: ${h.chunk.source}\n${h.chunk.text}`).join("\n\n");
```

Then the prompt and the call (items 2 and 3), directly below. A backtick string keeps the quotes and line breaks. Paste it exactly, starting at the left margin, because any spaces you add in front of a line are sent to the model too:

```typescript
const prompt = `You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "${REFUSAL}"
  A question about "right now" is answered from the context, not refused: today is ${TODAY},
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
${context}

Question: ${question}

Answer:
`;
let answer = await generate(prompt);
```

Finally (item 4), delete the starter's `const response = ...` line and replace its `console.log(response.choices[0].message.content);` line with this. Step 5 moves it below the citation check:

```typescript
console.log(answer);
```

**Why:** two details of that prompt were settled by measurement, written up in `expected-output.md` under "Telling the Model What Day It Is". The date sits inside the refusal rule rather than standing as its own rule, because a date at the top of the prompt changed nothing (6 refusals in 16 runs against a baseline of 5), and a broadly worded currency rule in the Rules block fixed the trail question while dropping the Sperry campfire answer from 22 of 24 correct to 14 of 24. The refusal is one exact sentence rather than a general instruction so that your code can recognize a refusal when it sees one.

**Check:** the answer says no, wood fires are banned year-round at Sperry Chalet, and cites `[glacier-backcountry-camping-guide:04.2]`. An answer that opens "Yes" and then says fires are prohibited still passes. The failure is `You can have a campfire at Sperry Chalet in September, but only pressurized-gas stoves are permitted...` with no citation.

### Step 5: Check the citations against the chunks you sent

**Do:**
1. Make a set of the 3 `chunk_id` values in the context.
2. Find every `[...]` in the answer whose contents include a colon; split each on commas, trim, keep the pieces that still contain a colon. Those are the citations the model wrote.
3. Compare each against your set. For any not in the set, print `!! CITATION CHECK FAILED: [the id] not in the retrieved set`.
4. After the answer, print `[citations: N valid (the ids), M invalid]`, counting each distinct id once, so a repeated citation does not inflate N.

The helper that finds the citations (item 2) goes directly below `let answer = await generate(prompt);`. The regex `/\[([^\]]*:[^\]]*)\]/g` matches a `[`, then any run of characters that contains a colon, then `]`, and `m[1]` is the part between the brackets. `matchAll` finds every match, and `[...]` turns them into an array. `.flatMap((m) => m[1].split(","))` breaks a list like `[a:01, b:02]` into pieces and flattens them into one array, `trim()` removes the spaces, and `.filter` keeps the pieces that still contain a colon:

```typescript
function citations(text: string): string[] {
  return [...text.matchAll(/\[([^\]]*:[^\]]*)\]/g)]
    .flatMap((m) => m[1].split(","))
    .map((c) => c.trim())
    .filter((c) => c.includes(":"));
}
```

The rest goes directly below the helper, and its `console.log(answer);` is step 4's line moved down, so delete the old one. `new Set(...)` builds the set of retrieved ids (item 1) and `.has(c)` asks whether an id is in it. `invalidCitations` keeps each cited id that is not in the set (item 3); `let bad` stays `let` because the repair stretch goal reassigns it. `cited` keeps the ones that are in the set (item 4):

```typescript
const invalidCitations = (text: string, valid: Set<string>) => [...new Set(citations(text).filter((c) => !valid.has(c)))];
const retrievedIds = new Set(top.map((h) => h.chunk.chunk_id));
let bad = invalidCitations(answer, retrievedIds);
if (bad.length > 0) {
  console.log(`!! CITATION CHECK FAILED: ${bad.map((c) => `[${c}]`).join(", ")} not in the retrieved set`);
}

console.log(answer);

const cited = [...new Set(citations(answer).filter((c) => retrievedIds.has(c)))];
console.log(`\n[citations: ${cited.length} valid (${cited.join(", ")}), ${bad.length} invalid]`);
```

Spreading a `Set` drops repeats, so a citation the model wrote twice counts once in both lists.

**Why:** the model sometimes writes an id that looks real but was never in the context, most often stapled to its own refusal; on the EV charging question it cited `glacier-bear-safety-advisory:02` in 8 of 20 runs, when the chunk you sent was `:03`. Nothing in the answer text tells you that, and only this check does.

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
npm run starter -- "Is the Avalanche Lake Trail open right now?"
```

The `--` tells npm that the rest belongs to your program, and the quotes keep the question together. The starter already does this. `process.argv.slice(2)` holds the words after `--`, and with no argument the question falls back to question 1:

```typescript
const question = process.argv.slice(2).join(" ") || "Can I have a campfire at Sperry Chalet in September?";
```

If you added the flag loop from Running, its last line does the same job.

**Why:** these come from `data/questions.json`, four objects each with `id`, `question`, an `answerable` flag, and `answer_lives_in` (the document/section the answer sits in, or `nowhere` for question 4).

**Check:** question 2 says eight and cites `[glacier-backcountry-permit-regulations:04]`. Question 3 says closed effective June 20, 2026, citing `glacier-seasonal-closures-2026:04.1` or `glacier-visitor-faq:02`. Question 4 replies `The provided documents don't say.` and claims no charger. If question 3 refuses, the date line from step 4 is missing or moved. `No, fuel is not available anywhere within the Park` on question 4 is a miss. The model answered the question next door.

### Step 7: Run question 1 twenty times

**Do:** from `typescript/`. `--silent` keeps npm's own banner lines out of the last three:

```bash
for i in $(seq 20); do npm run --silent starter 2>/dev/null | tail -3; done
```

The last three lines of each run are the answer, a blank line, and the `[citations: ...]` summary from step 5.

**Why:** the answer to this question varies run to run, so one run tells you what the program did once, not how often it does it. Looping twenty times turns "it worked when I tried it" into a rate you can quote, which is what the Check below asks you to count.

**Check:** the answer lands on no, and the citation is a real id. Count how many open with "Yes" before getting to the ban; when I measured it on `llama3.2` the rate was about 40 percent, and it is not a retrieval bug. Expect a run here and there to cite nothing usable, too (your numbers will move around; the model is non-deterministic). What fails is a run that never reaches the ban at all. The failure to watch for: `campfires are permitted at Sperry Chalet area (site code SPE) when the posted fire danger rating is below Very High`.

### Stretch goals

Pick any. Each one is already built in `complete/`, and the measurements that justify it are in [`expected-output.md`](../expected-output.md).

- **Add a keyword score.** Cosine alone treats "campfire regulations" from five parks as nearly the same thing and barely notices the word "Sperry". Fix that with a second score: count the question's words in each chunk, weight each word by how few chunks contain it (a rare word like "Sperry" counts for more than a common one like "campfire"), rescale both scores to 0..1, and blend `0.6 * cosine + 0.4 * keyword`. Print the top 8 both ways. The full tokenizer, stop-word list, and BM25 formula are in `complete/`; the shape is below.

  This stretch goal needs the `--top-k`, `--alpha`, and `--retrieval-only` flags, so add the flag loop from Running first if you have not.

  First, a tokenizer: lowercase, split into words, drop short and filler words, and trim a plural `s`. Copy it and its stop-word list from `complete/index.ts` exactly; a shorter tokenizer changes which chunks win, and the Check below will not match. `new Set([...])` makes a list you can ask `.has(word)` of. `.match(/[a-z0-9]+/g)` returns every run of letters and digits as an array of words, or `null` when there are none, which `?? []` turns into an empty array. Both go directly below step 3's `cosineSimilarity` function, above `const [questionVector] = ...`:

  ```typescript
  const STOP_WORDS = new Set([
    "the", "and", "for", "are", "but", "not", "you", "your", "with", "that", "this", "these",
    "those", "from", "have", "has", "had", "was", "were", "been", "being", "can", "could",
    "will", "would", "shall", "should", "may", "might", "must", "does", "did", "doing",
    "what", "when", "where", "which", "who", "whom", "why", "how", "any", "all", "some",
    "there", "here", "then", "than", "them", "they", "their", "its", "his", "her", "our",
    "get", "got", "still", "now", "right", "just", "about", "into", "onto", "over", "under",
    "out", "off", "per", "via", "one", "two", "also", "more", "most", "much", "many", "each",
    "other", "such", "only", "own", "same", "too", "very", "let", "need", "want",
  ]);

  function tokenize(text: string): string[] {
    return (text.toLowerCase().match(/[a-z0-9]+/g) ?? [])
      .filter((t) => t.length > 2 && !STOP_WORDS.has(t))
      .map((t) => (t.length > 3 && t.endsWith("s") && !t.endsWith("ss") ? t.slice(0, -1) : t));
  }
  ```

  The rescaling helper goes directly below `tokenize`. It maps the lowest score in an object to 0 and the highest to 1. `Object.values` lists the scores, `Math.min(...values)` spreads them into one call, and `Object.fromEntries(Object.entries(raw).map(...))` builds a new object with the same keys:

  ```typescript
  function minMax(raw: Record<string, number>): Record<string, number> {
    const values = Object.values(raw);
    const lo = Math.min(...values), hi = Math.max(...values);
    const span = hi - lo;
    return Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, span > 1e-9 ? (v - lo) / span : 0]));
  }
  ```

  Next, turn step 3's ranking into an object of cosine scores keyed by `chunk_id`, so there is something to blend with. Keep `const [questionVector] = ...`, and replace step 3's three-line `const scored = ...` statement and its `for` line with:

  ```typescript
  const cosine: Record<string, number> = Object.fromEntries(chunks.map((c) => [c.chunk_id, cosineSimilarity(questionVector, index[c.chunk_id])]));
  ```

  Then count how many chunks contain each word. These lines go directly below. `tokenized` holds each chunk's word list, `avgLength` is the average list length (`reduce` adds up the lengths), and `docFreq` is a `Map` from word to chunk count. Looping over `new Set(terms)` visits each distinct word once, so a word repeated inside one chunk still counts that chunk once. `df(t)` reads the count, or 0 for a word no chunk contains:

  ```typescript
  const tokenized: Record<string, string[]> = Object.fromEntries(chunks.map((c) => [c.chunk_id, tokenize(c.text)]));
  const avgLength = Object.values(tokenized).reduce((sum, t) => sum + t.length, 0) / chunks.length;
  const docFreq = new Map<string, number>();
  for (const terms of Object.values(tokenized)) {
    for (const term of new Set(terms)) docFreq.set(term, (docFreq.get(term) ?? 0) + 1);
  }
  const df = (t: string) => docFreq.get(t) ?? 0;
  ```

  Then the rare-word weight (`idf`) for each word of the question, directly below. `K1` and `B` are the two BM25 tuning constants, and `[...new Set(tokenize(question))]` drops repeated words while keeping their order:

  ```typescript
  const K1 = 1.2, B = 0.3;
  const n = chunks.length;
  const queryTerms = [...new Set(tokenize(question))];
  const idf: Record<string, number> = Object.fromEntries(queryTerms.map((t) => [t, Math.log(1 + (n - df(t) + 0.5) / (df(t) + 0.5))]));
  ```

  Then the keyword score for every chunk (`lexical`), directly below. `counts` says how many times each word appears in this chunk, so `tf` is the count for one question word, or `undefined` when the chunk does not contain it, and `continue` skips to the next word:

  ```typescript
  const lexical: Record<string, number> = {};
  for (const c of chunks) {
    const terms = tokenized[c.chunk_id];
    const counts = new Map<string, number>();
    for (const t of terms) counts.set(t, (counts.get(t) ?? 0) + 1);
    let score = 0;
    for (const t of queryTerms) {
      const tf = counts.get(t);
      if (!tf) continue;
      score += idf[t] * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * terms.length / avgLength));
    }
    lexical[c.chunk_id] = score;
  }
  ```

  Now the blend, directly below. Each chunk becomes an object carrying the chunk, both raw scores, both rescaled scores, and the blend; `.sort((a, b) => b.combined - a.combined)` puts the highest blend first, and `top` is the first `topK` of them:

  ```typescript
  const semanticNorm = minMax(cosine);
  const lexicalNorm = minMax(lexical);

  const scored = chunks
    .map((c) => ({
      chunk: c,
      cosine: cosine[c.chunk_id],
      semanticNorm: semanticNorm[c.chunk_id],
      lexical: lexical[c.chunk_id],
      lexicalNorm: lexicalNorm[c.chunk_id],
      combined: alpha * semanticNorm[c.chunk_id] + (1 - alpha) * lexicalNorm[c.chunk_id],
    }))
    .sort((a, b) => b.combined - a.combined);
  const top = scored.slice(0, topK);
  ```

  Print the table and the rank-1 margin directly below, then stop early when `--retrieval-only` is set. `padStart(8)` pads a number with spaces on the left so the columns line up:

  ```typescript
  console.log(`[retrieved top ${topK}]  combined = ${alpha.toFixed(2)} * semantic + ${(1 - alpha).toFixed(2)} * lexical`);
  console.log("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id");
  top.forEach((h, r) => {
    console.log(`  ${String(r + 1).padStart(4)}  ${h.combined.toFixed(4).padStart(8)}   ${h.semanticNorm.toFixed(3).padStart(5)} (${h.cosine.toFixed(4)})   ${h.lexicalNorm.toFixed(3).padStart(5)} (${h.lexical.toFixed(2).padStart(5)})   ${h.chunk.chunk_id}`);
  });
  const margin = scored.length > 1 ? scored[0].combined - scored[1].combined : 0;
  console.log(`  margin over rank 2: ${margin.toFixed(4)}\n`);

  if (retrievalOnly) process.exit(0);
  ```

  Finally, delete step 4's `const top = scored.slice(0, 3);` line; the new `top` already has `topK` entries, and a second `const top` stops the program with `ERROR: The symbol "top" has already been declared`. Each new entry still has a `chunk` field, so step 4's `context` line and step 5's `retrievedIds` line work unchanged.

  Compare the two in your program:

  ```bash
  npm run starter -- --top-k 8 --alpha 1.0 --retrieval-only
  npm run starter -- --top-k 8 --retrieval-only
  ```

  In `complete/`, compare the two with:

  ```bash
  npm run complete -- --top-k 8 --alpha 1.0 --retrieval-only
  npm run complete -- --top-k 8 --retrieval-only
  ```

  **Check:** the top 3 do not change, the rank-1 margin grows from 0.1630 to 0.2321, Acadia disappears from ranks 4 through 8, and ranks 4-6 become Glacier documents that name Sperry Chalet. Then try the rephrasings listed in `expected-output.md`.
- **Repair a bad citation.** When step 5 finds an invalid id, send the prompt again with an extra line that lists the 3 legal ids and asks for a rewrite. If the second answer is still wrong, replace the bad id with `invalid-citation-removed`. One exception: when the answer is the refusal sentence with a citation attached, just delete the citation and skip the retry. **Check:** no invalid id ever reaches the printed answer, and the refusal string comes back word for word.

  The check has to run twice now, once on the first answer and once on the retry. Step 5's `invalidCitations` helper already does it, and step 5 declared `bad` with `let` so it can be reassigned. The repair replaces step 5's three-line `if (bad.length > 0) { ... }` statement and sits above `console.log(answer);`. First the exception: a refusal with a citation attached. `answer.includes(REFUSAL)` is true when the refusal sentence appears anywhere in the answer, and the fix is to keep only the sentence:

  ```typescript
  if (bad.length > 0 && answer.includes(REFUSAL)) {
    console.log(`!! CITATION CHECK FAILED: ${bad.map((c) => `[${c}]`).join(", ")} not in the retrieved set`);
    console.log("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed\n");
    answer = REFUSAL;
    bad = [];
  ```

  Then the retry, directly below, continuing the same `if`. ``prompt + `...` `` appends the extra instruction to the original prompt, and `[...retrievedIds].map((id) => "  " + id).join("\n")` puts each legal id on its own line. If the second answer is still wrong, `answer.replaceAll` swaps each bad id for the marker:

  ```typescript
  } else if (bad.length > 0) {
    console.log(`!! CITATION CHECK FAILED: ${bad.map((c) => `[${c}]`).join(", ")} not in the retrieved set`);
    console.log("!! retrying once with the valid chunk_ids spelled out\n");
    const retryPrompt = prompt + `

  Your previous answer cited ${bad.map((c) => `[${c}]`).join(", ")}, which is not a real chunk_id.
  The only chunk_ids you may cite are, exactly:
  ${[...retrievedIds].map((id) => "  " + id).join("\n")}
  Rewrite the answer using only those.
  `;
    answer = await generate(retryPrompt);
    bad = invalidCitations(answer, retrievedIds);
    if (bad.length > 0) {
      console.log(`!! STILL INVALID after retry: ${bad.map((c) => `[${c}]`).join(", ")}`);
      console.log("!! stripping them; the answer below is unverified where the citation was removed\n");
      for (const c of bad) answer = answer.replaceAll(c, "invalid-citation-removed");
    }
  }
  ```

  The four lines inside the backtick string after the blank line start at the left margin of your file, not indented like the code around them; indentation there would be sent to the model.
- **Point generation at the cloud.** Do the keyword-score stretch goal first: the Check below assumes the blended ranking. Build the chat client from `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` when they're set, falling back to `llama3.2` when they're not. The endpoint is `https://trailhead-ai-workshop.openai.azure.com`, the deployment is the name the feature uses, and the key is handed out in the room. Retrieval stays local either way.

  No extra package: `AzureOpenAI` ships in the same `openai` package the starter imports, and it exposes the same `chat.completions.create` call, so `generate()` does not change. Only the client and model name it uses do. Retrieval keeps using `ollama`. This import replaces the starter's `import OpenAI from "openai";` at the top of `index.ts`:

  ```typescript
  import OpenAI, { AzureOpenAI } from "openai";
  ```

  The rest replaces the `let chatClient: OpenAI = ollama;` and `let chatModel: string = localModel;` lines from step 4. Keep `let localModel = "llama3.2";` directly above it, because the fallback uses it. `process.env` holds the environment variables, and the first line copies three of them into `endpoint`, `key`, and `deployment` (each is `undefined` when not set). `if (endpoint && key && deployment)` is true only when all three have a value:

  ```typescript
  const { AZURE_OPENAI_ENDPOINT: endpoint, AZURE_OPENAI_KEY: key, AZURE_OPENAI_DEPLOYMENT: deployment } = process.env;
  let chatClient: OpenAI;
  let chatModel: string;
  if (endpoint && key && deployment) {
    chatClient = new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment });
    chatModel = deployment;
    console.log(`[generation: Azure OpenAI, deployment '${deployment}']`);
  } else {
    chatClient = ollama;
    chatModel = localModel;
    console.log(`[generation: AZURE_OPENAI_* not set, falling back to local ${localModel}]`);
  }
  ```

  Set the three variables in the same terminal before `npm run starter` (fill in the key from the room and the deployment name):

  ```bash
  export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
  export AZURE_OPENAI_KEY=<the key>
  export AZURE_OPENAI_DEPLOYMENT=<the deployment name>
  ```

  With the three variables exported, a `[generation: ...]` line printed before the answer names the deployment.

  Then ask a question whose answer is spread across three chunks from two documents:

  ```text
  Our party of nine wants to camp at Sperry. Do we need a backcountry permit, can we all go on one permit, and can we have a wood fire there?
  ```

  ```bash
  npm run starter -- "Our party of nine wants to camp at Sperry. Do we need a backcountry permit, can we all go on one permit, and can we have a wood fire there?"
  ```

  **Check:** retrieval returns `glacier-backcountry-camping-guide:04.2`, `glacier-backcountry-permit-regulations:04`, and `glacier-backcountry-camping-guide:03`. `gpt-4.1` says a permit is required, says nine is over the limit of eight so the group must split into separate permits, and says wood fires are banned at Sperry year-round with stoves only. It cites `glacier-backcountry-permit-regulations:04` and `glacier-backcountry-camping-guide:04.2`, with 0 invalid citations; some runs also cite `glacier-backcountry-camping-guide:03`. Measured over 5 runs: all 5 matched, and 2 of 5 cited all three chunks.
- **Build an evaluation loop.** Write ten more questions, each with the chunk_id that should win, then sweep the blend weight from 0 to 1 and record recall@3 at each setting (the share of questions whose correct chunk landed in the top 3). **Check:** a table, and an answer to whether the weight that wins on question 1 wins on the other ten. The rephrasings table in `expected-output.md` seeds the first four rows.

  This builds on the keyword-score stretch goal; do that one first. Move the ranking into a function that takes the question and the blend weight and returns the 3 winning ids. It reuses `tokenized`, `avgLength`, `df`, `n`, `K1`, and `B`, which do not depend on the question, and redoes everything that does. Fill in the `lexical` loop body from the stretch goal, using `qTerms` and `qIdf` in place of `queryTerms` and `idf`:

  ```typescript
  // Hint:
  async function top3Ids(q: string, a: number): Promise<string[]> {
    const [qVector] = await embed([q]);
    const qCosine: Record<string, number> = Object.fromEntries(chunks.map((c) => [c.chunk_id, cosineSimilarity(qVector, index[c.chunk_id])]));
    const qTerms = [...new Set(tokenize(q))];
    const qIdf: Record<string, number> = Object.fromEntries(qTerms.map((t) => [t, Math.log(1 + (n - df(t) + 0.5) / (df(t) + 0.5))]));
    const qLexical: Record<string, number> = {};
    for (const c of chunks) {
      const terms = tokenized[c.chunk_id];
      const counts = new Map<string, number>();
      for (const t of terms) counts.set(t, (counts.get(t) ?? 0) + 1);
      let score = 0;
      // <the "for (const t of queryTerms)" loop from the stretch goal, with qTerms and qIdf>
      qLexical[c.chunk_id] = score;
    }
    const sem = minMax(qCosine), lex = minMax(qLexical);
    const blend = (c: Chunk) => a * sem[c.chunk_id] + (1 - a) * lex[c.chunk_id];
    const ranked = [...chunks].sort((x, y) => blend(y) - blend(x));
    return ranked.slice(0, 3).map((c) => c.chunk_id);
  }
  ```

  Then the table. Put the function and this loop directly above the stretch goal's `if (retrievalOnly) process.exit(0);` line, and run with `--retrieval-only` so no answer is generated. `step` counts 0 to 10, so `a` goes 0.0, 0.1, ... 1.0, and `hits` counts the questions whose expected id is in the top 3:

  ```typescript
  // Hint:
  const evalSet: [string, string][] = [
    ["Can I have a campfire at Sperry Chalet in September?", "glacier-backcountry-camping-guide:04.2"],
    // ten more ["question", "chunk_id that should win"] rows
  ];
  for (let step = 0; step <= 10; step++) {
    const a = step / 10;
    let hits = 0;
    for (const [q, expectedId] of evalSet) {
      if ((await top3Ids(q, a)).includes(expectedId)) hits++;
    }
    console.log(`alpha ${a.toFixed(1)}  recall@3 ${(hits / evalSet.length).toFixed(2)}`);
  }
  ```

  Each row embeds its question 11 times, once per weight, so with ten rows expect the table to take a minute or two.

## What Is in This Folder

- `data/park-docs/`: the corpus, 25 fictional park documents across six parks, described at step 0.
- `data/chunks.jsonl`: those documents cut into 250 chunks, one per line, each with `chunk_id`, `source`, and `text`. The chunking rule is spelled out at step 1.
- `data/chunk-embeddings.json`: the 250 `nomic-embed-text` vectors from step 2, keyed by `chunk_id`, in case you want to skip the 40 seconds.
- `data/questions.json`: the four test questions, with an `answerable` flag and where each answer lives.
- `data/build-chunks.py`: the script that made `chunks.jsonl`. Not needed for the lab. Run it with a different word ceiling or floor (`python3 build-chunks.py out.jsonl 400 0`, for example) to change the chunking and see what breaks; the outcomes are already measured.
- `expected-output.md`: real retrieval scores and real answers for all four questions, plus the measurements behind every choice above: chunk size, the keyword blend, citation checking, and why the date is in the prompt.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
