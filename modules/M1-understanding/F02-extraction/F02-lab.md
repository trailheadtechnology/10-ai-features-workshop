# Lab 02: Extraction

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** turn a trip report into one JSON record of eight trail facts, `null` for anything unstated.
- **Input:** `data/tr-0007.md`, the fact-rich report; `data/tr-0011.md`, the fact-sparse one.
- **How:** POST to Ollama's chat endpoint with `format` set to a JSON schema. `http/ollama.http` holds both requests, ready to run.
- **Model:** `llama3.2`, local. No key.

### Step 1: Extract from `tr-0007.md`

Request 1 in `http/ollama.http`, or your track's starter, on `data/tr-0007.md`. The prompt, above the report:

```text
Extract the trail facts from this trip report. Use null for any field the report does not state, and empty arrays when nothing applies. Do not guess.
```

The schema in the request's `format` field:

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

**Check:** bare JSON matching request 1 in `expected-output.md`: `Sperry Chalet Trail`, `Glacier National Park`, `2026-07-04`, `12.8` (not the 6.4 one-way figure), `3400`. A preamble, a fence, or field names that change between runs means `format` is not reaching the request.

### Step 2: Extract from `tr-0011.md`

Request 2: the same prompt and schema on `data/tr-0011.md`, the sparse report. Run it three or four times; the inventions vary.

**Check:** missing facts come back `null` (`park` as `"Yosemite"` or `null` both pass). `"trail_name": "Yosemite Falls Trail"`, `"distance_mi": 7.2`, `elevation_gain_ft: 0`, or `date_hiked: "last month (exact date not specified)"` is an invention; step 3 catches it.

### Step 3: Fix the schema, then write the validator

Change the `distance_mi` and `elevation_gain_ft` descriptions to say this, and re-run request 2:

```text
null, never 0, when the report gives no figure
```

Then write the validator (`Validate` and `Clean` in `dotnet/complete/Program.cs`; `dotnet run` there prints verdicts and what would be stored). Coerce to `null`, never throw: `date_hiked` must parse against an explicit format list; `distance_mi` and `elevation_gain_ft` reject `0`, negatives, over 100 mi, over 20,000 ft; reject empty strings; `trail_name` and `park` must have their distinctive words in the report text.

**Check:** on `tr-0011.md`, the rejections in `expected-output.md` (`REJECT trail_name ""`, `REJECT date_hiked last month`, `REJECT distance_mi 0 mi`, `REJECT elevation_gain_ft 0 ft`), and `July 4, 2026` passes as `2026-07-04`. A `0` in the "what we would store" block means the range rule is missing.

### Stretch goal: per-field confidence, or one record per trail

Add a sibling `"confidence"` object to the schema, same keys, values 0 to 1; or make the top level an array of the record and feed it a report covering two trails.

**Check:** `tr-0011.md`'s confidences come back lower than `tr-0007.md`'s; the array version returns one record per trail. Every confidence at `1` on tr-0011 means the field is being ignored.

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

- `data/tr-0007.md`: the fact-rich report (a note-taker hikes Sperry Chalet on July 4, 2026; 12.8 miles, 3,400 feet, bears, goats, and pie, all stated outright)
- `data/tr-0011.md`: the fact-sparse report (a Yosemite trip written up a month late; no trail name, no date, no mileage, no elevation, because the little book is in the car at the mechanic's)
- `expected-output.md`: real `llama3.2` outputs for both requests and for the .NET demo, plus the success checks and an honest list of where the model still slips

The last step of the lab is the validator. The schema gets you JSON that parses; it does not get you JSON that is true. Once you have an extracted record, write the rejection rules:

- a date that no parser accepts is not a date, and "last month" belongs in `null`
- `0` miles and `0` feet are not measurements, they are a missing fact wearing a number
- an empty string is not a trail name
- a trail or park whose words never appear in the report is a name the model supplied

Reject means coerce to `null`, not throw. A gap is something a human can fill; a plausible wrong value is something nobody ever notices. The .NET version in `dotnet/complete/Program.cs` is roughly seventy lines of ordinary code, and `expected-output.md` has real rejections from four consecutive runs so you know what you are aiming at.
