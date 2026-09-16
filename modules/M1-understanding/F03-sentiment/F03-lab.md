# Lab 03: Sentiment

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20.
- **How:** send one prompt per review through your track's chat client. Keep the one-word label that comes back. Compare it with the hand label. Same prompt bytes everywhere, temperature 0. On the HTTP track the same calls are requests 1 to 5 in `http/ollama.http` and requests 1 to 3 in `http/azure.http`.
- **Model:** `phi3` is the small model. The big model is `gpt-4.1` on Azure. To use it, paste the room key over `<KEY FROM INSTRUCTOR>` in `azure.http`, or set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` in a code track. With no key, `llama3.2` on Ollama fills in for the big model, and the whole lab runs offline.

Each step below adds one thing to the program. Each code track's `starter/` already classifies one review on `phi3`. Edit it until it does all six steps. Look at `complete/` when you get stuck. If you are on the HTTP track, run the numbered requests in `http/ollama.http` and `http/azure.http` where a step names one. Both files are hand-written request files that ship with the workshop; nothing generates them. `ollama.http` holds five requests to the local `phi3` model: two easy-set reviews, two hard-set reviews, and the second hard one rerun on `llama3.2` so you can see a disagreement offline. `azure.http` holds the same prompt in three requests to the workshop's `gpt-4.1` deployment, endpoint already filled in and only the key missing. Every request inlines one review's text after `Review: ` so it runs as-is.

### Step 0: Run the starter on the sarcastic review

This step reads `data/hard.jsonl`: 10 reviews, one JSON object per line with `id`, `product`, `rating`, `reviewer`, and `text`, picked by hand out of the workshop's full review corpus (feature 06's `data/gear-reviews.jsonl`, where all 20 ids in this lab appear with the same text) because the text and the star rating pull in opposite directions. There is no script behind the pick; the file ships with the workshop.

Run `starter/` as it is (`dotnet run`, `uv run main.py`, or `npm run starter`, depending on your track). It loads `gr-0007` from `data/hard.jsonl`. It prints the product, the star rating, the reviewer, and the text. Then it sends this prompt to `phi3` and prints the label:

```text
Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: Absolutely love it when the mesh blew out at the pinky toe inside two weeks the second day of a trip. Five-star experience, truly, if the stars are measuring my personal growth through adversity. Rating it what it deserves.
```

The four line breaks in the prompt matter. Do not join the lines. Every later step sends this same prompt and only changes the review text after `Review: `.

**Check:** `phi3 says: negative`. The review has two stars and says "five-star experience, truly". That is sarcasm, and the small model gets it right. Pass a different id as the argument (`gr-0034`, for instance) and it classifies that review instead.

### Step 1: Load the reviews and the reference labels

1. Open `../../data/easy.jsonl`. It is the other half of the sample: 10 reviews picked by hand from the same corpus because the text and the star rating agree. Every line is one JSON object with the same five fields as `hard.jsonl`: `id`, `product`, `rating`, `reviewer`, `text`.
2. Read the file one line at a time. Skip blank lines. Parse each line and add the result to a list.
3. Open `../../data/reference-labels.json`. It is one JSON object keyed by review `id`. Each value has a `set` (`easy` or `hard`), a `label` (`positive`, `negative`, or `mixed`), and on hard cases a `rationale`.
4. Parse it into a dictionary keyed by `id`. You only need `label` in code. Ignore `rationale`. The labels were written by hand when the sample was picked, before any model ran, and one of them changed later: `gr-0004` started as `positive` and became `mixed` after `gpt-4.1` called it `mixed` on every soak-test run and a reread agreed that a two-star review with a five-star product inside it is genuinely split. `expected-output.md` tells that story. No script builds this file; it ships with the workshop.

**Check:** the easy list has 10 reviews and the first `id` is `gr-0002`. The label dictionary has 20 keys, and `gr-0002` maps to `negative`.

### Step 2: Classify the easy set on `phi3` and score it

1. Keep the starter's classify function as it is. It takes a review text. It puts that text after `Review: ` in the prompt from step 0. It sends the prompt as a single user message through your track's chat client with model `phi3` and temperature 0. Requests 1 (`gr-0003`) and 2 (`gr-0002`) in `http/ollama.http` are that call for two of the easy reviews.
2. Lowercase the reply. Look for `positive`, `negative`, and `mixed` in it. Keep whichever one shows up first. If none of the three shows up, keep the trimmed reply as the label. Small models sometimes wrap the label in a sentence, and this step still gives you a label you can score.
3. Loop over the easy list. For each review, call classify with its `text`. Look up the reference label by `id`.
4. Print one row per review: `id`, reference label, `phi3` label.
5. Count the reviews where the `phi3` label matches the reference label. After the loop, print `phi3 N/10`.

Request 1's review, `gr-0003`:

```text
Review: Easy recommendation. No blisters in 200 straight miles. The heel cushion alone would be worth half the price.
```

Request 2's review, `gr-0002`:

```text
Review: Returned it. The seam taping peeled after one season. For this price that's inexcusable.
```

**Check:** `gr-0003` comes back `positive`, `gr-0002` `negative`, one word each. Recorded `phi3` score 9/10, missing `gr-0074`. A 7/10 with every miss `mixed` means you reflowed the prompt onto one line.

### Step 3: Add the big model and classify every review twice

1. Read the environment variables `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT`.
2. Build a second chat client from those three `AZURE_OPENAI_*` variables, falling back to `llama3.2` when they are not set. When all three are set, name the client `azure:` plus the deployment name for printing.
3. When any one is missing, print `AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.` and name the fallback client `llama3.2`. Request 5 in `http/ollama.http` is that call for `gr-0013`. Requests 1 to 3 in `http/azure.http` are the Azure version for `gr-0013`, `gr-0004`, and `gr-0074`.
4. Change nothing in classify. It already takes a client and a text, so the second model is just a different client.
5. In the loop from step 2, call classify twice per review: once with `phi3` and once with the big model. Store a result record with the review, the set name, the reference label, the small label, and the big label. Keep every record in one list.
6. Add a fourth column to the printed row for the big model's label. When the two labels differ, add `  <- disagree` to the end of the row.

The three `azure.http` reviews, in request order:

```text
Review: Five stars for the return process, which I got to experience because the battery meter jumps from 40 percent to dead on day one. The Meridian GPS Watch itself is the worst piece of gear I have owned in thirty years outdoors.
```

```text
Review: Two stars for the instruction booklet, which is a crime against paper. Once I figured it out on my own, wow — it carries 40 pounds like it's 25. The Cascade 65 Backpack deserves five, the manual deserves jail.
```

```text
Review: It pitches in three minutes flat. That's it, that's the review.
```

**Check:** the easy table has four columns, and `gr-0074` is the one row flagged `<- disagree`: reference `positive`, `phi3` says `mixed`, `gpt-4.1` says `positive`. On the `llama3.2` stand-in the row is still flagged, but the big label is `negative`.

### Step 4: Run the hard set too and print accuracy per set

1. Put the loop from step 3 inside an outer loop over two set names, `easy` then `hard`. Read `../../data/{set}.jsonl` for each one. Print a heading line with the set name above its table. Requests 3 (`gr-0034`) and 4 (`gr-0013`) in `http/ollama.http` are the `phi3` call for two of the hard reviews.
2. After both tables, loop over the two set names again. For each set, take only the records from that set. Count the records where the small label matches the reference label. Count the records where the big label matches the reference label.
3. Print one line per set: the set name, `phi3 N/10`, and the big model's name with its `N/10`.

Request 3's review, `gr-0034`:

```text
Review: Five stars for the return process, which I got to experience because the frame stay poked through the back panel on day one. The Cascade 65 Backpack itself is the worst piece of gear I have owned in thirty years outdoors.
```

**Check:** `gr-0034` comes back `negative` from both models. On `gr-0013`, `phi3` says `mixed` and the big model says `negative`, the reference label. Recorded scores: `phi3` 9/10 easy and 7/10 hard, `gpt-4.1` 10/10 and 10/10, the `llama3.2` stand-in 9/10 and 8/10.

### Step 5: Print the disagreement list

1. Take only the records where the small label and the big label differ.
2. Print a heading with the count: `disagreements (N of 20)`.
3. For each one, pick the verdict. If the big label matches the reference, the verdict is the big model's name plus ` right`. Otherwise, if the small label matches the reference, it is `phi3 right`. Otherwise it is `both wrong`.
4. Print one line per disagreement: `id`, the set in square brackets, `ref=`, `phi3=`, the big model's name with `=`, and the verdict in parentheses.
5. On the next line, print the review text in quotes. If it is longer than 100 characters, cut it to 100 and add `...`.
6. If the list is empty, print `(none this run)`.

**Check:** your version of the two tables in `expected-output.md`: recorded 4 of 20 against `gpt-4.1` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0021`) and 4 of 20 against `llama3.2` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0089`). Different numbers are fine; no tables is not.

### Stretch goals

Pick any. The first two are already built in `complete/`.

- **Add `--easy` and `--hard` flags.** Read the program's arguments. If `--easy` is there, run only the easy set. If `--hard` is there, run only the hard set. Otherwise run both. Nothing else has to change, because every later section already loops over the list of set names. **Check:** `--hard` prints one table, one accuracy line, and a disagreement count out of 10 instead of 20.
- **Reflow the prompt onto one line and measure the damage.** Replace the four line breaks inside the prompt with spaces. Change nothing else. Rerun both sets on both models. Then put the line breaks back. **Check:** recorded `phi3` drops from 9/10 to 7/10 on easy and from 7/10 to 4/10 on hard, with every miss coming back `mixed`. `llama3.2` scores the same either way. The small model is the one that cares about prompt shape. That is why the lab locks the prompt bytes before it changes the model.
- **Aspect-based sentiment.** Change the prompt to ask for this shape instead of one word, and add a `format` schema for it:

  ```text
  {"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}
  ```

  Parse the reply as JSON and print it per review. **Check:** parseable JSON every time, with an aspect left `null` when the review never mentions it. A `price` sentiment for a review that never mentions price is the failure to look for.

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

- `data/easy.jsonl`: 10 straightforward reviews chosen by hand from the full review set (feature 06's `data/gear-reviews.jsonl`). The text says what it means and the star rating agrees.
- `data/hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `data/reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case explaining what the rating is really about. One label (`gr-0004`) was revised after the soak test; `expected-output.md` records why.
- `http/ollama.http` and `http/azure.http`: the same classify prompt as hand-written requests, five against local models and three against `gpt-4.1`, for the HTTP track.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.

The reviews keep their original ids, so any of them can be traced back to the full corpus.
