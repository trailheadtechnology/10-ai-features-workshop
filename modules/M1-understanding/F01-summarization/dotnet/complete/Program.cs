using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the demo script in docs/slides/outlines:
//   dotnet run                                the naive prompt (the book report)
//   dotnet run -- --briefing                  3-bullet hiker briefing
//   dotnet run -- --headline                  one-line trail status for a card UI
//   dotnet run -- --briefing --audience ranger
//   Any non-flag argument is a path to a different trip report.

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var reportPath = "../../data/tr-0004.md";
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
    // Deliberately the weak prompt. It is kept so the two prompts can be run
    // back to back against the same report; nothing about it should be fixed.
    "naive" => $"Summarize this trip report.\n\n{report}",

    // The last two lines are load-bearing, not politeness. The bullets require a
    // hazards slot and require any hazard to come first, so on a report with no
    // hazard the model will promote the nearest noun (a bear, a creek, the word
    // "avalanche" in the trail name) into a closure. Giving it a legal way to
    // report nothing is what stops that. Measurements in ../../expected-output.md.
    "briefing" => $"""
        You are helping {audienceFocus}.
        From the trip report below, produce exactly 3 bullets covering:
        current trail conditions, hazards or closures, and crowding.
        Ignore gear talk, personal stories, and scenery.
        Report only what the trip report states. Do not turn a wildlife sighting into a
        hazard or a closure, and write "no closures or hazards reported" when it says none.
        If the report does state a closure or hazard, it must appear in the first bullet.

        {report}
        """,

    // Same client, same report, same call. Only the instruction changes to fit a
    // different UI slot, so no new infrastructure is needed for a new surface.
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
