# Module 4: Doing

**The capstone.** 15:30 to 16:30.

One feature, double the time, and it uses everything from the rest of the day. A user types a plain-language request and the app carries out a sequence of real actions to fulfill it.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Core** | [10 Agentic Workflows](F10-agentic-workflows/F10-spec.md) | "Plan me a 3-day trip in Glacier" becomes real tool calls | Azure OpenAI (Ollama fallback) |

Everyone attempts this one. There is no menu here.

## How the hands-on works

I present the feature and run the agent live, then you get about 35 minutes. The lab ships the five tool definitions and a transcript of a complete successful run, so you can work against a known-good reference when your own run goes sideways. It will go sideways; see below.

If you finish, the stretch goals are in the feature README, and they are the good kind: add a tool, break the agent, and watch what it does when a tool it expected is missing.

## Why this one is different

An agent is a language model given tools and a goal, running in a loop. The model chooses which tool to call, your code executes it, the result goes back, and it continues until the goal is met. Your code never stops being in charge of what actually happens. The model only ever chooses.

That is the entire idea, and the reason it goes last is that it fails in every way the first nine features fail, at once, with the failures compounding.

## The thread to watch

This is the feature where reliability stops being a nice-to-have and becomes the whole engineering problem.

Across roughly two dozen local runs while building this, the model called a tool with the literal argument `[insert trail IDs here]`, announced it had called every necessary tool and then wrote nothing, emitted tool calls as plain text so none of them executed, and once read the reports saying a bridge was washed out and scheduled the closed trail anyway. Before any scaffolding was added, fewer than one run in five reached all four planning tools.

None of that is in the feature README as a warning. It is in there as measured fact, with counts, because it is the honest answer to "should we ship an agent?" The answer is yes, with a step budget, validated arguments, tools that return errors instead of throwing, a persisted trace, and a human approving anything irreversible. Which is to say: with feature 09.

This is also the one feature where the workshop pays for a frontier model rather than running local, and the reason is the numbers above rather than a preference.

## The leadership beat

Row 10 of the [decision framework](../../docs/decision-framework.md), and the one your board has already asked about.
