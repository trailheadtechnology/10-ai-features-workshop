# 08 · Anomaly detection

Block 3 (Deciding) · Runs on Ollama embeddings (`nomic-embed-text`) plus plain math

## The user problem

Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

## The concept

This module has a secret, and the secret is that it's barely an AI module: it's embeddings plus arithmetic. Embed every condition report for a trail, and the routine ones ("muddy," "buggy," "fine") form a dense cluster in vector space. Average them and you get a centroid, the mathematical center of "normal" for that trail. A new report's distance from that centroid is an anomaly score. "Bridge washed out" sits far from the mud cluster, so it jumps out of the pile with no large model, no classifier, and no training, just cosine distance and a threshold.

Beyond one weird report, clusters of anomalies are the interesting signal. One outlier might be a rambling hiker; three outliers in a week that also sit close to each other are an event. The pattern generalizes to any stream of routine text: support tickets, log messages, form submissions, review streams. Define normal from the data itself, and let distance flag what deserves human eyes. It also pairs naturally with module 07: classification handles the categories you knew to define, and anomaly detection catches the things you didn't.

## Demo outline (13 min, .NET)

1. Scroll the condition-report stream. Boring on purpose. Ask the room to find the problem; nobody can, and that's the situation the park is in.
2. Embed one trail's reports with the module 04 embedding code, average the vectors into a centroid, and put the idea on one slide: the center of normal.
3. Print every report's distance from the centroid, sorted. The payoff: "bridge washed out completely at the crossing" tops the list by a wide margin, and the mud reports pack the bottom.
4. Show the cluster signal: the three washout reports are all far from normal but close to each other. That's not noise, that's an event.
5. Turn it into an alert in a few lines: distance beyond threshold, more than N reports in a window, flag the trail. Show it also catching the planted bear-activity spike on the other trail.
6. Count the model calls: embeddings only. Everything else was subtraction. Some AI features are mostly arithmetic wearing an AI badge.

## Lab spec (13 min, any language)

- **Goal:** flag the anomalous condition reports for one trail using centroid distance.
- **Input:** `lab/` provides one trail's reports from `data/condition-reports.jsonl` (about 40, including a planted washout cluster) plus precomputed embeddings if you'd rather skip the embedding step.
- **How:** Ollama embeddings via `lab/ollama.http` (or the precomputed vectors), then your own vector averaging and cosine distance.
- **Steps:**
  1. Compute the centroid of all report embeddings for the trail.
  2. Score every report by distance from the centroid and sort descending.
  3. Success check: the washout reports occupy the top of your list, cleanly separated from the routine reports (compare `lab/expected-output.md`).
- **Stretch goal:** compute the centroid from a sliding 30-day window instead of all time, so "normal" adapts by season, or run the second trail's data and catch the bear-activity spike.

## Leadership beat

- **When to reach for this:** any stream of routine text where the rare exception is expensive to miss. Tickets, logs, safety reports, transaction notes, review streams.
- **Rough cost & effort:** days, and cheap to run forever. Embeddings are the only model cost; the detection itself is arithmetic your team already knows.
- **The one-liner for your CTO:** "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

This card is row 8 of the [decision framework](../../docs/decision-framework.md).
