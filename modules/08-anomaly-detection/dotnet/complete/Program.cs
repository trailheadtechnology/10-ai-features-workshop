using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                    trail-0117, live embeddings, distance table + cluster alerts
//   dotnet run -- --raw           same trail without nomic's task prefix (the failure mode)
//   dotnet run -- --trail 0042    stretch goal: the bear-activity spike on the other trail
//   dotnet run -- --sigma 1.5     tighter threshold
//   dotnet run -- --window 30     wider clustering window, in days
//
// Embeddings are the only model calls. Everything after them is arithmetic.

var trail = "0117";
var sigma = 1.0;
var window = 14;
var prefix = "classification: ";   // nomic-embed-text is trained with task prefixes; see ../README.md
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
var reports = File.ReadLines($"../../lab/reports-{trail}.jsonl")
    .Select(line => JsonSerializer.Deserialize<Report>(line, json)!)
    .ToList();

// The only model calls in the whole feature.
IEmbeddingGenerator<string, Embedding<float>> embedder =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

var generated = await embedder.GenerateAsync(reports.Select(r => prefix + r.Text));
var vectors = generated.Select(e => Normalize(e.Vector.ToArray())).ToList();

// Step 1: the centroid. The mathematical center of "normal" for this trail.
var dimensions = vectors[0].Length;
var centroid = new float[dimensions];
foreach (var vector in vectors)
    for (var i = 0; i < dimensions; i++)
        centroid[i] += vector[i] / vectors.Count;
centroid = Normalize(centroid);

// Step 2: score every report by how far it sits from normal.
var scored = reports
    .Select((r, i) => new Scored(r, CosineDistance(vectors[i], centroid)))
    .OrderByDescending(s => s.Distance)
    .ToList();

// Step 3: a threshold the data picks for itself, rather than a magic number.
var mean = scored.Average(s => s.Distance);
var deviation = Math.Sqrt(scored.Average(s => Math.Pow(s.Distance - mean, 2)));
var threshold = mean + sigma * deviation;

Console.WriteLine($"trail-{trail} · {reports.Count} reports · nomic-embed-text ({dimensions} dims)"
    + (prefix.Length == 0 ? " · NO task prefix" : $" · prefix \"{prefix.Trim()}\""));
Console.WriteLine($"mean distance {mean:F4} · sd {deviation:F4} · threshold mean+{sigma:0.#}sd = {threshold:F4}\n");

Console.WriteLine("  dist    id       date        report");
foreach (var s in scored)
    Console.WriteLine($"{(s.Distance > threshold ? " !" : "  ")}{s.Distance:F4}  {s.Report.Id}  {s.Report.Date}  {Truncate(s.Report.Text, 62)}");

// Step 4: one outlier is a rambling hiker. Several outliers close together in time are an event.
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

// Both vectors are unit length here, so the dot product is the cosine similarity.
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
