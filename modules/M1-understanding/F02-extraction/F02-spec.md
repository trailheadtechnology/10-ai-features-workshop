# 02 · Extraction (structured output)

Module 1: Understanding · Runs on Ollama (`llama3.2`, JSON mode)

## The user problem

Trailhead Guides wants a "trail stats" panel: which trails were hiked, when, how far, what wildlife showed up, what shape the trail was in. All of that information already exists, scattered through forty rambling trip reports as prose. Today a human would have to re-read every report and re-type the facts into a form, which is why the panel doesn't exist.

## The concept

Extraction is summarization's sibling with one critical difference: the output is structured data your code can consume, not text a human reads. You hand the model a document and a schema, and you get back JSON ready for a database insert or an API response. This is the feature where the LLM stops being a chat feature and becomes a data-pipeline component.

Two mechanics matter. First, JSON mode: modern models, including local ones, can be constrained to emit valid JSON matching your schema, so you're not regex-parsing prose and hoping. Second, the schema does most of the prompting. Field names, descriptions, and an explicit "use `null` when the report doesn't say" rule do more for accuracy than any clever prompt wording.

Then there's the part most tutorials skip. The classic extraction failure is a missing fact hallucinated into a field, and the schema reduces that failure without eliminating it. On the sparse report in this lab, `llama3.2` returns `elevation_gain_ft: 0` where the honest answer is `null`, which is the more dangerous kind of miss: zero is a value, and a pipeline will store it without complaint. Run it a few times and you will also see invented distances and date strings that no date parser will accept. A structured shape guarantees the JSON parses; it does not guarantee the JSON is true. That is why this feature ends in validation code rather than in a prompt, and why what you ship is the schema plus a rejection rule. All of it works on a local model, which matters here more than anywhere: extraction pipelines often process private data at volume, where free and on-prem beats per-token pricing.

## Demo outline (about 12 min, .NET)

1. Show the same messy trip report from feature 01. This time the goal isn't a summary, it's a database row.
2. Define a C# record (`TripFacts`: trail, park, date, distance, wildlife, conditions) and use Microsoft.Extensions.AI's typed-response support to request it directly. The schema is code.
3. Run it. Prose in, populated .NET object out, with no parsing step. This is the payoff moment.
4. Break it on purpose: run a report that never mentions distance, and watch the model invent `distance_mi: 5.0`.
5. Fix it in the schema with nullable fields and "null when not stated" descriptions. Re-run: the invented distance usually goes away, and something else usually doesn't. Run it two or three times live so the room sees the variance rather than one lucky result. Then say the quiet part: this is better, and it is not a guarantee. The last mile is a validator that rejects a date like "early last month" and a `0` that should have been `null`, which is ordinary code your team already knows how to write.
6. Zoom out: loop over ten reports and print rows. That's an ingestion pipeline in thirty lines, and the "trail stats" panel is now just a query.

## Lab spec (Challenge lab, any language)

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Core lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** turn a trip report into a clean JSON record: `{trail, park, date, distance_mi, wildlife[], conditions}`.
- **Input:** two trip reports in `lab/`, one rich in facts and one missing several fields, drawn from `data/trip-reports/`.
- **How:** POST to Ollama's chat endpoint (`llama3.2`) with `format` set to a JSON schema. `lab/ollama.http` has the exact request, schema payload included.
- **Steps:**
  1. Extract from report #1 and check your JSON against `lab/expected-output.md`.
  2. Extract from report #2, the sparse one. Success check: most missing facts come back as `null` rather than plausible inventions, and you can name the ones that didn't. Compare against `lab/expected-output.md`, which records a real run where `elevation_gain_ft` came back `0` instead of `null`.
  3. If you got hallucinated values, first fix what you can in the schema descriptions rather than pleading in the prompt (`"null, never 0, when the report gives no figure"` is the fix for that one). Then write the validator for what the schema didn't catch: reject unparseable dates, absurd distances, and empty strings that should be `null`.
- **Stretch goal:** add a per-field `confidence` value, or handle multi-day reports by extracting an array of records, one per trail mentioned.

## Leadership beat

- **When to reach for this:** whenever valuable data is trapped in documents. Invoices, resumes, support emails, contracts, form uploads, legacy records.
- **Rough cost & effort:** days to a working pipeline. The real work is schema design and spot-checking accuracy. Local models keep volume processing free and private.
- **The one-liner for your CTO:** "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it."

This card is row 2 of the [decision framework](../../../docs/decision-framework.md).
