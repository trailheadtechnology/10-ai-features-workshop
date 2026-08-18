using System.Text.Json;

// Demo starting point: the "you might also like" box, picking trails at random.
// Run: dotnet run [trail id or name]   (default: trail-0117, Avalanche Lake Trail)

var trails = JsonSerializer.Deserialize<List<Trail>>(
    await File.ReadAllTextAsync("../../data/trails.json"),
    new JsonSerializerOptions { PropertyNamingPolicy = JsonNamingPolicy.SnakeCaseLower })!;

var query = args.Length > 0 ? string.Join(' ', args) : "trail-0117";
var target = trails.FirstOrDefault(t =>
        t.Id.Equals(query, StringComparison.OrdinalIgnoreCase) ||
        t.Name.Contains(query, StringComparison.OrdinalIgnoreCase))
    ?? throw new ArgumentException($"No trail matches '{query}'.");

Console.WriteLine($"You liked: {target.Name} ({target.Park})");
Console.WriteLine("You might also like (picked at random, which is the current feature):\n");

foreach (var t in trails.Where(t => t.Id != target.Id).OrderBy(_ => Random.Shared.Next()).Take(5))
{
    Console.WriteLine($"  {t.Name} ({t.Park}, {t.Difficulty})");
}

record Trail(string Id, string Name, string Park, double DistanceMi, int ElevationFt,
    string Difficulty, string[] Features, string Description);
