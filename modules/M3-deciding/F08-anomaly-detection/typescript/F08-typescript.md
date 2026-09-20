<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 08: Anomaly Detection (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F08-dotnet.md), [Python](../python/F08-python.md). Lab overview: [F08-lab.md](../F08-lab.md).*

**The User Problem:** Trail-condition reports trickle into Trailhead Guides all season, about 500 of them across 200 trails. Almost all say some version of "muddy in spots, otherwise fine." Then over one week, three separate hikers report a washed-out bridge on the same trail, and a fourth mentions aggressive bear activity two trails over. Nobody notices, because nobody reads 500 routine reports. The park finds out about the bridge from a one-star review a month later.

*A Challenge lab. Do it if you finished [Module 3](../../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** find the condition reports for one trail that do not look like the rest, using distance from a centroid. Then raise one alert when several of them land close together in time.
- **Input:** `data/reports-0117.jsonl`, 40 reports for trail-0117 with the planted washout cluster; `data/embeddings-0117.json`, their `nomic-embed-text` vectors, unnormalized, `classification: ` prefixed; `data/reports-0042.jsonl`, trail-0042, for a stretch goal.
- **How:** embed the reports through your track's embeddings client against local Ollama. The centroid, distances, threshold, and alert rule are plain arithmetic you write yourself.
- **Model:** `nomic-embed-text`, local. Every track's `starter/` runs offline on the precomputed vectors. Only step 5 onward needs Ollama.

## The Concept

This feature is barely an AI feature: it's embeddings plus arithmetic. Embed every condition report for a trail, and the routine ones ("muddy," "buggy," "fine") cluster together in vector space. Average them and you get a centroid, the mathematical center of "normal" for that trail. A report's distance from that centroid is an anomaly score. "Bridge washed out" sits farther from the mud cluster than the mud reports sit from each other, so it rises without a large model or any training. Cosine distance and a threshold do most of the job.

The word "most" is doing real work in that sentence, and this feature is honest about it. Ranking single reports by distance is noisy: routine reports about parking or wildflowers can outrank a genuine hazard, and when 8 of 40 reports describe the same washout, the anomalies drag the centroid toward themselves and partially hide. Two things rescue it, and both are the actual lesson. First, embedding models have contracts: `nomic-embed-text` is trained with task prefixes, and embedding `"classification: " + text` instead of the bare text moves the first washout report from rank 11 to rank 2. Second, the alert rule beats the ranking. Requiring two flagged reports within a two-week window fires exactly one alert on this trail, all three of its reports genuine, zero false positives. One outlier might be a rambling hiker; several outliers in a week that also sit near each other are an event.

The pattern generalizes to any stream of routine text: support tickets, log messages, form submissions, review streams. Define normal from the data itself, and let distance flag what deserves human eyes. It also pairs naturally with feature 07: classification handles the categories you knew to define, and anomaly detection catches the things you didn't.

Every step below is one thing to make the program do. The `starter/` already does steps 1 through 4 using the precomputed vectors. Read those four steps next to the starter code and find each one. Then edit the starter until it does steps 5 through 7. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts:

- `starter/index.ts`: no model, no network. Loads the precomputed vectors from `../../data/embeddings-0117.json`, averages them into a centroid, ranks every report by cosine distance from it. Runs with Ollama down.
- `complete/index.ts`: the finished demo as shown on stage. Embeds live with the `classification:` task prefix, derives the threshold from the corpus (mean plus sigma standard deviations), and applies the alert rule: two or more flagged reports within a 14-day window. One alert fires, three genuine washout reports in it.

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. Run everything from the `typescript/` folder, where `package.json` lives:

```bash
npm install                        # once
npm run starter                    # precomputed vectors, no model
npm run complete                   # trail-0117, live embeddings, distance table + cluster alerts
npm run complete -- --raw          # same trail with the task prefix removed
npm run complete -- --trail 0042   # the bear-activity trail
npm run complete -- --sigma 1.5
npm run complete -- --window 30
```

The starter takes no flags. `--raw`, `--sigma`, `--window`, and `--trail` exist only in `complete/`, so hard-code the value or add the same argument parsing.

### Step 0: Run the starter and read the ranking

**Do:** run `starter/` as it is. It needs no model and no network.

```bash
npm run starter
```

**Check:** a 40-row table, one report per row, sorted by distance with the largest first. The top row is `cr-0496` at `0.2561` and the bottom row is `cr-0014` at `0.1484`. The washout reports rise toward the top, but parking and wildflower reports are mixed in with them. That is what the technique does out of the box.

### Step 1: Load the reports

**Do:**
1. Open `../../data/reports-0117.jsonl`: `id`, `trail_id`, `date`, `text` per line.
2. Review how the existing code from the starter project reads the file line by line, parses each line, and keeps the results in a list in file order.

   `DATA` is the `data/` folder two levels above `index.ts`, and `Report` describes the fields of one line:

   ```typescript
   const DATA = resolve(import.meta.dirname, "../../data");
   type Report = { id: string; trail_id: string; date: string; text: string };
   ```

   The `reports` line reads the whole file as text, splits it into lines, drops blank ones with `filter`, and turns each line into an object with `JSON.parse`, so `r.id` and `r.text` work:

   ```typescript
   const reports: Report[] = readFileSync(resolve(DATA, "reports-0117.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
   ```

   To see the Check for yourself, add a temporary line below it and delete it afterward:

   ```typescript
   // Hint: temporary, prints the count and the first report
   console.log(reports.length, reports[0].id, reports[0].date);
   ```

**Why:** the file is the 40 reports for trail-0117, dated 2025-05-04 to 2026-07-22, copied out of the workshop's 500-report stream in feature 10's data. Eight of the 40 are the planted washout: five from 2026-06-18 to 2026-06-24 (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`). The other 32 are mud, ice, bugs, wildflowers, parking, and blowdown.

**Check:** the list has 40 entries. The first `id` is `cr-0009`, dated 2025-05-04.

### Step 2: Load the precomputed vectors and normalize them

**Do:**
1. Review how the existing code from the starter project opens `../../data/embeddings-0117.json`. Its `embeddings` field maps report `id` to an array of 768 numbers.

   `JSON.parse` turns the file into an object, and `.embeddings` pulls out the `id`-to-numbers record:

   ```typescript
   const raw: Record<string, number[]> = JSON.parse(readFileSync(resolve(DATA, "embeddings-0117.json"), "utf8")).embeddings;
   ```

2. Review how it normalizes a vector: find its length (square every component, sum, square root), then divide every component by that length.

   It is a function near the top of `index.ts`. `reduce` adds up `v * v` for every component, starting from `0`, and `map` builds a new array with every component divided by the length:

   ```typescript
   function normalize(vector: number[]): number[] {
     const length = Math.sqrt(vector.reduce((s, v) => s + v * v, 0));
     return vector.map((v) => v / length);
   }
   ```

3. Review how it normalizes every vector as it loads it, storing results in a dictionary keyed by `id`.

   `Object.entries(raw)` gives `[id, numbers]` pairs, `map` keeps each `id` and normalizes its numbers, and `Object.fromEntries` turns the pairs back into a record keyed by `id`:

   ```typescript
   const vectors = Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, normalize(v)]));
   ```

   To see the Check for yourself, add a temporary line below it (a unit vector's length prints as `1` or very close to it):

   ```typescript
   // Hint: temporary, prints the key count, the size, and the length of one vector
   const first = Object.values(vectors)[0];
   console.log(Object.keys(vectors).length, first.length, Math.sqrt(first.reduce((s, v) => s + v * v, 0)));
   ```

**Why:** the vectors were captured once from local Ollama with `nomic-embed-text`, from `"classification: " + text`, stored raw and unnormalized.

**Check:** 40 keys, each holding 768 numbers, and every stored vector now has length 1.0.

### Step 3: Build the centroid

**Do:**
1. Review how the existing code from the starter project makes an array of 768 zeros. This is the centroid.

   `all` is every vector as an array, `dimensions` is the length of any one of them (768), and `new Array<number>(dimensions).fill(0)` is an array of that many zeros:

   ```typescript
   const all = Object.values(vectors);
   const dimensions = all[0].length;
   const centroid = new Array<number>(dimensions).fill(0);
   ```

2. Review how it loops over the 40 normalized vectors; for each vector and position `i`, it adds `vector[i] / 40` to `centroid[i]`.

   It happens in one line: the outer `for` visits each vector, the inner `for` visits each position `i`, and `all.length` is 40:

   ```typescript
   for (const vector of all) for (let i = 0; i < dimensions; i++) centroid[i] += vector[i] / all.length;
   ```

3. Review how it normalizes the centroid with the same function from step 2.

   `centroid` is a `const`, so the normalized result gets a new name, `center`, and step 4 measures against `center`:

   ```typescript
   const center = normalize(centroid);
   ```

**Check:** one 768-dimension unit vector for trail-0117. If your step 4 distances are off in the third decimal from `expected-output.md`, you skipped one of the two normalizations.

### Step 4: Score every report and sort

**Do:**
1. Review how the existing code from the starter project computes the cosine distance between two vectors: 1 minus the dot product (this only works because both vectors are unit length).

   It is a function below `normalize`. The loop multiplies matching positions of `a` and `b` and adds them up:

   ```typescript
   function cosineDistance(a: number[], b: number[]): number {
     let dot = 0;
     for (let i = 0; i < a.length; i++) dot += a[i] * b[i];
     return 1 - dot;
   }
   ```

2. Review how it looks up each report's vector by `id` and computes its distance from the centroid.
3. Review how it sorts by distance, largest first.

   Items 2 and 3 happen in one statement. `map` turns each report into a `{ report, distance }` object, looking up `vectors[r.id]`, and `sort((a, b) => b.distance - a.distance)` puts the largest distance first:

   ```typescript
   const scored = reports
     .map((r) => ({ report: r, distance: cosineDistance(vectors[r.id], center) }))
     .sort((a, b) => b.distance - a.distance);
   ```

4. Review how it prints one line per report: distance to four decimals, `id`, `date`, and `text` cut to 62 characters.

   It happens at the bottom of `index.ts`. `toFixed(4)` is four decimals, and `for (const { report, distance } of scored)` takes each object apart into its two fields:

   ```typescript
   console.log(`trail-0117 · ${reports.length} reports · ${dimensions}-dim nomic-embed-text vectors\n`);
   console.log("  dist    id       date        report");
   for (const { report, distance } of scored) {
     console.log(`  ${distance.toFixed(4)}  ${report.id}  ${report.date}  ${truncate(report.text, 62)}`);
   }
   ```

   The cut is a one-line arrow function below `cosineDistance`:

   ```typescript
   const truncate = (text: string, n: number) => (text.length <= n ? text : text.slice(0, n - 1) + "…");
   ```

**Check:** `cr-0496` is rank 1 at `0.2561` and `cr-0429` is rank 2 at `0.2399`. Routine reports are mixed in with the washout reports and there is no gap anywhere in the distances. Washout reports ranked in the 30s mean the prefix is missing.

### Step 5: Embed live with the task prefix

**Do:**
1. Delete the code that loads `../../data/embeddings-0117.json`.
2. Build a list of strings, one per report in list order: `"classification: "` followed by the report's `text`.
3. Embed all 40 strings in **one call** with model `nomic-embed-text`. Ollama listens on `http://localhost:11434`.

   The starter imports no AI package. The client is the `openai` package (already installed by `npm install`) pointed at Ollama. Add the import directly below the starter's `import { resolve } from "node:path";` line:

   ```typescript
   import OpenAI from "openai";
   ```

   Add the client directly above the `const DATA = ...` line. `apiKey` can be any text; Ollama ignores it, but the package refuses to start without one:

   ```typescript
   const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
   ```

   Then replace the two lines that start `const raw` and `const vectors` (they read `embeddings-0117.json`) with these three. `reports.map((r) => prefix + r.text)` builds the 40 prefixed strings from item 2, in report order, and `client.embeddings.create` sends them all in one request. Top-level `await` (an `await` outside any function) works here because `package.json` sets `"type": "module"`:

   ```typescript
   let prefix = "classification: ";
   const response = await client.embeddings.create({ model: "nomic-embed-text", input: reports.map((r) => prefix + r.text) });
   const vectors = response.data.map((d) => normalize(d.embedding));
   ```

