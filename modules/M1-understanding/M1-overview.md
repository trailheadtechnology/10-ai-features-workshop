# Module 1: Understanding

**Making sense of messy content.** About 90 minutes.

The three features in this module all take unstructured text and give you back something a person or a program can use, and they share a shape: one model call and one carefully written instruction, with no vector database and no training anywhere in sight. If you only remember one thing from the module, remember that the wording of the instruction does more work than the choice of model.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Recommended** | [01 Summarization](F01-summarization/F01-spec.md) | Long trip report to a 3-bullet conditions briefing | Ollama `llama3.2` |
| Challenge | [02 Extraction](F02-extraction/F02-spec.md) | Messy prose to validated structured JSON | Ollama `llama3.2` |
| Challenge | [03 Sentiment](F03-sentiment/F03-spec.md) | Gear reviews labeled, small model against big | Ollama `phi3` + cloud |

## How the Hands-On Works

The first 30 minutes are mine: the theme, and a demo of all three features. The remaining 60 are yours to build. **Start with the Recommended lab.** After that, take a Challenge lab if you have time, or take the Recommended lab's stretch goal if you would rather go deeper than wide; finishing one lab well beats skimming three.

Start with 01 regardless of your experience. It is the shortest path from nothing to a working AI feature, and everything else today assumes you have made one model call and seen what comes back.

If you finish everything, the best use of the remaining minutes is helping someone near you, since explaining a feature is how it sticks.

## What Each Lab Costs You

- **01 Summarization** is the easiest lab of the day: one endpoint, one prompt, and two provided reports. If you are new to this, budget the full time and do the stretch goal.
- **02 Extraction** adds a JSON schema and a validator, so it is more code than 01 and the most directly reusable at work.
- **03 Sentiment** is the least code and the most measurement: you run 20 reviews through two models and score the disagreements. Choose this if the question you brought to the workshop is "which model should we pay for."

## The Thread to Watch

All three features fail in the same direction, and watching for it here will save you in the afternoon. Asked for something it cannot find in the source, a language model will supply something plausible rather than nothing: feature 01 invents a trail closure from a bear sighting, feature 02 reports a distance the report never gives, and feature 03 reads a sarcastic five-star review at face value.

A better model does not fix any of that. What does is a more specific instruction, a schema that permits `null`, and code that checks the output before anything downstream trusts it.

## The Leadership Beats

We collect all three at the debrief. Each feature's README ends with the same card: when to reach for it, roughly what it costs, and the one-liner for your CTO. Those become rows 1 through 3 of the [decision framework](../../docs/decision-framework.md).
