# Lab 08: Anomaly Detection

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** find the condition reports for one trail that do not look like the rest, using distance from a centroid. Then raise one alert when several of them land close together in time.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, unnormalized, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for a stretch goal.
- **How:** embed the reports through your track's embeddings client against local Ollama; `http/ollama.http` holds the same four embed requests for the HTTP track. The centroid, distances, threshold, and alert rule are plain arithmetic you write yourself.
- **Model:** `nomic-embed-text`, local. Every code track's `starter/` runs offline on the precomputed vectors. Only step 5 onward needs Ollama.

Every step below is one thing to make the program do. Each code track's `starter/` already does steps 1 through 4 using the precomputed vectors. Read those four steps next to the starter code and find each one. Then edit the starter until it does steps 5 through 7. Compare against `complete/` when you get stuck. HTTP-track readers run the numbered requests in `http/ollama.http` where a step names one.

### Step 0: Run the starter and read the ranking

Run `starter/` as it is (`dotnet run`, `uv run main.py`, or `npm run starter`, from the track folder). It needs no model and no network.

**Check:** a 40-row table, one report per row, sorted by distance with the largest first. The top row is `cr-0496` at 0.2561 and the bottom row is `cr-0014` at 0.1484. The washout reports rise toward the top, but parking and wildflower reports are mixed in with them. That is what the technique does out of the box.

### Step 1: Load the reports

