# Instructor Runsheet

The whole day on one page: what to have open, what to run, what to cut. Timings are lengths, not clock times. The spoken words live in each deck's notes; this page is the checklist that sits next to the laptop.

## Before the room opens

- Ollama running, three models present: `ollama list` shows `llama3.2`, `phi3`, `nomic-embed-text`.
- Terminal 1 at the repo root with the Azure variables loaded. Paste key1 (yours, not the room key) into `.env`, then `set -a; source .env; set +a`. Every `dotnet run` below that needs the cloud inherits it from this terminal.
- Terminal 2 with no Azure variables, for the demos that must stay local.
- All six decks open in PowerPoint, in order, presenter view checked once.
- VS Code open at the repo root, REST Client extension installed, for the `.http` labs.
- Room key (key2) written on the whiteboard before Module 0's smoke test, never on a slide. The command to fetch it is in `instructor.local.md`.
- Every .NET project built once so first runs are not compile waits: `dotnet build workshop.slnx` at the repo root builds all twenty.

## Module 0: Opening, 30 min

Deck `M0-opening.pptx`.

| At | Do |
|---|---|
| 0:00 | Title slide up while people settle. Get Set Up slide, point at `SETUP.md`. |
| 0:03 | Smoke test: everyone runs `modules/M0-opening/F00-setup-and-framing/http/smoke-test.http`. Three requests, three JSON responses. Key from the whiteboard goes into request 3. |
| 0:08 | Walk the room. Anyone red gets the fallback: local-only for the day, key later. |
| 0:12 | The Saturday 5:58 a.m. story, the promise, the wrong question and the better one. |
| 0:20 | The day, how modules work, any language, local first, Trailhead Guides, row 0. |
| 0:30 | Straight into Module 1. |

## Module 1: Understanding, 90 min

Deck `M1-understanding.pptx`. Terminal 2 for 01 and 02, Terminal 1 for 03.

| At | Do |
|---|---|
| 0:00 | This Module slide. Two minutes. |
| 0:02 | 01 Summarization, 8 min. `modules/M1-understanding/F01-summarization/dotnet/complete`. Open a report from `data/trip-reports/` first. `dotnet run`, then `dotnet run -- --headline` for the card-UI reshape. |
| 0:10 | 02 Extraction, 10 min. `modules/M1-understanding/F02-extraction/dotnet/complete`. `dotnet run`. Run the sparse report two or three times so the room sees the variance. |
| 0:20 | 03 Sentiment, 10 min. `modules/M1-understanding/F03-sentiment/dotnet/complete`, Terminal 1. `dotnet run -- --easy`, then `dotnet run -- --hard`. Expect phi3 7/10, gpt-4.1 10/10 on the hard set. If gpt-4.1 shows 9/10 on the easy set, that is `gr-0014` flickering and it is in expected-output. |
| 0:30 | Lab slide up. 01 Recommended, 02 and 03 Challenge. Sticky check at 0:50. |
| 1:25 | Debrief. Two people, what surprised them. |
| 1:30 | Break, 15 min. |

Cut order if behind: the break-it beats in 02 and 03, then the headline reshape in 01, then say a card in one sentence.

## Module 2: Finding, 90 min

Deck `M2-finding.pptx`. Terminal 2 for 04 and 06, Terminal 1 for 05.

| At | Do |
|---|---|
| 0:00 | This Module and the Same First Half diagram. Four minutes. |
| 0:04 | 04 Semantic Search, 8 min. `modules/M2-finding/F04-semantic-search/dotnet/complete`. `dotnet run`. The kids query lands on Taft Point; that is the lesson, not a bug. |
| 0:12 | 05 RAG, 12 min. `modules/M2-finding/F05-rag/dotnet/complete`, Terminal 1. In order: `dotnet run -- --no-context`, `dotnet run -- --top-k 8 --alpha 1.0 --retrieval-only`, `dotnet run -- --top-k 8 --retrieval-only`, `dotnet run`, then the EV question: `dotnet run -- "Are there EV charging stations in Glacier National Park?"`. |
| 0:24 | 06 Recommendations, 6 min. `modules/M2-finding/F06-recommendations/dotnet/complete`. `dotnet run`, then `dotnet run -- --gear` for the daypack failure. |
| 0:30 | Lab slide up. 04 Recommended, 05 and 06 Challenge. 05 needs the whiteboard key. Sticky check at 0:50. |
| 1:25 | Debrief. |
| 1:30 | Lunch, 60 min. |

