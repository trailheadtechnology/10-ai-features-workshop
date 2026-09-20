<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 07: Classification & Routing (.NET)

*You are on the .NET track. Other tracks: [Python](../python/F07-python.md), [TypeScript](../typescript/F07-typescript.md). Lab overview: [F07-lab.md](../F07-lab.md).*

**The User Problem:** Every message to Trailhead Guides lands in one inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, and, occasionally, someone reporting an actual emergency. A ranger triages the pile by hand once or twice a day. The permit request waits behind the granola questions, the complaint goes to the wrong person twice, and the emergency sits unread for four hours. The users' real problem is that their message goes into a hole, and how fast it comes out depends on luck.

*This is the Recommended lab for [Module 3](../../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** sort visitor messages into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and send each one to the right queue.
- **Input:** `data/inquiries-slice.jsonl` (20 messages: `id`, `channel`, `received`, `text`; emergencies `inq-0013` and `inq-0041`, ambiguous `inq-0035`), `data/reference-labels.json` (the correct category for each id, and the queue for each category), `data/inquiries.jsonl` (the full 100-message inbox the 20 came from; the demo scrolls it, the lab does not read it).
- **How:** send each message to `llama3.2` with the taxonomy prompt. Use a JSON schema so the model can only answer with one of the seven labels. Loop all 20 at temperature 0. Print the routing table, then score the results against the reference labels.
- **Model:** `llama3.2`, local. No key.

## The Concept

This is classification again (feature 03 was the warm-up), but now the label has consequences: it decides where the message goes and how fast. An LLM makes a solid zero-shot classifier. You describe the categories in plain language and it labels messages with no training data, which is exactly the situation most teams are in on day one.

Two design decisions carry the feature. The first is that the taxonomy is the product: category names and one-sentence descriptions in the prompt are where accuracy lives, and when the model misfiles something, you usually fix the description rather than the model. The second is that errors are not symmetric, since misrouting a complaint costs a day of annoyance and misrouting an emergency is a headline. So the system needs an "unsure" route to a human, and it should be tuned so the expensive class never slips through, even at the cost of extra false alarms. Recall on the class that matters, not overall accuracy, is the number to watch. A local model does this fine, and at inbox volume, free matters.

Every step below is one thing to make the program do. The starter classifies one message and prints whatever text the model sends back. Edit it until it does all six steps. Compare against `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two console projects named `Triage`, both built on Microsoft.Extensions.AI over OllamaSharp. `starter/` has one `IChatClient`, one classify call, and one inquiry, answered as free text. `complete/` is the finished demo as shown on stage: it classifies all 20 messages through `GetResponseAsync<TriageResult>` with a C# enum, prints emergencies first, prints the routing table, and scores itself against the reference labels.

```bash
cd starter && dotnet run              # classify inq-0005
cd starter && dotnet run inq-0041     # any id from the slice
cd complete && dotnet run             # all 20, routed and scored
```

### Step 0: Run the starter and read the free-text label

**Do:** run `starter/` as it is. It reads `inq-0005` from the slice, builds the taxonomy prompt, puts the message after `Message:`, and sends it to `llama3.2` as one user message. Then run it again with `inq-0035` as the argument, and again with `inq-0013`.

```bash
dotnet run
dotnet run -- inq-0035
dotnet run -- inq-0013
```

The starter already does this. It picks the id from the command line (or `inq-0005`), reads the slice, and keeps the matching line:

```csharp
var wantedId = args.Length > 0 ? args[0] : "inq-0005";
var inquiry = File.ReadLines("../../data/inquiries-slice.jsonl")
    .Select(line => JsonSerializer.Deserialize<Inquiry>(line)!)
    .First(i => i.id == wantedId);
```

The starter already does this too. The prompt is one big `$"""` string that ends with the message text, and the call sends it as one user message and prints whatever text comes back:

```csharp
    Message:
    {inquiry.text}
    """;

var response = await client.GetResponseAsync(prompt);
Console.WriteLine($"{inquiry.id}: {response.Text}");
```

**Check:** one label per run, printed after the id. `inq-0013` says `emergency`. `inq-0035` usually says `conditions`, and some runs say `unsure`, `emergency`, or `complaint`: measured over 30 runs across the three tracks, 24 `conditions`, 4 `unsure`, 1 `emergency`, 1 `complaint`. That spread on one message is the problem. Nothing in the starter stops a run from printing a label that is not one of the seven names, such as `Emergency.` or a sentence, either. The rest of the lab closes both gaps.

### Step 1: Load the 20 inquiries and the reference labels

**Do:**
1. Open `../../data/inquiries-slice.jsonl`. Every line is one JSON object with `id`, `channel`, `received`, `text`. Read it line by line, skip blanks, parse each line, keep the results in a list. Your track's block below says how to point at that `data/` folder; use the same way for both files.
2. Open `../../data/reference-labels.json` and parse it: an object with `routing` (category to queue name), `labels` (id to correct category), and `notes` (why `inq-0013`/`inq-0041` are emergencies and why `inq-0035` is `unsure`). Keep the `routing` and `labels` dictionaries.

Load `../../data/inquiries-slice.jsonl` line by line into a list of `Inquiry` records (the starter already declares `Inquiry`) and deserialize `../../data/reference-labels.json` into a record with `Routing` and `Labels` dictionaries. `[JsonPropertyName]` maps the lowercase JSON keys; it needs `using System.Text.Json.Serialization;`, which the starter does not have. Add that `using` line under `using System.Text.Json;` at the top.

The .NET starter has no data-folder constant; it writes the relative path `../../data/...` inline, so do the same. Put these two statements right below the `IChatClient client = ...` line. Leave the starter's `wantedId`/`inquiry` lookup in place for now; step 3 removes it.

```csharp
var inquiries = File.ReadLines("../../data/inquiries-slice.jsonl")
    .Select(line => JsonSerializer.Deserialize<Inquiry>(line)!)
    .ToList();

var reference = JsonSerializer.Deserialize<ReferenceLabels>(
    File.ReadAllText("../../data/reference-labels.json"))!;
```

The record goes at the bottom of the file, below the `record Inquiry(...)` line. In a file with top-level statements, every `record`, `enum`, and `static` helper declaration goes after the last top-level statement:

```csharp
record ReferenceLabels(
    [property: JsonPropertyName("routing")] Dictionary<string, string> Routing,
    [property: JsonPropertyName("labels")] Dictionary<string, string> Labels);
```

To see the Check numbers, print them once below the two load statements (delete the line afterward):

```csharp
// Hint: a throwaway print for the Check
Console.WriteLine($"{inquiries.Count} {inquiries[0].id} {reference.Routing.Count} {reference.Labels.Count}");
```

**Why:** these 20 are drawn from the fictional Trailhead Guides 100-message inbox, chosen to mirror the inbox's mix, including both emergencies and the one ambiguous message. `unsure` was added to the taxonomy, labels, and routing table together once the ambiguous message needed a queue.

**Check:** the inquiry list has 20 entries. The first `id` is `inq-0001`. `routing` has 7 keys and `labels` has 20.

### Step 2: Pin the answer to the seven labels with structured output

**Do:**
1. Delete the line `Answer with the category name only.` from the starter's prompt. The schema below does that job now.

   In `Program.cs` it is this line near the end of the `var prompt = $"""` string, just above `Message:`. Delete it and the blank line under it:

   ```csharp
       Answer with the category name only.
   ```
2. Define the category as an enum type with exactly seven allowed values, wrapped in a result type with one field, `category`. If your language cannot spell `lost-and-found` as an identifier, map that member to the exact JSON name. The generated schema must list `lost-and-found`, or three of the 20 messages can never match their reference label.

   C# identifiers cannot contain a hyphen, so `lost-and-found` needs `[JsonStringEnumMemberName]`, and the enum needs the `JsonStringEnumConverter` attribute so the generated schema lists the JSON names rather than `LostAndFound`. Both attributes need the `using System.Text.Json.Serialization;` line you added in step 1.

   Paste the enum and the one-field result record at the bottom of `Program.cs`, below `record ReferenceLabels(...)`. They are declarations, so they must come after all the top-level statements:

   ```csharp
   record TriageResult(Category Category);

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
   ```

3. Send the message through your chat client with that result type as the structured-output schema, temperature 0.

   The typed call is `GetResponseAsync<TriageResult>`, temperature goes in `ChatOptions`, and the parsed record is `response.Result`. Pin temperature at 0; anything above it makes a scored comparison against fixed labels meaningless.

   First make the options object once. Put it right below the `IChatClient client = ...` line:

   ```csharp
   var options = new ChatOptions { Temperature = 0 };
   ```

   Then replace the starter's last two top-level lines (`var response = await client.GetResponseAsync(prompt);` and the `Console.WriteLine` under it). Keep the starter's single-id lookup for this step; step 3 adds the loop.

   ```csharp
   // Hint: the typed call on the one inquiry, then two prints
   var response = await client.GetResponseAsync<TriageResult>(prompt, options);
   Console.WriteLine($"{inquiry.id}: {response.Result.Category}");
   Console.WriteLine(JsonSerializer.Serialize(response.Result));
   ```

   The first print shows the C# member name (`Conditions`, `LostAndFound`); step 4 adds a helper that turns it back into `lost-and-found`. The second print is the wire form the step 2 Check describes. C# writes it as `{"Category":"conditions"}`, with a capital `C` and no space, because `JsonSerializer` keeps the record's property name by default; the value is the part to check.

4. Keep the prompt text:

```text
You are the triage system for the Trailhead Guides shared inbox. Classify the visitor message into exactly one category.

- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, including billing problems and missing confirmations for a permit application.
- conditions: asking whether a trail, road, or area is open, safe, or passable right now: snow, water levels, washouts, wildlife activity, closures.
- complaint: unhappy about a park facility, service, or staff member and wants it acknowledged or fixed.
- lost-and-found: reporting a lost or found physical item.
- emergency: a person may be hurt, missing, or in danger right now and needs immediate human attention.
- general: anything else: park rules, fees, trip planning, questions that fit none of the above.
- unsure: two different queues both have to act before this message can be resolved, so no single queue owns it. The case that qualifies: the sender asks about trail conditions AND asks someone to change, refund, or cancel a booking. Trail info cannot issue a refund, and the permits office does not decide whether a trail is passable, so a human reads this queue and splits the work. Also use unsure when the message fits none of the categories above.

Decide in this order. First, if anyone might be hurt, missing, or in danger, answer emergency and stop; never answer unsure for those, even when the message also mentions permits, conditions, or a lost item. Second, if one queue can resolve the whole message on its own, answer that queue; a booking or reservation problem with nothing else attached is permit, not unsure. Third, only if two queues must both act, answer unsure. Unsure is not a catch-all for anything hard.

Message:
```

The starter already does this: its `var prompt = $"""` string holds that exact text, and it ends by putting the message after `Message:`. After item 1 the only change is the deleted line, so the end of the string still reads:

```csharp
    Message:
    {inquiry.text}
    """;
```

This is the schema your client generates from that type:

```json
{
  "type": "object",
  "properties": {
    "category": {
      "type": "string",
      "enum": ["permit", "conditions", "complaint", "lost-and-found", "emergency", "general", "unsure"]
    }
  },
  "required": ["category"]
}
```

5. Run the program on `inq-0005`, then `inq-0041`, then `inq-0035`. The message text for `inq-0041`:

```text
My dad slipped on scree on the Beehive descent, we're just past the ladder section. His ankle is swollen bad, can't put weight on it. We have water and I have 2 bars of signal. 2 adults 1 teen. How do we get help up here? Submitting this because 911 kept dropping.
```

And for `inq-0035`:

```text
Hi, I have a backcountry permit that includes a night at the Avalanche Lake area on June 24 (conf #GL-2026-07733). With the bridge out, is my itinerary even doable, and if not, will you let me swap that night for a different site without penalty, or refund it? I need to know before we leave Thursday. Thanks, Priya
```

From `starter/`:

```bash
dotnet run -- inq-0005
dotnet run -- inq-0041
dotnet run -- inq-0035
```

**Why:** `inq-0035` now lands in `unsure`, while step 0's free-text version called it `conditions`. The enum made `unsure` a real choice for the model.

**Check:** the parsed result's `category` is `conditions` for `inq-0005`, `emergency` for `inq-0041`, and `unsure` for `inq-0035`. Serialize the result back to JSON to see the wire form. How the key is spelled and spaced differs by track, and your track's block above says which you get; the value is the part to check, and it is always one of the seven strings.

### Step 3: Classify all 20 in a loop

**Do:**
1. Replace the starter's single-id lookup with a loop over the step 1 list.
2. For each inquiry, build the step 2 prompt with that inquiry's `text` after `Message:`, send it with the same schema and temperature 0.
3. Read `category` from the response, store the (inquiry, category) pair in a results list.
4. Print a `.` after each call to show progress, then a blank line after the loop.

For item 1, delete the starter's single-id lookup. These are the lines to remove (the `inquiries` list from step 1 replaces them, and `complete/` takes no id argument):

```csharp
var wantedId = args.Length > 0 ? args[0] : "inq-0005";
var inquiry = File.ReadLines("../../data/inquiries-slice.jsonl")
    .Select(line => JsonSerializer.Deserialize<Inquiry>(line)!)
    .First(i => i.id == wantedId);
```

Also delete step 2's single call and its two prints; the loop below replaces them.

For item 2, turn the starter's `var prompt = $"""` string into a function `Prompt(string text)` that puts `text` after `Message:`. Delete the `var prompt` block from the top-level code and paste this at the bottom of the file, above `record Inquiry(...)`. Inside it the message is `{text}` instead of `{inquiry.text}`, because there is no `inquiry` variable in a static function:

```csharp
static string Prompt(string text) => $"""
    You are the triage system for the Trailhead Guides shared inbox.
    Classify the visitor message into exactly one category.

    - permit: reserving, changing, canceling, or paying for a permit, pass,
      or reservation, including billing problems and missing confirmations
      for a permit application.
    - conditions: asking whether a trail, road, or area is open, safe, or
      passable right now: snow, water levels, washouts, wildlife activity,
      closures.
    - complaint: unhappy about a park facility, service, or staff member
      and wants it acknowledged or fixed.
    - lost-and-found: reporting a lost or found physical item.
    - emergency: a person may be hurt, missing, or in danger right now and
      needs immediate human attention.
    - general: anything else: park rules, fees, trip planning, questions
      that fit none of the above.
    - unsure: two different queues both have to act before this message can
      be resolved, so no single queue owns it. The case that qualifies: the
      sender asks about trail conditions AND asks someone to change, refund,
      or cancel a booking. Trail info cannot issue a refund, and the permits
      office does not decide whether a trail is passable, so a human reads
      this queue and splits the work. Also use unsure when the message fits
      none of the categories above.

    Decide in this order. First, if anyone might be hurt, missing, or in
    danger, answer emergency and stop; never answer unsure for those, even
    when the message also mentions permits, conditions, or a lost item.
    Second, if one queue can resolve the whole message on its own, answer
    that queue; a booking or reservation problem with nothing else attached
    is permit, not unsure. Third, only if two queues must both act, answer
    unsure. Unsure is not a catch-all for anything hard.

    Message:
    {text}
    """;
```

For items 3 and 4, declare `results` as a list of `(Inquiry Inquiry, Category Category)` pairs (a tuple with two named parts) before the loop, and print a dot inside it. Put this where the deleted single call was, below the step 1 load statements and the `options` line. `Console.WriteLine('\n')` prints the newline character plus the line end, which gives the blank line:

```csharp
var results = new List<(Inquiry Inquiry, Category Category)>();
foreach (var inquiry in inquiries)
{
    var response = await client.GetResponseAsync<TriageResult>(Prompt(inquiry.text), options);
    results.Add((inquiry, response.Result.Category));
    Console.Write('.');
}
Console.WriteLine('\n');
```

**Check:** 20 dots, then 20 stored pairs, every category one of the seven strings. The run takes under a minute.

### Step 4: Print emergencies first, then the routing table

**Do:**
1. Filter results to `category == "emergency"`.
2. If any, print `!!! EMERGENCY: route to dispatch, page the duty ranger now !!!`, then one line per emergency with `!!! `, the id, and the first 70 characters of the text. Blank line after.
3. Print a header line (`id`, `category`, `routed to`) and a rule of 62 dashes.
4. Sort so emergencies come first. For each pair, print the id padded to 10 characters, category padded to 15, and the queue from `routing` for that category. If your category is an enum, convert it back to its JSON name (`lost-and-found`, not `LostAndFound`) before the `routing` lookup, and reuse that helper in step 5.

`routing` and `labels` are both keyed by the JSON names. The enum has to go back to its wire string before either lookup, or `routing[...]` throws on the first `lost-and-found` row and accuracy scores 0/20. One helper does it; use it in the routing table print, the accuracy count, and the miss lines. Paste it at the bottom of the file next to `Prompt`:

```csharp
static string Wire(Category c) => c switch
{
    Category.LostAndFound => "lost-and-found",
    _ => c.ToString().ToLowerInvariant(),
};
```

Items 1 and 2, the emergency block. It goes right below the loop's `Console.WriteLine('\n');`. `Where` keeps the matching pairs, and `foreach (var (inquiry, _) in ...)` splits each pair into its parts, with `_` ignoring the category:

```csharp
var emergencies = results.Where(r => r.Category == Category.Emergency).ToList();
if (emergencies.Count > 0)
{
    Console.WriteLine("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!");
    foreach (var (inquiry, _) in emergencies)
        Console.WriteLine($"!!! {inquiry.id}  {Clip(inquiry.text, 70)}");
    Console.WriteLine();
}
```

Items 3 and 4, the routing table, directly below the emergency block. In `{inquiry.id,-10}` the `,-10` pads the value with spaces to 10 characters, left-aligned. `OrderBy(r => r.Category != Category.Emergency)` sorts on `false` before `true`, so emergencies come first:

```csharp
Console.WriteLine($"{"id",-10} {"category",-15} routed to");
Console.WriteLine(new string('-', 62));
foreach (var (inquiry, category) in results.OrderBy(r => r.Category != Category.Emergency))
    Console.WriteLine($"{inquiry.id,-10} {Wire(category),-15} {reference.Routing[Wire(category)]}");
```

`Clip(text, max)` cuts the text at `max` characters and appends `...`. It is another helper for the bottom of the file:

```csharp
static string Clip(string text, int max) =>
    text.Length <= max ? text : text[..max] + "...";
```

**Check:** two lines in the emergency block, `inq-0013` and `inq-0041`, both before the table. The table has 20 rows. `inq-0035` reads `unsure` and routes to `human-review-queue (a ranger reads it and picks the queue)`.

### Step 5: Score accuracy, then score emergency recall separately

**Do:**
1. Count results where `category == labels[id]`. That count is the accuracy numerator.
2. Collect ids in `labels` valued `emergency`; count how many appear in the step 4 emergency list. That count is emergency recall.
3. Print `Accuracy vs reference labels: N/20` and `Emergency recall: N/2`.
4. For each mismatch, print `miss: ` plus the id, the model's category, and the reference category.

This goes below the routing table, at the end of the top-level code. `results.Count(r => ...)` counts the pairs where the test is true (item 1). The `emergencyIds` line keeps the `labels` entries whose value is `emergency` and takes their keys (item 2). `Wire` turns the enum back into the JSON name before each comparison:

```csharp
var correct = results.Count(r => Wire(r.Category) == reference.Labels[r.Inquiry.id]);
var emergencyIds = reference.Labels.Where(l => l.Value == "emergency").Select(l => l.Key).ToList();
var caught = emergencies.Count(e => emergencyIds.Contains(e.Inquiry.id));

Console.WriteLine();
Console.WriteLine($"Accuracy vs reference labels: {correct}/{results.Count}");
Console.WriteLine($"Emergency recall: {caught}/{emergencyIds.Count} " +
    (caught == emergencyIds.Count ? "(all caught; the metric that matters)" : "(MISSED ONE; this fails, whatever the accuracy says)"));
foreach (var (inquiry, category) in results.Where(r => Wire(r.Category) != reference.Labels[r.Inquiry.id]))
    Console.WriteLine($"  miss: {inquiry.id} got {Wire(category)}, reference says {reference.Labels[inquiry.id]}");
```

Run from `starter/`:

```bash
dotnet run
```

**Why:** accuracy is the headline number, but emergency recall is the number that decides whether this is safe to ship. A missed emergency is a person waiting in a queue nobody is watching.

**Check:** one recorded run scored 18/20 with two misses: `inq-0030` (wedding photographer) got `general`, reference `permit`; `inq-0051` (Sperry campfires) got `conditions`, reference `general`. Another recorded 17/20 with one more: `inq-0001` got `unsure`, reference `permit`. Either way, emergency recall is 2/2, no routine message is `emergency`, and `inq-0035` is `unsure`. An emergency anywhere but `emergency` fails, whatever the accuracy.

### Step 6: Fix the misses by editing the category descriptions, not the code

`complete/` stops at step 5 with the original descriptions. This step's edits are yours, and the Check below is the only answer key.

**Do:**
1. Extend the `permit` description to also cover whether an activity needs a permit at all:

```text
- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, and questions about whether an activity requires a permit at all, including billing problems and missing confirmations.
```

2. Narrow `conditions` to mean only whether a trail, road, or area is physically passable.
3. Let `general` own park rules and regulations.
4. Leave `Decide in this order` and the `unsure` description alone.
5. Run all 20 again and read the scoreboard.

All four edits happen inside the `static string Prompt(string text) => $"""` string at the bottom of `Program.cs`. It is a raw string, so each description is plain text: keep every line indented at least as far as the closing `"""`, and a description may wrap onto more lines. For item 1, replace the three `- permit:` lines with the new wording:

```csharp
// Hint: the permit lines inside Prompt, rewritten (wrap where you like)
    - permit: reserving, changing, canceling, or paying for a permit, pass,
      or reservation, and questions about whether an activity requires a
      permit at all, including billing problems and missing confirmations.
```

Items 2 and 3 are the same kind of edit on the `- conditions:` and `- general:` lines. The wording is yours:

```csharp
// Hint: same shape, your words
    - conditions: <only whether a trail, road, or area is physically passable>
    - general: <park rules and regulations, plus what general already covers>
```

Then from `starter/`:

```bash
dotnet run
```

This program prints the routing table, not JSON. Where the Check below says `{"category": "emergency"}` or `{"category": "unsure"}`, read the `category` column of the table for that id.

**Why:** the misses came from the wording, not the model. `permit` described only transactions on a reservation, so a question about whether a permit is needed read as trip planning, and `conditions` was loose enough to swallow a question about campfire rules. When a message lands in the wrong queue, sharpen the description of the queue that should have owned it before you touch the code.

**Check:** when I ran it, `inq-0051` (Sperry campfires) moved to `general` and `inq-0001` moved from `unsure` to `permit`, while `inq-0008` (Half Dome lottery) stayed `permit`. Those two moves are the ones your edit is aiming at, and they do not land every run (your results will vary; the model is non-deterministic). `inq-0030` (wedding photographer) usually stays `general` on `llama3.2`, even with the permit description above; a commercial-photography permit is a hard call for a small model, so treat it as a known miss rather than a sign your edit failed. `inq-0005` may move to `unsure`. `inq-0041` and `inq-0013` still return `{"category": "emergency"}`, and `inq-0035` still returns `{"category": "unsure"}`. At most one or two messages besides `inq-0035` sit in `unsure`. Accuracy lands between 17 and 19 out of 20; measured on `llama3.2`, the .NET and Python runs scored 18/20, missing `inq-0030` and `inq-0005`, and the TypeScript runs scored 19/20 because `inq-0005` stayed `conditions`. A run at 20/20 means check whether the descriptions now fit only these 20 messages. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves. `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goals

Pick either. Neither is built in `complete/`. The reasoning is in [`expected-output.md`](../expected-output.md) under "Stretch Goal".

- **Add a priority field.** Add `priority` to the result type next to `category`, with its own small set of allowed values. Add a line to the prompt that says what each priority value means (for example, high when someone's safety or health is at risk), or the model rates almost everything high. Print it in the routing table. Priority is a second axis. It says how fast, and category says where. Mixing the two is how a lost inhaler ends up in line behind a lost wedding ring. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority.

  Give priority its own enum, built like `Category`, and add it to the record. Both go at the bottom of the file:

  ```csharp
  // Hint: a second field on the result record, and a second small enum
  record TriageResult(Category Category, Priority Priority);

  [JsonConverter(typeof(JsonStringEnumConverter<Priority>))]
  enum Priority
  {
      [JsonStringEnumMemberName("<value>")] <Member>,
      [JsonStringEnumMemberName("<value>")] <Member>,
  }
  ```

  Then carry it through the loop and the table. The tuple gains a third part, and every line that builds or splits a pair changes to match:

  ```csharp
  // Hint: three parts instead of two
  var results = new List<(Inquiry Inquiry, Category Category, Priority Priority)>();
  results.Add((inquiry, response.Result.Category, response.Result.Priority));
  foreach (var (inquiry, category, priority) in results.OrderBy(r => r.Category != Category.Emergency))
      Console.WriteLine($"{inquiry.id,-10} {Wire(category),-15} {priority,-8} {reference.Routing[Wire(category)]}");
  ```

  The step 4 emergency `foreach` and the step 5 miss `foreach` split the pair too; give them a third `_` or name.

- **Add a confidence threshold.** Add a numeric `confidence` field to the result type and schema. After the loop, change the category to `unsure` on any result whose confidence is under a threshold you pick. Run the scoreboard again. **Check:** both emergencies stay `emergency` with high confidence, `inq-0035` stays `unsure`, and the `unsure` queue does not fill up with ordinary permit questions. If a third of the slice lands in `unsure`, the threshold is too high and you have rebuilt the unsorted inbox.

  A `double` on the record becomes a number in the generated schema:

  ```csharp
  // Hint: a number field next to Category
  record TriageResult(Category Category, double Confidence);
  ```

  Store the confidence in the tuple the way the priority stretch stores priority. The tuple gains a third part, and every line that builds or splits a pair changes to match. Printing the confidence in the routing table lets you check it:

  ```csharp
  // Hint: three parts instead of two
  var results = new List<(Inquiry Inquiry, Category Category, double Confidence)>();
  results.Add((inquiry, response.Result.Category, response.Result.Confidence));
  foreach (var (inquiry, category, confidence) in results.OrderBy(r => r.Category != Category.Emergency))
      Console.WriteLine($"{inquiry.id,-10} {Wire(category),-15} {confidence,-5:0.00} {reference.Routing[Wire(category)]}");
  ```

  The step 4 emergency `foreach` and the step 5 miss `foreach` split the pair too; give them a third `_` or name.

  Then rewrite `results` after the loop and before the emergency block. `results` is declared with `var`, so it can be reassigned:

  ```csharp
  // Hint: after the loop, pick your own threshold
  const double threshold = <your number>;
  results = results
      .Select(r => r.Confidence < threshold ? (r.Inquiry, Category.Unsure, r.Confidence) : r)
      .ToList();
  ```

## What Is in This Folder

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder, part of the workshop corpus and shared with feature 09), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix matches the full inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the hand-assigned category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus, the source the reference labels were drawn from. Not for handout.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