1. Open `../../data/reports-0117.jsonl`. Every line is one JSON object with four string fields: `id`, `trail_id`, `date`, `text`, for example `{"id": "cr-0009", "trail_id": "trail-0117", "date": "2025-05-04", "text": "Waterbars are doing their job, tread is in great shape. ..."}`. The file is the 40 reports for trail-0117, dated 2025-05-04 to 2026-07-22, copied line for line out of the workshop's 500-report stream in feature 10's `data/condition-reports.jsonl` (no script; it is the `trail-0117` lines in stream order). Eight of the 40 are the planted washout: five from 2026-06-18 to 2026-06-24 (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`). The other 32 are mud, ice, bugs, wildflowers, parking, and blowdown.
2. Read the file line by line, parse each line, and keep the results in a list in file order.

**Check:** the list has 40 entries. The first `id` is `cr-0009`, dated 2025-05-04.

### Step 2: Load the precomputed vectors and normalize them

1. Open `../../data/embeddings-0117.json`. It is one JSON object. Its `embeddings` field is a dictionary from report `id` to an array of 768 numbers. The file ships with the workshop and there is no build script: the vectors were captured once from local Ollama with `nomic-embed-text`, one per trail-0117 report, and stored raw, keyed by `id`, without normalizing. Its other fields (`model`, `prefix`, `dimensions`, `note`) record that.
2. Write a `normalize` function. Find the vector's length: square every component, add them up, and take the square root. Then divide every component by that length.
3. Normalize every vector as you load it. Store the results in a dictionary keyed by report `id`.

Every vector in the file is exactly what Ollama returned for `"classification: " + text`. `http/ollama.http` is four hand-written embed requests for the HTTP track, all with inputs taken from `reports-0117.jsonl`: request 1 is a routine report (`cr-0355`), request 2 is washout report `cr-0429`, request 3 is a batch of four, and request 4 is `cr-0429` again without the prefix. Request 1 shows the input shape, prefix and trailing space included:

```text
classification: Muddy in the usual low spots, gaiters not a bad idea. Sunscreen is non-negotiable up there.
```

**Check:** 40 keys, each holding 768 numbers, and every stored vector now has length 1.0.

### Step 3: Build the centroid

1. Make an array of 768 zeros. This is the centroid.
2. Loop over the 40 normalized vectors. For each vector, and for each position `i`, add `vector[i] / 40` to `centroid[i]`.
3. Normalize the centroid with the same function from step 2.

**Check:** one 768-dimension unit vector for trail-0117. If your step 4 distances are off in the third decimal from `expected-output.md`, you skipped one of the two normalizations.

### Step 4: Score every report and sort

1. Write a `cosine_distance` function: 1 minus the dot product of two vectors. This only works because both vectors are unit length.
2. For each report in the list, look up its vector by `id` and compute its distance from the centroid.
3. Sort the reports by distance, largest first.
4. Print one line per report: the distance to four decimals, the `id`, the `date`, and the `text` cut to 62 characters.

**Check:** `cr-0496` is rank 1 at 0.2561 and `cr-0429` is rank 2 at 0.2399. Routine reports are mixed in with the washout reports and there is no gap anywhere in the distances. Washout reports ranked in the 30s mean the prefix is missing.

### Step 5: Embed live with the task prefix

1. Delete the code that loads `../../data/embeddings-0117.json`.
2. Build a list of strings, one per report in list order: `"classification: "` followed by the report's `text`.
3. Embed all 40 strings in one call through your track's embeddings client, with model `nomic-embed-text` and the whole list as the input. Request 3 in `http/ollama.http` is the same call with four strings.
4. The response holds one vector per input, in the same order as the list you sent. Normalize each vector and keep it at the same index as its report.
5. Rebuild the centroid from these 40 vectors (step 3). Then score and sort again (step 4), looking up each report's vector by index instead of by `id`.

**Check:** the same 40-row table as step 4, matching `expected-output.md` to four decimals. Your program made 40 embeddings and 0 chat completions. Distances that differ in the third decimal mean a normalization is missing.

### Step 6: Derive the threshold

1. Compute the mean of the 40 distances.
2. Compute the standard deviation. Subtract the mean from each distance, square the result, average those 40 squares (divide by 40), and take the square root.
3. Set `threshold = mean + sigma * sd`, where `sigma` defaults to 1.0 (`--sigma` in `complete/`).
4. Print one line before the table: `mean distance <mean> · sd <sd> · threshold mean+1sd = <threshold>`, all to four decimals.
5. In the table, put a `!` in front of every row whose distance is above the threshold.

**Check:** mean 0.1935, sd 0.0274, threshold 0.2208. Seven rows carry the `!`. The last marked row is `cr-0329` at 0.2219, and the first unmarked row is `cr-0125` at 0.2206.

### Step 7: Group the flagged reports into alerts

1. Take every report whose distance is above the threshold and sort those by `date`, oldest first.
2. Print `7 of 40 reports above threshold. Clustering them within 14 days:`. The window defaults to 14 days (`--window` in `complete/`).
3. Walk the sorted list. Start a group at position `i`. Move `j` forward while the date at `j` minus the date at `j - 1` is 14 days or less. The group is positions `i` up to but not including `j`.
4. If the group holds 2 or more reports, print `ALERT trail-0117: <count> anomalous reports between <first date> and <last date>`. Then print one line per report with its `id`, `date`, and `text` cut to 70 characters.
5. If the group holds only 1 report, print `(ignored) <id> <date> is a lone outlier, not an event`.
6. Set `i = j` and repeat until the list is used up.
7. Print `<N> alert(s). Model calls: 40 embeddings, 0 chat completions.`

**Check:** 7 reports above the 0.2208 threshold and exactly one alert, `cr-0429`, `cr-0436`, `cr-0464`, 2026-06-18 to 2026-07-05. Four lone outliers are ignored: `cr-0282`, `cr-0329`, `cr-0354`, and `cr-0496`. All three reports in the alert are genuine washout reports.

### Stretch goals

Pick any. Each one is already built in `complete/`, and [`expected-output.md`](expected-output.md) shows the measured reason for it.

- **Drop the prefix and watch it fail.** Change the prefix in step 5 to an empty string and run again (`--raw` in `complete/`). Requests 2 and 4 in `http/ollama.http` are washout report `cr-0429` both ways, so you can see that the two vectors differ:

  ```text
  classification: The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  ```text
  The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  **Check:** bare, `cr-0429` drops to rank 11, the threshold moves to 0.2789, and two alerts fire, one on October 2025 mud and one on May 2026 glacier lilies. Not one washout report is flagged. Nothing throws and the table still looks reasonable. Put the prefix back and `cr-0429` returns to rank 2 with three washout reports in the top 5.
- **Tune the threshold.** Run with `sigma` at 1.5 (`--sigma 1.5`). **Check:** 4 reports are flagged instead of 7. Neither setting is correct. The value is a business choice about how much review you can afford.
- **Build the baseline before the anomalies arrived.** Build the centroid in step 3 from only the 32 reports dated before 2026-06-18, then score all 40 against it. Eight of the 40 reports are about the bridge. In the main path they pull the centroid toward themselves and partly hide. **Check:** all eight washout reports (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`, `cr-0464`, `cr-0480`, `cr-0496`) land in the top 10.
- **Run the other trail.** Point step 1 at `../../data/reports-0042.jsonl` (`--trail 0042` in `complete/`). Same record shape and same source as `reports-0117.jsonl`: the 25 `trail-0042` lines from feature 10's stream, dated 2025-05-04 to 2026-07-07. The planted event is bear activity, four reports from 2026-06-25 to 2026-07-02 (`cr-0446`, `cr-0449`, `cr-0453`, `cr-0455`). There are no precomputed vectors for this trail, so this goal needs Ollama. **Check:** `cr-0446` is rank 1 at 0.3067 with a visible gap to second place, the threshold is 0.2292, and one alert fires on `cr-0446` and `cr-0449`, 2026-06-25 to 2026-06-27.

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

- `data/reports-0117.jsonl`: the 40 trail-0117 reports, a slice of feature 10's full stream. Eight describe the footbridge washout (details in step 1).
- `data/reports-0042.jsonl`: the 25 trail-0042 reports from the same stream, with the bear cluster, for the last stretch goal.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for the 40 trail-0117 reports, 768 dimensions each, keyed by `id`, unnormalized (details in step 2). The starter uses these so steps 1 through 4 run without a model.
- `http/ollama.http`: the four embed requests the HTTP track runs, with inputs drawn from `reports-0117.jsonl` (listed in step 2).
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this actually works on this data.

## One Thing That Will Cost You an Hour if You Miss It

Every input in `http/ollama.http` and every vector in `embeddings-0117.json` was
produced from `"classification: " + report.text`, not from the report text alone.
`nomic-embed-text` is trained with task prefixes. Leave the prefix off and nothing
fails loudly. You just quietly get worse vectors. On this data that drops the first
washout report from rank 2 to rank 11 and makes the detector fire on October mud
instead. `expected-output.md` shows both rankings side by side.
