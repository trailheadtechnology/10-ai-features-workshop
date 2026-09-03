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
01: forty trip reports become a three-bullet briefing.
~
02: the same prose becomes a database row, and your code decides whether to believe it.
~
03: three hundred reviews get a label, and you measure which model deserves the money.
The thread to watch all module: asked for something it cannot find, a model supplies something plausible rather than nothing. 01 invents a closure, 02 invents a distance, 03 believes the sarcasm. A better model does not fix it; a tighter instruction, a null-able schema, and checking code do.
Set the rhythm expectation out loud: three quick demos with a card for your boss after each, then the room builds for an hour: 01 recommended, 02 and 03 for anyone with time left.

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

Notes: ~8 min · `cd modules/M1-understanding/F01-summarization/dotnet`
Files: `tr-0004.md` = the story's report, bridge buried mid-text · `tr-0001.md` = clean report, nothing closed (hallucination bait)
Flags: none = naive "Summarize this" · `--briefing` = 3 hiker bullets · `--headline` = one-line trail card · `--audience ranger` = same report, different reader
1. Open `data/tr-0004.md`, scroll. The bridge is one sentence, mid-report.
2. Before: `cd starter && dotnet run -- ../../data/tr-0004.md`. Book report. Right model, wrong instruction.
3. After: `cd ../complete && dotnet run -- --briefing`. Same report, three bullets, washout leads. Show the prompt.
4. `dotnet run -- --briefing ../../data/tr-0001.md`: the clean report, nothing closed. Last two prompt lines: 11 of 24 became 1 of 24.
5. `dotnet run -- --headline`: one line for a card.
6. Line 11: `localhost:11434`. No key, nothing left the room.
Cut: 5, then 4.

## F01 · Summarization · Leadership Card

- **When:** Too much content for people to read, easy-to-miss detail buried in avalanche of text.
- **Think:** "Our users are drowning in text we already have. A few hours of work turns it into automatic answers."

Difficulty: easy

Notes: Reviews, tickets, reports, meeting notes, threads. One call per document.
~
The line for your boss; read it. Row 1 of the framework.
~
Easy. If asked about cost: days of work, free local models, no new infrastructure.

---

## [section] F02 · Extraction

Notes: The user problem, spoken, 60 seconds: Trailhead wants a "trail stats" panel — which trails, when, how far, what wildlife, what condition. All of it exists as prose across the forty reports, and today a human would re-read and re-type it. So the panel doesn't exist.
This one gets the module's one concept slide, because the pipeline — and where it lies to you — is the lesson.

## F02 · How It Works

Flow: Trip report -> Prompt: fields, types, examples, "null when absent" -> llama3.2 in JSON mode -> Parse and validate against the schema -> Store, or reject and log

- The model writes JSON; your code decides whether to believe it
- The `null` rule and the validator are the feature
- A rejected record is a good outcome

Notes: The input is prose: the same trip report from 01.
~
The prompt is the schema in words: fields, types, examples, and "null when absent."
~
llama3.2 in JSON mode. JSON mode guarantees syntax, not truth.
~
Parse and validate against the schema. This is the box people skip.
~
Store, or reject and log. A silently stored zero is the bug.
~
The model writes JSON; your code decides whether to believe it.
~
The null rule and the validator are the feature. Say plainly that extraction without validation is a data-corruption feature.
~
A rejected record is a good outcome; it is the pipeline telling you the truth.

## [demo] **DEMO** · F02 Extraction

## [static] F02 · Extraction · Demo: What to Watch

- A C# record (`TripFacts`) · typed response · no parsing
- ❌ No distance in the report → `distance_mi: 5.0`
- ✅ Nullable fields + "null when not stated"
- Run it 3×: better ≠ guaranteed
- The last mile is a validator — ordinary code

Notes: ~10 min · `cd modules/M1-understanding/F02-extraction/dotnet`
Files: `tr-0007.md` = full report, every field present · `tr-0011.md` = sparse report, no distance stated (the null test)
Args: a report path; complete with no args runs both
1. Same kind of report as 01. This time the goal is a database row.
2. Before: `cd starter && dotnet run -- ../../data/tr-0007.md`. JSON asked for in the prompt. Fences, preamble, drifting field names.
3. After: `cd ../complete && dotnet run -- ../../data/tr-0007.md`. `TripFacts` record, typed response, no parsing.
4. `dotnet run -- ../../data/tr-0011.md`, three times. Nulls where there is nothing, and watch what still slips.
5. Point at the validator: PASS or REJECT. A rejected record is a good outcome.
Cut: 4's repeats.

## F02 · Extraction · Leadership Card

- **When:** Valuable data trapped in documents nobody will re-type.
- **Think:** "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it."

Difficulty: medium

Notes: Invoices, resumes, support emails, contracts, legacy records.
~
The line for your boss. Row 2.
~
Medium. If asked about cost: days to a working pipeline; the real work is schema design and spot-checking accuracy.

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

Notes: ~10 min · `cd modules/M1-understanding/F03-sentiment/dotnet`
Files: `easy.jsonl` = 10 plain reviews · `hard.jsonl` = 10 sarcastic or split reviews · `reference-labels.json` = the hand labels
Flags: `--easy` / `--hard` = one set; none = both · starter takes a review id (default gr-0007, the sarcastic one)
1. Open `data/hard.jsonl`. Find a two-star review with glowing text. Stars lie.
2. Before: `cd starter && dotnet run`. One method, one label, gr-0007. Show the DI line: the swap is one line.
3. After: `cd ../complete && dotnet run -- --easy`. phi3 9/10, gpt-4.1 10/10. Paid for one review.
4. `dotnet run -- --hard`: phi3 7/10, gpt-4.1 10/10. Diff on screen. Every miss goes the frontier model's way.
5. The recipe: labeled sample, run both, count disagreements, price the errors.
Cut: 3, go straight to 4.

## F03 · Sentiment · Leadership Card

- **When:** A high-volume text stream needing a judgment call.
- **Think:** "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun."

Difficulty: easy

Notes: Reviews, NPS verbatims, tickets, mentions, survey answers.
~
The line for your boss, and today you measured it. Row 3.
~
Easy. If asked about cost: days; the classifier is trivial, the diligence is a labeled sample and an error count.

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
