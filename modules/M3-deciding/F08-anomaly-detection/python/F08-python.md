<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 08: Anomaly Detection (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F08-dotnet.md), [TypeScript](../typescript/F08-typescript.md). Lab overview: [F08-lab.md](../F08-lab.md).*

**The User Problem:** Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

*A Challenge lab. Do it if you finished [Module 3](../../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** find the condition reports for one trail that do not look like the rest, using distance from a centroid. Then raise one alert when several of them land close together in time.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for a stretch goal.
- **How:** embed the reports through your track's embeddings client against local Ollama. The centroid, distances, threshold, and alert rule are plain arithmetic you write yourself.
- **Model:** `nomic-embed-text`, local. Every track's `starter/` runs offline on the precomputed vectors. Only step 5 onward needs Ollama.

## The Concept

This feature is barely an AI feature: it's embeddings plus arithmetic. Embed every condition report for a trail, and the routine ones ("muddy," "buggy," "fine") cluster together in vector space. Average them and you get a centroid, the mathematical center of "normal" for that trail. A report's distance from that centroid is an anomaly score. "Bridge washed out" sits farther from the mud cluster than the mud reports sit from each other, so it rises without a large model or any training. Cosine distance and a threshold do most of the job.

The word "most" is doing real work in that sentence, and this feature is honest about it. Ranking single reports by distance is noisy: routine reports about parking or wildflowers can outrank a genuine hazard, and when 8 of 40 reports describe the same washout, the anomalies drag the centroid toward themselves and partially hide. Two things rescue it, and both are the actual lesson. First, embedding models have contracts: `nomic-embed-text` is trained with task prefixes, and embedding `"classification: " + text` instead of the bare text moves the first washout report from rank 11 to rank 2. Second, the alert rule beats the ranking. Requiring two flagged reports within a two-week window fires exactly one alert on this trail, all three of its reports genuine, zero false positives. One outlier might be a rambling hiker; several outliers in a week that also sit near each other are an event.

The pattern generalizes to any stream of routine text: support tickets, log messages, form submissions, review streams. Define normal from the data itself, and let distance flag what deserves human eyes. It also pairs naturally with feature 07: classification handles the categories you knew to define, and anomaly detection catches the things you didn't.

Every step below is one thing to make the program do. The `starter/` already does steps 1 through 4 using the precomputed vectors. Read those four steps next to the starter code and find each one. Then edit the starter until it does steps 5 through 7. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts:

- `starter/main.py`: no model, no network. Loads the precomputed vectors from `../../data/embeddings-0117.json`, averages them into a centroid, ranks every report by cosine distance from it. Runs with Ollama down.
- `complete/main.py`: the finished demo as shown on stage. Embeds live with the `classification:` task prefix, derives the threshold from the corpus (mean plus sigma standard deviations), and applies the alert rule: two or more flagged reports within a 14-day window. One alert fires, three genuine washout reports in it.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [SETUP.md](../../../../SETUP.md)), and `uv run` finds it from any folder. Run `uv run main.py` from the `starter/` folder.

From `complete/` (`starter/main.py` takes no flags):

```bash
uv run main.py                  # trail-0117, live embeddings, distance table + cluster alerts
uv run main.py --raw            # same trail with the task prefix removed
uv run main.py --trail 0042     # the bear-activity trail
uv run main.py --sigma 1.5      # tighter threshold
uv run main.py --window 30      # wider clustering window, in days
```

`--raw`, `--sigma`, `--window`, and `--trail` exist only in `complete/`, so hard-code the value or add the same argument parsing.

### Step 0: Run the starter and read the ranking

**Do:** run `starter/` as it is. It needs no model and no network.

```bash
uv run main.py
```

**Check:** a 40-row table, one report per row, sorted by distance with the largest first. The top row is `cr-0496` at `0.2561` and the bottom row is `cr-0014` at `0.1484`. The washout reports rise toward the top, but parking and wildflower reports are mixed in with them. That is what the technique does out of the box.

### Step 1: Load the reports

