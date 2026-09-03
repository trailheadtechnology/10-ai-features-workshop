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
ollama.com/download: one installer, no GPU needed.
~
The pulls, one per command: `ollama pull llama3.2` then `ollama pull phi3` then `ollama pull nomic-embed-text`.
Models are about 5 GB total.
~
REST Client from the terminal: `code --install-extension humao.rest-client`.
JetBrains users can skip REST Client; the built-in HTTP client opens the same .http files.
~
Runtimes: .NET 10 SDK, uv for Python, or Node. Pick the one you read fastest.
~
The clone: `git clone https://github.com/trailheadtechnology/10-ai-features-workshop.git`
~
Azure endpoint and deployment names are already in the lab files; the key is handed out in the room.
Ask for hands: who has all three models pulled? Anyone without them starts copying from the USB drives now, during the framing, rather than in the first lab.

## First, Prove Your Machine Works

- `http/smoke-test.http`
- 💬 local chat · 🔢 local embeddings · ☁️ Azure chat
- ✅ Three JSON responses, no red text
- 🖐️ Broken? Raise a hand now.

Notes: Full path: modules/M0-opening/F00-setup-and-framing/http/smoke-test.http.
Do this before the framing, about ten minutes in, because the most expensive twenty minutes of a hands-on workshop is the twenty minutes in the first lab when a third of the room finds out their setup doesn't work.
~
Three requests: local chat, local embeddings, Azure chat. Paste the room key over <KEY FROM INSTRUCTOR> for the Azure request.
Room key (key2): in `instructor.local.md` at the repo root (untracked). Write it on the whiteboard, not a slide. Regenerate it after the workshop; the command is in the same file.
~
Three JSON responses and no red text is a pass. Curl versions are in the lab README.
~
Fallbacks exist for exactly this moment. Walk the room while people run it.

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

Notes: Every company is asking this question right now.
~
Today asks the second one.
This is the thesis. Say it plainly and come back to it in every module: every feature today opens with a user who is stuck, and the AI only shows up as the answer.
The technology is the last thing we talk about in each feature, not the first.

## "Users" Is a Broad Word

- 🥾 The hiker planning Saturday
- 📊 The PM facing 300 reviews
- 🏞️ The ranger facing a full inbox
- Features 03, 07, and 08 might serve the people running the product.

Notes: The obvious user: the hiker from the story, planning Saturday.
~
The PM with 300 gear reviews and no time to read them.
~
The ranger with a full inbox and an emergency somewhere in it.
~
Stakeholders are users too. "Nobody has time to read all of this" is a real user problem; it just belongs to a user on your payroll.

## What You Leave Able to Do

- Purpose-shaped summaries · validated records — locally
- Embedding search · grounded answers, with verified citations
- Routing that never misses the expensive class · a human gate with an audit trail
- A bounded agent loop you can read
- See all ten · build four

Notes: Who this is for: you write software for a living, you can make an HTTP request, and you haven't shipped an LLM feature yet, or you've shipped one and want the other nine.
Say the four outcomes out loud, one per reveal; these are the promises the day is measured against, and they come back one at a time in each module's debrief. One outcome per module.
Module 1: summaries shaped for a purpose and records your code validated, all on your laptop.
~
Module 2: search by meaning, and answers grounded in your documents with citations your code checks.
~
Module 3: routing that never misses the expensive class, and a human gate with an audit trail.
~
Module 4: an agent loop bounded by a step budget, small enough to read.
~
You see all ten demoed; you build four, one per module.
If someone in the room has shipped all of this already, they are a helper for the day; say so now.

## The Modules

- 1 · Understanding — summarize, extract, sentiment
- 2 · Finding — search, RAG, recommend
- 3 · Deciding — route, detect, approve
- 4 · Doing — the agent capstone

Notes: Module 1, Understanding: summarize, extract, sentiment. 90 minutes.
~
Module 2, Finding: search, RAG, recommend. 90 minutes, then an hour for lunch.
~
Module 3, Deciding: route, detect, approve. 90 minutes.
~
Module 4, Doing: the agent capstone, everyone. 60 minutes. Then the closing, 30 minutes.
Breaks between modules and an hour for lunch after Module 2.

## [static] The Schedule

- Hand-made table in the deck (module · features · minutes); not generated

Notes: 30 · 90 · 90 · 90 · 60 · 30. Leave it up for a beat; the README has the same table.

## [about] I'm J. Tower

Notes: Quick intro, then the free offer: tinyurl.com/th-offer, on screen again at the end.

## How Each Module Works

- 30 min mine · 60 min yours
- 1 lab **Recommended** · 2 **Challenge**
- The capstone: everyone
- Finishing one lab well = success

Notes: The instructor half: the module's theme, then all three features demoed live in .NET. The build half is yours.
~
Pick the Recommended lab unless you have a reason not to; the other two are for anyone with time left, and every lab has a stretch goal.
~
The capstone splits 10 and 50 instead of 30 and 60, and there is no menu: everyone builds it.
~
Set expectations hard here. Ten labs in a day is not the goal, and nobody is behind. Helping the person next to you is a good use of leftover time.

## Any Language

- `.http` files: VS Code · JetBrains · curl
- Can you make an HTTP request? You're equipped.
- `dotnet/` · `python/` · `typescript/` — each with `starter/` and `complete/`

Notes: Every lab ships as raw .http files against Ollama and Azure OpenAI. VS Code with REST Client, JetBrains, or curl.
~
Java, Go, Rust, anything with an HTTP client is welcome.
~
The .NET projects are the demo; the Python and TypeScript ports produce the same output, so pick whichever you read fastest.

