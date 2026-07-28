# What passing looks like

Wording varies run to run; the checks below are what has to be true. Every sample
here came from an actual run against local Ollama: `nomic-embed-text` for the 250
chunks in `chunks.jsonl`, `llama3.2` for generation. All scores are real.

Retrieval in the finished demo is **hybrid**: normalized cosine similarity blended
with a BM25-lite lexical score over the chunk text, weighted by `--alpha` (default
0.6). `--alpha 1.0` turns the lexical half off. Both sets are below.

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

This is module 04's search, verbatim.

**Q1. "Can I have a campfire at Sperry Chalet in September?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-backcountry-camping-guide:04.2` | 0.7422 |
| 2 | `glacier-frontcountry-campground-regulations:04.1-2` | 0.6913 |
| 3 | `yosemite-campfire-regulations:03` | 0.6812 |

The right chunk ranks first by 0.0509 of raw cosine. Ranks 2 and 3 are the frontcountry
campfire-season rule from the same park and a campfire rule from Yosemite, both of which
say fires **are** permitted under conditions. That is the shape of this question's
danger: the near-misses are not irrelevant, they are relevant and wrong.

**Q2. "What is the maximum group size on a Glacier backcountry permit?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-backcountry-permit-regulations:04` | 0.8412 |
| 2 | `glacier-backcountry-permit-regulations:05` | 0.7786 |
| 3 | `glacier-backcountry-permit-regulations:03` | 0.7774 |

Clean. 0.84 with a wide gap to second place, all three hits from the correct document.
This is what retrieval looks like when the question's vocabulary matches the source text.

