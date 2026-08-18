# Deck 0: Opening (about 30 minutes)

Format for every slide below: **title**, then the bullets that appear on the slide, then `Notes:` for what to say. Slides are plain black-on-white; no imagery yet.

---

## 1. Building Pragmatic AI: 10 AI Features Your Users Actually Want

- WeAreDevelopers World Congress North America · September 23, 2026
- J. Tower · Trailhead Technology Partners
- github.com/trailheadtechnology/10-ai-features-workshop

Notes: Get the repo URL on screen and leave it there while people settle. Ask who did the SETUP.md pre-work. Anyone who didn't should say so now, not in the first lab.

## 2. Install and Setup

- Ollama, from ollama.com/download, with three models pulled: `llama3.2`, `phi3`, `nomic-embed-text` (about 5 GB, no GPU)
- VS Code with the REST Client extension (`humao.rest-client`): every lab's `.http` file runs from it with one click
- A language extension for whatever you'll code in: C# Dev Kit, Python, or the built-in TypeScript support
- The runtime to match: .NET 10 SDK, Python 3, or Node
- A clone of the repo: github.com/trailheadtechnology/10-ai-features-workshop
- Azure OpenAI: nothing to install; the endpoint and deployment names are already in the lab files, and the key is handed out in the room

Notes: This is SETUP.md on one slide. Most people did it; this is for the ones who didn't and for confirming what "done" means. JetBrains users can skip REST Client, since the built-in HTTP client opens the same `.http` files. Ask for hands: who has all three models pulled? Anyone without them starts copying from the USB drives now, during the framing, rather than in the first lab.

## 3. First, Prove Your Machine Works

- Open `modules/M0-opening/F00-setup-and-framing/http/smoke-test.http`
- Three requests: local chat (`llama3.2`), local embeddings (`nomic-embed-text`), Azure chat (paste the room key over `<KEY FROM INSTRUCTOR>`)
- Three JSON responses, no red text = cleared for the day
- Broken? Raise a hand now. Fallbacks exist for exactly this moment.

Notes: Do this before the framing, about ten minutes in, because the most expensive twenty minutes of a hands-on workshop is the twenty minutes in the first lab when a third of the room finds out their setup doesn't work. Walk the room while people run it. Curl versions are in the lab README.

## 4. The Wrong Question and the Better One

- Every company right now: "Where can we add AI?"
- Today: "What problems can AI best solve for our users?"
- Every feature today opens with a user who is stuck, and the AI only shows up as the answer

Notes: This is the thesis. Say it plainly and come back to it in every module. The technology is the last thing we talk about in each feature, not the first.

## 5. "Users" Is a Broad Word

- The hiker planning Saturday
- The product manager who can't read 300 reviews
- The ranger staring at a full inbox
- Three of today's ten features (03, 07, 08) serve the people running the product

Notes: Stakeholders are users too. "Nobody has time to read all of this" is a real user problem; it just belongs to a user on your payroll.

## 6. Who This Is For, and What You Leave Able to Do

- You write software for a living, you can make an HTTP request, and you haven't shipped an LLM feature yet (or you've shipped one and want the other nine)
- By the end of the day you will be able to:
  - Turn a document into a purpose-shaped summary and a validated record with a local model
  - Build embedding search over your own catalog and ground answers in your own documents, with verified citations
  - Route an inbox so the expensive class is never missed, and gate anything irreversible behind a human with an audit trail
  - Wire tools into a bounded agent loop and read its trace
- One outcome per module. You see all ten features; you build four yourself.

Notes: Say the four out loud; these are the promises the day is measured against, and they come back one at a time in each module's debrief. If someone in the room has shipped all of this already, they are a helper for the day; say so now.

## 7. The Day

- Module 1: Understanding (90 min): summarization, extraction, sentiment
- Module 2: Finding (90 min): semantic search, RAG, recommendations
- Module 3: Deciding (90 min): classification & routing, anomaly detection, human-in-the-loop
- Module 4: Doing (60 min): agentic workflows (the capstone)
- Closing (30 min): the decision framework and pitching this to leadership

Notes: Breaks between modules and an hour for lunch after Module 2.

