using System.ComponentModel;
using System.Text.Json;
using System.Text.Json.Nodes;
using Azure;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished capstone demo, matching the outline in ../../README.md:
// five tools over the mock APIs, registered via Microsoft.Extensions.AI
// function invocation. The loop narrates itself: every tool call prints
// as it happens, the permit step pauses for a human yes, and a step
// budget keeps the loop from running forever.
//
//   dotnet run                                   the capstone request
//   dotnet run -- Plan me a trip on Avalanche Lake Trail in September
//   dotnet run -- --yes <request>                auto-approve the permit gate
//
// Model: Azure OpenAI when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY /
// AZURE_OPENAI_DEPLOYMENT are set; otherwise falls back to Ollama llama3.2.

var autoApprove = args.Contains("--yes");
var requestArgs = args.Where(a => a != "--yes").ToArray();
var request = requestArgs.Length > 0
    ? string.Join(" ", requestArgs)
    : "Plan me a 3-day trip in Glacier National Park for September 14-16.";

Trailhead.AutoApprovePermits = autoApprove;

IChatClient inner = CreateChatClient();
IChatClient client = new ChatClientBuilder(inner)
    .UseFunctionInvocation(configure: c => c.MaximumIterationsPerRequest = 12) // step budget: the loop cannot run forever
    .Build();

var options = new ChatOptions
{
    Tools =
    [
        AIFunctionFactory.Create(Trailhead.SearchTrails),
        AIFunctionFactory.Create(Trailhead.GetWeather),
        AIFunctionFactory.Create(Trailhead.GetTrailConditions),
        AIFunctionFactory.Create(Trailhead.CheckCampsites),
        AIFunctionFactory.Create(Trailhead.RequestPermit),
    ],
};

List<ChatMessage> messages =
[
    new(ChatRole.System, """
        You are the trip-planning agent for Trailhead Guides, a hiking app.
        Today's date is September 11, 2026.

        Plan trips using your tools; never invent trails, weather, availability,
        or conditions. Every trail name, forecast, campground, and condition in
        your answer must have come back from a tool call in this conversation.

        Call the tools one at a time, in this order, and do not write any part of
        the itinerary until all of them have been called:
        1. get_weather for the park.
        2. search_trails for candidate trails that fit the request.
        3. get_trail_conditions for EVERY trail you intend to recommend, one call
           per trail, using the trail id returned by search_trails.
           If the newest reports for a trail mention a closure, a washout, a bridge
           that is out, or any other reason hikers are turning around, that trail is
           CLOSED. Do not schedule a day on a closed trail. Replace it with another
           trail from search_trails and state plainly, in the itinerary, that the
           original trail is closed and why.
        4. check_campsites for where to stay each night.
        5. request_permit once, only if a backcountry site or permit zone is involved.

        If you have not yet called search_trails and get_trail_conditions, your
        next move is a tool call, not prose.

        Then write the final itinerary: one section per day with trail, campsite,
        and how the forecast shaped the choice (put harder or more exposed hiking
        on the drier days). End with the permit status.
        """),
    new(ChatRole.User, request),
];

Console.WriteLine($"Request: {request}");
Console.WriteLine(new string('=', 60));

// Small-model insurance. Azure OpenAI runs the whole plan from one call.
// llama3.2 often makes one tool call and then starts writing the itinerary,
// so when a required tool has not been called yet we say so and let the loop
// continue. Three nudges maximum; that is the outer half of the step budget.
string[] required = ["get_weather", "search_trails", "get_trail_conditions", "check_campsites"];
var response = await client.GetResponseAsync(messages, options);

for (var nudge = 0; nudge < 3; nudge++)
{
    var missing = required.Where(t => !Trailhead.Called.Contains(t)).ToArray();
    if (missing.Length == 0) break;

    Console.WriteLine($"[nudge] still missing: {string.Join(", ", missing)}");
    messages.AddRange(response.Messages);
    var hint = missing.Contains("get_trail_conditions") && Trailhead.LastResultIds.Count > 0
        ? $" Use one of these trail ids: {string.Join(", ", Trailhead.LastResultIds)}."
        : "";
    messages.Add(new ChatMessage(ChatRole.User,
        $"You have not called these tools yet: {string.Join(", ", missing)}. " +
        $"Call the next one now with real arguments.{hint} Do not write the itinerary yet."));
    response = await client.GetResponseAsync(messages, options);
}

