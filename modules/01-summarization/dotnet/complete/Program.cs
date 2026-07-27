using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                                the naive prompt (the book report)
//   dotnet run -- --briefing                  3-bullet hiker briefing
//   dotnet run -- --headline                  one-line trail status for a card UI
//   dotnet run -- --briefing --audience ranger
//   Any non-flag argument is a path to a different trip report.

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var reportPath = "../../lab/tr-0004.md";
var mode = "naive";
var audience = "hiker";
for (var i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--briefing": mode = "briefing"; break;
        case "--headline": mode = "headline"; break;
        case "--audience": audience = args[++i]; break;
        default: reportPath = args[i]; break;
    }
}

var report = StripFrontMatter(await File.ReadAllTextAsync(reportPath));

var audienceFocus = audience switch
{
    "ranger" => "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery",
    _ => "a hiker planning to hike this trail within the next week",
};

var prompt = mode switch
{
    // Step 2 of the demo: right model, wrong instruction.
    "naive" => $"Summarize this trip report.\n\n{report}",

    // Step 4: a summary with a purpose. The instruction carries the feature.
    "briefing" => $"""
        You are helping {audienceFocus}.
        From the trip report below, produce exactly 3 bullets covering:
        current trail conditions, hazards or closures, and crowding.
        Ignore gear talk, personal stories, and scenery.
        If the report mentions a closure or hazard, it must appear in the first bullet.

        {report}
        """,

    // Step 5: same call, different product surface.
    "headline" => $"""
        From the trip report below, write ONE line of at most 12 words,
        suitable for a status badge on a trail card in an app.
        Lead with the most important condition or closure. No preamble.

        {report}
        """,

    _ => throw new ArgumentException(mode),
};

var response = await client.GetResponseAsync(prompt);
Console.WriteLine(response.Text);

static string StripFrontMatter(string markdown)
{
    var parts = markdown.Split("---", 3, StringSplitOptions.None);
    return parts.Length == 3 ? parts[2].Trim() : markdown.Trim();
}
