# .NET Demo for 03 Sentiment

Two console projects, both built on Microsoft.Extensions.AI:

- `starter/`: one `IChatClient` over Ollama (`phi3`), one `Classify` method, one review, one word back.
- `complete/`: both sets through both models, with a label table, accuracy per set per model, and the disagreement list.

From `starter/`:

```bash
dotnet run                # gr-0007, the sarcastic trail runner review
dotnet run -- gr-0034     # any id from ../../data/easy.jsonl or hard.jsonl
```

From `complete/`:

```bash
dotnet run                # both sets, both models, table + accuracy + disagreements
dotnet run -- --easy      # easy set only (lab steps 3-4 shape)
dotnet run -- --hard      # hard set only (lab step 5 shape)
```

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT` to build the second `IChatClient` against Azure OpenAI:

```bash
export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
export AZURE_OPENAI_KEY=<KEY FROM INSTRUCTOR>
export AZURE_OPENAI_DEPLOYMENT=gpt-4.1
```

Unset, the app falls back to `llama3.2` on Ollama, so the whole comparison runs offline. That fallback is a few lines around one constructor; everything downstream, `Classify` included, is untouched — point at those lines on stage, it's the provider-flexibility claim made real.

## Lab Walkthrough: From `starter/` to `complete/`

The steps are in [`../F03-lab.md`](../F03-lab.md); this maps each onto `starter/Program.cs` → `complete/Program.cs`. Edit the starter in place (or copy it first); `complete/`'s comments say why each piece is there.

### Lab step 0: run the starter

```bash
dotnet run
```

Check: `phi3 says: negative` on `gr-0007`.

### Lab steps 1-2: load reference labels, loop the easy set, score it

```csharp
var labels = JsonSerializer.Deserialize<Dictionary<string, RefLabel>>(
    await File.ReadAllTextAsync("../../data/reference-labels.json"))!;
var correct = 0; var total = 0;
foreach (var line in File.ReadLines("../../data/easy.jsonl"))
{
    var review = JsonSerializer.Deserialize<Review>(line)!;
    var label = await Classify(client, review.text);
    var reference = labels[review.id].label;
    Console.WriteLine($"{review.id,-9} {reference,-10} {label,-10}");
    total++; if (label == reference) correct++;
}
Console.WriteLine($"phi3 {correct}/{total}");
```

Check: 9/10 on the easy set in the recorded runs; yours may differ by one.

### Lab steps 3-4: second client, hard set through both models

```csharp
IChatClient big = new AzureOpenAIClient(new Uri(endpoint), new ApiKeyCredential(key))
    .GetChatClient(deployment).AsIChatClient();
// or, offline: new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2")
var small = await Classify(phi3, review.text);
var bigLabel = await Classify(big, review.text);
```

Check: 7/10 for `phi3` on the hard set, 10/10 for `gpt-4.1`, 8/10 for the `llama3.2` stand-in.

### Lab step 5: disagreement list

```csharp
foreach (var d in results.Where(r => r.Small != r.Big))
{
    var verdict = d.Big == d.Reference ? "big right" : d.Small == d.Reference ? "phi3 right" : "both wrong";
    Console.WriteLine($"{d.Review.id} [{d.Set}] ref={d.Reference} phi3={d.Small} big={d.Big}  ({verdict})");
}
```

Check: your version of the two tables in `../expected-output.md`. Stretch: change the label to `{overall, aspects: {comfort, durability, price}}` with structured output.
