# Module 3: Deciding

**Triage and judgment.** About 90 minutes.

The first two modules produced answers for a user. This one produces decisions for the people running the product: which queue does this go in, which of these 500 reports deserves attention, and what is the software allowed to send without a human reading it first.

The user of these three features is usually a colleague, a ranger or an ops lead or a support manager, rather than a hiker. Stakeholders are users too, and "nobody has time to read all of this" is a real user problem.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Recommended** | [07 Classification & Routing](F07-classification-routing/F07-spec.md) | Sorts an inbox, never misses an emergency | Ollama `llama3.2` |
| Challenge | [08 Anomaly Detection](F08-anomaly-detection/F08-spec.md) | Finds the washed-out bridge in 500 routine reports | Ollama embeddings + math |
| Challenge | [09 Human-in-the-Loop](F09-human-in-the-loop/F09-spec.md) | AI drafts, a ranger approves, everything is logged | Whatever the draft needs |

## How the Hands-On Works

The first 30 minutes are mine: the theme, and a demo of all three features. The remaining 60 are yours to build. 07 is the Recommended start, and after that you pick.

**Choose by energy as much as by interest.** It is mid-afternoon and you have been coding since morning. Feature 08 is the most code of the three, and feature 09 is the least, since it is mostly a policy worksheet and a decision about what your software may do unsupervised. If you are running out of gas, take 09; it is also the feature most likely to matter when you get back to work, because it is the pattern that makes the other nine shippable.

## What Each Lab Costs You

- **07 Classification** is structured output with an enum, plus an honest scoring pass. The number that matters is not overall accuracy; it is whether both emergencies were caught.
- **08 Anomaly Detection** is embeddings plus arithmetic, and it ships precomputed vectors so you can skip straight to the math if you want.
- **09 Human-in-the-Loop** is drafts, an approval loop, and an audit trail, and half of it is judgment rather than typing.

## The Thread to Watch

Errors here are not symmetric: misrouting a complaint costs somebody a day, and misrouting the message about the overdue hiker is a headline. Feature 07 measures recall on the class that matters rather than overall accuracy, and adds an `unsure` route so an ambiguous message goes to a person instead of confidently into the wrong queue.

Feature 09 makes the sharpest version of the point, and it makes it by failing. Asked to draft a reply to a woman whose husband is four hours overdue, the model ignores its escalation instruction and writes her a warm, reassuring, useless note, every single time. Move the instruction to the top of the prompt and it announces `ESCALATE` and then writes the note anyway. The lesson is not that the prompt needs more engineering. A prompt instruction is a request and a policy lane is a guarantee, so the finished demo refuses to send emergencies to the model at all.

## The Leadership Beats

Collected at the debrief, and they become rows 7 through 9 of the [decision framework](../../docs/decision-framework.md).
