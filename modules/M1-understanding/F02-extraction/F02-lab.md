# Lab 02: Extraction

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** turn a trip report into one JSON record of eight trail facts, `null` for anything unstated, then reject the values the model made up.
- **Input:** `data/tr-0007.md`, the fact-rich report; `data/tr-0011.md`, the fact-sparse one.
- **How:** send each report to `llama3.2` with a JSON schema as the response format, parse the reply into a record, then run five plain-code rules over the record. `http/ollama.http` holds both requests for the HTTP track, ready to run.
- **Model:** `llama3.2`, local. No key.

Every step below is one thing to make the program do. Each code track's `starter/` is the simple version that does not work yet. Edit it until it does all six steps. Look at `complete/` when you get stuck. On the HTTP track, run the numbered request in `http/ollama.http` when a step names one, and write the validator in any language you like.

### Step 0: Run the starter and read the shape you get

Run `starter/` as it is (`dotnet run` from `dotnet/starter/`, `uv run main.py` from `python/starter/`, `npm run starter` from `typescript/`). It reads `data/tr-0007.md`, drops the front matter (the block between the two `---` lines at the top), and sends this prompt with the report pasted under it. There is no schema yet.

`data/tr-0007.md` is one of the forty synthetic Trailhead Guides trip reports that feature 01 carries in its `data/`, copied here so this lab is self-contained. Each report is a markdown file with a short front matter block (`id`, `author`, `date`, `park`) between two `---` lines, then the report itself as prose. This one was picked because its author is a note-taker who states every fact the schema will ask for outright: trail name, park, date, round-trip distance, elevation gain, wildlife. The corpus was written for the workshop; there is no script that generated it.

```text
Extract the details of this trip report as JSON.
```

**Check:** a prose preamble, a markdown fence, and a nested shape the model invented, as in the starter block of `expected-output.md`. The shape of the mess is the same on every track. Run it twice and the field names change. Nothing here can be inserted into a database. The rest of the lab fixes that.

### Step 1: Send the schema with the request and extract from `tr-0007.md`

1. Read `data/tr-0007.md` into a string. The starter already resolves the path from its own folder; keep that.
2. Strip the front matter. Split the text on `---` into at most three parts. Keep the third part, trimmed. If there are not three parts, keep the whole text.
3. Build the user message. It is this prompt, then a blank line, then the report text.

```text
Extract the trail facts from this trip report. Use null for any field the report does not state, and empty arrays when nothing applies. Do not guess.
```

4. Define the record the model has to fill in. It has eight fields: five single values that may be `null`, and three arrays of strings. Every field has a description that says when to use `null`. On the code tracks, define it as a typed record or schema object named `TripFacts` in your language, with each description attached to its field; your track's walkthrough shows the mechanism, and the client turns it into this JSON schema for you. On the HTTP track, it is this schema, sent verbatim as the response format:

```json
{
  "type": "object",
  "properties": {
    "trail_name": {
      "type": [
        "string",
        "null"
      ],
      "description": "The name of the trail hiked. null if the report never names the trail."
    },
    "park": {
      "type": [
        "string",
        "null"
      ],
      "description": "The park the trail is in. null if the report never names the park."
    },
    "date_hiked": {
      "type": [
        "string",
        "null"
      ],
      "description": "The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date."
    },
    "distance_mi": {
      "type": [
        "number",
        "null"
      ],
      "description": "Round-trip distance in miles, as stated in the report. null if the report gives no distance. Never estimate."
    },
    "elevation_gain_ft": {
      "type": [
        "number",
        "null"
      ],
      "description": "Elevation gain in feet, as stated in the report. null if the report gives no elevation figure. Never estimate."
    },
    "wildlife": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Animals the author actually saw on this hike. Empty array if none are mentioned."
    },
    "conditions": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none."
    },
    "hazards": {
      "type": "array",
      "items": {
        "type": "string"
      },
      "description": "Hazards or closures the report mentions. Empty array if none."
    }
  },
  "required": [
    "trail_name",
    "park",
    "date_hiked",
    "distance_mi",
    "elevation_gain_ft",
    "wildlife",
    "conditions",
    "hazards"
  ]
}
```

