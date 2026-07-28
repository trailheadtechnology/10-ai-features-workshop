using System.Text.Json;

// Starting point: ranks trail reports by distance from the trail's baseline.
// Run: dotnet run
//
// Makes no model calls and needs no network. The vectors in
// ../../lab/embeddings-0117.json were precomputed with nomic-embed-text so this
// runs with Ollama down, which is the fallback when the room's network is not
// cooperating. Those vectors were embedded with the "classification: " task
// prefix nomic requires, so anything you add to this corpus must be embedded the
// same way or its distances will not be comparable to these.
//
// complete/ embeds live and adds the alert rule on top of this ranking.

var lab = "../../lab";
var json = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
var reports = File.ReadLines($"{lab}/reports-0117.jsonl")
    .Select(line => JsonSerializer.Deserialize<Report>(line, json)!)
    .ToList();

using var doc = JsonDocument.Parse(File.ReadAllText($"{lab}/embeddings-0117.json"));
var vectors = doc.RootElement.GetProperty("embeddings")
    .EnumerateObject()
    .ToDictionary(p => p.Name, p => Normalize(p.Value.EnumerateArray().Select(v => v.GetSingle()).ToArray()));

// The centroid is this trail's definition of normal, and it is built from the
// same reports it is about to judge. Anomalies that appear often enough pull the
// centroid toward themselves and stop looking anomalous.
var dimensions = vectors.Values.First().Length;
var centroid = new float[dimensions];
foreach (var vector in vectors.Values)
    for (var i = 0; i < dimensions; i++)
        centroid[i] += vector[i] / vectors.Count;
centroid = Normalize(centroid);

var scored = reports
    .Select(r => (Report: r, Distance: CosineDistance(vectors[r.Id], centroid)))
    .OrderByDescending(x => x.Distance)
    .ToList();

Console.WriteLine($"trail-0117 · {reports.Count} reports · {dimensions}-dim nomic-embed-text vectors\n");
Console.WriteLine("  dist    id       date        report");
foreach (var (report, distance) in scored)
    Console.WriteLine($"  {distance:F4}  {report.Id}  {report.Date}  {Truncate(report.Text, 62)}");

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
