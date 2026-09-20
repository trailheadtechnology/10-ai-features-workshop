<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 03: Sentiment

**The User Problem:** Trailhead Guides sells gear, and the Cascade 65 backpack has 300 reviews. The product team asks a simple question: are people happy with it, and what are they mad about? Star ratings lie: "4 stars, but the hip belt broke on day two" is not a happy customer. Someone would have to read all 300, and every new product adds to the pile. The team doesn't need eloquent analysis, just a reliable happy/unhappy/mixed signal at scale.

The user in this feature is the product team, not the hiker, and that's deliberate. Stakeholders are users too. Their version of the problem is that nobody has time to read every review, so a defect surfaces only when returns spike. Track the same happy/unhappy signal weekly and it surfaces months earlier, as a chart sliding downhill. One product in this corpus has exactly that kind of problem buried in its reviews, and the demo gets to find it.

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20. All three are hand-picked from feature 06's `data/gear-reviews.jsonl`; no script builds them.
- **How:** send one prompt per review through your track's chat client. Keep the one-word label that comes back. Compare it with the hand label. Same prompt bytes everywhere, temperature 0.
- **Model:** `phi3` is the small model. The big model is `gpt-4.1` on Azure. Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` to use it. With no key, `llama3.2` on Ollama fills in for the big model, and the whole lab runs offline.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each track's lab has the full steps plus the exact calls, imports, and run commands for that language.

| Track | Open this lab | What you edit |
|---|---|---|
| .NET | [`dotnet/F03-dotnet.md`](dotnet/F03-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F03-python.md`](python/F03-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F03-typescript.md`](typescript/F03-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/easy.jsonl`: 10 straightforward reviews chosen by hand from the full review set (feature 06's `data/gear-reviews.jsonl`). The text says what it means and the star rating agrees.
- `data/hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `data/reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case. One label (`gr-0004`) was revised after the soak test; `expected-output.md` records why.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key

The reviews keep their original ids, so any of them can be traced back to the full corpus.
