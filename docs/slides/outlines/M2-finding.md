# Deck 2: Module 2, Finding (about 90 minutes)

Three features built on one idea: embeddings. Same math, three products, and the point in the day where the work stops being about prompts.

This file is the instructor script for the module. The bullets under each slide are what is on screen; the `Notes:` text and the "Demo script" steps under each "Demo: What to Watch" slide are what you say and click. The specs in `modules/` are written for attendees and no longer carry the demo steps.

Runsheet (lengths, not clock times):
- 30 min instructor: slides 1 to 4 (the module thread and the comparison diagram), then for each feature the problem, concept, how-it-works diagram, demo (about 8 min each), and leadership card. Feature 04 first, then 05, then 06.
- 60 min build: slide "Hands-On" up; attendees pick 04 (recommended), 05, or 06. Sticky check at about 20 min. 05 needs the Foundry key on the board.
- Cut if behind: the zero-overlap second query in demo 04 (step 6), the drone and Avalanche Lake questions in demo 05 (steps 8 and 8b), and the gear failure in demo 06 (step 5).
- Last 5 min: debrief slide.

---

## 1. Module 2: Finding

- Surfacing the right thing
- 04 Semantic Search (Recommended) · 05 RAG · 06 Recommendations
- One piece of infrastructure, three features

## 2. One Idea, Three Products

- Embeddings turn text into vectors; distance between vectors means similarity of meaning
- Search: rank a catalog against a query
- RAG: retrieve passages, hand them to a model to answer from
- Recommendations: rank a catalog against an item instead of a query

## 3. The Thread to Watch

- Embeddings capture what text is *about*, and nothing about whether it is suitable, safe, current, or complete
- Search returns a cliff-edge trail for "somewhere quiet to take my kids"
- Recommendations answer "you bought the 65-litre pack" with the 40-litre pack
- RAG answers correctly only because a rule and its exception stayed in one chunk
- The fix is metadata, filters, thresholds, and how you split documents, rather than a bigger model

Notes: Feature 05 has the measurements: better chunking moved its flagship question from 75 to 97 percent correct; a model ten times the size bought only the last three points.

---

## 4. Search, RAG, Recommendations: Same First Half

Flow (04 Search): Query text -> Embed -> Nearest stored vectors -> Show the list
Flow (05 RAG): Question -> Embed -> Nearest chunks -> Model writes an answer from them -> Show the answer, with citations
Flow (06 Recommend): The item on screen -> Its stored vector -> Nearest stored vectors -> Show the list

- All three start with "turn text into a vector and find the nearest ones"
- RAG adds a model at the end that reads the results instead of showing them
- Recommendations swap the query for an item the user is already looking at, so there is no model call at request time
- This is why the demos reuse code: 05 and 06 import 04's search

Notes: Two minutes. This is the one slide most people photograph. If someone asks "is RAG just search," the honest answer is "search plus a model that reads the results," and this picture shows it.

## 5. 04 · Semantic Search: The User Problem

- "dog-friendly waterfall hike, not too steep"
- The catalog has a dozen matches. Keyword search returns almost nothing: no description says "not too steep." One says "a gentle grade shaded by cedars."
- Three bad results, and the user goes back to asking strangers on Reddit

## 6. 04 · The Concept

- Embed every trail description once; embed the query at search time; rank by cosine similarity
- Cosine similarity is a few lines of code in any language
- `nomic-embed-text` is small, free, local; 200 trails embed in seconds
- The search box stays a search box. Users learn nothing new.

Notes: Foundation feature for the rest of the day: 05, 06, and 08 all reuse the idea, and 06 and 08 reuse the actual vectors.

## 7. 04 · How It Works

Flow (Index, once): 30 trail descriptions -> nomic-embed-text -> One vector per trail, kept in memory
Flow (Search, each query): "quiet waterfall hike for kids" -> nomic-embed-text -> Cosine similarity against every stored vector -> Top 5, with scores

- An embedding is a list of about 768 numbers. Text that means similar things lands close together.
- No text is generated. The only model call turns text into numbers.
- Cosine similarity is one method; a real store (pgvector, Azure AI Search) does the same math at scale

Notes: Show the raw float array once so the word "vector" stops being magic. Then show the cosine method: it fits on a slide.

