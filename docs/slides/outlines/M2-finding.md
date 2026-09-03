# Deck 2: Module 2, Finding (about 90 minutes)

Demo-centric rhythm: name the three features and the one idea under them, then per feature — divider, DEMO, the lab, the leadership card. Concept slides only where a picture earns it: the shared embeddings pipeline, 04's index/search flow, and 05's RAG flow. Everything J. says lives in Notes; demo scripts live under each "What to Watch" slide.

Runsheet (lengths, not clock times):
- 30 min instructor: opener 4 min (This Module + the Same First Half diagram), then per feature — spoken problem setup over the section slide, demo, leadership card. Demo budget: 04 ≈ 8, 05 ≈ 12 (the big one), 06 ≈ 6.
- 60 min build: the Hands-On slide up; 04 is Recommended, 05 and 06 are Challenge. Sticky check at about 20 min. 05 needs the Foundry key on the board.
- Cut if behind: the zero-overlap second query in demo 04 (step 6), the drone and Avalanche Lake questions in demo 05 (steps 8 and 8b), the gear failure in demo 06 (step 5), then say a card in one sentence.
- Last 5 min: debrief.

---

## [deck-title] Module 2: Finding

Notes: Surfacing the right thing, and the point in the day where the work stops being about prompts.

## This Module

Icon (search): F04 Search
Icon (document): F05 RAG
Icon (route): F06 Recommend

Notes: One piece of infrastructure, embeddings, and the rhythm again: three demos, then your lab hour.
Embeddings turn text into vectors; distance between vectors means similarity of meaning.
04: a search box that finds what users mean.
~
05: the same retrieval, plus a model that reads what it found and cites it.
~
06: the same vectors, with the query swapped for an item. Same math, three products.
The thread to watch: embeddings capture what text is *about*, and nothing about whether it's suitable, safe, or current. Search hands a cliff edge to "quiet, for kids"; recommendations answer the 65-litre pack with the 40-litre pack. The fix is metadata, filters, thresholds, and how you split documents, not a bigger model.

## Search, RAG, Recommendations: Same First Half

Flow (04 Search): Query text -> Embed -> Nearest stored vectors -> Show the list
Flow (05 RAG): Question -> Embed -> Nearest chunks -> Model writes an answer from them -> Show the answer, with citations
Flow (06 Recommend): The item on screen -> Its stored vector -> Nearest stored vectors -> Show the list

- All three: embed, then find the nearest
- RAG adds a model that *reads* the results
- Recommendations swap the query for an item

Notes: Two minutes, one row per advance. This is the one slide most people photograph.
Row one, search: embed the query, find the nearest stored vectors, show the list.
~
Row two, RAG: the same first half, then a model writes an answer from the nearest chunks, with citations.
~
Row three, recommendations: the item on screen already has a vector; find its neighbors. No model call at request time.
~
All three: embed, then find the nearest. 05 and 06 literally import 04's search code.
~
If someone asks "is RAG just search," the honest answer is "search plus a model that reads the results," and this picture shows it.
~
Recommendations swap the query for an item, and nothing else changes.

---

## [section] F04 · Semantic Search

Notes: The user problem, spoken, 60 seconds: "dog-friendly waterfall hike, not too steep." The catalog has a dozen matches and keyword search returns almost nothing, because no description says "not too steep" — the best trail says "a gentle grade shaded by cedars." Three bad results and the user goes back to asking strangers on Reddit.

## [define] Search by what words **mean**, not by which words appear.

Icon (search): F04

Notes: The textbook version (Wikipedia): search with meaning, as distinguished from lexical search, where the engine looks for literal matches of the query words. Read the slide, let it sit a beat, then the problem.

## F04 · How It Works

Flow (Index, once): 30 trail descriptions -> nomic-embed-text -> One vector per trail, kept in memory
Flow (Search, each query): "quiet waterfall hike for kids" -> nomic-embed-text -> Cosine similarity against every stored vector -> Top 5, with scores

- ~768 numbers · similar meaning lands close
- Nothing is generated
- pgvector / Azure AI Search: same math at scale

Notes: Index, once: thirty trail descriptions through nomic-embed-text, one vector per trail, kept in memory. That happens at startup, not per search.
~
Search, each query: embed the query the same way, cosine similarity against every stored vector, top five with scores. The only model call turns text into numbers.
~
About 768 numbers per text, and similar meaning lands close. Show the raw float array once in the demo so the word "vector" stops being magic.
~
Nothing is generated. Cosine similarity is a few lines of code, and the search box stays a search box.
~
pgvector or Azure AI Search is the same math at scale. Foundation for the rest of the day: 05, 06, and 08 reuse the idea, and 06 and 08 reuse the actual vectors.

## [demo] **DEMO** · F04 Semantic Search

