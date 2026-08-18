# .NET Demo for 06 Recommendations

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. It loads the trail catalog and fills the "you might also like" box with five random trails, which is roughly what that box does in most apps today and takes no model at all.
- `complete/`: the finished demo as shown on stage. It embeds the 30 trail descriptions through `IEmbeddingGenerator`, ranks every other trail by cosine similarity to the one you name, skips the trail itself, and prints the top 5 with scores.

Both run against Ollama (`nomic-embed-text`), matching the demo script in [docs/slides/outlines/M2-finding.md](../../../../docs/slides/outlines/M2-finding.md). From `complete/`:

```bash
dotnet run                             # more like this, for Avalanche Lake Trail
dotnet run -- trail-0008               # any trail id from ../../data/trails.json
dotnet run -- Trail of the Cedars      # or any name, or part of one
dotnet run -- --gear Cascade 65        # step 5: same trick over gear review text
```

Vectors land in `complete/embeddings.json` and `complete/gear-embeddings.json` on the first run, so every run after that is instant. Delete a cache file to re-embed.

`Cosine` is a dozen lines at the bottom of `complete/Program.cs`, and the recommendation itself is the LINQ query above it. That is the demo's point: this is the feature 04 search code with the query text swapped for an item's vector.

Real output for all four commands, including the neighbors that are obviously wrong, is in [../expected-output.md](../expected-output.md).

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F06-lab.md`](../F06-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter: The Random Carousel

This is the recommendation feature most apps ship: five trails picked at random under "you might also like". Run it twice for the same trail and get two different lists.

Run:

```bash
dotnet run
```

Check: Nothing about the five relates to Avalanche Lake Trail.

### Step 2: Get a Vector for Every Trail (lab step 1, first half)

If you did feature 04, this is the same code and the same model. If you did not, `../data/trail-embeddings.json` has the 30 vectors precomputed (keyed by trail id, embedded from `description`) and you can load them instead of calling the model at all.

```csharp
// Option A: embed live (same as feature 04)
var embeddings = await generator.GenerateAsync(trails.Select(t => t.Description));
var vectors = trails.Zip(embeddings).ToDictionary(p => p.First.Id, p => p.Second.Vector.ToArray());
// Option B: precomputed
var vectors = JsonSerializer.Deserialize<Dictionary<string, float[]>>(
    await File.ReadAllTextAsync("../../data/trail-embeddings.json"))!;
```

Check: Whichever way, `vectors["trail-0117"]` is 768 floats. If you loaded the precomputed file, check its top-level shape first; it may wrap the vectors in an object.

### Step 3: Rank Every Other Trail by Similarity to the Target (lab step 1, second half)

"More like this" is feature 04's search with the query vector replaced by the target trail's own vector. Skip the target itself, take five.

```csharp
var hits = trails
    .Where(t => t.Id != target.Id)
    .Select(t => (Trail: t, Score: Cosine(vectors[target.Id], vectors[t.Id])))
    .OrderByDescending(h => h.Score).Take(5);
foreach (var (trail, score) in hits)
    Console.WriteLine($"  {score:F4}  {trail.Name} ({trail.Park}, {trail.Difficulty}; {string.Join(", ", trail.Features)})");
```

Run:

```bash
dotnet run
```

Check: Gunsight Lake Approach at 0.7849 on top for Avalanche Lake Trail. Read the difficulty column: the target is a moderate family walk and most neighbors are hard. Difficulty is not in the description text, so the embedding cannot see it.

### Step 4: Do the Other Two Targets and Judge Whether You Would Ship Them (lab steps 2 and 3)

Run the other targets from `../F06-lab.md`, compare against the acceptable sets in `../expected-output.md` (there is more than one right answer), and then read your own output as a product owner. One target in this slice has no real neighbors at all; a shipping product should show nothing rather than five weak guesses.

Run:

```bash
dotnet run -- trail-0008
dotnet run -- Trail of the Cedars
```

Check: Substantial overlap with the acceptable sets, and a sentence from you on whether you would ship each list. Stretch: average two trails' vectors and rank against the average, or filter to the same park or an easier difficulty before ranking. `complete/` also has `--gear`, where the top hit for the Cascade 65 pack is the Cascade 40 pack: substitutes, not complements.
