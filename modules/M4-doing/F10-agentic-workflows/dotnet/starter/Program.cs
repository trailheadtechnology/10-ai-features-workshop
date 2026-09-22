using Azure;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// The trip request goes to a plain chat completion, with no tools and no loop.
// Nothing here can reach the trail catalog, the weather feed, or the condition
// reports, so the itinerary that comes back is fluent, generic, and books
// nothing. That gap is the point of this project; the fix lives in ../complete.
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
    var endpoint = "https://trailhead-ai-workshop.openai.azure.com";
    var key = "<KEY FROM INSTRUCTOR>";  // paste the room key between the quotes
    var deployment = "gpt-5.5";

    if (!key.StartsWith("<"))
    {
        return new AzureOpenAIClient(new Uri(endpoint), new AzureKeyCredential(key))
            .GetChatClient(deployment)
            .AsIChatClient();
    }

    Console.WriteLine("[note] no room key pasted in; falling back to Ollama llama3.2.");
    return new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
}
