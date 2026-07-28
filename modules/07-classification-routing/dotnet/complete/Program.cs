using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
// classify all 20 inquiries from ../../lab/inquiries-slice.jsonl with
// structured output (a C# enum, so the model cannot invent a category),
// flag emergencies at the top, print the routing table, and score against
// ../../lab/reference-labels.json.
// Run: dotnet run

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
var options = new ChatOptions { Temperature = 0 }; // stable labels run to run

var inquiries = File.ReadLines("../../lab/inquiries-slice.jsonl")
    .Select(line => JsonSerializer.Deserialize<Inquiry>(line)!)
    .ToList();

var reference = JsonSerializer.Deserialize<ReferenceLabels>(
    File.ReadAllText("../../lab/reference-labels.json"))!;

var results = new List<(Inquiry Inquiry, Category Category)>();
foreach (var inquiry in inquiries)
{
    var response = await client.GetResponseAsync<TriageResult>(Prompt(inquiry.text), options);
    results.Add((inquiry, response.Result.Category));
    Console.Write('.');
}
Console.WriteLine('\n');

// Emergencies first, loudly. This queue is the reason the system exists.
var emergencies = results.Where(r => r.Category == Category.Emergency).ToList();
if (emergencies.Count > 0)
{
    Console.WriteLine("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!");
    foreach (var (inquiry, _) in emergencies)
        Console.WriteLine($"!!! {inquiry.id}  {Clip(inquiry.text, 70)}");
    Console.WriteLine();
}

Console.WriteLine($"{"id",-10} {"category",-15} routed to");
Console.WriteLine(new string('-', 62));
foreach (var (inquiry, category) in results.OrderBy(r => r.Category != Category.Emergency))
    Console.WriteLine($"{inquiry.id,-10} {Wire(category),-15} {reference.Routing[Wire(category)]}");

// Score against the reference labels. Overall accuracy is the small number;
// emergency recall is the one that matters.
var correct = results.Count(r => Wire(r.Category) == reference.Labels[r.Inquiry.id]);
var emergencyIds = reference.Labels.Where(l => l.Value == "emergency").Select(l => l.Key).ToList();
var caught = emergencies.Count(e => emergencyIds.Contains(e.Inquiry.id));

Console.WriteLine();
Console.WriteLine($"Accuracy vs reference labels: {correct}/{results.Count}");
Console.WriteLine($"Emergency recall: {caught}/{emergencyIds.Count} " +
    (caught == emergencyIds.Count ? "(all caught; the metric that matters)" : "(MISSED ONE; this fails, whatever the accuracy says)"));
foreach (var (inquiry, category) in results.Where(r => Wire(r.Category) != reference.Labels[r.Inquiry.id]))
    Console.WriteLine($"  miss: {inquiry.id} got {Wire(category)}, reference says {reference.Labels[inquiry.id]}");

static string Prompt(string text) => $"""
    You are the triage system for the Trailhead Guides shared inbox.
    Classify the visitor message into exactly one category.

    - permit: reserving, changing, canceling, or paying for a permit, pass,
      or reservation, including billing problems and missing confirmations
      for a permit application.
    - conditions: asking whether a trail, road, or area is open, safe, or
      passable right now: snow, water levels, washouts, wildlife activity,
      closures.
    - complaint: unhappy about a park facility, service, or staff member
      and wants it acknowledged or fixed.
    - lost-and-found: reporting a lost or found physical item.
    - emergency: a person may be hurt, missing, or in danger right now and
      needs immediate human attention.
    - general: anything else: park rules, fees, trip planning, questions
      that fit none of the above.

    When a message could fit two categories, pick the one whose queue can
    actually act on it. If anyone might be in danger, it is always emergency.

    Message:
    {text}
    """;

static string Wire(Category c) => c switch
{
    Category.LostAndFound => "lost-and-found",
    _ => c.ToString().ToLowerInvariant(),
};

static string Clip(string text, int max) =>
    text.Length <= max ? text : text[..max] + "...";

record Inquiry(string id, string channel, string received, string text);

record TriageResult(Category Category);

[JsonConverter(typeof(JsonStringEnumConverter<Category>))]
enum Category
{
    [JsonStringEnumMemberName("permit")] Permit,
    [JsonStringEnumMemberName("conditions")] Conditions,
    [JsonStringEnumMemberName("complaint")] Complaint,
    [JsonStringEnumMemberName("lost-and-found")] LostAndFound,
    [JsonStringEnumMemberName("emergency")] Emergency,
    [JsonStringEnumMemberName("general")] General,
}

record ReferenceLabels(
    [property: JsonPropertyName("routing")] Dictionary<string, string> Routing,
    [property: JsonPropertyName("labels")] Dictionary<string, string> Labels);