**Q3. "Is the Avalanche Lake Trail open right now?"**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-seasonal-closures-2026:04.1` | 0.7114 |
| 2 | `glacier-seasonal-closures-2026:03` | 0.7078 |
| 3 | `glacier-seasonal-closures-2026:08` | 0.6850 |

Rank 1 is the closure notice itself, by 0.0036 over the road-status section of the same
document. Note what changed here: this used to be won by `glacier-visitor-faq:02`, the
plain-language FAQ. Splitting the 223-word closures section into its numbered
subsections gave the authoritative notice a chunk of its own, and it now outranks the FAQ
paraphrase of itself.

**Q4. "Are there EV charging stations in Glacier National Park?" (unanswerable)**

| rank | chunk_id | cosine |
|---|---|---|
| 1 | `glacier-going-to-the-sun-road-guide:07` | 0.6933 |
| 2 | `glacier-going-to-the-sun-road-guide:03` | 0.6637 |
| 3 | `glacier-going-to-the-sun-road-guide:06` | 0.6615 |

Retrieval never returns nothing. It returns the three least-bad chunks at scores in the
same range as Q1's winning chunk, which is exactly why a similarity threshold is a bad
refusal mechanism and the prompt has to carry the refusal.

## Step 1b: hybrid retrieval, top 3 (default, `--alpha 0.6`)

`combined = 0.60 * semantic + 0.40 * lexical`, where both signals are rescaled to
0..1 across all 250 chunks for this question. The tool prints all three numbers so
you can point at the reason a chunk won.

**Q1. "Can I have a campfire at Sperry Chalet in September?"**

Query terms by IDF: `sperry` (3.17, in 10 chunks), `chalet` (3.08, 11), `september`
(2.46, 21), `campfire` (1.75, 43).

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **0.9700** | 1.000 (0.7422) | 0.925 (8.47) | `glacier-backcountry-camping-guide:04.2` |
| 2 | 0.7379 | 0.837 (0.6913) | 0.589 (5.40) | `glacier-frontcountry-campground-regulations:04.1-2` |
| 3 | 0.7210 | 0.805 (0.6812) | 0.596 (5.46) | `yosemite-campfire-regulations:03` |

**Margin over rank 2: 0.2321, up from 0.1630 on the semantic signal alone.** The top 3
is the same set either way, which is itself worth saying on stage: once the chunks are
the right size, this question no longer needs hybrid to get rank 1 right. What hybrid
still buys you is visible at `--top-k 8` (bottom of this file), where pure cosine fills
the rest of the context with Acadia and Yosemite and hybrid fills it with Glacier.

**Q2. "What is the maximum group size on a Glacier backcountry permit?"**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **1.0000** | 1.000 (0.8412) | 1.000 (14.82) | `glacier-backcountry-permit-regulations:04` |
| 2 | 0.8150 | 0.762 (0.7574) | 0.894 (13.26) | `glacier-backcountry-camping-guide:03` |
| 3 | 0.7863 | 0.819 (0.7774) | 0.737 (10.93) | `glacier-backcountry-permit-regulations:03` |

Rank 1 wins on both signals at once, margin 0.1850. Section 4 of the permit regulations
is 110 words, under the split threshold, so it stays whole and the chunk_id has no
subsection number on it.

**Q3. "Is the Avalanche Lake Trail open right now?"**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **0.9749** | 1.000 (0.7114) | 0.937 (9.57) | `glacier-seasonal-closures-2026:04.1` |
| 2 | 0.9647 | 0.987 (0.7078) | 0.931 (9.51) | `glacier-seasonal-closures-2026:03` |
| 3 | 0.9422 | 0.904 (0.6846) | 1.000 (10.22) | `glacier-visitor-faq:02` |

Margin 0.0101, the tightest of the four, but all three chunks are from Glacier and two
of the three say the trail is closed. Rank 2 earns its place honestly: the road-status
section carries the forward-overlap line `(continues in 4.1) Avalanche Lake Trail: CLOSED
effective June 20, 2026, until further notice.` This file used to record that hybrid did
**not** fix the FAQ outranking the regulation. Chunking did.

**Q4. "Are there EV charging stations in Glacier National Park?" (unanswerable)**

| rank | combined | semantic (cos) | lexical (bm25) | chunk_id |
|---|---|---|---|---|
| 1 | **0.7876** | 1.000 (0.6933) | 0.469 (3.96) | `glacier-going-to-the-sun-road-guide:07` |
| 2 | 0.7424 | 0.571 (0.5996) | 1.000 (8.44) | `glacier-bear-safety-advisory:03` |
| 3 | 0.6838 | 0.764 (0.6417) | 0.564 (4.76) | `glacier-backcountry-permit-regulations:03` |

Rank 2 is the funniest result in the module: "charging" appears in exactly one chunk of
the whole corpus, and it is the bear advisory telling you what to do when a grizzly
charges. IDF loves a rare word and has no idea what it means. This is the honest cost of
the lexical half, and it is why alpha is 0.6 and not 0.2. The refusal holds anyway, which
is the point: the prompt carries the refusal, not the retrieval scores.

## Step 1c: the rephrasings

Same question, four ways a visitor might type it. The number is the rank of
`glacier-backcountry-camping-guide:04.2`, the only chunk that actually answers it.

| question | rank, `--alpha 1.0` | rank, `--alpha 0.6` |
|---|---|---|
| "Can I have a campfire at Sperry Chalet in September?" | 1 (by 0.1630) | 1 (by 0.2321) |
| "Are campfires allowed at Sperry Chalet in the fall?" | 1 (by 0.1375) | 1 (by 0.1595) |
| "Can I build a fire at the Sperry Chalet campsites in September?" | 1 (by 0.1779) | 1 (by 0.2865) |
| "What are the campfire rules at Sperry Chalet?" | 1 (by 0.0345) | 1 (by 0.0772) |

**Read this table next to the one it replaced.** On the old section-level chunks, row
four put `acadia-campfire-and-campground-regulations:04` first, four Yosemite sections
after it, and the correct chunk at rank **7** under pure cosine, and hybrid scoring was
what rescued it. That demo is gone, and the reason is worth more than the demo was:
Section 4 of the Glacier guide used to be 256 words of which only one sentence was about
Sperry, so "Sperry" was 1/256th of the vector. The same sentence now sits in a 130-word
chunk, and pure cosine finds it from every phrasing. Chunk size moved this question
further than the scoring function did.

Row four is still the thinnest margin of the four and still the one to run on stage;
hybrid roughly doubles it.

## Step 2: the grounded answers

Two hundred runs at default settings: 60 on Q1, 40 on Q2, 60 on Q3, 40 on Q4. Twenty
runs is enough to see a one-in-five defect and not enough to claim a 95 percent success
rate, so the samples here are sized for the claim being made.

**Q1** (58/60 correct, plus 1 that says both things at once):

> [glacier-backcountry-camping-guide:04.2] According to section 4.2 of the Glacier
> Backcountry Camping Guide, wood fires are prohibited year-round at all campsites in
> the Sperry Chalet area (site code SPE), regardless of season or posted fire danger;
> only pressurized-gas stoves are permitted for cooking at these sites. This prohibition
> applies in September and during the shoulder seasons, and is not lifted when fire
> danger is Low.

**Check:** correct answer (no, and not because of September), a real chunk_id in
brackets, and no Yosemite or frontcountry content bleeding in. Fifty-eight of sixty state
the year-round prohibition cleanly, and not one of the sixty says fires are permitted when
conditions allow. The old failure mode is gone.

The one failure is a different animal and worth reading:

> You can have a campfire at Sperry Chalet in September, but only pressurized-gas stoves
> are permitted for cooking at these sites year-round.

No citation, no prohibition stated, and the first clause is false. It did not come from
the wrong chunk; the right chunk was rank 1 with a 0.2321 margin, as it is on every run.
This is generation variance in a 3B model, and the honest reading is that chunking took
this question from a 1-in-5 systematic defect to a 1-in-60 random one.

**The wording problem this file will not paper over.** Roughly 45 percent of the correct
answers open with the word "Yes":

> Yes, you can have a pressurized-gas stove for cooking at the Sperry Chalet site, but
> wood fires are prohibited year-round there, regardless of season or posted fire danger.

> Yes, wood fires are prohibited year-round at all campsites in the Sperry Chalet area.

The first is answering a question nobody asked. The second is agreeing with itself. Both
are substantively right, both would fail a human review on a safety question, and one run
in sixty went all the way to "Yes, you can have a campfire at Sperry Chalet in September,
but it's prohibited year-round," which is a sentence that contradicts itself inside twenty
words. Chunking does not touch this. It is a prompt and a model-size question, and both were
measured: the prompt route made it worse in all three variants tried, and `qwen3:32b` drops
it to zero. See "The directness rule that did not work" and "What a bigger model buys".

**Q2** (40/40 correct):

> [glacier-backcountry-permit-regulations:04] The maximum party size on any
> backcountry permit is eight (8) persons.

**Check:** eight, cited. Format varies; the local model likes putting the citation first.

**Q3** (55/60 correct, 5/60 refused; 59/60 correct before the chunking change):

> [glacier-seasonal-closures-2026:04.1] No, the Avalanche Lake Trail is currently CLOSED
> effective June 20, 2026, until further notice due to the washed-out footbridge over
> Avalanche Creek.

**Check:** closed, with a date and a citation. Roughly half the runs cite the closure
notice, roughly half cite `glacier-visitor-faq:02`, and a few cite both.

**Q4, the refusal** (37/40 refused, 0/40 asserted that a charging station exists):

> The provided documents don't say.

**Check:** no run claimed the park has charging stations. Twenty-seven of forty used the
exact sentence somewhere in the answer and twelve used it alone with nothing else
attached; the rest declined in wording like "There is no information available in the
provided documents about electric vehicle (EV) charging stations within Glacier National
Park," which is substantively the same refusal but no longer a string anything downstream
can match on.

Q4 is the weakest of the four at 37 of 40, and the misses are all the same shape.
Three answered the question next door:

> No, fuel is not available anywhere within the Park [glacier-going-to-the-sun-road-guide:07].

The corpus says nothing about electric vehicles. It says fuel is unavailable, the model
treats that as close enough, and the answer opens with "No" to a question it never
addressed. The conclusion happens to be plausible, which is what makes it dangerous: the
grounding is wrong and the output looks fine. One earlier run in this series answered with a bare
`[glacier-bear-safety-advisory:03]` and no sentence at all, which is not a refusal, not
an answer, and not something any check in this demo catches.

## Chunking

This is the beat that was missing from the module, and it is the most expensive mistake
in it.

**The bug.** The first `chunks.jsonl` split every document at its numbered section
headings: one chunk per `## 4. Fires and Stoves`, 241 chunks in total. That is a
defensible default and it is what most people write first. Section 4 of the Glacier
backcountry guide is 256 words and it contains this, in this order:

