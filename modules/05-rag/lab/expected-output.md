# What passing looks like

Wording varies run to run; the checks below are what has to be true. Every sample
here came from an actual run against local Ollama: `nomic-embed-text` for the 241
chunks in `chunks.jsonl`, `llama3.2` for generation. All scores are real.

Retrieval in the finished demo is **hybrid**: normalized cosine similarity blended
with a BM25-lite lexical score over the chunk text, weighted by `--alpha` (default
0.6). `--alpha 1.0` turns the lexical half off and gives you the pure-cosine
numbers this file used to report. Both sets are below, because the difference is
the lesson.

## Step 0: the Sperry question with no context (the failure you're supposed to get)

`llama3.2`, asked the question cold:

> Sperry Chalet is a popular destination in the San Gabriel Mountains of California,
> and it's known for its scenic views and outdoor recreational activities. However,
> I couldn't find any information on whether campfires are allowed at Sperry Chalet.
>
> According to the US Forest Service website, which manages the San Gabriel National
> Forest where Sperry Chalet is located, campfires are generally prohibited in the
> forest during periods of high fire danger [...]

A second run relocated it to the Los Angeles National Forest and suggested contacting
"the Sperry Chalet management team." Sperry Chalet is in Glacier National Park, Montana.
The model is wrong about the state, wrong about the managing agency, and fluent about
both. It hedges on the fire rule, which reads as caution but is really just the same
guesswork wearing a different hat: the actual rule is absolute and the model never
finds it. Note that the answer text changes every run while the confidence does not.

## Step 1a: pure cosine retrieval, top 3 (`--alpha 1.0`)

This is module 04's search, verbatim. Run it first so the room sees the problem.

**Q1. "Can I have a campfire at Sperry Chalet in September?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-backcountry-camping-guide:04` | 0.6933 |
| 2 | `acadia-campfire-and-campground-regulations:07` | 0.6929 |
| 3 | `yosemite-campfire-regulations:03` | 0.6851 |

The right chunk ranks first by **0.0004** of raw cosine. Ranks 2 and 3 are campfire
rules from the wrong parks. Five of the 25 documents are about campfires, and the
embedder is matching "campfire regulations," not "Sperry." That margin is a coin
flip, and Step 1c shows a rephrasing that loses it.

**Q2. "What is the maximum group size on a Glacier backcountry permit?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-backcountry-permit-regulations:04` | 0.8457 |
| 2 | `glacier-backcountry-permit-regulations:05` | 0.7802 |
| 3 | `glacier-backcountry-permit-regulations:01` | 0.7769 |

Clean. 0.85 with a wide gap to second place, all three hits from the correct document.
This is what retrieval looks like when the question's vocabulary matches the source text.

**Q3. "Is the Avalanche Lake Trail open right now?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-visitor-faq:02` | 0.6996 |
| 2 | `glacier-seasonal-closures-2026:04` | 0.6906 |
| 3 | `glacier-seasonal-closures-2026:07` | 0.6896 |

The FAQ answer beats the authoritative closure notice. Both say the trail is closed,
so the answer survives, but the citation points at the FAQ rather than the regulation.

