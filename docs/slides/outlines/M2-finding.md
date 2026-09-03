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

Notes: One piece of infrastructure — embeddings — and the rhythm again: three demos, then your lab hour.
Embeddings turn text into vectors; distance between vectors means similarity of meaning. Same math, three products.
The thread to watch: embeddings capture what text is *about* — nothing about whether it's suitable, safe, or current. Search hands a cliff edge to "quiet, for kids"; recommendations answer the 65-litre pack with the 40-litre pack. The fix is metadata, filters, thresholds, and how you split documents — not a bigger model.

## Search, RAG, Recommendations: Same First Half

Flow (04 Search): Query text -> Embed -> Nearest stored vectors -> Show the list
Flow (05 RAG): Question -> Embed -> Nearest chunks -> Model writes an answer from them -> Show the answer, with citations
Flow (06 Recommend): The item on screen -> Its stored vector -> Nearest stored vectors -> Show the list

- All three: embed, then find the nearest
- RAG adds a model that *reads* the results
- Recommendations swap the query for an item

Notes: Two minutes, one row per advance. This is the one slide most people photograph. Recommendations make no model call at request time, and 05 and 06 literally import 04's search code. If someone asks "is RAG just search," the honest answer is "search plus a model that reads the results," and this picture shows it.

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

Notes: The only model call turns text into numbers; cosine similarity is a few lines of code, and the search box stays a search box. Show the raw float array once in the demo so the word "vector" stops being magic.
Foundation for the rest of the day: 05, 06, and 08 reuse the idea, and 06 and 08 reuse the actual vectors.

## [demo] **DEMO** · F04 Semantic Search

## [static] F04 · Semantic Search · Demo: What to Watch

- ❌ Keyword baseline: junk
- The raw float array
- 30 trails embedded in < 2 s
- Cosine: one visible method
- ✅ Same query, semantic: the right trails, with scores
- "Quiet, for kids" → Taft Point. A 3,000-foot drop.

Notes: Taft Point is a perfect topical match and terrible advice, because embeddings capture what a text is about and not whether it suits the asker. Note the scores: this query tops out near 0.49 vs 0.77 for query one. That number is what a score floor keys on. The fix is metadata filters and a floor rather than a better model.

Demo script (~8 min; step 6 is the first cut):
1. Run the "dog-friendly waterfall hike, not too steep" query against naive keyword search over `data/trails-slice.json` and get junk results. That's today's baseline.
2. Show what an embedding is: embed one sentence via Microsoft.Extensions.AI's `IEmbeddingGenerator` against Ollama and look at the raw float array on screen for a moment, because it is worth seeing that an embedding is nothing more than a list of numbers.
3. Embed the trail catalog in a loop and time it live. The checked-in demo embeds the 30-trail slice in `data/trails-slice.json` in under two seconds. The point lands: seconds, not minutes, and it happens once at startup rather than per search.
4. Write cosine similarity on screen; it fits in one visible method, which surprises people.
5. Re-run the same query semantically. The payoff: the gentle shaded waterfall trails rise to the top, with the similarity scores visible next to each hit.
6. Show one more query with zero keyword overlap ("somewhere quiet to take my kids"), which proves it isn't a fluke and then hands you the feature's best moment. The top hit is Taft Point, whose own description warns that the ground gives no second chances beside a 3,000-foot drop. It is a perfect topical match and terrible advice. Say what that means: embeddings capture what text is about, not whether it's suitable, and the fix is metadata filters and a score floor, not a better model. Note that everything ran locally, and that the scores for this query top out near 0.49 against 0.77 for query one, which is the number a score floor would key on.

## F04 · Semantic Search · Leadership Card

- **When:** A search box users complain about — they describe what they want, it matches what things are called.
- **Think:** "Our search finds what users mean, not just what they type."

Difficulty: easy

Notes: Also: any catalog where people describe what they want instead of naming it. Row 4.
If asked about cost: days to a prototype on an existing catalog; embeddings cheap to free; the work is evaluation and tuning.

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

Notes: Reading instead of remembering is why the answer can be current and can be cited. Point at the second row and say "the top row happened before anyone asked."
The bad-split lesson gets its numbers in the demo: a chunk-per-section put a conditional fire rule and its overriding exception together and answered right; splitting them apart produced "campfires are fine" 4 runs in 20; re-splitting at subsection boundaries, 0 in 60 — while every retrieval metric said the system was working. Chunking is a correctness decision, upstream of model choice.

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

Notes: The date beat: "CLOSED effective June 20, 2026, until further notice" is unanswerable without a calendar. Without the date the model refused this question 10 runs in 18; with it, 1 in 20. Almost every real knowledge base is a corpus of dated notices — most transferable thing in the feature. Small models invent chunk ids that look right and point nowhere.
Where the money goes, one sentence for the close: swapping the 3B local model for a 32B one closes every remaining defect at ~20x the latency (0.9 s to 16.6 s) — and swapping the model before fixing the chunk boundary buys a smarter model reading the wrong context.