## 8. 04 · Demo: What to Watch

- Keyword search over `trails.json` returns junk. That's the baseline.
- Embed one sentence and look at the raw float array. It's just numbers.
- Embed the catalog and time it: under two seconds for the 30-trail slice
- Cosine similarity on screen: one visible method
- Same query, semantic: the gentle shaded waterfall trails rise, scores visible
- "somewhere quiet to take my kids": the top hit is Taft Point, beside a 3,000-foot drop

Notes: A perfect topical match and terrible advice, because embeddings capture what a text is about and not whether it suits the asker. Note the scores: this query tops out near 0.49 vs 0.77 for query one. That number is what a score floor keys on. The fix is metadata filters and a floor rather than a better model.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Run the "dog-friendly waterfall hike, not too steep" query against naive keyword search over `data/trails-slice.json` and get junk results. That's today's baseline.
2. Show what an embedding is: embed one sentence via Microsoft.Extensions.AI's `IEmbeddingGenerator` against Ollama and look at the raw float array on screen for a moment, because it is worth seeing that an embedding is nothing more than a list of numbers.
3. Embed the trail catalog in a loop and time it live. The checked-in demo embeds the 30-trail slice in `data/trails-slice.json` in under two seconds. The point lands: seconds, not minutes, and it happens once at startup rather than per search.
4. Write cosine similarity on screen; it fits in one visible method, which surprises people.
5. Re-run the same query semantically. The payoff: the gentle shaded waterfall trails rise to the top, with the similarity scores visible next to each hit.
6. Show one more query with zero keyword overlap ("somewhere quiet to take my kids"), which proves it isn't a fluke and then hands you the feature's best moment. The top hit is Taft Point, whose own description warns that the ground gives no second chances beside a 3,000-foot drop. It is a perfect topical match and terrible advice. Say what that means: embeddings capture what text is about, not whether it's suitable, and the fix is metadata filters and a score floor, not a better model. Note that everything ran locally, and that the scores for this query top out near 0.49 against 0.77 for query one, which is the number a score floor would key on.

## 9. 04 · Leadership Card

- When: any search box users complain about; any catalog where people describe what they want instead of naming it
- Cost: days to a prototype on an existing catalog. Embeddings cheap to free; the work is evaluation and tuning.
- "Our search finds what users mean, not just what they type."

---

## 10. 05 · RAG: The User Problem

- "Can I have a campfire at Sperry Chalet in September?"
- The answer exists, in paragraph four of a 12-page regulations document nobody will read
- Search returns the document, not the answer. A plain chatbot answers fluently and makes it up.
- A confidently wrong answer about fire regulations is worse than no answer

## 11. 05 · The Concept

- Feature 04's retrieval bolted onto generation: "Answer using only this context. If it doesn't cover the question, say so. Cite the source."
- Three mechanics: chunking, retrieval (embeddings plus a lexical signal it turns out to need), grounded prompting
- A fourth that's easy to leave out: tell the model what "now" is

Notes: The model becomes a reading assistant for your content rather than an oracle.

## 12. 05 · Chunking Is a Correctness Decision

- One chunk per numbered section put a conditional fire rule and the absolute exception that overrides it in the same 256-word chunk
- Retrieval ranked that chunk first every time. The model read the conditional, stopped, and said a campfire was fine: 4 runs in 20.
- Split oversized sections at their own subsection boundaries: 0 in 60. Retrieval scores barely moved.
- Every retrieval metric said the system was working

Notes: Open Section 4 of the Glacier backcountry guide and read 4.1 then 4.2 aloud. Ask the room which one a model answers from. Worth 90 seconds.

## 13. 05 · How It Works

Flow (Index, once): Park docs -> Split into chunks -> Embed each chunk -> Store vector + text + source id
Flow (Answer, each question): Question -> Embed -> Retrieve top chunks (vector + keyword) -> Prompt: chunks + question + "answer only from these, cite, or say you don't know" -> gpt-4.1 -> Answer with citations your code checks

- Retrieval is feature 04. Generation is feature 01 with the reading material supplied.
- The model reads instead of remembers, which is why the answer can be current and can be cited
- Two places to be wrong: retrieval brings the wrong chunk, or the chunk splits a rule from its exception

