# .NET Demo for 02 Extraction

Two console projects, both built on Microsoft.Extensions.AI over OllamaSharp:

- `starter/`: the demo's starting point. One `IChatClient`, one naive prompt ("Extract the details of this trip report as JSON."), no schema, so you get whatever shape the model feels like emitting.
- `complete/`: the finished demo as shown on stage. A C# record `TripFacts` with nullable fields and `[Description]` attributes, handed to `GetResponseAsync<TripFacts>()`. The schema is code, and there is no parsing step. Then the validator, which is the half of the demo that ships.

Each report prints three blocks: what the model gave us, what the validator says field by field, and what we would actually store. The rules are plain C# in `Program.cs`, no model involved:

- `date_hiked` must parse against an explicit format list or be `null`. "July 4, 2026" passes and is normalized to `2026-07-04`; "last month" is rejected.
- `distance_mi` and `elevation_gain_ft` reject `0`, negatives, and the absurd (over 100 mi, over 20,000 ft). Zero is the near-miss worth pausing on: a pipeline stores it without complaint.
- Empty and whitespace-only strings are rejected.
- `trail_name` and `park` get a grounding check against the source text, which is what catches an invented trail name.

Anything rejected becomes `null` in the stored object. Null is a gap someone can fill later; a plausible wrong number never gets questioned.

Both run against Ollama (`llama3.2`, JSON mode), matching the demo script in [docs/slides/outlines/M1-understanding.md](../../../../docs/slides/outlines/M1-understanding.md). From `complete/`:

```bash
dotnet run                                       # both lab reports, extracted and validated
dotnet run -- ../../data/tr-0011.md               # just the sparse one, for the null check
dotnet run -- ../../../F01-summarization/data/tr-0002.md   # any report path works
```

Run the sparse report three or four times on stage. The output moves, and that is the demo: some runs come back clean, and some hand you `0` or "last month" and the validator says so out loud. Real output from four consecutive runs, rejections included, lives in [../expected-output.md](../expected-output.md).

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F02-lab.md`](../F02-lab.md), done in .NET: start from `starter/Program.cs` and end where `complete/Program.cs` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with `dotnet run`; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Look at What "JSON in the Prompt" Gets You

The starter asks for JSON in prose. Run it twice on `tr-0007.md` and compare: field names drift, there may be a preamble or a markdown fence, and nothing guarantees it parses. This is what the schema replaces.

Run:

```bash
dotnet run
```

Check: Two runs, two shapes.

### Step 2: Make the Schema Code and Ask for a Typed Response (lab step 1)

Define the record with every scalar nullable and a description on every field, then ask for that type instead of prose. The descriptions do the prompting; "null if the report does not state" on each field is the half of the hallucination fix that no plea in the prompt can do.

```csharp
record TripFacts(
    [property: Description("The name of the trail hiked. null if the report never names the trail.")] string? TrailName,
    [property: Description("The park the trail is in. null if the report never names the park.")] string? Park,
    [property: Description("The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.")] string? DateHiked,
    [property: Description("Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.")] double? DistanceMi,
    [property: Description("Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.")] double? ElevationGainFt,
    [property: Description("Animals the author actually saw on this hike. Empty array if none are mentioned.")] string[] Wildlife,
    [property: Description("Short phrases describing trail conditions the report mentions. Empty array if none.")] string[] Conditions,
    [property: Description("Hazards or closures the report mentions. Empty array if none.")] string[] Hazards);

var response = await client.GetResponseAsync<TripFacts>($"""
    Extract the trail facts from this trip report.
    Use null for any field the report does not state, and empty arrays
    when nothing applies. Do not guess.

    {report}
    """);
var facts = response.Result;
```

Run:

```bash
dotnet run
```

Check: A populated object, no parsing step, and the values match the `tr-0007.md` block in `../expected-output.md` (Sperry Chalet Trail, 2026-07-04, 12.8 mi, 3400 ft).

### Step 3: Run the Sparse Report and Count What It Made up (lab step 2)

`tr-0011.md` never names the trail, gives no distance, no elevation, and no exact date. Run it three or four times and write down every field that came back with a value the report does not contain. The recorded runs in `../expected-output.md` show `elevation_gain_ft: 0` and `date_hiked: "early last month"`.

Run:

```bash
dotnet run -- ../../data/tr-0011.md
```

Check: Most missing facts come back `null`, and you can name the ones that did not. That list is what the next step is for.

### Step 4: Fix What the Schema Can Fix, Then Write the Validator for the Rest (lab step 3)

First tighten the descriptions (the "null, never 0" wording above is that fix). Then add rules in code for what the schema cannot express: a date must parse in an explicit format, a measurement of 0 is not a measurement, a name must appear in the source text. Anything that fails is coerced to `null` before it could reach a database. The two rules below catch the two recorded failures; `complete/` has all five plus the grounding check, and the small `Verdict` type they return (field, value, passed, reason, optional normalized value) is defined there too.

```csharp
static Verdict InRange(string field, double? value, double max, string unit)
{
    if (value is null) return Verdict.Pass(field, null);
    if (value.Value == 0)
        return Verdict.Fail(field, $"{value:0.#} {unit}", "0 is not a measurement; the report gave no figure, so this should be null");
    return value.Value > max ? Verdict.Fail(field, $"{value:0.#} {unit}", $"implausible: over {max:0} {unit}") : Verdict.Pass(field, $"{value:0.#} {unit}");
}

static Verdict ValidDate(string field, string? value)
{
    if (value is null) return Verdict.Pass(field, null);
    string[] formats = ["yyyy-MM-dd", "yyyy/MM/dd", "MM/dd/yyyy", "MMMM d, yyyy", "MMM d, yyyy", "d MMMM yyyy"];
    return DateOnly.TryParseExact(value.Trim(), formats, CultureInfo.InvariantCulture, DateTimeStyles.None, out var d)
        ? Verdict.Pass(field, value) with { Normalized = d.ToString("yyyy-MM-dd") }
        : Verdict.Fail(field, value, "does not parse as a date; no parser can store this");
}
```

Run:

```bash
dotnet run -- ../../data/tr-0011.md   # several times
```

Check: Every rejected field prints a reason, and "what we would store" has `null` where the model had `0` or prose. Zero is the dangerous one: it is a value, and a pipeline will store it without complaint. Stretch: add a per-field confidence, or extract an array of records for a multi-day report.
