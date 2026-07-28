# Module 0: Opening

**Setup and framing.** 9:00 to 9:30.

| | Feature | What it does |
|---|---|---|
| **Core** | [00 Setup & Framing](F00-setup-and-framing/) | Environment check, the day's thesis, a tour of the data |

Not a feature you build, but the half hour that decides whether the other four modules go well. Everyone runs the same three-request smoke test, broken setups surface while there is still slack to fix them, and the room starts Module 1 together.

## What happens here

The environment check comes first, at about 9:10, because the most expensive twenty minutes in a hands-on workshop is the twenty minutes in the first lab when a third of the room discovers their setup does not work. Three requests: local chat, local embeddings, and the cloud endpoint from the card handed out at the door.

Then the framing. Every company is currently asking "where can we add AI?" This workshop spends the day practicing the better question: what problems can AI solve for my users? And "users" is defined broadly on purpose, since three of the ten features exist for the people running the product rather than the people using it.

Then a five-minute tour of the Trailhead Guides corpus in [`data/`](../../data/), so no later feature has to stop and explain the dataset.

## How the day is shaped

Four working modules after this one, each 90 minutes: **Understanding**, **Finding**, **Deciding**, and **Doing**. Each module presents three features with live demos, then gives you roughly 45 minutes of hands-on time.

Every module marks one feature as **Core** and the rest as **Challenge**. Everyone does the Core lab. What you do after that is yours: another feature, a stretch goal, or helping the person next to you. Doing one lab properly is a good outcome, and it is the intended one for most people. Ten labs in a day is not the goal, and nobody is behind.

Every feature carries the same closing card: when to reach for it, roughly what it costs, and the one-liner for your CTO. Those accumulate into the [decision framework](../../docs/decision-framework.md) we assemble at 16:30, which is the artifact you actually take back to work.

## Before you sit down

Everything runs on your laptop. Three models, about 5 GB total, no GPU required, and an 8 GB machine is enough. The pre-work is in [`SETUP.md`](../../SETUP.md). If you did none of it, say so during this half hour rather than at 9:35.