## Every Feature Ends with the Same Card

- When to reach for it
- Think: the line for your boss
- Difficulty: easy, medium, or hard
- 10 cards → the decision framework

Notes: When to reach for it: the user problem this feature answers.
~
Think: the one line you say to your boss.
~
Difficulty: easy, medium, or hard, as measured by building all ten.
~
The cards become the framework we assemble in the closing session; it's the artifact you take back to work. Point at docs/decision-framework.md in the repo now so people know it exists.

## Local First, Cloud When Needed

- 💻 `llama3.2` · `phi3` · `nomic-embed-text` — ~5 GB, no GPU
- ☁️ Foundry only where quality matters: F03, F05, F10
- Free, private, offline-tolerant

Notes: Three local models, about 5 GB, no GPU. Free, private, and they run on the laptop in front of you.
~
Foundry only where model quality changes the answer: 03's comparison, 05's generation, and the capstone.
~
Two payoffs: you learn where free/private/small is enough, and most of the day survives conference wifi.
Rows 1 through 8 of the framework mostly run on free local models. That is a finding your leadership will care about on its own.

## The Shape of Every Feature

Flow: Your data (a report, a query, an inbox) -> Prompt: instruction + data + the shape you want back -> Model, over HTTP (Ollama or Foundry) -> Output your code checks -> Your UI

- Ten features, one shape
- The model is one box; the rest is ordinary software
- Later modules add boxes, never replace them

Notes: Build the diagram box by box as you talk, and leave the picture in people's heads for the day. It starts with your data: a report, a query, an inbox.
~
A prompt: the instruction, the data, and the shape you want back.
~
The model, over HTTP. Ollama on the laptop or Foundry in the cloud; same box.
~
Output your code checks before anything trusts it.
~
Your UI. The user never sees the model.
~
Ten features, one shape. What changes per feature is the data going in, the instruction, and what your code does with what comes back.
~
The model is one box; the rest is ordinary software your team already writes.
~
Modules 2 to 4 add a vector store, a retrieval step, a tool loop. Every module opens with a version of this diagram with more boxes; point back to this one each time.

## Every Model Is an HTTP Call

- `localhost:11434` → `/api/chat` · `/api/embed`
- `…openai.azure.com/…/chat/completions` + `api-key`
- Same request shape · a URL and a key apart
- The SDKs are wrappers over these POSTs

Notes: Ollama listens on localhost:11434: /api/chat and /api/embed. It also exposes an OpenAI-compatible /v1.
~
Foundry serves gpt-4.1 and gpt-5.5 per deployment name, at the chat/completions path with an api-key header.
~
Show one request to each on screen: the smoke test's local chat and its cloud chat, side by side. Point at the URL and the header. Same request shape, a URL and a key apart.
~
The http/ track is the real thing with the SDK removed; Microsoft.Extensions.AI and the openai package are wrappers over these POSTs. This is why "any language" is true and why local-to-cloud is a config change, which the next slide makes concrete in code.

## One Line of Code, Any Provider

- `IChatClient` · `IEmbeddingGenerator`
- Ollama ↔ Azure = one DI line
- F03 and F05 (both optional) do the swap live

Notes: Every .NET demo uses Microsoft.Extensions.AI: IChatClient for chat, IEmbeddingGenerator for vectors.
~
Ollama today, Azure OpenAI tomorrow, is a change in DI registration, not in the feature. Show the two registration lines side by side.
~
03 and 05 do the swap live, and both are optional labs.
This is the one "framework" slide of the day; move on. That is all the framework talk the day needs.

## Meet Trailhead Guides

- The app from the story
- 🗺️ 200 trails · 40 trip reports
- 🎒 ~300 gear reviews · 📜 25 regulation docs
- 📥 an inquiry inbox · ~500 condition reports
- 🔌 mock weather / campsite / permit APIs
- ⚠️ 100% synthetic

Notes: This is the app from the story. Fictional, built for today, and every feature you see runs inside it.
Two minutes, not a tour; each lab doc describes its own data when people get there.
~
Two hundred trails in the catalog. Forty trip reports, about 1,200 words each, written by people who love gear more than brevity.
Switch to VS Code: modules/M1-understanding/F01-summarization/data/tr-0004.md. Scroll slowly. Do not read it aloud.
Airport sandwiches. Huckleberry ice cream. And somewhere in the middle, one sentence about a bridge. That is the problem feature 01 solves in about twenty minutes.
~
Three hundred gear reviews with star ratings that lie. Feature 03 measures how often.
Twenty-five regulation documents, written like real ones. Feature 05 answers questions from them, with citations.
~
One hundred visitor inquiries; two are emergencies, and feature 07 must never miss them.
Five hundred trail-condition reports, boring on purpose. Feature 08 finds the two clusters that are not.
~
Weather, campsites, permits: JSON files in a folder. Nothing today calls a real park. Feature 10 treats them as real.
~
Real park names, invented rules. Do not plan a trip from it.
Two facts are planted in this corpus and you will meet them all day. The footbridge on Avalanche Lake Trail, trail-0117, washed out in June 2026: her bridge from the story. And a bear-activity spike on Hidden Meadow Trail, trail-0042, late June into early July, which nobody mentions until feature 08 finds it.

## Row 0 of the Framework

- When: before any AI feature at all
- Cost: one workshop
- "List the ten things users hate doing. The features will pick themselves."

Notes: Before any AI feature at all: the framing question is the cheapest AI work your team will ever do.
~
The cost is one workshop, and it happens in a meeting room, not a codebase.
~
Read the line. Then break straight into Module 1.
