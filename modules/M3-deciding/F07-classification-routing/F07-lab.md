# Lab 07: Classification & Routing

*This is the Recommended lab for [Module 3](../M3-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** classify visitor inquiries into `permit | conditions | complaint | lost-and-found | emergency | general | unsure` and route them.
- **Input:** `data/` provides 20 inquiries from the full inbox in `data/inquiries.jsonl` (including 2 emergencies and at least one deliberately ambiguous message) plus reference labels.
- **How:** POST to Ollama's chat endpoint (`llama3.2`). `http/ollama.http` has the request with a starter taxonomy prompt.
- **Steps:**
  1. Classify all 20 with the starter taxonomy and score against the reference labels.
  2. Find your misclassifications and fix them by rewording the category descriptions.
  3. Success check: both emergencies classified as `emergency`, and the ambiguous message in `unsure` rather than confidently wrong (see `expected-output.md`). Missing an emergency fails the lab even at 19/20 accuracy; that's the lesson.
- **Stretch goal:** add a `priority` field alongside the category, or return a confidence score and route anything below a threshold to `unsure`.

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
