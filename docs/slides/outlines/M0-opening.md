# Deck 0: Opening (about 30 minutes)

Format: `## [marker] Title`, then the bullets that appear on the slide, then `Notes:` for what to say. Markers: `[title]`/`[promise]`/`[about]`/`[thanks]` clone the themed boilerplate slides, `[section]` and `[demo]` are full-bleed divider slides, `[big]` is one centered statement, `[static]` shows a list all at once; everything else reveals item by item across consecutive slides. See `build.py`.

This deck is hand-maintained in PowerPoint, and this outline was reverse-engineered from it on 2026-09-17. `build.py` no longer builds it; edit the deck, then update this file to match.

On-slide bullets are cues and tokens, not sentences: everything J. says lives in the Notes.

Run order: title while people settle, the story, the problem and the promise, what you leave able to do, about J., then environment setup and the smoke test, which users, modules and schedule, languages, models, and meet the app and its data.

---

## [title] Building Pragmatic AI

- 10 AI Features Your Users Actually Want

Notes: The repo URL comes up on the Get Set Up slide; advance to it while people settle.
Ask who did the SETUP.md pre-work. Anyone who didn't should say so now, not in the first lab.

## [big] Saturday. 5:58 a.m.

Notes: A hiker. Alarm before dawn, coffee in the car, ninety minutes to the trailhead.
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

## [big] Mile 2 is where she found out.

Notes: Two hours of driving, two miles of walking, and the trip is over at a creek she can't cross.
Beat of silence, then the turn:

## [big] The answer was there, but it was buried in a mountain of text.

Notes: The data was in the product. A user needed one sentence of it, and the product's job was to surface that sentence.
Every feature you build today exists to close exactly this kind of gap.
And this bridge, this trail, this body of data: you'll meet them all day.
That washout is planted in the workshop data, and by this afternoon your code will find it.

## [section] The Problem

## The Wrong Question and the Better One

- ❌ "Where can we use AI?"
- ✅ "What problems can AI best solve for our users?"

Notes: Every company is asking this question right now.
~
Today asks the second one.
This is the thesis. Say it plainly and come back to it in every module: every feature today opens with a user who is stuck, and the AI only shows up as the answer.
The technology is the last thing we talk about in each feature, not the first.

## [static] Ten shippable AI features that solve problems your users already have, and the judgment to know when to build each.

Notes: The promise of the day.
Not a survey, not vendor slides: you build these, you measure them, and you leave with a decision framework for picking the first one to ship at work.

## What You Leave Able to Do

Icon grid: one icon above a two-line label per advance, logo bottom-right.

- Summaries & Validations
- Search & Cited Answers
- Routing & a Human Gate
- A Bounded Agent Loop
- See 10; Build 4+

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

## [about] I'm J. Tower

Notes: Quick intro, then the free offer: tinyurl.com/th-offer, on screen again at the end.

## [section] Environment Setup

## Get Set Up

- 🦙 ollama.com/download
- ⬇️ `ollama pull` · `llama3.2` · `phi3` · `nomic-embed-text`
- 🧰 VS Code + REST Client (`humao.rest-client`)
- ⌨️ C#, Python, or TypeScript, and its runtime
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

## [static] Verify Your Setup

- `/modules/M0-opening/F00-setup-and-framing/http/smoke-test.http`
- 💬 Local chat
- 🔢 Local embedding
- ☁️ Azure chat
- ✅ Three JSON responses, no red text
- 🖐️ Broken? Raise a hand now.

Notes: Full path: `modules/M0-opening/F00-setup-and-framing/http/smoke-test.http`. Paste the room key over <KEY FROM INSTRUCTOR> for the Azure request.
Room key (key2): in instructor.local.md at the repo root (untracked). Write it on the whiteboard, not a slide. Regenerate it after the workshop; the command is in the same file.
Do this before the framing, about ten minutes in, because the most expensive twenty minutes of a hands-on workshop is the twenty minutes in the first lab when a third of the room finds out their setup doesn't work.
Fallbacks exist for exactly this moment. Walk the room while people run it. Curl versions are in the lab README.

## [section] Which "Users"?

## "Users" Is a Broad Word

