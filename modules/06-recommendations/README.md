# 06 · Recommendations

Block 2 (Finding) · Runs on Ollama embeddings (`nomic-embed-text`), reusing module 04's infrastructure

## The user problem

A hiker just finished Avalanche Lake Trail and loved it. Trailhead Guides says nothing. The obvious next screen ("you'd probably like these three trails") never got built, because everyone assumes recommendations require a data-science team, a ratings matrix, and six months. Meanwhile the gear store has the same gap: someone who bought the Cascade 65 gets shown a random carousel instead of the products that actually go with it.

## The concept

The classical answer to recommendations is collaborative filtering over user-behavior data, and it's real work with a real cold-start problem: it can't say anything about a new trail nobody has rated. This module shows the shortcut that gets you most of the value: content-based recommendations from the embeddings you already have. If module 04 gave every trail a position in meaning-space, then "trails similar to the one you just loved" is nothing more than nearest neighbors of that trail's vector. You don't train anything or build a ratings matrix, new items have no cold-start problem, and the infrastructure already exists.

That's the deliberate narrative beat of this module: one embedding investment keeps paying. Search (04), recommendations (06), and anomaly detection (08) are three features from one piece of infrastructure. The honest caveat gets said out loud too. Content similarity recommends things that are alike, and it will never discover that people who hike waterfalls also buy headlamps. When behavior data accumulates, collaborative filtering complements this; it doesn't replace it on day one.

## Demo outline (13 min, .NET)

1. Open the Avalanche Lake Trail page and ask the room what should be at the bottom of this screen. Everyone knows; nobody builds it.
2. Bring back the embedded trail catalog from module 04, already sitting in memory. Nothing new is created in this step, which is the point.
3. Write "more like this": take one trail's vector, rank every other trail by cosine similarity, skip itself, and take five. It's the search code with the query swapped for an item.
4. Run it for a few trails and let the room judge the results by reading the names, because they can. The alpine lake hikes do cluster. They also come back mostly rated hard when the trail you started from is a moderate family walk, and one neighbor is a Smokies hike with no lake at all. Difficulty and park are right there in the catalog, but they are not in the description text, so the embedding cannot see them. That is the whole lesson in one screen: this gets you candidates, and metadata re-ranking gets you recommendations.
5. Do the same with gear, and let it fail on purpose. The top result for the Cascade 65 Backpack is the Cascade 40 Daypack, the single product that buyer will never need. Content similarity finds substitutes; a store wants complements. Then show the review co-mention counts, where the Summit Bear Canister sits at the top, and name the difference: what a thing is like versus what people actually use with it.
6. Close on what this can't do. Cross-category discovery is out of reach for description similarity, one target trail in the slice has no real neighbors at all and a shipping product should show nothing rather than five weak guesses, and behavior data is what fixes both.

## Lab spec (13 min, any language)

- **Goal:** build "more like this" for trails from item embeddings.
- **Input:** `lab/` provides the same trail slice as module 04 and three target trails. If you finished module 04's lab, reuse your vectors; if not, `lab/` includes precomputed embeddings so you can start here.
- **How:** Ollama embeddings via `lab/ollama.http` (only if computing fresh), then your own cosine-similarity ranking from module 04.
- **Steps:**
  1. For target trail #1, rank all other trails by similarity and print the top 5.
  2. Repeat for the other two targets. Success check: your top 5 substantially overlaps the acceptable sets in `lab/expected-output.md` (there is more than one right answer, and the file says which neighbors are defensible).
  3. Read your own results and ask whether you'd ship them. That judgment call is the actual skill.
- **Stretch goal:** recommend for a user who liked two trails by averaging their vectors, or filter recommendations by metadata (only same park, only easier difficulty).

## Leadership beat

- **When to reach for this:** any catalog or content library where users finish one thing and get no next step. Products, articles, courses, templates, trails.
- **Rough cost & effort:** nearly free if you already built semantic search, since it's the same vectors queried differently. Days, even from scratch. Collaborative filtering can come later when behavior data exists.
- **The one-liner for your CTO:** "The same vectors that power our search give us 'more like this' for free."

This card is row 6 of the [decision framework](../../docs/decision-framework.md).