4. Normalize each returned vector and keep it at the same index as its report. Replace the `id`-keyed dictionary with a plain list in report order, so anything that indexed by `id` now indexes by position.

   `response.data` comes back in input order, so `vectors[i]` belongs to `reports[i]`. `vectors` is now an array, not a record keyed by `id`, so `Object.values(vectors)` and `vectors[r.id]` no longer fit. Replace the starter's lines from `const all = Object.values(vectors);` down to the end of the `const scored = ...` statement (its `.sort(...)` line) with these. The `(r, i)` in `map` gives each report its position `i`, so `vectors[i]` is that report's vector:

   ```typescript
   const dimensions = vectors[0].length;
   const centroid = new Array<number>(dimensions).fill(0);
   for (const vector of vectors) for (let i = 0; i < dimensions; i++) centroid[i] += vector[i] / vectors.length;
   const center = normalize(centroid);

   const scored = reports
     .map((r, i) => ({ report: r, distance: cosineDistance(vectors[i], center) }))
     .sort((a, b) => b.distance - a.distance);
   ```

5. Rebuild the centroid from these 40 vectors (step 3), then score and sort again (step 4), looking up each vector by index instead of by `id`. Run it with Ollama up:

   ```bash
   npm run starter
   ```

**Check:** the same 40-row table as step 4, matching `expected-output.md` to four decimals. Your program made 40 embeddings and 0 chat completions. Distances that differ in the third decimal mean a normalization is missing.

