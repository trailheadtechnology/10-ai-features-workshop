# Module 4: Doing

**The capstone.** About 60 minutes.

The last module is a single feature with a double-length slot, and it draws on everything the day has built up to this point. The user types a plain-language request, and the app works out and carries out the sequence of real actions needed to fulfill it, checking the forecast, searching for trails, reading condition reports, and filing a permit request, instead of handing back a list of links.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Everyone** | [10 Agentic Workflows](F10-agentic-workflows/F10-spec.md) | "Plan me a 3-day trip in Glacier" becomes real tool calls | `gpt-5.5` on Microsoft Foundry (Ollama fallback) |

There is no Recommended-or-Challenge choice in this module; everyone attempts the capstone, and the lab is sized so that a partial result still teaches the lesson.

## How the Hands-On Works

The first 10 minutes are mine: the feature and the agent running live. The remaining 50 are yours to build. The lab ships the five tool definitions and a transcript of a complete successful run, so when your own run goes sideways, and it will, you have a known-good reference to compare against rather than a guess.

If you finish, the stretch goals are in the feature README, and they are the good kind: add a tool, break the agent, and watch what it does when a tool it expected is missing.

## Why This One Is Different

An agent is a language model given tools and a goal, running in a loop. The model chooses which tool to call, your code executes it, the result goes back, and it continues until the goal is met. Your code never stops being in charge of what actually happens; the model only chooses.

That is the entire idea. It goes last because an agent fails in every way the first nine features fail, all at once, and the failures compound: a hallucinated fact becomes a tool argument, a misclassified message becomes an action, and an ungrounded answer becomes a booking.

## The Thread to Watch

Reliability turns out to be the whole engineering problem. Across roughly two dozen local runs while building this, the model called a tool with the literal argument `[insert trail IDs here]`, announced it had called every necessary tool and then wrote nothing, emitted tool calls as plain text so none of them executed, and once read the reports saying a bridge was washed out and scheduled the closed trail anyway. Before any scaffolding was added, fewer than one run in five reached all four planning tools.

The feature README carries those counts as measured fact, because they are the honest answer to "should we ship an agent?" The answer is yes, provided you ship it with a step budget, validated arguments, tools that return errors instead of throwing, a persisted trace, and a human approving anything irreversible, which is another way of saying that feature 09 comes along for the ride.

It is also the one feature where the workshop pays for a frontier model rather than running local, and the numbers above are the reason for that choice, not a preference for the cloud.

## The Leadership Beat

Row 10 of the [decision framework](../../docs/decision-framework.md), and the one your board has already asked about.
