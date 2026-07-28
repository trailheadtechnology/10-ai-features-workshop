using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: one chat client, one classify method, one inquiry.
// The taxonomy lives in the prompt as plain-language category descriptions;
// the model answers with a bare label as free text.
// Run: dotnet run [inquiry-id]     (default inq-0005)

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var wantedId = args.Length > 0 ? args[0] : "inq-0005";
var inquiry = File.ReadLines("../../lab/inquiries-slice.jsonl")
    .Select(line => JsonSerializer.Deserialize<Inquiry>(line)!)
    .First(i => i.id == wantedId);

var prompt = $"""
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
    Answer with the category name only.

    Message:
    {inquiry.text}
    """;

var response = await client.GetResponseAsync(prompt);
Console.WriteLine($"{inquiry.id}: {response.Text}");

record Inquiry(string id, string channel, string received, string text);
