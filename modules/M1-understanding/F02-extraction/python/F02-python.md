<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 02: Extraction (Python)

*You are on the Python track. Other tracks: [.NET](../dotnet/F02-dotnet.md), [TypeScript](../typescript/F02-typescript.md). Lab overview: [F02-lab.md](../F02-lab.md).*

**The User Problem:** Trailhead Guides wants a "trail stats" panel: which trails were hiked, when, how far, what wildlife showed up, what shape the trail was in. All of that information already exists, scattered through forty rambling trip reports as prose. Today a human would have to re-read every report and re-type the facts into a form, which is why the panel doesn't exist.

*A Challenge lab. Do it if you finished [Module 1](../../M1-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** turn a trip report into one JSON record of eight trail facts, `null` for anything unstated, then reject the values the model made up.
- **Input:** `data/tr-0007.md`, the fact-rich report (copied from feature 01's corpus); `data/tr-0011.md`, the fact-sparse one.
- **How:** send each report to `llama3.2` with a JSON schema as the response format, parse the reply into a record, then run five plain-code rules over the record.
- **Model:** `llama3.2`, local. No key.

## The Concept

Extraction is summarization's sibling with one critical difference: the output is structured data your code can consume, not text a human reads. You hand the model a document and a schema, and you get back JSON ready for a database insert or an API response. This is the feature where the LLM stops being a chat feature and becomes a data-pipeline component.

Two mechanics matter. The first is JSON mode: modern models, including local ones, can be constrained to emit valid JSON matching your schema, so you're not regex-parsing prose and hoping. The second is that the schema does most of the prompting. Field names, descriptions, and an explicit "use `null` when the report doesn't say" rule do more for accuracy than any clever prompt wording.

Then there's the part most tutorials skip. The classic extraction failure is a missing fact hallucinated into a field, and the schema reduces that failure without eliminating it. On the sparse report in this lab, `llama3.2` returns `elevation_gain_ft: 0` where the honest answer is `null`, which is the more dangerous kind of miss: zero is a value, and a pipeline will store it without complaint. Run it a few times and you will also see invented distances and date strings that no date parser will accept. A structured shape guarantees the JSON parses; it does not guarantee the JSON is true. That is why this feature ends in validation code rather than in a prompt, and why what you ship is the schema plus a rejection rule. All of it works on a local model, which matters here more than anywhere: extraction pipelines often process private data at volume, where free and on-prem beats per-token pricing.

Every step below is one thing to make the program do. The `starter/` is the simple version that does not work yet. Edit it until it does all six steps. Look at `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Switching the provider later is a different constructor and nothing else. `starter/main.py` asks for JSON in the prompt and prints whatever comes back. `complete/main.py` is the finished demo as shown on stage: a pydantic `TripFacts` model handed to `.parse()`, then the validator. The repo root's `pyproject.toml` is the one install for all ten features: run `uv sync` there once (see [SETUP.md](../../../../SETUP.md)), and `uv run` finds it from any folder, with no venv to activate. Run `uv run main.py` from the `starter/` folder; the starter takes one optional report path as its only argument.

`complete/` runs both lab reports by default and takes any report paths instead. From `complete/`:

```bash
uv run main.py                              # both reports: extract, validate, show what we would store
uv run main.py ../../data/tr-0011.md         # just the sparse one, for the null check
uv run main.py ../../../F01-summarization/data/tr-0002.md   # any report path works
```

### Step 0: Run the starter and read the shape you get

**Do:** run `starter/` as it is. It reads `data/tr-0007.md`, drops the front matter, and sends `Extract the details of this trip report as JSON.` with the report pasted under it. There is no schema yet. Run it twice:

```bash
uv run main.py
```

Here is how it works, in `starter/main.py`. Line one is the path (the first argument, or `tr-0007.md`), line two reads the file and drops the front matter, and the call sends the naive prompt and prints the raw reply text:

```python
report_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DATA / "tr-0007.md"
report = strip_front_matter(report_path.read_text())

response = client.chat.completions.create(
    model="llama3.2",
    messages=[{"role": "user", "content": f"Extract the details of this trip report as JSON.\n\n{report}"}],
)
print(response.choices[0].message.content)
```

**Why:** `tr-0007.md` is a note-taker's report that states every fact the schema will later ask for outright (trail name, park, date, distance, elevation, wildlife), so any gap you see here is the model's fault, not the source's.

**Check:** the reply opens with a line of prose, wraps the JSON in a markdown fence, and uses a nested shape the model made up on the spot, as in the starter block of [`expected-output.md`](../expected-output.md). Run it a few times and the shape moves around, usually the field names (when I ran it, `trip_date` became `log_date` and the nesting changed). Nothing in it is wrong about the report; it just isn't a contract, so no database insert can consume it.

### Step 1: Send the schema with the request and extract from `tr-0007.md`

**Do:**
1. Review how the existing code from the starter project reads `data/tr-0007.md` and strips the front matter (split on `---`, keep the third part, trimmed).

   `strip_front_matter` is the function near the top of `starter/main.py`. `split("---", 2)` splits at most twice, so the report body (which may contain its own `---`) stays in one piece as `parts[2]`:

   ```python
   def strip_front_matter(markdown: str) -> str:
       parts = markdown.split("---", 2)
       return parts[2].strip() if len(parts) == 3 else markdown.strip()
   ```

   and the read is the line under `report_path = ...`:

   ```python
   report = strip_front_matter(report_path.read_text())
   ```

2. Build the user message: the prompt below, a blank line, then the report text.

   ```text
   Extract the trail facts from this trip report. Use null for any field the report does not state, and empty arrays when nothing applies. Do not guess.
   ```

   In Python the prompt, the blank line, and `{report}` all go into one `f"""..."""` string, written straight into the call in item 4 below. The `f` makes `{report}` insert the report text. The lines after the first one start at the very left edge, with no indentation, because every space inside the quotes is sent to the model.

3. Define the record: eight fields, five scalars that may be `null` and three string arrays, each with a description saying when to use `null`. Define it as a typed record/schema object named `TripFacts`; your client turns it into a schema equivalent to this (key casing may follow your language's convention; the descriptions are what matter). Attaching a description in the wrong place silently drops every description from the schema, and the Check below will not catch that, so copy the form exactly:

   ```json
   {
     "type": "object",
     "properties": {
       "trail_name": {"type": ["string", "null"], "description": "The name of the trail hiked. null if the report never names the trail."},
       "park": {"type": ["string", "null"], "description": "The park the trail is in. null if the report never names the park."},
       "date_hiked": {"type": ["string", "null"], "description": "The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date."},
       "distance_mi": {"type": ["number", "null"], "description": "Round-trip distance in miles, as stated in the report. null if the report gives no distance. Never estimate."},
       "elevation_gain_ft": {"type": ["number", "null"], "description": "Elevation gain in feet, as stated in the report. null if the report gives no elevation figure. Never estimate."},
       "wildlife": {"type": "array", "items": {"type": "string"}, "description": "Animals the author actually saw on this hike. Empty array if none are mentioned."},
       "conditions": {"type": "array", "items": {"type": "string"}, "description": "Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none."},
       "hazards": {"type": "array", "items": {"type": "string"}, "description": "Hazards or closures the report mentions. Empty array if none."}
     },
     "required": ["trail_name", "park", "date_hiked", "distance_mi", "elevation_gain_ft", "wildlife", "conditions", "hazards"]
   }
   ```

4. Send the message as a single user message to `llama3.2` with the schema as the response format, streaming off. Hand `TripFacts` to the client's structured-output call, so you never write the schema JSON by hand.

   - The record is a `pydantic` `BaseModel` (pydantic ships with the `openai` package, nothing to install). The description goes in `Field(description=...)` as the field's default value; a comment or docstring next to the field never reaches the schema. The two numbers are `float | None`, the three lists are `list[str]`.
   - The call is `client.chat.completions.parse(...)` (not `.create`) with `response_format=TripFacts`. The SDK builds the JSON schema from the model and sends it; you never write the schema by hand.
   - The parsed record is `response.choices[0].message.parsed`, already a `TripFacts` instance.

   The two number descriptions below already carry step 3's `null, never 0,` wording; leave those two words out for now if you want to see step 2's inventions first.

   First, add this import at the top of `main.py`, under `from openai import OpenAI`:

   ```python
   from pydantic import BaseModel, Field
   ```

   Then paste the class under the `DATA = ...` line, above `def strip_front_matter`. A class must be defined above any code that uses it. Each line is one field: its name, its type (`str | None` means "a string or `None`", and `None` is Python's `null`), and `= Field(description=...)`:

   ```python
   class TripFacts(BaseModel):
       trail_name: str | None = Field(description="The name of the trail hiked. null if the report never names the trail.")
       park: str | None = Field(description="The park the trail is in. null if the report never names the park.")
       date_hiked: str | None = Field(description="The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.")
       distance_mi: float | None = Field(description="Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.")
       elevation_gain_ft: float | None = Field(description="Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.")
       wildlife: list[str] = Field(description="Animals the author actually saw on this hike. Empty array if none are mentioned.")
       conditions: list[str] = Field(description="Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none.")
       hazards: list[str] = Field(description="Hazards or closures the report mentions. Empty array if none.")
   ```

   Last, replace the starter's call (every line from `response = client.chat.completions.create(` through `print(response.choices[0].message.content)`) with the typed call. `response_format=TripFacts` is the only thing that turns on the schema:

   ```python
   response = client.chat.completions.parse(
       model="llama3.2",
       messages=[{"role": "user", "content": f"""Extract the trail facts from this trip report.
   Use null for any field the report does not state, and empty arrays
   when nothing applies. Do not guess.

   {report}"""}],
       response_format=TripFacts,
   )
   ```

5. Take the reply as the record (already parsed).

   Directly under the call's closing `)`, `.parsed` is already a `TripFacts`. The `assert` stops the program with an error if the model sent nothing parseable back:

   ```python
   raw = response.choices[0].message.parsed
   assert raw is not None
   ```

6. Print a heading with the file name, then, under a `-- what the model gave us --` line, the eight fields labelled `trail:`, `park:`, `date:`, `distance: N mi`, `elev gain: N ft`, `wildlife: [...]`, `conditions: [...]`, `hazards: [...]`, as in the complete-demo block of `expected-output.md`. Print `null` for a null.

   A `show()` function gives you the eight labelled lines, formatting the two floats so `3400.0` prints as `3400 ft`, matching `expected-output.md`. Put these three lines under `assert raw is not None`. `report_path.name` is just the file name, without the folders, and the `\n` adds the blank line:

   ```python
   print(f"== {report_path.name} ==\n")
   print("-- what the model gave us --")
   show(raw)
   ```

   and the function, copied from `complete/main.py`, under `def strip_front_matter` (so below `class TripFacts`, above `report_path = ...`). `f.trail_name or 'null'` prints the text `null` when the value is `None`; `fmt` does the same for a number and formats it otherwise; `', '.join(...)` glues a list into one comma-separated string:

   ```python
   def show(f: TripFacts) -> None:
       fmt = lambda v, spec: "null" if v is None else format(v, spec)
       print(f"  trail:      {f.trail_name or 'null'}")
       print(f"  park:       {f.park or 'null'}")
       print(f"  date:       {f.date_hiked or 'null'}")
       print(f"  distance:   {fmt(f.distance_mi, '.1f').rstrip('0').rstrip('.') if f.distance_mi is not None else 'null'} mi")
       print(f"  elev gain:  {fmt(f.elevation_gain_ft, '.0f')} ft")
       print(f"  wildlife:   [{', '.join(f.wildlife or [])}]")
       print(f"  conditions: [{', '.join(f.conditions or [])}]")
       print(f"  hazards:    [{', '.join(f.hazards or [])}]")
   ```

Run it:

```bash
uv run main.py
```

**Why:** the field descriptions do the prompting here, not the sentence above the report. Saying "null if the report does not state" on every field is what stops the model from filling a blank with a plausible guess on a report like `tr-0011.md`, which never names the trail and gives no date or mileage. It does not get you all the way there: `llama3.2` still returns `0` for `elevation_gain_ft` on some runs, and `0` is a value a pipeline will store without complaint.

**Check:** the fields line up with request 1 in [`expected-output.md`](../expected-output.md): `Sperry Chalet Trail`, `Glacier National Park`, `2026-07-04`, `12.8` (not the 6.4 one-way figure), `3400`. Expect the odd run to miss one, most often `park` coming back null. The wildlife list catches most of the menagerie; array phrasing wandering between runs is fine. A preamble, a fence, or field names that change between runs means the schema is not reaching the request.

### Step 2: Extract from `tr-0011.md` and write down what it invented

**Do:**
1. Review how the existing code from the starter project makes the report path a command-line argument, defaulting to `tr-0007.md`.

   `sys.argv[1]` is whatever you type after `uv run main.py`, and `DATA` is the feature's `data/` folder:

   ```python
   report_path = Path(sys.argv[1]) if len(sys.argv) > 1 else DATA / "tr-0007.md"
   ```

2. Run it on `data/tr-0011.md`, three or four times.

   ```bash
   uv run main.py ../../data/tr-0011.md
   ```

3. Each time, write down every field that came back with a value the report does not contain.

**Why:** `tr-0011.md` never names the trail and gives no date, mileage, or elevation, so every scalar the schema asks for is a fact the model has to leave `null` or invent.

**Check:** missing facts come back `null` (`park` as `"Yosemite"` or `null` both pass). `"trail_name": "Yosemite Falls Trail"`, `"distance_mi": 7.2`, `elevation_gain_ft: 0`, or `date_hiked: "last month (exact date not specified)"` is an invention. Zero is the one to watch: it is a value, and a pipeline will store it. Steps 3 and 4 catch these.

### Step 3: Tighten the two number descriptions

**Do:** the `distance_mi` and `elevation_gain_ft` descriptions need `, never 0,` before their `null` clause. Step 1's code already carries those two words, so unless you left them out to watch step 2 invent numbers, this step is a read rather than an edit:

```text
Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.
Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.
```

In `class TripFacts` near the top of `main.py`, the `distance_mi` and `elevation_gain_ft` lines should end up reading exactly like this (if you pasted step 1's class as shown, they already do):

```python
    distance_mi: float | None = Field(description="Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.")
    elevation_gain_ft: float | None = Field(description="Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.")
```

```bash
uv run main.py ../../data/tr-0011.md
```

Run `tr-0011.md` three or four more times.

**Check:** `distance_mi` and `elevation_gain_ft` come back `null` on most runs. Some runs still return `0`, and some still put prose like `last month` in `date_hiked`. The schema got you JSON that parses; it does not get you JSON that is true. Step 4 is the fix for what the schema cannot express.

### Step 4: Write the validator, one rule per scalar field

**Do:**
1. Define a `Verdict` type: `field`, `value` (as a string, or `null`), `passed`, `reason` (set on failure), `normalized` (set when a value passed only after being rewritten). Give it a pass helper and a fail helper for building one.

   Copied from `complete/main.py`. A `@dataclass` is a class whose fields are listed like `TripFacts`'s, and Python writes the constructor for you: `Verdict(field, value, passed, reason, normalized)`, where the last two are optional. The two `@staticmethod` helpers are called on the class itself, as `Verdict.ok(...)` and `Verdict.fail(...)`. First, add these imports at the top of `main.py`, under `import sys`:

   ```python
   from dataclasses import dataclass
   from datetime import date, datetime
   ```

   Then paste the class under `class TripFacts`:

   ```python
   @dataclass
   class Verdict:
       field: str
       value: str | None
       passed: bool
       reason: str | None = None
       # Set when a field passed only after being rewritten, e.g. a date the model
       # wrote as "July 4, 2026" that we store as "2026-07-04".
       normalized: str | None = None

       @staticmethod
       def ok(field, value, normalized=None):
           return Verdict(field, value, True, None, normalized)

       @staticmethod
       def fail(field, value, reason):
           return Verdict(field, value, False, reason)
   ```

   Each rule below is a function under `def show`, above `report_path = ...`, that takes a field name and a value and returns one `Verdict`.

2. **Non-empty rule:** `null` passes; an empty/whitespace-only string fails with `empty or whitespace-only string; should be null`; anything else passes.

   Copied from `complete/main.py`. `value.strip()` removes the spaces from both ends, and an empty string counts as false, so `not value.strip()` means "nothing but whitespace":

   ```python
   def non_empty(field: str, value: str | None) -> Verdict:
       if value is None:
           return Verdict.ok(field, None)
       if not value.strip():
           return Verdict.fail(field, f'"{value}"', "empty or whitespace-only string; should be null")
       return Verdict.ok(field, value)
   ```

3. **Grounding rule** (for `trail_name`, `park`): run the non-empty rule first. Otherwise split the value on spaces/hyphens/commas/apostrophes, keep words ≥4 characters that are not in this boilerplate list (`national`, `park`, `state`, `trail`, `trailhead`, `loop`, `canyon`, `falls`, `the`), and require every kept word to appear in the report text (case-insensitive). If no words were kept, look for the whole trimmed value instead. Fail with `not grounded in the source report (no mention of "word")` if any is missing.

   `complete/` has the full `grounded` function. The shape, with the parts for you to fill in marked `...`. The list inside `[...]` is a comprehension: it keeps each `w` from the split value that passes the `if`. `x in text` is `True` when `x` appears anywhere in `text`, and lowering both sides makes it ignore case:

   ```python
   # Hint: the building blocks of grounded(field, value, source)
   BOILERPLATE = {"national", "park", ...}  # the rest of the list

   def grounded(field: str, value: str | None, source: str) -> Verdict:
       basic = non_empty(field, value)
       if not basic.passed or value is None:
           return basic
       words = [w for w in value.replace("-", " ").replace(",", " ").replace("'", " ").split()
                if len(w) >= 4 and w.lower() not in BOILERPLATE]
       if not words:
           words = [...]  # no words kept: check the whole trimmed value instead
       missing = [w for w in words if w.lower() not in source.lower()]
       if missing:
           return Verdict.fail(field, value, f'not grounded in the source report (no mention of "{missing[0]}")')
       return Verdict.ok(field, value)
   ```

4. **Date rule** (for `date_hiked`): run the non-empty rule first. Try only an explicit list of formats (`2026-07-04`, `2026/07/04`, `7/4/2026`, `July 4, 2026`, `Jul 4, 2026`, `4 July 2026`, `July 4 2026`), never a lenient parser. Fail with `does not parse as a date; no parser can store this` if none match, or `parses, but the year N is not plausible` if the year is before 1900 or after 2100. Otherwise pass with `normalized` set to the ISO form.

   Copied from `complete/main.py`. `DATE_FORMATS` is the format list above as `strptime` codes (`%Y` year, `%m` month number, `%d` day, `%B` full month name, `%b` short month name). `datetime.strptime` raises `ValueError` when the text does not match a format, so the loop catches that and tries the next one; `break` leaves the loop at the first match. `isoformat()` writes the date as `2026-07-04`:

   ```python
   DATE_FORMATS = ["%Y-%m-%d", "%Y/%m/%d", "%m/%d/%Y", "%B %d, %Y", "%b %d, %Y", "%d %B %Y", "%B %d %Y"]


   def valid_date(field: str, value: str | None) -> Verdict:
       basic = non_empty(field, value)
       if not basic.passed or value is None:
           return basic
       parsed: date | None = None
       for fmt in DATE_FORMATS:
           try:
               parsed = datetime.strptime(value.strip(), fmt).date()
               break
           except ValueError:
               continue
       if parsed is None:
           return Verdict.fail(field, value, "does not parse as a date; no parser can store this")
       if parsed.year < 1900 or parsed.year > 2100:
           return Verdict.fail(field, value, f"parses, but the year {parsed.year} is not plausible")
       # Parseable but off-format gets normalized on the way to storage.
       return Verdict.ok(field, value, normalized=parsed.isoformat())
   ```

5. **Range rule** (for the two numbers, `max`/`unit` supplied): `null` passes; exactly `0` fails with `0 is not a measurement; the report gave no figure, so this should be null`; negative fails with `negative value is impossible`; over `max` fails with `implausible: over {max} {unit} for a single day hike`. Call it with `max=100, unit="mi"` for `distance_mi` and `max=20000, unit="ft"` for `elevation_gain_ft`. The verdict's `value` for a number is the number followed by its unit (`0 mi`, `3400 ft`), which is what the PASS/REJECT line prints.

   Copied from `complete/main.py`. `{value:g}` is what turns `3400.0` into the `3400 ft` the PASS/REJECT line prints, and `{maximum:.0f}` prints the ceiling with no decimals:

   ```python
   def in_range(field: str, value: float | None, maximum: float, unit: str) -> Verdict:
       if value is None:
           return Verdict.ok(field, None)
       shown = f"{value:g} {unit}"
       if value == 0:
           return Verdict.fail(field, shown, "0 is not a measurement; the report gave no figure, so this should be null")
       if value < 0:
           return Verdict.fail(field, shown, "negative value is impossible")
       if value > maximum:
           return Verdict.fail(field, shown, f"implausible: over {maximum:.0f} {unit} for a single day hike")
       return Verdict.ok(field, shown)
   ```

6. Write `validate(record, report_text)` returning five verdicts in order: grounding on `trail_name`, grounding on `park`, date rule on `date_hiked`, range rule on `distance_mi`, range rule on `elevation_gain_ft`. The three arrays get no rule.

   A function under the rules, above `report_path = ...`. It returns a list of the five verdicts:

   ```python
   # Hint: five rule calls, in the order the lab lists them
   def validate(f: TripFacts, source: str) -> list[Verdict]:
       return [
           grounded("trail_name", f.trail_name, source),
           # ...park, then the date rule, then distance_mi...
           in_range("elevation_gain_ft", f.elevation_gain_ft, 20000, "ft"),
       ]
   ```

7. After the step 1 printout, print a `-- what the validator says --` block: one line per verdict, `PASS`/`REJECT`, field name, value (`null` when null); under a `REJECT` print `reason: `; under a `PASS` whose `normalized` differs from the value, print `normalized to: `.

   Copied from `complete/main.py`. Put it directly under `show(raw)`, at the left edge like `show(raw)` (the indented lines belong to the `for` and the `if`). `{v.field:<18}` pads the field name to 18 characters so the values line up:

   ```python
   verdicts = validate(raw, report)

   print("\n-- what the validator says --")
   for v in verdicts:
       mark = "PASS  " if v.passed else "REJECT"
       print(f"  {mark}  {v.field:<18} {v.value if v.value is not None else 'null'}")
       if not v.passed:
           print(f"          reason: {v.reason}")
       elif v.normalized is not None and v.normalized != v.value:
           print(f"          normalized to: {v.normalized}")
   ```

8. Run the sparse report several times:

   ```bash
   uv run main.py ../../data/tr-0011.md   # several times
   ```

**Why:** `trail_name` and `park` are the only fields that get the grounding rule, because a name is something you can go looking for in the report text. The rule ignores boilerplate words like "National", "Park", and "Trail", so `Glacier National Park` still grounds on a report that only ever says "Glacier". A free-text entry like "some wet spots" gives you nothing to match on, and neither does a number, which is why `distance_mi` gets a range check instead.

**Check:** on `tr-0011.md`, the rejections in [`expected-output.md`](../expected-output.md) (`REJECT trail_name ""`, `REJECT date_hiked last month`, `REJECT distance_mi 0 mi`, `REJECT elevation_gain_ft 0 ft`), each with its reason on the next line. `llama3.2` sometimes returns the text `"null"` instead of a real null; the validator rejects it (for a name, with `no mention of "null"`), and that rejection is correct, because a string that says null is not a missing value. On `tr-0007.md`, `July 4, 2026` passes with `normalized to: 2026-07-04`. Never throw from a rule; every field gets a verdict.

### Step 5: Coerce rejected fields to `null` and print what you would store

**Do:**
1. Write `clean(record, verdicts)` returning a new record: for each of the five scalar fields, keep the (trimmed, if string) value if its verdict passed, else store `null`; for `date_hiked`, store the verdict's `normalized` value when set.

   A function under `validate`, above `report_path = ...`. A pydantic model is built by naming every field, so this makes a new `TripFacts` rather than changing the old one. `v` is a dictionary from field name to verdict, so `v["park"]` is the park's verdict, and `ok("park")` is `True` when it passed:

   ```python
   # Hint: build a new TripFacts, keeping each scalar only if its verdict passed
   def clean(f: TripFacts, verdicts: list[Verdict]) -> TripFacts:
       v = {x.field: x for x in verdicts}
       ok = lambda name: v[name].passed
       return TripFacts(
           trail_name=f.trail_name.strip() if ok("trail_name") and f.trail_name else None,
           park=...,  # same pattern as trail_name
           date_hiked=(v["date_hiked"].normalized or ...) if ok("date_hiked") else None,
           distance_mi=f.distance_mi if ok("distance_mi") else None,
           elevation_gain_ft=...,  # same pattern as distance_mi
           wildlife=f.wildlife,  # item 2 replaces these three lines
           conditions=f.conditions,
           hazards=f.hazards,
       )
   ```

2. Clean the three arrays: drop empty/whitespace-only entries, trim the rest. Nothing in an array is ever rejected.

   Copied from `complete/main.py`. The helper goes above `def clean`; the comprehension keeps each entry that is not blank and `s.strip()` trims it:

   ```python
   def clean_list(items: list[str] | None) -> list[str]:
       return [s.strip() for s in (items or []) if s and s.strip()]
   ```

   Then change the three array lines inside `clean`'s `TripFacts(...)` to call it, one per array:

   ```python
   wildlife=clean_list(f.wildlife),
   ```

3. Count failed verdicts; print `-- what we would store (nothing rejected this run) --` when 0, else `-- what we would store (N fields coerced to null) --` (`1 field`, not `1 fields`).

   Copied from `complete/main.py`. Put it under the validator printout loop, back at the left edge. `sum(1 for v in verdicts if not v.passed)` counts the verdicts that failed:

   ```python
   rejected = sum(1 for v in verdicts if not v.passed)
   print()
   print("-- what we would store (nothing rejected this run) --" if rejected == 0
         else f"-- what we would store ({rejected} {'field' if rejected == 1 else 'fields'} coerced to null) --")
   ```

4. Print the cleaned record with the same eight-line printout from step 1.

   Copied from `complete/main.py`, directly under the heading lines above:

   ```python
   show(clean(raw, verdicts))
   print()
   ```

5. Loop over a list of report paths, defaulting to both `tr-0007.md` and `tr-0011.md` when none is given. Run with no arguments.

   `complete/main.py` has `clean`, and its loop reads `sys.argv[1:]`, defaulting to both data files.

   Replace the starter's `report_path = ...` line with a list of paths and a `for` loop. `sys.argv[1:]` is every argument after `main.py`; an empty list counts as false, so `or` falls back to the two data files:

   ```python
   report_paths = [Path(a) for a in sys.argv[1:]] or [DATA / "tr-0007.md", DATA / "tr-0011.md"]

   for report_path in report_paths:
       report = strip_front_matter(report_path.read_text())
   ```

   Python has no braces: the loop body is every line indented under the `for`. Indent everything from `response = client.chat.completions.parse(` down to the last `print()` by four more spaces. The lines of the prompt inside `f"""..."""` that start at the left edge stay there.

   ```bash
   uv run main.py
   ```

**Check:** both reports print three blocks each: what the model gave us, what the validator says, what we would store. A `0` in the "what we would store" block means the range rule is missing. A `""` there means the non-empty rule is missing. `distance: null mi` and `elev gain: null ft` under a run that rejected two zeros is the passing output; it matches "Run 4" in [`expected-output.md`](../expected-output.md). Reject means coerce to `null`, not throw: a gap is something a human can fill later, and a plausible wrong number is something nobody ever notices.

### Stretch goals

Pick any. The first two are not in `complete/`, so the checks are your only answer key.

- **Per-field confidence.** Add a second object named `"confidence"` to the schema, next to the eight fields, with the same eight keys, each a number 0-1. Print it next to each field. **Check:** `tr-0011.md`'s confidences come back lower than `tr-0007.md`'s. Every confidence at `1` on tr-0011 means the model is ignoring the field.

  A second model becomes a nested object in the schema. Put the new class above `class TripFacts`, and add one more field at the end of `TripFacts`:

  ```python
  # Hint: one float per field, each with its own description
  class FieldConfidence(BaseModel):
      trail_name: float = Field(description="How sure you are of trail_name, from 0 to 1.")
      ...  # the other seven fields, same pattern

  # new last field of TripFacts, under hazards:
      confidence: FieldConfidence = Field(description="A confidence from 0 to 1 for each field above.")

  # in show:
      print(f"  trail:      {f.trail_name or 'null'}  ({f.confidence.trail_name})")

  # in clean, TripFacts(...) needs the new field too, or pydantic raises a ValidationError:
          confidence=f.confidence,
  ```

- **One record per trail.** Make the top level of the schema an array of the same record. Feed it a report that covers two trails. **Check:** the array version returns one record per trail, and each record's `trail_name` grounds in the report on its own.

  `response_format` must be a model whose top level is an object, so wrap the array in one model with a single list field, put it under `class TripFacts`, and loop over that list. For a two-trail report, paste two reports into one file (for example `tr-0007.md` followed by feature 01's `tr-0002.md`) and pass its path:

  ```python
  # Hint: a model holding a list of TripFacts
  class TripList(BaseModel):
      trips: list[TripFacts] = Field(description="One record per trail hiked in the report.")

  response = client.chat.completions.parse(
      model="llama3.2",
      messages=[...],  # same prompt, saying there may be several trails
      response_format=TripList,
  )
  for raw in response.choices[0].message.parsed.trips:
      ...  # print, validate, clean each record as before
  ```

- **Any report.** Point the program at one of the other reports in feature 01's `data/`, for example `tr-0002.md`. **Check:** every scalar prints a `PASS` or a `REJECT` with a reason, and the validator never throws on a report it has not seen.

  No code to write. From `starter/`:

  ```bash
  uv run main.py ../../../F01-summarization/data/tr-0002.md
  ```

- **What the validator misses.** Run `tr-0011.md` until a run returns a made-up number that is still under the ceiling. `expected-output.md` records one such run: `distance_mi: 40`. **Check:** the range rule passes it. No cheap check can ground a number the way substring matching grounds a name, so what remains is sampling and review, which is a management answer, not a code answer.

  No code to write. Run this until a distance or elevation passes that the report never states:

  ```bash
  uv run main.py ../../data/tr-0011.md
  ```

## What Is in This Folder

- `data/tr-0007.md`: the fact-rich report (a note-taker hikes Sperry Chalet on July 4, 2026; 12.8 miles, 3,400 feet, bears, goats, and pie, all stated outright). Copied from feature 01's forty-report corpus, front matter and all.
- `data/tr-0011.md`: the fact-sparse report (a Yosemite trip written up a month late; no trail name, no date, no mileage, no elevation, because the little book is in the car at the mechanic's).
- `expected-output.md`: real `llama3.2` outputs for both reports and for the finished demo, the real rejections from four consecutive runs, the success checks, and an honest list of where the model still slips
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
