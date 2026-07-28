using System.ClientModel;
using System.Numerics.Tensors;
using System.Text.Json;
using System.Text.RegularExpressions;
using Azure.AI.OpenAI;
using Microsoft.Extensions.AI;
using OllamaSharp;

// Finished demo, matching the outline in ../../README.md:
//   dotnet run                                    the Sperry Chalet question, grounded
//   dotnet run -- "Is the Avalanche Lake Trail open right now?"
//   dotnet run -- --no-context                    step 1: the confident wrong answer
//   dotnet run -- --alpha 1.0                     pure cosine, the wrong-park neighbors
//   dotnet run -- --retrieval-only                print the score table and stop
//   dotnet run -- --top-k 8 "your question"       vary retrieval depth
//
// Retrieval is hybrid: normalized cosine similarity blended with a BM25-lite
// lexical score, so a distinctive proper noun like "Sperry" counts for something.
// Chunks are one numbered subsection each wherever a section ran long enough to
// hold a rule and the exception that overrides it; see ../../lab/expected-output.md.
// Every generated answer's [chunk-id] citations are validated against the chunks
// that were actually retrieved.
//
// Retrieval always runs locally (nomic-embed-text). Generation uses Azure OpenAI
// when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY / AZURE_OPENAI_DEPLOYMENT are set,
// and falls back to local llama3.2 when they are not.

var chunksPath = "../../lab/chunks.jsonl";
var cachePath = "embeddings.json";
var topK = 3;
var alpha = 0.6;          // weight on the semantic signal; 1.0 = cosine only
var noContext = false;
var retrievalOnly = false;
var questionParts = new List<string>();
for (var i = 0; i < args.Length; i++)
{
    switch (args[i])
    {
        case "--no-context": noContext = true; break;
        case "--retrieval-only": retrievalOnly = true; break;
        case "--top-k": topK = int.Parse(args[++i]); break;
        case "--alpha": alpha = double.Parse(args[++i]); break;
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

// ---------------------------------------------------------------------------
// Signal 1: semantic. Module 04's cosine search, verbatim, over the chunk index.
// ---------------------------------------------------------------------------
var questionVector = (await embedder.GenerateAsync(question)).Vector.ToArray();
var cosine = chunks.ToDictionary(
    c => c.chunk_id,
    c => (double)TensorPrimitives.CosineSimilarity(questionVector, index[c.chunk_id]));

// ---------------------------------------------------------------------------
// Signal 2: lexical. BM25-lite over the chunk text. The embedder collapses
// "campfire regulations" from five different parks onto nearly the same point;
// the word "Sperry" appears in exactly one document, and IDF makes that count.
// ---------------------------------------------------------------------------
var tokenized = chunks.ToDictionary(c => c.chunk_id, c => Tokenize(c.text));
var avgLength = tokenized.Values.Average(t => (double)t.Count);
var docFreq = new Dictionary<string, int>();
foreach (var terms in tokenized.Values)
    foreach (var term in terms.Distinct())
        docFreq[term] = docFreq.GetValueOrDefault(term) + 1;

const double K1 = 1.2, B = 0.3;
var n = chunks.Count;
var queryTerms = Tokenize(question).Distinct().ToList();
// IDF: a term in 1 of 250 chunks is worth far more than one in 200 of them.
var idf = queryTerms.ToDictionary(
    t => t,
    t => Math.Log(1 + (n - docFreq.GetValueOrDefault(t) + 0.5) / (docFreq.GetValueOrDefault(t) + 0.5)));

var lexical = new Dictionary<string, double>();
foreach (var c in chunks)
{
    var terms = tokenized[c.chunk_id];
    var counts = terms.GroupBy(t => t).ToDictionary(g => g.Key, g => (double)g.Count());
    var score = 0.0;
    foreach (var t in queryTerms)
    {
        if (!counts.TryGetValue(t, out var tf)) continue;
        score += idf[t] * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * terms.Count / avgLength));
    }
    lexical[c.chunk_id] = score;
}

// Rescale both signals to 0..1 across the corpus for this question, so alpha
// means what it looks like it means and the two numbers are comparable on screen.
var semanticNorm = MinMax(cosine);
var lexicalNorm = MinMax(lexical);

