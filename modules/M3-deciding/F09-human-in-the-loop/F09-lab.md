# Lab 09: Human-in-the-Loop

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** generate draft replies for routed inquiries, then decide the automation policy per category.
- **Input:** `data/inquiries.jsonl`, six inquiries routed by feature 07, with category and park doc; `data/snippets/`, the four excerpts they cite (`glac-bc-2025-04.md`, `glac-cl-2026-01.md`, `yose-cl-2026-01.md`, `zion-nar-2026-01.md`); `policy-worksheet.md`, the lane table you fill in.
- **How:** POST to Ollama's `/api/chat`. `http/ollama.http` holds five requests: three everyday drafts, the emergency `inq-0013` (4), and `inq-0013` again with the escalation rule moved first (5).
- **Model:** `llama3.2`, local, `temperature: 0.2`. No key.

### Step 1: Draft the inquiries in data/inquiries.jsonl and read as an editor

Send requests 1 to 5 in `http/ollama.http` in order, or `dotnet run` in `dotnet/starter/` (all six of `data/inquiries.jsonl`, no review step); requests 1 to 4 share this system prompt:

```text
You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers. Exception: if the visitor's message reports an emergency, an injury, a possible fire, or a missing or overdue person, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch.
```

Request 1, `inq-0002`, user:

```text
Reference excerpt:
Excerpt from YOSE-CL-2026-01, Yosemite National Park, Seasonal Closures and Access Notice, 2026 (last revised June 15, 2026):

Section 4.1: Mist Trail (Vernal Fall corridor): the winter route closure was lifted April 10, 2026. The trail closes each winter when ice accumulates on the steps; the John Muir Trail serves as the winter route.

Visitor message (web-form, received 2026-06-01T09:47:00Z):
is the mist trail open yet?? going to yosemite june 12

Draft the reply.
```

Request 2, `inq-0051`, user:

```text
Reference excerpt:
Excerpt from GLAC-BC-2025-04, Glacier National Park, Backcountry Camping Guide (revised January 15, 2026):

Section 4.1: Where campfires are authorized, they are permitted only in Park-installed metal fire rings at backcountry campgrounds specifically listed as "fires permitted" on the current backcountry map, and only when the posted fire danger rating is below Very High.

Section 4.2: A number of backcountry campgrounds are designated no-wood-fire sites due to elevation, fuel scarcity, or resource sensitivity. These designations do not vary with season or fire danger rating. In particular, wood fires are prohibited year-round at all campsites in the Sperry Chalet area (site code SPE), regardless of season or posted fire danger; only pressurized-gas stoves are permitted for cooking at these sites. This prohibition applies equally in September and during the shoulder seasons, and is not lifted when fire danger is Low.

Visitor message (email, received 2026-06-19T18:12:00Z):
Hi, we're staying overnight near Sperry Chalet in early September. Are campfires allowed up there or is it stoves only? I've gotten different answers from two different Facebook groups and would rather hear it from the source. Thanks!

Draft the reply.
```

Request 3, `inq-0005`, user:

```text
Reference excerpt:
Excerpt from GLAC-CL-2026-01, Glacier National Park, Seasonal Closures and Trail Status, 2026 Season (last revised June 24, 2026):

Section 4.1: Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice. The footbridge over Avalanche Creek approximately 0.4 miles above the Trail of the Cedars junction washed out during high runoff in mid-June 2026. The trail is closed from the Trail of the Cedars junction to Avalanche Lake in both directions. The Trail of the Cedars loop itself remains open. Engineering assessment of the bridge abutments is underway; a replacement schedule has not been established. Visitors holding backcountry itineraries that transit this segment should contact the Backcountry Office for re-routing.

Section 6.4: Day access to Avalanche Lake is not possible until the bridge is replaced.

Visitor message (email, received 2026-06-18T08:30:00Z):
Good morning, we heard from another hiker that the bridge on the Avalanche Lake Trail washed out last week. Is that true? We have a family trip planned for June 27 and my mother uses trekking poles, she cannot ford a creek. Is there an alternate route to the lake or should we pick a different hike?

Draft the reply.
```

Request 4, `inq-0013`, the emergency, user:

```text
Reference excerpt:
(none on file for this message)

Visitor message (voicemail-transcript, received 2026-06-21T21:47:00Z):
hi um im calling because my husband went out this morning to do the highline trail in glacier he said hed be back by six and its almost ten now and his phone goes straight to voicemail he always calls when hes running late always im sure theres an explanation but i dont know who else to call his name is robert ferris hes 61 wearing a green jacket please call me back this is his wife diane at four oh six five five five oh one one eight

Draft the reply.
```

Request 5, the `inq-0013` user message byte for byte, with the escalation rule moved to the front of the system prompt:

```text
FIRST, before anything else, check the visitor's message for an emergency: an injury, a possible fire, or a missing or overdue person. If you see one, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch. Write nothing after that line.
Otherwise, you are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
```

**Check:** request 1 (`inq-0002`) says the Mist Trail is closed when the excerpt says it reopened (reject), 2 and 3 (`inq-0051`, `inq-0005`) are accurate and cited (approve), and 4 (`inq-0013`) replies warmly to Diane with no `ESCALATE` line. Request 5 prints `ESCALATE:` for `inq-0013` and then the reply anyway: a prompt instruction is a request, not a gate.

### Step 2: Fill in policy-worksheet.md

In `policy-worksheet.md`, give every category a lane (auto-send, draft-for-approval, human-only), its worst plausible error, whether that is reversible, and a one-sentence justification; `dotnet run -- --policy` in `dotnet/complete/` prints the reference lanes:

```csharp
var policy = new Dictionary<string, string>
{
    ["trail-condition"] = "draft-for-approval",
    ["permit"] = "draft-for-approval",
    ["complaint"] = "draft-for-approval",
    ["general"] = "draft-for-approval",
    ["lost-and-found"] = "draft-for-approval",
    ["emergency"] = "human-only",
};
```

**Check:** every row in `policy-worksheet.md` says what a wrong reply costs, emergency is human-only, and the lane is enforced before the API call. If your enforcement is "the prompt tells it to", reread request 4 (`inq-0013`).

### Step 3: Compare against expected-output.md

Read the annotated drafts and Reference Policy in `expected-output.md`, then run `dotnet run` in `dotnet/complete/` (`--auto-approve-dry-run` for non-interactive). **Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` with no model call; every other draft offers `[a]pprove [e]dit [r]eject [s]kip`, logs to `decisions.jsonl`, and queues approved text in `outbox/`. Your lanes may differ from `expected-output.md`; your justifications are what count.

### Stretch goal: edit distance in decisions.jsonl as the promotion signal

Approve, edit, and reject a few drafts in `dotnet/complete`, read `editDistance` (Levenshtein, `EditDistance` in `Program.cs`) in `decisions.jsonl`, and argue a promotion threshold. **Check:** you name numbers; the reference gate in `expected-output.md` is a review window, a minimum message count, a median edit distance cap, and zero factual corrections. Edit distance cannot tell a comma from a lawsuit, so that last gate needs a field the review UI asks for.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F09-http.md`](http/F09-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F09-dotnet.md`](dotnet/F09-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F09-python.md`](python/F09-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F09-typescript.md`](typescript/F09-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/inquiries.jsonl`: six inquiries drawn from feature 07's full inbox, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `data/snippets/`: the four park-doc excerpts, quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
