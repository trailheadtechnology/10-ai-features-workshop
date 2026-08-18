# .NET Demo for 08 Anomaly Detection

Two console projects, both named `Anomaly`, matching the demo script in
[docs/slides/outlines/M3-deciding.md](../../../../docs/slides/outlines/M3-deciding.md):

- `starter/`: the demo's starting point. No packages, no network, no model. It loads the precomputed vectors from `../../data/embeddings-0117.json`, averages them into a centroid, and prints every report's cosine distance from it. About forty lines, all of it arithmetic, and it runs with Ollama switched off.
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

## What the Demo Is Actually Showing

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

See [../expected-output.md](../expected-output.md) for real output from
every command above, including where this technique is weaker than the feature card
implies.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F08-lab.md`](../F08-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: Centroid Distance from Precomputed Vectors

Lab steps 1 and 2 are already in the starter, and it needs no model: it loads the 40 vectors in `../data/embeddings-0117.json` (embedded with the `classification:` prefix), averages them into a centroid, and ranks every report by cosine distance from it. Read the top of the list.

Run:

```bash
dotnet run
```

Check: Washout reports rise toward the top, but not cleanly: routine reports about parking or wildflowers are mixed in. Compare the ranking in `../expected-output.md`. That is what the technique does out of the box.

### Step 2: Embed Live, First Without the Task Prefix (lab step 3, first half)

Replace the precomputed file with a live embedding call, and deliberately embed the bare text. `nomic-embed-text` expects a task prefix on every input; without one it still returns a well-formed vector, so nothing throws, but the vectors land off-distribution and the ranking degrades. Watch where the first washout report lands.

```csharp
IEmbeddingGenerator<string, Embedding<float>> embedder =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");
var prefix = "";                       // step 2: bare text. step 3: "classification: "
var generated = await embedder.GenerateAsync(reports.Select(r => prefix + r.Text));
var vectors = generated.Select(e => Normalize(e.Vector.ToArray())).ToList();
```

Run:

```bash
dotnet run
```

Check: The first washout report sits around rank 11 (`complete/ --raw` reproduces this). Then set the prefix to `"classification: "` and re-run: it jumps to rank 2 and the mud reports settle to the bottom. Reading the model card is engineering work.

### Step 3: Derive a Threshold and Add the Alert Rule (lab step 3, second half, and where the feature actually lives)

A threshold from the corpus (mean plus one standard deviation) flags outliers; the alert rule requires two or more flagged reports within 14 days of each other. One outlier is a rambling hiker; several in a week that also sit near each other is an event.

```csharp
var mean = scored.Average(s => s.Distance);
var sd = Math.Sqrt(scored.Average(s => Math.Pow(s.Distance - mean, 2)));
var threshold = mean + sigma * sd;

var flagged = scored.Where(s => s.Distance > threshold).OrderBy(s => s.Report.Date).ToList();
for (var i = 0; i < flagged.Count;)
{
    var j = i + 1;
    while (j < flagged.Count && (Date(flagged[j].Report) - Date(flagged[j - 1].Report)).Days <= window) j++;
    var group = flagged[i..j];
    if (group.Count >= 2)
        Console.WriteLine($"  ALERT trail-{trail}: {group.Count} anomalous reports between {group[0].Report.Date} and {group[^1].Report.Date}");
    else
        Console.WriteLine($"  (ignored) {group[0].Report.Id} is a lone outlier, not an event");
    i = j;
}
```

Run:

```bash
dotnet run
```

Check: Exactly one alert on trail-0117, three genuine washout reports in it (`cr-0429`, `cr-0436`, `cr-0464`), lone outliers ignored. Count the model calls: 40 embeddings, zero chat completions. Stretch: build the centroid only from reports before the washout window and watch all eight washout reports reach the top 10; or run `--trail 0042` and catch the bear-activity spike, which separates more sharply.
