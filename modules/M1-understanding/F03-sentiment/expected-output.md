# What Passing Looks Like

Every number on this page came from real runs scored against `reference-labels.json`: `phi3` and `llama3.2` through the Ollama API, and `gpt-4.1` on the workshop's Microsoft Foundry deployment (the same one the room key points at). The local-only version of this comparison, with `llama3.2` standing in for the big model, is kept below because its result is instructive on its own.

## Accuracy, Measured

| Model | Easy set | Hard set |
|---|---|---|
| `phi3` (2GB, local, free) | 9/10 | 7/10 |
| `gpt-4.1` (Microsoft Foundry) | 10/10 | 10/10 |
| `llama3.2` (local stand-in, offline fallback) | 9/10 | 7/10 |

The easy set is close to a tie: the free local model gets nine of ten and the frontier model gets the tenth, which is `gr-0074`, the deadpan four-star rave. On the hard set the frontier model earns its price: 10/10 against 7/10, and every one of `phi3`'s misses is one `gpt-4.1` gets right. So both halves of the module's argument hold at once. Straightforward reviews do not need the big model, and the sarcastic, contradictory slice does, and you know which slice is which because you measured it.

The local stand-in tells a different story, and it is worth keeping: `llama3.2` ties `phi3` on both sets. A bigger local model is not the same thing as a frontier model, and a comparison run against the wrong stand-in would have told you the gap does not exist.

## The Disagreement List

Against `gpt-4.1` (4 of 20, all four resolved the same way):

| id | set | reference | phi3 | gpt-4.1 | who's right |
|---|---|---|---|---|---|
| gr-0074 | easy | positive | mixed | positive | gpt-4.1 |
| gr-0004 | hard | positive | negative | positive | gpt-4.1 |
| gr-0013 | hard | negative | mixed | negative | gpt-4.1 |
| gr-0021 | hard | positive | negative | positive | gpt-4.1 |

Against the `llama3.2` stand-in (4 of 20, split):

| id | set | reference | phi3 | llama3.2 | who's right |
|---|---|---|---|---|---|
| gr-0074 | easy | positive | mixed | negative | neither |
| gr-0004 | hard | positive | negative | mixed | neither |
| gr-0013 | hard | negative | mixed | negative | llama3.2 |
| gr-0089 | hard | positive | positive | mixed | phi3 |

One each in the local pairing. The two local models split the disagreements they actually resolved, and lost the other two together.

- `gr-0074` is the miss worth showing the room. "It pitches in three minutes flat. That's it, that's the review." is a four-star rave, and both models read the deadpan brevity as a complaint. Nothing about it is sarcastic or contradictory; it just isn't effusive. Your hard cases are not always the ones you nominated.
- `gr-0013` and `gr-0089` are the classic shape: the star rating is about something other than the product (a return process, a customer-service email). Each model gets one and misses the other.

Also worth knowing: both local models missed `gr-0021` the same way, calling the camp chair review `negative` when the chair is described as "a sitting cloud" and the star rating is about the reviewer's ex. Because they agree, it never shows up as a disagreement in the local pairing; it only surfaces once a model that reads it correctly (`gpt-4.1` does) is in the comparison. Agreement is not correctness, which is why you hand-label the sample before you run anything.

## The Finding Nobody Was Looking for

The prompt in `http/ollama.http` is wrapped across four short lines. Reflowing that identical prompt onto a single line, changing nothing but the newlines, moved `phi3` from 9/10 to 7/10 on the easy set and from 7/10 to 4/10 on the hard set. `llama3.2` scored the same either way (9/10 and 7/10).

Both variants were measured, twice, at temperature 0. Under the single-line prompt phi3's failure mode was a consistent retreat to `mixed` on anything containing a negative word, which is the worst possible error for a product team: it launders anger into ambivalence and flattens the trend line the feature exists to surface.

Line breaks are not magic. The lesson is that the small model is more sensitive to prompt formatting than the bigger one, so a comparison that changes both the model and the prompt shape is measuring nothing. Fix the prompt bytes, then vary the model.

## Success Check

You pass when you can show your own version of the two tables above: accuracy split easy vs. hard for both models, plus the reviews where they disagreed and your call on which model was right. Getting different numbers than these is fine; a different deployment or a different day will move a review or two. Not being able to produce the tables is not.

## Stretch Goal

Ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` instead of one word. Watch two things: whether each model returns parseable JSON every time, and whether it invents aspects the review never mentions rather than leaving them null. That second failure is where the models separate more cleanly than they did on the single label.
