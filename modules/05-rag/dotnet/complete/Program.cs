using System.ClientModel;
using System.Numerics.Tensors;
using System.Text.Json;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                                    the Sperry Chalet question, grounded
//   dotnet run -- "Is the Avalanche Lake Trail open right now?"
//   dotnet run -- --no-context                    step 1: the confident wrong answer
//   dotnet run -- --top-k 8 "your question"       stretch goal: vary retrieval depth
//
// Retrieval always runs locally (nomic-embed-text). Generation uses Azure OpenAI
// when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY / AZURE_OPENAI_DEPLOYMENT are set,
// and falls back to local llama3.2 when they are not.

var chunksPath = "../../lab/chunks.jsonl";
var cachePath = "embeddings.json";
var topK = 3;
var noContext = false;
var questionParts = new List<string>();
for (var i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--no-context": noContext = true; break;
        case "--top-k": topK = int.Parse(args[++i]); break;
        default: questionParts.Add(args[i]); break;
    }
}
var question = questionParts.Count > 0
    ? string.Join(" ", questionParts)
    : "Can I have a campfire at Sperry Chalet in September?";

// Generation: Azure OpenAI if configured, local llama3.2 otherwise. This is the
// one-line client swap from step 7 of the demo; everything downstream is identical.
var endpoint = Environment.GetEnvironmentVariable("AZURE_OPENAI_ENDPOINT");
var key = Environment.GetEnvironmentVariable("AZURE_OPENAI_KEY");
var deployment = Environment.GetEnvironmentVariable("AZURE_OPENAI_DEPLOYMENT");
IChatClient chatClient;
if (!string.IsNullOrEmpty(endpoint) && !string.IsNullOrEmpty(key) && !string.IsNullOrEmpty(deployment))
{
    chatClient = new AzureOpenAIClient(new Uri(endpoint), new ApiKeyCredential(key))
        .GetChatClient(deployment)
        .AsIChatClient();
    Console.WriteLine($"[generation: Azure OpenAI, deployment '{deployment}']");
}
else
{
    chatClient = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
    Console.WriteLine("[generation: AZURE_OPENAI_* not set, falling back to local llama3.2]");
}

Console.WriteLine($"Q: {question}\n");

if (noContext)
{
    // Step 1 of the demo: no retrieval, no context, pure model memory.
    Console.WriteLine((await chatClient.GetResponseAsync(question)).Text);
    return;
}

// Load the pre-chunked park docs.
var chunks = File.ReadLines(chunksPath)
    .Select(line => JsonSerializer.Deserialize<Chunk>(line)!)
    .ToList();

// Embed the chunks with local nomic-embed-text, caching to embeddings.json so
// only the first run pays the ~40 seconds of embedding time.
IEmbeddingGenerator<string, Embedding<float>> embedder =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");

Dictionary<string, float[]> index;
if (File.Exists(cachePath))
{
    index = JsonSerializer.Deserialize<Dictionary<string, float[]>>(File.ReadAllText(cachePath))!;
}
else
{
    Console.WriteLine($"[embedding {chunks.Count} chunks, one-time; caching to {cachePath}]");
    index = new Dictionary<string, float[]>();
    foreach (var batch in chunks.Chunk(32))
    {
        var embeddings = await embedder.GenerateAsync(batch.Select(c => c.text));
        foreach (var (chunk, embedding) in batch.Zip(embeddings))
            index[chunk.chunk_id] = embedding.Vector.ToArray();
    }
    File.WriteAllText(cachePath, JsonSerializer.Serialize(index));
}

// Retrieval: embed the question, cosine similarity against every chunk, take top-k.
// This is module 04's search, verbatim, pointed at the chunk index.
var questionVector = (await embedder.GenerateAsync(question)).Vector.ToArray();
var top = chunks
    .Select(c => (chunk: c, score: TensorPrimitives.CosineSimilarity(questionVector, index[c.chunk_id])))
    .OrderByDescending(x => x.score)
    .Take(topK)
    .ToList();

Console.WriteLine($"[retrieved top {topK}]");
foreach (var (chunk, score) in top)
    Console.WriteLine($"  {score:F4}  {chunk.chunk_id}");
Console.WriteLine();

// Grounded prompt: context in, citations out, refusal when the context is silent.
var context = string.Join("\n\n", top.Select(x =>
    $"chunk_id: {x.chunk.chunk_id}\nsource: {x.chunk.source}\n{x.chunk.text}"));

var prompt = $"""
    You are a park information assistant. Answer the visitor's question using ONLY the context below.
    Rules:
    - Base every statement on the context. Do not use outside knowledge.
    - Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
    - If, and only if, none of the context is relevant to the question, reply exactly: "The provided documents don't say."

    Context:
    {context}

    Question: {question}

    Answer:
    """;

var response = await chatClient.GetResponseAsync(prompt);
Console.WriteLine(response.Text);

record Chunk(string chunk_id, string source, string text);
