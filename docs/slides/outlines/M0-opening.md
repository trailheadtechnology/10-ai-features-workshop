# Deck 0: Opening (about 30 minutes)

Format: `## [marker] Title`, then the bullets that appear on the slide, then `Notes:` for what to say. Markers: `[title]`/`[promise]`/`[about]`/`[thanks]` clone the themed boilerplate slides, `[section]` and `[demo]` are full-bleed divider slides, `[big]` is one centered statement, `[static]` shows a list all at once; everything else reveals item by item across consecutive slides. See `build.py`.

On-slide bullets are cues and tokens, not sentences: everything J. says lives in the Notes.

Run order: repo slide up while people settle, setup + smoke test in the first ten minutes, then the story and framing.

---

## [title] Building Pragmatic AI

- 10 AI Features Your Users Actually Want

Notes: The repo URL comes up on the Get Set Up slide; advance to it while people settle.
Ask who did the SETUP.md pre-work. Anyone who didn't should say so now, not in the first lab.

## Get Set Up

- 🦙 ollama.com/download
- ⬇️ `ollama pull` · `llama3.2` · `phi3` · `nomic-embed-text`
- 🧰 VS Code + REST Client (`humao.rest-client`)
- ⌨️ C#, Python, or TypeScript — and its runtime
- 📦 `git clone` github.com/trailheadtechnology/10-ai-features-workshop
- ☁️ Azure OpenAI: nothing to install

Notes: This is SETUP.md, revealed line by line; the finished checklist stays up.
The pulls, one per command: `ollama pull llama3.2` then `ollama pull phi3` then `ollama pull nomic-embed-text`.
REST Client from the terminal: `code --install-extension humao.rest-client`.
The clone: `git clone https://github.com/trailheadtechnology/10-ai-features-workshop.git`
Models are about 5 GB total, no GPU needed. Runtimes: .NET 10 SDK, Python 3, or Node.
JetBrains users can skip REST Client; the built-in HTTP client opens the same .http files.
Azure endpoint and deployment names are already in the lab files; the key is handed out in the room.
Ask for hands: who has all three models pulled? Anyone without them starts copying from the USB drives now, during the framing, rather than in the first lab.

## First, Prove Your Machine Works

- `http/smoke-test.http`
- 💬 local chat · 🔢 local embeddings · ☁️ Azure chat
- ✅ Three JSON responses, no red text
- 🖐️ Broken? Raise a hand now.

Notes: Full path: modules/M0-opening/F00-setup-and-framing/http/smoke-test.http. Paste the room key over <KEY FROM INSTRUCTOR> for the Azure request.
Room key (key2): in `instructor.local.md` at the repo root (untracked). Write it on the whiteboard, not a slide. Regenerate it after the workshop; the command is in the same file.
Do this before the framing, about ten minutes in, because the most expensive twenty minutes of a hands-on workshop is the twenty minutes in the first lab when a third of the room finds out their setup doesn't work.
Fallbacks exist for exactly this moment. Walk the room while people run it. Curl versions are in the lab README.

## [big] Saturday. 5:58 a.m.

