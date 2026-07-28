using System.Diagnostics;
using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run -- dog-friendly waterfall hike, not too steep
//   dotnet run -- somewhere quiet to take my kids
// Embeds all 30 trail descriptions once (cached to embeddings.json next to
// the binary), embeds the query, ranks by cosine similarity, prints top 5.

var query = args.Length > 0
    ? string.Join(' ', args)
    : "dog-friendly waterfall hike, not too steep";

IEmbeddingGenerator<string, Embedding<float>> generator =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

var jsonOptions = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower };
var json = await File.ReadAllTextAsync("../../lab/trails-slice.json");
var trails = JsonSerializer.Deserialize<List<Trail>>(json, jsonOptions)!;

// Step 3 of the demo: embed the whole catalog. Once. It takes seconds,
// so we cache the vectors next to the binary and never pay again.
var cachePath = Path.Combine(AppContext.BaseDirectory, "embeddings.json");
Dictionary<string, float[]> vectors;
if (File.Exists(cachePath))
{
    vectors = JsonSerializer.Deserialize<Dictionary<string, float[]>>(
        await File.ReadAllTextAsync(cachePath))!;
    Console.WriteLine($"Loaded {vectors.Count} cached vectors from embeddings.json");
}
else
{
    var sw = Stopwatch.StartNew();
    var embeddings = await generator.GenerateAsync(trails.Select(t => t.Description));
    vectors = trails.Zip(embeddings)
        .ToDictionary(p => p.First.Id, p => p.Second.Vector.ToArray());
    await File.WriteAllTextAsync(cachePath, JsonSerializer.Serialize(vectors));
    Console.WriteLine($"Embedded {vectors.Count} trail descriptions in {sw.ElapsedMilliseconds} ms");
}

// Embed the query the same way, then rank the catalog against it.
var queryVector = (await generator.GenerateAsync([query]))[0].Vector.ToArray();

var results = trails
    .Select(t => (Trail: t, Score: CosineSimilarity(queryVector, vectors[t.Id])))
    .OrderByDescending(r => r.Score)
    .Take(5);

Console.WriteLine($"\nSemantic search: \"{query}\"\n");
foreach (var (trail, score) in results)
{
    Console.WriteLine($"{score:F4}  {trail.Id}  {trail.Name} " +
        $"({trail.Difficulty}, {trail.DistanceMi} mi)  [{string.Join(", ", trail.Features)}]");
}

// Step 4 of the demo: the entire math of semantic search, on one screen.
static float CosineSimilarity(float[] a, float[] b)
{
    float dot = 0, magA = 0, magB = 0;
    for (var i = 0; i < a.Length; i++)
    {
        dot += a[i] * b[i];
        magA += a[i] * a[i];
        magB += b[i] * b[i];
    }
    return dot / (MathF.Sqrt(magA) * MathF.Sqrt(magB));
}

record Trail(string Id, string Name, string Park, double DistanceMi,
    int ElevationFt, string Difficulty, List<string> Features, string Description);