**Q4. "Are there EV charging stations in Glacier National Park?" (unanswerable)**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-going-to-the-sun-road-guide:07` | 0.6991 |
| 2 | `glacier-backcountry-permit-regulations:07` | 0.6701 |
| 3 | `glacier-frontcountry-campground-regulations:06` | 0.6663 |

Retrieval never returns nothing. It returns the three least-bad chunks at scores
indistinguishable from Q1's winning chunk, which is exactly why a similarity threshold
is a bad refusal mechanism and the prompt has to carry the refusal.

## Step 1b: hybrid retrieval, top 3 (default, `--alpha 0.6`)

`combined = 0.60 * semantic + 0.40 * lexical`, where both signals are rescaled to
0..1 across all 241 chunks for this question. The tool prints all three numbers so
you can point at the reason a chunk won.

**Q1. "Can I have a campfire at Sperry Chalet in September?"**

Query terms by IDF: `sperry` (3.35, in 8 chunks), `chalet` (3.24, 9), `september`
(2.47, 20), `campfire` (2.01, 32).

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **0.9674** | 1.000 (0.6933) | 0.919 (8.77) | `glacier-backcountry-camping-guide:04` |
| 2 | 0.8230 | 0.969 (0.6851) | 0.604 (5.77) | `yosemite-campfire-regulations:03` |
| 3 | 0.8006 | 0.822 (0.6465) | 0.768 (7.34) | `glacier-seasonal-closures-2026:06` |

**Margin over rank 2: 0.1444, up from 0.0014.** Roughly a hundredfold. Acadia falls
out of the top 3 entirely, and the chunk that replaces it at rank 3 is a Glacier
document that names Sperry Chalet by site code. Rank 3 got there on the lexical
signal: at 0.822 normalized cosine it does not appear in the pure-cosine top 8 at all.

**Q2. "What is the maximum group size on a Glacier backcountry permit?"**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **1.0000** | 1.000 (0.8457) | 1.000 (15.34) | `glacier-backcountry-permit-regulations:04` |
| 2 | 0.8494 | 0.788 (0.7671) | 0.942 (14.46) | `glacier-backcountry-camping-guide:03` |
| 3 | 0.7482 | 0.673 (0.7248) | 0.860 (13.20) | `great-smoky-mountains-backcountry-guide:03` |

Rank 1 is unchanged and now wins on both signals at once. Ranks 2 and 3 changed, but
the answer never depended on them.

**Q3. "Is the Avalanche Lake Trail open right now?"**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **1.0000** | 1.000 (0.6996) | 1.000 (11.31) | `glacier-visitor-faq:02` |
| 2 | 0.9418 | 0.967 (0.6906) | 0.904 (10.22) | `glacier-seasonal-closures-2026:04` |
| 3 | 0.9356 | 0.924 (0.6788) | 0.952 (10.77) | `glacier-seasonal-closures-2026:06` |

Same winner, margin 0.0328 to 0.0582. Note what hybrid did **not** fix: the FAQ still
outranks the authoritative closure notice, because the FAQ genuinely uses more of the
question's words. If your product needs the regulation cited rather than the FAQ, the
fix is source-level weighting, not a lexical signal.

**Q4. "Are there EV charging stations in Glacier National Park?" (unanswerable)**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **0.8026** | 1.000 (0.6991) | 0.506 (4.36) | `glacier-going-to-the-sun-road-guide:07` |
| 2 | 0.7691 | 0.615 (0.6027) | 1.000 (8.60) | `glacier-bear-safety-advisory:03` |
| 3 | 0.7145 | 0.788 (0.6460) | 0.604 (5.20) | `glacier-backcountry-permit-regulations:03` |

Rank 1 is unchanged. Rank 2 is the funniest result in the module: "charging" appears
in exactly one chunk of the whole corpus, and it is the bear advisory telling you what
to do when a grizzly charges. IDF loves a rare word and has no idea what it means.
This is the honest cost of the lexical half, and it is why alpha is 0.6 and not 0.2.
The refusal holds anyway, which is the point: the prompt carries the refusal, not the
retrieval scores.

## Step 1c: the rephrasing that used to break it

Same question, four ways a visitor might type it. The number is the rank of
`glacier-backcountry-camping-guide:04`, the only chunk that actually answers it.

| question | rank, `--alpha 1.0` | rank, `--alpha 0.6` |
|---|---|---|
| "Can I have a campfire at Sperry Chalet in September?" | 1 (by 0.0014) | 1 (by 0.1444) |
| "Are campfires allowed at Sperry Chalet in the fall?" | 1 (by 0.0189) | 1 (by 0.0977) |
| "Can I build a fire at the Sperry Chalet campsites in September?" | 1 (by 0.0419) | 1 (by 0.1606) |
| **"What are the campfire rules at Sperry Chalet?"** | **7** | **1 (by 0.0494)** |

Row four is the one to run on stage. Drop the month and the verb, and pure cosine
puts `acadia-campfire-and-campground-regulations:04` first, four Yosemite sections
after it, and the correct chunk at rank 7, outside any sane top-k. The question still
contains the word "Sperry." The embedder does not care. Hybrid puts it back at rank 1.

## Step 2: the grounded answers

Seventy-two runs at default settings: 20 each on Q1, Q3, and Q4, 12 on Q2.

**Q1** (15/20 correct):

> [glacier-backcountry-camping-guide:04]
>
> According to Section 4.2 of the Glacier Backcountry Camping Guide, wood fires are
> prohibited year-round at all campsites in the Sperry Chalet area (site code SPE),
> regardless of season or posted fire danger.

**Check:** correct answer (no, and not because of September), a real chunk_id in
brackets, and no Acadia or Yosemite content bleeding in. One run cited two chunks as
`[glacier-backcountry-camping-guide:04, glacier-seasonal-closures-2026:06]`, a
comma-separated list inside one pair of brackets. Both ids are real and both were
retrieved, so it passes; the validator splits on commas for exactly this reason.

Five of the twenty went wrong, and they go wrong the same way every time: the model reads
Section 4.1's conditional "permitted only when the posted fire danger rating is below Very
High", stops before 4.2's absolute year-round ban, and answers yes. This is the same
failure the `--top-k 8` caveat below describes, and it is the reason Q1 rather than Q3 is
the question to watch when you change the prompt. It is unchanged by the date fix: the
build before it scored 14 of 16 on the same check.

**Q2** (12/12 correct):

> [glacier-backcountry-permit-regulations:04] The maximum party size on any
> backcountry permit is eight (8) persons.

**Check:** eight, cited. Format varies; the local model likes putting the citation first.

**Q3** (19/20 correct; it used to be roughly half, see "Telling the model what day it is"):

> [glacier-visitor-faq:02] No, the Avalanche Lake Trail is closed effective June 20, 2026,
> until further notice.

**Check:** closed, with a date and a citation.

**Q4, the refusal** (20/20 refused, 16 of them word for word):

> The provided documents don't say.

**Check:** no helpful speculation about charging infrastructure. Chunks about shuttle
parking, permit fees, and what to do when a bear charges were sitting right there in the
context, and the model declined to build an answer out of adjacent material every single
time. Twelve of the twenty runs appended a fabricated citation to the refusal; see the
next section.

Four of the twenty declined without using the exact sentence, in wording like "There is
no information about EV charging stations in the provided documents." Substantively that
is the same refusal, and none of the four invented a charging station. But it is no longer
a string anything downstream can match on, which is the difference between a refusal your
product can route on and a refusal a human has to read. All four came out of the citation
retry, which is the one code path that hands a finished refusal back to the model and asks
it to write the answer again.

## Citation validation

After generation, the demo pulls every `[...]` token containing a colon out of the
answer, splits comma-separated lists, and checks each id against the set of chunk_ids
actually placed in the context. Anything not in that set is a failure, printed loudly:

```
!! CITATION CHECK FAILED: [glacier-bear-safety-advisory:02] not in the retrieved set
!! retrying once with the valid chunk_ids spelled out
```

**Behavior on failure:** retry once with the valid ids listed verbatim in the prompt,
then strip whatever is still wrong and label it `[invalid-citation-removed]`. The one
exception is a refusal with a citation stapled to it, which is fixed in code without a
retry, for reasons measured below. The reasoning is in
[../dotnet/README.md](../dotnet/README.md).

What it caught across the runs above:

| question | invalid citation emitted | what happened |
|---|---|---|
| Q4, 12 of 20 runs | `glacier-bear-safety-advisory:02` | refusal detected in code, citation dropped, no retry |
| Q1, 5 of 20 runs | `glacier-backcountry-camping-guide:04.2` and `:04:04` | retry, then `[invalid-citation-removed]` where it stuck |

All of them are the same failure mode: a real document name with a section number the
model made up. `glacier-bear-safety-advisory:03` was in the context; `:02` was not.
Nothing in the answer text tells you that. Only the check does.

**Why the refusal case skips the retry.** `The provided documents don't say.` plus a
fabricated chunk_id is the single most common invalid citation in the module, and it is
not a question the model needs to reconsider: the answer is already right, and a refusal
by definition has no sources. Sending it back anyway was measurably harmful. Told to
"rewrite the answer using only those ids", `llama3.2` rewrites the refusal along with the
citation and returns "There is no information about EV charging stations in the provided
context." Still a refusal, no longer the contracted string. Deleting a citation is a
string operation, so the demo does it in code. Q4's word-for-word refusal rate went from
12 of 16 to 16 of 20 on that change alone, and every run that reached the model at all
still refused.

