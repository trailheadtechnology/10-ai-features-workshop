using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Ranks trail reports by distance from the trail's own baseline and alerts when
// several outliers land close together in time. Needs Ollama running with
// nomic-embed-text pulled.
//
//   dotnet run                    trail-0117, live embeddings, distance table + cluster alerts
//   dotnet run -- --raw           same trail with the task prefix removed
//   dotnet run -- --trail 0042    the other trail in the data folder
//   dotnet run -- --sigma 1.5     tighter threshold
//   dotnet run -- --window 30     wider clustering window, in days
//
// Embeddings are the only model calls. Everything after them is arithmetic, so
// the cost of this feature is one embedding per report and nothing per query.

var trail = "0117";
var sigma = 1.0;
var window = 14;
// nomic-embed-text is trained with task prefixes (search_query:, search_document:,
// clustering:, classification:) and expects one on every input. Embedding bare text
// still returns a well-formed vector, which is why --raw fails silently rather than
// throwing, but the vectors land off-distribution and the ranking degrades badly.
// Do not drop this prefix, and if you change it, change it for every input in the
// same corpus: vectors embedded under different prefixes are not comparable.
var prefix = "classification: ";
for (var i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--trail": trail = args[++i]; break;
        case "--sigma": sigma = double.Parse(args[++i]); break;
        case "--window": window = int.Parse(args[++i]); break;
        case "--raw": prefix = ""; break;
        default: throw new ArgumentException($"unknown argument: {args[i]}");
    }
}

var json = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
var reports = File.ReadLines($"../../data/reports-{trail}.jsonl")
    .Select(line => JsonSerializer.Deserialize<Report>(line, json)!)
    .ToList();

IEmbeddingGenerator<string, Embedding<float>> embedder =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

var generated = await embedder.GenerateAsync(reports.Select(r => prefix + r.Text));
var vectors = generated.Select(e => Normalize(e.Vector.ToArray())).ToList();

// The centroid is this trail's definition of normal, and it is built from the
// same reports it is about to judge. Anomalies that appear often enough pull the
// centroid toward themselves and stop looking anomalous, so a long-running
// detector should rebuild this from a trailing window rather than the full history.
var dimensions = vectors[0].Length;
var centroid = new float[dimensions];
foreach (var vector in vectors)
    for (var i = 0; i < dimensions; i++)
        centroid[i] += vector[i] / vectors.Count;
centroid = Normalize(centroid);

var scored = reports
    .Select((r, i) => new Scored(r, CosineDistance(vectors[i], centroid)))
    .OrderByDescending(s => s.Distance)
    .ToList();

// The threshold is derived from this corpus rather than hard-coded, so it travels
// to a trail with a different spread of distances. It is still a business choice,
// not a boundary the data hands you: sigma decides how much review you are willing
// to pay for, and there is no value of it that separates incidents from oddities.
var mean = scored.Average(s => s.Distance);
var deviation = Math.Sqrt(scored.Average(s => Math.Pow(s.Distance - mean, 2)));
var threshold = mean + sigma * deviation;

Console.WriteLine($"trail-{trail} · {reports.Count} reports · nomic-embed-text ({dimensions} dims)"
    + (prefix.Length == 0 ? " · NO task prefix" : $" · prefix \"{prefix.Trim()}\""));
Console.WriteLine($"mean distance {mean:F4} · sd {deviation:F4} · threshold mean+{sigma:0.#}sd = {threshold:F4}\n");

Console.WriteLine("  dist    id       date        report");
foreach (var s in scored)
    Console.WriteLine($"{(s.Distance > threshold ? " !" : "  ")}{s.Distance:F4}  {s.Report.Id}  {s.Report.Date}  {Truncate(s.Report.Text, 62)}");

// Corroboration is what makes this alertable. A single report far from normal is
// usually just an unusual subject, not an incident; two or more inside the window
// mean several people independently noticed the same thing. Requiring two is what
// keeps the alert queue small enough that someone still reads it.
var flagged = scored.Where(s => s.Distance > threshold).OrderBy(s => s.Report.Date).ToList();
Console.WriteLine($"\n{flagged.Count} of {scored.Count} reports above threshold. Clustering them within {window} days:\n");

var alerts = 0;
for (var i = 0; i < flagged.Count;)
{
    var j = i + 1;
    while (j < flagged.Count && (Date(flagged[j].Report) - Date(flagged[j - 1].Report)).Days <= window) j++;
    var group = flagged[i..j];
    if (group.Count >= 2)
    {
        alerts++;
        Console.WriteLine($"  ALERT trail-{trail}: {group.Count} anomalous reports between {group[0].Report.Date} and {group[^1].Report.Date}");
        foreach (var s in group)
            Console.WriteLine($"        {s.Report.Id} {s.Report.Date}  {Truncate(s.Report.Text, 70)}");
    }
    else
    {
        Console.WriteLine($"  (ignored) {group[0].Report.Id} {group[0].Report.Date} is a lone outlier, not an event");
    }
    i = j;
}
Console.WriteLine($"\n{alerts} alert(s). Model calls: {reports.Count} embeddings, 0 chat completions.");

static DateTime Date(Report r) => DateTime.Parse(r.Date);

static float[] Normalize(float[] vector)
{
    var length = MathF.Sqrt(vector.Sum(v => v * v));
    return vector.Select(v => v / length).ToArray();
}

// Correct only for unit-length inputs, where the dot product is already the cosine
// similarity. Every vector reaching this method has been through Normalize; pass an
// unnormalized one and it returns a number that still looks plausible.
static float CosineDistance(float[] a, float[] b)
{
    var dot = 0f;
    for (var i = 0; i < a.Length; i++) dot += a[i] * b[i];
    return 1f - dot;
}

static string Truncate(string text, int max) =>
    text.Length <= max ? text : text[..(max - 1)] + "…";

record Report(string Id, string Trail_Id, string Date, string Text);
record Scored(Report Report, float Distance);
