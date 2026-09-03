# Deck 1: Module 1, Understanding (about 90 minutes)

Demo-centric rhythm: name the three features, then for each one — divider, DEMO, the lab, the leadership card. Slides appear only where a concept needs a picture (in this module: extraction's validate-or-reject pipeline). Everything J. says lives in Notes; the demo scripts live under each "What to Watch" slide.

Runsheet (lengths, not clock times):
- 30 min instructor: opener 2 min, then per feature — spoken problem setup 1–2 min over the section slide, demo, leadership card. Demo budget: 01 ≈ 8, 02 ≈ 10 (it has the one concept slide), 03 ≈ 10.
- 60 min build: the Hands-On slide up; 01 is Recommended, 02 and 03 are Challenge. Walk the room. Sticky check at about 20 min: green if output matches `expected-output.md`, red if stuck.
- Cut if behind: the break-it-on-purpose beats in demos 02 and 03, then the "one-line trail status" reshape in demo 01, then say a leadership card in one sentence instead of revealing it.
- Last 5 min: debrief; ask two people what surprised them.

---

## [deck-title] Module 1: Understanding

Notes: Making sense of messy content, and everything runs on your laptop.

## This Module

Icon (document): F01 Summarize
Icon (table): F02 Extract
Icon (gauge): F03 Sentiment

Notes: One model call + one careful instruction; the rhythm is three demos, then your lab hour.
Three features, one shape: a single chat call and a carefully written instruction. No vector database, no training, no cloud.
The thread to watch all module: asked for something it cannot find, a model supplies something plausible rather than nothing. 01 invents a closure, 02 invents a distance, 03 believes the sarcasm. A better model does not fix it; a tighter instruction, a null-able schema, and checking code do.
Set the rhythm expectation out loud: three quick demos with a card for your boss after each, then the room builds for an hour — 01 recommended, 02 and 03 for anyone with time left.

---

## [section] F01 · Summarization

Notes: The user problem, spoken over this slide, 60 seconds: forty trip reports for Avalanche Lake Trail, 1,200 words each. Somewhere in there: is the bridge out, and are the mosquitoes bad? Nobody reads forty essays; they skim three and miss the warning in the fourth. The hiker at mile two from the opening story is this feature's user.
The concept needs no slide: one chat call, a document, an instruction. "Summarize this" produces a book report; a real product asks for a summary with a purpose and a shape. The demo shows exactly that.

## [demo] **DEMO** · F01 Summarization

## [static] F01 · Summarization · Demo: What to Watch

- The raw report first
- ❌ Naive prompt: faithful, generic, useless
- ✅ Purpose-built prompt: the bridge surfaces
- Grounding lines — without them, ~half of runs invent a closure
- Same call, new shape: a one-line trail status

Notes: A required "hazards" slot with nothing honest to put in it is an invitation to make something up. The invented bear-related closure happens on the *clean* report; numbers are in `expected-output.md`. Point at the Ollama endpoint at the end: there was no API key, and no data left the room.

Demo script (~8 min; the reshape in step 6 is the first cut):
1. Open a raw trip report from `data/trip-reports/` on screen and scroll through it slowly, so the room feels the problem before seeing the fix.
2. Starter project: `IChatClient` wired to Ollama via Microsoft.Extensions.AI. One method, one prompt: "Summarize this trip report."
3. Run it and get a faithful, generic, useless book report, then name the failure out loud: right model, wrong instruction.
4. Iterate the prompt live into the hiker-focused version (conditions, hazards, crowding, ignore everything else). Run again. This is the payoff: three bullets, and the bridge warning surfaces.
5. Point at the last two lines of that prompt, the grounding lines. Without them, this exact prompt run against the clean report invented a bear-related trail closure in roughly half of runs, because a required "hazards" slot with nothing honest to put in it is an invitation to make something up. The numbers and the failing outputs are in `expected-output.md`.
6. Change the shape: same call, but output a one-line "trail status" headline for a card UI. Same feature, different product surface.
7. Point at the Ollama endpoint in the code. This ran entirely on the laptop, with no API key and no data leaving the room.

## F01 · Summarization · Leadership Card

- **When:** Too much content for people to read, easy-to-miss detail buried in avalanche of text.
- **Think:** "Our users are drowning in text we already have. A few hours of work turns it into automatic answers."

Difficulty: easy

Notes: Reviews, tickets, reports, meeting notes, threads. One call per document. Row 1 of the framework.
If asked about cost: days of work, free local models, no new infrastructure.

---

## [section] F02 · Extraction

Notes: The user problem, spoken, 60 seconds: Trailhead wants a "trail stats" panel — which trails, when, how far, what wildlife, what condition. All of it exists as prose across the forty reports, and today a human would re-read and re-type it. So the panel doesn't exist.
This one gets the module's one concept slide, because the pipeline — and where it lies to you — is the lesson.

## F02 · How It Works

Flow: Trip report -> Prompt: fields, types, examples, "null when absent" -> llama3.2 in JSON mode -> Parse and validate against the schema -> Store, or reject and log

- The model writes JSON; your code decides whether to believe it
- The `null` rule and the validator are the feature
- A rejected record is a good outcome

Notes: JSON mode guarantees syntax, not truth. A silently stored zero is the bug.
The last box is the one people skip. Say plainly that extraction without validation is a data-corruption feature.

## [demo] **DEMO** · F02 Extraction

## [static] F02 · Extraction · Demo: What to Watch

- A C# record (`TripFacts`) · typed response · no parsing
- ❌ No distance in the report → `distance_mi: 5.0`
- ✅ Nullable fields + "null when not stated"
- Run it 3×: better ≠ guaranteed
- The last mile is a validator — ordinary code

Notes: After the schema fix, something else usually still slips: `elevation_gain_ft: 0` where the honest answer is `null`, or "early last month" as a date. Zero is the dangerous miss: it is a value, and a pipeline will store it without complaint. Ship the schema plus a rejection rule.

Demo script (~10 min; the break-it beat in step 4 is the first cut):
1. Show the same messy trip report from feature 01. This time the goal isn't a summary, it's a database row.
2. Define a C# record (`TripFacts`: trail, park, date, distance, wildlife, conditions) and use Microsoft.Extensions.AI's typed-response support to request it directly. The schema is code.
3. Run it: prose in, populated .NET object out, with no parsing step. This is the payoff moment.
4. Break it on purpose: run a report that never mentions distance, and watch the model invent `distance_mi: 5.0`.
5. Fix it in the schema with nullable fields and "null when not stated" descriptions. Re-run: the invented distance usually goes away, and something else usually doesn't. Run it two or three times live so the room sees the variance rather than one lucky result. Then say the quiet part: this is better, and it is not a guarantee. The last mile is a validator that rejects a date like "early last month" and a `0` that should have been `null`, which is ordinary code your team already knows how to write.
6. Zoom out: loop over ten reports and print rows. That's an ingestion pipeline in thirty lines, and the "trail stats" panel is now just a query.

## F02 · Extraction · Leadership Card

- **When:** Valuable data trapped in documents nobody will re-type.
- **Think:** "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it."

Difficulty: medium

Notes: Invoices, resumes, support emails, contracts, legacy records. Row 2.
If asked about cost: days to a working pipeline; the real work is schema design and spot-checking accuracy.

---

## [section] F03 · Sentiment

Notes: The user problem, spoken, 60 seconds: the Cascade 65 backpack has 300 reviews. Are people happy, and what are they mad about? Star ratings lie — "4 stars, but the hip belt broke on day two." The user here is the product team, not the hiker; track the signal weekly and a defect surfaces months before returns spike.
No concept slide: this is classification, and the demo IS the lesson — the day's model-selection experiment. phi3 free and local vs gpt-4.1 on Azure, same prompt, and you measure instead of assume.

## [demo] **DEMO** · F03 Sentiment

## [static] F03 · Sentiment · Demo: What to Watch

- One method · `positive | negative | mixed` · swap = one DI line
- 20 easy through `phi3`: fast, free, correct
- Same 20 through Azure: 9 of 10 identical
- The hard set through both — diff on screen
- 📊 `phi3` 9/10 · 7/10 — `gpt-4.1` 10/10 · 10/10

Notes: Both halves of the argument hold: ordinary reviews barely need the big model, and the sarcastic slice does, 7/10 vs 10/10, with every disagreement going the frontier model's way. The local stand-in (`llama3.2`) tied phi3, so a comparison against the wrong big model would have said the gap does not exist. Side finding worth 20 seconds: reflowing the identical prompt onto one line dropped phi3 to 7/10 and 4/10. Small models are sensitive to formatting.
Close with the decision recipe: labeled sample, run both, count disagreements, price the errors. That recipe generalizes to every feature today.

Demo script (~10 min; the hard-set diff in step 5 is the heart — cut step 4's Azure easy-run first if behind):
1. Show the Cascade 65's reviews and point at a 4-star review with furious text, which is the whole case for reading the text rather than the stars.
2. Starter project: one classify method whose prompt returns exactly `positive | negative | mixed`. Because of Microsoft.Extensions.AI, the same code runs against both providers, and the swap is one line in DI registration. Say that out loud, since it's the provider-flexibility slide made real.
3. Run 20 easy reviews through `phi3` locally, where it is fast, free, and correct, and let the small model win the first round.
4. Run the same 20 through `gpt-4.1` on Foundry and get the same labels on nine of ten (the frontier model also gets the deadpan four-star rave that `phi3` calls mixed), which is the first payoff: on ordinary reviews you'd have paid for very little.
5. Now the hard set, sarcasm and mixed reviews. Run both and diff the labels on screen. Second payoff: `phi3` drops to 7/10 and `gpt-4.1` holds at 10/10, and every disagreement is one the frontier model gets right. This is the slice that earns its price, and you know that because you measured it rather than assumed it. Check `expected-output.md` for what happened when this was built, and be ready for the room's answer to differ from yours.

## F03 · Sentiment · Leadership Card

- **When:** A high-volume text stream needing a judgment call.
- **Think:** "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun."

Difficulty: easy

Notes: Reviews, NPS verbatims, tickets, mentions, survey answers. Row 3.
If asked about cost: days; the classifier is trivial, the diligence is a labeled sample and an error count.

---

## [static] Lab 1: ~60 Minutes

- ⭐ Recommended: **F01 Summarization**
- ⛰️ Challenge: **F02 Extraction** · **F03 Sentiment**
- `FNN-lab.md` · `http/ollama.http` · `expected-output.md` · `data/`
- ✅ Done = your output similar to `expected-output.md`

Notes: Start with 01 regardless of experience: one endpoint, one prompt, two reports; everything else today assumes you've made one model call and seen what comes back. New to this? Budget the full time and do the stretch goal.
Pick 03 if your question at work is "which model should we pay for" — its Azure half needs the room key. 02's done-check: the sparse report comes back with nulls and the validator says PASS/REJECT honestly.
Finished everything? Help someone near you.

## [static] Lab 1: Debrief

- How'd it go?				  Observations?			        Questions?

Notes: Ask two people what surprised them.
Prompts if quiet: what surfaced the buried bridge, and what invented one? Rows 1–3 of the framework are done. Break, then Module 2.
