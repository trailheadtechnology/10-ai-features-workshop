# Policy Worksheet

Feature 07 sorted the inbox into six categories. Your job here is to decide what the software is allowed to do with each one. Pick a lane and defend it in one sentence. The justification is the graded part; reasonable people land on different lanes.

The lanes:

- **auto-send**: the model's reply goes to the visitor with no human in the path.
- **draft-for-approval**: the model writes it, a ranger approves, edits, or rejects it, and the decision is logged.
- **human-only**: no draft is generated at all. A person writes the reply.

Ask two questions per category. What does the worst plausible wrong reply cost? And can we take it back after we send it?

| Category | Lane | Worst plausible error | Reversible? | Justification |
|---|---|---|---|---|
| permit | draft-for-approval | Tells someone no permit is needed and they lose a once-a-year trip at check-in. | Sometimes; not if they arrive before correction. | Permit mistakes can derail a trip and create enforcement problems, so a ranger must approve. |
| trail-condition | draft-for-approval | Says a closed or hazardous trail is open and a family heads into dangerous conditions. | Not reliably; harm can happen before correction is seen. | Safety and closure status can change quickly, so a human review is required before sending. |
| complaint | draft-for-approval | Promises a refund, transfer, or policy exception that the park will not honor. | Partly; screenshots of promises persist and trust damage remains. | The model can draft tone, but commitments need a ranger to avoid false promises. |
| lost-and-found | draft-for-approval | Confidently says an item was or was not found when that status is wrong. | Usually yes; a follow-up can correct it. | Lower risk than safety topics, but still needs review until logs show near-zero edits. |
| general | draft-for-approval | Gives a confident, cited-sounding answer that is wrong on a policy detail. | Sometimes; visitors may act before seeing a correction. | General questions still carry policy risk, so keep a reviewer in the loop for now. |
| emergency | human-only | Sends a reassuring draft instead of initiating dispatch for an overdue/injured person. | No; life-safety delays are irreversible. | Emergency traffic must bypass generation entirely and route straight to a human dispatcher. |

Two follow-up questions once the table is full:

1. I would require at least 90 days, at least 200 reviewed messages per category, median edit distance under 5% of draft length, and zero factual-correction edits before promoting any category to auto-send.
2. The lane must live in code as a category-to-lane policy map checked before the model call, with `human-only` short-circuiting to no draft.

Reference answers and reasoning are in `expected-output.md`.
