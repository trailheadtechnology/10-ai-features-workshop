using Microsoft.Extensions.AI;
using OllamaSharp;
using System.Text.Json;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                          "more like this" for Avalanche Lake Trail
//   dotnet run -- trail-0008            any trail id works
//   dotnet run -- Trail of the Cedars   so does any name (or part of one)
//   dotnet run -- --gear Cascade 65     the same trick on gear, from review text
//
// Vectors are cached in embeddings.json / gear-embeddings.json in the project
// directory. The cache is only checked for missing keys, so an edited description
// or a different embedding model leaves the stale vectors in place. Delete the
// cache file whenever the source text or the model changes.

IEmbeddingGenerator<string, Embedding<float>> generator =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

if (args.Length > 0 && args[0] == "--gear")
{
    await RecommendGear(generator, string.Join(' ', args.Skip(1)));
    return;
}

// Same catalog and same embedding model as feature 04. Recommendations need no
// new model and no new data, only the vectors search already produced.
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

// "More like this" is the search from feature 04 with the query vector replaced
// by an item's own vector. That means it ranks on what the descriptions talk
// about, so whatever the prose leaves out is invisible: difficulty and distance
// live in structured fields and are never mentioned in the text, and a moderate
// family hike will cheerfully return a list of hard all-day climbs. If those
// fields matter to the user, filter or re-rank on them after the similarity pass.
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

// Products have no descriptions, so each vector comes from that product's reviews
// concatenated: "similar" here means "reviewers describe them the same way".
//
// Content similarity finds substitutes, not complements. The nearest neighbor to
// a backpack is usually another size of the same backpack, which is the one item
// its owner will never buy. Complements come from behavior data (what people buy
// or mention together), and no embedding of the product text can supply it.
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

// Embed each text once and cache the vectors to disk. Note the staleness trap:
// the cache is accepted whenever it holds every key, so changed text under an
// existing key keeps its old vector. Delete the file to force a re-embed.
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