var scored = chunks
    .Select(c => new Hit(
        c,
        cosine[c.chunk_id],
        semanticNorm[c.chunk_id],
        lexical[c.chunk_id],
        lexicalNorm[c.chunk_id],
        alpha * semanticNorm[c.chunk_id] + (1 - alpha) * lexicalNorm[c.chunk_id]))
    .OrderByDescending(h => h.Combined)
    .ToList();
var top = scored.Take(topK).ToList();

var distinctive = string.Join("  ", queryTerms
    .OrderByDescending(t => idf[t])
    .Take(5)
    .Select(t => $"{t}({idf[t]:F2}/{docFreq.GetValueOrDefault(t)} chunks)"));
Console.WriteLine($"[query terms by IDF]  {distinctive}");
Console.WriteLine($"[retrieved top {topK}]  combined = {alpha:F2} * semantic + {1 - alpha:F2} * lexical");
Console.WriteLine("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id");
for (var r = 0; r < top.Count; r++)
{
    var h = top[r];
    Console.WriteLine($"  {r + 1,4}  {h.Combined,8:F4}   {h.SemanticNorm,5:F3} ({h.Cosine:F4})   " +
                      $"{h.LexicalNorm,5:F3} ({h.Lexical,5:F2})   {h.Chunk.chunk_id}");
}
var margin = scored.Count > 1 ? scored[0].Combined - scored[1].Combined : 0;
Console.WriteLine($"  margin over rank 2: {margin:F4}\n");

if (retrievalOnly) return;

// Grounded prompt: context in, citations out, refusal when the context is silent.
//
// The corpus is full of dated notices ("CLOSED effective June 20, 2026, until further
// notice"). A model with no idea what day it is treats "is the trail open right now?" as
// a question its documents cannot speak to, and refuses. So we tell it the date. Two
// measured caveats, both worth a sentence on stage: the date on its own changes nothing
// (6/16 refusals with it, 5/16 without), and a broadly worded currency rule in the Rules
// block fixes this question while wrecking Q1, where the model starts applying effective
// dates to a year-round fire ban. Attaching the date to the refusal clause, where the
// refusal decision is actually made, is the version that helps without collateral damage.
// Production passes DateTime.Today here; this demo pins a date so the recorded outputs in
// lab/expected-output.md stay reproducible.
const string today = "September 23, 2026";
const string refusal = "The provided documents don't say.";

var retrievedIds = top.Select(x => x.Chunk.chunk_id).ToHashSet();
var context = string.Join("\n\n", top.Select(x =>
    $"chunk_id: {x.Chunk.chunk_id}\nsource: {x.Chunk.source}\n{x.Chunk.text}"));

var prompt = $"""
    You are a park information assistant. Answer the visitor's question using ONLY the context below.
    Rules:
    - Base every statement on the context. Do not use outside knowledge.
    - Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
    - Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
      and do not combine parts of two chunk_ids.
    - If, and only if, none of the context is relevant to the question, reply exactly: "{refusal}"
      A question about "right now" is answered from the context, not refused: today is {today},
      and a notice that is in effect "until further notice" is still in effect right now.

    Context:
    {context}

    Question: {question}

    Answer:
    """;

var answer = (await chatClient.GetResponseAsync(prompt)).Text;
var bad = InvalidCitations(answer, retrievedIds);