## [static] F04 · Semantic Search · Demo: What to Watch

- ❌ Keyword baseline: junk
- The raw float array
- 30 trails embedded in < 2 s
- Cosine: one visible method
- ✅ Same query, semantic: the right trails, with scores
- "Quiet, for kids" → Taft Point. A 3,000-foot drop.

Notes: ~8 min · `cd modules/M2-finding/F04-semantic-search/dotnet`
Files: `trails-slice.json` = 30 trails · `queries.json` = the three demo queries · `embeddings.json` = cache, delete to embed live
Args: the query, in plain words; none = "dog-friendly waterfall hike, not too steep"
1. Before: `cd starter && dotnet run`. Keyword search. Junk.
2. After: `cd ../complete && rm -f embeddings.json && dotnet run`. 30 trails embedded in under 2 s, same query. Right trails, scores beside them. Show `queryVector` in `Program.cs`: 768 floats.
3. Show the cosine method. One screen.
4. `dotnet run -- somewhere quiet to take my kids`: Taft Point, a 3,000-foot drop. Scores top out near 0.49 vs 0.77. Filters and a floor, not a better model.
Cut: 4.

## F04 · Semantic Search · Leadership Card

- **When:** A search box users complain about — they describe what they want, it matches what things are called.
- **Think:** "Our search finds what users mean, not just what they type."

Difficulty: easy

Notes: Also: any catalog where people describe what they want instead of naming it.
~
The line for your boss. Row 4.
~
Easy. If asked about cost: days to a prototype on an existing catalog; embeddings cheap to free; the work is evaluation and tuning.

---

## [section] F05 · RAG

Notes: The user problem, spoken, 60 seconds: "Can I have a campfire at Sperry Chalet in September?" The answer exists — paragraph four of a 12-page regulations document nobody will read. Search returns the document, not the answer; a plain chatbot answers fluently and makes it up, and a confidently wrong answer about fire rules is worse than no answer.

## [define] Look up your documents first; the model **answers from what it finds**.

Icon (document): F05

Notes: The textbook version (AWS/NVIDIA): retrieval-augmented generation has the model consult an authoritative knowledge source outside its training data before generating, so answers are grounded in your content and can cite it.

## F05 · How It Works

Flow (Index, once): Park docs -> Split into chunks -> Embed each chunk -> Store vector + text + source id
Flow (Answer, each question): Question -> Embed -> Retrieve top chunks (vector + keyword) -> Prompt: chunks + question + "answer only from these, cite, or say you don't know" -> gpt-4.1 -> Answer with citations your code checks

- Retrieval = 04 · generation = 01, with reading material
- The model reads instead of remembers
- Two failure spots: the wrong chunk, or a bad split

Notes: Index, once: split the park docs into chunks, embed each chunk, store the vector with its text and a source id. Say "the top row happened before anyone asked."
~
Answer, each question: embed the question, retrieve the top chunks by vector and keyword, build a prompt that says "answer only from these, cite, or say you don't know," and check the citations in code.
~
Retrieval is 04; generation is 01 with reading material.
~
The model reads instead of remembers. That is why the answer can be current and can be cited.
~
Two failure spots: the wrong chunk, or a bad split. The bad-split lesson gets its numbers in the demo: a chunk-per-section put a conditional fire rule and its overriding exception together and answered right; splitting them apart produced "campfires are fine" 4 runs in 20; re-splitting at subsection boundaries, 0 in 60, while every retrieval metric said the system was working. Chunking is a correctness decision, upstream of model choice.

## [demo] **DEMO** · F05 RAG

## [static] F05 · RAG · Demo: What to Watch

- ❌ No context: confident and wrong
- Embeddings only: 5 of 8 chunks, wrong park
- ✅ Hybrid: "Sperry" starts to count
- ✅ Grounded prompt: right answer, cited
- Citations are strings — validate the ids
- "Drone?" → "The documents don't say."
- "Open right now?" needs the date in the prompt
- One-line swap to `gpt-4.1`

