using System.ClientModel;
using System.Text.Json;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                 both sets, both models, table + accuracy + disagreements
//   dotnet run -- --easy       easy set only (demo steps 3 and 4)
//   dotnet run -- --hard       hard set only (demo step 5)
//
// The big model is Azure OpenAI when AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_KEY,
// and AZURE_OPENAI_DEPLOYMENT are set. When they aren't, llama3.2 on Ollama
// stands in so the whole comparison runs offline. Either way, the swap is the
// one line building `bigModel` below; nothing downstream changes.

var sets = args.Contains("--easy") ? new[] { "easy" }
         : args.Contains("--hard") ? new[] { "hard" }
         : new[] { "easy", "hard" };

// Model 1: the small local model. Free, private, 2GB.
IChatClient phi3 = new OllamaApiClient(new Uri("http://localhost:11434"), "phi3");

// Model 2: the big model, or its local stand-in.
var endpoint = Environment.GetEnvironmentVariable("AZURE_OPENAI_ENDPOINT");
var key = Environment.GetEnvironmentVariable("AZURE_OPENAI_KEY");
var deployment = Environment.GetEnvironmentVariable("AZURE_OPENAI_DEPLOYMENT");

IChatClient bigModel;
string bigName;
if (!string.IsNullOrEmpty(endpoint) && !string.IsNullOrEmpty(key) && !string.IsNullOrEmpty(deployment))
{
    bigModel = new AzureOpenAIClient(new Uri(endpoint), new ApiKeyCredential(key))
        .GetChatClient(deployment)
        .AsIChatClient();
    bigName = $"azure:{deployment}";
}
else
{
    Console.WriteLine("AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.");
    Console.WriteLine();
    bigModel = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
    bigName = "llama3.2";
}

var labels = JsonSerializer.Deserialize<Dictionary<string, RefLabel>>(
    await File.ReadAllTextAsync("../../lab/reference-labels.json"))!;

var results = new List<Result>();
foreach (var set in sets)
{
    Console.WriteLine($"── {set} set ──");
    Console.WriteLine($"{"id",-9} {"reference",-10} {"phi3",-10} {bigName,-10}");
    foreach (var line in File.ReadLines($"../../lab/{set}.jsonl"))
    {
        var review = JsonSerializer.Deserialize<Review>(line)!;
        var reference = labels[review.id].label;
        var small = await Classify(phi3, review.text);
        var big = await Classify(bigModel, review.text);
        results.Add(new(review, set, reference, small, big));
        var flag = small != big ? "  <- disagree" : "";
        Console.WriteLine($"{review.id,-9} {reference,-10} {small,-10} {big,-10}{flag}");
    }
    Console.WriteLine();
}

Console.WriteLine("── accuracy vs. reference labels ──");
foreach (var set in sets)
{
    var batch = results.Where(r => r.Set == set).ToList();
    var smallOk = batch.Count(r => r.Small == r.Reference);
    var bigOk = batch.Count(r => r.Big == r.Reference);
    Console.WriteLine($"{set,-5}  phi3 {smallOk}/{batch.Count}   {bigName} {bigOk}/{batch.Count}");
}
Console.WriteLine();

var disagreements = results.Where(r => r.Small != r.Big).ToList();
Console.WriteLine($"── disagreements ({disagreements.Count} of {results.Count}) ──");
foreach (var d in disagreements)
{
    var verdict = d.Big == d.Reference ? $"{bigName} right"
                : d.Small == d.Reference ? "phi3 right"
                : "both wrong";
    Console.WriteLine($"{d.Review.id} [{d.Set}] ref={d.Reference} phi3={d.Small} {bigName}={d.Big}  ({verdict})");
    Console.WriteLine($"  \"{Truncate(d.Review.text, 100)}\"");
}
if (disagreements.Count == 0) Console.WriteLine("(none this run)");

// Same method as the starter: one prompt, one word back, any IChatClient.
static async Task<string> Classify(IChatClient client, string text)
{
    // Both models get this exact prompt, and it is byte-identical to the one in
    // lab/ollama.http and lab/azure.http, line breaks included. Reflowing these
    // four lines into one costs phi3 measured accuracy on both sets while leaving
    // llama3.2 unchanged, so varying the prompt shape and the model in the same
    // run measures nothing. See lab/expected-output.md.
    var prompt = $"""
        Classify this gear review as exactly one word: positive, negative, or mixed.
        Positive means the reviewer is happy with the product, negative means unhappy,
        mixed means genuinely both. Judge the review text only; ignore any star rating
        it mentions. Reply with only the label.

        Review: {text}
        """;

    var response = await client.GetResponseAsync(prompt, new ChatOptions { Temperature = 0 });
    var raw = response.Text.ToLowerInvariant();

    var first = new[] { "positive", "negative", "mixed" }
        .Select(l => (label: l, at: raw.IndexOf(l, StringComparison.Ordinal)))
        .Where(x => x.at >= 0)
        .OrderBy(x => x.at)
        .Select(x => x.label)
        .FirstOrDefault();
    return first ?? raw.Trim();
}

static string Truncate(string s, int max) =>
    s.Length <= max ? s : s[..max].TrimEnd() + "...";

record Review(string id, string product, int rating, string reviewer, string text);
record RefLabel(string set, string label, string? rationale = null);
record Result(Review Review, string Set, string Reference, string Small, string Big);
