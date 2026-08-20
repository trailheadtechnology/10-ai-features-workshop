# Deck 1: Module 1, Understanding (about 90 minutes)

Three features that share one shape: a single model call and a carefully written instruction, with no vector database and no training.

This file is the instructor script for the module. The bullets under each slide are what is on screen; the `Notes:` text and the "Demo script" steps under each "Demo: What to Watch" slide are what you say and click. The specs in `modules/` are written for attendees and no longer carry the demo steps.

Runsheet (lengths, not clock times):
- 30 min instructor: slides 1 to 4 (the module thread and the comparison diagram), then for each feature the problem, concept, how-it-works diagram, demo (about 8 min each), and leadership card. Feature 01 first, then 02, then 03.
- 60 min build: slide "Hands-On" up; attendees pick 01 (recommended), 02, or 03. Walk the room. Sticky check at about 20 min: green if `complete/` output matches `expected-output.md`, red if stuck.
- Cut if behind: the "one-line trail status" reshape in demo 01, the break-it-on-purpose beats in demos 02 and 03, then say the 02 and 03 leadership cards in one sentence each.
- Last 5 min: debrief slide; ask two people what surprised them.

---

## 1. Module 1: Understanding

- Making sense of messy content
- 01 Summarization (Recommended) · 02 Extraction · 03 Sentiment
- All three run on your laptop

Notes: The wording of the instruction does more work than the choice of model. If they remember one thing from this module, that is it.

## 2. The Thread to Watch

- Asked for something it cannot find, a language model supplies something plausible rather than nothing
- 01 invents a trail closure from a bear sighting
- 02 reports a distance the report never gives
- 03 reads a sarcastic five-star review at face value
- A better model does not fix this. A more specific instruction, a schema that permits `null`, and code that checks the output do.

Notes: Plant this now; every feature in the module pays it off, and it saves them in the afternoon.

---

## 3. Three Features, One Call

Flow (01 Summarize): Trip report -> "In 3 bullets, for a hiker: conditions, hazards, crowding" -> llama3.2 -> Prose or bullets
Flow (02 Extract): Trip report -> Prompt + JSON schema, nulls allowed -> llama3.2, JSON mode -> Validated JSON
Flow (03 Classify): Gear review -> Prompt: the label set + examples -> phi3 or gpt-4.1 -> One label

- The pipeline is identical. Only the instruction and the output shape change.
- Which is why the labs feel similar and why you should pick by what you would ship, not by difficulty

Notes: Thirty seconds. The point is that "understanding" is one HTTP call three ways, and the module thread (a wrong-shaped answer is a bug in the instruction) applies to all three.

## 4. 01 · Summarization: The User Problem

- Forty trip reports for Avalanche Lake Trail, 1,200 words each
- Somewhere in there: is the bridge out, and are the mosquitoes bad?
- Nobody reads forty essays. They skim three, miss the warning in the fourth, and have a bad Saturday.

## 5. 01 · The Concept

- The simplest LLM feature: one chat call, a document, an instruction
- "Summarize this" produces a book report
- Real products ask for a summary with a purpose: "In 3 bullets, for a hiker planning a trip this week: conditions, hazards, crowding. Ignore gear talk."
- Shape it to the UI slot: prose, bullets, template, a single headline

Notes: Small local model handles all of this. This feature never touches the cloud.

## 6. 01 · How It Works

Flow: 1,200-word trip report -> Prompt: purpose, output shape, and two grounding lines -> llama3.2 (local, no key) -> Three bullets a hiker can act on

- One request, one response. There is no pipeline to build.
- The instruction carries the feature: who it is for, what to keep, what to drop, and "only from the text"
- The output shape is a product decision: bullets for a card, a headline for a list, prose for an email

Notes: Read the prompt aloud once. Everything they build in the lab is a change to the middle box.

## 7. 01 · Demo: What to Watch

- Raw report on screen first, so the room feels the problem
- Naive prompt: faithful, generic, useless
- Purpose-built prompt: three bullets, and the bridge warning surfaces
- The last two lines of the prompt are grounding lines: without them, roughly half of runs invented a bear-related closure on the *clean* report
- Same call, different shape: a one-line trail-status headline

