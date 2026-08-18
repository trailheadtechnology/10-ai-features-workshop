using System.Text.Json;
using System.Text.RegularExpressions;

// Demo starting point: keyword search over the trail slice. Split the query into
// words and count how many appear, as whole words, in each trail's name and
// description.
// Run: dotnet run -- <query>          (defaults to the demo query)
//
// Matching on words alone has no access to meaning. A description saying the
// trail avoids the steep section still matches "steep", and a trail whose text
// reads "leashed dogs are welcome" never matches "dog-friendly". That gap is
// what the complete project replaces with embeddings.

var query = args.Length > 0
    ? string.Join(' ', args)
    : "dog-friendly waterfall hike, not too steep";

var jsonOptions = new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower };
var json = await File.ReadAllTextAsync("../../data/trails-slice.json");
var trails = JsonSerializer.Deserialize<List<Trail>>(json, jsonOptions)!;

var tokens = Regex.Matches(query.ToLowerInvariant(), "[a-z]+")
    .Select(m => m.Value)
    .Where(w => w.Length >= 3)
    .Distinct()
    .ToList();

var results = trails
    .Select(t =>
    {
        var haystack = $"{t.Name} {t.Description}".ToLowerInvariant();
        var hits = tokens.Where(w => Regex.IsMatch(haystack, $@"\b{w}\b")).ToList();
        return (Trail: t, Hits: hits);
    })
    .Where(r => r.Hits.Count > 0)
    .OrderByDescending(r => r.Hits.Count)
    .ThenBy(r => r.Trail.Id)
    .Take(5)
    .ToList();

Console.WriteLine($"Keyword search: \"{query}\"");
Console.WriteLine($"Query words: {string.Join(", ", tokens)}\n");

if (results.Count == 0)
{
    Console.WriteLine("No results. Not one trail contains those words.");
    return;
}

foreach (var (trail, hits) in results)
{
    Console.WriteLine($"{hits.Count} word(s) [{string.Join(", ", hits)}]  " +
        $"{trail.Id}  {trail.Name} ({trail.Difficulty}, {trail.DistanceMi} mi)");
}

record Trail(string Id, string Name, string Park, double DistanceMi,
    int ElevationFt, string Difficulty, List<string> Features, string Description);
