# HTTP Walkthrough for 08 Anomaly Detection

The lab steps in [`../F08-lab.md`](../F08-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: embedding requests against `nomic-embed-text`, single and batched, plus one deliberately unprefixed request for comparison.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Steps 1 and 2: Centroid and Ranking

You can skip the model entirely to start: `data/embeddings-0117.json` has all 40 vectors, embedded with the `classification:` prefix. Normalize them, average them into a centroid, normalize that, and rank every report by cosine distance from it (1 minus the dot product of unit vectors).

Check: washout reports rise toward the top, but not cleanly; routine reports are mixed in. Compare the ranking in [`../expected-output.md`](../expected-output.md).

### Lab Step 3, Part One: The Task Prefix

Requests 1 to 3 embed live: one routine report, one washout report, and a batch. Request 4 embeds the same washout report with no prefix so you can see it is a different vector. Embed all 40 with request 3's shape, once bare and once with `classification:` in front of every text, and rank both ways.

Check: bare, the first washout report sits around rank 11; prefixed, rank 2, and the mud reports settle to the bottom.

### Lab Step 3, Part Two: The Alert Rule

Threshold at mean plus one standard deviation of the distances, flag everything above it, sort the flagged reports by date, and alert only when two or more fall within 14 days of each other.

Check: exactly one alert on trail-0117, three genuine washout reports in it, lone outliers ignored. Model calls: 40 embeddings, zero chat completions.

### Stretch

Build the centroid only from reports before the washout window and watch all eight washout reports reach the top 10, or run `data/reports-0042.jsonl` and catch the bear-activity spike.
