using System.ComponentModel;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                          extract both lab reports as typed rows
//   dotnet run -- path1.md [path2.md]   extract any report(s) instead
// The schema is the C# record below. Nullable fields plus the [Description]
// attributes ("null if not stated") are what keep the sparse report honest.

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var reportPaths = args.Length > 0
    ? args
    : ["../../lab/tr-0007.md", "../../lab/tr-0011.md"];

foreach (var reportPath in reportPaths)
{
    var report = StripFrontMatter(await File.ReadAllTextAsync(reportPath));

    // Steps 2-3 of the demo: prose in, populated .NET object out. No parsing.
    var response = await client.GetResponseAsync<TripFacts>(
        $"""
        Extract the trail facts from this trip report.
        Use null for any field the report does not state, and empty arrays
        when nothing applies. Do not guess.

        {report}
        """);

    var f = response.Result;
    Console.WriteLine($"== {Path.GetFileName(reportPath)} ==");
    Console.WriteLine($"  trail:      {f.TrailName ?? "null"}");
    Console.WriteLine($"  park:       {f.Park ?? "null"}");
    Console.WriteLine($"  date:       {f.DateHiked ?? "null"}");
    Console.WriteLine($"  distance:   {f.DistanceMi?.ToString("0.#") ?? "null"} mi");
    Console.WriteLine($"  elev gain:  {f.ElevationGainFt?.ToString("0") ?? "null"} ft");
    Console.WriteLine($"  wildlife:   [{string.Join(", ", f.Wildlife)}]");
    Console.WriteLine($"  conditions: [{string.Join(", ", f.Conditions)}]");
    Console.WriteLine($"  hazards:    [{string.Join(", ", f.Hazards)}]");
    Console.WriteLine();
}

static string StripFrontMatter(string markdown)
{
    var parts = markdown.Split("---", 3, StringSplitOptions.None);
    return parts.Length == 3 ? parts[2].Trim() : markdown.Trim();
}

// Step 5 of the demo: the schema does the prompting. Every scalar is nullable
// and every description says when to use null; that, not prompt pleading, is
// the hallucination fix.
record TripFacts(
    [property: Description("The name of the trail hiked. null if the report never names the trail.")]
    string? TrailName,
    [property: Description("The park the trail is in. null if the report never names the park.")]
    string? Park,
    [property: Description("The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.")]
    string? DateHiked,
    [property: Description("Round-trip distance in miles, as stated in the report. null if the report gives no distance. Never estimate.")]
    double? DistanceMi,
    [property: Description("Elevation gain in feet, as stated in the report. null if the report gives no elevation figure. Never estimate.")]
    double? ElevationGainFt,
    [property: Description("Animals the author actually saw on this hike. Empty array if none are mentioned.")]
    string[] Wildlife,
    [property: Description("Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none.")]
    string[] Conditions,
    [property: Description("Hazards or closures the report mentions. Empty array if none.")]
    string[] Hazards);
