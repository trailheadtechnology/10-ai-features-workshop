using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: one chat client, one method, one naive prompt.
// Run: dotnet run [path-to-trip-report.md]

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var reportPath = args.Length > 0 ? args[0] : "../../lab/tr-0001.md";
var report = StripFrontMatter(await File.ReadAllTextAsync(reportPath));

var response = await client.GetResponseAsync($"Summarize this trip report.\n\n{report}");
Console.WriteLine(response.Text);

static string StripFrontMatter(string markdown)
{
    var parts = markdown.Split("---", 3, StringSplitOptions.None);
    return parts.Length == 3 ? parts[2].Trim() : markdown.Trim();
}