Notes: Point at the second row and say "the top row happened before anyone asked." Then say where the failure in the next slide lives: in the split box, not in the model box.

## 14. 05 · Demo: What to Watch

- No context: a confident wrong answer. Open the regulation and let it sit.
- Pure embedding retrieval, top 8: five of eight chunks from the wrong park (Acadia, Yosemite)
- Hybrid retrieval: cosine plus a rarity-weighted keyword score. "Sperry" starts to count. Wrong-park chunks gone.
- Grounded prompt: the right answer, with the regulation cited
- Citation check: a citation is only a string the model typed. Validate ids against what was retrieved; small models invent chunk ids that look right and point nowhere.
- Unanswerable question ("drone?") gets "The provided documents don't say."
- "Is Avalanche Lake Trail open right now?" works because the date is in the prompt
- Swap generation to `gpt-4.1` on Foundry with the one-line client change; compare a multi-document answer

Notes: The date beat: "CLOSED effective June 20, 2026, until further notice" is unanswerable without a calendar. Without the date the model refused this question 10 runs in 18; with it, 1 in 20. Almost every real knowledge base is a corpus of dated notices. Most transferable thing in the feature.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Ask the model the Sperry Chalet campfire question with no context. It answers confidently. Then open the actual regulation and show the answer is wrong. Let that sit for a beat; this is the whole reason RAG exists.
2. Show the chunked park docs, then open Section 4 of the Glacier backcountry guide in `data/park-docs/` and read 4.1 and 4.2 aloud in order. A conditional rule, then the absolute exception that cancels it. Ask the room which one a model answers from. This is the chunking beat and it is worth 90 seconds; the numbers are in `expected-output.md` under "Chunking".
3. Reuse the embedding search from feature 04 against the chunks. Retrieval is literally the previous feature. Show the score table for the Sperry question, then run `--top-k 8 --alpha 1.0` and point at ranks 4 through 8: Acadia and Yosemite campfire sections, five of eight chunks from the wrong park.
4. Turn on hybrid retrieval. Same cosine score, plus a keyword score weighted by word rarity, combined with a visible alpha. Re-run `--top-k 8` at the default alpha: Acadia is gone and the wrong-park chunks are replaced by Glacier documents that name Sperry. The printed table shows both component scores, so you can point at "sperry" doing the work.
5. Build the grounded prompt: retrieved chunks + question + "answer only from the context, cite the source, say when you don't know."
6. Run it. The payoff: the right answer, with the regulation document cited by name.
7. Show the citation check. Parse the ids out of the answer, compare against what was retrieved, print a loud failure on a mismatch. Run the unanswerable question a couple of times until the model appends an invented chunk id to its own refusal, which it does often enough to count on.
8. Ask it something the docs don't cover ("Can I bring a drone?" if absent) and get an honest "the provided documents don't say," not a guess.
8b. Ask "Is the Avalanche Lake Trail open right now?" and show the date sitting in the prompt. Explain why it is there: without it the same question got refused in 10 runs out of 18, not because retrieval missed but because "effective June 20, 2026, until further notice" is unanswerable without a calendar. Worth 60 seconds; it is the most transferable thing in the feature.
9. Swap generation from the local model to `gpt-4.1` on Foundry with the one-line Microsoft.Extensions.AI client change, re-run a multi-document question, and compare answer quality side by side.

## 15. 05 · Where the Money Goes

- Swapping the 3B local model for a 32B one closes every remaining defect
- Cost: about 20x the latency per answer, 0.9 s to 16.6 s
- Swapping the model before fixing the chunk boundary buys a smarter model reading the wrong context
- Chunking is upstream of model choice

## 16. 05 · Leadership Card

- When: internal knowledge bases, support and policy docs, product documentation. Anywhere the org already wrote the answer down and people still ask humans.
- Cost: weeks to production quality. The demo takes a day; the work is chunking, retrieval evaluation, keeping documents current.
- The uncomfortable one: the worst bug was a text-splitting rule, not a model or a prompt. Ask what your evaluation would have caught.
- "Our documents already answer these questions. This makes them answer directly, with receipts."

---

## 17. 06 · Recommendations: The User Problem

- A hiker just finished Avalanche Lake Trail and loved it. The app says nothing.
- "You'd probably like these three" never got built, because everyone assumes it needs a data-science team and six months
- Same gap in the gear store: buy the Cascade 65, get a random carousel

