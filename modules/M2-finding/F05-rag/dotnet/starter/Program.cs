using Microsoft.Extensions.AI;
using OllamaSharp;

// Demo starting point: one chat client, one question, no context.
// This is the "plain chatbot" from step 1 of the demo script. It answers
// confidently. It is also wrong: the model puts Sperry Chalet in California
// and guesses at fire rules the park wrote down years ago.
// Run: dotnet run [-- "your question"]

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

var question = args.Length > 0
    ? string.Join(" ", args)
    : "Can I have a campfire at Sperry Chalet in September?";

Console.WriteLine($"Q: {question}\n");
var response = await client.GetResponseAsync(question);
Console.WriteLine(response.Text);
