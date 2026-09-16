# Lab 08: Anomaly Detection

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** find the condition reports for one trail that do not look like the rest, using distance from a centroid. Then raise one alert when several of them land close together in time.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, unnormalized, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for a stretch goal.
- **How:** embed the reports through your track's embeddings client against local Ollama. The centroid, distances, threshold, and alert rule are plain arithmetic you write yourself.
- **Model:** `nomic-embed-text`, local. Every track's `starter/` runs offline on the precomputed vectors. Only step 5 onward needs Ollama.

Every step below is one thing to make the program do. Every track's `starter/` already does steps 1 through 4 using the precomputed vectors. Read those four steps next to the starter code and find each one. Then edit the starter until it does steps 5 through 7. Compare against `complete/` when you get stuck.

### Step 0: Run the starter and read the ranking

**Do:** run `starter/` as it is. It needs no model and no network.

**Check:** a 40-row table, one report per row, sorted by distance with the largest first. The top row is `cr-0496` at `0.2561` and the bottom row is `cr-0014` at `0.1484`. The washout reports rise toward the top, but parking and wildflower reports are mixed in with them. That is what the technique does out of the box.

### Step 1: Load the reports

**Do:**
1. Open `../../data/reports-0117.jsonl`: `id`, `trail_id`, `date`, `text` per line.
2. Read line by line, parse each line, keep the results in a list in file order.

