# Module 0: Opening

**Setup and framing.** About 30 minutes.

| | Feature | What it does |
|---|---|---|
| **Everyone** | [00 Setup & Framing](F00-setup-and-framing/F00-lab.md) | Environment check, the day's thesis, a tour of the data |

Nothing gets built in this half hour, but it decides whether the other four modules go well. Everyone runs the same three-request smoke test, broken setups surface while there is still slack to fix them, and the room starts Module 1 together.

## What Happens Here

The environment check comes first, about ten minutes in, because the most expensive twenty minutes in a hands-on workshop is the twenty minutes in the first lab when a third of the room discovers their setup does not work. It is three requests, one each against local chat, local embeddings, and the cloud endpoint, using the key handed out in the room.

After that comes the framing. Every company is currently asking "where can we add AI?", and this workshop spends the day practicing the better question, which is what problems AI can solve for your users. "Users" is defined broadly on purpose, since three of the ten features exist for the people running the product rather than the people using it.

Then two minutes on Trailhead Guides, the fictional trip-planning app every feature works inside: rambling trip reports, opinionated gear reviews, a trail catalog, park regulations, a visitor inbox, and a stream of trail-condition reports. Every feature folder carries the data its lab reads, and its lab doc describes it, so you meet each dataset when its feature does rather than all at once here.

## Who This Is For

This workshop is for people who write software for a living, can make an HTTP request in their language, and have not yet shipped an LLM feature, or have shipped one and want the other nine. No ML background is assumed.

Each module leaves you able to do one thing, and each module's debrief comes back to it: turn a document into a purpose-shaped summary and a validated record with a local model (Understanding); build embedding search over your own catalog and ground answers in your own documents with verified citations (Finding); route an inbox so the expensive class is never missed and gate anything irreversible behind a human with an audit trail (Deciding); wire tools into a bounded agent loop and read its trace (Doing).

## How the Day Is Shaped

Three working modules after this one: **Understanding** (10:30), **Finding** (11:45, with lunch from 12:30 to 1:30 in the middle of its lab), and **Deciding** (2:00, with the coffee break from 3:00 to 3:30 in the middle of its lab). Each one opens with about 30 minutes in which I introduce the theme and demo all three features, then gives you 40 minutes to build, then a short debrief. The capstone module, **Doing**, starts at 3:45 and runs an hour: 15 of demo, 40 building, and a debrief. Closing is 4:45 to 5:00.

Every module marks one feature as **Recommended** and the other two as **Challenge**. Start with the Recommended lab unless you have a reason not to. What you do after that is yours: another feature, a stretch goal, or helping the person next to you. Doing one lab properly is a good outcome, and it is the intended one for most people. Ten labs in a day is not the goal, and nobody is behind.

Every feature carries the same closing card: when to reach for it, the one-liner for your CTO, and a difficulty rating. Those accumulate into the [decision framework](../../docs/decision-framework.md) we assemble in the closing session, which is the artifact you actually take back to work.

## Before You Sit Down

Everything runs on your laptop: three models, about 5 GB in total, no GPU required, and an 8 GB machine is enough. The pre-work is in [`SETUP.md`](../../SETUP.md), and if you did none of it, say so during this half hour rather than in the first lab.