// Same insurance at the other end: llama3.2 likes to announce that it has
// finished calling tools and then stop, without ever writing the plan.
if (!response.Text.Contains("Day", StringComparison.OrdinalIgnoreCase))
{
    Console.WriteLine("[nudge] tools are done but no itinerary was written; asking for it.");
    messages.AddRange(response.Messages);
    messages.Add(new ChatMessage(ChatRole.User,
        "Every tool you need has been called. Write the final itinerary now, " +
        "using only what the tools returned. Do not call any more tools."));
    response = await client.GetResponseAsync(messages, options);
}

Console.WriteLine(new string('=', 60));
Console.WriteLine(response.Text);

static IChatClient CreateChatClient()
{
    var endpoint = Environment.GetEnvironmentVariable("AZURE_OPENAI_ENDPOINT");
    var key = Environment.GetEnvironmentVariable("AZURE_OPENAI_KEY");
    var deployment = Environment.GetEnvironmentVariable("AZURE_OPENAI_DEPLOYMENT");

    if (!string.IsNullOrEmpty(endpoint) && !string.IsNullOrEmpty(key) && !string.IsNullOrEmpty(deployment))
    {
        return new AzureOpenAIClient(new Uri(endpoint), new AzureKeyCredential(key))
            .GetChatClient(deployment)
            .AsIChatClient();
    }

    Console.WriteLine("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.");
    return new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
}

// The five tools. Ordinary C# methods over the workshop's fixture files;
// the [Description] attributes are the model's only manual. Each method
// prints itself on entry, which is what makes the loop visible on stage.
static class Trailhead
{
    const string DataDir = "../../../../data";
    public static bool AutoApprovePermits;
    public static readonly HashSet<string> Called = [];
    public static readonly List<string> LastResultIds = [];

    static readonly JsonSerializerOptions Pretty = new() { WriteIndented = false };

    [Description("Search the trail catalog. Returns matching trails with id, name, park, distance, elevation, difficulty, and features.")]
    public static string SearchTrails(
        [Description("Park name, e.g. 'Glacier National Park'. Partial names like 'Glacier' work.")] string park = "Glacier National Park",
        [Description("Optional feature keywords to look for, e.g. ['lake', 'waterfall'].")] string[]? features = null,
        [Description("Optional maximum difficulty: 'easy', 'moderate', or 'hard'.")] string? maxDifficulty = null)
    {
        Narrate("search_trails", new { park, features, max_difficulty = maxDifficulty });

        var trails = JsonNode.Parse(File.ReadAllText($"{DataDir}/trails.json"))!.AsArray();
        var rank = (string d) => d switch { "easy" => 0, "moderate" => 1, _ => 2 };
        var maxRank = maxDifficulty is null ? 2 : rank(maxDifficulty.ToLowerInvariant());

        var matches = trails
            .Where(t => ((string)t!["park"]!).Contains(park, StringComparison.OrdinalIgnoreCase))
            .Where(t => rank((string)t!["difficulty"]!) <= maxRank)
            .Where(t => features is null || features.Length == 0 || features.Any(f =>
                t!["features"]!.AsArray().Any(x => ((string)x!).Contains(f, StringComparison.OrdinalIgnoreCase))))
            .Take(8)
            .Select(t => new
            {
                id = (string)t!["id"]!,
                name = (string)t["name"]!,
                park = (string)t["park"]!,
                distance_mi = (double)t["distance_mi"]!,
                elevation_ft = (int)t["elevation_ft"]!,
                difficulty = (string)t["difficulty"]!,
                features = t["features"]!.AsArray().Select(x => (string)x!).ToArray(),
            });

        var found = matches.ToArray();
        LastResultIds.Clear();
        LastResultIds.AddRange(found.Select(t => t.id));
        return Result(JsonSerializer.Serialize(found, Pretty));
    }