**Why:** the file is the 40 reports for trail-0117, dated 2025-05-04 to 2026-07-22, copied out of the workshop's 500-report stream in feature 10's data. Eight of the 40 are the planted washout: five from 2026-06-18 to 2026-06-24 (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`). The other 32 are mud, ice, bugs, wildflowers, parking, and blowdown.

**Check:** the list has 40 entries. The first `id` is `cr-0009`, dated 2025-05-04.

### Step 2: Load the precomputed vectors and normalize them

**Do:**
1. Open `../../data/embeddings-0117.json`. Its `embeddings` field maps report `id` to an array of 768 numbers.
2. Write `normalize(vector)`: find its length (square every component, sum, square root), then divide every component by that length.
3. Normalize every vector as you load it, storing results in a dictionary keyed by `id`.

**Why:** the vectors were captured once from local Ollama with `nomic-embed-text`, from `"classification: " + text`, stored raw and unnormalized.

**Check:** 40 keys, each holding 768 numbers, and every stored vector now has length 1.0.

### Step 3: Build the centroid

**Do:**
1. Make an array of 768 zeros — the centroid.
2. Loop over the 40 normalized vectors; for each vector and position `i`, add `vector[i] / 40` to `centroid[i]`.
3. Normalize the centroid with the same function from step 2.

**Check:** one 768-dimension unit vector for trail-0117. If your step 4 distances are off in the third decimal from `expected-output.md`, you skipped one of the two normalizations.

### Step 4: Score every report and sort

**Do:**
1. Write `cosine_distance(a, b)`: 1 minus the dot product (this only works because both vectors are unit length).
2. For each report, look up its vector by `id` and compute its distance from the centroid.
3. Sort by distance, largest first.
4. Print one line per report: distance to four decimals, `id`, `date`, and `text` cut to 62 characters.

**Check:** `cr-0496` is rank 1 at `0.2561` and `cr-0429` is rank 2 at `0.2399`. Routine reports are mixed in with the washout reports and there is no gap anywhere in the distances. Washout reports ranked in the 30s mean the prefix is missing.

### Step 5: Embed live with the task prefix

**Do:**
1. Delete the code that loads `../../data/embeddings-0117.json`.
2. Build a list of strings, one per report in list order: `"classification: "` followed by the report's `text`.
3. Embed all 40 strings in **one call** with model `nomic-embed-text`.
4. Normalize each returned vector and keep it at the same index as its report.
5. Rebuild the centroid from these 40 vectors (step 3), then score and sort again (step 4), looking up each vector by index instead of by `id`.

**Check:** the same 40-row table as step 4, matching `expected-output.md` to four decimals. Your program made 40 embeddings and 0 chat completions. Distances that differ in the third decimal mean a normalization is missing.

### Step 6: Derive the threshold

**Do:**
1. Compute the mean of the 40 distances.
2. Compute the standard deviation: subtract the mean from each distance, square, average the 40 squares, take the square root.
3. Set `threshold = mean + sigma * sd`, `sigma` defaulting to 1.0 (`--sigma` in `complete/`).
4. Print `mean distance <mean> · sd <sd> · threshold mean+1sd = <threshold>`, all to four decimals.
5. In the table, put a `!` in front of every row above the threshold.

**Check:** mean `0.1935`, sd `0.0274`, threshold `0.2208`. Seven rows carry the `!`. The last marked row is `cr-0329` at `0.2219`, and the first unmarked row is `cr-0125` at `0.2206`.

### Step 7: Group the flagged reports into alerts

**Do:**
1. Take every report above the threshold and sort by `date`, oldest first.
2. Print `7 of 40 reports above threshold. Clustering them within 14 days:` (window defaults to 14 days, `--window` in `complete/`).
3. Walk the sorted list: start a group at position `i`, move `j` forward while the date at `j` minus the date at `j-1` is 14 days or less. The group is positions `i` up to but not including `j`.
4. If the group holds 2+ reports, print `ALERT trail-0117: <count> anomalous reports between <first date> and <last date>`, then one line per report (`id`, `date`, text cut to 70 characters).
5. If the group holds 1 report, print `(ignored) <id> <date> is a lone outlier, not an event`.
6. Set `i = j` and repeat until the list is used up.
7. Print `<N> alert(s). Model calls: 40 embeddings, 0 chat completions.`

**Check:** 7 reports above the `0.2208` threshold and exactly one alert, `cr-0429`, `cr-0436`, `cr-0464`, 2026-06-18 to 2026-07-05. Four lone outliers are ignored: `cr-0282`, `cr-0329`, `cr-0354`, and `cr-0496`. All three reports in the alert are genuine washout reports.

### Stretch goals

Pick any. Each one is already built in `complete/`, and [`expected-output.md`](expected-output.md) shows the measured reason for it.

- **Drop the prefix and watch it fail.** Change the prefix in step 5 to an empty string and run again (`--raw` in `complete/`). Compare the two versions of the same input:

  ```text
  classification: The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  ```text
  The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  **Check:** bare, `cr-0429` drops to rank 11, the threshold moves to `0.2789`, and two alerts fire, one on October 2025 mud and one on May 2026 glacier lilies. Not one washout report is flagged. Nothing throws and the table still looks reasonable. Put the prefix back and `cr-0429` returns to rank 2 with three washout reports in the top 5.
- **Tune the threshold.** Run with `sigma` at 1.5 (`--sigma 1.5`). **Check:** 4 reports are flagged instead of 7. Neither setting is correct — the value is a business choice about how much review you can afford.
- **Build the baseline before the anomalies arrived.** Build the centroid in step 3 from only the 32 reports dated before 2026-06-18, then score all 40 against it. **Why:** eight of the 40 reports are about the bridge; in the main path they pull the centroid toward themselves and partly hide. **Check:** all eight washout reports (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`, `cr-0464`, `cr-0480`, `cr-0496`) land in the top 10.
- **Run the other trail.** Point step 1 at `../../data/reports-0042.jsonl` (`--trail 0042` in `complete/`): the 25 `trail-0042` lines from feature 10's stream, dated 2025-05-04 to 2026-07-07. The planted event is bear activity, four reports from 2026-06-25 to 2026-07-02 (`cr-0446`, `cr-0449`, `cr-0453`, `cr-0455`). There are no precomputed vectors for this trail, so this goal needs Ollama. **Check:** `cr-0446` is rank 1 at `0.3067` with a visible gap to second place, the threshold is `0.2292`, and one alert fires on `cr-0446` and `cr-0449`, 2026-06-25 to 2026-06-27.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F08-dotnet.md`](dotnet/F08-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F08-python.md`](python/F08-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F08-typescript.md`](typescript/F08-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/reports-0117.jsonl`: the 40 trail-0117 reports, a slice of feature 10's full stream. Eight describe the footbridge washout (details in step 1).
- `data/reports-0042.jsonl`: the 25 trail-0042 reports from the same stream, with the bear cluster, for the last stretch goal.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for the 40 trail-0117 reports, 768 dimensions each, keyed by `id`, unnormalized (details in step 2). The starter uses these so steps 1 through 4 run without a model.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this actually works on this data.

## One Thing That Will Cost You an Hour if You Miss It

Every vector in `embeddings-0117.json` was produced from `"classification: " + report.text`,
not from the report text alone. `nomic-embed-text` is trained with task prefixes. Leave the
prefix off and nothing fails loudly — you just quietly get worse vectors. On this data that
drops the first washout report from rank 2 to rank 11 and makes the detector fire on October
mud instead. `expected-output.md` shows both rankings side by side.
