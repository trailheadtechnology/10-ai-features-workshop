# Lab 02: Extraction

*A Challenge lab. Do it if you finished [Module 1](../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** turn a trip report into a clean JSON record: `{trail, park, date, distance_mi, wildlife[], conditions}`.
- **Input:** two trip reports in `data/`, one rich in facts and one missing several fields, drawn from feature 01's `data/trip-reports/`.
- **How:** POST to Ollama's chat endpoint (`llama3.2`) with `format` set to a JSON schema. `http/ollama.http` has the exact request, schema payload included.
- **Steps:**
  1. Extract from report #1 and check your JSON against `expected-output.md`.
  2. Extract from report #2, the sparse one. Success check: most missing facts come back as `null` rather than plausible inventions, and you can name the ones that didn't. Compare against `expected-output.md`, which records a real run where `elevation_gain_ft` came back `0` instead of `null`.
  3. If you got hallucinated values, first fix what you can in the schema descriptions rather than pleading in the prompt (`"null, never 0, when the report gives no figure"` is the fix for that one). Then write the validator for what the schema didn't catch: reject unparseable dates, absurd distances, and empty strings that should be `null`.
- **Stretch goal:** add a per-field `confidence` value, or handle multi-day reports by extracting an array of records, one per trail mentioned.

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
