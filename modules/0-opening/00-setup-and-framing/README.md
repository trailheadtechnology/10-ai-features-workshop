# 00 · Setup & Framing

Module 0: Opening (9:00-9:30) · Runs on your laptop, before anything else does

## Why this feature exists

The most expensive twenty minutes of any hands-on workshop is the twenty minutes in Module 1 when a third of the room discovers their environment doesn't work. This feature spends its time up front instead: everyone runs one smoke test, broken setups surface while there's still slack to fix them, and the room starts Module 1 together.

It's also where the day gets its thesis. Every company right now is asking "where can we add AI?" This workshop spends seven hours practicing the better question: "what problems can AI best solve for my users?" Each of the ten features that follow opens with a user who is stuck, and the AI only shows up as the answer to that user's problem.

"Users" is defined broadly on purpose. The product manager who can't read every review counts, and so does the ranger staring down a full inbox. Three of the ten features (03, 07, and 08) solve problems for the people running the product rather than the people using it. Spotting problems early is a user problem too; it just belongs to a user on your payroll.

## How the day works (5 min)

- Five modules: **Opening**, **Understanding**, **Finding**, **Deciding**, **Doing**. Three 30-minute features per module after this one, except Doing, which is a single double-length agent capstone.
- Every feature has the same beat: 2 minutes on the user problem, a 13-minute live demo in .NET, a 13-minute lab you can do in any language, and 2 minutes on the leadership angle.
- Every leadership beat is a row in the [decision framework](../../../docs/decision-framework.md) we assemble at the end of the day. That document is the thing you take back to your CTO.
- Labs ship as raw `.http` files against Ollama and Azure OpenAI. If your language can make an HTTP request, you're equipped.

## Meet the corpus (5 min)

All ten features work through **Trailhead Guides**, a fictional national-park trip-planning app with a deliberately messy corpus in [`data/`](../../../data/): rambling trip reports, opinionated gear reviews, hundreds of trail descriptions, dry park regulations, and a queue of visitor inquiries. Each feature stands alone, but they all live in the same park, so by mid-afternoon you'll know this data well enough to focus on the feature instead of the dataset.

## Lab spec (10 min): the environment check

The most important lab of the day, because every other lab depends on it.

- **Goal:** prove your machine can reach a local model, an embedding model, and the venue's Azure OpenAI endpoint.
- **How:** run the three requests in `lab/smoke-test.http` (or the curl equivalents in `lab/README.md`), in order:
  1. Chat completion against Ollama (`llama3.2`), which proves Ollama is installed, running, and the model is pulled.
  2. Embedding against Ollama (`nomic-embed-text`), which proves the embedding model that powers Modules 2 and 3 is ready.
  3. Chat completion against Azure OpenAI using the key handed out at the door, which proves the cloud path for sentiment, RAG, and the capstone.
- **Success check:** three JSON responses, no red text. Compare `lab/expected-output.md`.
- **If something fails:** flag a helper now. Fallbacks (USB model copies, shared endpoints) exist precisely for this moment. Do not wait until Module 1 to mention it.
- **Stretch goal:** finished early? Pull up `data/trip-reports/` and skim one trip report end to end. Feel the problem feature 01 is about to solve.

## Leadership beat

- **When to reach for this:** before any AI feature at all. The framing question ("what are our users bad at, or sick of doing?") is the cheapest AI work your team will ever do, and it happens in a meeting room, not a codebase.
- **Rough cost & effort:** one workshop's worth of attention.
- **The one-liner for your CTO:** "Before we pick an AI feature, let's list the ten things our users hate doing. The features will pick themselves."

This is row 0 of the [decision framework](../../../docs/decision-framework.md), and the lens for everything that follows.