> 4.1 Where campfires are authorized, they are permitted only in Park-installed metal
> fire rings [...] and only when the posted fire danger rating is below Very High.
>
> 4.2 [...] wood fires are prohibited year-round at all campsites in the Sperry Chalet
> area (site code SPE), regardless of season or posted fire danger [...]

A conditional rule, then the absolute exception that overrides it. Retrieval put that
chunk at rank 1 on every phrasing of the question, so by every retrieval metric the
system worked. `llama3.2` read 4.1, stopped, and told the visitor a campfire was fine as
long as fire danger was below Very High.

**Measured, 20 runs each, before and after:**

| question | before (241 section chunks) | after (250 chunks) |
|---|---|---|
| **Q1, Sperry campfire, correct** | **15/20 (75%)** | **58/60 (97%)** |
| Q1, wrong answer ("yes, if fire danger allows") | 4/20 | **0/60** |
| Q1, refused | 1/20 | 0/60 |
| Q2, group size, correct | 11/12 (92%) | **40/40 (100%)** |
| Q3, trail closure, correct | 59/60 (98%) | 55/60 (92%) |
| Q4, EV charging, refused | 18/20 (90%) | 37/40 (92%) |
| Q4, claimed a charging station exists | 2/20 | **0/40** |
| Q4, refusal string used verbatim and alone | 12/20 (60%) | 12/40 (30%) |