## 18. 06 · The Concept

- Classical answer: collaborative filtering over behavior data. Real work, real cold-start problem.
- The shortcut: content-based recommendations from the vectors you already have. "More like this" is nearest neighbors of the item's vector.
- No training, no ratings matrix, no cold start, and the infrastructure exists
- Honest caveat: similarity finds things that are alike. It will never discover that waterfall hikers buy headlamps.

Notes: One embedding investment keeps paying: search, recommendations, anomaly detection from one piece of infrastructure.

## 19. 06 · How It Works

Flow: The trail the user is viewing -> Its vector, already computed by 04 -> Cosine against the rest of the catalog -> Drop itself, take five -> "More like this"

- No query, no prompt, no model call while the user waits
- Same vectors, same similarity code as search. The only new line is "skip the item itself."
- What it can't do: cross-category ("people who bought this pack also bought a stove") needs behavior data, not descriptions

Notes: The gear failure in the demo (65-litre pack recommends the 40-litre pack) is this diagram working exactly as drawn. Description similarity finds the most similar description.

## 20. 06 · Demo: What to Watch

- Same vectors from 04, still in memory; nothing new gets created
- "More like this": the search code with the query swapped for an item
- Alpine lake hikes do cluster. They also come back mostly *hard* when you started from a moderate family walk, and one neighbor is a Smokies hike with no lake at all.
- Difficulty and park are in the catalog, not in the description text, so the embedding can't see them
- Gear: top result for the Cascade 65 Backpack is the Cascade 40 Daypack. Substitutes, not complements. Review co-mentions put the bear canister on top.

Notes: This gets you candidates; metadata re-ranking gets you recommendations. One target trail in the slice has no real neighbors at all; a shipping product shows nothing rather than five weak guesses.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Open the Avalanche Lake Trail page and ask the room what should be at the bottom of this screen. Everyone knows the answer, and almost nobody builds it.
2. Bring back the embedded trail catalog from feature 04, already sitting in memory; nothing new is created in this step, which is the point.
3. Write "more like this": take one trail's vector, rank every other trail by cosine similarity, skip itself, and take five. It's the search code with the query swapped for an item.
4. Run it for a few trails and let the room judge the results by reading the names, because they can. The alpine lake hikes do cluster, but they also come back mostly rated hard when the trail you started from is a moderate family walk, and one neighbor is a Smokies hike with no lake at all. Difficulty and park are right there in the catalog, but they are not in the description text, so the embedding cannot see them. One screen carries the lesson: similarity gets you candidates, and metadata re-ranking turns them into recommendations.
5. Do the same with gear and let it fail on purpose. The top result for the Cascade 65 Backpack is the Cascade 40 Daypack, the single product that buyer will never need. Content similarity finds substitutes; a store wants complements. Then show the review co-mention counts, where the Summit Bear Canister sits at the top, and name the difference: what a thing is like versus what people actually use with it.
6. Close on what this can't do. Cross-category discovery is out of reach for description similarity, one target trail in the slice has no real neighbors at all and a shipping product should show nothing rather than five weak guesses, and behavior data is what fixes both.

## 21. 06 · Leadership Card

- When: any catalog or content library where users finish one thing and get no next step
- Cost: nearly free if you already built semantic search. Days from scratch. Collaborative filtering later, when behavior data exists.
- "The same vectors that power our search give us 'more like this' for free."

---

## 22. Hands-On: 60 Minutes

- **Start with 04 Semantic Search** (Recommended): embed 30 trails, embed a query, rank by cosine. Keyword baseline provided.
- **Two honest routes.** If RAG is why you came, **go straight to 05** and come back to 04 after. 05 ships precomputed chunk vectors.
- **06 Recommendations** is the shortest if you've done 04: same code, query swapped for an item

Notes: Nobody should leave this room having never built the thing they came for. 05's step 6 says run question 1 twenty times, not once; a wrong answer one run in five is invisible in a single run and is the only defect here that could hurt somebody.

## 23. Debrief

- Who got Taft Point for the kids query? Who got the Cascade 40?
- Rows 4 to 6 of the decision framework are done
- Lunch, then Module 3.
