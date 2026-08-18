# 03 · Sentiment (Small Local Vs. General Cloud)

Module 1: Understanding · Runs on Ollama (`phi3`) and `gpt-4.1` on Microsoft Foundry. The comparison is the feature.

## The User Problem

Trailhead Guides sells gear, and the Cascade 65 backpack has 300 reviews. The product team asks a simple question: are people happy with it, and what are they mad about? Star ratings lie: "4 stars, but the hip belt broke on day two" is not a happy customer. Someone would have to read all 300, and every new product adds to the pile. The team doesn't need eloquent analysis, just a reliable happy/unhappy/mixed signal at scale.

The user in this feature is the product team, not the hiker, and that's deliberate. Stakeholders are users too. Their version of the problem is that nobody has time to read every review, so a defect surfaces only when returns spike. Track the same happy/unhappy signal weekly and it surfaces months earlier, as a chart sliding downhill. One product in this corpus has exactly that kind of problem buried in its reviews, and the demo gets to find it.

## The Concept

Sentiment analysis is classification applied to text, and it's this workshop's vehicle for the most useful model-selection lesson of the day: you don't always need the big model. A small local model (`phi3`, about 2GB, free, private) labels straightforward reviews just as well as a frontier cloud model. At 300 reviews per product across a whole catalog, per-token pricing versus free-on-your-hardware is a real budget line.

The comparison cuts both ways, though. Feed both models the corpus's hard cases (sarcasm like "Great bag, if you enjoy shoulder pain", mixed feelings, ratings that contradict the text) and accuracy drops for everyone. How far it drops for each model is the thing you measure, and there is no verdict here for either side. The decision is measurable: run both on a labeled sample, count the disagreements, look at what the errors cost you, then choose. Most teams never run that experiment; you'll run it before lunch.

## The Lab

The hands-on lab is [F03-lab.md](F03-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** any high-volume text stream that needs a judgment call. Reviews, NPS verbatims, support tickets, social mentions, survey answers.
- **Rough cost & effort:** days. The classifier is trivial; the diligence is a labeled sample and an error count. Small local models often make the unit cost roughly zero.
- **The one-liner for your CTO:** "We measured: the free local model matches the expensive one on most of this task, and we know exactly which slice needs the big gun."

This card is row 3 of the [decision framework](../../../docs/decision-framework.md).