The obvious alternative, adding "if the context did not cover the question, answer exactly
`The provided documents don't say.`" to the retry prompt, was tried and rejected. It took
Q4 to 20 of 20 word for word and simultaneously made Q1 refuse in 6 of 20 runs: once the
retry prompt mentions the refusal, an answerable question that happens to trip the citation
check starts taking the exit. Offering a model an escape hatch in a repair prompt is not
free.

At `--top-k 8` on Q1, one run in five emitted
`[glacier-frontcountry-campground-regulations:02]`, again a real document with an
invented section. The retry produced a three-sentence answer citing three real
retrieved chunks. Earlier top-k 8 runs against pure-cosine retrieval produced
`[glacier-backcountry-campground-regulations:04]`, a chunk_id welded together out of
`glacier-backcountry-camping-guide` and `glacier-frontcountry-campground-regulations`,
pointing at a document that does not exist in any form. That is the whole argument for
validating in code instead of trusting the model to copy a string.

**Known gap:** the validator only recognizes square brackets. One run wrote
`(glacier-visitor-faq:02)` in parentheses and the check counted zero citations rather
than flagging anything. Worth mentioning if someone asks how airtight this is.

## Telling the model what day it is

Q3 used to be the most embarrassing thing in this module. "Is the Avalanche Lake Trail
open right now?" is a question the corpus answers twice over, retrieval puts the right
chunks at rank 1 and 2, and the model refused anyway. Measured on the build before this
fix: **10 refusals in 18 runs**, not the "about one in four" this file used to claim from
a three-run sample. On stage that is a coin flip on a question the lab presents as
answerable.

