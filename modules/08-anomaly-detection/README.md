# 08 · Anomaly detection

Block 3 (Deciding) · Runs on Ollama embeddings (`nomic-embed-text`) plus plain math

## The user problem

Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

## The concept

This module has a secret, and the secret is that it's barely an AI module: it's embeddings plus arithmetic. Embed every condition report for a trail, and the routine ones ("muddy," "buggy," "fine") cluster together in vector space. Average them and you get a centroid, the mathematical center of "normal" for that trail. A report's distance from that centroid is an anomaly score. "Bridge washed out" sits farther from the mud cluster than the mud reports sit from each other, so it rises without a large model or any training. Cosine distance and a threshold do most of the job.

The word "most" is doing real work in that sentence, and this module is honest about it. Ranking single reports by distance is noisy: routine reports about parking or wildflowers can outrank a genuine hazard, and when 8 of 40 reports describe the same washout, the anomalies drag the centroid toward themselves and partially hide. Two things rescue it, and both are the actual lesson. First, embedding models have contracts: `nomic-embed-text` is trained with task prefixes, and embedding `"classification: " + text` instead of the bare text moves the first washout report from rank 11 to rank 2. Second, the alert rule beats the ranking. Requiring two flagged reports within a two-week window fires exactly one alert on this trail, all three of its reports genuine, zero false positives. One outlier might be a rambling hiker; several outliers in a week that also sit near each other are an event.

The pattern generalizes to any stream of routine text: support tickets, log messages, form submissions, review streams. Define normal from the data itself, and let distance flag what deserves human eyes. It also pairs naturally with module 07: classification handles the categories you knew to define, and anomaly detection catches the things you didn't.

## Demo outline (13 min, .NET)

1. Scroll the condition-report stream. Boring on purpose. Ask the room to find the problem; nobody can, and that's the situation the park is in.
2. Embed one trail's reports with the module 04 embedding code, average the vectors into a centroid, and put the idea on one slide: the center of normal.
3. Run it with `--raw` first and print every report's distance, sorted. It underwhelms: the washout reports scatter through the middle of the list. Sit in that for a second, because this is what the technique actually does out of the box.
4. Add the model's task prefix (`classification:`) and re-run. A washout report jumps to rank 2 and the mud reports settle at the bottom. The lesson: embedding models have usage contracts, and reading the model card is engineering work, not homework.
5. Now the alert rule, which is where the feature actually lives: distance beyond threshold, plus two or more flagged reports within a two-week window. One alert fires on this trail, three genuine washout reports in it, nothing false. Show it also catching the bear-activity spike on the other trail, where the signal is even cleaner because that trail's routine chatter is more uniform. Accuracy here is a property of your corpus, not your code.
6. Count the model calls: embeddings only. Everything else was subtraction. Some AI features are mostly arithmetic wearing an AI badge.

## Lab spec (13 min, any language)

- **Goal:** flag the anomalous condition reports for one trail using centroid distance.
- **Input:** `lab/` provides one trail's reports from `data/condition-reports.jsonl` (about 40, including a planted washout cluster) plus precomputed embeddings if you'd rather skip the embedding step.
- **How:** Ollama embeddings via `lab/ollama.http` (or the precomputed vectors), then your own vector averaging and cosine distance.
- **Steps:**
  1. Compute the centroid of all report embeddings for the trail.
  2. Score every report by distance from the centroid and sort descending.
  3. Success check: washout reports rise toward the top, but not cleanly, and your ranking will have routine reports mixed in (compare `lab/expected-output.md`). Then add the `classification:` prefix to each text before embedding and watch the ranking improve. Then apply the alert rule (threshold plus two flagged reports within 14 days) and check that it fires once, on real reports.
- **Stretch goal:** compute the centroid only from reports before the washout window, so the anomalies stop dragging "normal" toward themselves. That puts all 8 washout reports in the top 10 and is the seed of a real sliding-window detector. Or run the second trail's data and catch the bear-activity spike, which separates more sharply.

## Leadership beat

- **When to reach for this:** any stream of routine text where the rare exception is expensive to miss. Tickets, logs, safety reports, transaction notes, review streams.
- **Rough cost & effort:** days, and cheap to run forever. Embeddings are the only model cost; the detection itself is arithmetic your team already knows.
- **The one-liner for your CTO:** "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

This card is row 8 of the [decision framework](../../docs/decision-framework.md).