### Step 6: Derive the threshold

**Do:**
1. Compute the mean of the 40 distances.
2. Compute the standard deviation: subtract the mean from each distance, square, average the 40 squares, take the square root.
3. Set `threshold = mean + sigma * sd`, `sigma` defaulting to 1.0 (`--sigma` in `complete/`).
4. Print `mean distance <mean> · sd <sd> · threshold mean+1sd = <threshold>`, all to four decimals.

   Put `sigma` near the top of `index.ts`, directly below the `const DATA = ...` line:

   ```typescript
   let sigma = 1.0;
   ```

   Directly below the `const scored = ...` statement (after its `.sort(...)` line), compute the mean and the population standard deviation of the 40 distances. `reduce((s, x) => s + x.distance, 0)` adds up every distance starting from `0`, `** 2` squares, and `Math.sqrt` is the square root:

   ```typescript
   const mean = scored.reduce((s, x) => s + x.distance, 0) / scored.length;
   const deviation = Math.sqrt(scored.reduce((s, x) => s + (x.distance - mean) ** 2, 0) / scored.length);
   const threshold = mean + sigma * deviation;
   ```

   Print it between the starter's `` console.log(`trail-0117 · ...`) `` header line and its `console.log("  dist    id       date        report");` line, so it lands above the table. `${sigma}` prints `1.0` as `1`:

   ```typescript
   console.log(`mean distance ${mean.toFixed(4)} · sd ${deviation.toFixed(4)} · threshold mean+${sigma}sd = ${threshold.toFixed(4)}\n`);
   ```

   The starter's header line ends in `\n`, so a blank line prints above this one. That is fine; only the numbers matter.