Retrieval was never the problem. The prompt was, and the specific problem was the calendar.

The corpus is written the way real operational documents are written, in dated notices:
"Avalanche Lake Trail: CLOSED effective June 20, 2026, until further notice." Answering
"is it open right now?" from that sentence takes one step the model cannot take, which is
knowing what "now" is. Given a notice with a start date, no end date, and no idea what
today is, refusing is not a malfunction. It is the correct answer to a question the model
genuinely cannot resolve.

Three experiments, sixteen runs each, pinned it down:

| what changed | Q3 refusals |
|---|---|
| nothing (baseline prompt) | 5/16 |
| drop "right now" from the question | **0/16** |
| add "Today's date is September 23, 2026." to the top of the prompt | 6/16 |

The second row proves the phrase "right now" is the trigger. The third row kills the
obvious fix: **the date alone does nothing.** A model handed a date and a dated notice
does not spontaneously connect them. You have to say what the date is *for*.

Where you say it turns out to matter more than what you say. A currency rule in the Rules
block, phrased broadly ("the context is the park's current status record, answer
present-tense questions from it"), fixed Q3 completely, held Q4's refusal at 32 of 32,
and quietly wrecked Q1: correct answers on the Sperry campfire question fell from 22 of 24
to **14 of 24**, with the model applying effective-date reasoning to a year-round fire ban
and concluding that campfires are allowed because nothing said the rule had expired. A
prompt rule you added for one question is a prompt rule that runs on every question.

What ships is narrower. The date rides along with the refusal clause, where the refusal
decision is actually made:

```
- If, and only if, none of the context is relevant to the question, reply exactly: "The provided documents don't say."
  A question about "right now" is answered from the context, not refused: today is September 23, 2026,
  and a notice that is in effect "until further notice" is still in effect right now.
```

Measured over 20 runs each on the finished demo:

| question | before | after |
|---|---|---|
| Q3, "open right now" | 10/18 refused | **1/20 refused** |
| Q4, EV charging (must refuse) | 12/16 word for word, 16/16 refused | 16/20 word for word, 20/20 refused |
| Q1, Sperry campfire | 0/16 refused, 14/16 correct | 0/20 refused, 15/20 correct |
| Q2, group size | 0/8 refused, 8/8 correct | 0/12 refused, 12/12 correct |

`today` is a constant in `Program.cs` with a comment saying so. Production passes
`DateTime.Today`; the demo pins a date so the outputs recorded in this file stay
reproducible.

**The leadership version of this:** a corpus of dated notices is useless to a model that
does not know the date, and almost every real knowledge base is a corpus of dated notices.
Policies with effective dates, incidents with open and close times, price lists, org charts,
on-call rotations. If your RAG system does not tell the model what "now" means, it will
either refuse questions it can answer or answer them as of an unknown date, and you will
not be able to tell which from the output.

## Caveats from twenty real runs

- **Q3 still refuses occasionally.** One run in twenty, down from better than half. If
  your lab refuses on a question you know is covered, run it twice before you go debugging
  retrieval.
- **Q4's refusal survives in substance, not always word for word.** Twenty of twenty runs
  declined; sixteen used the exact sentence. The gap is the citation retry, documented
  above. Do not tighten this by mentioning the refusal in the retry prompt; that trade was
  measured and it costs Q1 far more than it gains Q4.
- **Higher top-k makes Q1 worse, not better.** At `--top-k 8` the context picks up
  Glacier's frontcountry campground rules and Section 4.1's conditional "fires
  permitted" language, and one run in five concluded that campfires **are** allowed at
  Sperry in September. The correct chunk is still rank 1 with a 0.1444 margin. More
  context is not more grounding.
- **Scores are stable, answers are not.** Retrieval output is byte-identical run to
  run because the embeddings are cached. Everything downstream of generation varies.

## Stretch goal: watch the two signals fight at `--top-k 8`

`dotnet run -- --retrieval-only --top-k 8 --alpha 1.0` on Q1:

| rank | chunk_id |
|---|---|
| 1 | `glacier-backcountry-camping-guide:04` |
| 2 | `acadia-campfire-and-campground-regulations:07` |
| 3 | `yosemite-campfire-regulations:03` |
| 4 | `glacier-frontcountry-campground-regulations:04` |
| 5 | `acadia-campfire-and-campground-regulations:04` |
| 6 | `yosemite-campfire-regulations:05` |
| 7 | `yosemite-campfire-regulations:04` |
| 8 | `yosemite-campfire-regulations:06` |

Six of eight chunks are from the wrong park. Now `--alpha 0.6`:

| rank | chunk_id |
|---|---|
| 1 | `glacier-backcountry-camping-guide:04` |
| 2 | `yosemite-campfire-regulations:03` |
| 3 | `glacier-seasonal-closures-2026:06` |
| 4 | `glacier-frontcountry-campground-regulations:04` |
| 5 | `glacier-backcountry-camping-guide:02` |
| 6 | `yosemite-visitor-faq:03` |
| 7 | `glacier-visitor-faq:04` |
| 8 | `glacier-bear-safety-advisory:06` |

Acadia is gone, Yosemite drops from four chunks to two, and rank 8 is the bear
advisory that mentions Sperry's no-wood-fire status in passing, which is a legitimate
near-miss rather than a wrong-park one. Same corpus, same embedder, same question.
