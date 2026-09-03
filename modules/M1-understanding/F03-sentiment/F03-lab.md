# Lab 03: Sentiment

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20.
- **How:** POST to Ollama's chat endpoint, `http/ollama.http` (requests 1 to 5), and Azure OpenAI's, `http/azure.http` (requests 1 to 3). Same prompt bytes everywhere, temperature 0.
- **Model:** `phi3` for `ollama.http` requests 1 to 4, `llama3.2` for request 5, `gpt-4.1` for `azure.http` (paste the room key over `<KEY FROM INSTRUCTOR>`); .NET `complete/` uses Azure when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are set, else `llama3.2`.

### Step 1: The easy set on phi3

Requests 1 (`gr-0003`) and 2 (`gr-0002`) in `http/ollama.http`, or `dotnet run -- --easy` from `dotnet/complete/`. Request 1 in full; keep the line breaks:

```text
Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: Easy recommendation. No blisters in 200 straight miles. The heel cushion alone would be worth half the price.
```

Request 2's review, `gr-0002`:

```text
Review: Returned it. The seam taping peeled after one season. For this price that's inexcusable.
```

Swap in the other eight reviews from `data/easy.jsonl` and score against `data/reference-labels.json`.

**Check:** `gr-0003` comes back `positive`, `gr-0002` `negative`, one word each; recorded `phi3` score 9/10, missing `gr-0074`. A 7/10 with every miss `mixed` means you reflowed the prompt onto one line.

### Step 2: The hard set on both models

Requests 3 and 4 (`phi3`) and 5 (`llama3.2`) in `http/ollama.http`, then all three in `http/azure.http`; or `dotnet run -- --hard` from `dotnet/complete/`. Request 3's review, `gr-0034`:

```text
Review: Five stars for the return process, which I got to experience because the frame stay poked through the back panel on day one. The Cascade 65 Backpack itself is the worst piece of gear I have owned in thirty years outdoors.
```

Requests 4 and 5, and `azure.http` request 1, `gr-0013`:

```text
Review: Five stars for the return process, which I got to experience because the battery meter jumps from 40 percent to dead on day one. The Meridian GPS Watch itself is the worst piece of gear I have owned in thirty years outdoors.
```

`azure.http` requests 2 (`gr-0004`) and 3 (`gr-0074`):

```text
Review: Two stars for the instruction booklet, which is a crime against paper. Once I figured it out on my own, wow — it carries 40 pounds like it's 25. The Cascade 65 Backpack deserves five, the manual deserves jail.
```

```text
Review: It pitches in three minutes flat. That's it, that's the review.
```

Extend to the rest of `data/hard.jsonl` on both models.

**Check:** `gr-0034` comes back `negative`; on `gr-0013`, `phi3` says `mixed` and the other two say `negative`, the reference label. Recorded hard-set scores: `phi3` 7/10, `llama3.2` 8/10, `gpt-4.1` 10/10.

### Step 3: The disagreement list

No prompt: list every review where the two models differ, with the reference label and which model was right (`dotnet run` from `dotnet/complete/` prints this as `disagreements`).

**Check:** your version of the two tables in `expected-output.md`: recorded 4 of 20 against `gpt-4.1` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0021`) and 4 of 20 against `llama3.2` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0089`). Different numbers are fine; no tables is not.

### Stretch goal: aspect-based sentiment

Change the prompt to ask for this shape instead of one word, and add a `format` schema for it:

```text
{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}
```

**Check:** parseable JSON every time, with an aspect left `null` when the review never mentions it. A `price` sentiment for a review that never mentions price is the failure to look for.

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
