# 03 · Sentiment (small local vs. general cloud)

Module 1: Understanding · Runs on Ollama (`phi3`) and Azure OpenAI. The comparison is the feature.

## The user problem

Trailhead Guides sells gear, and the Cascade 65 backpack has 300 reviews. The product team asks a simple question: are people happy with it, and what are they mad about? Star ratings lie. "4 stars, but the hip belt broke on day two" is not a happy customer. Someone would have to read all 300, and every new product adds to the pile. The team doesn't need eloquent analysis, just a reliable happy/unhappy/mixed signal at scale.

The user in this feature is the product team, not the hiker, and that's deliberate. Stakeholders are users too. Their version of the problem is that nobody has time to read every review, so a defect surfaces only when returns spike. Track the same happy/unhappy signal weekly and it surfaces months earlier, as a chart sliding downhill. One product in this corpus has exactly that kind of problem buried in its reviews, and the demo gets to find it.

## The concept

Sentiment analysis is classification applied to text, and it's this workshop's vehicle for the most useful model-selection lesson of the day: you don't always need the big model. A small local model (`phi3`, about 2GB, free, private) labels straightforward reviews just as well as a frontier cloud model. At 300 reviews per product across a whole catalog, per-token pricing versus free-on-your-hardware is a real budget line.

The comparison cuts both ways, though. Feed both models the corpus's hard cases (sarcasm like "Great bag, if you enjoy shoulder pain", mixed feelings, ratings that contradict the text) and accuracy drops for everyone. How far it drops for each model is the thing you measure. The takeaway isn't a verdict for either side. It's that this is a measurable engineering decision: run both on a labeled sample, count the disagreements, look at what the errors cost you, then choose. Most teams never run that experiment. You'll run it before lunch.

## Demo outline (13 min, .NET)

1. Show the Cascade 65's reviews, and point at a 4-star review with furious text. Stars lie; text doesn't.
2. Starter project: one classify method whose prompt returns exactly `positive | negative | mixed`. Because of Microsoft.Extensions.AI, the same code runs against both providers, and the swap is one line in DI registration. Say that out loud, since it's the provider-flexibility slide made real.
3. Run 20 easy reviews through `phi3` locally: fast, free, and correct. Let the small model win the first round.
4. Run the same 20 through Azure OpenAI and get identical labels. First payoff: you'd have paid for nothing.
5. Now the hard set, sarcasm and mixed reviews. Run both and diff the labels on screen. Second payoff: both models drop several points, and the disagreements are where the interesting arguments live. Whether the big model actually earns its price on this slice is the measurement, not the assumption. Check `lab/expected-output.md` for what happened when this was built, and be ready for the room's answer to differ from yours.
6. Close with the decision recipe: labeled sample, run both, count disagreements, price the errors. That recipe generalizes to every feature today.

## Lab spec (13 min, any language)

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, and find where they disagree.
- **Input:** `lab/` provides about 20 reviews from `data/gear-reviews.jsonl`, split into an easy set and a hard set (sarcasm, contradictions), plus reference labels.
- **How:** `lab/ollama.http` (phi3) and `lab/azure.http` (Azure OpenAI, key handed out in the room). Same prompt, two endpoints.
- **Steps:**
  1. Classify the easy set with `phi3` and score against the reference labels.
  2. Classify the hard set with both models.
  3. Success check: produce the disagreement list. Which reviews got different labels, and which model was right? (See `lab/expected-output.md`.)
- **Stretch goal:** extend the label to aspect-based sentiment, `{overall, aspects: {comfort, durability, price}}`, and see which model can go deeper than a single label.

## Leadership beat

- **When to reach for this:** any high-volume text stream that needs a judgment call. Reviews, NPS verbatims, support tickets, social mentions, survey answers.
- **Rough cost & effort:** days. The classifier is trivial; the diligence is a labeled sample and an error count. Small local models often make the unit cost roughly zero.
- **The one-liner for your CTO:** "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun."

This card is row 3 of the [decision framework](../../../docs/decision-framework.md).
