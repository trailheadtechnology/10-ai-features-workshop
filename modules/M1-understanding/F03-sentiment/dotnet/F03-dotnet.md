<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 03: Sentiment (.NET)

*You are on the .NET track. Other tracks: [Python](../python/F03-python.md), [TypeScript](../typescript/F03-typescript.md). Lab overview: [F03-lab.md](../F03-lab.md).*

**The User Problem:** Trailhead Guides sells gear, and the Cascade 65 backpack has 300 reviews. The product team asks a simple question: are people happy with it, and what are they mad about? Star ratings lie: "4 stars, but the hip belt broke on day two" is not a happy customer. Someone would have to read all 300, and every new product adds to the pile. The team doesn't need eloquent analysis, just a reliable happy/unhappy/mixed signal at scale.

The user in this feature is the product team, not the hiker, and that's deliberate. Stakeholders are users too. Their version of the problem is that nobody has time to read every review, so a defect surfaces only when returns spike. Track the same happy/unhappy signal weekly and it surfaces months earlier, as a chart sliding downhill. One product in this corpus has exactly that kind of problem buried in its reviews, and the demo gets to find it.

*A Challenge lab. Do it if you finished [Module 1](../../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20. All three are hand-picked from feature 06's `data/gear-reviews.jsonl`; no script builds them.
- **How:** send one prompt per review through your track's chat client. Keep the one-word label that comes back. Compare it with the hand label. Same prompt bytes everywhere, temperature 0.
- **Model:** `phi3` is the small model. The big model is `gpt-4.1` on Azure. Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` to use it. With no key, `llama3.2` on Ollama fills in for the big model, and the whole lab runs offline.

## The Concept

Sentiment analysis is classification applied to text, and it's this workshop's vehicle for the most useful model-selection lesson of the day: you don't always need the big model. A small local model (`phi3`, about 2GB, free, private) labels straightforward reviews just as well as a frontier cloud model. At 300 reviews per product across a whole catalog, per-token pricing versus free-on-your-hardware is a real budget line.

The comparison cuts both ways, though. Feed both models the corpus's hard cases (sarcasm like "Great bag, if you enjoy shoulder pain", mixed feelings, ratings that contradict the text) and accuracy drops for everyone. How far it drops for each model is the thing you measure, and there is no verdict here for either side. The decision is measurable: run both on a labeled sample, count the disagreements, look at what the errors cost you, then choose. Most teams never run that experiment; you'll run it before lunch.

Each step below adds one thing to the program. The `starter/` already classifies one review on `phi3`. Edit it until it does all six steps. Look at `complete/` when you get stuck; its comments say why each piece is there.

## Running

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
dotnet run -- --easy      # easy set only
dotnet run -- --hard      # hard set only
```

### Step 0: Run the starter on the sarcastic review

**Do:** run `starter/` as it is. It loads `gr-0007` from `data/hard.jsonl`, prints the product, star rating, reviewer, and text, then sends this prompt to `phi3` and prints the label:

```text
Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: Absolutely love it when the mesh blew out at the pinky toe inside two weeks the second day of a trip. Five-star experience, truly, if the stars are measuring my personal growth through adversity. Rating it what it deserves.
```

The starter already does this. It picks the id from the command line (or `gr-0007`), finds that review in the two data files, prints it, and classifies it:

```csharp
var id = args.Length > 0 ? args[0] : "gr-0007";
var review = File.ReadLines("../../data/easy.jsonl")
    .Concat(File.ReadLines("../../data/hard.jsonl"))
    .Select(line => JsonSerializer.Deserialize<Review>(line)!)
    .First(r => r.id == id);

Console.WriteLine($"{review.product} ({review.rating} stars), reviewed by {review.reviewer}");
Console.WriteLine(review.text);
Console.WriteLine();

var label = await Classify(client, review.text);
Console.WriteLine($"phi3 says: {label}");
```

`args` is the command-line argument array; top-level statements get it for free. Run it:

```bash
dotnet run
```

**Why:** the line breaks in that prompt are load-bearing. Reflowing the same words onto one line, changing nothing but the newlines, drops `phi3` from 9/10 to 7/10 on the easy set and from 7/10 to 4/10 on the hard set. Keep it wrapped exactly as it is here.

The prompt lives inside the starter's `Classify` method, near the bottom of `Program.cs`. The starter already does this; `{text}` is where the review goes:

```csharp
    var prompt = $"""
        Classify this gear review as exactly one word: positive, negative, or mixed.
        Positive means the reviewer is happy with the product, negative means unhappy,
        mixed means genuinely both. Judge the review text only; ignore any star rating
        it mentions. Reply with only the label.

        Review: {text}
        """;
```

**Check:** `phi3 says: negative`. The review has two stars and says "five-star experience, truly". That is sarcasm, and the small model gets it right. Pass a different id (`gr-0034`, for instance) to classify a different review.

```bash
dotnet run -- gr-0034
```

### Step 1: Load the reviews and the reference labels

**Do:**
1. Review how the existing code from the starter project opens `../../data/easy.jsonl` (relative to `starter/`): 10 reviews where text and stars agree, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`).

   The .NET starter has no data-folder constant; it uses the literal `../../data/` paths, which work because `dotnet run` runs from `starter/`. Line 14 opens the file:

   ```csharp
   var review = File.ReadLines("../../data/easy.jsonl")
   ```

2. Read it one line at a time, skip blanks, parse each line into a list.

   `File.ReadLines` already hands you one line at a time, and the starter already turns each line into a `Review` with `JsonSerializer.Deserialize<Review>(line)!` (the `!` tells the compiler the result is not null). The `Review` record is at the bottom of the starter; its lowercase property names match the JSON keys exactly, which is why no JSON attributes are needed. To collect a list, put this below the `IChatClient client` line:

   ```csharp
   // Hint: filter out blank lines, parse the rest, collect them
   var easy = File.ReadLines("../../data/easy.jsonl")
       .Where(line => !string.IsNullOrWhiteSpace(line))
       .Select(line => JsonSerializer.Deserialize<Review>(line)!)
       .ToList();
   ```

   `complete/` skips the list: its loop in step 2 reads the file line by line directly, so you can delete `easy` once the check below passes.

3. Open `../../data/reference-labels.json`: one object keyed by review `id`, each value a `set` (`easy`/`hard`), a `label`, and on hard cases a `rationale`.

   The file has one value per id, so it needs its own record. Add it at the very bottom of `Program.cs`, right under the `Review` record (record declarations must come after all the top-level statements):

   ```csharp
   record RefLabel(string set, string label, string? rationale = null);
   ```

   `string?` with `= null` lets easy-set entries leave `rationale` out.

4. Parse it into a dictionary keyed by `id`. You only need `label` in code.

Put this below the `IChatClient client` line (it is the same in `complete/Program.cs`). `labels["gr-0002"].label` then gives you one label:

```csharp
var labels = JsonSerializer.Deserialize<Dictionary<string, RefLabel>>(
    await File.ReadAllTextAsync("../../data/reference-labels.json"))!;
```

To see the check below, print the counts once and then delete the line:

```csharp
// Hint: print what you loaded
Console.WriteLine($"{easy.Count} {easy[0].id} {labels.Count} {labels["gr-0002"].label}");
```

**Why:** `gr-0004`'s label changed from `positive` to `mixed`. `gpt-4.1` called it `mixed` on every soak-test run, and a reread agreed that a two-star review which praises the product is split. `expected-output.md` tells that story.

**Check:** the easy list has 10 reviews, first `id` is `gr-0002`. The label dictionary has 20 keys; `gr-0002` maps to `negative`.

### Step 2: Classify the easy set on `phi3` and score it

**Do:**
1. Review how the existing code from the starter project classifies a review: one review text in, one prompt out, `phi3`, temperature 0.

   `client` is the `phi3` client passed in, and `Temperature = 0` is set on every call:

   ```csharp
   static async Task<string> Classify(IChatClient client, string text)
   ```

   ```csharp
       var response = await client.GetResponseAsync(prompt, new ChatOptions { Temperature = 0 });
   ```

2. Review how it lowercases the reply, looks for `positive`/`negative`/`mixed`, and keeps whichever shows up first (or the trimmed reply if none do).

   It happens at the end of `Classify`. `IndexOf` finds where each label appears (-1 if absent), and the LINQ chain keeps the earliest one:

   ```csharp
       var raw = response.Text.ToLowerInvariant();

       // Small models sometimes wrap the label in a sentence; keep the first label mentioned.
       var first = new[] { "positive", "negative", "mixed" }
           .Select(l => (label: l, at: raw.IndexOf(l, StringComparison.Ordinal)))
           .Where(x => x.at >= 0)
           .OrderBy(x => x.at)
           .Select(x => x.label)
           .FirstOrDefault();
       return first ?? raw.Trim();
   ```

3. Loop over the easy list, classify each review's `text`, look up the reference label by `id`.
4. Print one row per review: `id`, reference label, `phi3` label.
5. Count matches; print `phi3 N/10`.

Delete the starter's single-review block (from `var id = ...` down to `Console.WriteLine($"phi3 says: {label}");`) and put this in its place, below your step 1 `labels` lines. It covers items 3 to 5. Like `complete/`, it reads `easy.jsonl` line by line instead of using the `easy` list. `{review.id,-9}` pads the value to 9 characters so the columns line up:

```csharp
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

**Why:** the prompt asks for one word, but a small model sometimes wraps the label in a sentence. Lowercasing the reply and keeping the first of `positive`, `negative`, or `mixed` that appears scores those replies as the label they contain, so the number you get measures the model's judgment and not its phrasing.

**Check:** `gr-0003` comes back `positive`, `gr-0002` `negative`. `phi3` should get most of the easy set (when I ran it, 9/10, missing `gr-0074`). A 7/10 with every miss `mixed` is the tell that you reflowed the prompt onto one line.

### Step 3: Add the big model and classify every review twice

**Do:**
1. Read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT`.

   ```bash
   export AZURE_OPENAI_ENDPOINT=https://trailhead-ai-workshop.openai.azure.com
   export AZURE_OPENAI_KEY=<KEY FROM INSTRUCTOR>
   export AZURE_OPENAI_DEPLOYMENT=gpt-4.1
   ```

2. Build a second chat client from those three variables, falling back to `llama3.2` when any is missing; name it `azure:<deployment>` or `llama3.2` for printing.

   The starter project only references `Microsoft.Extensions.AI` and `OllamaSharp`. The Azure client needs two more packages, and `AsIChatClient()` on the OpenAI client is marked experimental, so the build fails with `OPENAI001` unless you suppress it. Add these to `starter/Sentiment.csproj` just above the closing `</Project>` line (they match `complete/Sentiment.csproj`, which puts the `NoWarn` line in its first `PropertyGroup` and the two packages in its existing `ItemGroup`; either layout builds). `dotnet run` downloads the packages on the next build:

   ```xml
   <PropertyGroup>
     <NoWarn>$(NoWarn);OPENAI001</NoWarn>
   </PropertyGroup>
   <ItemGroup>
     <PackageReference Include="Azure.AI.OpenAI" Version="2.1.0" />
     <PackageReference Include="Microsoft.Extensions.AI.OpenAI" Version="10.8.1" />
   </ItemGroup>
   ```

   Then at the top of `Program.cs`, next to the existing `using` lines:

   ```csharp
   using System.ClientModel;   // ApiKeyCredential
   using Azure.AI.OpenAI;      // AzureOpenAIClient
   ```

   Rename the starter's `client` to `phi3`, as `complete/` does, then build the second client from the three env vars. The block below replaces the starter's `IChatClient client = ...` line, so your step 2 call `Classify(client, review.text)` must become `Classify(phi3, review.text)`. `Environment.GetEnvironmentVariable` returns `null` when a variable is not set. This also covers item 3:

   ```csharp
   IChatClient phi3 = new OllamaApiClient(new Uri("http://localhost:11434"), "phi3");

   var endpoint = Environment.GetEnvironmentVariable("AZURE_OPENAI_ENDPOINT");
   var key = Environment.GetEnvironmentVariable("AZURE_OPENAI_KEY");
   var deployment = Environment.GetEnvironmentVariable("AZURE_OPENAI_DEPLOYMENT");

   IChatClient bigModel;
   string bigName;
   if (!string.IsNullOrEmpty(endpoint) && !string.IsNullOrEmpty(key) && !string.IsNullOrEmpty(deployment))
   {
       bigModel = new AzureOpenAIClient(new Uri(endpoint), new ApiKeyCredential(key))
           .GetChatClient(deployment)
           .AsIChatClient();
       bigName = $"azure:{deployment}";
   }
   else
   {
       Console.WriteLine("AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.");
       Console.WriteLine();
       bigModel = new OllamaApiClient(new Uri("http://localhost:11434"), "llama3.2");
       bigName = "llama3.2";
   }
   ```

3. When falling back, print `AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.`
4. Change nothing in the body of classify. The second model is just a different client; some tracks pass that client in as a parameter, so its signature line may change.

5. In the step 2 loop, classify each review with both `phi3` and the big model; store a record with the review, set name, reference label, small label, big label.
6. Add a fourth column for the big model's label; append `  <- disagree` when the two differ.

   Add a record type at the very bottom of `Program.cs`, under `RefLabel`:

   ```csharp
   record Result(Review Review, string Set, string Reference, string Small, string Big);
   ```

   Then grow the step 2 loop to a four-column table. `set` arrives with step 4; until then declare it yourself, above the loop:

   ```csharp
   // Hint: a stand-in set name until step 4 adds the outer loop
   var set = "easy";
   ```

   Replace the step 2 `var correct = 0; var total = 0;` line with the first two lines below. Replace everything inside the loop's `{ }` with the other seven. Delete the `Console.WriteLine($"phi3 {correct}/{total}");` line after the loop; step 4 brings scoring back. `new(...)` builds a `Result`, because `results` is a `List<Result>`:

   ```csharp
   var results = new List<Result>();
   Console.WriteLine($"{"id",-9} {"reference",-10} {"phi3",-10} {bigName,-10}");
   var review = JsonSerializer.Deserialize<Review>(line)!;
   var reference = labels[review.id].label;
   var small = await Classify(phi3, review.text);
   var big = await Classify(bigModel, review.text);
   results.Add(new(review, set, reference, small, big));
   var flag = small != big ? "  <- disagree" : "";
   Console.WriteLine($"{review.id,-9} {reference,-10} {small,-10} {big,-10}{flag}");
   ```

**Why:** the swap is one extra client, built from `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT`, plus one extra `classify` call per review. The prompt, the body of `classify`, and the scoring all stay as they were, so whatever differs between the two label columns came from the model.

**Check:** the easy table has four columns, and `gr-0074` is flagged `<- disagree`: reference `positive`, `phi3` says `mixed`, and the big-model column (headed `azure:gpt-4.1`) says `positive`. On the `llama3.2` stand-in the row is still flagged, big label `negative`.

### Step 4: Run the hard set too and print accuracy per set

**Do:**
1. Put the step 3 loop inside an outer loop over `easy` then `hard`, reading `../../data/{set}.jsonl` each time, with a heading line per set.
2. After both tables, loop over the two set names again; count matches for small and big labels against the reference.
3. Print one line per set: name, `phi3 N/10`, big model's name with `N/10`.

`complete/` builds `sets` from the stretch-goal flags; with no flags it is `easy` then `hard`. Delete your step 3 `var set = "easy";` line. Then replace everything from `var results = new List<Result>();` to the end of the step 3 loop with the block below: the step 3 loop moves inside the outer loop, and the `$"../../data/{set}.jsonl"` path picks the file per set. The accuracy loop goes right after it. `Where`, `Count`, and `ToList` are LINQ; the `ImplicitUsings` setting in the `.csproj` already imports them:

```csharp
var sets = args.Contains("--easy") ? new[] { "easy" }
         : args.Contains("--hard") ? new[] { "hard" }
         : new[] { "easy", "hard" };

var results = new List<Result>();
foreach (var set in sets)
{
    Console.WriteLine($"── {set} set ──");
    Console.WriteLine($"{"id",-9} {"reference",-10} {"phi3",-10} {bigName,-10}");
    foreach (var line in File.ReadLines($"../../data/{set}.jsonl"))
    {
        var review = JsonSerializer.Deserialize<Review>(line)!;
        var reference = labels[review.id].label;
        var small = await Classify(phi3, review.text);
        var big = await Classify(bigModel, review.text);
        results.Add(new(review, set, reference, small, big));
        var flag = small != big ? "  <- disagree" : "";
        Console.WriteLine($"{review.id,-9} {reference,-10} {small,-10} {big,-10}{flag}");
    }
    Console.WriteLine();
}

Console.WriteLine("── accuracy vs. reference labels ──");
foreach (var set in sets)
{
    var batch = results.Where(r => r.Set == set).ToList();
    var smallOk = batch.Count(r => r.Small == r.Reference);
    var bigOk = batch.Count(r => r.Big == r.Reference);
    Console.WriteLine($"{set,-5}  phi3 {smallOk}/{batch.Count}   {bigName} {bigOk}/{batch.Count}");
}
Console.WriteLine();
```

**Check:** `gr-0034` comes back `negative` from both models. On `gr-0013`, `phi3` says `mixed`, big model says `negative` (the reference label). Your numbers will not be mine. When I ran it: `phi3` 9/10 easy and 7/10 hard, `gpt-4.1` 10/10 on both, `llama3.2` stand-in 9/10 and 8/10. What has to hold is the shape, the big model ahead of `phi3` on the hard set, not the digits.

### Step 5: Print the disagreement list

**Do:**
1. Take only records where small and big labels differ.
2. Print a heading: `disagreements (N of 20)`.
3. For each: verdict is the big model's name + ` right` if it matches the reference, `phi3 right` if the small label matches instead, else `both wrong`.
4. Print one line per disagreement: `id`, set in brackets, `ref=`, `phi3=`, big model `=`, verdict in parens; then on its own indented line the review text in quotes, cut to 100 characters with `...` if longer.
5. If empty, print `(none this run)`.

```csharp
var disagreements = results.Where(r => r.Small != r.Big).ToList();
Console.WriteLine($"── disagreements ({disagreements.Count} of {results.Count}) ──");
foreach (var d in disagreements)
{
    var verdict = d.Big == d.Reference ? $"{bigName} right"
                : d.Small == d.Reference ? "phi3 right"
                : "both wrong";
    Console.WriteLine($"{d.Review.id} [{d.Set}] ref={d.Reference} phi3={d.Small} {bigName}={d.Big}  ({verdict})");
    Console.WriteLine($"  \"{Truncate(d.Review.text, 100)}\"");
}
if (disagreements.Count == 0) Console.WriteLine("(none this run)");
```

Put the lines above after the accuracy loop. The `Truncate` helper below is a `static` method like `Classify`: put it at the bottom of `Program.cs`, under `Classify` and above the `record` lines. `s[..max]` is the first `max` characters:

```csharp
static string Truncate(string s, int max) =>
    s.Length <= max ? s : s[..max].TrimEnd() + "...";
```

**Check:** your version of the two tables in `expected-output.md`. A handful of disagreements per model is the normal shape (when I ran it, 4 of 20 against `gpt-4.1`, missing `gr-0074`, `gr-0004`, `gr-0013`, `gr-0021`, and 4 of 20 against `llama3.2`, missing `gr-0074`, `gr-0004`, `gr-0013`, `gr-0089`; the first three show up on both). Your numbers will differ. No tables at all is the failure.

### Stretch goals

Pick any. The first two are already built in `complete/`.

- **Add `--easy` and `--hard` flags.** If `--easy` is present, run only the easy set; if `--hard`, only the hard set; otherwise both. **Check:** `--hard` prints one table, one accuracy line, and a disagreement count out of 10 instead of 20.

  If you used the step 4 `var sets = args.Contains("--easy") ? ...` lines, this is already done; `args.Contains` checks whether that word was passed. The `--` stops `dotnet run` from reading the flag itself:

  ```bash
  dotnet run -- --hard
  ```

- **Reflow the prompt onto one line and measure the damage.** Replace the four line breaks with spaces, rerun both sets on both models, then put the line breaks back. **Check:** `phi3` drops from 9/10 to 7/10 easy and from 7/10 to somewhere between 4/10 and 6/10 hard (4/10 in the recorded run, 6/10 in two later runs), every miss `mixed`. `llama3.2` scores the same either way. The small model is the one that cares about prompt shape.

  The change is inside `Classify`. Join the four instruction lines into one and leave the blank line and the `Review:` line alone:

  ```csharp
  // Hint: same words, one line
  var prompt = $"""
      Classify this gear review as exactly one word: positive, negative, or mixed. Positive means ... Reply with only the label.

      Review: {text}
      """;
  ```

- **Aspect-based sentiment.** Ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` instead of one word, and tell the model to reply with that JSON and nothing else. Parse the reply, and catch the parse error so one bad reply doesn't stop the run. **Check:** the run finishes without crashing, and the replies that parse leave aspects `null` when the review never mentions them. Expect many `phi3` replies not to parse: in one measured run only 7 of 20 did, and the rest added an explanation after the JSON or wrapped it in a ```` ```json ```` fence. Those are the parse errors you catch, and they are why production code uses a structured-output schema like step 1 of feature 02. A `price` sentiment on a review that never mentions price is the failure to look for.

  `complete/` does not build this one. Write a second classify method with a prompt that asks for that JSON, and records at the bottom of `Program.cs` whose property names match the JSON keys. `string?` lets an aspect come back `null`. `complete/Program.cs` has no MEAI call for a `format` schema, so this hint parses the reply text with `JsonSerializer` instead.

  Careful with the prompt: the JSON's `{` and `}` clash with the `{text}` holes in a `$"""` string, and `{{` does not escape them there, so the build fails with `CS9006`. Drop the `$` and add the review text with `+` instead:

  ```csharp
  // Hint: no $ on this string, so the JSON braces are plain text
  var prompt = """
      Classify this gear review. Reply with only JSON shaped like
      {"overall": "positive|negative|mixed", "aspects": {"comfort": ..., "durability": ..., "price": ...}}
      Use null for any aspect the review never mentions.

      Review:
      """ + " " + text;
  ```

  If `Deserialize` throws a `JsonException`, print `response.Text`: a small model that adds a sentence after the JSON is the "parseable JSON every time" check failing.

  ```csharp
  // Hint: records shaped like the JSON (bottom of Program.cs)
  record Aspects(string? comfort, string? durability, string? price);
  record AspectResult(string overall, Aspects aspects);
  ```

  ```csharp
  // Hint: inside your new method, after the same GetResponseAsync call as Classify
  var parsed = JsonSerializer.Deserialize<AspectResult>(response.Text);
  Console.WriteLine($"{parsed?.overall} price={parsed?.aspects.price ?? "null"}");
  ```

## What Is in This Folder

- `data/easy.jsonl`: 10 straightforward reviews chosen by hand from the full review set (feature 06's `data/gear-reviews.jsonl`). The text says what it means and the star rating agrees.
- `data/hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `data/reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case. One label (`gr-0004`) was revised after the soak test; `expected-output.md` records why.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The reviews keep their original ids, so any of them can be traced back to the full corpus.