Read the sample sizes. The before column is 20 runs per question because that is what it
took to see the defect; the after column is 40 to 60 because that is what it takes to
claim a rate.

The four wrong Q1 answers all failed the same way. Two read 4.1 and stopped:

> According to the backcountry camping guide, campfires are permitted at Sperry Chalet
> area (site code SPE) when the posted fire danger rating is below Very High.

Two more borrowed a season from a different park's chunk in the same context:

> Since the Sperry Chalet area campsites opened for the season on July 1, and it's
> currently September, which falls outside of the restricted period (May 1 through
> September 30), you should be able to have a campfire at Sperry Chalet in September.

A visitor who acts on either of those lights a fire in a place where fires are banned
year-round. This module opens by saying a confidently wrong answer about fire regulations
is worse than no answer at all, and this was the module doing it on its own flagship
question, once every five runs.

**The fix, and why this one.** Four candidates were on the table.

- **Parent-section chunking, with a size budget.** Chosen. A section under 200 words is
  one idea and stays whole. Over that, it splits at the subsection boundaries the
  document already provides, and consecutive subsections are packed together until each
  chunk clears a 50-word floor. Five sections in the 25-document corpus were long enough
  to split, which is why the chunk count moved 241 to 250 rather than to 481. Section 4
  becomes `:04.1`, `:04.2`, and `:04.3-5`, and the Sperry prohibition is now a chunk that
  says one thing.
- **Contextual prefixing.** Adopted alongside it, not instead of it. Every chunk already
  carried its document title; it now carries its section heading too, so a retrieved
  `:04.2` still announces that it is Section 4, Fires and Stoves, of the Glacier
  backcountry guide. On its own this fixes nothing here: the model's problem was not that
  it did not know where the text came from.
