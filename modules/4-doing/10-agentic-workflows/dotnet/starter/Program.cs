using Azure;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: the trip request goes to a plain chat completion.
// No tools, no data, no loop. The model has never heard of our trails,
// our weather feed, or the washed-out bridge, so what comes back is a
// fluent, generic itinerary that books nothing. Step 1 of the demo.
//
// Run: dotnet run [-- your own trip request]

IChatClient client = CreateChatClient();

var request = args.Length > 0
    ? string.Join(" ", args)
    : "Plan me a 3-day trip in Glacier National Park for September 14-16.";

Console.WriteLine($"Request: {request}");
Console.WriteLine(new string('-', 60));

var response = await client.GetResponseAsync(
    $"""
    You are the trip planner for Trailhead Guides, a hiking app.

    {request}
    """);

Console.WriteLine(response.Text);
Console.WriteLine(new string('-', 60));
Console.WriteLine("Note: zero tool calls were made. No weather was checked, no");
Console.WriteLine("conditions were read, nothing was booked. Fluent and useless.");

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