Notes: A required "hazards" slot with nothing honest to put in it is an invitation to make something up. Numbers are in `expected-output.md`. Point at the Ollama endpoint at the end: there was no API key, and no data left the room.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Open a raw trip report from `data/trip-reports/` on screen and scroll through it slowly, so the room feels the problem before seeing the fix.
2. Starter project: `IChatClient` wired to Ollama via Microsoft.Extensions.AI. One method, one prompt: "Summarize this trip report."
3. Run it and get a faithful, generic, useless book report, then name the failure out loud: right model, wrong instruction.
4. Iterate the prompt live into the hiker-focused version (conditions, hazards, crowding, ignore everything else). Run again. This is the payoff: three bullets, and the bridge warning surfaces.
5. Point at the last two lines of that prompt, the grounding lines. Without them, this exact prompt run against the clean report invented a bear-related trail closure in roughly half of runs, because a required "hazards" slot with nothing honest to put in it is an invitation to make something up. The numbers and the failing outputs are in `expected-output.md`.
6. Change the shape: same call, but output a one-line "trail status" headline for a card UI. Same feature, different product surface.
7. Point at the Ollama endpoint in the code. This ran entirely on the laptop, with no API key and no data leaving the room.

## 8. 01 · Leadership Card

- When: anywhere users face long content they don't want to read. Reviews, tickets, reports, meeting notes, threads.
- Cost: days. One call per document, free local models, no new infrastructure.
- "Our users are drowning in text we already have. A weekend of prompt work turns it into answers."

---

## 9. 02 · Extraction: The User Problem

- Trailhead wants a "trail stats" panel: which trails, when, how far, what wildlife, what shape the trail was in
- All of it exists, scattered through forty reports as prose
- Today a human would re-read every report and re-type the facts. So the panel doesn't exist.

## 10. 02 · The Concept

- Summarization's sibling: the output is data your code consumes, not text a person reads
- JSON mode: the model is constrained to emit valid JSON matching your schema
- The schema does most of the prompting: field names, descriptions, "use `null` when the report doesn't say"
- Structured shape guarantees the JSON parses. It does not guarantee the JSON is true.

Notes: This is where the LLM stops being a chat feature and becomes a data-pipeline component. Extraction runs on private data at volume, so free and on-prem matters here more than anywhere.

## 11. 02 · How It Works

Flow: Trip report -> Prompt: fields, types, examples, "null when absent" -> llama3.2 in JSON mode -> Parse and validate against the schema -> Store, or reject and log

- The model writes JSON; your code decides whether to believe it
- JSON mode guarantees syntax, not truth. The schema check and the null rule are the feature.
- A rejected record is a good outcome. A silently stored zero is the bug.

Notes: The last box is the one people skip. Say plainly that extraction without validation is a data-corruption feature.

## 12. 02 · Demo: What to Watch

- Define a C# record (`TripFacts`); typed response, no parsing step
- Break it: a report that never mentions distance comes back with `distance_mi: 5.0`
- Fix in the schema: nullable fields, "null when not stated"
- Run it two or three times: it gets better, and it is not a guarantee. `elevation_gain_ft: 0` where the honest answer is `null`; "early last month" as a date.
- The last mile is a validator, which is ordinary code your team already knows how to write.

Notes: Zero is the dangerous miss: it is a value, and a pipeline will store it without complaint. Ship the schema plus a rejection rule.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Show the same messy trip report from feature 01. This time the goal isn't a summary, it's a database row.
2. Define a C# record (`TripFacts`: trail, park, date, distance, wildlife, conditions) and use Microsoft.Extensions.AI's typed-response support to request it directly. The schema is code.
3. Run it: prose in, populated .NET object out, with no parsing step. This is the payoff moment.
4. Break it on purpose: run a report that never mentions distance, and watch the model invent `distance_mi: 5.0`.
5. Fix it in the schema with nullable fields and "null when not stated" descriptions. Re-run: the invented distance usually goes away, and something else usually doesn't. Run it two or three times live so the room sees the variance rather than one lucky result. Then say the quiet part: this is better, and it is not a guarantee. The last mile is a validator that rejects a date like "early last month" and a `0` that should have been `null`, which is ordinary code your team already knows how to write.
6. Zoom out: loop over ten reports and print rows. That's an ingestion pipeline in thirty lines, and the "trail stats" panel is now just a query.

## 13. 02 · Leadership Card

- When: valuable data trapped in documents. Invoices, resumes, support emails, contracts, legacy records.
- Cost: days to a working pipeline. The real work is schema design and spot-checking accuracy.
- "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it."

---

## 14. 03 · Sentiment: The User Problem

- The Cascade 65 backpack has 300 reviews. Are people happy, and what are they mad about?
- Star ratings lie: "4 stars, but the hip belt broke on day two"
- The user here is the product team, not the hiker

