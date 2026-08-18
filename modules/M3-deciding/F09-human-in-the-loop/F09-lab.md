# Lab 09: Human-in-the-Loop

*A Challenge lab. Do it if you finished [Module 3](../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** generate draft replies for routed inquiries, then decide the automation policy per category.
- **Input:** `data/` provides 6 classified inquiries (boilerplate, two permit questions, a conditions question about the washout, a billing complaint, and one emergency), relevant park-doc snippets for grounding, and a policy worksheet.
- **How:** POST to Ollama's chat endpoint (`llama3.2`) with the grounding snippets. `http/ollama.http` has the drafting request.
- **Steps:**
  1. Generate a draft reply for each inquiry and read them as an editor: what would you change before this goes out under your name? Watch what the model does with the emergency in particular.
  2. Fill in the policy worksheet: for each of the five categories from feature 07, choose auto-send, draft-for-approval, or human-only, and write one sentence of justification based on error cost.
  3. Success check: compare against `expected-output.md`, which has reference drafts and a reasoned reference policy. Your policy may differ; your justifications are what count. This lab is deliberately part judgment, because that's the actual skill.
- **Stretch goal:** compute edit distance between a draft and your edited version, and sketch what threshold would earn a category promotion from draft-mode to auto-send.

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