5. In the table, put a `!` in place of the first leading space of every row above the threshold, so columns stay aligned.

   Replace the starter's three-line `for (const { report, distance } of scored) { ... }` loop at the bottom of `index.ts` with this one. `s.distance > threshold ? " !" : "  "` picks one of the two 2-character prefixes:

   ```typescript
   for (const s of scored) {
     console.log(`${s.distance > threshold ? " !" : "  "}${s.distance.toFixed(4)}  ${s.report.id}  ${s.report.date}  ${truncate(s.report.text, 62)}`);
   }
   ```

**Check:** mean `0.1935`, sd `0.0274`, threshold `0.2208`. Seven rows carry the `!`. The last marked row is `cr-0329` at `0.2219`, and the first unmarked row is `cr-0125` at `0.2206`.

### Step 7: Group the flagged reports into alerts

**Do:**
1. Take every report above the threshold and sort by `date`, oldest first.

   Everything in this step goes at the very bottom of `index.ts`, after the table loop. `filter` keeps only the objects above the threshold, and `a.report.date.localeCompare(b.report.date)` sorts them by the report's `date` as text (ISO text like `2026-06-18` sorts oldest first):

   ```typescript
   const flagged = scored.filter((s) => s.distance > threshold).sort((a, b) => a.report.date.localeCompare(b.report.date));
   ```