5. Send the message from action 3 as a single user message to `llama3.2` through your track's chat client, with the schema as the response format and streaming off. On the code tracks you do not build the schema JSON by hand: hand the `TripFacts` record to the client's structured-output call and let it send the schema for you; your walkthrough shows the call. On the HTTP track, request 1 in `http/ollama.http` is this exact call with the report already pasted in. That file holds two numbered requests separated by `###` lines, one per report, each with the same prompt and schema and the report body (front matter already removed) pasted into the user message, so they run as-is. It was written by hand for this lab alongside the schema above; its number descriptions are the step 1 wording, so step 3's tightening is yours to apply there too.
6. Take the reply as the record. On the code tracks the structured-output call hands you the record already parsed. On the HTTP track the reply's message content is bare JSON with exactly the eight fields. Either way there is nothing to strip.
7. Print a heading with the file name. Then print the eight fields, one per line: trail, park, date, distance in mi, elev gain in ft, then the three arrays in square brackets. Print `null` for a null.

**Check:** bare JSON matching request 1 in `expected-output.md`: `Sperry Chalet Trail`, `Glacier National Park`, `2026-07-04`, `12.8` (not the 6.4 one-way figure), `3400`. The wildlife list catches most of the menagerie; the array phrasing wanders between runs, and that is fine. A preamble, a fence, or field names that change between runs means the schema is not reaching the request.

### Step 2: Extract from `tr-0011.md` and write down what it invented

1. Make the report path a command-line argument. Default it to `tr-0007.md`. Now the same program can run on either file.
2. Run it on `data/tr-0011.md`. Request 2 in `http/ollama.http` is the same prompt and schema with the sparse report pasted in. This file is the second copy from feature 01's forty reports, same shape as `tr-0007.md`, picked as its opposite: the author writes up a Yosemite waterfall hike a month late, never names the trail, and gives no date, mileage, or elevation, so every scalar the schema asks for is a fact the model has to leave `null` or invent. Its front matter does carry a `date` and a `park`, which is why the stripping in step 1 matters: the model sees only the body.
3. Run it three or four more times. Each time, write down every field that came back with a value the report does not contain.

**Check:** missing facts come back `null` (`park` as `"Yosemite"` or `null` both pass). `"trail_name": "Yosemite Falls Trail"`, `"distance_mi": 7.2`, `elevation_gain_ft: 0`, or `date_hiked: "last month (exact date not specified)"` is an invention. Zero is the one to watch: it is a value, and a pipeline will store it. Steps 3 and 4 catch these.

### Step 3: Tighten the two number descriptions

1. Change the `distance_mi` description to exactly this:

```text
Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.
```

2. Change the `elevation_gain_ft` description to exactly this:

```text
Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.
```

3. Run `tr-0011.md` three or four times again.

**Check:** `distance_mi` and `elevation_gain_ft` come back `null` on most runs. Some runs still return `0`, and some still put prose like `last month` in `date_hiked`. The schema got you JSON that parses; it does not get you JSON that is true. Step 4 is the fix for what the schema cannot express.

### Step 4: Write the validator, one rule per scalar field

1. Define a `Verdict` type with five parts. `field` is the field name. `value` is the value as a string, or `null`. `passed` is true or false. `reason` is set when the field failed. `normalized` is set when the value passed only after being rewritten.
2. Write the non-empty rule. It takes a field name and a value. A `null` passes. An empty or whitespace-only string fails with the reason `empty or whitespace-only string; should be null`. Anything else passes.
3. Write the grounding rule. It takes a field name, a value, and the report text, and it is for `trail_name` and `park`. Run the non-empty rule first. If it failed, or the value is `null`, return that verdict. Otherwise split the value on spaces, hyphens, commas, and apostrophes. Keep only the words that are at least 4 characters long and are not in this boilerplate list: `national`, `park`, `state`, `trail`, `trailhead`, `loop`, `canyon`, `falls`, `the`. Every kept word must appear somewhere in the report text, ignoring case. If no words were kept, look for the whole trimmed value in the report instead. If any word is missing, fail with the reason `not grounded in the source report (no mention of "word")`.
4. Write the date rule. It takes a field name and a value, and it is for `date_hiked`. Run the non-empty rule first, same as above. Trim the value. Try to parse it against an explicit short list of formats and no others, never a lenient parser: the ISO form `2026-07-04`, the same with slashes `2026/07/04`, the US slash form `7/4/2026` with one or two digits for month and day, and the spelled-out forms `July 4, 2026`, `Jul 4, 2026`, `4 July 2026`, and `July 4 2026`. If none of them parse, fail with the reason `does not parse as a date; no parser can store this`. If the year is before 1900 or after 2100, fail with `parses, but the year N is not plausible`. Otherwise pass, and set `normalized` to the date written in the ISO form, `2026-07-04`.
5. Write the range rule. It takes a field name, a value, a `max`, and a `unit`, and it is for the two numbers. A `null` passes. Exactly `0` fails with the reason `0 is not a measurement; the report gave no figure, so this should be null`. A negative number fails with `negative value is impossible`. A number over `max` fails with `implausible: over {max} {unit} for a single day hike`. Anything else passes. Call it with `max` 100 and unit `mi` for `distance_mi`. Call it with `max` 20000 and unit `ft` for `elevation_gain_ft`.
6. Write a validate function that takes the record and the report text. It returns five verdicts in this order: grounding on `trail_name`, grounding on `park`, the date rule on `date_hiked`, the range rule on `distance_mi`, the range rule on `elevation_gain_ft`. The three arrays get no rule. Looking for a name in the report text works. Looking for a free-text phrase like "some wet spots" does not.
7. After the step 1 printout, print a `-- what the validator says --` block. Print one line per verdict: `PASS` or `REJECT`, the field name, and the value (`null` when null). Under a `REJECT`, print `reason: ` and the reason. Under a `PASS` whose `normalized` differs from the value, print `normalized to: ` and the normalized value.

