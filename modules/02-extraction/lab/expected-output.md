# What passing looks like

Exact values vary run to run; the checks below are what has to be true. These samples came from actual `llama3.2` runs of `ollama.http`.

## Request 1: fact-rich report tr-0007 (expect every field populated)

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

## Request 2: fact-sparse report tr-0011 (success check: nulls, not inventions)

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

## The .NET starter (`dotnet run` from `dotnet/starter/`, the failure you're supposed to get)

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

Real output from one run:

```text
== tr-0007.md ==
  trail:      Sperry Chalet Trail
  park:
  date:       2026-07-04
  distance:   12.8 mi
  elev gain:  3400 ft
  wildlife:   [black bear sow with two cubs, mountain goats (9 total, including 2 kids), hoary marmot]
  conditions: [dry tread, recently cleared, with fresh sawcuts on two big deadfall logs, substantial snowmelt flow in some streams, one lingering snow patch crossed the trail in a shaded gully, good rock lines for creek crossings]
  hazards:    [burned area from years back (snags standing silver against green regrowth)]

== tr-0011.md ==
  trail:      null
  park:       Yosemite
  date:       last month (exact date not specified)
  distance:   null mi
  elev gain:  null ft
  wildlife:   []
  conditions: []
  hazards:    []
```

Same checks, plus two honest wrinkles from `llama3.2` worth pointing at on stage: `park` came back as an empty string for tr-0007 in this run (other runs say "Glacier National Park"), and the sparse report's date came back as the string `"last month (exact date not specified)"` instead of `null`. The model extracted rather than invented, but it ignored the "null if no exact date" instruction, which is why real pipelines validate model output (is this parseable as a date?) instead of trusting it. A bigger model follows the null rule more reliably; on `llama3.2`, treat it as a demo of why the schema-plus-validation pair exists.

## Stretch goal

Per-field confidence: add a sibling `"confidence"` object to the schema (same keys, values 0 to 1) and see whether the sparse report's confidences drop. Multi-day reports: change the schema's top level to an array of the same record and feed it a report that covers two trails.
