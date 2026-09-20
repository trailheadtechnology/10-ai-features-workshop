<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 07: Classification & Routing (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F07-dotnet.md), [Python](../python/F07-python.md). Lab overview: [F07-lab.md](../F07-lab.md).*

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

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Swapping the provider later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. `starter/index.ts` classifies a single inquiry and prints the free-text label. `complete/index.ts` classifies the whole slice through structured output into a zod enum, prints emergencies first, and scores accuracy and emergency recall.

Setup once (`npm install` in the `typescript/` folder, where `package.json` lives), then run everything from that folder:

```bash
npm run complete             # all 20, scored
npm run starter -- inq-0013  # one inquiry, free text
```

`complete/` takes no arguments: keep the starter's positional id through step 2 and drop it when step 3 loops the slice.

### Step 0: Run the starter and read the free-text label

**Do:** run `starter/` as it is. It reads `inq-0005` from the slice, builds the taxonomy prompt, puts the message after `Message:`, and sends it to `llama3.2` as one user message. Then run it again with `inq-0035` as the argument, and again with `inq-0013`.

```bash
npm run starter
npm run starter -- inq-0035
npm run starter -- inq-0013
```

The starter already does this. It picks the id from the command line (`process.argv[2]`, or `inq-0005` when you give none; `??` means "use the right side when the left is missing"), reads the slice, splits it into lines, skips blank ones, parses each with `JSON.parse`, and keeps the one whose `id` matches:

```typescript
const wanted = process.argv[2] ?? "inq-0005";
const inquiry: Inquiry = readFileSync(resolve(DATA, "inquiries-slice.jsonl"), "utf8").split("\n")
  .filter((l) => l.trim()).map((l) => JSON.parse(l)).find((i: Inquiry) => i.id === wanted);
```

The starter already does this too. The prompt is one big backtick template string that ends with the message text (`${inquiry.text}` drops the value into the string), and the call sends it as one user message and prints whatever text comes back. The `await` sits at the top level of the file, outside any function, which works because `package.json` sets `"type": "module"`:

```typescript
Message:
${inquiry.text}`;

