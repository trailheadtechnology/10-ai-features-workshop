# Policy worksheet

Feature 07 sorted the inbox into six categories. Your job here is to decide what the software is allowed to do with each one. Pick a lane and defend it in one sentence. The justification is the graded part; reasonable people land on different lanes.

The lanes:

- **auto-send**: the model's reply goes to the visitor with no human in the path.
- **draft-for-approval**: the model writes it, a ranger approves, edits, or rejects it, and the decision is logged.
- **human-only**: no draft is generated at all. A person writes the reply.

Ask two questions per category. What does the worst plausible wrong reply cost? And can we take it back after we send it?

| Category | Lane | Worst plausible error | Reversible? | Justification |
|---|---|---|---|---|
| permit | | | | |
| conditions | | | | |
| complaint | | | | |
| lost-and-found | | | | |
| general | | | | |
| emergency | | | | |

Two follow-up questions once the table is full:

1. For every category you did not put in auto-send, what evidence would change your mind? Name the number, not the feeling. "Ninety days at under 5 percent edit rate with zero factual corrections" is an answer; "when it seems reliable" is not.
2. Where does your lane live in the code? If the answer is "the prompt tells it to," reread the emergency section of `expected-output.md`.

Reference answers and reasoning are in `expected-output.md`.
