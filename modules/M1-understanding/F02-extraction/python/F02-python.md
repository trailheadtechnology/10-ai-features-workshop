# Python Demo for 02 Extraction

Two scripts, both reading from [`../data/`](../data/):

- `starter/main.py`: the naive approach: ask for JSON in the prompt and print whatever comes back, preamble, markdown fences, drifting field names and all.
- `complete/main.py`: the finished demo as shown on stage. A pydantic `TripFacts` model with nullable fields and per-field descriptions, handed to `.parse()` so the reply comes back typed. Then the validator: grounding checks for names, strict date parsing, range and zero checks for the numbers, and rejected fields coerced to `null` before anything would be stored.

No setup here: the repo root has the `pyproject.toml`, and `uv sync` there (see [`SETUP.md`](../../../../SETUP.md)) is the one install for all ten features. `uv run` finds it from any folder. From `complete/`: (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
uv run main.py                              # both reports: extract, validate, show what we would store
uv run main.py ../../data/tr-0011.md         # just the sparse one, for the null check
uv run main.py ../../../F01-summarization/data/trip-reports/tr-0002.md   # any report path works
```

Run the sparse report three or four times. The output moves, and that is the demo: some runs come back clean, and some hand you `0` or "early last month", which is what the validator is for. The measured runs are in [`../expected-output.md`](../expected-output.md).

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), the Python equivalent of the .NET demo's Microsoft.Extensions.AI clients: swapping the provider is a different constructor and nothing else.

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F02-lab.md`](../F02-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter and Look at What "JSON in the Prompt" Gets You

The starter asks for JSON in prose. Run it twice on `tr-0007.md` and compare: field names drift, there may be a preamble or a markdown fence, and nothing guarantees it parses. This is what the schema replaces.

Run:

```bash
uv run main.py
```

Check: Two runs, two shapes.

### Step 2: Make the Schema Code and Ask for a Typed Response (lab step 1)

Define the record with every scalar nullable and a description on every field, then ask for that type instead of prose. The descriptions do the prompting; "null if the report does not state" on each field is the half of the hallucination fix that no plea in the prompt can do.

```python
from pydantic import BaseModel, Field

class TripFacts(BaseModel):
    trail_name: str | None = Field(description="The name of the trail hiked. null if the report never names the trail.")
    park: str | None = Field(description="The park the trail is in. null if the report never names the park.")
    date_hiked: str | None = Field(description="The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.")
    distance_mi: float | None = Field(description="Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.")
    elevation_gain_ft: float | None = Field(description="Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.")
    wildlife: list[str] = Field(description="Animals the author actually saw on this hike. Empty array if none are mentioned.")
    conditions: list[str] = Field(description="Short phrases describing trail conditions the report mentions. Empty array if none.")
    hazards: list[str] = Field(description="Hazards or closures the report mentions. Empty array if none.")

response = client.chat.completions.parse(
    model="llama3.2",
    messages=[{"role": "user", "content": f"""Extract the trail facts from this trip report.
Use null for any field the report does not state, and empty arrays
when nothing applies. Do not guess.

{report}"""}],
    response_format=TripFacts,
)
facts = response.choices[0].message.parsed
```

Run:

```bash
uv run main.py
```

Check: A populated object, no parsing step, and the values match the `tr-0007.md` block in `../expected-output.md` (Sperry Chalet Trail, 2026-07-04, 12.8 mi, 3400 ft).

### Step 3: Run the Sparse Report and Count What It Made up (lab step 2)

`tr-0011.md` never names the trail, gives no distance, no elevation, and no exact date. Run it three or four times and write down every field that came back with a value the report does not contain. The recorded runs in `../expected-output.md` show `elevation_gain_ft: 0` and `date_hiked: "early last month"`.

Run:

```bash
uv run main.py ../../data/tr-0011.md
```

Check: Most missing facts come back `null`, and you can name the ones that did not. That list is what the next step is for.

### Step 4: Fix What the Schema Can Fix, Then Write the Validator for the Rest (lab step 3)

First tighten the descriptions (the "null, never 0" wording above is that fix). Then add rules in code for what the schema cannot express: a date must parse in an explicit format, a measurement of 0 is not a measurement, a name must appear in the source text. Anything that fails is coerced to `null` before it could reach a database. The two rules below catch the two recorded failures; `complete/` has all five plus the grounding check, and the small `Verdict` type they return (field, value, passed, reason, optional normalized value) is defined there too.

```python
def in_range(field, value, maximum, unit):
    if value is None:
        return Verdict.ok(field, None)
    if value == 0:
        return Verdict.fail(field, f"{value:g} {unit}", "0 is not a measurement; the report gave no figure, so this should be null")
    if value > maximum:
        return Verdict.fail(field, f"{value:g} {unit}", f"implausible: over {maximum:.0f} {unit}")
    return Verdict.ok(field, f"{value:g} {unit}")

DATE_FORMATS = ["%Y-%m-%d", "%Y/%m/%d", "%m/%d/%Y", "%B %d, %Y", "%b %d, %Y", "%d %B %Y"]

def valid_date(field, value):
    if value is None:
        return Verdict.ok(field, None)
    for fmt in DATE_FORMATS:
        try:
            return Verdict.ok(field, value, normalized=datetime.strptime(value.strip(), fmt).date().isoformat())
        except ValueError:
            pass
    return Verdict.fail(field, value, "does not parse as a date; no parser can store this")
```

Run:

```bash
uv run main.py ../../data/tr-0011.md   # several times
```

Check: Every rejected field prints a reason, and "what we would store" has `null` where the model had `0` or prose. Zero is the dangerous one: it is a value, and a pipeline will store it without complaint. Stretch: add a per-field confidence, or extract an array of records for a multi-day report.
