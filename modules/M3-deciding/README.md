# Module 3: Deciding

**Triage and judgment.** 13:45 to 15:15.

The first two modules produced answers for a user. This one produces decisions for the people running the product: which queue does this go in, which of these 500 reports deserves attention, and what is the software allowed to send without a human reading it first.

Worth saying out loud, because it reframes the whole day: the user of these three features is usually a colleague. A ranger, an ops lead, a support manager. Stakeholders are users too, and "nobody has time to read all of this" is a real user problem.

| | Feature | What it does | Runs on |
|---|---|---|---|
| **Core** | [07 Classification & Routing](F07-classification-routing/) | Sorts an inbox, never misses an emergency | Ollama `llama3.2` |
| Challenge | [08 Anomaly Detection](F08-anomaly-detection/) | Finds the washed-out bridge in 500 routine reports | Ollama embeddings + math |
| Challenge | [09 Human-in-the-Loop](F09-human-in-the-loop/) | AI drafts, a ranger approves, everything is logged | Whatever the draft needs |

## How the hands-on works

I present all three with live demos, then you get about 45 minutes. Everyone does 07. Then pick.

**A note on choosing by energy, not just interest.** It is mid-afternoon and you have been coding since morning. Feature 08 is the most code of the three. Feature 09 is the least: it is mostly a policy worksheet and a decision about what your software may do unsupervised. If you are running out of gas, take 09. It is also the feature most likely to matter when you get back to work, because it is the pattern that makes the other nine shippable.

## What each lab costs you

- **07 Classification** is structured output with an enum, plus an honest scoring pass. The number that matters is not overall accuracy; it is whether both emergencies were caught.
- **08 Anomaly Detection** is embeddings plus arithmetic, and it ships precomputed vectors so you can skip straight to the math if you want.
- **09 Human-in-the-Loop** is drafts, an approval loop, and an audit trail. Half of it is judgment rather than typing.

## The thread to watch

Errors are not symmetric, and this module is where that stops being a slogan.

Misrouting a complaint costs somebody a day. Misrouting the message about the overdue hiker is a headline. Feature 07 measures recall on the class that matters rather than overall accuracy, and adds an `unsure` route so an ambiguous message goes to a person instead of confidently into the wrong queue.

Feature 09 makes the sharpest version of the point, and it makes it by failing. Asked to draft a reply to a woman whose husband is four hours overdue, the model ignores its escalation instruction and writes her a warm, reassuring, useless note. Every time. Move the instruction to the top and it announces `ESCALATE` and then writes the note anyway. The lesson is not that the prompt needed more work. It is that a prompt is a request and a policy lane is a guarantee, so the finished demo refuses to send emergencies to the model at all.

## The leadership beats

Collected at the debrief, becoming rows 7 through 9 of the [decision framework](../../docs/decision-framework.md).