**Do:**
1. Open `../../data/reports-0117.jsonl`: `id`, `trail_id`, `date`, `text` per line.
2. Review how the existing code from the starter project reads the file line by line, parses each line, and keeps the results in a list in file order.

   `DATA` is the `data/` folder two levels above `main.py`. The `reports` line reads the whole file, splits it into lines, skips blank ones, and turns each line into a dict with `json.loads`, so `r["id"]` and `r["text"]` work:

   ```python
   DATA = Path(__file__).resolve().parents[2] / "data"
   ```

   ```python
   reports = [json.loads(l) for l in (DATA / "reports-0117.jsonl").read_text().splitlines() if l.strip()]
   ```

   To see the Check for yourself, you can add a temporary line below it and delete it afterward:

   ```python
   # Hint: temporary, prints the count and the first report
   print(len(reports), reports[0]["id"], reports[0]["date"])
   ```

**Why:** the file is the 40 reports for trail-0117, dated 2025-05-04 to 2026-07-22, copied out of the workshop's 500-report stream in feature 10's data. Eight of the 40 are the planted washout: five from 2026-06-18 to 2026-06-24 (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`). The other 32 are mud, ice, bugs, wildflowers, parking, and blowdown.

**Check:** the list has 40 entries. The first `id` is `cr-0009`, dated 2025-05-04.

### Step 2: Load the precomputed vectors and normalize them

**Do:**
1. Review how the existing code from the starter project opens `../../data/embeddings-0117.json`. Its `embeddings` field maps report `id` to an array of 768 numbers.

   It happens inside the `vectors` line (item 3): `json.loads((DATA / "embeddings-0117.json").read_text())["embeddings"]` reads the file and pulls out the `id`-to-numbers dict.

2. Review how it normalizes a vector: find its length (square every component, sum, square root), then divide every component by that length.

   It is a function near the top of `main.py` (`import math` is at the top of the file):

   ```python
   def normalize(vector: list[float]) -> list[float]:
       length = math.sqrt(sum(v * v for v in vector))
       return [v / length for v in vector]
   ```

3. Review how it normalizes every vector as it loads it, storing results in a dictionary keyed by `id`.

   It happens with a dict comprehension: for each `k, v` pair in the `embeddings` dict, the key stays the report `id` and the value becomes `normalize(v)`:

   ```python
   vectors = {k: normalize(v) for k, v in json.loads((DATA / "embeddings-0117.json").read_text())["embeddings"].items()}
   ```

   To see the Check for yourself, add a temporary line below it (a unit vector's length prints as `1.0` or very close to it):

   ```python
   # Hint: temporary, prints the key count, the size, and the length of one vector
   first = next(iter(vectors.values()))
   print(len(vectors), len(first), math.sqrt(sum(v * v for v in first)))
   ```

**Why:** the vectors were captured once from local Ollama with `nomic-embed-text`, from `"classification: " + text`. `nomic-embed-text` already returns unit-length vectors, so normalizing these is a no-op and the check below passes either way. Write the function anyway: step 5 replaces this file with vectors you embed yourself, cosine distance is only 1 minus the dot product when both sides are unit length, and an embedding model that does not normalize for you is the common case.

**Check:** 40 keys, each holding 768 numbers, and every stored vector has length 1.0. It did before you normalized them too; the point is that your `normalize` works, so test it on something that is not unit length, such as `[3, 4]`, which should come back `[0.6, 0.8]`.

### Step 3: Build the centroid

**Do:**
1. Review how the existing code from the starter project makes an array of 768 zeros. This is the centroid.

   `dimensions` is the length of any one vector (768), and `[0.0] * dimensions` is a list of that many zeros:

   ```python
   dimensions = len(next(iter(vectors.values())))
   centroid = [0.0] * dimensions
   ```

2. Review how it loops over the 40 normalized vectors; for each vector and position `i`, it adds `vector[i] / 40` to `centroid[i]`.

   `enumerate(vector)` gives each position `i` together with its number `v`, and `len(vectors)` is 40:

   ```python
   for vector in vectors.values():
       for i, v in enumerate(vector):
           centroid[i] += v / len(vectors)
   ```

3. Review how it normalizes the centroid with the same function from step 2.

   It happens in one line:

   ```python
   centroid = normalize(centroid)
   ```

**Check:** one 768-dimension unit vector for trail-0117. If your step 4 distances are off in the third decimal from `expected-output.md`, you skipped one of the two normalizations.

### Step 4: Score every report and sort

**Do:**
1. Review how the existing code from the starter project computes the cosine distance between two vectors: 1 minus the dot product (this only works because both vectors are unit length).

   It is a function below `normalize`. `zip(a, b)` walks both lists side by side, so `x * y` multiplies matching positions:

   ```python
   def cosine_distance(a: list[float], b: list[float]) -> float:
       return 1.0 - sum(x * y for x, y in zip(a, b))
   ```

2. Review how it looks up each report's vector by `id` and computes its distance from the centroid.
3. Review how it sorts by distance, largest first.

   Items 2 and 3 happen in one line. The inner part builds a `(distance, report)` pair for every report, looking up `vectors[r["id"]]`, and `sorted` with `key=lambda x: -x[0]` puts the largest distance first:

   ```python
   scored = sorted(((cosine_distance(vectors[r["id"]], centroid), r) for r in reports), key=lambda x: -x[0])
   ```

4. Review how it prints one line per report: distance to four decimals, `id`, `date`, and `text` cut to 62 characters.

   It happens at the bottom of `main.py`. `{distance:.4f}` is four decimals, and each pair in `scored` unpacks into `distance, report`:

   ```python
   print("  dist    id       date        report")
   for distance, report in scored:
       print(f"  {distance:.4f}  {report['id']}  {report['date']}  {truncate(report['text'], 62)}")
   ```

   The cut is a function below `cosine_distance`:

   ```python
   def truncate(text: str, n: int) -> str:
       return text if len(text) <= n else text[: n - 1] + "…"
   ```

**Check:** `cr-0496` is rank 1 at `0.2561` and `cr-0429` is rank 2 at `0.2399`. Routine reports are mixed in with the washout reports and there is no gap anywhere in the distances. Washout reports ranked in the 30s mean the prefix is missing.

### Step 5: Embed live with the task prefix

**Do:**
1. Delete the code that loads `../../data/embeddings-0117.json`.
2. Build a list of strings, one per report in list order: `"classification: "` followed by the report's `text`.
3. Embed all 40 strings in **one call** with model `nomic-embed-text`. Ollama listens on `http://localhost:11434`.

   The starter imports no AI package. The client is the `openai` package (already installed by `uv sync`) pointed at Ollama. Add the import below the starter's `from pathlib import Path` line, and the client directly above the `DATA = ...` line. `api_key` can be any text; Ollama ignores it, but the package refuses to start without one:

   ```python
   from openai import OpenAI
   ```

   ```python
   client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
   ```

   Then replace the `vectors = {...}` line that reads `embeddings-0117.json` with these three lines. The `input=[...]` list comprehension builds the 40 prefixed strings from item 2, in report order, and `client.embeddings.create` sends them all in one request:

   ```python
   prefix = "classification: "
   response = client.embeddings.create(model="nomic-embed-text", input=[prefix + r["text"] for r in reports])
   vectors = [normalize(d.embedding) for d in response.data]
   ```

4. Normalize each returned vector and keep it at the same index as its report. Replace the `id`-keyed dictionary with a plain list in report order, so anything that indexed by `id` now indexes by position.

   `response.data` comes back in input order, so `vectors[i]` belongs to `reports[i]`. `vectors` is now a list, not a dict keyed by `id`, so `vectors.values()` no longer works. Replace the starter's centroid lines, from `dimensions = ...` down to `centroid = normalize(centroid)`, with:

   ```python
   dimensions = len(vectors[0])
   centroid = [0.0] * dimensions
   for vector in vectors:
       for k, v in enumerate(vector):
           centroid[k] += v / len(vectors)
   centroid = normalize(centroid)
   ```

   Then replace the starter's `scored = ...` line so it looks up by position. `zip(vectors, reports)` pairs the first vector with the first report, the second with the second, and so on:

   ```python
   scored = sorted(((cosine_distance(v, centroid), r) for v, r in zip(vectors, reports)), key=lambda x: -x[0])
   ```

5. Rebuild the centroid from these 40 vectors (step 3), then score and sort again (step 4), looking up each vector by index instead of by `id`. Run it with Ollama up:

   ```bash
   uv run main.py
   ```

**Check:** the same 40-row table as step 4, matching `expected-output.md` to four decimals. Your program made 40 embeddings and 0 chat completions. Distances that differ in the third decimal mean a normalization is missing.

### Step 6: Derive the threshold

**Do:**
1. Compute the mean of the 40 distances.
2. Compute the standard deviation: subtract the mean from each distance, square, average the 40 squares, take the square root.
3. Set `threshold = mean + sigma * sd`, `sigma` defaulting to 1.0 (`--sigma` in `complete/`).
4. Print `mean distance <mean> · sd <sd> · threshold mean+1sd = <threshold>`, all to four decimals.

   Put `sigma` near the top of `main.py`, directly below the `DATA = ...` line:

   ```python
   sigma = 1.0
   ```

   Directly below the `scored = ...` line, compute the mean and the population standard deviation of the 40 distances. Each item in `scored` is a `(distance, report)` pair, so `for d, _ in scored` takes the distance as `d` and ignores the report (`_`). `** 2` squares, and `math.sqrt` is the square root:

   ```python
   mean = sum(d for d, _ in scored) / len(scored)
   deviation = math.sqrt(sum((d - mean) ** 2 for d, _ in scored) / len(scored))
   threshold = mean + sigma * deviation
   ```

   Print it between the starter's `print(f"trail-0117 · ...")` header line and its `print("  dist    id       date        report")` line, so it lands above the table. `{sigma:g}` prints `1.0` as `1`:

   ```python
   print(f"mean distance {mean:.4f} · sd {deviation:.4f} · threshold mean+{sigma:g}sd = {threshold:.4f}\n")
   ```

   The starter's header line ends in `\n`, so a blank line prints above this one. That is fine; only the numbers matter.

5. In the table, put a `!` in place of the first leading space of every row above the threshold, so columns stay aligned.

   Replace the `print(...)` line inside the starter's `for distance, report in scored:` loop (the last line of `main.py`) with this one, keeping its four-space indent. `' !' if distance > threshold else '  '` picks one of the two 2-character prefixes:

   ```python
       print(f"{' !' if distance > threshold else '  '}{distance:.4f}  {report['id']}  {report['date']}  {truncate(report['text'], 62)}")
   ```

**Check:** mean `0.1935`, sd `0.0274`, threshold `0.2208`. Seven rows carry the `!`. The last marked row is `cr-0329` at `0.2219`, and the first unmarked row is `cr-0125` at `0.2206`.

### Step 7: Group the flagged reports into alerts

**Do:**
1. Take every report above the threshold and sort by `date`, oldest first.

   Everything in this step goes at the very bottom of `main.py`, after the table loop, with no indent. The comprehension keeps only the `(distance, report)` pairs above the threshold, and `key=lambda x: x[1]["date"]` sorts them by the report's `date` (ISO text like `2026-06-18` sorts oldest first):

   ```python
   flagged = sorted(((d, r) for d, r in scored if d > threshold), key=lambda x: x[1]["date"])
   ```

2. Print `7 of 40 reports above threshold. Clustering them within 14 days:` (window defaults to 14 days, `--window` in `complete/`).

   Put `window` near the top of `main.py`, below the `sigma = 1.0` line from step 6:

   ```python
   window = 14
   ```

   Then, below the `flagged = ...` line:

   ```python
   print(f"\n{len(flagged)} of {len(scored)} reports above threshold. Clustering them within {window} days:\n")
   ```

3. Walk the sorted list: start a group at position `i`, move `j` forward while the date at `j` minus the date at `j-1` is 14 days or less. The group is positions `i` up to but not including `j`. `date` is a string; parse it to a real date before subtracting (ISO strings sort correctly as text but do not subtract).

   `date.fromisoformat` from the standard library turns the ISO string into a real date, and subtracting two of those gives a `timedelta` whose `.days` is the gap. Add the import at the top of `main.py`, below `import math`:

   ```python
   from datetime import date
   ```

   Then, below the print from item 2, start the counter and the outer loop. `flagged[j][1]["date"]` is the `date` of the report in pair `j` (`[0]` is the distance, `[1]` is the report), and `flagged[i:j]` is a new list of positions `i` up to but not including `j`:

   ```python
   alerts = 0
   i = 0
   while i < len(flagged):
       j = i + 1
       while j < len(flagged) and (date.fromisoformat(flagged[j][1]["date"]) - date.fromisoformat(flagged[j - 1][1]["date"])).days <= window:
           j += 1
       group = flagged[i:j]
   ```

   Items 4 to 6 go inside this `while i < len(flagged):` loop, indented four spaces like `group = ...`.

4. If the group holds 2+ reports, print `ALERT trail-0117: <count> anomalous reports between <first date> and <last date>`, then one line per report (`id`, `date`, text cut to 70 characters).

   `complete/` names the trail in a variable so the last stretch goal can change it. Put it near the top of `main.py`, directly below the `DATA = ...` line:

   ```python
   trail = "0117"
   ```

   Then, below `group = flagged[i:j]`, at the same indent. `group[0]` is the first pair and `group[-1]` the last, and `for _, r in group` takes each report and ignores its distance:

   ```python
       if len(group) >= 2:
           alerts += 1
           print(f"  ALERT trail-{trail}: {len(group)} anomalous reports between {group[0][1]['date']} and {group[-1][1]['date']}")
           for _, r in group:
               print(f"        {r['id']} {r['date']}  {truncate(r['text'], 70)}")
   ```

5. If the group holds 1 report, print `(ignored) <id> <date> is a lone outlier, not an event`.

   Directly below the `if` block, lined up with the `if`:

   ```python
       else:
           print(f"  (ignored) {group[0][1]['id']} {group[0][1]['date']} is a lone outlier, not an event")
   ```

6. Set `i = j` and repeat until the list is used up.

   Last line inside the `while i < len(flagged):` loop, lined up with the `if` and `else`. Without it the loop never ends:

   ```python
       i = j
   ```

7. Print `<N> alert(s). Model calls: 40 embeddings, 0 chat completions.`

   After the loop, with no indent, as the last line of `main.py`:

   ```python
   print(f"\n{alerts} alert(s). Model calls: {len(reports)} embeddings, 0 chat completions.")
   ```

Put together, the bottom of `main.py` is the `flagged` line, the item 2 print, then the loop, with the final print below it. Run:

```bash
uv run main.py
```

**Check:** 7 reports above the `0.2208` threshold and exactly one alert, `cr-0429`, `cr-0436`, `cr-0464`, 2026-06-18 to 2026-07-05. Four lone outliers are ignored: `cr-0282`, `cr-0329`, `cr-0354`, and `cr-0496`. All three reports in the alert are genuine washout reports.

### Stretch goals

Pick any. Each one is already built in `complete/`, and [`expected-output.md`](../expected-output.md) shows the measured reason for it.

- **Drop the prefix and watch it fail.** Change the prefix in step 5 to an empty string and run again (`--raw` in `complete/`). Compare the two versions of the same input:

  ```text
  classification: The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  ```text
  The footbridge over the gorge is completely gone. Creek is raging and there is no safe way across, we turned back.
  ```

  **Check:** bare, `cr-0429` drops to rank 11, the threshold moves to `0.2789`, and two alerts fire, one on October 2025 mud and one on May 2026 glacier lilies. Not one washout report is flagged. Nothing throws and the table still looks reasonable. Put the prefix back and `cr-0429` returns to rank 2 with three washout reports in the top 5.

  Change the one `prefix` line you added in step 5. Nothing else changes, because every input is built as `prefix + r["text"]`:

  ```python
  # Hint: the step 5 line, with an empty string instead of "classification: "
  prefix = ""
  ```

- **Tune the threshold.** Run with `sigma` at 1.5 (`--sigma 1.5`). **Check:** 4 reports are flagged instead of 7. Neither setting is correct. The value is a business choice about how much review you can afford.

  Change the `sigma` line you added in step 6:

  ```python
  # Hint: the step 6 line, with a new value
  sigma = <new sigma>
  ```

- **Build the baseline before the anomalies arrived.** Build the centroid in step 3 from only the 32 reports dated before 2026-06-18, then score all 40 against it. **Why:** eight of the 40 reports are about the bridge; in the main path they pull the centroid toward themselves and partly hide. **Check:** all eight washout reports (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`, `cr-0464`, `cr-0480`, `cr-0496`) land in the top 10.

  Only the centroid loop changes. First pick out the vectors whose report is older than the cutoff (`zip(vectors, reports)` pairs each vector with its report), then average only those. Put it in place of the `for vector in vectors:` loop from step 5, keep `centroid = normalize(centroid)` after it, and leave the `scored = ...` line alone so all 40 are still scored. `date` is the import you added in step 7:

  ```python
  # Hint: filter with zip, then divide by the smaller count
  baseline = [v for v, r in zip(vectors, reports) if date.fromisoformat(r["date"]) < date.fromisoformat("<cutoff date>")]
  for vector in baseline:
      for k, v in enumerate(vector):
          centroid[k] += v / len(baseline)
  ```

