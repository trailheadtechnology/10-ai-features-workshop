# 09 · Human-in-the-loop

Block 3 (Deciding) · Pattern module; uses whatever the demo needs (Ollama for drafting)

## The user problem

Module 07 routed the inbox, so now a ranger stares at forty messages that all need replies. Most answers are boilerplate the ranger has typed a hundred times, and typing them eats the afternoon. The obvious move is to let the AI answer. The obvious disaster is the AI telling a visitor that campfires are fine during a burn ban, on official park letterhead. The ranger's problem is drudgery; the park's problem is that full automation of outbound communication is how you end up apologizing publicly.

## The concept

Human-in-the-loop is a product pattern, not a model feature, and it's the difference between AI features that ship and AI features that get killed in legal review. The core move: the AI drafts, the human approves, edits, or rejects, and the system remembers what happened. The user-facing risk drops to near zero while most of the typing still disappears.

The design question is where to put the human, and the answer comes from error cost and reversibility, which connects straight back to module 07's asymmetry lesson. A sensible policy has three lanes: full automation for cheap, reversible, low-stakes replies; draft-plus-approval for the middle; human-only for the expensive and irreversible (in Trailhead Guides terms, emergencies never get an AI draft at all). Two practical details do a lot of work in real systems. Keep an audit trail of what was drafted, who approved it, and what they changed. And measure the gap between draft and final text, because how much humans edit tells you whether trust in each lane is earned, and edit patterns show you exactly where the drafts fall short.

## Demo outline (13 min, .NET)

1. Pick up where module 07 left off: a routed queue of condition questions awaiting replies.
2. Generate a draft reply for one inquiry, grounded in the park docs the same way module 05 grounded answers. Show the draft on screen; it's good but not perfect.
3. Run the approval flow in a small console UI: approve one draft untouched, edit a second before sending, reject a third and write it by hand. Three inquiries handled in the time one used to take.
4. Show the audit log the flow just produced: draft, decision, final text, edit distance per reply.
5. Put the routing policy on screen as a table: which categories auto-send, which get drafts, which stay human-only. Point at the emergency row: no AI draft, ever, by policy.
6. Close the loop: the edits collected in step 3 are tomorrow's prompt improvements. The human is the feedback signal as much as the safety check.

## Lab spec (13 min, any language)

- **Goal:** generate draft replies for routed inquiries, then decide the automation policy per category.
- **Input:** `lab/` provides 3 classified inquiries (a conditions question, a complaint, a permit request), relevant park-doc snippets for grounding, and a policy worksheet.
- **How:** POST to Ollama's chat endpoint (`llama3.2`) with the grounding snippets. `lab/ollama.http` has the drafting request.
- **Steps:**
  1. Generate a draft reply for each inquiry and read them as an editor: what would you change before this goes out under your name?
  2. Fill in the policy worksheet: for each of the five categories from module 07, choose auto-send, draft-for-approval, or human-only, and write one sentence of justification based on error cost.
  3. Success check: compare against `lab/expected-output.md`, which has reference drafts and a reasoned reference policy. Your policy may differ; your justifications are what count. This lab is deliberately part judgment, because that's the actual skill.
- **Stretch goal:** compute edit distance between a draft and your edited version, and sketch what threshold would earn a category promotion from draft-mode to auto-send.

## Leadership beat

- **When to reach for this:** any customer-facing or high-stakes generation. Support replies, quotes, claims decisions, medical or legal drafts, anything sent under your organization's name.
- **Rough cost & effort:** the drafting is trivial; the approval UI, audit trail, and policy design are the real work, and that's ordinary product engineering measured in weeks. This pattern is often what makes the other nine features shippable.
- **The one-liner for your CTO:** "The AI does the typing, our people keep the judgment, and nothing goes out without a human OK until the data says it can."

This card is row 9 of the [decision framework](../../docs/decision-framework.md).
