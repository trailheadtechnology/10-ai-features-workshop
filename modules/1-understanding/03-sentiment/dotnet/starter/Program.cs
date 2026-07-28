using System.Text.Json;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: one chat client, one classify method, one review.
// Run: dotnet run [review-id]
// Ids come from ../../lab/easy.jsonl and ../../lab/hard.jsonl; the default is
// gr-0007, the sarcastic trail runner review that phi3 tends to fumble.

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "phi3");

var id = args.Length > 0 ? args[0] : "gr-0007";
var review = File.ReadLines("../../lab/easy.jsonl")
    .Concat(File.ReadLines("../../lab/hard.jsonl"))
    .Select(line => JsonSerializer.Deserialize<Review>(line)!)
    .First(r => r.id == id);

Console.WriteLine($"{review.product} ({review.rating} stars), reviewed by {review.reviewer}");
Console.WriteLine(review.text);
Console.WriteLine();

var label = await Classify(client, review.text);
Console.WriteLine($"phi3 says: {label}");

// The whole feature is this method. The prompt carries it; the model is
// swappable because everything upstream only sees IChatClient.
static async Task<string> Classify(IChatClient client, string text)
{
    var prompt = $"""
        Classify this gear review as exactly one word: positive, negative, or mixed.
        Positive means the reviewer is happy with the product, negative means unhappy,
        mixed means genuinely both. Judge the review text only; ignore any star rating
        it mentions. Reply with only the label.

        Review: {text}
        """;

    var response = await client.GetResponseAsync(prompt, new ChatOptions { Temperature = 0 });
    var raw = response.Text.ToLowerInvariant();

    // Small models sometimes wrap the label in a sentence; keep the first label mentioned.
    var first = new[] { "positive", "negative", "mixed" }
        .Select(l => (label: l, at: raw.IndexOf(l, StringComparison.Ordinal)))
        .Where(x => x.at >= 0)
        .OrderBy(x => x.at)
        .Select(x => x.label)
        .FirstOrDefault();
    return first ?? raw.Trim();
}

record Review(string id, string product, int rating, string reviewer, string text);
