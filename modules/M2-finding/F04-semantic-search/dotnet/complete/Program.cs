using System.Diagnostics;
using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the demo script in docs/slides/outlines:
//   dotnet run -- dog-friendly waterfall hike, not too steep
//   dotnet run -- somewhere quiet to take my kids
// Embeds every trail description once, embeds the query, ranks by cosine
// similarity, prints the top 5.

var query = args.Length > 0
    ? string.Join(' ', args)
    : "dog-friendly waterfall hike, not too steep";

IEmbeddingGenerator<string, Embedding<float>> generator =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

var jsonOptions = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower };
var json = await File.ReadAllTextAsync("../../data/trails-slice.json");
var trails = JsonSerializer.Deserialize<List<Trail>>(json, jsonOptions)!;

// Embedding the catalog takes seconds, so the vectors are cached next to the
// binary. The cache is keyed by trail id and nothing else: if a description
// changes, or the embedding model changes, delete embeddings.json. Otherwise
// every later query is ranked against vectors for text that no longer exists.
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

// The query has to go through the same model that produced the cached vectors.
// Vectors from two different models are not comparable, and cosine similarity
// will still return confident-looking numbers if you mix them.
var queryVector = (await generator.GenerateAsync([query]))[0].Vector.ToArray();

// This ranks on topic, not on suitability. An embedding cannot tell "great for
// kids" from "dangerous for kids" or "easy" from "never uses the word easy but
// is a cliff", so a top result can be about the right subject and still be the
// worst possible recommendation. Read the absolute scores as well: a top 5
// bunched together at a low score means nothing in the catalog is a real match
// and the order is mostly noise. Anything you can express as a filter over the
// metadata you already have (difficulty, features) is cheaper and more reliable
// than hoping the vector carries it.
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
