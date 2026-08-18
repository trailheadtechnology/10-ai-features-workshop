# 09 · Human-in-the-Loop

Module 3: Deciding · Pattern feature; uses whatever the demo needs (Ollama for drafting)

## The User Problem

Feature 07 routed the inbox, so now a ranger stares at forty messages that all need replies. Most answers are boilerplate the ranger has typed a hundred times, and typing them eats the afternoon. The obvious move is to let the AI answer, and the obvious disaster is the AI telling a visitor that campfires are fine during a burn ban, on official park letterhead. The ranger's problem is drudgery; the park's problem is that full automation of outbound communication is how you end up apologizing publicly.

## The Concept

Human-in-the-loop is a product pattern, not a model feature, and it's the difference between AI features that ship and AI features that get killed in legal review. The core move: the AI drafts, the human approves, edits, or rejects, and the system remembers what happened. The user-facing risk drops to near zero while most of the typing still disappears.

The design question is where to put the human, and the answer comes from error cost and reversibility, which connects straight back to feature 07's asymmetry lesson. A sensible policy has three lanes: full automation for cheap, reversible, low-stakes replies; draft-plus-approval for the middle; human-only for the expensive and irreversible (in Trailhead Guides terms, emergencies never get an AI draft at all). Two practical details do a lot of work in real systems. Keep an audit trail of what was drafted, who approved it, and what they changed. And measure the gap between draft and final text, because how much humans edit tells you whether trust in each lane is earned, and edit patterns show you exactly where the drafts fall short.

## The Lab

The hands-on lab is [F09-lab.md](F09-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** any customer-facing or high-stakes generation. Support replies, quotes, claims decisions, medical or legal drafts, anything sent under your organization's name.
- **Rough cost & effort:** the drafting is trivial; the approval UI, audit trail, and policy design are the real work, and that's ordinary product engineering measured in weeks. This pattern is often what makes the other nine features shippable.
- **The one-liner for your CTO:** "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can."

This card is row 9 of the [decision framework](../../../docs/decision-framework.md).