2. Print `7 of 40 reports above threshold. Clustering them within 14 days:` (window defaults to 14 days, `--window` in `complete/`).

   Put `window` near the top of `index.ts`, below the `let sigma = 1.0;` line from step 6:

   ```typescript
   let window = 14;
   ```

   Then, below the `const flagged = ...` line:

   ```typescript
   console.log(`\n${flagged.length} of ${scored.length} reports above threshold. Clustering them within ${window} days:\n`);
   ```

3. Walk the sorted list: start a group at position `i`, move `j` forward while the date at `j` minus the date at `j-1` is 14 days or less. The group is positions `i` up to but not including `j`. `date` is a string; parse it to a real date before subtracting (ISO strings sort correctly as text but do not subtract).

   `Date.parse` turns the ISO string into milliseconds since 1970, so the gap in days is the difference divided by `86_400_000` (the milliseconds in one day). Add this helper directly below the starter's `const truncate = ...` line:

   ```typescript
   const days = (a: string, b: string) => Math.round((Date.parse(a) - Date.parse(b)) / 86_400_000);
   ```

   Then, below the print from item 2, start the counter and the outer loop. The `for` has no step part (nothing after the second `;`) because item 6 moves `i` forward. `flagged[j].report.date` is the `date` of the report at position `j`, and `flagged.slice(i, j)` is a new array of positions `i` up to but not including `j`:

   ```typescript
   let alerts = 0;
   for (let i = 0; i < flagged.length;) {
     let j = i + 1;
     while (j < flagged.length && days(flagged[j].report.date, flagged[j - 1].report.date) <= window) j++;
     const group = flagged.slice(i, j);
   ```

   Items 4 to 6 go inside this loop, below `const group = ...`. The loop's closing `}` comes in item 6.

4. If the group holds 2+ reports, print `ALERT trail-0117: <count> anomalous reports between <first date> and <last date>`, then one line per report (`id`, `date`, text cut to 70 characters).

   `complete/` names the trail in a variable so the last stretch goal can change it. Put it near the top of `index.ts`, directly below the `const DATA = ...` line:

   ```typescript
   let trail = "0117";
   ```

   Then, below `const group = ...`, inside the loop. `group[0]` is the first item and `group[group.length - 1]` the last:

   ```typescript
     if (group.length >= 2) {
       alerts++;
       console.log(`  ALERT trail-${trail}: ${group.length} anomalous reports between ${group[0].report.date} and ${group[group.length - 1].report.date}`);
       for (const s of group) console.log(`        ${s.report.id} ${s.report.date}  ${truncate(s.report.text, 70)}`);
   ```

5. If the group holds 1 report, print `(ignored) <id> <date> is a lone outlier, not an event`.

   Directly below the `for (const s of group)` line. The `} else {` closes the `if` block from item 4:

   ```typescript
     } else {
       console.log(`  (ignored) ${group[0].report.id} ${group[0].report.date} is a lone outlier, not an event`);
     }
   ```

6. Set `i = j` and repeat until the list is used up.

   Last line inside the loop, below the `}` that closes `else`, followed by the `}` that closes the `for` from item 3. Without `i = j;` the loop never ends:

   ```typescript
     i = j;
   }
   ```

7. Print `<N> alert(s). Model calls: 40 embeddings, 0 chat completions.`

   After the loop, as the last line of `index.ts`:

   ```typescript
   console.log(`\n${alerts} alert(s). Model calls: ${reports.length} embeddings, 0 chat completions.`);
   ```

