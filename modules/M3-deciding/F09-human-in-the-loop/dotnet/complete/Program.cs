using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Review queue: the model drafts, a human decides, and every decision is logged
// to decisions.jsonl.
//
//   dotnet run                          review the queue: [a]pprove / [e]dit / [r]eject / [s]kip
//   dotnet run -- --policy              print the routing policy table and exit
//   dotnet run -- --auto-approve-dry-run   non-interactive run for testing and CI
//   dotnet run -- ../../lab/inquiries.jsonl   any queue file works
//
// SAFETY INVARIANT, load-bearing, do not weaken:
// emergencies never reach the model. The policy table below routes them to
// human-only and the loop skips the API call entirely, in code, before any
// request is built. The system prompt also tells the model to escalate instead
// of drafting, but that instruction is a request and this lane is a guarantee.
// A model that ignores the instruction and writes a warm, fluent, confident
// reply to someone reporting an overdue hiker is not a hypothetical; it is the
// documented behavior of the model this demo ships with. Anyone editing this
// file must keep the emergency path free of model calls. Adding a "just draft
// it and let the reviewer catch it" shortcut here puts a reassuring lie in front
// of a person who needed a dispatcher.

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

// The lane is chosen by what a wrong answer costs, not by how good the model is
// at the category. Everything reversible can be drafted; emergency is
// irreversible and stays human-only. Note the lookup below defaults an unknown
// category to human-only, so a category added upstream fails closed rather than
// quietly acquiring a draft lane.
var policy = new Dictionary<string, string>
{
    ["trail-condition"] = "draft-for-approval",
    ["permit"] = "draft-for-approval",
    ["complaint"] = "draft-for-approval",
    ["general"] = "draft-for-approval",
    ["lost-and-found"] = "draft-for-approval",
    ["emergency"] = "human-only",
};

var inquiriesPath = "../../lab/inquiries.jsonl";
var outboxDir = "outbox";
var decisionsPath = "decisions.jsonl";
var autoApprove = false;

for (var i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--auto-approve-dry-run": autoApprove = true; break;
        case "--outbox": outboxDir = args[++i]; break;
        case "--decisions": decisionsPath = args[++i]; break;
        case "--policy": PrintPolicy(policy); return;
        default: inquiriesPath = args[i]; break;
    }
}

var labDir = Path.GetDirectoryName(Path.GetFullPath(inquiriesPath))!;
Directory.CreateDirectory(outboxDir);

IChatClient client = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
var reviewer = autoApprove ? "auto-approve-dry-run" : Environment.UserName;
var counts = new Dictionary<string, int>();

PrintPolicy(policy);
if (autoApprove)
    Console.WriteLine("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.\n");

