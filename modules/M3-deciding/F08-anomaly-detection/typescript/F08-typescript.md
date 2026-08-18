# TypeScript Demo for 08 Anomaly Detection

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: no model, no network. Loads the precomputed vectors from `../data/embeddings-0117.json`, averages them into a centroid, ranks every report by cosine distance from it. Runs with Ollama down.
- `complete/index.ts`: the finished demo as shown on stage. Embeds live with the `classification:` task prefix, derives the threshold from the corpus (mean plus sigma standard deviations), and applies the alert rule: two or more flagged reports within a 14-day window. One alert fires, three genuine washout reports in it.

Setup once (`npm install` in this directory), then:

```bash
npm run starter                    # precomputed vectors, no model
npm run complete                   # trail-0117, live embeddings, distance table + cluster alerts
npm run complete -- --raw          # same trail with the task prefix removed
npm run complete -- --trail 0042   # the bear-activity trail
npm run complete -- --sigma 1.5
npm run complete -- --window 30
```

Embeddings are the only model calls; everything after them is arithmetic. The recorded ranking and the single alert are in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the TypeScript equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F08-lab.md`](../F08-lab.md), done in TypeScript: start from `starter/index.ts` and end where `complete/index.ts` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run with `npm run starter` from the `typescript/` directory; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: Centroid Distance from Precomputed Vectors

Lab steps 1 and 2 are already in the starter, and it needs no model: it loads the 40 vectors in `../data/embeddings-0117.json` (embedded with the `classification:` prefix), averages them into a centroid, and ranks every report by cosine distance from it. Read the top of the list.

Run:

```bash
npm run starter
```

Check: Washout reports rise toward the top, but not cleanly: routine reports about parking or wildflowers are mixed in. Compare the ranking in `../expected-output.md`. That is what the technique does out of the box.

### Step 2: Embed Live, First Without the Task Prefix (lab step 3, first half)

Replace the precomputed file with a live embedding call, and deliberately embed the bare text. `nomic-embed-text` expects a task prefix on every input; without one it still returns a well-formed vector, so nothing throws, but the vectors land off-distribution and the ranking degrades. Watch where the first washout report lands.

```typescript
const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
let prefix = "";                       // step 2: bare text. step 3: "classification: "
const response = await client.embeddings.create({ model: "nomic-embed-text", input: reports.map((r) => prefix + r.text) });
const vectors = response.data.map((d) => normalize(d.embedding));
```

Run:

```bash
npm run starter
```

Check: The first washout report sits around rank 11 (`complete/ --raw` reproduces this). Then set the prefix to `"classification: "` and re-run: it jumps to rank 2 and the mud reports settle to the bottom. Reading the model card is engineering work.

### Step 3: Derive a Threshold and Add the Alert Rule (lab step 3, second half, and where the feature actually lives)

A threshold from the corpus (mean plus one standard deviation) flags outliers; the alert rule requires two or more flagged reports within 14 days of each other. One outlier is a rambling hiker; several in a week that also sit near each other is an event.

```typescript
const mean = scored.reduce((s, x) => s + x.distance, 0) / scored.length;
const sd = Math.sqrt(scored.reduce((s, x) => s + (x.distance - mean) ** 2, 0) / scored.length);
const threshold = mean + sigma * sd;

const flagged = scored.filter((s) => s.distance > threshold).sort((a, b) => a.report.date.localeCompare(b.report.date));
for (let i = 0; i < flagged.length;) {
  let j = i + 1;
  while (j < flagged.length && days(flagged[j].report.date, flagged[j - 1].report.date) <= window) j++;
  const group = flagged.slice(i, j);
  if (group.length >= 2) console.log(`  ALERT trail-${trail}: ${group.length} anomalous reports between ${group[0].report.date} and ${group[group.length - 1].report.date}`);
  else console.log(`  (ignored) ${group[0].report.id} is a lone outlier, not an event`);
  i = j;
}
```

Run:

```bash
npm run starter
```

Check: Exactly one alert on trail-0117, three genuine washout reports in it (`cr-0429`, `cr-0436`, `cr-0464`), lone outliers ignored. Count the model calls: 40 embeddings, zero chat completions. Stretch: build the centroid only from reports before the washout window and watch all eight washout reports reach the top 10; or run `--trail 0042` and catch the bear-activity spike, which separates more sharply.
