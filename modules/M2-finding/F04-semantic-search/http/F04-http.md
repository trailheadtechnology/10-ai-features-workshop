# HTTP Walkthrough for 04 Semantic Search

The lab steps in [`../F04-lab.md`](../F04-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: three requests against Ollama (`nomic-embed-text`). This endpoint is the only network call the lab makes; the ranking is arithmetic you write yourself.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Requests 1 and 2, Embed the Catalog

Request 1 embeds one sentence: look at the response once, it is 768 floats. Request 2 embeds a batch (an array in `input`), which is how you embed all 30 descriptions from `data/trails-slice.json` in one call. Do that in your language and keep the vectors in a dictionary keyed by trail id.

### Lab Step 2: Request 3, Embed the Query, Then Rank

Request 3 embeds the demo query through the same model and the same endpoint. Now write cosine similarity (a dozen lines; pseudocode is in [`../expected-output.md`](../expected-output.md)), score every trail against the query vector, and print the top 5 with scores.

Check: the gentle shaded waterfall trails at the top, scores around 0.77.

### Lab Step 3: The Other Queries

`data/queries.json` has all three test queries and their expected top hits. Change the `input` in request 3 and re-rank. Sit with the kids query: Taft Point at 0.4876 is a perfect topical match to a cliff edge.

Check: the expected trail is in your top 3 for each query.

### Stretch

Filter on the metadata you already have (`features` contains `dog-friendly`, `distance_mi < 6`) before ranking, or blend the keyword-hit count into the score.
