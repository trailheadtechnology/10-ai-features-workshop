# Module 1: Understanding

**Making sense of messy content.** 9:30 to 11:00.

Three features that all take unstructured text and give you back something a person or a program can use. They share a shape: one model call, one carefully written instruction, no vector database and no training. If you only remember one thing from this module, it is that the instruction is the feature. The model is a commodity; what you ask it for is the product.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Core** | [01 Summarization](F01-summarization/F01-spec.md) | Long trip report to a 3-bullet conditions briefing | Ollama `llama3.2` |
| Challenge | [02 Extraction](F02-extraction/F02-spec.md) | Messy prose to validated structured JSON | Ollama `llama3.2` |
| Challenge | [03 Sentiment](F03-sentiment/F03-spec.md) | Gear reviews labeled, small model against big | Ollama `phi3` + cloud |

## How the hands-on works

I present all three with live demos, then you get about 45 minutes. **Everyone does the Core lab.** After that, take a Challenge if you have time, or take the Core lab's stretch goal if you would rather go deeper than wide. Finishing one lab well beats skimming three.

Start with 01 regardless of your experience. It is the shortest path from nothing to a working AI feature, and everything else today assumes you have made one model call and seen what comes back.

If you finish everything, the most valuable use of the remaining minutes is helping someone near you. Explaining a feature is how it sticks.

## What each lab costs you

- **01 Summarization** is the easiest lab of the day. One endpoint, one prompt, two provided reports. If you are new to this, budget the full time and do the stretch goal.
- **02 Extraction** adds a JSON schema and a validator. More code than 01, and the most directly reusable at work.
- **03 Sentiment** is the least code and the most measurement: you run 20 reviews through two models and score the disagreements. Choose this if the question you brought to the workshop is "which model should we pay for."

## The thread to watch

All three features fail in the same direction, and watching for it here will save you in the afternoon. Asked for something it cannot find in the source, a language model will supply something plausible rather than nothing. Feature 01 invents a trail closure from a bear sighting. Feature 02 reports a distance the report never gives. Feature 03 reads a sarcastic five-star review at face value.

None of that is fixed by a better model. It is fixed by a more specific instruction, a schema that permits `null`, and code that checks the output before anything downstream trusts it. That is the actual job.

## The leadership beats

We collect all three at the debrief. Each feature's README ends with the same card: when to reach for it, roughly what it costs, and the one-liner for your CTO. Those become rows 1 through 3 of the [decision framework](../../docs/decision-framework.md).
