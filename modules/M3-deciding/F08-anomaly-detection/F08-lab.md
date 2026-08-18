# Lab 08: Anomaly Detection

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** flag the anomalous condition reports for one trail using centroid distance.
- **Input:** `data/` provides one trail's reports from the full stream (feature 10's `data/condition-reports.jsonl`) (about 40, including a planted washout cluster) plus precomputed embeddings if you'd rather skip the embedding step.
- **How:** Ollama embeddings via `http/ollama.http` (or the precomputed vectors), then your own vector averaging and cosine distance.
- **Steps:**
  1. Compute the centroid of all report embeddings for the trail.
  2. Score every report by distance from the centroid and sort descending.
  3. Success check: washout reports rise toward the top, but not cleanly, and your ranking will have routine reports mixed in (compare `expected-output.md`). Then add the `classification:` prefix to each text before embedding and watch the ranking improve. Then apply the alert rule (threshold plus two flagged reports within 14 days) and check that it fires once, on real reports.
- **Stretch goal:** compute the centroid only from reports before the washout window, so the anomalies stop dragging "normal" toward themselves. That puts all 8 washout reports in the top 10 and is the seed of a real sliding-window detector. Or run the second trail's data and catch the bear-activity spike, which separates more sharply.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F08-http.md`](http/F08-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F08-dotnet.md`](dotnet/F08-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F08-python.md`](python/F08-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F08-typescript.md`](typescript/F08-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/reports-0117.jsonl`: all 40 condition reports for trail-0117, straight out of the full stream, feature 10's `data/condition-reports.jsonl`. Five of them (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) describe a footbridge washing out between 2026-06-18 and 2026-06-24, and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`) say it is still out. The other 32 are mud, ice, bugs, and blowdown.
- `data/reports-0042.jsonl`: the 25 reports for trail-0042, for the stretch goal. The bear cluster runs `cr-0446` through `cr-0455`, late June into early July 2026.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for all 40 trail-0117 reports, 768 dimensions each, exactly as Ollama returned them. Use these if you would rather skip the embedding step and get straight to the arithmetic. They are unnormalized, so L2-normalize before you use them.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this actually works on this data.

## One Thing That Will Cost You an Hour if You Miss It

Every input in `http/ollama.http` and every vector in `embeddings-0117.json` was
produced from `"classification: " + report.text`, not from the report text alone.
`nomic-embed-text` is trained with task prefixes, and omitting one does not fail
loudly, it just quietly hands you worse vectors. On this data that drops the first
washout report from rank 2 to rank 11 and makes the detector fire on October mud
instead. `expected-output.md` shows both rankings side by side.
