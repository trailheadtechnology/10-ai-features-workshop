using System.Text.Json;

// Demo starting point: no model calls at all.
// It loads the precomputed nomic-embed-text vectors from ../../lab/embeddings-0117.json,
// averages them into a centroid, and prints every report's cosine distance from it.
// Run: dotnet run
//
// This is the whole idea of the feature in about forty lines of arithmetic.
// The complete/ project does the embedding live and adds the alert rule.

var lab = "../../lab";
var json = new JsonSerializerOptions { PropertyNameCaseInsensitive = true };
var reports = File.ReadLines($"{lab}/reports-0117.jsonl")
    .Select(line => JsonSerializer.Deserialize<Report>(line, json)!)
    .ToList();

using var doc = JsonDocument.Parse(File.ReadAllText($"{lab}/embeddings-0117.json"));
var vectors = doc.RootElement.GetProperty("embeddings")
    .EnumerateObject()
    .ToDictionary(p => p.Name, p => Normalize(p.Value.EnumerateArray().Select(v => v.GetSingle()).ToArray()));

// The centroid: the mathematical center of "normal" for this trail.
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
