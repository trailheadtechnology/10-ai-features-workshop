# Lab 07: Classification & Routing

*This is the Recommended lab for [Module 3](../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** sort visitor messages into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and send each one to the right queue.
- **Input:** `data/inquiries-slice.jsonl` (20 messages: `id`, `channel`, `received`, `text`; emergencies `inq-0013` and `inq-0041`, ambiguous `inq-0035`), `data/reference-labels.json` (the correct category for each id, and the queue for each category), `data/inquiries.jsonl` (the full 100-message inbox the 20 came from; the demo scrolls it, the lab does not read it).
- **How:** send each message to `llama3.2` with the taxonomy prompt. Use a JSON schema so the model can only answer with one of the seven labels. Loop all 20 at temperature 0. Print the routing table. Then score the results against the reference labels. `http/ollama.http` holds the request on three messages: 1 `inq-0005` (conditions), 2 `inq-0041` (emergency), 3 `inq-0035` (unsure).
- **Model:** `llama3.2`, local. No key.

Each step below is one thing to make the program do. Every code track's `starter/` classifies one message and prints whatever text the model sends back. Edit it until it does all six steps. Compare against `complete/` when you get stuck. HTTP-track readers run the numbered requests in `http/ollama.http` when a step names one.

### Step 0: Run the starter and read the free-text label

Run `starter/` as it is (`dotnet run`, `uv run main.py`, or `npm run starter`, from the track's folder). It reads `inq-0005` from the slice. It builds the taxonomy prompt and puts the message after `Message:`. It sends that to `llama3.2` as one user message. Then run it again with `inq-0035` as the one argument (`dotnet run -- inq-0035`, `uv run main.py inq-0035`, `npm run starter -- inq-0035`), and again with `inq-0013`.

HTTP-track readers: `http/ollama.http` is three copies of the same classify request, one per inlined message, in the shape Ollama's chat API expects: the model name, one user message holding the taxonomy prompt plus the inquiry text, and the category schema from step 2 already attached. Request 1 is this step's call on `inq-0005`, so its answer is already pinned to the seven labels. The free-text wobble in the check below only shows on the code tracks.

**Check:** one label per run, printed after the id. `inq-0013` says `emergency`. `inq-0035` flips between `conditions` and `permit` across runs. At least one run prints something that is not exactly one of the seven category names, such as `Emergency.` or a sentence. Nothing in the starter stops that. The rest of the lab does.

### Step 1: Load the 20 inquiries and the reference labels

1. Open `../../data/inquiries-slice.jsonl`. Every line is one JSON object with four string fields: `id`, `channel`, `received`, `text`. These are 20 of the 100 messages in `data/inquiries.jsonl`, the fictional Trailhead Guides inbox that ships with the workshop corpus: visitor messages that arrived by email, web form, or voicemail transcript, unlabeled on purpose because labeling is this lab. The slice keeps the original ids and fields untouched and was chosen to mirror the inbox's mix, including both emergencies and the one ambiguous message; no script builds it, it ships with the workshop as is. Feature 09 reads six of these same 100 messages, with feature 07's category added.
2. Read the file line by line. Skip blank lines. Parse each line. Keep the results in a list.
3. Open `../../data/reference-labels.json` and parse it. It is one JSON object with three members: `routing`, `labels`, and `notes` (three short paragraphs on why `inq-0013` and `inq-0041` are emergencies and why `inq-0035` is `unsure`). The labels were assigned by hand from the instructor answer key that was written alongside the corpus (`answer-key.md`, which marks the emergencies and the ambiguous messages in all 100); `unsure` was added to the taxonomy, the labels, and the routing table together once the ambiguous message needed a queue. Keep two dictionaries from it: `routing` (category to queue name) and `labels` (id to correct category).

**Check:** the inquiry list has 20 entries. The first `id` is `inq-0001`. `routing` has 7 keys and `labels` has 20.

### Step 2: Pin the answer to the seven labels with structured output

1. Delete the line `Answer with the category name only.` from the starter's prompt. The schema below does that job now.
2. Define the category as an enum type with exactly seven allowed values, in whatever form your language's structured-output support takes (the raw request carries it as a JSON schema `enum`). Wrap it in a result type with one field, `category`.
3. Send the message through your track's chat client with that result type as the structured-output schema. Set temperature to 0.
4. Keep the prompt text the same. This is the prompt, followed by the inquiry text:

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

Whatever your track's client generates from that type is this JSON schema, which is also what the requests in `http/ollama.http` carry verbatim:

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

5. Run the program on `inq-0005`, then `inq-0041`, then `inq-0035` (requests 1, 2, and 3 in `http/ollama.http`). The message after `Message:` in request 2 is:

```text
My dad slipped on scree on the Beehive descent, we're just past the ladder section. His ankle is swollen bad, can't put weight on it. We have water and I have 2 bars of signal. 2 adults 1 teen. How do we get help up here? Submitting this because 911 kept dropping.
```

And in request 3:

```text
Hi, I have a backcountry permit that includes a night at the Avalanche Lake area on June 24 (conf #GL-2026-07733). With the bridge out, is my itinerary even doable, and if not, will you let me swap that night for a different site without penalty, or refund it? I need to know before we leave Thursday. Thanks, Priya
```

**Check:** `inq-0005` returns `{"category": "conditions"}`, `inq-0041` returns `{"category": "emergency"}`, and `inq-0035` returns `{"category": "unsure"}`. The value is always one of the seven strings. Notice that `inq-0035` now lands in `unsure`. With free text in step 0, the same prompt called it `conditions`. The enum made `unsure` a real choice for the model.

### Step 3: Classify all 20 in a loop

1. Replace the starter's single-id lookup with a loop over the list from step 1.
2. For each inquiry, build the prompt from step 2 with that inquiry's `text` after `Message:`. Send it as one user message to `llama3.2` with the same structured-output schema and temperature 0.
3. Read `category` from the parsed response. Store the pair (inquiry, category) in a results list.
4. Print a `.` after each call so you can see progress. Print a blank line after the loop.

HTTP-track readers: copy request 1 once per message in the slice and swap the text after `Message:`. Or loop the body in your language.

**Check:** 20 dots, then 20 stored pairs, every category one of the seven strings. The run takes under a minute.

### Step 4: Print emergencies first, then the routing table

1. Filter the results down to the ones whose category is `emergency`.
2. If there are any, print the line `!!! EMERGENCY: route to dispatch, page the duty ranger now !!!`. Then print one line per emergency with `!!! `, the id, and the first 70 characters of the text. Print a blank line after the block.
3. Print a header line with `id`, `category`, and `routed to`. Then print a rule of 62 dashes.
4. Sort the results so emergencies come first. For each pair, print the id padded to 10 characters, the category padded to 15, and the queue from `routing` for that category.

**Check:** two lines in the emergency block, `inq-0013` and `inq-0041`, both before the table. The table has 20 rows. `inq-0035` reads `unsure` and routes to `human-review-queue (a ranger reads it and picks the queue)`.

### Step 5: Score accuracy, then score emergency recall separately

1. Count the results whose category equals `labels[id]`. That is the accuracy numerator.
2. Collect the ids in `labels` whose value is `emergency`. Count how many of them appear in the emergency list from step 4. That is emergency recall.
3. Print `Accuracy vs reference labels: N/20` and `Emergency recall: N/2`.
4. For each result whose category does not match `labels[id]`, print `miss: ` plus the id, the category the model gave, and the reference category.

Accuracy is the headline number. Emergency recall is the number that decides whether this is safe to ship. A missed emergency is a person waiting in a queue nobody is watching.

**Check:** a raw HTTP run recorded 18/20 with two misses: `inq-0030` (wedding photographer) got `general`, reference `permit`; `inq-0051` (Sperry campfires) got `conditions`, reference `general`. The code tracks recorded 17/20 with one more: `inq-0001` got `unsure`, reference `permit`. Either way, emergency recall is 2/2, no routine message is `emergency`, and `inq-0035` is `unsure`. An emergency anywhere but `emergency` fails, whatever the accuracy.

### Step 6: Fix the misses by editing the category descriptions, not the code

The category descriptions in the prompt are where accuracy lives. When the model files something wrong, you fix the description first.

1. In the prompt, extend the `permit` description so it also covers questions about whether an activity needs a permit at all. The walkthroughs use this line:

```text
- permit: reserving, changing, canceling, or paying for a permit, pass, or reservation, and questions about whether an activity requires a permit at all, including billing problems and missing confirmations.
```

2. Narrow the `conditions` description so it means only whether a trail, road, or area is physically passable.
3. Let the `general` description own park rules and regulations.
4. Leave the `Decide in this order` paragraph alone. Leave the `unsure` description narrow.
5. Run all 20 again and read the scoreboard.

**Check:** `inq-0030` moves to `permit` and `inq-0008` (Half Dome lottery) stays there. `inq-0051` moves to `general`. `inq-0041` and `inq-0013` still return `{"category": "emergency"}`, and `inq-0035` still returns `{"category": "unsure"}`. At most one or two messages besides `inq-0035` sit in `unsure`. Accuracy lands between 17 and 19 out of 20. A run at 20/20 means check whether the descriptions now fit only these 20 messages. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves. `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goals

Pick either. Neither is built in `complete/`. The reasoning is in [`expected-output.md`](expected-output.md) under "Stretch Goal".

- **Add a priority field.** Add `priority` to the result type next to `category`, with its own small set of allowed values. In `http/ollama.http`, add it to the schema's `properties` and `required`. Print it in the routing table. Priority is a second axis: it says how fast, and category says where. Mixing the two is how a lost inhaler ends up in line behind a lost wedding ring. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority.
- **Add a confidence threshold.** Add a numeric `confidence` field to the result type and the schema. After the loop, change the category to `unsure` on any result whose confidence is under a threshold you pick. Run the scoreboard again. **Check:** both emergencies stay `emergency` with high confidence, `inq-0035` stays `unsure`, and the `unsure` queue does not fill up with ordinary permit questions. If a third of the slice lands in `unsure`, the threshold is too high and you have rebuilt the unsorted inbox.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F07-http.md`](http/F07-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F07-dotnet.md`](dotnet/F07-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F07-python.md`](python/F07-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F07-typescript.md`](typescript/F07-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder, part of the workshop corpus and shared with feature 09), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix matches the full inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the hand-assigned category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it actually got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus, the source the reference labels were drawn from. Not for handout.