// Citation validation. An invalid citation is a product defect, not a style
// issue: it is a receipt pointing at a document nobody retrieved, and sometimes
// at a document that does not exist. Retry once with the valid ids spelled out,
// then strip whatever is still wrong so a bad receipt never reaches the visitor.
if (bad.Count > 0 && answer.Contains(refusal))
{
    // A refusal with a chunk_id stapled to it is the most common invalid citation in the
    // whole demo, and it is not a question the model needs to think about again: the answer
    // is already correct and by definition it has no sources. Sending it back through the
    // model was measurably harmful. Asked to "rewrite the answer using only those ids", it
    // rewrites the refusal too, and the exact wording the product depends on comes back as
    // "There is no information about EV charging stations in the provided context." Truthful,
    // still refusing, but no longer the string anything downstream can match on. Deleting a
    // citation is a string operation. Do it in code and skip the round trip.
    Console.WriteLine($"!! CITATION CHECK FAILED: {string.Join(", ", bad.Select(c => $"[{c}]"))} not in the retrieved set");
    Console.WriteLine("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed\n");
    answer = refusal;
    bad = [];
}
else if (bad.Count > 0)
{
    Console.WriteLine($"!! CITATION CHECK FAILED: {string.Join(", ", bad.Select(c => $"[{c}]"))} not in the retrieved set");
    Console.WriteLine("!! retrying once with the valid chunk_ids spelled out\n");

    var retryPrompt = prompt + $"""


        Your previous answer cited {string.Join(", ", bad.Select(c => $"[{c}]"))}, which is not a real chunk_id.
        The only chunk_ids you may cite are, exactly:
        {string.Join("\n", retrievedIds.Select(id => "  " + id))}
        Rewrite the answer using only those.
        """;
    answer = (await chatClient.GetResponseAsync(retryPrompt)).Text;
    bad = InvalidCitations(answer, retrievedIds);

    if (bad.Count > 0)
    {
        Console.WriteLine($"!! STILL INVALID after retry: {string.Join(", ", bad.Select(c => $"[{c}]"))}");
        Console.WriteLine("!! stripping them; the answer below is unverified where the citation was removed\n");
        foreach (var c in bad)
            answer = answer.Replace(c, "invalid-citation-removed");
    }
}

Console.WriteLine(answer);

var cited = Citations(answer).Where(retrievedIds.Contains).Distinct().ToList();
Console.WriteLine($"\n[citations: {cited.Count} valid ({string.Join(", ", cited)}), {bad.Count} invalid]");

// Any bracketed token containing a colon is a citation attempt, including the
// comma-separated lists llama3.2 sometimes writes. Being generous about what
// counts as an attempt is the point: we want to catch the near-misses.
static List<string> Citations(string text) =>
    Regex.Matches(text, @"\[([^\]]*:[^\]]*)\]")
        .SelectMany(m => m.Groups[1].Value.Split(','))
        .Select(c => c.Trim())
        .Where(c => c.Contains(':'))
        .ToList();

static List<string> InvalidCitations(string text, HashSet<string> valid) =>
    Citations(text).Where(c => !valid.Contains(c)).Distinct().ToList();

// Lowercase, split on non-alphanumerics, drop question filler, and knock the
// plural off so "campfires" in a document matches "campfire" in a question.
// Crude on purpose: an attendee can read all of it in ten seconds.
static List<string> Tokenize(string text) =>
    Regex.Matches(text.ToLowerInvariant(), @"[a-z0-9]+")
        .Select(m => m.Value)
        .Where(t => t.Length > 2 && !StopWords.Contains(t))
        .Select(t => t.Length > 3 && t.EndsWith('s') && !t.EndsWith("ss") ? t[..^1] : t)
        .ToList();

static Dictionary<string, double> MinMax(Dictionary<string, double> raw)
{
    var min = raw.Values.Min();
    var max = raw.Values.Max();
    var range = max - min;
    return raw.ToDictionary(kv => kv.Key, kv => range > 1e-9 ? (kv.Value - min) / range : 0.0);
}

// Question filler. Without this, "can" and "have" carry as much weight as "Sperry"
// simply because no park document says "can I have".
static class StopWords
{
    static readonly HashSet<string> Words =
    [
        "the", "and", "for", "are", "but", "not", "you", "your", "with", "that", "this", "these",
        "those", "from", "have", "has", "had", "was", "were", "been", "being", "can", "could",
        "will", "would", "shall", "should", "may", "might", "must", "does", "did", "doing",
        "what", "when", "where", "which", "who", "whom", "why", "how", "any", "all", "some",
        "there", "here", "then", "than", "them", "they", "their", "its", "his", "her", "our",
        "get", "got", "still", "now", "right", "just", "about", "into", "onto", "over", "under",
        "out", "off", "per", "via", "one", "two", "also", "more", "most", "much", "many", "each",
        "other", "such", "only", "own", "same", "too", "very", "let", "need", "want",
    ];

    public static bool Contains(string term) => Words.Contains(term);
}

record Chunk(string chunk_id, string source, string text);
record Hit(Chunk Chunk, double Cosine, double SemanticNorm, double Lexical, double LexicalNorm, double Combined);