Demo script (~12 min; steps 8 and 8b are the first cuts):
1. Ask the model the Sperry Chalet campfire question with no context. It answers confidently. Then open the actual regulation and show the answer is wrong. Let that sit for a beat; this is the whole reason RAG exists.
2. Show the chunked park docs, then open Section 4 of the Glacier backcountry guide in `data/park-docs/` and read 4.1 and 4.2 aloud in order. A conditional rule, then the absolute exception that cancels it. Ask the room which one a model answers from. This is the chunking beat and it is worth 90 seconds; the numbers are in `expected-output.md` under "Chunking".
3. Reuse the embedding search from feature 04 against the chunks. Retrieval is literally the previous feature. Show the score table for the Sperry question, then run `dotnet run -- --top-k 8 --alpha 1.0 --retrieval-only` and point at ranks 4 through 8: Acadia and Yosemite campfire sections, five of eight chunks from the wrong park.
4. Turn on hybrid retrieval. Same cosine score, plus a keyword score weighted by word rarity, combined with a visible alpha. Re-run at the default alpha, `dotnet run -- --top-k 8 --retrieval-only`: Acadia is gone and the wrong-park chunks are replaced by Glacier documents that name Sperry. The printed table shows both component scores, so you can point at "sperry" doing the work.
5. Build the grounded prompt: retrieved chunks + question + "answer only from the context, cite the source, say when you don't know."
6. Run it. The payoff: the right answer, with the regulation document cited by name.
7. Show the citation check. Parse the ids out of the answer, compare against what was retrieved, print a loud failure on a mismatch. Run the unanswerable question a couple of times until the model appends an invented chunk id to its own refusal, which it does often enough to count on.
8. Ask it something the docs don't cover ("Can I bring a drone?" if absent) and get an honest "the provided documents don't say," not a guess.
8b. Ask "Is the Avalanche Lake Trail open right now?" and show the date sitting in the prompt. Explain why it is there: without it the same question got refused in 10 runs out of 18, not because retrieval missed but because "effective June 20, 2026, until further notice" is unanswerable without a calendar. Worth 60 seconds; it is the most transferable thing in the feature.
9. Swap generation from the local model to `gpt-4.1` on Foundry with the one-line Microsoft.Extensions.AI client change, re-run a multi-document question, and compare answer quality side by side.

## F05 · RAG · Leadership Card

- **When:** The org already wrote the answer down, and people still ask humans.
- **Think:** "Our documents already answer these questions. This makes them answer directly, with receipts."

Difficulty: hard

Notes: Internal knowledge bases, support and policy docs, product documentation. The production work is chunking, retrieval evaluation, keeping documents current.
If asked about cost: a day to demo, weeks to production quality.
The uncomfortable line for leadership: the worst bug was a text-splitting rule, not a model or a prompt. Ask what your evaluation would have caught. Row 5.

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

Notes: Similarity gets you candidates; metadata re-ranking gets you recommendations. Review co-mentions put the bear canister on top for the Cascade 65. One target trail in the slice has no real neighbors at all; a shipping product shows nothing rather than five weak guesses. And the honest caveat: description similarity will never discover that waterfall hikers buy headlamps — cross-category needs behavior data.

Demo script (~6 min; step 5 is the first cut):
1. Open the Avalanche Lake Trail page and ask the room what should be at the bottom of this screen. Everyone knows the answer, and almost nobody builds it.
2. Bring back the embedded trail catalog from feature 04, already sitting in memory; nothing new is created in this step, which is the point.
3. Write "more like this": take one trail's vector, rank every other trail by cosine similarity, skip itself, and take five. It's the search code with the query swapped for an item.
4. Run it for a few trails and let the room judge the results by reading the names, because they can. The alpine lake hikes do cluster, but they also come back mostly rated hard when the trail you started from is a moderate family walk, and one neighbor is a Smokies hike with no lake at all. Difficulty and park are right there in the catalog, but they are not in the description text, so the embedding cannot see them. One screen carries the lesson: similarity gets you candidates, and metadata re-ranking turns them into recommendations.
5. Do the same with gear and let it fail on purpose. The top result for the Cascade 65 Backpack is the Cascade 40 Daypack, the single product that buyer will never need. Content similarity finds substitutes; a store wants complements. Then show the review co-mention counts, where the Summit Bear Canister sits at the top, and name the difference: what a thing is like versus what people actually use with it.
6. Close on what this can't do. Cross-category discovery is out of reach for description similarity, one target trail in the slice has no real neighbors at all and a shipping product should show nothing rather than five weak guesses, and behavior data is what fixes both.

## F06 · Recommendations · Leadership Card

- **When:** Users finish one thing and the product offers no next step.
- **Think:** "The same vectors that power our search give us 'more like this' for free."

Difficulty: easy

Notes: Any catalog or content library. Collaborative filtering comes later, when behavior data exists. Row 6.
If asked about cost: nearly free once search exists; days from scratch.

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
