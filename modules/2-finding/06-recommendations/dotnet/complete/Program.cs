using Microsoft.Extensions.AI;
using OllamaSharp;
using System.Text.Json;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                          "more like this" for Avalanche Lake Trail
//   dotnet run -- trail-0008            any trail id works
//   dotnet run -- Trail of the Cedars   so does any name (or part of one)
//   dotnet run -- --gear Cascade 65     step 5: same trick on gear, from review text
//
// Vectors are cached in embeddings.json / gear-embeddings.json next to this
// file; delete a cache to re-embed fresh.

IEmbeddingGenerator<string, Embedding<float>> generator =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

if (args.Length > 0 && args[0] == "--gear")
{
    await RecommendGear(generator, string.Join(' ', args.Skip(1)));
    return;
}

// Step 2 of the demo: the embedded trail catalog from feature 04. One item,
// one description, one vector. Nothing new is created here, which is the point.
var trails = JsonSerializer.Deserialize<List<Trail>>(
    await File.ReadAllTextAsync("../../lab/trails.json"),
    new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower })!;

var vectors = await EmbedWithCache(generator, "embeddings.json",
    trails.ToDictionary(t => t.Id, t => t.Description));

var query = args.Length > 0 ? string.Join(' ', args) : "trail-0117";
var target = trails.FirstOrDefault(t =>
        t.Id.Equals(query, StringComparison.OrdinalIgnoreCase) ||
        t.Name.Contains(query, StringComparison.OrdinalIgnoreCase))
    ?? throw new ArgumentException($"No trail matches '{query}'.");

// Step 3: "more like this" is the search code with the query swapped for an
// item: take this trail's vector, rank every other trail, skip itself, take 5.
Console.WriteLine($"You liked: {target.Name} ({target.Park})");
Console.WriteLine("You might also like:\n");

var hits = trails
    .Where(t => t.Id != target.Id)
    .Select(t => (Trail: t, Score: Cosine(vectors[target.Id], vectors[t.Id])))
    .OrderByDescending(h => h.Score)
    .Take(5);

foreach (var (trail, score) in hits)
{
    Console.WriteLine($"  {score:F4}  {trail.Name} ({trail.Park}, {trail.Difficulty}; {string.Join(", ", trail.Features)})");
}

// Step 5: gear. No descriptions here, so each product's vector comes from its
// combined review text: "similar products" means "products reviewers describe
// the same way".
static async Task RecommendGear(IEmbeddingGenerator<string, Embedding<float>> generator, string query)
{
    var reviews = File.ReadLines("../../../../../data/gear-reviews.jsonl")
        .Select(line => JsonSerializer.Deserialize<Review>(line,
            new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower })!)
        .ToList();

    var reviewText = reviews
        .GroupBy(r => r.Product)
        .ToDictionary(g => g.Key, g => string.Join("\n", g.Select(r => r.Text)));

    var vectors = await EmbedWithCache(generator, "gear-embeddings.json", reviewText);

    var target = reviewText.Keys.FirstOrDefault(p => p.Contains(query, StringComparison.OrdinalIgnoreCase))
        ?? throw new ArgumentException($"No product matches '{query}'.");

    Console.WriteLine($"You bought: {target}");
    Console.WriteLine("Goes well with:\n");

    var hits = vectors
        .Where(v => v.Key != target)
        .Select(v => (Product: v.Key, Score: Cosine(vectors[target], v.Value)))
        .OrderByDescending(h => h.Score)
        .Take(5);

    foreach (var (product, score) in hits)
        Console.WriteLine($"  {score:F4}  {product}");
}

// Embed each text once and cache the vectors; a second run costs nothing.
static async Task<Dictionary<string, float[]>> EmbedWithCache(
    IEmbeddingGenerator<string, Embedding<float>> generator, string cachePath,
    Dictionary<string, string> texts)
{
    if (File.Exists(cachePath))
    {
        var cached = JsonSerializer.Deserialize<Dictionary<string, float[]>>(
            await File.ReadAllTextAsync(cachePath))!;
        if (texts.Keys.All(cached.ContainsKey)) return cached;
    }

    var keys = texts.Keys.ToList();
    var embeddings = await generator.GenerateAsync(keys.Select(k => texts[k]));
    var vectors = keys.Zip(embeddings, (k, e) => (k, e.Vector.ToArray()))
        .ToDictionary(p => p.k, p => p.Item2);

    await File.WriteAllTextAsync(cachePath, JsonSerializer.Serialize(vectors));
    return vectors;
}

// Step 4: the math at the center of the feature, small enough to read.
static float Cosine(float[] a, float[] b)
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

record Trail(string Id, string Name, string Park, double DistanceMi, int ElevationFt,
    string Difficulty, string[] Features, string Description);

record Review(string Id, string Product, int Rating, string Reviewer, string Text);
