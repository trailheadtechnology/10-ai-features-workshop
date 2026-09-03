# .NET Demo for 03 Sentiment

Two console projects, both built on Microsoft.Extensions.AI:

- `starter/`: the demo's starting point. One `IChatClient` over Ollama (`phi3`), one `Classify` method, one review, one word back.
- `complete/`: the finished demo. Both sets through both models, with a label table, accuracy per set per model, and the disagreement list.

From `starter/`:

```bash
dotnet run                # gr-0007, the sarcastic trail runner review
dotnet run -- gr-0034     # any id from ../../data/easy.jsonl or hard.jsonl
```

From `complete/`:

```bash
dotnet run                # both sets, both models, table + accuracy + disagreements
dotnet run -- --easy      # easy set only (demo steps 3 and 4)
dotnet run -- --hard      # hard set only (demo step 5)
```

## The Two Models

`complete/` builds two `IChatClient`s and then never mentions a provider again. The first is `phi3` on Ollama. The second is Azure OpenAI when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are all set:

```bash
export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
export AZURE_OPENAI_KEY=<KEY FROM INSTRUCTOR>
export AZURE_OPENAI_DEPLOYMENT=gpt-4.1
```

With those unset, the app prints a note and falls back to `llama3.2` on Ollama as a stand-in, so the whole comparison runs offline. That fallback is a few lines around one constructor, and everything downstream, the classify method included, is untouched. Point at those lines on stage: it is the provider-flexibility claim made real, and the same swap works in the other direction when you move to production.

Both models get a byte-identical prompt. That is deliberate and it is load-bearing. The same prompt reflowed onto one line costs `phi3` a measurable amount of accuracy while leaving `llama3.2` unchanged, so a comparison that varies prompt shape along with the model is measuring noise. See [../expected-output.md](../expected-output.md).

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F03-lab.md`](../F03-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter on the Sarcastic Review

One client, one `classify` function, one review: `gr-0007`, two stars, "five-star experience, truly". The prompt is the whole feature and it is byte-identical to `../ollama.http`; keep its line breaks, because reflowing it onto one line costs `phi3` measured accuracy.

Run:

```bash
dotnet run
```

Check: `phi3 says: negative`. Try `gr-0034` or any other id from `easy.jsonl` / `hard.jsonl`.

### Step 2: Loop the Easy Set and Score It Against the Reference Labels (lab step 1)

Replace the single review with a loop over `easy.jsonl`, look each id up in `reference-labels.json`, and count matches.

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
// record RefLabel(string set, string label, string? rationale = null);
```

Run:

```bash
dotnet run
```

Check: 9/10 on the easy set in the recorded runs. Yours may differ by one.

### Step 3: Add the Second Model and Run the Hard Set Through Both (lab step 2)

Build a second client: Azure OpenAI if you have the room key in `AZURE_OPENAI_KEY` (with `AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com` and `AZURE_OPENAI_DEPLOYMENT=gpt-4.1`), otherwise `llama3.2` on the same Ollama as a stand-in. Nothing in `classify` changes; that is the provider-swap point of the whole module. Then run `hard.jsonl` through both.

```csharp
IChatClient big = new AzureOpenAIClient(new Uri(endpoint), new ApiKeyCredential(key))
    .GetChatClient(deployment).AsIChatClient();
// or, offline: new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2")
var small = await Classify(phi3, review.text);
var bigLabel = await Classify(big, review.text);
```

Run:

```bash
dotnet run
```

Check: Two columns of labels for the hard set. Recorded: 7/10 for `phi3`, 10/10 for `gpt-4.1` on Azure, and 8/10 for the `llama3.2` stand-in. The frontier model earns its price on this slice; the local stand-in would have told you the gap is one review wide.

### Step 4: Print the Disagreement List and Call Each One (lab step 3, the success check)

Every review where the two models differ, with the reference label and a verdict on who was right. This list is the actual deliverable of the feature: it is what tells you which slice of your traffic needs the expensive model.

```csharp
foreach (var d in results.Where(r => r.Small != r.Big))
{
    var verdict = d.Big == d.Reference ? "big right" : d.Small == d.Reference ? "phi3 right" : "both wrong";
    Console.WriteLine($"{d.Review.id} [{d.Set}] ref={d.Reference} phi3={d.Small} big={d.Big}  ({verdict})");
}
```

Check: Your version of the two tables in `../expected-output.md`: accuracy per set per model, and the disagreements with your call on each. Stretch: change the label to `{overall, aspects: {comfort, durability, price}}` with structured output and see which model can go deeper.