- **Run the other trail.** Point step 1 at `../../data/reports-0042.jsonl` (`--trail 0042` in `complete/`): the 25 `trail-0042` lines from feature 10's stream, dated 2025-05-04 to 2026-07-07. The planted event is bear activity, four reports from 2026-06-25 to 2026-07-02 (`cr-0446`, `cr-0449`, `cr-0453`, `cr-0455`). There are no precomputed vectors for this trail, so this goal needs Ollama. **Check:** `cr-0446` is rank 1 at `0.3067` with a visible gap to second place, the threshold is `0.2292`, and one alert fires on `cr-0446` and `cr-0449`, 2026-06-25 to 2026-06-27.

  Set `trail = "0042"` (the line you added in step 7) and build the file name from it. This replaces the starter's `reports = ...` line:

  ```python
  reports = [json.loads(l) for l in (DATA / f"reports-{trail}.jsonl").read_text().splitlines() if l.strip()]
  ```

  The step 7 ALERT line already uses `trail`. The starter's header line still says `trail-0117`; put `trail` into it too:

  ```python
  # Hint: same idea as the ALERT line in step 7
  print(f"trail-{trail} · {len(reports)} reports · <rest of the header>")
  ```

## What Is in This Folder

- `data/reports-0117.jsonl`: the 40 trail-0117 reports, a slice of feature 10's full stream. Eight describe the footbridge washout (details in step 1).
- `data/reports-0042.jsonl`: the 25 trail-0042 reports from the same stream, with the bear cluster, for the last stretch goal.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for the 40 trail-0117 reports, 768 dimensions each, keyed by `id`, already unit length as `nomic-embed-text` returns them (details in step 2). The starter uses these so steps 1 through 4 run without a model.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this works on this data.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

## One Thing That Will Cost You an Hour if You Miss It

Every vector in `embeddings-0117.json` was produced from `"classification: " + report.text`,
not from the report text alone. `nomic-embed-text` is trained with task prefixes. Leave the
prefix off and nothing fails loudly. You just quietly get worse vectors. On this data that
drops the first washout report from rank 2 to rank 11 and makes the detector fire on October
mud instead. `expected-output.md` shows both rankings side by side.
