# What passing looks like

Every number on this page came from a real run of both sets through `phi3` and `llama3.2` via the Ollama API, scored against `reference-labels.json`. Azure OpenAI was not available when this was written, so `llama3.2` stands in for the big model. That substitution matters, and the results say why.

## Accuracy, measured

| Model | Easy set | Hard set |
|---|---|---|
| `phi3` (2GB, local, free) | 9/10 | 7/10 |
| `llama3.2` (big-model stand-in) | 9/10 | 7/10 |

A dead tie, on both sets. If you expected the small model to fall apart on sarcasm, so did we. It didn't, at least not against this stand-in. That is the honest result and it is worth more than the tidy one: the easy-set half of the argument lands hard (you would have paid for nothing), and the hard-set half turns into a real question rather than a foregone conclusion. When you run `azure.http` with the key from the card, you are testing whether a frontier model actually opens the gap that `llama3.2` does not.

## The disagreement list (4 of 20)

| id | set | reference | phi3 | llama3.2 | who's right |
|---|---|---|---|---|---|
| gr-0074 | easy | positive | mixed | negative | neither |
| gr-0004 | hard | positive | negative | mixed | neither |
| gr-0013 | hard | negative | mixed | negative | llama3.2 |
| gr-0089 | hard | positive | positive | mixed | phi3 |

One each. The two models split the disagreements they actually resolved, and lost the other two together.

- `gr-0074` is the quiet lesson of the whole lab. "It pitches in three minutes flat. That's it, that's the review." is a four-star rave, and both models read the deadpan brevity as a complaint. Nothing about it is sarcastic or contradictory; it just isn't effusive. Your hard cases are not always the ones you nominated.
- `gr-0013` and `gr-0089` are the classic shape: the star rating is about something other than the product (a return process, a customer-service email). Each model gets one and misses the other.

Also worth knowing: both models missed `gr-0021` the same way, calling the camp chair review `negative` when the chair is described as "a sitting cloud" and the star rating is about the reviewer's ex. Because they agree, it never shows up as a disagreement. Agreement is not correctness, which is exactly why you hand-label the sample before you run anything.

## The finding nobody was looking for

The prompt in `ollama.http` is wrapped across four short lines. Reflowing that identical prompt onto a single line, changing nothing but the newlines, moved `phi3` from 9/10 to 7/10 on the easy set and from 7/10 to 4/10 on the hard set. `llama3.2` scored the same either way (9/10 and 7/10).

Both variants were measured, twice, at temperature 0. Under the single-line prompt phi3's failure mode was a consistent retreat to `mixed` on anything containing a negative word, which is the worst possible error for a product team: it launders anger into ambivalence and flattens the trend line the feature exists to surface.

The takeaway is not that line breaks are magic. It is that the small model is more sensitive to prompt formatting than the bigger one, so a comparison that changes both the model and the prompt shape is measuring nothing. Fix the prompt bytes, then vary the model.

## Success check

You pass when you can show your own version of the two tables above: accuracy split easy vs. hard for both models, plus the reviews where they disagreed and your call on which model was right. Getting different numbers than these is fine, especially against a real Azure deployment. Not being able to produce the tables is not.

## Stretch goal

Ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` instead of one word. Watch two things: whether each model returns parseable JSON every time, and whether it invents aspects the review never mentions rather than leaving them null. That second failure is where the models separate more cleanly than they did on the single label.