Notes: Everyone's settled, machines are proven. Now the story that frames the whole day.
A hiker. Alarm before dawn, coffee in the car, ninety minutes to the trailhead.
She planned this trip in a trip-planning app. (It's our fictional app, Trailhead Guides; the room will live in its data all day.)

## [big] Avalanche Lake Trail

Notes: The trail she picked. Moderate, family-rated, a lake at the end.
The app showed photos, distance, elevation. All accurate.

## [big] 41 trip reports

Notes: The app also had forty-one trip reports for this trail. Real hikers, real conditions, about 1,200 words each.
Ask the room: who reads forty-one essays before a day hike?

## [big] She read three.

Notes: The top three. "Beautiful." "Crowded by ten." "Bring bug spray."

## [big] "The footbridge at the creek is gone. We had to turn around."

- Report #14 of 41

Notes: Report fourteen. One sentence, buried in a paragraph about huckleberries.
She never saw it.

## [big] Mile two. That's where she found out.

Notes: Two hours of driving, two miles of walking, and the trip is over at a creek she can't cross.
Beat of silence, then the turn:

## [big] The answer existed. The product buried it.

Notes: This is not an AI story yet. This is a product failure story.
The data was in the product. A user needed one sentence of it, and the product's job was to surface that sentence.
Every feature you build today exists to close exactly this kind of gap. And this bridge, this trail, this corpus: you'll meet them all day. That washout is planted in the workshop data, and by this afternoon your code will find it.

## [promise] Ten **shippable** AI features that solve problems your users **already have** — and the judgment to know **which to build first**

Notes: The promise of the day.
Not a survey, not vendor slides: you build these, you measure them, and you leave with a decision framework for picking the first one to ship at work.

## The Wrong Question and the Better One

- ❌ "Where can we add AI?"
- ✅ "What problems can AI best solve for our users?"

Notes: Every company is asking the first question right now. Today asks the second.
This is the thesis. Say it plainly and come back to it in every module: every feature today opens with a user who is stuck, and the AI only shows up as the answer.
The technology is the last thing we talk about in each feature, not the first.

## "Users" Is a Broad Word

- 🥾 The hiker planning Saturday
- 📊 The PM facing 300 reviews
- 🏞️ The ranger facing a full inbox
- Features 03 · 07 · 08 serve the people running the product

Notes: Stakeholders are users too. "Nobody has time to read all of this" is a real user problem; it just belongs to a user on your payroll.

## What You Leave Able to Do

- Purpose-shaped summaries · validated records — locally
- Embedding search · grounded answers, with verified citations
- Routing that never misses the expensive class · a human gate with an audit trail
- A bounded agent loop you can read
- See all ten · build four

Notes: Who this is for: you write software for a living, you can make an HTTP request, and you haven't shipped an LLM feature yet, or you've shipped one and want the other nine.
Say the four outcomes out loud, one per reveal; these are the promises the day is measured against, and they come back one at a time in each module's debrief. One outcome per module.
If someone in the room has shipped all of this already, they are a helper for the day; say so now.

## The Day

- 1 · Understanding — summarize, extract, sentiment (90 min)
- 2 · Finding — search, RAG, recommend (90 min)
- 3 · Deciding — route, detect, approve (90 min)
- 4 · Doing — the agent capstone (60 min)
- Closing — the decision framework (30 min)

Notes: Breaks between modules and an hour for lunch after Module 2.

## [about] I'm J. Tower

Notes: Quick intro, then the free offer: tinyurl.com/th-offer, on screen again at the end.

## How Each Module Works

- 30 min mine · 60 min yours
- 1 lab **Recommended** · 2 **Challenge**
- The capstone: everyone
- Finishing one lab well = success

Notes: The instructor half: the module's theme, then all three features demoed live in .NET. The build half: pick the Recommended lab unless you have a reason not to; the other two are for anyone with time left. The capstone splits 10 and 50 instead of 30 and 60, and every lab has a stretch goal.
Set expectations hard here. Ten labs in a day is not the goal, and nobody is behind. Helping the person next to you is a good use of leftover time.

## Any Language

- `.http` files: VS Code · JetBrains · curl
- Can you make an HTTP request? You're equipped.
- `dotnet/` · `python/` · `typescript/` — each with `starter/` and `complete/`

Notes: Every lab ships as raw .http files against Ollama and Azure OpenAI. Java, Go, Rust, anything with an HTTP client is welcome.
The .NET projects are the demo; the Python and TypeScript ports produce the same output, so pick whichever you read fastest.

## Every Feature Ends with the Same Card

- When to reach for it
- Cost & effort
- The CTO one-liner
- 10 cards → the decision framework

Notes: The cards become the framework we assemble in the closing session; it's the artifact you take back to work. Point at docs/decision-framework.md in the repo now so people know it exists.

## Local First, Cloud Where It Earns It

- 💻 `llama3.2` · `phi3` · `nomic-embed-text` — ~5 GB, no GPU
- ☁️ Foundry only where quality matters: 03, 05, the capstone
- Free, private, offline-tolerant

Notes: Two payoffs: you learn where free/private/small is enough, and most of the day survives conference wifi.
Rows 1 through 8 of the framework mostly run on free local models. That is a finding your leadership will care about on its own.

## The Shape of Every Feature

Flow: Your data (a report, a query, an inbox) -> Prompt: instruction + data + the shape you want back -> Model, over HTTP (Ollama or Foundry) -> Output your code checks -> Your UI

- Ten features, one shape
- The model is one box; the rest is ordinary software
- Later modules add boxes, never replace them

Notes: Build the diagram box by box as you talk, and leave the picture in people's heads for the day. What changes per feature is the data going in, the instruction, and what your code does with what comes back. Modules 2 to 4 add a vector store, a retrieval step, a tool loop.
Every module opens with a version of this diagram with more boxes; point back to this one each time.

## Every Model Is an HTTP Call

- `localhost:11434` → `/api/chat` · `/api/embed`
- `…openai.azure.com/…/chat/completions` + `api-key`
- Same request shape · a URL and a key apart
- The SDKs are wrappers over these POSTs

Notes: Ollama also exposes an OpenAI-compatible /v1; Foundry serves gpt-4.1 and gpt-5.5 per deployment name.
Show one request to each on screen: the smoke test's local chat and its cloud chat, side by side. Point at the URL and the header.
The http/ track is the real thing with the SDK removed; Microsoft.Extensions.AI and the openai package are wrappers over these POSTs. This is why "any language" is true and why local-to-cloud is a config change, which the next slide makes concrete in code.

## One Line of Code, Any Provider

- `IChatClient` · `IEmbeddingGenerator`
- Ollama ↔ Azure = one DI line
- 03 and 05 do the swap live

Notes: Every .NET demo uses Microsoft.Extensions.AI. Ollama today, Azure OpenAI tomorrow, is a change in DI registration, not in the feature.
This is the one "framework" slide of the day. Show the two registration lines side by side, and then move on; that is all the framework talk the day needs.

## Meet Trailhead Guides

- The app from the story
- 🗺️ 200 trails · 40 trip reports
- 🎒 ~300 gear reviews · 📜 25 regulation docs
- 📥 an inquiry inbox · ~500 condition reports
- 🔌 mock weather / campsite / permit APIs
- ⚠️ 100% synthetic

Notes: A fictional national-park trip-planning app, one deliberately messy corpus, reused all day; each feature's data/ holds what its lab reads. Do not plan a real trip from it.
Two minutes, not a tour; each lab doc describes its own data when people get there.
Open one trip report on screen and scroll it slowly; that is the problem feature 01 is about to solve.
Mention the two planted facts people will meet repeatedly: the washed-out footbridge on Avalanche Lake Trail (trail-0117), from the story, and a bear-activity spike two trails over.

## Row 0 of the Framework

- When: before any AI feature at all
- Cost: one workshop
- "List the ten things users hate doing. The features will pick themselves."

Notes: The framing question is the cheapest AI work your team will ever do, and it happens in a meeting room, not a codebase.
Then break straight into Module 1.
