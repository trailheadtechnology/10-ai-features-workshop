using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Classifies every inquiry in ../../lab/inquiries-slice.jsonl and scores the
// result against ../../lab/reference-labels.json.
// Run: dotnet run
//
// The category comes back as a C# enum through structured output, so the model
// can only return a label the routing table already knows how to handle. Adding
// a category here means adding a routing destination for it too, or the lookup
// on the routing table will throw.

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
// Anything above 0 makes the same message land in different queues on different
// runs, which makes a scored comparison against fixed reference labels meaningless.
var options = new ChatOptions { Temperature = 0 };

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

// Emergencies print before the routing table and are sorted to the top of it.
// A person scanning this output under time pressure must not have to read past
// the first screen to find one.
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

// Two scores, and they are not equally important. Overall accuracy is the
// headline number; recall on the emergency class is the one that decides
// whether this taxonomy is safe to ship. A missed emergency is a person waiting
// in a queue nobody is watching, and no amount of accuracy elsewhere offsets it.
// If you tune the category descriptions, judge the change on emergency recall
// first and treat a drop there as a failure even when accuracy improves.
var correct = results.Count(r => Wire(r.Category) == reference.Labels[r.Inquiry.id]);
var emergencyIds = reference.Labels.Where(l => l.Value == "emergency").Select(l => l.Key).ToList();
var caught = emergencies.Count(e => emergencyIds.Contains(e.Inquiry.id));

Console.WriteLine();
Console.WriteLine($"Accuracy vs reference labels: {correct}/{results.Count}");
Console.WriteLine($"Emergency recall: {caught}/{emergencyIds.Count} " +
    (caught == emergencyIds.Count ? "(all caught; the metric that matters)" : "(MISSED ONE; this fails, whatever the accuracy says)"));
foreach (var (inquiry, category) in results.Where(r => Wire(r.Category) != reference.Labels[r.Inquiry.id]))
    Console.WriteLine($"  miss: {inquiry.id} got {Wire(category)}, reference says {reference.Labels[inquiry.id]}");

// These descriptions are the taxonomy, and editing them changes behavior more
// than any code below. Two rules constrain any rewrite. Emergency wins over
// every other category, including messages that also mention a permit or a lost
// item, so the ordering paragraph at the end must stay. And unsure has to stay
// narrow: it means two queues must both act on one message, not that the model
// found the message hard. Widen it and it fills up with ordinary traffic, which
// is the unsorted inbox this system replaced.
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
    - unsure: two different queues both have to act before this message can
      be resolved, so no single queue owns it. The case that qualifies: the
      sender asks about trail conditions AND asks someone to change, refund,
      or cancel a booking. Trail info cannot issue a refund, and the permits
      office does not decide whether a trail is passable, so a human reads
      this queue and splits the work. Also use unsure when the message fits
      none of the categories above.

    Decide in this order. First, if anyone might be hurt, missing, or in
    danger, answer emergency and stop; never answer unsure for those, even
    when the message also mentions permits, conditions, or a lost item.
    Second, if one queue can resolve the whole message on its own, answer
    that queue; a booking or reservation problem with nothing else attached
    is permit, not unsure. Third, only if two queues must both act, answer
    unsure. Unsure is not a catch-all for anything hard.

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
    [JsonStringEnumMemberName("unsure")] Unsure,
}

record ReferenceLabels(
    [property: JsonPropertyName("routing")] Dictionary<string, string> Routing,
    [property: JsonPropertyName("labels")] Dictionary<string, string> Labels);
