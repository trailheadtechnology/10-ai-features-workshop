# 06 · Recommendations

Module 2: Finding · Runs on Ollama embeddings (`nomic-embed-text`), reusing feature 04's infrastructure

## The User Problem

A hiker just finished Avalanche Lake Trail and loved it. Trailhead Guides says nothing. The obvious next screen ("you'd probably like these three trails") never got built, because everyone assumes recommendations require a data-science team, a ratings matrix, and six months. Meanwhile the gear store has the same gap: someone who bought the Cascade 65 gets shown a random carousel instead of the products that actually go with it.

## The Concept

The classical answer to recommendations is collaborative filtering over user-behavior data, and it's real work with a real cold-start problem: it can't say anything about a new trail nobody has rated. This feature shows the shortcut that gets you most of the value: content-based recommendations from the embeddings you already have. If feature 04 gave every trail a position in meaning-space, then "trails similar to the one you just loved" is nothing more than nearest neighbors of that trail's vector. You don't train anything or build a ratings matrix, new items have no cold-start problem, and the infrastructure already exists.

That's the deliberate narrative beat of this feature: one embedding investment keeps paying. Search (04), recommendations (06), and anomaly detection (08) are three features from one piece of infrastructure. The honest caveat gets said out loud too. Content similarity recommends things that are alike, and it will never discover that people who hike waterfalls also buy headlamps. When behavior data accumulates, collaborative filtering complements this; it doesn't replace it on day one.

## The Lab

The hands-on lab is [F06-lab.md](F06-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** any catalog or content library where users finish one thing and get no next step. Products, articles, courses, templates, trails.
- **Rough cost & effort:** nearly free if you already built semantic search, since it's the same vectors queried differently. Days, even from scratch. Collaborative filtering can come later when behavior data exists.
- **The one-liner for your CTO:** "The same vectors that power our search give us 'more like this' for free."

This card is row 6 of the [decision framework](../../../docs/decision-framework.md).