Put together, the bottom of `index.ts` is the `flagged` line, the item 2 print, then the loop, with the final print below it. Run:

```bash
npm run starter
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

  Change the one `prefix` line you added in step 5. Nothing else changes, because every input is built as `prefix + r.text`:

  ```typescript
  // Hint: the step 5 line, with an empty string instead of "classification: "
  let prefix = "";
  ```

- **Tune the threshold.** Run with `sigma` at 1.5 (`--sigma 1.5`). **Check:** 4 reports are flagged instead of 7. Neither setting is correct. The value is a business choice about how much review you can afford.

  Change the `sigma` line you added in step 6:

  ```typescript
  // Hint: the step 6 line, with a new value
  let sigma = <new sigma>;
  ```

- **Build the baseline before the anomalies arrived.** Build the centroid in step 3 from only the 32 reports dated before 2026-06-18, then score all 40 against it. **Why:** eight of the 40 reports are about the bridge; in the main path they pull the centroid toward themselves and partly hide. **Check:** all eight washout reports (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`, `cr-0464`, `cr-0480`, `cr-0496`) land in the top 10.

  Only the centroid loop changes. First pick out the vectors whose report is older than the cutoff (`filter((v, index) => ...)` gives each vector its position, and `vectors[index]` belongs to `reports[index]`), then average only those. Put it in place of the one-line `for (const vector of vectors) ...` loop from step 5, keep `const center = normalize(centroid);` after it, and leave the `const scored` statement alone so all 40 are still scored:

  ```typescript
  // Hint: filter by position, then divide by the smaller count
  const baseline = vectors.filter((v, index) => Date.parse(reports[index].date) < Date.parse("<cutoff date>"));
  for (const vector of baseline) for (let i = 0; i < dimensions; i++) centroid[i] += vector[i] / baseline.length;
  ```

- **Run the other trail.** Point step 1 at `../../data/reports-0042.jsonl` (`--trail 0042` in `complete/`): the 25 `trail-0042` lines from feature 10's stream, dated 2025-05-04 to 2026-07-07. The planted event is bear activity, four reports from 2026-06-25 to 2026-07-02 (`cr-0446`, `cr-0449`, `cr-0453`, `cr-0455`). There are no precomputed vectors for this trail, so this goal needs Ollama. **Check:** `cr-0446` is rank 1 at `0.3067` with a visible gap to second place, the threshold is `0.2292`, and one alert fires on `cr-0446` and `cr-0449`, 2026-06-25 to 2026-06-27.

  Set `let trail = "0042";` (the line you added in step 7, which already sits above `const reports`) and build the file name from it. This replaces the starter's `const reports = ...` line:

  ```typescript
  const reports: Report[] = readFileSync(resolve(DATA, `reports-${trail}.jsonl`), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
  ```

  The step 7 ALERT line already uses `trail`. The starter's header line still says `trail-0117`; put `trail` into it too:

  ```typescript
  // Hint: same idea as the ALERT line in step 7
  console.log(`trail-${trail} · ${reports.length} reports · <rest of the header>`);
  ```

## What Is in This Folder

- `data/reports-0117.jsonl`: the 40 trail-0117 reports, a slice of feature 10's full stream. Eight describe the footbridge washout (details in step 1).
- `data/reports-0042.jsonl`: the 25 trail-0042 reports from the same stream, with the bear cluster, for the last stretch goal.
- `data/embeddings-0117.json`: real `nomic-embed-text` vectors for the 40 trail-0117 reports, 768 dimensions each, keyed by `id`, unnormalized (details in step 2). The starter uses these so steps 1 through 4 run without a model.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this works on this data.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

## One Thing That Will Cost You an Hour if You Miss It

Every vector in `embeddings-0117.json` was produced from `"classification: " + report.text`,
not from the report text alone. `nomic-embed-text` is trained with task prefixes. Leave the
prefix off and nothing fails loudly. You just quietly get worse vectors. On this data that
drops the first washout report from rank 2 to rank 11 and makes the detector fire on October
mud instead. `expected-output.md` shows both rankings side by side.
