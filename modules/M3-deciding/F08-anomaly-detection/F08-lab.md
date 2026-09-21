<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 08: Anomaly Detection

**The User Problem:** Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** find the condition reports for one trail that do not look like the rest, using distance from a centroid. Then raise one alert when several of them land close together in time.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for a stretch goal.
- **How:** embed the reports through your track's embeddings client against local Ollama. The centroid, distances, threshold, and alert rule are plain arithmetic you write yourself.
- **Model:** `nomic-embed-text`, local. Every track's `starter/` runs offline on the precomputed vectors. Only step 5 onward needs Ollama.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each track's lab has the full steps plus the exact calls, imports, and run commands for that language.

| Track | Open this lab | What you edit |
|---|---|---|
| .NET | [`dotnet/F08-dotnet.md`](dotnet/F08-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F08-python.md`](python/F08-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F08-typescript.md`](typescript/F08-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/reports-0117.jsonl`: the 40 trail-0117 reports, a slice of feature 10's full stream. Eight describe the footbridge washout (details in step 1).
- `data/reports-0042.jsonl`: the 25 trail-0042 reports from the same stream, with the bear cluster, for the last stretch goal.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for the 40 trail-0117 reports, 768 dimensions each, keyed by `id`, already unit length as `nomic-embed-text` returns them (details in step 2). The starter uses these so steps 1 through 4 run without a model.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this works on this data.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