const response = await client.chat.completions.create({ model: "llama3.2", messages: [{ role: "user", content: prompt }] });
console.log(`${inquiry.id}: ${response.choices[0].message.content}`);
```

**Check:** one label per run, printed after the id. `inq-0013` says `emergency`. `inq-0035` usually says `conditions`, and some runs say `unsure`, `emergency`, or `complaint`: measured over 30 runs across the three tracks, 24 `conditions`, 4 `unsure`, 1 `emergency`, 1 `complaint`. That spread on one message is the problem. Nothing in the starter stops a run from printing a label that is not one of the seven names, such as `Emergency.` or a sentence, either. The rest of the lab closes both gaps.

### Step 1: Load the 20 inquiries and the reference labels

**Do:**
1. Open `../../data/inquiries-slice.jsonl`. Every line is one JSON object with `id`, `channel`, `received`, `text`. Read it line by line, parse each line, keep the results in a list, and skip any blank line so a trailing newline cannot become an empty record. Your track's block below says how to point at that `data/` folder and how it handles the blank-line case; use the same way for both files.

   The starter resolves the data folder as `DATA`, and `resolve(DATA, "inquiries-slice.jsonl")` is the file inside it. This one line reads the whole file as text, `.split("\n")` cuts it into lines, `.filter((l) => l.trim())` drops blank lines, and `.map((l) => JSON.parse(l))` turns each line into an object, giving an array of `Inquiry`. Put it right below the starter's `const inquiry: Inquiry = ...` lookup (the two lines ending in `.find(...)`). Leave the starter's `wanted`/`inquiry` lines in place for now; step 3 removes them.

   ```typescript
   const inquiries: Inquiry[] = readFileSync(resolve(DATA, "inquiries-slice.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
   ```
2. Open `../../data/reference-labels.json` and parse it: an object with `routing` (category to queue name), `labels` (id to correct category), and `notes` (why `inq-0013`/`inq-0041` are emergencies and why `inq-0035` is `unsure`). Keep the `routing` and `labels` dictionaries.

   Put this directly below the `inquiries` line. The part after `const reference:` is only a type label: an object with two dictionaries, where `Record<string, string>` means "string keys, string values". `reference.routing` and `reference.labels` are the two dictionaries; there is nothing else to unpack.

   ```typescript
   const reference: { routing: Record<string, string>; labels: Record<string, string> } = JSON.parse(readFileSync(resolve(DATA, "reference-labels.json"), "utf8"));
   ```

   To see the Check numbers, print them once below the two load lines (delete the line afterward). `Object.keys(...)` gives an array of a dictionary's keys, so its `.length` is the key count:

   ```typescript
   // Hint: a throwaway print for the Check
   console.log(inquiries.length, inquiries[0].id, Object.keys(reference.routing).length, Object.keys(reference.labels).length);
   ```

   Run it from the `typescript/` folder:

   ```bash
   npm run starter
   ```

**Why:** these 20 are drawn from the fictional Trailhead Guides 100-message inbox, chosen to mirror the inbox's mix, including both emergencies and the one ambiguous message. `unsure` was added to the taxonomy, labels, and routing table together once the ambiguous message needed a queue.

**Check:** the inquiry list has 20 entries. The first `id` is `inq-0001`. `routing` has 7 keys and `labels` has 20.

### Step 2: Pin the answer to the seven labels with structured output

**Do:**
1. Delete the line `Answer with the category name only.` from the starter's prompt. The schema below does that job now.

   In `index.ts` it is this line near the end of the `` const prompt = ` `` template string, just above `Message:`. Delete it and the blank line under it:

   ```typescript
   Answer with the category name only.
   ```
2. Define the category as an enum type with exactly seven allowed values, wrapped in a result type with one field, `category`. If your language cannot spell `lost-and-found` as an identifier, map that member to the exact JSON name. The generated schema must list `lost-and-found`, or three of the 20 messages can never match their reference label.

   A zod enum takes the string values directly, so `lost-and-found` needs no mapping and the generated schema lists it as written. Two imports the starter does not have: `import { zodResponseFormat } from "openai/helpers/zod";` and `import { z } from "zod";` (both packages are in the `typescript/` folder's `package.json`).

   Add the two imports at the top of `index.ts`, right below `import OpenAI from "openai";`:

   ```typescript
   import { zodResponseFormat } from "openai/helpers/zod";
   import { z } from "zod";
   ```

   Then paste these three lines below the `const DATA = ...` line, above everything that uses them. `z.enum([...])` builds a zod schema that accepts only those seven strings. `type Category = z.infer<typeof Category>;` asks zod for the matching TypeScript type (the union `"permit" | "conditions" | ...`), so the same name works as a value (the schema) and as a type. `z.object({ category: Category })` is the schema for an object with one field, `category`, that must be one of the seven:

   ```typescript
   const Category = z.enum(["permit", "conditions", "complaint", "lost-and-found", "emergency", "general", "unsure"]);
   type Category = z.infer<typeof Category>;
   const TriageResult = z.object({ category: Category });
   ```

3. Send the message through your chat client with that result type as the structured-output schema, temperature 0.

   The typed call is `client.chat.completions.parse` with `response_format: zodResponseFormat(TriageResult, "triage")`, `temperature: 0` is a property on that same options object, and the parsed object is `response.choices[0].message.parsed` (typed as possibly null, hence the `!`). Keep the starter's single-id lookup and swap the call. `JSON.stringify(parsed)` prints the wire form the step 2 Check describes.

   Replace the starter's last two lines (`const response = await client.chat.completions.create(...)` and the `console.log` under it) with this. `zodResponseFormat(TriageResult, "triage")` turns the zod schema into the JSON schema the API expects (`"triage"` is just a name for it), and `.parse` checks the reply against that schema and hands back a typed object. `JSON.stringify` writes the wire form without the space, `{"category":"conditions"}`; the value is the part to check:

   ```typescript
   // Hint: the typed call on the one inquiry, then two prints
   const response = await client.chat.completions.parse({
     model: "llama3.2",
     messages: [{ role: "user", content: prompt }],
     response_format: zodResponseFormat(TriageResult, "triage"),
     temperature: 0,
   });
   const parsed = response.choices[0].message.parsed!;
   console.log(`${inquiry.id}: ${parsed.category}`);
   console.log(JSON.stringify(parsed));
   ```

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

The starter already does this: its `` const prompt = ` `` template string holds that exact text, and it ends by putting the message after `Message:`. After item 1 the only change is the deleted line, so the end of the string still reads:

```typescript
Message:
${inquiry.text}`;
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

From the `typescript/` folder:

```bash
npm run starter -- inq-0005
npm run starter -- inq-0041
npm run starter -- inq-0035
```

**Why:** `inq-0035` now lands in `unsure`, while step 0's free-text version called it `conditions`. The enum made `unsure` a real choice for the model.

**Check:** the parsed result's `category` is `conditions` for `inq-0005`, `emergency` for `inq-0041`, and `unsure` for `inq-0035`. Serialize the result back to JSON to see the wire form. How the key is spelled and spaced differs by track, and your track's block above says which you get; the value is the part to check, and it is always one of the seven strings.

### Step 3: Classify all 20 in a loop

**Do:**
1. Replace the starter's single-id lookup with a loop over the step 1 list.

   Delete the starter's single-id lookup. These are the three lines to remove (the `inquiries` array from step 1 replaces them, so the program no longer takes an id argument):

   ```typescript
   const wanted = process.argv[2] ?? "inq-0005";
   const inquiry: Inquiry = readFileSync(resolve(DATA, "inquiries-slice.jsonl"), "utf8").split("\n")
     .filter((l) => l.trim()).map((l) => JSON.parse(l)).find((i: Inquiry) => i.id === wanted);
   ```

   Also delete step 2's single `parse` call, its `const parsed = ...` line, and the two prints; the loop in item 3 replaces them. Keep the `type Inquiry = ...` line; the loop still uses it.
2. For each inquiry, build the step 2 prompt with that inquiry's `text` after `Message:`, send it with the same schema and temperature 0.

   Turn the starter's `prompt` string into a function `prompt(text)` that puts `text` after `Message:` (as `complete/` does). Two lines change. The first line of the template string becomes an arrow function: `(text: string) => ...` takes one string and returns the template string built from it. Replace the starter's first prompt line with:

   ```typescript
   const prompt = (text: string) => `You are the triage system for the Trailhead Guides shared inbox.
   ```

   The end of the string uses `${text}` instead of `${inquiry.text}`, because inside the function there is no `inquiry` variable. Replace the last line of the string with:

   ```typescript
   ${text}`;
   ```

   Everything between those two lines stays as it is, at the left edge; every space inside the backticks is sent to the model. A `const` has to be defined above the line that first uses it, and the prompt already sits above where the loop goes.

   The loop in item 3 uses `MODEL`. Put this constant right below the `const client = new OpenAI(...)` line:

   ```typescript
   const MODEL = "llama3.2";
   ```
3. Read `category` from the response, store the (inquiry, category) pair in a results list.

   Put the loop at the end of the file, below the step 1 load lines. `const results: { inquiry: Inquiry; category: Category }[] = [];` makes an empty array; the part after the colon is only a type label saying each entry is an object with an `inquiry` and a `category`. `const` stops you from pointing `results` at a different array, but `results.push(...)` can still add to it. `{ inquiry, category: ... }` is short for `{ inquiry: inquiry, category: ... }`:

   ```typescript
   const results: { inquiry: Inquiry; category: Category }[] = [];
   for (const inquiry of inquiries) {
     const response = await client.chat.completions.parse({
       model: MODEL,
       messages: [{ role: "user", content: prompt(inquiry.text) }],
       response_format: zodResponseFormat(TriageResult, "triage"),
       temperature: 0,
     });
     results.push({ inquiry, category: response.choices[0].message.parsed!.category });
     process.stdout.write(".");
   }
   console.log("\n");
   ```
4. Print a `.` after each call to show progress, then a blank line after the loop.

   The item 3 loop already does this. `process.stdout.write(".")` prints a dot with no line break, unlike `console.log`. `console.log("\n")` after the loop prints a newline plus its own line end, which gives the blank line. Run it from the `typescript/` folder:

   ```bash
   npm run starter
   ```

**Check:** 20 dots, then 20 stored pairs, every category one of the seven strings. The run takes under a minute.

### Step 4: Print emergencies first, then the routing table

**Do:**
1. Filter results to `category == "emergency"`.

   Put this below the loop's `console.log("\n");`. `.filter(...)` walks `results` and returns a new array with only the entries where the test is true. `(r) => r.category === "emergency"` is that test; `r` is one `{ inquiry, category }` entry. A zod enum value is already the plain string, so comparing to `"emergency"` works with no conversion:

   ```typescript
   const emergencies = results.filter((r) => r.category === "emergency");
   ```
2. If any, print `!!! EMERGENCY: route to dispatch, page the duty ranger now !!!`, then one line per emergency with `!!! `, the id, and the first 70 characters of the text. Blank line after.

   `emergencies.length > 0` is true when the array is not empty. In `for (const { inquiry } of emergencies)`, the `{ inquiry }` pulls the `inquiry` field out of each entry and ignores `category`. Put this directly below the `emergencies` line:

   ```typescript
   if (emergencies.length > 0) {
     console.log("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!");
     for (const { inquiry } of emergencies) console.log(`!!! ${inquiry.id}  ${clip(inquiry.text, 70)}`);
     console.log();
   }
   ```

   `clip(text, n)` is a one-line helper in `complete/` that cuts the text at `n` characters and appends `...`. `a ? b : c` means "if `a`, then `b`, else `c`", and `text.slice(0, n)` is the first `n` characters. Paste it below the `prompt` function (after its closing `` `; `` line), above the load lines:

   ```typescript
   const clip = (text: string, n: number) => (text.length <= n ? text : text.slice(0, n) + "...");
   ```
3. Print a header line (`id`, `category`, `routed to`) and a rule of 62 dashes.

   Directly below the emergency block. `"id".padEnd(10)` returns `id` with spaces added on the right until it is 10 characters long. `"-".repeat(62)` repeats the dash 62 times:

   ```typescript
   console.log(`${"id".padEnd(10)} ${"category".padEnd(15)} routed to`);
   console.log("-".repeat(62));
   ```
4. Sort so emergencies come first. For each pair, print the id padded to 10 characters, category padded to 15, and the queue from `routing` for that category. If your category is an enum, convert it back to its JSON name (`lost-and-found`, not `LostAndFound`) before the `routing` lookup, and reuse that helper in step 5.

   `reference.routing` and `reference.labels` are keyed by the JSON names, and a zod enum value already is that string (`lost-and-found`), so no helper is needed: `reference.routing[category]` works as is.

   `.sort(...)` changes the array it is called on, so `[...results]` makes a copy first and the original order stays for step 5. The comparator `(a, b) => ...` gets two entries and returns a negative number when `a` should come first, positive when `b` should. `Number(a.category !== "emergency")` is `0` for an emergency and `1` for everything else, so subtracting the two puts emergencies on top. `for (const { inquiry, category } of ...)` pulls both fields out of each entry. Put this directly below the header lines:

   ```typescript
   for (const { inquiry, category } of [...results].sort((a, b) => Number(a.category !== "emergency") - Number(b.category !== "emergency"))) {
     console.log(`${inquiry.id.padEnd(10)} ${category.padEnd(15)} ${reference.routing[category]}`);
   }
   ```

**Check:** two lines in the emergency block, `inq-0013` and `inq-0041`, both before the table. The table has 20 rows. `inq-0035` reads `unsure` and routes to `human-review-queue (a ranger reads it and picks the queue)`.

### Step 5: Score accuracy, then score emergency recall separately

**Do:**
1. Count results where `category == labels[id]`. That count is the accuracy numerator.

   This goes below the routing table, at the end of the file. `.filter(...)` keeps the entries where the test is true, and `.length` counts them. `labels[r.inquiry.id]` is the reference category for that inquiry:

   ```typescript
   const labels = reference.labels;
   const correct = results.filter((r) => r.category === labels[r.inquiry.id]).length;
   ```
2. Collect ids in `labels` valued `emergency`; count how many appear in the step 4 emergency list. That count is emergency recall.

   `Object.entries(labels)` gives an array of `[id, category]` pairs. In `([, v]) => v === "emergency"` the leading comma skips the id and names the category `v`; `.map(([k]) => k)` then keeps only the id. The second line counts the step 4 emergencies whose id is in that list (`.includes` is true when the array holds the value). Put both below the `correct` line:

   ```typescript
   const emergencyIds = Object.entries(labels).filter(([, v]) => v === "emergency").map(([k]) => k);
   const caught = emergencies.filter((e) => emergencyIds.includes(e.inquiry.id)).length;
   ```
3. Print `Accuracy vs reference labels: N/20` and `Emergency recall: N/2`.

   Directly below. `complete/` adds a verdict after the recall numbers; the `+` joins two strings, and `test ? A : B` picks one of two texts:

   ```typescript
   console.log();
   console.log(`Accuracy vs reference labels: ${correct}/${results.length}`);
   console.log(`Emergency recall: ${caught}/${emergencyIds.length} ` +
     (caught === emergencyIds.length ? "(all caught; the metric that matters)" : "(MISSED ONE; this fails, whatever the accuracy says)"));
   ```
4. For each mismatch, print `miss: ` plus the id, the model's category, and the reference category.

   At the very end of the file. The `.filter(...)` keeps only the mismatches, and the loop prints one line for each:

   ```typescript
   for (const { inquiry, category } of results.filter((r) => r.category !== labels[r.inquiry.id])) {
     console.log(`  miss: ${inquiry.id} got ${category}, reference says ${labels[inquiry.id]}`);
   }
   ```

Run from the `typescript/` folder:

```bash
npm run starter
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

All four edits happen inside the `prompt` function's template string in `index.ts`. Each description is plain text; a description may wrap onto more lines, and the wrapped lines keep their two leading spaces like the ones around them. For item 1, replace the three `- permit:` lines with the new wording. Paste only the three text lines, not the `// Hint:` line, since anything inside the backticks is sent to the model:

```typescript
// Hint: the permit lines inside prompt, rewritten (wrap where you like)
- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, and questions about whether an activity requires a
  permit at all, including billing problems and missing confirmations.
```

Items 2 and 3 are the same kind of edit on the `- conditions:` and `- general:` lines. The wording is yours:

```typescript
// Hint: same shape, your words
- conditions: <only whether a trail, road, or area is physically passable>
- general: <park rules and regulations, plus what general already covers>
```

Then from the `typescript/` folder:

```bash
npm run starter
```

This program prints the routing table, not JSON. Where the Check below says `{"category": "emergency"}` or `{"category": "unsure"}`, read the `category` column of the table for that id.

**Why:** the misses came from the wording, not the model. `permit` described only transactions on a reservation, so a question about whether a permit is needed read as trip planning, and `conditions` was loose enough to swallow a question about campfire rules. When a message lands in the wrong queue, sharpen the description of the queue that should have owned it before you touch the code.

**Check:** when I ran it, `inq-0051` (Sperry campfires) moved to `general` and `inq-0001` moved from `unsure` to `permit`, while `inq-0008` (Half Dome lottery) stayed `permit`. Those two moves are the ones your edit is aiming at, and they do not land every run (your results will vary; the model is non-deterministic). `inq-0030` (wedding photographer) usually stays `general` on `llama3.2`, even with the permit description above; a commercial-photography permit is a hard call for a small model, so treat it as a known miss rather than a sign your edit failed. `inq-0005` may move to `unsure`. `inq-0041` and `inq-0013` still return `{"category": "emergency"}`, and `inq-0035` still returns `{"category": "unsure"}`. At most one or two messages besides `inq-0035` sit in `unsure`. Accuracy lands between 17 and 19 out of 20; measured on `llama3.2`, the .NET and Python runs scored 18/20, missing `inq-0030` and `inq-0005`, and the TypeScript runs scored 19/20 because `inq-0005` stayed `conditions`. A run at 20/20 means check whether the descriptions now fit only these 20 messages. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves. `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goals

Pick either. Neither is built in `complete/`. The reasoning is in [`expected-output.md`](../expected-output.md) under "Stretch Goal".

- **Add a priority field.** Add `priority` to the result type next to `category`, with its own small set of allowed values. Add a line to the prompt that says what each priority value means (for example, high when someone's safety or health is at risk), or the model rates almost everything high. Print it in the routing table. Priority is a second axis. It says how fast, and category says where. Mixing the two is how a lost inhaler ends up in line behind a lost wedding ring. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority.

  Give priority its own zod enum, built like `Category`, and add a second field to `TriageResult`. Both go where the step 2 lines are, and `Priority` must sit above `TriageResult`, which uses it:

  ```typescript
  // Hint: a second small zod enum, and a second field on the result schema
  const Priority = z.enum(["<value>", "<value>"]);
  type Priority = z.infer<typeof Priority>;
  const TriageResult = z.object({ category: Category, priority: Priority });
  ```

  The prompt line goes inside the `prompt` template string, above `Decide in this order`, with a blank line on each side. Keep it general and say it is separate from category. A test run whose line named a specific item ("a lost medicine") pulled `inq-0006` into `emergency`; the plain wording kept it `lost-and-found` with `high`:

  ```typescript
  // Hint: one line in the prompt, your words
  Priority is separate from category: <value> when <meaning>; <value> for everything else.
  ```

  Then carry it through the loop and the table. Each entry gains a `priority` field, and the lines that build or print an entry change to match:

  ```typescript
  // Hint: a third field on each entry
  const results: { inquiry: Inquiry; category: Category; priority: Priority }[] = [];
  const parsed = response.choices[0].message.parsed!;
  results.push({ inquiry, category: parsed.category, priority: parsed.priority });
  for (const { inquiry, category, priority } of [...results].sort((a, b) => Number(a.category !== "emergency") - Number(b.category !== "emergency"))) {
    console.log(`${inquiry.id.padEnd(10)} ${category.padEnd(15)} ${priority.padEnd(8)} ${reference.routing[category]}`);
  }
  ```

  The `const parsed` line and the new `results.push` replace the old `results.push` inside the loop. The step 4 emergency loop and the step 5 lines pull fields out by name (`{ inquiry }`, `r.category`), so they need no change.

- **Add a confidence threshold.** Add a numeric `confidence` field to the result type and schema. After the loop, change the category to `unsure` on any result whose confidence is under a threshold you pick. Run the scoreboard again. **Check:** both emergencies stay `emergency` with high confidence, `inq-0035` stays `unsure`, and the `unsure` queue does not fill up with ordinary permit questions. If a third of the slice lands in `unsure`, the threshold is too high and you have rebuilt the unsorted inbox.

  `z.number()` is zod's number type, and it becomes a number in the generated schema. It is not in `complete/`:

  ```typescript
  // Hint: a number field next to category
  const TriageResult = z.object({ category: Category, confidence: z.number() });
  ```

  Store the confidence on each entry the way the priority stretch stores priority. Printing it in the routing table lets you check it; `confidence.toFixed(2)` writes the number with 2 decimals:

  ```typescript
  // Hint: a third field on each entry
  const results: { inquiry: Inquiry; category: Category; confidence: number }[] = [];
  const parsed = response.choices[0].message.parsed!;
  results.push({ inquiry, category: parsed.category, confidence: parsed.confidence });
  for (const { inquiry, category, confidence } of [...results].sort((a, b) => Number(a.category !== "emergency") - Number(b.category !== "emergency"))) {
    console.log(`${inquiry.id.padEnd(10)} ${category.padEnd(15)} ${confidence.toFixed(2).padEnd(5)} ${reference.routing[category]}`);
  }
  ```

  Then change the low-confidence entries after the loop's `console.log("\n");` and before the `emergencies` line. `results` is a `const` array, but its entries can still be changed in place:

  ```typescript
  // Hint: after the loop, pick your own threshold
  const THRESHOLD = <your number>;
  for (const r of results) if (r.confidence < THRESHOLD) r.category = "unsure";
  ```

## What Is in This Folder

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder, part of the workshop corpus and shared with feature 09), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix matches the full inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the hand-assigned category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus, the source the reference labels were drawn from. Not for handout.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
