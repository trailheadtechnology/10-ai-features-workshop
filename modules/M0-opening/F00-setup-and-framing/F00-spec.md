# 00 · Setup & Framing

Module 0: Opening (about 30 minutes) · Runs on your laptop, before anything else does

## Why This Feature Exists

The most expensive twenty minutes of any hands-on workshop is the twenty minutes in Module 1 when a third of the room discovers their environment doesn't work. This feature spends its time up front instead: everyone runs one smoke test, broken setups surface while there's still slack to fix them, and the room starts Module 1 together.

It's also where the day gets its thesis. Every company right now is asking "where can we add AI?" This workshop spends seven hours practicing the better question: "what problems can AI best solve for my users?" Each of the ten features that follow opens with a user who is stuck, and the AI only shows up as the answer to that user's problem.

"Users" is defined broadly on purpose. The product manager who can't read every review counts, and so does the ranger staring down a full inbox. Three of the ten features (03, 07, and 08) solve problems for the people running the product rather than the people using it. Spotting problems early is a user problem too; it just belongs to a user on your payroll.

## How the Day Works (5 Min)

- Who it is for: people who write software for a living, can make an HTTP request in their language, and have not yet shipped an LLM feature (or have shipped one and want the other nine). No ML background assumed.
- What they leave able to do, one outcome per module: turn a document into a purpose-shaped summary and a validated record with a local model; build embedding search over their own catalog and ground answers in their own documents with verified citations; route an inbox so the expensive class is never missed and gate anything irreversible behind a human with an audit trail; wire tools into a bounded agent loop and read its trace. Say these out loud; each module's debrief comes back to its one.
- Five modules: **Opening**, **Understanding**, **Finding**, **Deciding**, **Doing**. The three middle modules are 90 minutes each and cover three features; Doing is a 60-minute agent capstone.
- Every module has the same beat: 30 minutes in which I introduce the theme and demo all three features in .NET (10 minutes for the capstone's single feature), then 60 minutes of hands-on time (50 for the capstone) in which you start with the Recommended lab and anyone with time left picks a Challenge lab; the capstone is the one lab everyone does. Every lab can be done in any language.
- Every leadership beat is a row in the [decision framework](../../../docs/decision-framework.md) we assemble at the end of the day. That document is the thing you take back to your CTO.
- Every model today is an HTTP endpoint. Ollama is a local server on `localhost:11434` (native `/api/chat` and `/api/embed`, plus an OpenAI-compatible `/v1`); Foundry serves `gpt-4.1` and `gpt-5.5` at `…openai.azure.com/openai/deployments/<name>/chat/completions` with an `api-key` header. Same request shape both places. Labs ship as raw `.http` files against both, so if your language can make an HTTP request, you're equipped, and the SDKs are wrappers over the same calls. Every feature also ships the demo code three ways, .NET, Python, and TypeScript, each with a starter and a complete version.

## Meet Trailhead Guides (2 Min)

All ten features work through **Trailhead Guides**, a fictional national-park trip-planning app with a deliberately messy corpus: rambling trip reports, opinionated gear reviews, hundreds of trail descriptions, dry park regulations, and a queue of visitor inquiries. Each feature stands alone, but they all live in the same park, so by mid-afternoon you'll know this data well enough to focus on the feature instead of the dataset.

## The Lab (10 Min)

The environment check is [F00-lab.md](F00-lab.md): three requests, three JSON responses, and anyone whose machine cannot make them raises a hand now rather than in the first lab. It is the most important lab of the day, because every other lab depends on it.

## Leadership Beat

- **When to reach for this:** before any AI feature at all. The framing question ("what are our users bad at, or sick of doing?") is the cheapest AI work your team will ever do, and it happens in a meeting room, not a codebase.
- **Rough cost & effort:** one workshop's worth of attention.
- **The one-liner for your CTO:** "Before we pick an AI feature, let's list the ten things our users hate doing. The features will pick themselves."

This is row 0 of the [decision framework](../../../docs/decision-framework.md), and the question the other nine features are answers to.