- 🥾 The hiker planning Saturday
- 📊 An admin facing 300 reviews
- 🏞️ The ranger facing a full inbox

Notes: The obvious user: the hiker from the story, planning Saturday.
~
The PM with 300 gear reviews and no time to read them.
~
The ranger with a full inbox and an emergency somewhere in it.

## [section] Modules and Schedule

## The Modules

- 1 · Understanding: summarize, extract, sentiment

Notes: Module 1, Understanding: summarize, extract, sentiment. 90 minutes.

## [static] The Schedule

| Module | Features | Time |
|---|---|---|
| 00 Opening | Setup, context | 30 min |
| 01 Understanding | F01 Summarize, F02 Extract, F03 Sentiment | 90 min |
| 02 Finding | F04 Search, F05 RAG, F06 Recommend | 90 min |
| 03 Deciding | F07 Route, F08 Detect, F09 Approve | 90 min |
| 04 Doing | F10 Agent capstone | 60 min |
| 05 Closing | A decision framework | 30 min |

Notes: Breaks between modules and an hour for lunch after Module 2.

## [static] How Each Module Works

- 3 90-Min Modules
-   3 AI features per module
-   20 min mine for overview
-   70 min yours for labs
-   1 lab Recommended · 2 Challenge
- A 60-minute capstone module
-   1 AI feature
-   1 lab for everyone
- Recap
- NOTE: Breaks any time, especially during labs

Notes: The instructor half: the module's theme, then all three features demoed live in .NET. The build half is yours.

## [section] Languages and Frameworks

## Language Choices

Three logos: .NET, Python, TypeScript.

Notes: The .NET projects are the demo; the Python and TypeScript ports produce the same output, so pick whichever you read fastest.
~
The .NET projects are the demo; the Python and TypeScript ports produce the same output, so pick whichever you read fastest.

## [section] Models

## [static] Local Models, Cloud When Needed

- Local Models (~5 GB, no GPU):
-   💻 `llama3.2`
-   💻 `phi3`
-   💻 `nomic-embed-text`
- Cloud Model (my subscription):
-   ☁️ Cloud models where it helps: F03, F05, F10*
- *NOTE: F10 is the only recommended lab, others are optional

Notes: Two payoffs: you learn where free/private/small is enough, and most of the day survives conference wifi.
Rows 1 through 8 of the framework mostly run on free local models. That is a finding your leadership will care about on its own.

## Every Model Is an HTTP Call

- `localhost:11434` → `/api/chat` · `/api/embed`
- …openai.azure.com/…/chat/completions + `api-key`
- Same request shape · a URL and a key apart
- The SDKs are wrappers over these POSTs

Notes: Ollama listens on localhost:11434: /api/chat and /api/embed. It also exposes an OpenAI-compatible /v1.
~
Foundry serves gpt-4.1 and gpt-5.5 per deployment name, at the chat/completions path with an api-key header.
~
Show one request to each on screen: the smoke test's local chat and its cloud chat, side by side. Point at the URL and the header. Same request shape, a URL and a key apart.
~
Feature 00's smoke test sends these requests raw, with the SDK removed; Microsoft.Extensions.AI and the openai package are wrappers over these POSTs. This is why the three language tracks look alike and why local-to-cloud is a config change, which the next slide makes concrete in code.

## [section] Meet The App

## [static] Meet Trailhead Guides

Picture: `assets/app-mockups/trailhead-guides-trail-devices.png` (laptop and phone mockup of the app).

Notes: This is the app from the story. Fictional, built for today, and every feature you see runs inside it.
Two minutes, not a tour; each lab doc describes its own data when people get there.

## [static] Meet the Data

- 🗺️ 200 trails
- 📃 40 trip reports
- 🎒 300 gear reviews
- 📜 25 regulation docs
- 📥 A full email inbox
- 🌧️ 500 condition reports
- 🔌 Mock APIs (weather / campsites / permits)
- ⚠️ 100% synthetic

Notes: Three hundred gear reviews with star ratings that lie. Feature 03 measures how often.
Twenty-five regulation documents, written like real ones. Feature 05 answers questions from them, with citations.

