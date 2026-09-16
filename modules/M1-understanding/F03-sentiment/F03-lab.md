# Lab 03: Sentiment

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** classify gear reviews as `positive | negative | mixed` with two models, score both, and list where they disagree.
- **Input:** `data/easy.jsonl`, 10 reviews where text and stars agree; `data/hard.jsonl`, 10 where they fight; `data/reference-labels.json`, hand labels for all 20. All three are hand-picked from feature 06's `data/gear-reviews.jsonl`; no script builds them.
- **How:** send one prompt per review through your track's chat client. Keep the one-word label that comes back. Compare it with the hand label. Same prompt bytes everywhere, temperature 0.
- **Model:** `phi3` is the small model. The big model is `gpt-4.1` on Azure. Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` to use it. With no key, `llama3.2` on Ollama fills in for the big model, and the whole lab runs offline.

Each step below adds one thing to the program. Every track's `starter/` already classifies one review on `phi3`. Edit it until it does all six steps. Look at `complete/` when you get stuck.

### Step 0: Run the starter on the sarcastic review

**Do:** run `starter/` as it is. It loads `gr-0007` from `data/hard.jsonl`, prints the product, star rating, reviewer, and text, then sends this prompt to `phi3` and prints the label:

```text
Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: Absolutely love it when the mesh blew out at the pinky toe inside two weeks the second day of a trip. Five-star experience, truly, if the stars are measuring my personal growth through adversity. Rating it what it deserves.
```

**Why:** the four line breaks matter; every later step sends this same prompt and only changes the text after `Review: `.

**Check:** `phi3 says: negative`. The review has two stars and says "five-star experience, truly" — that's sarcasm, and the small model gets it right. Pass a different id (`gr-0034`, for instance) to classify a different review.

### Step 1: Load the reviews and the reference labels

**Do:**
1. Open `../../data/easy.jsonl`: 10 reviews where text and stars agree, one JSON object per line (`id`, `product`, `rating`, `reviewer`, `text`).
2. Read it one line at a time, skip blanks, parse each line into a list.
3. Open `../../data/reference-labels.json`: one object keyed by review `id`, each value a `set` (`easy`/`hard`), a `label`, and on hard cases a `rationale`.
4. Parse it into a dictionary keyed by `id`. You only need `label` in code.

**Why:** `gr-0004`'s label changed from `positive` to `mixed` after `gpt-4.1` called it `mixed` on every soak-test run and a reread agreed a two-star review praising the product inside it is genuinely split — `expected-output.md` tells that story.

**Check:** the easy list has 10 reviews, first `id` is `gr-0002`. The label dictionary has 20 keys; `gr-0002` maps to `negative`.

### Step 2: Classify the easy set on `phi3` and score it

**Do:**
1. Keep the starter's classify function as-is: one review text in, one prompt out, `phi3`, temperature 0.
2. Lowercase the reply, look for `positive`/`negative`/`mixed`, keep whichever shows up first (or the trimmed reply if none do).
3. Loop over the easy list, classify each review's `text`, look up the reference label by `id`.
4. Print one row per review: `id`, reference label, `phi3` label.
5. Count matches; print `phi3 N/10`.

**Why:** step 2's lowercase-and-search handles small models wrapping the label in a sentence.

**Check:** `gr-0003` comes back `positive`, `gr-0002` `negative`. Recorded `phi3` score 9/10, missing `gr-0074`. A 7/10 with every miss `mixed` means you reflowed the prompt onto one line.

### Step 3: Add the big model and classify every review twice

**Do:**
1. Read `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT`.
2. Build a second chat client from those three variables, falling back to `llama3.2` when any is missing; name it `azure:<deployment>` or `llama3.2` for printing.
3. When falling back, print `AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.`
4. Change nothing in classify — the second model is just a different client.
5. In the step 2 loop, classify each review with both `phi3` and the big model; store a record with the review, set name, reference label, small label, big label.
6. Add a fourth column for the big model's label; append `  <- disagree` when the two differ.

**Why:** the provider swap is a few lines around one client constructor, nothing downstream changes — that's the point of this step.

**Check:** the easy table has four columns, and `gr-0074` is flagged `<- disagree`: reference `positive`, `phi3` says `mixed`, `gpt-4.1` says `positive`. On the `llama3.2` stand-in the row is still flagged, big label `negative`.

### Step 4: Run the hard set too and print accuracy per set

**Do:**
1. Put the step 3 loop inside an outer loop over `easy` then `hard`, reading `../../data/{set}.jsonl` each time, with a heading line per set.
2. After both tables, loop over the two set names again; count matches for small and big labels against the reference.
3. Print one line per set: name, `phi3 N/10`, big model's name with `N/10`.

**Check:** `gr-0034` comes back `negative` from both models. On `gr-0013`, `phi3` says `mixed`, big model says `negative` (the reference label). Recorded: `phi3` 9/10 easy, 7/10 hard; `gpt-4.1` 10/10 both; `llama3.2` stand-in 9/10 and 8/10.

### Step 5: Print the disagreement list

**Do:**
1. Take only records where small and big labels differ.
2. Print a heading: `disagreements (N of 20)`.
3. For each: verdict is the big model's name + ` right` if it matches the reference, `phi3 right` if the small label matches instead, else `both wrong`.
4. Print one line per disagreement: `id`, set in brackets, `ref=`, `phi3=`, big model `=`, verdict in parens; then the review text in quotes, cut to 100 characters with `...` if longer.
5. If empty, print `(none this run)`.

**Check:** your version of the two tables in `expected-output.md`: recorded 4 of 20 against `gpt-4.1` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0021`) and 4 of 20 against `llama3.2` (`gr-0074`, `gr-0004`, `gr-0013`, `gr-0089`). Different numbers are fine; no tables is not.

### Stretch goals

Pick any. The first two are already built in `complete/`.

- **Add `--easy` and `--hard` flags.** If `--easy` is present, run only the easy set; if `--hard`, only the hard set; otherwise both. **Check:** `--hard` prints one table, one accuracy line, and a disagreement count out of 10 instead of 20.
- **Reflow the prompt onto one line and measure the damage.** Replace the four line breaks with spaces, rerun both sets on both models, then put the line breaks back. **Check:** recorded `phi3` drops from 9/10 to 7/10 easy and 7/10 to 4/10 hard, every miss `mixed`. `llama3.2` scores the same either way — the small model is the one that cares about prompt shape.
- **Aspect-based sentiment.** Ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` instead of one word; add a `format` schema for it. **Check:** parseable JSON every time, aspects left `null` when the review never mentions them. A `price` sentiment on a review that never mentions price is the failure to look for.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
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

The reviews keep their original ids, so any of them can be traced back to the full corpus.
