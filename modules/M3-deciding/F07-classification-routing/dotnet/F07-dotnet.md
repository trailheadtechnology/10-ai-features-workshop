# .NET Demo for 07 Classification & Routing

Two console projects named `Triage`, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one classify call, one inquiry. The taxonomy is plain-language category descriptions in the prompt and the model answers as free text, which is fine until it answers "Emergency." or "the category is emergency" and your `switch` falls through to the default queue.
- `complete/`: the finished demo as shown on stage. Same taxonomy, but through `GetResponseAsync<TriageResult>` with a C# enum, so the label is one of seven values or the call fails. It classifies all 20 messages in `../../data/inquiries-slice.jsonl`, prints anything it called an emergency at the top in a block you cannot scroll past, prints the routing table (id, category, queue), and scores itself against `../../data/reference-labels.json`.

Both run against Ollama (`llama3.2`), matching the demo script in [docs/slides/outlines/M3-deciding.md](../../../../docs/slides/outlines/M3-deciding.md).

```bash
cd starter && dotnet run              # classify inq-0005
cd starter && dotnet run inq-0041     # any id from the slice
cd complete && dotnet run             # all 20, routed and scored
```

A real run of `complete/` scored 17/20 against the reference labels with both emergencies caught and no false emergencies, and it put the ambiguous `inq-0035` in `unsure`, which is the reference label. The three misses were `inq-0030` (a permit-rules question filed as `general`, the misclassification you fix live by widening the `permit` description), `inq-0051` (campfire rules filed as `conditions`), and `inq-0001` (a stuck permit booking sent to `unsure` rather than `permit`, one over-cautious call worth pointing at). See [../expected-output.md](../expected-output.md) for the full labeling and what it means.

One thing worth showing on stage: `starter/`, with the same taxonomy but free-text output, still calls `inq-0035` `conditions`. The enum in `complete/` is what makes `unsure` a live option for the model rather than a paragraph it skims past. Structured output is not only about parsing.

Emergency recall is the number to watch on stage. Accuracy moves a point or two between runs; missing an emergency is the failure the demo exists to talk about.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F07-lab.md`](../F07-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Try a Few Ids

One inquiry, free-text label. Try `inq-0035`, the ambiguous one, and watch the answer vary between `conditions` and `permit`; try `inq-0013` and check that it says `emergency`. Nothing stops the model returning "Emergency." or a sentence.

Run:

```bash
dotnet run
dotnet run -- inq-0035
dotnet run -- inq-0013
```

Check: A label per run, and at least one that is not exactly one of the seven category names.

### Step 2: Make the Label an Enum Through Structured Output, and Loop All 20 (lab step 1)

The category becomes a type with seven values, so the model can only return a label the routing table knows how to handle. Keep the taxonomy prompt from the starter, pin temperature at 0 (anything above it makes a scored comparison against fixed labels meaningless), and loop the slice.

```csharp
[JsonConverter(typeof(JsonStringEnumConverter<Category>))]
enum Category
{
    [JsonStringEnumMemberName("permit")] Permit,
    [JsonStringEnumMemberName("conditions")] Conditions,
    [JsonStringEnumMemberName("complaint")] Complaint,
    [JsonStringEnumMemberName("lost-and-found")] LostAndFound,
    [JsonStringEnumMemberName("emergency")] Emergency,
    [JsonStringEnumMemberName("general")] General,
    [JsonStringEnumMemberName("unsure")] Unsure,
}
record TriageResult(Category Category);

var options = new ChatOptions { Temperature = 0 };
foreach (var inquiry in inquiries)
{
    var response = await client.GetResponseAsync<TriageResult>(Prompt(inquiry.text), options);
    results.Add((inquiry, response.Result.Category));
}
```

Check: Twenty labels, every one of them one of the seven strings. Note that `inq-0035` now lands in `unsure`: the enum made it a live option rather than a paragraph the model skims past.

### Step 3: Score Against the Reference Labels, and Score Emergency Recall Separately (lab step 1, the scoring pass)

`../data/reference-labels.json` has `labels` (id to category) and `routing` (category to queue). Two numbers: overall accuracy, and recall on the emergency class. They are not equally important. Missing an emergency fails the lab at 19/20.

```csharp
var correct = results.Count(r => Wire(r.Category) == reference.Labels[r.Inquiry.id]);
var emergencyIds = reference.Labels.Where(l => l.Value == "emergency").Select(l => l.Key).ToList();
var caught = results.Count(r => r.Category == Category.Emergency && emergencyIds.Contains(r.Inquiry.id));
Console.WriteLine($"Accuracy vs reference labels: {correct}/{results.Count}");
Console.WriteLine($"Emergency recall: {caught}/{emergencyIds.Count}");
```

Run:

```bash
dotnet run
```

Check: Recorded: 17/20 and 2/2. Print each miss with what the model said and what the reference says; that list drives the next step.

### Step 4: Fix a Miss by Editing a Description, Not the Code (lab step 2)

`inq-0030` (a wedding photographer asking whether a special-use permit is required) lands in `general` because the `permit` description talks about reserving and paying. Widen the description by a clause and re-run. Two rules constrain any edit: the ordering paragraph that makes emergency win stays, and `unsure` stays narrow.

```csharp
- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, and questions about whether an activity requires a
  permit at all, including billing problems and missing confirmations.
```

Run:

```bash
dotnet run
```

Check: Lab step 3, the success check: both emergencies classified `emergency`, `inq-0035` in `unsure`, and your accuracy at or above where it started. Judge every taxonomy edit on emergency recall first. Stretch: add a `priority` field to the result type, or a confidence score routed to `unsure` below a threshold.
