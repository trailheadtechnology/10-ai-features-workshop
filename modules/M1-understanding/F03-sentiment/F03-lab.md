# Lab 03: Sentiment

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, and find where they disagree.
- **Input:** `data/` provides about 20 reviews from the ~300 in feature 06's `data/gear-reviews.jsonl`, split into an easy set and a hard set (sarcasm, contradictions), plus reference labels.
- **How:** `http/ollama.http` (phi3) and `http/azure.http` (Azure OpenAI, key handed out in the room). Same prompt, two endpoints.
- **Steps:**
  1. Classify the easy set with `phi3` and score against the reference labels.
  2. Classify the hard set with both models.
  3. Success check: produce the disagreement list. Which reviews got different labels, and which model was right? (See `expected-output.md`.)
- **Stretch goal:** extend the label to aspect-based sentiment, `{overall, aspects: {comfort, durability, price}}`, and see which model can go deeper than a single label.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F03-http.md`](http/F03-http.md) | the requests in `http/ollama.http` and `http/azure.http`, or a port of them in your language |
| .NET | [`dotnet/F03-dotnet.md`](dotnet/F03-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F03-python.md`](python/F03-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F03-typescript.md`](typescript/F03-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/easy.jsonl`: 10 straightforward reviews from the full review set (feature 06's `data/gear-reviews.jsonl`). The text says what it means and the star rating agrees.
- `data/hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `data/reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case explaining what the rating is really about.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.

The reviews keep their original ids, so any of them can be traced back to the full corpus.
