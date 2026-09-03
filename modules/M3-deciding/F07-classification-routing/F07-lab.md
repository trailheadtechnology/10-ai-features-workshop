# Lab 07: Classification & Routing

*This is the Recommended lab for [Module 3](../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** classify visitor inquiries into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and route them.
- **Input:** `data/inquiries-slice.jsonl` (20 messages: `id`, `channel`, `received`, `text`; emergencies `inq-0013` and `inq-0041`, ambiguous `inq-0035`), `data/reference-labels.json` (reference category per id and each category's queue), `data/inquiries.jsonl` (the full 100-message inbox the 20 were drawn from; the demo scrolls it, the lab does not read it).
- **How:** Ollama's chat endpoint with a `format` JSON schema pinning the category to the seven labels. `http/ollama.http`: 1 `inq-0005` (conditions), 2 `inq-0041` (emergency), 3 `inq-0035` (unsure); for the other 17, copy request 1 and swap the text after `Message:`. .NET `complete/` scores all 20: `dotnet run`, no flags, temperature 0.
- **Model:** `llama3.2`, local. No key.

### Step 1: Classify all 20 in data/inquiries-slice.jsonl with the starter taxonomy

Send requests 1, 2, and 3, then the other 17 through a copy of request 1, and score every label against `data/reference-labels.json` (or `dotnet run` in `complete/`); the prompt, followed by the inquiry text:

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

The `format` field on every request:

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

**Check:** about 18 of 20 right, `inq-0013` and `inq-0041` in `emergency`, `inq-0035` in `unsure`; expect `inq-0030` (wedding photographer) in `general` and `inq-0051` (Sperry campfires) in `conditions`. Either emergency anywhere but `emergency` fails, whatever the accuracy.

### Step 2: Fix inq-0030 and inq-0051 by rewording the category descriptions

Edit the description lines only, then re-run all 20: extend `permit` to permit availability, eligibility, and rules; narrow `conditions` to physical passability; let `general` own rules.

**Check:** `inq-0030` moves to `permit` and `inq-0008` (Half Dome lottery) stays there, `inq-0051` moves to `general`, and at most one or two messages besides `inq-0035` sit in `unsure`. `inq-0013` or `inq-0041` leaving `emergency` fails, even when accuracy improves.

### Step 3: Re-check inq-0041, inq-0013, and inq-0035 with the edited taxonomy

Send request 2 (`inq-0041`) and request 3 (`inq-0035`) again, and the same prompt on `inq-0013`; the messages after `Message:` in requests 2 and 3:

```text
My dad slipped on scree on the Beehive descent, we're just past the ladder section. His ankle is swollen bad, can't put weight on it. We have water and I have 2 bars of signal. 2 adults 1 teen. How do we get help up here? Submitting this because 911 kept dropping.
```

```text
Hi, I have a backcountry permit that includes a night at the Avalanche Lake area on June 24 (conf #GL-2026-07733). With the bridge out, is my itinerary even doable, and if not, will you let me swap that night for a different site without penalty, or refund it? I need to know before we leave Thursday. Thanks, Priya
```

**Check:** `inq-0041` and `inq-0013` return `{"category": "emergency"}`, `inq-0035` returns `{"category": "unsure"}`. Missing either emergency fails the lab even at 19/20; `inq-0035` confidently in `conditions` or `permit` fails too.

### Stretch goal: a priority field for inq-0006, or a confidence threshold

Add `priority` to the schema's `properties` and `required` next to `category`, or add a confidence score and route anything under a threshold to `unsure`. **Check:** `inq-0006` (lost daypack with a child's inhaler) stays `lost-and-found` with a high priority; folding priority into category puts the inhaler behind the wedding ring.

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

- `data/inquiries-slice.jsonl`: 20 messages pulled from `inquiries.jsonl` (the full 100-message inbox, also in this folder), one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix is representative of the real inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `data/reference-labels.json`: the category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it actually got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus. Not for handout.
