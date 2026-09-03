"""Finished demo, matching the demo script in docs/slides/outlines:
  uv run main.py                        extract both data/ reports, then validate
  uv run main.py path1.md [path2.md]    extract any report(s) instead
The schema is the pydantic model at the bottom. Optional fields plus the
Field descriptions ("null if not stated") are what keep the sparse report
honest. The validator underneath is what catches the times they don't.
"""

import sys
from dataclasses import dataclass
from datetime import date, datetime
from pathlib import Path

from openai import OpenAI
from pydantic import BaseModel, Field

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
MODEL = "llama3.2"
DATA = Path(__file__).resolve().parents[2] / "data"


# The schema does the prompting. Every scalar is optional and every description
# says when to use null, which is the half of the hallucination fix that a plea
# in the prompt cannot do; the validator below is the other half. Making a field
# required here forces the model to invent a value for it.
class TripFacts(BaseModel):
    trail_name: str | None = Field(description="The name of the trail hiked. null if the report never names the trail.")
    park: str | None = Field(description="The park the trail is in. null if the report never names the park.")
    date_hiked: str | None = Field(description="The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date.")
    distance_mi: float | None = Field(description="Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate.")
    elevation_gain_ft: float | None = Field(description="Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate.")
    wildlife: list[str] = Field(description="Animals the author actually saw on this hike. Empty array if none are mentioned.")
    conditions: list[str] = Field(description="Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none.")
    hazards: list[str] = Field(description="Hazards or closures the report mentions. Empty array if none.")


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


def strip_front_matter(markdown: str) -> str:
    parts = markdown.split("---", 2)
    return parts[2].strip() if len(parts) == 3 else markdown.strip()


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


def non_empty(field: str, value: str | None) -> Verdict:
    if value is None:
        return Verdict.ok(field, None)
    if not value.strip():
        return Verdict.fail(field, f'"{value}"', "empty or whitespace-only string; should be null")
    return Verdict.ok(field, value)


# Cheap grounding check: a name whose distinctive words never appear in the
# report is a name the model supplied. Deliberately crude, because a check worth
# having costs a dozen lines rather than a research project. The boilerplate list
# is what lets "Glacier National Park" ground on a report that only says
# "Glacier"; shortening it will start rejecting correct park names.
BOILERPLATE = {"national", "park", "state", "trail", "trailhead", "loop", "canyon", "falls", "the"}


def grounded(field: str, value: str | None, source: str) -> Verdict:
    basic = non_empty(field, value)
    if not basic.passed or value is None:
        return basic
    lower = source.lower()
    words = [w for w in value.replace("-", " ").replace(",", " ").replace("'", " ").split()
             if len(w) >= 4 and w.lower() not in BOILERPLATE]
    if not words:
        # Nothing distinctive to check: fall back to the whole string.
        missing = [] if value.strip().lower() in lower else [value.strip()]
    else:
        missing = [w for w in words if w.lower() not in lower]
    if not missing:
        return Verdict.ok(field, value)
    return Verdict.fail(field, value, "not grounded in the source report (no mention of " + ", ".join(f'"{m}"' for m in missing) + ")")


# Explicit formats only. A real date in an odd format gets normalized;
# "last month (exact date not specified)" is prose, and prose in a date column
# is a bug waiting for a reporting query.
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


# The rejection rules. Ordinary code, no model involved. Each rule covers output
# llama3.2 has actually returned for these two reports, which is why none of them
# look defensive until you see the run that needs them (../../expected-output.md).
def validate(f: TripFacts, source: str) -> list[Verdict]:
    return [
        # A name the source text never contains does not get to be a fact, no
        # matter how plausible it reads.
        grounded("trail_name", f.trail_name, source),
        grounded("park", f.park, source),
        valid_date("date_hiked", f.date_hiked),
        # 0 is the dangerous near-miss: it is a value, so a pipeline stores it and
        # nothing downstream ever questions it. The honest answer is null.
        in_range("distance_mi", f.distance_mi, 100, "mi"),
        in_range("elevation_gain_ft", f.elevation_gain_ft, 20000, "ft"),
    ]


def clean_list(items: list[str] | None) -> list[str]:
    return [s.strip() for s in (items or []) if s and s.strip()]


# Rejected fields become null. Storing nothing beats storing a plausible lie:
# null is a gap a human can fill, 0 is a number nobody will ever question.
def clean(f: TripFacts, verdicts: list[Verdict]) -> TripFacts:
    v = {x.field: x for x in verdicts}
    ok = lambda name: v[name].passed
    return TripFacts(
        trail_name=f.trail_name.strip() if ok("trail_name") and f.trail_name else None,
        park=f.park.strip() if ok("park") and f.park else None,
        date_hiked=(v["date_hiked"].normalized or (f.date_hiked or "").strip() or None) if ok("date_hiked") else None,
        distance_mi=f.distance_mi if ok("distance_mi") else None,
        elevation_gain_ft=f.elevation_gain_ft if ok("elevation_gain_ft") else None,
        # The arrays get no rule of their own: substring matching cannot ground a
        # free-text condition the way it grounds a name, so they are only cleaned.
        wildlife=clean_list(f.wildlife),
        conditions=clean_list(f.conditions),
        hazards=clean_list(f.hazards),
    )


report_paths = [Path(a) for a in sys.argv[1:]] or [DATA / "tr-0007.md", DATA / "tr-0011.md"]

for report_path in report_paths:
    report = strip_front_matter(report_path.read_text())

    # Prose in, populated object out. Nothing here strips a markdown fence or a
    # preamble, because .parse() with a pydantic model makes the shape the
    # model's problem: the schema goes up as response_format, and the reply comes
    # back already validated against it.
    response = client.chat.completions.parse(
        model=MODEL,
        messages=[{"role": "user", "content": f"""Extract the trail facts from this trip report.
Use null for any field the report does not state, and empty arrays
when nothing applies. Do not guess.

{report}"""}],
        response_format=TripFacts,
    )
    raw = response.choices[0].message.parsed
    assert raw is not None

    print(f"== {report_path.name} ==\n")
    print("-- what the model gave us --")
    show(raw)

    # The schema guarantees the JSON parses, not that it is true. Every scalar
    # goes through a rule here, and anything that fails is coerced to null before
    # it could reach a database. Do not shortcut this to store `raw` directly.
    verdicts = validate(raw, report)

    print("\n-- what the validator says --")
    for v in verdicts:
        mark = "PASS  " if v.passed else "REJECT"
        print(f"  {mark}  {v.field:<18} {v.value if v.value is not None else 'null'}")
        if not v.passed:
            print(f"          reason: {v.reason}")
        elif v.normalized is not None and v.normalized != v.value:
            print(f"          normalized to: {v.normalized}")

    rejected = sum(1 for v in verdicts if not v.passed)
    print()
    print("-- what we would store (nothing rejected this run) --" if rejected == 0
          else f"-- what we would store ({rejected} {'field' if rejected == 1 else 'fields'} coerced to null) --")
    show(clean(raw, verdicts))
    print()