Notes: ~12 min · `cd modules/M2-finding/F05-rag/dotnet`
Files: `park-docs/` = 25 regulation docs · `chunks.jsonl` = 250 chunks · `chunk-embeddings.json` = their vectors, precomputed · `questions.json` = the four lab questions
Flags: `--retrieval-only` = no answer, just ranks · `--top-k N` = chunks retrieved (default 3) · `--alpha` = semantic weight, 1.0 = cosine only, default 0.6 · `--no-context` = plain chatbot · a quoted question replaces the Sperry default
1. Before: `cd starter && dotnet run`. The Sperry campfire question, no context. Confident. Wrong.
2. Open `data/park-docs/glacier-backcountry-camping-guide.md`, section 4. Read 4.1 then 4.2 aloud. Which one does a model answer from?
3. `cd ../complete && dotnet run -- --top-k 8 --alpha 1.0 --retrieval-only`: ranks 4 to 8 are Acadia and Yosemite.
4. `dotnet run -- --top-k 8 --retrieval-only`: hybrid. Acadia gone, "sperry" doing the work.
5. After: `dotnet run`. Right answer, cited by name. Point at the citation check line.
6. `dotnet run -- "Are there EV charging stations in Glacier National Park?"`: "The documents don't say."
7. `dotnet run -- "Is the Avalanche Lake Trail open right now?"`: the date in the prompt. Without it, refused 10 of 18.
8. Point at the DI line: this was gpt-4.1. Local would be one line.
Cut: 7, then 6.

## F05 · RAG · Leadership Card

- **When:** The org already wrote the answer down, and people still ask humans.
- **Think:** "Our documents already answer these questions. This makes them answer directly, with receipts."

Difficulty: hard

Notes: Internal knowledge bases, support and policy docs, product documentation. The production work is chunking, retrieval evaluation, keeping documents current.
~
The line for your boss. The uncomfortable one underneath it: the worst bug was a text-splitting rule, not a model or a prompt. Ask what your evaluation would have caught. Row 5.
~
Hard. If asked about cost: a day to demo, weeks to production quality.

---

## [section] F06 · Recommendations

Notes: The user problem, spoken, 60 seconds: a hiker just finished Avalanche Lake Trail and loved it, and the app says nothing. "You'd probably like these three" never got built because everyone assumes it needs a data-science team and six months. Same gap in the gear store: buy the Cascade 65, get a random carousel.
No concept slide: the Same First Half diagram already told this story — same vectors, same similarity code, the query swapped for an item, no model call while the user waits.

## [define] Suggest the **next thing** a user will want, from what they liked already.

Icon (route): F06

Notes: The textbook version: a recommender system predicts what a user will prefer. Today's shortcut is content-based (similarity of descriptions); the classical version is collaborative filtering over behavior data.

## [demo] **DEMO** · F06 Recommendations

## [static] F06 · Recommendations · Demo: What to Watch

- Same vectors, still in memory
- "More like this" = search, query → item
- Lakes cluster — but mostly *hard*, and one has no lake
- Difficulty and park aren't in the description text
- Cascade 65 → Cascade 40: substitutes, not complements

Notes: ~6 min · `cd modules/M2-finding/F06-recommendations/dotnet`
Files: `trails.json` = 30 trails · `trail-embeddings.json` = 04's vectors, precomputed · `gear-reviews.jsonl` = 300 reviews
Args: a trail id or name (default trail-0117, Avalanche Lake) · `--gear <product>` = same trick over review text
1. Avalanche Lake Trail page. What belongs at the bottom of this screen?
2. Before: `cd starter && dotnet run`. Random picks. That is what most apps ship.
3. After: `cd ../complete && dotnet run`. Same vectors as 04, nothing new embedded. Lakes cluster, but mostly hard, and one has no lake.
4. Difficulty and park are in the catalog, not the text. Similarity gets candidates; metadata re-ranks.
5. `dotnet run -- --gear Cascade 65`: the Cascade 40. Substitutes, not complements. What people use with it lives in the reviews: the bear canister, in `expected-output.md`.
Cut: 5.

## F06 · Recommendations · Leadership Card

- **When:** Users finish one thing and the product offers no next step.
- **Think:** "The same vectors that power our search give us 'more like this' for free."

Difficulty: easy

Notes: Any catalog or content library. Collaborative filtering comes later, when behavior data exists.
~
The line for your boss. Row 6.
~
Easy. If asked about cost: nearly free once search exists; days from scratch.

---

## [static] Lab 2: ~60 Minutes

- ⭐ Recommended: **F04 Semantic Search**
- ⛰️ Challenge: **F05 RAG** (☁️ room key) · **F06 Recommendations**
- `FNN-lab.md` · `http/` · `expected-output.md` · `data/`
- ✅ Done = your output similar to `expected-output.md`

Notes: 04: embed 30 trails, embed a query, rank by cosine; keyword baseline provided. Two honest routes: if RAG is why you came, go straight to 05 and come back to 04 after — nobody should leave having never built the thing they came for. 06 is 04's code with the query swapped for an item.
05's step that says run question 1 twenty times matters: a wrong answer one run in five is invisible in a single run and is the only defect here that could hurt somebody.

## [static] Lab 2: Debrief

- How'd it go?				  Observations?			        Questions?

Notes: Ask two people what surprised them.
Prompts if quiet: who got Taft Point for the kids query? Who got the Cascade 40? Rows 4–6 of the framework are done. Lunch, then Module 3.