foreach (var line in await File.ReadAllLinesAsync(inquiriesPath))
{
    if (string.IsNullOrWhiteSpace(line)) continue;
    var inquiry = JsonSerializer.Deserialize<Inquiry>(line)!;
    var lane = policy.GetValueOrDefault(inquiry.Category, "human-only");

    Console.WriteLine(new string('-', 72));
    Console.WriteLine($"{inquiry.Id}  ·  {inquiry.Category}  ·  {inquiry.Channel}  ·  lane: {lane}");
    Console.WriteLine(new string('-', 72));
    Console.WriteLine(Indent(inquiry.Text));
    Console.WriteLine();

    // THE GATE. This must stay above the API call, and the API call must stay
    // below it. A human-only message is escalated and logged without a single
    // token being spent on it, so there is no draft to leak, no reviewer fatigue
    // to survive, and no sampling luck involved. The ESCALATE handling further
    // down is a backstop for emergencies that arrive miscategorized; it is never
    // the control, because it runs after the model has already had its say.
    if (lane == "human-only")
    {
        Console.WriteLine("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.\n");
        await LogAsync(decisionsPath, new Decision(DateTimeOffset.UtcNow, inquiry.Id, inquiry.Category,
            lane, "escalated", reviewer, null, null, 0));
        Bump(counts, "escalated");
        continue;
    }

    var snippetPath = Path.Combine(labDir, "snippets", inquiry.Doc);
    var snippet = !string.IsNullOrEmpty(inquiry.Doc) && File.Exists(snippetPath)
        ? (await File.ReadAllTextAsync(snippetPath)).Trim()
        : "(none on file for this message)";

    Console.Write("  drafting...");
    var response = await client.GetResponseAsync(
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
    var draft = response.Text.Trim();
    Console.WriteLine("\r  draft:      \n");
    Console.WriteLine(Indent(draft));
    Console.WriteLine();

    // Second layer, for an emergency that reached here under the wrong category.
    // An ESCALATE prefix is a hard stop: the draft is logged for the audit trail
    // but never offered for approval, because a reviewer presented with a
    // sendable-looking reply may send it. Do not soften this into a warning.
    if (draft.StartsWith("ESCALATE", StringComparison.OrdinalIgnoreCase))
    {
        Console.WriteLine("  Model asked to escalate. Draft discarded, routing to a human.\n");
        await LogAsync(decisionsPath, new Decision(DateTimeOffset.UtcNow, inquiry.Id, inquiry.Category,
            lane, "escalated", reviewer, draft, null, 0));
        Bump(counts, "escalated");
        continue;
    }

    string decision;
    string? final = null;

    if (autoApprove)
    {
        decision = "approved";
        final = draft;
        Console.WriteLine("  [auto] approved\n");
    }
    else
    {
        Console.Write("  [a]pprove  [e]dit  [r]eject  [s]kip > ");
        var key = (Console.ReadLine() ?? "s").Trim().ToLowerInvariant();
        Console.WriteLine();
        switch (key)
        {
            case "a":
                decision = "approved";
                final = draft;
                break;
            case "e":
                decision = "edited";
                final = ReadEdited(draft);
                break;
            case "r":
                decision = "rejected";
                break;
            default:
                decision = "skipped";
                break;
        }
    }

    if (final is not null)
    {
        var path = Path.Combine(outboxDir, $"{inquiry.Id}.txt");
        await File.WriteAllTextAsync(path, final + Environment.NewLine);
        Console.WriteLine($"  -> {decision}, queued at {path}\n");
    }
    else
    {
        Console.WriteLine($"  -> {decision}, nothing queued\n");
    }

    await LogAsync(decisionsPath, new Decision(DateTimeOffset.UtcNow, inquiry.Id, inquiry.Category,
        lane, decision, reviewer, draft, final, EditDistance(draft, final ?? "")));
    Bump(counts, decision);
}

Console.WriteLine(new string('=', 72));
Console.WriteLine("Queue done: " + string.Join(", ", counts.Select(c => $"{c.Value} {c.Key}")));
Console.WriteLine($"Audit trail: {decisionsPath}   ·   Outbox: {outboxDir}/");

static void PrintPolicy(Dictionary<string, string> policy)
{
    Console.WriteLine("Routing policy (error cost decides the lane):");
    foreach (var (category, lane) in policy)
        Console.WriteLine($"  {category,-16} {lane}");
    Console.WriteLine();
}

static string Indent(string text) =>
    string.Join(Environment.NewLine, text.Split('\n').Select(l => "  | " + l.TrimEnd()));

static string ReadEdited(string draft)
{
    Console.WriteLine("  Type the reply you want to send. End with a single '.' on its own line.");
    Console.WriteLine("  Press Enter on the first line to start from the draft text instead.\n");
    var sb = new StringBuilder();
    var first = true;
    while (true)
    {
        var line = Console.ReadLine();
        if (line is null || line == ".") break;
        if (first && line.Length == 0)
        {
            sb.AppendLine(draft);
            Console.WriteLine("  (draft copied in; keep typing to append, '.' to finish)");
        }
        else
        {
            sb.AppendLine(line);
        }
        first = false;
    }
    var edited = sb.ToString().Trim();
    return edited.Length == 0 ? draft : edited;
}

static async Task LogAsync(string path, Decision decision) =>
    await File.AppendAllTextAsync(path, JsonSerializer.Serialize(decision) + Environment.NewLine);

static void Bump(Dictionary<string, int> counts, string key) =>
    counts[key] = counts.GetValueOrDefault(key) + 1;

// Logged on every decision so the promotion question has data behind it rather
// than a feeling. It measures how much someone typed, not whether they were
// fixing a comma or preventing a lawsuit, so it can support an argument for
// promoting a lane and must never be the only evidence for one.
// O(a*b) and unbounded by draft length; fine for a review queue, not for bulk.
static int EditDistance(string a, string b)
{
    var previous = new int[b.Length + 1];
    var current = new int[b.Length + 1];
    for (var j = 0; j <= b.Length; j++) previous[j] = j;
    for (var i = 1; i <= a.Length; i++)
    {
        current[0] = i;
        for (var j = 1; j <= b.Length; j++)
        {
            var cost = a[i - 1] == b[j - 1] ? 0 : 1;
            current[j] = Math.Min(Math.Min(current[j - 1] + 1, previous[j] + 1), previous[j - 1] + cost);
        }
        (previous, current) = (current, previous);
    }
    return previous[b.Length];
}

record Inquiry(
    [property: JsonPropertyName("id")] string Id,
    [property: JsonPropertyName("channel")] string Channel,
    [property: JsonPropertyName("received")] string Received,
    [property: JsonPropertyName("category")] string Category,
    [property: JsonPropertyName("doc")] string Doc,
    [property: JsonPropertyName("text")] string Text);

record Decision(
    [property: JsonPropertyName("at")] DateTimeOffset At,
    [property: JsonPropertyName("inquiryId")] string InquiryId,
    [property: JsonPropertyName("category")] string Category,
    [property: JsonPropertyName("lane")] string Lane,
    [property: JsonPropertyName("decision")] string Value,
    [property: JsonPropertyName("reviewer")] string Reviewer,
    [property: JsonPropertyName("draft")] string? Draft,
    [property: JsonPropertyName("final")] string? Final,
    [property: JsonPropertyName("editDistance")] int EditDistance);
