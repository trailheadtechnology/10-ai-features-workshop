# 08 · Anomaly Detection

Module 3: Deciding · Runs on Ollama embeddings (`nomic-embed-text`) plus plain math

## The User Problem

Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

## The Concept

This feature is barely an AI feature: it's embeddings plus arithmetic. Embed every condition report for a trail, and the routine ones ("muddy," "buggy," "fine") cluster together in vector space. Average them and you get a centroid, the mathematical center of "normal" for that trail. A report's distance from that centroid is an anomaly score. "Bridge washed out" sits farther from the mud cluster than the mud reports sit from each other, so it rises without a large model or any training. Cosine distance and a threshold do most of the job.

The word "most" is doing real work in that sentence, and this feature is honest about it. Ranking single reports by distance is noisy: routine reports about parking or wildflowers can outrank a genuine hazard, and when 8 of 40 reports describe the same washout, the anomalies drag the centroid toward themselves and partially hide. Two things rescue it, and both are the actual lesson. First, embedding models have contracts: `nomic-embed-text` is trained with task prefixes, and embedding `"classification: " + text` instead of the bare text moves the first washout report from rank 11 to rank 2. Second, the alert rule beats the ranking. Requiring two flagged reports within a two-week window fires exactly one alert on this trail, all three of its reports genuine, zero false positives. One outlier might be a rambling hiker; several outliers in a week that also sit near each other are an event.

The pattern generalizes to any stream of routine text: support tickets, log messages, form submissions, review streams. Define normal from the data itself, and let distance flag what deserves human eyes. It also pairs naturally with feature 07: classification handles the categories you knew to define, and anomaly detection catches the things you didn't.

## The Lab

The hands-on lab is [F08-lab.md](F08-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** any stream of routine text where the rare exception is expensive to miss. Tickets, logs, safety reports, transaction notes, review streams.
- **Rough cost & effort:** days, and cheap to run forever. Embeddings are the only model cost; the detection itself is arithmetic your team already knows.
- **The one-liner for your CTO:** "We hear about the washed-out bridge from the first three reports, not from a one-star review a month later."

This card is row 8 of the [decision framework](../../../docs/decision-framework.md).
