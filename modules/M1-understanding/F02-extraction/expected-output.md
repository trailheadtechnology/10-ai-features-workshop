# What Passing Looks Like

Exact values vary run to run; the checks below are what has to be true. These samples came from actual `llama3.2` runs of `http/ollama.http`.

## Request 1: Fact-Rich Report tr-0007 (Expect Every Field Populated)

```json
{
    "trail_name": "Sperry Chalet Trail",
    "park": "Glacier National Park",
    "date_hiked": "2026-07-04",
    "distance_mi": 12.8,
    "elevation_gain_ft": 3400,
    "wildlife": [
        "black bear sow with two cubs",
        "mountain goats",
        "hoary marmot",
        "mule deer",
        "ground squirrels"
    ],
    "conditions": [
        "dry tread, recently cleared, with fresh sawcuts",
        "stream crossings with snowmelt flow",
        "lingering snow patch",
        "open subalpine terrain",
        "cool breeze"
    ],
    "hazards": [
        "black bear sow with two cubs",
        "mountain goats on ledges"
    ]
}
```

**Check:** trail, park, date, 12.8 (round trip, not the 6.4 one-way figure), and 3400 all land, and the wildlife list catches most of the menagerie. Array phrasing wanders between runs; that's fine. This is a database row extracted from five paragraphs of pie appreciation.

## Request 2: Fact-Sparse Report tr-0011 (Success Check: Nulls, Not Inventions)

```json
{
    "trail_name": null,
    "park": "Yosemite",
    "date_hiked": null,
    "distance_mi": null,
    "elevation_gain_ft": 0,
    "wildlife": [],
    "conditions": [],
    "hazards": []
}
```

**Check:** the report never names the trail, gives no date, no mileage, no elevation. If your output says `"trail_name": "Yosemite Falls Trail"` or `"distance_mi": 7.2`, the model invented it and your schema descriptions need work. `park` may come back `"Yosemite"` (the body does say "the Yosemite trip") or `null`; both are defensible.

**Honest teaching points from real runs:**

- In the run above, `elevation_gain_ft` came back `0` instead of `null`. That is a near-miss hallucination: zero is a value, and a pipeline would happily insert it. Some runs return `null` correctly. If you see `0`, tighten the description ("null, never 0, when the report gives no figure") and re-run.
- Other runs populated `conditions` with `["fine", "some wet spots"]` and `hazards` with the fellow doing "the arm-windmill thing". Also defensible: the report does mention wet spots and a slip. Empty arrays and those entries are both passing; a fabricated bear sighting is not.

## The .NET Starter (`dotnet run` from `dotnet/starter/`, the Failure You're Supposed to Get)

The naive prompt, no schema. Real output:

````text
Here's the extracted trip report details as JSON:

```json
{
  "Trip": {
    "Name": "Sperry Chalet Trail",
    "Distance": { "OneWay": 6.4, "RoundTrip": 12.8 },
    "ElevationGain": 3400,
    ...
````

Three problems in the first six lines: a prose preamble your parser has to strip, a markdown fence around the payload, and a nested shape the model invented on the spot. Run it twice and the field names change. Nothing here is wrong, exactly. It just isn't a contract, and no database insert can consume it.

## The .NET complete demo (`dotnet run` from `dotnet/complete/`)

Real output from a run where the model behaved and the validator had nothing to do:

```text
== tr-0007.md ==

-- what the model gave us --
  trail:      Sperry Chalet Trail
  park:       Glacier National Park
  date:       2026-07-04
  distance:   12.8 mi
  elev gain:  3400 ft
  wildlife:   [black bear sow with two cubs, mountain goats (9 total), hoary marmot]
  conditions: [dry tread, recently cleared, open subalpine terrain, lingering snow patch (40ft)]
  hazards:    []

-- what the validator says --
  PASS    trail_name         Sperry Chalet Trail
  PASS    park               Glacier National Park
  PASS    date_hiked         2026-07-04
  PASS    distance_mi        12.8 mi
  PASS    elevation_gain_ft  3400 ft

