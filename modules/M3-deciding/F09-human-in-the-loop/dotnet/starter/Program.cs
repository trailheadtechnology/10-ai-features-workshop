using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Starting point, and deliberately unsafe. Every draft goes straight out with no
// human between the model and the visitor, and nothing is logged, so there is no
// record of what was sent or any way to find out later.
// Run: dotnet run [path-to-inquiries.jsonl]
//
// Nothing here treats an emergency differently. The system prompt asks the model
// to escalate instead of drafting, and the model is free to ignore that and
// often does, which is why complete/ moves the decision out of the prompt and
// into a policy check in code. Do not use this shape on a real inbox.

const string SystemPrompt = """
    You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides.
    A human ranger reviews your draft before anything is sent, so write it ready to send:
    friendly, plain, professional, at most two short paragraphs, signed
    'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure,
    state the rule and cite the source document number and section (for example
    GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided;
    if the excerpt does not answer the question, say a ranger will follow up with
    specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
    Exception: if the visitor's message reports an emergency, an injury, a possible fire,
    or a missing or overdue person, do not draft a reply at all. Output exactly one line
    beginning with ESCALATE: followed by a one-line reason, so the message goes straight
    to dispatch.
    """;

var inquiriesPath = args.Length > 0 ? args[0] : "../../lab/inquiries.jsonl";
var labDir = Path.GetDirectoryName(Path.GetFullPath(inquiriesPath))!;

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");

foreach (var line in await File.ReadAllLinesAsync(inquiriesPath))
{
    if (string.IsNullOrWhiteSpace(line)) continue;
    var inquiry = JsonSerializer.Deserialize<Inquiry>(line)!;

    var snippetPath = Path.Combine(labDir, "snippets", inquiry.Doc);
    var snippet = !string.IsNullOrEmpty(inquiry.Doc) && File.Exists(snippetPath)
        ? (await File.ReadAllTextAsync(snippetPath)).Trim()
        : "(none on file for this message)";

    var draft = await client.GetResponseAsync(
    [
        new ChatMessage(ChatRole.System, SystemPrompt),
        new ChatMessage(ChatRole.User, $"""
            Reference excerpt:
            {snippet}

            Visitor message ({inquiry.Channel}, received {inquiry.Received}):
            {inquiry.Text}

            Draft the reply.
            """),
    ]);

    // No review step and no audit record. In a real deployment this line is the
    // send call, and by the time anyone reads the output the mail has gone.
    Console.WriteLine($"=== SENT to visitor · {inquiry.Id} ({inquiry.Category}) ===");
    Console.WriteLine(draft.Text.Trim());
    Console.WriteLine();
}

Console.WriteLine("All replies sent. Nobody read them. Nothing was logged.");

record Inquiry(
    [property: JsonPropertyName("id")] string Id,
    [property: JsonPropertyName("channel")] string Channel,
    [property: JsonPropertyName("received")] string Received,
    [property: JsonPropertyName("category")] string Category,
    [property: JsonPropertyName("doc")] string Doc,
    [property: JsonPropertyName("text")] string Text);
