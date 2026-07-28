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

    Answer with the category name only.

    Message:
    {inquiry.text}
    """;

var response = await client.GetResponseAsync(prompt);
Console.WriteLine($"{inquiry.id}: {response.Text}");

record Inquiry(string id, string channel, string received, string text);
