# What passing looks like

Wording varies run to run; the checks below are what has to be true. Every sample
here came from an actual run against local Ollama: `nomic-embed-text` for the 241
chunks in `chunks.jsonl`, `llama3.2` for generation. Cosine scores are real.

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

## Step 1: retrieval, top 3 per question

Real cosine similarity, question vector against all 241 chunk vectors.

**Q1. "Can I have a campfire at Sperry Chalet in September?"**

| rank | chunk_id | score |
|---|---|---|
| 1 | `glacier-backcountry-camping-guide:04` | 0.6933 |
| 2 | `acadia-campfire-and-campground-regulations:07` | 0.6929 |
| 3 | `yosemite-campfire-regulations:03` | 0.6851 |

The right chunk ranks first, by 0.0004. Ranks 2 and 3 are campfire rules from the
wrong parks entirely. This is the honest state of retrieval on a 25-document corpus
where five documents are all about campfires: the embedding model is matching on
"campfire regulations," not on "Sperry." Drop to top-1 and you get lucky; a slightly
different phrasing of the question could easily push Acadia to the top. Worth saying
out loud in the demo.

**Q2. "What is the maximum group size on a Glacier backcountry permit?"**

| rank | chunk_id | score |
|---|---|---|
| 1 | `glacier-backcountry-permit-regulations:04` | 0.8457 |
| 2 | `glacier-backcountry-permit-regulations:05` | 0.7802 |
| 3 | `glacier-backcountry-permit-regulations:01` | 0.7769 |

Clean. 0.85 with a wide gap to second place, and all three hits are from the correct
document. This is what retrieval looks like when the question's vocabulary matches the
source text.

**Q3. "Is the Avalanche Lake Trail open right now?"**

| rank | chunk_id | score |
|---|---|---|
| 1 | `glacier-visitor-faq:02` | 0.6996 |
| 2 | `glacier-seasonal-closures-2026:04` | 0.6906 |
| 3 | `glacier-seasonal-closures-2026:07` | 0.6896 |

The FAQ answer beats the authoritative closure notice. Both say the trail is closed,
so the answer survives, but the citation points at the FAQ rather than the regulation.
If your product needs the regulation cited, retrieval scoring is where you fix that,
not the prompt.

**Q4. "Are there EV charging stations in Glacier National Park?" (unanswerable)**

| rank | chunk_id | score |
|---|---|---|
| 1 | `glacier-going-to-the-sun-road-guide:07` | 0.6991 |
| 2 | `glacier-backcountry-permit-regulations:07` | 0.6701 |
| 3 | `glacier-frontcountry-campground-regulations:06` | 0.6663 |

Retrieval never returns nothing. It returns the three least-bad chunks at scores
indistinguishable from Q1's winning chunk, which is exactly why a similarity threshold
is a bad refusal mechanism and the prompt has to carry the refusal.

## Step 2: the grounded answers

**Q1**, from the top-3 context above:

> According to section 4.2 of the backcountry camping guide,
> [glacier-backcountry-camping-guide:04], wood fires are prohibited year-round at all
> campsites in the Sperry Chalet area (site code SPE), regardless of season or posted
> fire danger. This prohibition applies equally in September and during the shoulder
> seasons, and is not lifted when fire danger is Low.

**Check:** correct answer (no, and not because of September), a chunk_id in brackets,
and no Acadia or Yosemite content bleeding in from ranks 2 and 3. `llama3.2` sometimes
writes the citation as `[glacier-backcountry-camping-guide:04.2]`, fusing the chunk_id
with the section number. Sloppy but traceable; a real product would validate citations
against the retrieved set.

**Q2:**

> [glacier-backcountry-permit-regulations:04]
>
> The maximum party size on any backcountry permit is eight (8) persons.

**Check:** eight, cited. Format varies; the local model likes putting the citation first.

**Q3:**

> The Avalanche Lake Trail is currently closed effective June 20, 2026, until further
> notice. [glacier-visitor-faq:02]

**Check:** closed, with a date and a citation.

**Q4 (the refusal):**

> The provided documents don't say.

**Check:** exactly this, with no citation and no helpful speculation about charging
infrastructure. Three chunks about parking, permit fees, and campground utilities were
sitting right there in the context, and the model declined to build an answer out of
adjacent material.

## Prompt wording matters more than you want it to

An earlier version of the grounded prompt ended with "If the context does not contain
the answer, reply exactly: ..." That version refused on **Q3**, even though the context
plainly said the trail was closed. The model read "is the trail open right now" and
decided a document could not speak to "right now." Rewording to "If, and only if, none
of the context is relevant to the question" fixed Q3 and left the Q4 refusal intact.

Same retrieval, same model, opposite outcome, one clause of difference. If your lab
answers "the documents don't say" to a question you know is covered, suspect the
refusal clause before you suspect retrieval.

## Stretch goal: vary top-k on Q1

Real runs of `dotnet run -- --top-k N` from `dotnet/complete/`.

**k=1** (only `glacier-backcountry-camping-guide:04` in context):

> According to the rule for the Sperry Chalet area, wood fires are prohibited year-round
> there, regardless of season or posted fire danger rating. [glacier-backcountry-camping-guide:04.2]

Correct and tight, because the one chunk retrieved happened to be the right one, by 0.0004.

**k=8** pulls in five more campfire sections from Yosemite, Acadia, and Glacier's
frontcountry regs. The answer stays correct and even quotes the regulation verbatim,
but the citation degrades:

> [glacier-backcountry-campground-regulations:04]
>
> According to Section 4.2, "wood fires are prohibited year-round at all campsites in
> the Sperry Chalet area (site code SPE) [...]"

`glacier-backcountry-campground-regulations:04` is not a real chunk_id. The model fused
`glacier-backcountry-camping-guide` with `glacier-frontcountry-campground-regulations`,
two documents that were both sitting in the context. The receipt points at a document
that does not exist. That is the whole argument for validating citations against the
retrieved set in code rather than trusting the model to copy a string.