- **Chunk overlap.** Adopted in one direction only. Every chunk ends with the opening
  sentence of whatever comes next in the document, marked `(continues in 4.2)`. In a
  regulation the exception follows the rule it modifies, so carrying the *next* unit's
  first sentence forward means a permissive rule can never be retrieved without a pointer
  to what qualifies it, while the prohibition is never preceded by someone else's
  permission. Symmetric overlap would have put 4.1's "fires permitted" language back at
  the top of 4.2's chunk, which is the bug.
- **Retrieve-then-expand**, pulling sibling subsections into the context after retrieval.
  Rejected. Expanding `:04.2` back out to its siblings reassembles the 256-word section
  that caused the failure. It is the right tool when a chunk is *missing* context; it is
  the wrong tool when a chunk contains context that contradicts it.

**Two honest costs.**

Full subsection splitting, with no size budget at all, was tried first: 481 chunks. It
took Q1 to 20/20 and Q3 to 0 refusals in 20, and it broke Q4. `glacier-going-to-the-sun-road-guide:07.2`
alone is nineteen words, "Fuel is not available anywhere within the Park. The nearest
stations are in West Glacier and St. Mary," and a nineteen-word chunk at rank 1 reads
like an answer even when it is not. Q4's refusals fell from 18/20 to 15/20,
with runs like "According to the park's guide, fuel (including electric vehicle charging)
is not available anywhere within the Park." Chunks that are too small invent confidence
the same way chunks that are too large hide the exception. The 200-word budget keeps
short list-like sections whole and is why that regression is not in the shipped numbers.