**Check:** on `tr-0011.md`, the rejections in `expected-output.md` (`REJECT trail_name ""`, `REJECT date_hiked last month`, `REJECT distance_mi 0 mi`, `REJECT elevation_gain_ft 0 ft`), each with its reason on the next line. On `tr-0007.md`, `July 4, 2026` passes with `normalized to: 2026-07-04`. Never throw from a rule; every field gets a verdict.

### Step 5: Coerce rejected fields to `null` and print what you would store

1. Write a clean function that takes the record and the five verdicts. It returns a new record. For each of the five single-value fields, look at its verdict. If it passed, keep the value, trimmed if it is a string. If it failed, store `null`. For `date_hiked`, store the verdict's `normalized` value when it has one.
2. Clean the three arrays. Drop empty and whitespace-only entries. Trim the rest. Nothing in an array is ever rejected.
3. Count the failed verdicts. When the count is 0, print `-- what we would store (nothing rejected this run) --`. Otherwise print `-- what we would store (N fields coerced to null) --`. Say `1 field`, not `1 fields`.
4. Print the cleaned record with the same eight-line printout from step 1.
5. Make the program loop over a list of report paths. When no path is given, default to both `tr-0007.md` and `tr-0011.md`. Run your program with no arguments.

**Check:** both reports print three blocks each: what the model gave us, what the validator says, what we would store. A `0` in the "what we would store" block means the range rule is missing. A `""` there means the non-empty rule is missing. `distance: null mi` and `elev gain: null ft` under a run that rejected two zeros is the passing output; it matches "Run 4" in `expected-output.md`. Reject means coerce to `null`, not throw: a gap is something a human can fill later, and a plausible wrong number is something nobody ever notices.

### Stretch goals

Pick any. The first two are not in `complete/`, so the checks are your only answer key.

- **Per-field confidence.** Add a second object named `"confidence"` to the schema, next to the eight fields. It has the same eight keys, and each value is a number from 0 to 1. Print it next to each field. **Check:** `tr-0011.md`'s confidences come back lower than `tr-0007.md`'s. Every confidence at `1` on tr-0011 means the model is ignoring the field.
- **One record per trail.** Make the top level of the schema an array of the same record. Feed it a report that covers two trails. **Check:** the array version returns one record per trail, and each record's `trail_name` grounds in the report on its own.
- **Any report.** Point the program at one of the other reports in feature 01's `data/`, for example `tr-0002.md`, by passing its path as the argument. **Check:** every scalar prints a `PASS` or a `REJECT` with a reason, and the validator never throws on a report it has not seen.
- **What the validator misses.** Run `tr-0011.md` until a run returns a made-up number that is still under the ceiling. `expected-output.md` records one such run: `distance_mi: 40`. **Check:** the range rule passes it. No cheap check can ground a number the way substring matching grounds a name, so what remains is sampling and review, which is a management answer, not a code answer.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F02-http.md`](http/F02-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F02-dotnet.md`](dotnet/F02-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F02-python.md`](python/F02-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F02-typescript.md`](typescript/F02-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tr-0007.md`: the fact-rich report (a note-taker hikes Sperry Chalet on July 4, 2026; 12.8 miles, 3,400 feet, bears, goats, and pie, all stated outright). Copied from feature 01's forty-report corpus, front matter and all.
- `data/tr-0011.md`: the fact-sparse report (a Yosemite trip written up a month late; no trail name, no date, no mileage, no elevation, because the little book is in the car at the mechanic's). The other copy from feature 01, chosen because the body states none of the scalars.
- `http/ollama.http`: the two requests for the HTTP track, one per report, with the prompt, the step 1 schema, and the report body pasted in
- `expected-output.md`: real `llama3.2` outputs for both requests and for the finished demo, the real rejections from four consecutive runs, the success checks, and an honest list of where the model still slips
