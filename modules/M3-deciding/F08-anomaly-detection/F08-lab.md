# Lab 08: Anomaly Detection

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** flag the anomalous condition reports for one trail using centroid distance.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, unnormalized, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for the stretch goal.
- **How:** POST to Ollama's `/api/embed`; `http/ollama.http` holds four embed requests. The centroid, distances, threshold, and alert rule are arithmetic you write.
- **Model:** `nomic-embed-text`, local. `dotnet/starter` runs offline on the precomputed vectors; only step 3 needs Ollama.

### Step 1: The trail-0117 centroid

Run `dotnet run` in `dotnet/starter/`, or load `data/embeddings-0117.json`, L2-normalize every vector, average them, then normalize the average (`Normalize` in `dotnet/complete/Program.cs`, applied at both points); request 1 in `http/ollama.http` is the input shape every vector came from, prefix and trailing space included:

```text
classification: Muddy in the usual low spots, gaiters not a bad idea. Sunscreen is non-negotiable up there.
```

**Check:** one 768-dimension unit vector for trail-0117. Step 2 distances off in the third decimal from `expected-output.md` mean you skipped a normalization.

### Step 2: Score and sort trail-0117

Score each report in `data/reports-0117.jsonl` as 1 minus the dot product of its unit vector and the centroid (`CosineDistance` in `dotnet/complete/Program.cs`), sort descending. **Check:** `cr-0496` and `cr-0429` are in the top six, routine reports mixed in, no gap in the distances. Washout reports ranked in the 30s mean the prefix is missing.

### Step 3: The task prefix on trail-0117, then the alert rule

Send all 40 texts from `data/reports-0117.jsonl` through request 3 twice, bare and with `classification: ` in front of every text, and rank both (.NET: `dotnet run -- --raw` and `dotnet run` in `dotnet/complete/`); requests 2 and 4 are washout report `cr-0429` both ways:

```text
classification: The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
```

```text
The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
```

**Check:** bare, `cr-0429` is rank 11 and two false alerts fire. Prefixed, `cr-0429` is rank 2 and three washout reports are in the top 5.

Then the rule on the trail-0117 distances: threshold at mean plus one standard deviation (`--sigma`, default 1.0), flag everything above it, sort the flagged by date, and alert when two or more fall within 14 days (`--window`, default 14). **Check:** 7 reports above the 0.2208 threshold and exactly one alert, `cr-0429`, `cr-0436`, `cr-0464`, 2026-06-18 to 2026-07-05. Lone outliers like `cr-0496` are ignored.

### Stretch goal: a baseline the anomalies did not help build

Rebuild the trail-0117 centroid from only the 32 reports dated before 2026-06-18 and re-score all 40, or run `dotnet run -- --trail 0042` on `data/reports-0042.jsonl`. **Check:** all eight trail-0117 washout reports land in the top 10. On trail-0042, `cr-0446` is rank 1 and one alert fires on `cr-0446` and `cr-0449`.

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