## 8. How Each Module Works

- The first 30 minutes are mine: the module's theme, then all three features demoed live in .NET
- The next 60 minutes are yours to build (the capstone is 10 and 50)
- One feature per module is **Recommended**: start there unless you have a reason not to
- The other two are **Challenge**: for anyone with time left
- The capstone is the one lab everyone does
- Every lab has a stretch goal beyond that

Notes: Set expectations hard here. Finishing one lab properly is the intended outcome for most people. Ten labs in a day is not the goal, and nobody is behind. Helping the person next to you is a good use of leftover time.

## 9. Any Language

- Every lab ships as raw `.http` files (VS Code REST Client, JetBrains) against Ollama and Azure OpenAI
- If your language can make an HTTP request, you're equipped
- Following along in code? Each feature has `dotnet/`, `python/`, and `typescript/`, each with a `starter/` and a `complete/`

Notes: Java, Go, Rust, anything with an HTTP client is welcome. The .NET projects are the demo; the Python and TypeScript ports produce the same output, so pick whichever you read fastest.

## 10. Every Feature Ends with the Same Card

- When to reach for this
- Rough cost & effort
- The one-liner for your CTO
- Those ten cards become the decision framework we assemble in the closing session

Notes: This is the artifact you take back to work. Point at `docs/decision-framework.md` in the repo now so people know it exists.

## 11. Local First, Cloud Where It Earns It

- Ollama on your laptop: `llama3.2`, `phi3`, `nomic-embed-text` (~5 GB total, no GPU)
- Frontier models on Microsoft Foundry only where quality genuinely matters: the sentiment comparison, RAG generation, the capstone
- Two payoffs: you learn where free/private/small is enough, and most of the day survives conference wifi

Notes: Rows 1 through 8 of the framework mostly run on free local models. That is a finding your leadership will care about on its own.

## 12. Every Model Is an HTTP Call

- Ollama is a local server on `localhost:11434`: native `/api/chat` and `/api/embed`, plus an OpenAI-compatible `/v1`
- Foundry serves `gpt-4.1` and `gpt-5.5` at `…openai.azure.com/openai/deployments/<name>/chat/completions` with an `api-key` header
- Same request shape both places, so the SDKs differ by a constructor and the `.http` files differ by a URL and a key
- The `http/` track is the real thing with the SDK removed; Microsoft.Extensions.AI and the `openai` package are wrappers over these POSTs

Notes: Show one request to each on screen: the smoke test's local chat and its cloud chat, side by side. Point at the URL and the header. This is why "any language" is true and why local-to-cloud is a config change, which the next slide makes concrete in code.

## 13. One Line of Code, Any Provider

- Every .NET demo uses Microsoft.Extensions.AI: `IChatClient`, `IEmbeddingGenerator`
- Ollama today, Azure OpenAI tomorrow, is a change in DI registration, not in the feature
- Feature 03 does the swap live; feature 05 does it again for generation

Notes: This is the one "framework" slide of the day. Show the two registration lines side by side, and then move on; that is all the framework talk the day needs.

## 14. Meet Trailhead Guides

- A fictional national-park trip-planning app, one deliberately messy corpus, reused all day; each feature's `data/` holds what its lab reads
- 40 rambling trip reports · ~300 gear reviews · 200 trails · 25 park regulation docs · a visitor-inquiry inbox · ~500 trail-condition reports · mock weather/campsite/permit APIs
- Everything is synthetic. Do not plan a real trip from it.

Notes: Two minutes, not a tour; each lab doc describes its own data when people get there. Open one trip report on screen and scroll it slowly; that is the problem feature 01 is about to solve. Mention two planted facts people will meet repeatedly: the washed-out footbridge on Avalanche Lake Trail (trail-0117), and a bear-activity spike two trails over.

## 15. Row 0 of the Framework

- When: before any AI feature at all
- Cost: one workshop's worth of attention
- The one-liner: "Before we pick an AI feature, let's list the ten things our users hate doing. The features will pick themselves."

Notes: The framing question is the cheapest AI work your team will ever do, and it happens in a meeting room, not a codebase. Then break straight into Module 1.
