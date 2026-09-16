# Lab 07: Classification & Routing

*This is the Recommended lab for [Module 3](../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** sort visitor messages into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and send each one to the right queue.
- **Input:** `data/inquiries-slice.jsonl` (20 messages: `id`, `channel`, `received`, `text`; emergencies `inq-0013` and `inq-0041`, ambiguous `inq-0035`), `data/reference-labels.json` (the correct category for each id, and the queue for each category), `data/inquiries.jsonl` (the full 100-message inbox the 20 came from; the demo scrolls it, the lab does not read it).
- **How:** send each message to `llama3.2` with the taxonomy prompt. Use a JSON schema so the model can only answer with one of the seven labels. Loop all 20 at temperature 0. Print the routing table, then score the results against the reference labels.
- **Model:** `llama3.2`, local. No key.

Every step below is one thing to make the program do. Every track's `starter/` classifies one message and prints whatever text the model sends back. Edit it until it does all six steps. Compare against `complete/` when you get stuck.

### Step 0: Run the starter and read the free-text label

**Do:** run `starter/` as it is. It reads `inq-0005` from the slice, builds the taxonomy prompt, puts the message after `Message:`, and sends it to `llama3.2` as one user message. Then run it again with `inq-0035` as the argument, and again with `inq-0013`.

**Check:** one label per run, printed after the id. `inq-0013` says `emergency`. `inq-0035` flips between `conditions` and `permit` across runs. At least one run prints something that is not exactly one of the seven category names, such as `Emergency.` or a sentence. Nothing in the starter stops that — the rest of the lab does.

### Step 1: Load the 20 inquiries and the reference labels

**Do:**
1. Open `../../data/inquiries-slice.jsonl`. Every line is one JSON object with `id`, `channel`, `received`, `text`. Read it line by line, skip blanks, parse each line, keep the results in a list.
2. Open `../../data/reference-labels.json` and parse it: an object with `routing` (category → queue name), `labels` (id → correct category), and `notes` (why `inq-0013`/`inq-0041` are emergencies and why `inq-0035` is `unsure`). Keep the `routing` and `labels` dictionaries.

**Why:** these 20 are drawn from the fictional Trailhead Guides 100-message inbox, chosen to mirror the inbox's mix, including both emergencies and the one ambiguous message. `unsure` was added to the taxonomy, labels, and routing table together once the ambiguous message needed a queue.

**Check:** the inquiry list has 20 entries. The first `id` is `inq-0001`. `routing` has 7 keys and `labels` has 20.

### Step 2: Pin the answer to the seven labels with structured output

**Do:**
1. Delete the line `Answer with the category name only.` from the starter's prompt — the schema below does that job now.
2. Define the category as an enum type with exactly seven allowed values, wrapped in a result type with one field, `category`.
3. Send the message through your chat client with that result type as the structured-output schema, temperature 0.
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

**Why:** `inq-0035` now lands in `unsure`, while step 0's free-text version called it `conditions` — the enum made `unsure` a real choice for the model.

**Check:** `inq-0005` returns `{"category": "conditions"}`, `inq-0041` returns `{"category": "emergency"}`, `inq-0035` returns `{"category": "unsure"}`. The value is always one of the seven strings.

### Step 3: Classify all 20 in a loop

**Do:**
1. Replace the starter's single-id lookup with a loop over the step 1 list.
2. For each inquiry, build the step 2 prompt with that inquiry's `text` after `Message:`, send it with the same schema and temperature 0.
3. Read `category` from the response, store the (inquiry, category) pair in a results list.
4. Print a `.` after each call to show progress, then a blank line after the loop.

**Check:** 20 dots, then 20 stored pairs, every category one of the seven strings. The run takes under a minute.

### Step 4: Print emergencies first, then the routing table

**Do:**
1. Filter results to `category == "emergency"`.
2. If any, print `!!! EMERGENCY: route to dispatch, page the duty ranger now !!!`, then one line per emergency with `!!! `, the id, and the first 70 characters of the text. Blank line after.
3. Print a header line (`id`, `category`, `routed to`) and a rule of 62 dashes.
4. Sort so emergencies come first. For each pair, print the id padded to 10 characters, category padded to 15, and the queue from `routing` for that category.

**Check:** two lines in the emergency block, `inq-0013` and `inq-0041`, both before the table. The table has 20 rows. `inq-0035` reads `unsure` and routes to `human-review-queue (a ranger reads it and picks the queue)`.

### Step 5: Score accuracy, then score emergency recall separately

**Do:**
1. Count results where `category == labels[id]` — the accuracy numerator.
2. Collect ids in `labels` valued `emergency`; count how many appear in the step 4 emergency list — emergency recall.
3. Print `Accuracy vs reference labels: N/20` and `Emergency recall: N/2`.
4. For each mismatch, print `miss: ` plus the id, the model's category, and the reference category.

**Why:** accuracy is the headline number, but emergency recall is the number that decides whether this is safe to ship — a missed emergency is a person waiting in a queue nobody is watching.

**Check:** a raw-request run recorded 18/20 with two misses: `inq-0030` (wedding photographer) got `general`, reference `permit`; `inq-0051` (Sperry campfires) got `conditions`, reference `general`. The code tracks recorded 17/20 with one more: `inq-0001` got `unsure`, reference `permit`. Either way, emergency recall is 2/2, no routine message is `emergency`, and `inq-0035` is `unsure`. An emergency anywhere but `emergency` fails, whatever the accuracy.

### Step 6: Fix the misses by editing the category descriptions, not the code

**Do:**
1. Extend the `permit` description to also cover whether an activity needs a permit at all:

```text
- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, and questions about whether an activity requires a permit at all, including billing problems and missing confirmations.
```

2. Narrow `conditions` to mean only whether a trail, road, or area is physically passable.
3. Let `general` own park rules and regulations.
4. Leave `Decide in this order` and the `unsure` description alone.
5. Run all 20 again and read the scoreboard.

**Why:** the category descriptions in the prompt are where accuracy lives — when the model files something wrong, fix the description before touching the code.

**Check:** `inq-0030` moves to `permit` and `inq-0008` (Half Dome lottery) stays there. `inq-0051` moves to `general`. `inq-0041` and `inq-0013` still return `{"category": "emergency"}`, and `inq-0035` still returns `{"category": "unsure"}`. At most one or two messages besides `inq-0035` sit in `unsure`. Accuracy lands between 17 and 19 out of 20. A run at 20/20 means check whether the descriptions now fit only these 20 messages. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves. `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goals

Pick either. Neither is built in `complete/`. The reasoning is in [`expected-output.md`](expected-output.md) under "Stretch Goal".

- **Add a priority field.** Add `priority` to the result type next to `category`, with its own small set of allowed values. Print it in the routing table. Priority is a second axis: it says how fast, and category says where. Mixing the two is how a lost inhaler ends up in line behind a lost wedding ring. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority.
- **Add a confidence threshold.** Add a numeric `confidence` field to the result type and schema. After the loop, change the category to `unsure` on any result whose confidence is under a threshold you pick. Run the scoreboard again. **Check:** both emergencies stay `emergency` with high confidence, `inq-0035` stays `unsure`, and the `unsure` queue does not fill up with ordinary permit questions. If a third of the slice lands in `unsure`, the threshold is too high and you have rebuilt the unsorted inbox.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F07-dotnet.md`](dotnet/F07-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F07-python.md`](python/F07-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F07-typescript.md`](typescript/F07-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder, part of the workshop corpus and shared with feature 09), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix matches the full inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the hand-assigned category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it actually got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus, the source the reference labels were drawn from. Not for handout.
