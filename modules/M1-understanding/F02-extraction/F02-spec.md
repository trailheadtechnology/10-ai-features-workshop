# 02 · Extraction (Structured Output)

Module 1: Understanding · Runs on Ollama (`llama3.2`, JSON mode)

## The User Problem

Trailhead Guides wants a "trail stats" panel: which trails were hiked, when, how far, what wildlife showed up, what shape the trail was in. All of that information already exists, scattered through forty rambling trip reports as prose. Today a human would have to re-read every report and re-type the facts into a form, which is why the panel doesn't exist.

## The Concept

Extraction is summarization's sibling with one critical difference: the output is structured data your code can consume, not text a human reads. You hand the model a document and a schema, and you get back JSON ready for a database insert or an API response. This is the feature where the LLM stops being a chat feature and becomes a data-pipeline component.

Two mechanics matter. The first is JSON mode: modern models, including local ones, can be constrained to emit valid JSON matching your schema, so you're not regex-parsing prose and hoping. The second is that the schema does most of the prompting. Field names, descriptions, and an explicit "use `null` when the report doesn't say" rule do more for accuracy than any clever prompt wording.

Then there's the part most tutorials skip. The classic extraction failure is a missing fact hallucinated into a field, and the schema reduces that failure without eliminating it. On the sparse report in this lab, `llama3.2` returns `elevation_gain_ft: 0` where the honest answer is `null`, which is the more dangerous kind of miss: zero is a value, and a pipeline will store it without complaint. Run it a few times and you will also see invented distances and date strings that no date parser will accept. A structured shape guarantees the JSON parses; it does not guarantee the JSON is true. That is why this feature ends in validation code rather than in a prompt, and why what you ship is the schema plus a rejection rule. All of it works on a local model, which matters here more than anywhere: extraction pipelines often process private data at volume, where free and on-prem beats per-token pricing.

## The Lab

The hands-on lab is [F02-lab.md](F02-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** whenever valuable data is trapped in documents. Invoices, resumes, support emails, contracts, form uploads, legacy records.
- **Rough cost & effort:** days to a working pipeline. The real work is schema design and spot-checking accuracy. Local models keep volume processing free and private.
- **The one-liner for your CTO:** "We have years of data trapped in documents. This turns it into queryable rows without anyone re-typing it."

This card is row 2 of the [decision framework](../../../docs/decision-framework.md).