Q4's word-for-word refusal rate fell from 12/20 to 6/20 even in the shipped build, and
the mechanism is a good one to say out loud. In the old build the model attached a
fabricated chunk_id to its refusal in 12 of 20 runs, the citation check caught it, and
the code path that strips a citation off a refusal replaced the whole answer with the
exact contract string. Better chunks mean fewer fabricated citations, which means that
repair path fires less often, which means the raw model wording survives more often.
The refusal itself got *more* reliable (18/20 to 19/20, with the two fabricated "no EV
charging available" answers gone); the string got less uniform. If
your product depends on an exact string, get it from code, not from the model.

## The directness rule that did not work

Two defects in the numbers above have the same shape: the model reaches for the nearest
available assertion instead of answering the question that was asked. On Q1 it opens with
"Yes" and then explains that fires are prohibited, because it is answering about the gas
stove the context also mentions. On Q4 it says "No, fuel is not available anywhere within
the Park" to a question about electric vehicles. Neither is a retrieval problem. Both look
like a prompt problem, so a prompt rule was written, placed two ways, and measured.

The rule, in the Rules block (variant A):

```
- Answer the question that was asked, not a related one. If the context covers only a
  neighboring subject, that is not an answer. When the question is a yes-or-no question
  and the context does answer it, make the first word of your reply Yes or No.
```

The same rule attached to the answer instruction instead, just above `Answer:` (variant B).
Then a third try (variant C), same placement as A, with the Yes-or-No sentence removed and
replaced by "Lead with the answer to the question that was asked. If the context permits
one thing and prohibits another, say which one applies to what the visitor asked about
before mentioning the other."

Forty runs per question per variant, graded the same way as everything else. A terse reply
is graded on its merits: a bare "No" to "can I have a campfire" counts as correct.

| | Q1 correct | Q1 opens "Yes" | Q2 correct | Q3 correct | Q3 refused | Q4 refused | Q4 claimed a station exists |
|---|---|---|---|---|---|---|---|
| **shipped prompt** | **58/60 (97%)** | 24/60 (40%) | **40/40** | **55/60** | 5/60 | **37/40 (92%)** | **0/40** |
| A, Rules block | 7/40 (18%) | 35/40 (88%) | 33/40 | 34/40 | 2/40 | 34/40 (85%) | 0/40 |
| B, answer instruction | 5/40 (12%) | 36/40 (90%) | 21/40 | 33/40 | 1/40 | 13/40 (32%) | 0/40 |
| C, no Yes-or-No wording | 15/40 (38%) | 32/40 (80%) | 40/40 | 35/40 | 0/40 | 33/40 (82%) | 3/40 |

**All three made the metric they were written to fix worse.** The "Yes" opener rate went
from 40 percent to 80 or 90 percent. Q1 correctness fell from 97 percent to between 12 and
38. Variant C broke the one constraint that is not negotiable, producing three answers in
forty that told a visitor the park has EV charging.

**Why.** `llama3.2` reads any instruction about the shape of the opening as an instruction
to be brief. Under variant A the median answer to Q1 was the entire string
`Yes [glacier-backcountry-camping-guide:04.2].` Twenty words became two. The sentence that
got deleted is the one that carried the year-round prohibition, so the answer stopped being
correct. Then, forced to produce a bare verdict on "Can I have a campfire at Sperry Chalet
in September?", the model reached for the agreeable token. A rule written to stop the model
saying "Yes" doubled how often it said "Yes", and removed the explanation that had been
quietly making the answer right.

Variant B is the same failure with a wider blast radius. Moving the rule next to `Answer:`
puts it closest to the generation and it dominates everything above it, including the
refusal clause: Q4's refusal rate fell from 92 percent to 32.

**Placement mattered less than wording.** That is worth saying, because placement is what
this module's earlier date experiment turned on. Here A and B differ only in where the
identical sentence sits, and both land in the same ditch on Q1 while differing wildly on Q4.
Wording drove the Q1 result; placement drove how much collateral damage it did elsewhere.

**Shipped: nothing.** The prompt is unchanged. The 40 percent "Yes" opener rate is a real
wart, and the honest position is that a 3B model with a 20-word budget for hedging cannot
reliably lead with a verdict it has to qualify. See the next section for what fixes it.

## What a bigger model buys, and what it costs

The demo stays on `llama3.2`. This section exists because module 03 teaches measuring the
model-size question rather than arguing about it, and module 05 should be able to answer it
about its own pipeline. Same 250 chunks, same `nomic-embed-text` retrieval, same prompt,
same graders. Generation swapped to `qwen3:32b`, twenty runs per question. `qwen3` is a
reasoning model; no thinking output reached the answer text, and the graders strip
`<think>` blocks in any case.

| | llama3.2 (3B) | qwen3:32b |
|---|---|---|
| Q1 correct | 58/60 (97%) | **20/20 (100%)** |
| Q1 opens with "Yes" | 24/60 (40%) | **0/20 (0%)** |
| Q1 opens with "No" | 3/60 (5%) | 13/20 (65%) |
| Q2 correct | 40/40 (100%) | 20/20 (100%) |
| Q3 correct | 55/60 (92%) | **20/20 (100%)** |
| Q3 refused | 5/60 (8%) | **0/20** |
| Q4 substantive refusal | 37/40 (92%) | **20/20 (100%)** |
| Q4 refusal string, verbatim and alone | 13/40 (33%) | **20/20 (100%)** |
| Q4 claimed a station exists | 0/40 | 0/20 |
| invalid citations emitted | 9 in 200 runs | **0 in 80 runs** |
| median latency per answer | **0.9 s** | 16.6 s |

**What it buys.** Every open defect in this module closes. The "Yes" opener disappears
completely, and the answer the prompt could not produce is the one the bigger model writes
without being asked:

> No, wood fires are prohibited year-round at all campsites in the Sperry Chalet area (site
> code SPE), regardless of season or fire danger rating. Only pressurized-gas stoves are
> permitted for cooking at these sites [glacier-backcountry-camping-guide:04.2].

Q3's occasional refusal is gone. Q4 returns the contracted refusal string, alone, twenty
times out of twenty, against 33 percent for the small model. And in eighty runs it never
once fabricated a chunk_id.

**What it does not buy.** Retrieval is identical, because retrieval never changed: the same
embedder, the same 250 chunks, the same rank-1 margins. A bigger generation model would not
have found the Sperry prohibition inside a 256-word chunk that led with the opposite rule.
Chunking is upstream of model choice, and this is the cleanest statement of it in the
module: the chunking fix took Q1 from 75 to 97 percent on a 3B model, and the 32B model adds
the last 3 points. If you spend the money without fixing the chunk boundary, you are paying
20x to have a smarter model read the wrong context more fluently.

**What it costs.** Roughly 20x the latency, 0.9 seconds to 16.6 seconds per answer, on the
same machine with the same retrieval. That is the difference between a demo you can run
live and one where you talk over the pause. It is also 20 GB of resident model against 2 GB.

**The consequence for this module's own demo.** Two of its set pieces only work because the
model is small. Citation validation never fires on `qwen3:32b`, so the loud
`!! CITATION CHECK FAILED` moment is a `llama3.2` phenomenon. So is the "Yes" opener. Say
that out loud rather than letting the room conclude the check is unnecessary: the check is
correct either way, it is cheap either way, and the reason to keep it is that you do not
control which model you will be running on next quarter.

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
| Q4, 8 of 20 runs | `glacier-bear-safety-advisory:02` | refusal detected in code, citation dropped, no retry |
| Q4, 1 of 20 runs | `glacier-visitor-faq:00` | a real document, a section nobody retrieved; retry fixed it |
| Q3, 2 of 60 runs | `chunk_id: glacier-seasonal-closures-2026:04.1` | the word `chunk_id:` pasted inside the brackets; retry fixed it |
| Q1, 0 of 20 runs | none | see below |

`glacier-bear-safety-advisory:03` was in the context; `:02` was not. Nothing in the
answer text tells you that. Only the check does.

Note one thing the chunking change took away. On the old chunks, Q1's most common invalid
citation was `glacier-backcountry-camping-guide:04.2`: the model fusing the chunk_id
`:04` with the section number `4.2` it had just read. That id is now real, and the model
writing it is now correct. The failure did not get fixed, it got legislated out of
existence, which is worth a sentence on stage: half of "the model can't copy an
identifier" was really "the identifier didn't match the thing the model was reading."

**Why the refusal case skips the retry.** `The provided documents don't say.` plus a
fabricated chunk_id is the single most common invalid citation in the module, and it is
not a question the model needs to reconsider: the answer is already right, and a refusal
by definition has no sources. Sending it back anyway was measurably harmful. Told to
"rewrite the answer using only those ids", `llama3.2` rewrites the refusal along with the
citation and returns "There is no information about EV charging stations in the provided
context." Still a refusal, no longer the contracted string. Deleting a citation is a
string operation, so the demo does it in code.

The obvious alternative, adding "if the context did not cover the question, answer exactly
`The provided documents don't say.`" to the retry prompt, was tried and rejected. It took
Q4 to 20 of 20 word for word and simultaneously made Q1 refuse in 6 of 20 runs: once the
retry prompt mentions the refusal, an answerable question that happens to trip the citation
check starts taking the exit. Offering a model an escape hatch in a repair prompt is not
free.

**Known gap:** the validator only recognizes square brackets. One run wrote
`(glacier-visitor-faq:02)` in parentheses and the check counted zero citations rather
than flagging anything. Worth mentioning if someone asks how airtight this is.

## Telling the model what day it is

Q3 used to be the most embarrassing thing in this module. "Is the Avalanche Lake Trail
open right now?" is a question the corpus answers twice over, retrieval puts the right
chunks at rank 1 and 2, and the model refused anyway. Measured on the build before this
fix: **10 refusals in 18 runs**. On stage that is a coin flip on a question the lab
presents as answerable.

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

`today` is a constant in `Program.cs` with a comment saying so. Production passes
`DateTime.Today`; the demo pins a date so the outputs recorded in this file stay
reproducible.

**The leadership version of this:** a corpus of dated notices is useless to a model that
does not know the date, and almost every real knowledge base is a corpus of dated notices.
Policies with effective dates, incidents with open and close times, price lists, org charts,
on-call rotations. If your RAG system does not tell the model what "now" means, it will
either refuse questions it can answer or answer them as of an unknown date, and you will
not be able to tell which from the output.

## Caveats from the runs above

- **Q3 refusals went up, slightly.** 1 in 60 before the chunking change, 5 in 60 after.
  At that sample size the difference is not statistically distinguishable from noise, but
  it is a climb and it is reported as one. All five refusals cite `glacier-visitor-faq:02`
  and then decline, with the closure notice sitting at rank 1 in the same context. If your
  lab refuses on a question you know is covered, run it twice before you go debugging
  retrieval.
- **Q4's refusal survives in substance, more reliably than before, in wording less
  reliably.** Twenty of twenty declined; six used the exact sentence alone. See the
  chunking section for why those two numbers moved in opposite directions. Do not tighten
  this by mentioning the refusal in the retry prompt; that trade was measured and it costs
  Q1 far more than it gains Q4.
- **Forty percent of correct Q1 answers open with "Yes."** They then say fires are
  prohibited. This is the biggest remaining gap in the module and it is not a chunking gap:
  the retrieval is right, the substance is right, and the first word is wrong on a safety
  question. Both available fixes were measured. A prompt rule about answering directly made
  it worse in all three variants tried, taking the opener rate to 80 or 90 percent and Q1
  correctness to between 12 and 38 percent. `qwen3:32b` takes it to zero at 20x the latency.
  The prompt is therefore unchanged and the wart ships, with numbers attached.
- **Q4's four misses answer the question next door.** Three said "No, fuel is not
  available anywhere within the Park" to a question about EV charging. Same class of error
  as the "Yes" openers: the model reaching for the nearest available assertion rather than
  declining. Both would be caught by an answer-relevance check, which this demo does not
  have.
- **Higher top-k is no longer a trap on Q1, but do not assume that.** The `--top-k 8`
  context still picks up two chunks that say campfires are permitted under conditions.
  The rank-1 margin is 0.2321 and the answer held across the runs measured here, but more
  context is still not more grounding.
- **Scores are stable, answers are not.** Retrieval output is byte-identical run to
  run because the embeddings are cached. Everything downstream of generation varies.

## Stretch goal: watch the two signals fight at `--top-k 8`

`dotnet run -- --retrieval-only --top-k 8 --alpha 1.0` on Q1:

| rank | chunk_id |
|---|---|
| 1 | `glacier-backcountry-camping-guide:04.2` |
| 2 | `glacier-frontcountry-campground-regulations:04.1-2` |
| 3 | `yosemite-campfire-regulations:03` |
| 4 | `acadia-campfire-and-campground-regulations:07` |
| 5 | `yosemite-campfire-regulations:05` |
| 6 | `acadia-campfire-and-campground-regulations:04.1` |
| 7 | `glacier-frontcountry-campground-regulations:04.3-4` |
| 8 | `acadia-campfire-and-campground-regulations:04.4-5` |

Five of eight chunks are from the wrong park. Now `--alpha 0.6`:

| rank | chunk_id |
|---|---|
| 1 | `glacier-backcountry-camping-guide:04.2` |
| 2 | `glacier-frontcountry-campground-regulations:04.1-2` |
| 3 | `yosemite-campfire-regulations:03` |
| 4 | `glacier-seasonal-closures-2026:06` |
| 5 | `glacier-backcountry-camping-guide:02` |
| 6 | `glacier-backcountry-camping-guide:01` |
| 7 | `yosemite-visitor-faq:03` |
| 8 | `glacier-backcountry-permit-regulations:01` |

Acadia is gone entirely, Yosemite drops from three chunks to two, and ranks 4 through 6
are Glacier documents that name Sperry Chalet or its site code. Same corpus, same
embedder, same question. The top 3 did not move, which is the point: with the right
chunk size, hybrid scoring is no longer rescuing rank 1 on this question. It is cleaning
up everything behind it, and everything behind it is what the model reads next.