    [Description("Get the multi-day weather forecast and advisories for a park.")]
    public static string GetWeather(
        [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park")
    {
        Narrate("get_weather", new { park });

        var all = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/weather.json"))!.AsObject();
        var entry = all.FirstOrDefault(kv =>
            kv.Key.Contains(park, StringComparison.OrdinalIgnoreCase) ||
            park.Contains(kv.Key, StringComparison.OrdinalIgnoreCase) ||
            kv.Key.Contains(park.Split(' ')[0], StringComparison.OrdinalIgnoreCase));

        return Result(entry.Value is null
            ? $"{{\"error\": \"No forecast available for '{park}'.\"}}"
            : entry.Value.ToJsonString(Pretty));
    }

    [Description("Get the most recent hiker-submitted condition reports for a trail. Always check this before recommending a trail; reports surface closures and hazards such as washouts.")]
    public static string GetTrailConditions(
        [Description("The trail id from search_trails, e.g. 'trail-0117'. A trail name also works.")] string? trailId = null)
    {
        Narrate("get_trail_conditions", new { trail_id = trailId });

        // Small models routinely call this with no id at all. Hand the ids back
        // rather than failing, so the loop can keep going.
        if (string.IsNullOrWhiteSpace(trailId) || trailId is "null" or "string")
        {
            var candidates = LastResultIds.Count > 0 ? string.Join(", ", LastResultIds) : "call search_trails first";
            return Result($"{{\"error\": \"trailId is required. Call this tool again with one of these ids: {candidates}.\"}}");
        }

        // Tolerate a trail name instead of an id; local models do that sometimes.
        if (!trailId.StartsWith("trail-", StringComparison.OrdinalIgnoreCase))
        {
            var trails = JsonNode.Parse(File.ReadAllText($"{DataDir}/trails.json"))!.AsArray();
            var byName = trails.FirstOrDefault(t =>
                ((string)t!["name"]!).Contains(trailId, StringComparison.OrdinalIgnoreCase));
            if (byName is not null) trailId = (string)byName["id"]!;
        }

        var id = trailId;
        var reports = File.ReadLines($"{DataDir}/condition-reports.jsonl")
            .Select(l => JsonNode.Parse(l)!)
            .Where(r => string.Equals((string)r["trail_id"]!, id, StringComparison.OrdinalIgnoreCase))
            .OrderByDescending(r => (string)r["date"]!)
            .Take(4)
            .Select(r => new { date = (string)r["date"]!, report = (string)r["text"]! })
            .ToArray();

        return Result(reports.Length == 0
            ? $"{{\"error\": \"No condition reports found for '{trailId}'.\"}}"
            : JsonSerializer.Serialize(reports, Pretty));
    }

    [Description("Check campground availability in a park. Returns campgrounds with open sites per date, type (frontcountry or backcountry), and notes.")]
    public static string CheckCampsites(
        [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park")
    {
        Narrate("check_campsites", new { park });

        var all = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/campsites.json"))!.AsObject();
        var entry = all.FirstOrDefault(kv =>
            kv.Key.Contains(park, StringComparison.OrdinalIgnoreCase) ||
            park.Contains(kv.Key, StringComparison.OrdinalIgnoreCase) ||
            kv.Key.Contains(park.Split(' ')[0], StringComparison.OrdinalIgnoreCase));

        return Result(entry.Value is null
            ? $"{{\"error\": \"No campsite data for '{park}'.\"}}"
            : entry.Value.ToJsonString(Pretty));
    }

    [Description("Submit a backcountry permit request. This files a real request, so use it once, at the end, after the plan is settled.")]
    public static string RequestPermit(
        [Description("Park name, e.g. 'Glacier National Park'.")] string park = "Glacier National Park",
        [Description("Permit zone, e.g. 'Lake McDonald / Sperry'.")] string zone = "Lake McDonald / Sperry",
        [Description("Trip dates, e.g. '2026-09-14 to 2026-09-16'.")] string dates = "unspecified",
        [Description("Number of people in the group.")] int groupSize = 2)
    {
        Narrate("request_permit", new { park, zone, dates, group_size = groupSize });

        // The module 09 gate: a human confirms before anything irreversible.
        Console.WriteLine($"  [gate] About to file a permit request: {park}, zone '{zone}', {dates}, group of {groupSize}.");
        bool approved;
        if (AutoApprovePermits)
        {
            Console.WriteLine("  [gate] --yes supplied; auto-approved.");
            approved = true;
        }
        else
        {
            Console.Write("  [gate] File it? [y/N] ");
            approved = Console.ReadLine()?.Trim().ToLowerInvariant() is "y" or "yes";
        }

        if (!approved)
            return Result("{\"status\": \"cancelled\", \"message\": \"The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed.\"}");

        var permits = JsonNode.Parse(File.ReadAllText($"{DataDir}/mock-apis/permits.json"))!;
        return Result(permits["submit_response"]!.ToJsonString(Pretty));
    }

    static void Narrate(string tool, object args)
    {
        Called.Add(tool);
        Console.WriteLine($"[tool] {tool} {JsonSerializer.Serialize(args)}");
    }

    static string Result(string json)
    {
        var preview = json.Length > 120 ? json[..120] + "..." : json;
        Console.WriteLine($"  [result] {preview}");
        return json;
    }
}
