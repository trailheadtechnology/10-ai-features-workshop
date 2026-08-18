# .NET Demo for 04 Semantic Search

Two console projects, both named `Search`, both reading the 30-trail slice in [`../data/trails-slice.json`](../data/trails-slice.json):

- `starter/`: the demo's starting point, and today's baseline. Keyword substring search, no AI at all: lowercase the query, keep words of three letters or more, count whole-word hits in each trail's name and description. It is 40 lines, and it is genuinely how most search boxes work.
- `complete/`: the finished demo as shown on stage. Microsoft.Extensions.AI's `IEmbeddingGenerator` over OllamaSharp, `nomic-embed-text`, cosine similarity in one visible method, top 5 with scores.

Both take the query as arguments and default to the demo query. From either directory:

```bash
dotnet run                                       # dog-friendly waterfall hike, not too steep
dotnet run -- somewhere quiet to take my kids
dotnet run -- an easy hike to a great view
```

Run `starter` first, get junk, then run `complete` on the same words. The 30 descriptions embed in under two seconds on a laptop, and `complete` caches the vectors to `embeddings.json` next to the binary, so only the first run pays even that. Delete that file to re-embed live and show the timing. Real output from both projects is recorded in [`../expected-output.md`](../expected-output.md).

**Starting point:** J.'s existing talk demo [trailheadtechnology/dotnet-semantic-search](https://github.com/trailheadtechnology/dotnet-semantic-search) ("Warm and Fuzzy: Semantic Search in .NET") already covers this feature's stack: Microsoft.Extensions.AI, Ollama, and vectorization. The build-out here adapts that code to the Trailhead Guides trail catalog instead of starting from scratch.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F04-lab.md`](../F04-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Keyword Baseline

The starter is today's search box: lowercase, split into words, count whole-word hits. Run the demo query and then the kids query. This is what you are beating.

Run:

```bash
dotnet run
dotnet run -- somewhere quiet to take my kids
```

Check: Junk for the first, one trail ("kids") for the second. Compare the keyword blocks in `../expected-output.md`.

### Step 2: Embed the 30 Descriptions Once (lab step 1)

Replace the keyword scoring with an embedding client and embed every trail's description. `nomic-embed-text` returns a 768-float vector per text; keep them in a dictionary keyed by trail id. Time it: it is seconds, and it happens once.

```csharp
IEmbeddingGenerator<string, Embedding<float>> generator =
    new OllamaApiClient(new Uri("http://localhost:11434"), "nomic-embed-text");
var embeddings = await generator.GenerateAsync(trails.Select(t => t.Description));
var vectors = trails.Zip(embeddings).ToDictionary(p => p.First.Id, p => p.Second.Vector.ToArray());
```

Check: 30 vectors of 768 floats. Print one and look at it: it is just numbers. `complete/` caches them to `embeddings.json`; keep the cache keyed by id and delete it if the text or the model changes.

### Step 3: Embed the Query, Write Cosine Similarity, Print the Top 5 (lab step 2)

The query goes through the same model as the catalog (vectors from two models are not comparable, and cosine will still return confident numbers if you mix them). Cosine similarity fits in one visible function.

```csharp
var queryVector = (await generator.GenerateAsync([query]))[0].Vector.ToArray();

static float CosineSimilarity(float[] a, float[] b)
{
    float dot = 0, magA = 0, magB = 0;
    for (var i = 0; i < a.Length; i++) { dot += a[i] * b[i]; magA += a[i] * a[i]; magB += b[i] * b[i]; }
    return dot / (MathF.Sqrt(magA) * MathF.Sqrt(magB));
}

var results = trails
    .Select(t => (Trail: t, Score: CosineSimilarity(queryVector, vectors[t.Id])))
    .OrderByDescending(r => r.Score).Take(5);
foreach (var (trail, score) in results)
    Console.WriteLine($"{score:F4}  {trail.Id}  {trail.Name} ({trail.Difficulty}, {trail.DistanceMi} mi)");
```

Run:

```bash
dotnet run
```

Check: The gentle shaded waterfall trails at the top for the demo query, scores around 0.77. Compare `../expected-output.md`.

### Step 4: Run the Other Two Queries and Read the Scores, Not Just the Order (lab step 3)

The three test queries and their expected top hits are in `../data/queries.json`. The kids query is the one to sit with: the top hit is Taft Point, a cliff edge, at 0.4876. Perfect topical match, terrible advice, and the score is a third lower than query one's.

Run:

```bash
dotnet run -- somewhere quiet to take my kids
dotnet run -- an easy hike to a great view
```

Check: The expected trail is in your top 3 for each query. Stretch: filter on the metadata you already have (`features` contains `dog-friendly`, `distance_mi < 6`) before ranking, or blend the keyword count into the score. Either is a few lines, and either fixes Taft Point in a way no better model would.
