using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: the naive approach. Ask for JSON in the prompt, get
// back whatever the model feels like: prose preamble, markdown fences,
// drifting field names. This is what the typed response replaces.
// Run: dotnet run [path-to-trip-report.md]

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var reportPath = args.Length > 0 ? args[0] : "../../data/tr-0007.md";
var report = StripFrontMatter(await File.ReadAllTextAsync(reportPath));

var response = await client.GetResponseAsync(
    $"Extract the details of this trip report as JSON.\n\n{report}");
Console.WriteLine(response.Text);

static string StripFrontMatter(string markdown)
{
    var parts = markdown.Split("---", 3, StringSplitOptions.None);
    return parts.Length == 3 ? parts[2].Trim() : markdown.Trim();
}