Cut order: the zero-overlap second query in 04, the drone and Avalanche Lake questions in 05, the gear failure in 06, then a card in one sentence.

## Module 3: Deciding, 90 min

Deck `M3-deciding.pptx`. Terminal 2 for all three.

| At | Do |
|---|---|
| 0:00 | This Module and the Three Ways to Decide diagram. Four minutes. |
| 0:04 | 07 Classification, 8 min. `modules/M3-deciding/F07-classification-routing/dotnet/complete`. Scroll `data/inquiries.jsonl` first. `dotnet run`. inq-0030 is the planted miss; fix the category description, re-run. |
| 0:12 | 08 Anomaly Detection, 9 min. `modules/M3-deciding/F08-anomaly-detection/dotnet/complete`. `dotnet run -- --raw` first (underwhelms on purpose), then `dotnet run`. |
| 0:21 | 09 Human-in-the-Loop, 9 min. `modules/M3-deciding/F09-human-in-the-loop/dotnet/complete`. `dotnet run` is interactive: approve one, edit one, reject one. `dotnet run -- --policy` prints the routing table for the close. |
| 0:30 | Lab slide up. 07 Recommended, 08 and 09 Challenge. Sticky check at 0:50. Steer tired people to 09. |
| 1:25 | Debrief. |
| 1:30 | Break, 15 min. |

Cut order: the unsure route in 07, the `--raw` pass in 08, the audit log in 09, then a card in one sentence.

## Module 4: Doing, 60 min

Deck `M4-doing.pptx`. Terminal 1, with `AZURE_OPENAI_DEPLOYMENT=gpt-5.5` exported over the default.

| At | Do |
|---|---|
| 0:00 | This Module, the definition, the two diagrams. |
| 0:03 | Demo, 7 min. `modules/M4-doing/F10-agentic-workflows/dotnet/starter`: `dotnet run` for the fluent, useless plain completion. Then `dotnet/complete`: `dotnet run` and narrate the trace. Without `--yes` the permit gate waits for you, which is the point. For the washout, use the request exactly as written in the lab: it names `trail-0117`. |
| 0:10 | Lab slide up. Everyone does 10. Whiteboard gets the key and the deployment name `gpt-5.5`. Sticky check at 0:30. |
| 0:55 | Debrief, then hand off to the closing deck. |

Cut order: the campsites-full re-run (it is a hand edit of `data/mock-apis/campsites.json`, so it is the first cut anyway), then the card in one sentence.

Known behavior from the soak test: the rain day sometimes gets a moderate hike instead of an easy one. The hard hike still lands on a dry day. Say "it shaped the plan, it didn't perfect it."

## Closing, 30 min

Deck `M5-closing.pptx`. Framework table from `docs/decision-framework.md` on screen. Q&A. Thanks slide.

## If the network dies

- Module 0: skip request 3, tell the room the key will work when the network does.
- 03 Sentiment: the code falls back to llama3.2 on its own. The story changes from "frontier earns its price" to "the bigger local model is one review better." Say that out loud; the numbers are in expected-output.
- 05 RAG: retrieval is local. Generation falls back to llama3.2; the answers are still right on the four questions, with weaker citations.
- 10 Agent: llama3.2 is not reliable enough to demo. Read `reference-transcript.md` on screen instead and narrate it as the trace.

## After the workshop

Regenerate key2 the same evening. The command is in `instructor.local.md`.
