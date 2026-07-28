# .NET demo for 08 Anomaly Detection

Two console projects, both named `Anomaly`, matching the demo outline in
[../README.md](../README.md):

- `starter/`: the demo's starting point. No packages, no network, no model. It loads the precomputed vectors from `../../lab/embeddings-0117.json`, averages them into a centroid, and prints every report's cosine distance from it. About forty lines, all of it arithmetic, and it runs with Ollama switched off.
- `complete/`: the finished demo as shown on stage. Built on Microsoft.Extensions.AI over OllamaSharp, it embeds all 40 reports live with `nomic-embed-text`, prints the same table with a data-derived threshold applied, and turns the flagged reports into alerts.

From `starter/`:

```bash
dotnet run                    # trail-0117 distance ranking, offline
```

From `complete/`:

```bash
dotnet run                    # trail-0117: live embeddings, ranking, threshold, alerts
dotnet run -- --raw           # same run without nomic's task prefix (the failure mode)
dotnet run -- --trail 0042    # stretch goal: the bear-activity spike on the other trail
dotnet run -- --sigma 1.5     # tighter threshold (default is mean + 1 sd)
dotnet run -- --window 30     # wider clustering window in days (default 14)
```

Both projects print the same distances for trail-0117, because
`embeddings-0117.json` holds exactly the vectors `complete` asks Ollama for. If
they disagree, one of you skipped the normalization step.

## What the demo is actually showing

`--raw` is worth the thirty seconds it takes. `nomic-embed-text` expects a task
prefix on its input, `complete` sends `"classification: "`, and dropping it moves
the washed-out-bridge report from rank 2 to rank 11 and makes the alert rule fire
on October mud. Nothing throws, nothing warns, the table still looks reasonable.
That is the shape most embedding bugs have.

The other beat is the alert rule. The distance ranking on its own is decent and not
great: it puts four washout reports in the top six but also flags a parking
complaint and a wildflower report. Requiring two flagged reports inside 14 days
before raising an alert drops every one of those false positives, because they are
lone oddities on quiet weeks, and leaves exactly one alert holding three genuine
bridge reports. One weird report is a rambling hiker. Three in a week is an event.

See [../lab/expected-output.md](../lab/expected-output.md) for real output from
every command above, including where this technique is weaker than the module card
implies.