Notes: Track the happy/unhappy signal weekly and a defect surfaces months before returns spike. One product in the corpus has that problem buried in its reviews.

## 15. 03 · The Concept

- Classification applied to text, and the day's model-selection lesson: you don't always need the big model
- `phi3` (2 GB, free, private) vs Azure OpenAI, same prompt
- Easy reviews: expect a tie. Hard reviews (sarcasm, mixed, rating contradicts text): measure the gap.
- The decision is measurable: labeled sample, run both, count disagreements, price the errors

Notes: Most teams never run that experiment. The room runs it before lunch.

## 16. 03 · How It Works

Flow: Gear review -> Prompt: the three labels, what each means, two examples -> Model (phi3 local, or gpt-4.1 on Foundry) -> One word -> Compare with the gold label

- Same shape as extraction with a one-word output, which makes it the cleanest place to compare models
- Everything else held equal: same prompt, same reviews, same gold labels. Swap the model box and count.
- Small model for the easy pile, big model for the sarcastic slice: the number tells you where the money goes

Notes: This is the slide that sets up the "which model" argument the demo measures.

## 17. 03 · Demo: What to Watch

- One classify method, `positive | negative | mixed`; the provider swap is one DI line
- 20 easy reviews through `phi3`: fast, free, and correct
- Same 20 through Azure: nine of ten identical, and the tenth goes to the frontier model
- The hard set through both; diff the labels on screen. The disagreements are where the arguments live.
- Measured: `phi3` 9/10 easy, 7/10 hard. `gpt-4.1` on Azure: 10/10 and 10/10. The local stand-in (`llama3.2`): identical to phi3.

Notes: Both halves of the argument hold: ordinary reviews barely need the big model, and the sarcastic slice does, 7/10 vs 10/10, with every disagreement going the frontier model's way. Also worth 20 seconds: the local stand-in tied phi3, so a comparison against the wrong big model would have said the gap does not exist. Side finding worth 20 seconds: reflowing the identical prompt onto one line dropped phi3 to 7/10 and 4/10. Small models are sensitive to formatting.

Demo script (the demo outline that used to live in the spec, sized for the 30-minute block: skip the break-it-on-purpose beats first if you are running long):
1. Show the Cascade 65's reviews and point at a 4-star review with furious text, which is the whole case for reading the text rather than the stars.
2. Starter project: one classify method whose prompt returns exactly `positive | negative | mixed`. Because of Microsoft.Extensions.AI, the same code runs against both providers, and the swap is one line in DI registration. Say that out loud, since it's the provider-flexibility slide made real.
3. Run 20 easy reviews through `phi3` locally, where it is fast, free, and correct, and let the small model win the first round.
4. Run the same 20 through `gpt-4.1` on Foundry and get the same labels on nine of ten (the frontier model also gets the deadpan four-star rave that `phi3` calls mixed), which is the first payoff: on ordinary reviews you'd have paid for very little.
5. Now the hard set, sarcasm and mixed reviews. Run both and diff the labels on screen. Second payoff: `phi3` drops to 7/10 and `gpt-4.1` holds at 10/10, and every disagreement is one the frontier model gets right. This is the slice that earns its price, and you know that because you measured it rather than assumed it. Check `expected-output.md` for what happened when this was built, and be ready for the room's answer to differ from yours.
6. Close with the decision recipe: labeled sample, run both, count disagreements, price the errors. That recipe generalizes to every feature today.

## 18. 03 · Leadership Card

- When: any high-volume text stream needing a judgment call. Reviews, NPS verbatims, tickets, mentions, survey answers.
- Cost: days. The classifier is trivial; the diligence is a labeled sample and an error count.
- "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun."

---

## 19. Hands-On: 60 Minutes

- **Start with 01 Summarization** (Recommended). One endpoint, one prompt, two reports. New to this? Budget the full time and do the stretch goal.
- Then, if you have time: **02 Extraction** (more code, most reusable at work) or **03 Sentiment** (least code, most measurement; pick this if your question is "which model should we pay for")
- Every feature folder: `FNN-lab.md`, `http/ollama.http`, `expected-output.md`, and `data/`

Notes: Start with 01 regardless of experience. Everything else today assumes you've made one model call and seen what comes back. Finished everything? Help someone near you.

## 20. Debrief

- What surfaced the buried bridge, and what invented one?
- Rows 1 to 3 of the decision framework are done
- Break, then Module 2.