-- what we would store (nothing rejected this run) --
  trail:      Sperry Chalet Trail
  park:       Glacier National Park
  date:       2026-07-04
  distance:   12.8 mi
  elev gain:  3400 ft
  ...
```

Same checks as the raw JSON above. The variance is still there run to run: `park` has come back as an empty string for tr-0007, and the sparse report's date has come back as `"last month (exact date not specified)"` instead of `null`. The model extracted rather than invented, but it ignored the "null if no exact date" instruction, which is exactly why the run does not end at the extracted object. A bigger model follows the null rule more reliably; on `llama3.2`, treat it as a live demo of why the schema-plus-validation pair exists. What the validator does with those runs is the next section.

## The validator (the part you actually ship)

The complete demo no longer stops at the extracted object. It prints three blocks per report: what the model gave us, what the validator says field by field, and what we would store after rejected fields are coerced to `null`. The rules are ordinary code:

- `date_hiked` must parse against an explicit format list, or be `null`. A real date in an odd format ("July 4, 2026") passes and gets normalized to `2026-07-04`. Prose like "last month" is rejected outright.
- `distance_mi` and `elevation_gain_ft` reject `0`, reject negatives, and reject the absurd (over 100 mi, over 20,000 ft). Zero is the near-miss this rule exists for: it is a value, and nothing downstream will ever question it.
- Empty and whitespace-only strings are rejected. `""` is not a trail name.
- `trail_name` and `park` get a grounding check: the distinctive words in the value have to appear somewhere in the source report. Boilerplate ("National", "Park", "Trail") is ignored, so `"Glacier National Park"` still grounds on a report that only ever says "Glacier", while a trail name the report never mentions does not.

### Real rejections, observed across four consecutive runs

Every one of these came out of an actual `dotnet run` against `llama3.2`. No run was clean across both reports two times running.

**Run 1, tr-0011 gave an empty string for the trail:**

```text
-- what the validator says --
  REJECT  trail_name         ""
          reason: empty or whitespace-only string; should be null
  PASS    park               Yosemite
  PASS    date_hiked         null
```

**Run 3, tr-0011 gave prose in the date column:**

```text
  REJECT  date_hiked         last month
          reason: does not parse as a date; no parser can store this
```

**Run 4, tr-0011 gave the textbook near-miss, `0` for both figures:**

```text
== tr-0011.md ==

-- what the model gave us --
  trail:      null
  park:       null
  date:       null
  distance:   0 mi
  elev gain:  0 ft
  wildlife:   []
  conditions: [fine, some wet spots]
  hazards:    [fellow slipped and caught himself]

-- what the validator says --
  PASS    trail_name         null
  PASS    park               null
  PASS    date_hiked         null
  REJECT  distance_mi        0 mi
          reason: 0 is not a measurement; the report gave no figure, so this should be null
  REJECT  elevation_gain_ft  0 ft
          reason: 0 is not a measurement; the report gave no figure, so this should be null

-- what we would store (2 fields coerced to null) --
  trail:      null
  park:       null
  date:       null
  distance:   null mi
  elev gain:  null ft
  wildlife:   []
  conditions: [fine, some wet spots]
  hazards:    [fellow slipped and caught himself]
```

**Run 1, tr-0007 got the date right in the wrong format, and the validator fixed rather than rejected it:**

```text
  PASS    date_hiked         July 4, 2026
          normalized to: 2026-07-04
```

That last one matters on stage. Some model output is correct and merely off-format, and normalizing it instead of rejecting it is what keeps people from switching the validator off.

### What the validator does not catch

One earlier run returned `distance_mi: 40` for tr-0011, a report that states no mileage anywhere. Forty miles is under the 100-mile ceiling, so the range rule passed it, and no cheap check can ground a number the way substring matching grounds a name. Say this out loud: validation moves the failure from silent-and-wrong to loud-and-wrong for a whole class of bugs, and it does not get you to zero. What remains is a sampling and review problem, which is a management answer, not a code answer.

## Stretch goal

Per-field confidence: add a sibling `"confidence"` object to the schema (same keys, values 0 to 1) and see whether the sparse report's confidences drop. Multi-day reports: change the schema's top level to an array of the same record and feed it a report that covers two trails.
