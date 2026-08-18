// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete                          extract both data/ reports, then validate
//   npm run complete -- path1.md [path2.md]   extract any report(s) instead
// The schema is the zod object below. Nullable fields plus the .describe()
// text ("null if not stated") are what keep the sparse report honest. The
// validator underneath is what catches the times they don't.

import { readFileSync } from "node:fs";
import { basename, resolve } from "node:path";
import OpenAI from "openai";
import { zodResponseFormat } from "openai/helpers/zod";
import { z } from "zod";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const MODEL = "llama3.2";
const DATA = resolve(import.meta.dirname, "../../data");

// The schema does the prompting. Every scalar is nullable and every description
// says when to use null, which is the half of the hallucination fix that a plea
// in the prompt cannot do; the validator below is the other half. Making a field
// non-nullable here forces the model to invent a value for it.
const TripFacts = z.object({
  trail_name: z.string().nullable().describe("The name of the trail hiked. null if the report never names the trail."),
  park: z.string().nullable().describe("The park the trail is in. null if the report never names the park."),
  date_hiked: z.string().nullable().describe("The date of the hike in YYYY-MM-DD format. null if the report does not give an exact date. Never guess or infer a date."),
  distance_mi: z.number().nullable().describe("Round-trip distance in miles, as stated in the report. null, never 0, if the report gives no distance. Never estimate."),
  elevation_gain_ft: z.number().nullable().describe("Elevation gain in feet, as stated in the report. null, never 0, if the report gives no elevation figure. Never estimate."),
  wildlife: z.array(z.string()).describe("Animals the author actually saw on this hike. Empty array if none are mentioned."),
  conditions: z.array(z.string()).describe("Short phrases describing trail conditions the report mentions (mud, snow, water crossings, dry tread). Empty array if none."),
  hazards: z.array(z.string()).describe("Hazards or closures the report mentions. Empty array if none."),
});
type TripFacts = z.infer<typeof TripFacts>;

type Verdict = {
  field: string;
  value: string | null;
  passed: boolean;
  reason?: string;
  // Set when a field passed only after being rewritten, e.g. a date the model
  // wrote as "July 4, 2026" that we store as "2026-07-04".
  normalized?: string;
};
const pass = (field: string, value: string | null, normalized?: string): Verdict => ({ field, value, passed: true, normalized });
const fail = (field: string, value: string | null, reason: string): Verdict => ({ field, value, passed: false, reason });

function stripFrontMatter(markdown: string): string {
  const parts = markdown.split("---");
  return parts.length >= 3 ? parts.slice(2).join("---").trim() : markdown.trim();
}

function show(f: TripFacts): void {
  console.log(`  trail:      ${f.trail_name ?? "null"}`);
  console.log(`  park:       ${f.park ?? "null"}`);
  console.log(`  date:       ${f.date_hiked ?? "null"}`);
  console.log(`  distance:   ${f.distance_mi ?? "null"} mi`);
  console.log(`  elev gain:  ${f.elevation_gain_ft ?? "null"} ft`);
  console.log(`  wildlife:   [${(f.wildlife ?? []).join(", ")}]`);
  console.log(`  conditions: [${(f.conditions ?? []).join(", ")}]`);
  console.log(`  hazards:    [${(f.hazards ?? []).join(", ")}]`);
}

function nonEmpty(field: string, value: string | null): Verdict {
  if (value === null) return pass(field, null);
  return value.trim().length === 0
    ? fail(field, `"${value}"`, "empty or whitespace-only string; should be null")
    : pass(field, value);
}

// Cheap grounding check: a name whose distinctive words never appear in the
// report is a name the model supplied. Deliberately crude, because a check worth
// having costs a dozen lines rather than a research project. The boilerplate list
// is what lets "Glacier National Park" ground on a report that only says
// "Glacier"; shortening it will start rejecting correct park names.
const BOILERPLATE = new Set(["national", "park", "state", "trail", "trailhead", "loop", "canyon", "falls", "the"]);

function grounded(field: string, value: string | null, source: string): Verdict {
  const basic = nonEmpty(field, value);
  if (!basic.passed || value === null) return basic;
  const lower = source.toLowerCase();
  const words = value.split(/[ \-,']+/).filter((w) => w.length >= 4 && !BOILERPLATE.has(w.toLowerCase()));
  // Nothing distinctive to check: fall back to the whole string.
  const missing = words.length === 0
    ? (lower.includes(value.trim().toLowerCase()) ? [] : [value.trim()])
    : words.filter((w) => !lower.includes(w.toLowerCase()));
  return missing.length === 0
    ? pass(field, value)
    : fail(field, value, `not grounded in the source report (no mention of ${missing.map((m) => `"${m}"`).join(", ")})`);
}

// Explicit formats only. A real date in an odd format gets normalized;
// "last month (exact date not specified)" is prose, and prose in a date column
// is a bug waiting for a reporting query.
const MONTHS = ["january", "february", "march", "april", "may", "june", "july", "august", "september", "october", "november", "december"];
function parseDate(text: string): { y: number; m: number; d: number } | null {
  const s = text.trim();
  let m: RegExpMatchArray | null;
  if ((m = s.match(/^(\d{4})[-/](\d{1,2})[-/](\d{1,2})$/))) return { y: +m[1], m: +m[2], d: +m[3] };
  if ((m = s.match(/^(\d{1,2})\/(\d{1,2})\/(\d{4})$/))) return { y: +m[3], m: +m[1], d: +m[2] };
  if ((m = s.match(/^([A-Za-z]+)\.? (\d{1,2}),? (\d{4})$/)) || (m = s.match(/^(\d{1,2}) ([A-Za-z]+) (\d{4})$/))) {
    const [name, day, year] = /^\d/.test(s) ? [m[2], m[1], m[3]] : [m[1], m[2], m[3]];
    const idx = MONTHS.findIndex((mo) => mo.startsWith(name.toLowerCase().slice(0, 3)));
    if (idx >= 0) return { y: +year, m: idx + 1, d: +day };
  }
  return null;
}

function validDate(field: string, value: string | null): Verdict {
  const basic = nonEmpty(field, value);
  if (!basic.passed || value === null) return basic;
  const parsed = parseDate(value);
  if (!parsed || parsed.m < 1 || parsed.m > 12 || parsed.d < 1 || parsed.d > 31) {
    return fail(field, value, "does not parse as a date; no parser can store this");
  }
  if (parsed.y < 1900 || parsed.y > 2100) {
    return fail(field, value, `parses, but the year ${parsed.y} is not plausible`);
  }
  // Parseable but off-format gets normalized on the way to storage.
  const canonical = `${parsed.y}-${String(parsed.m).padStart(2, "0")}-${String(parsed.d).padStart(2, "0")}`;
  return pass(field, value, canonical);
}

function inRange(field: string, value: number | null, max: number, unit: string): Verdict {
  if (value === null) return pass(field, null);
  const shown = `${value} ${unit}`;
  if (value === 0) return fail(field, shown, "0 is not a measurement; the report gave no figure, so this should be null");
  if (value < 0) return fail(field, shown, "negative value is impossible");
  return value > max ? fail(field, shown, `implausible: over ${max} ${unit} for a single day hike`) : pass(field, shown);
}

// The rejection rules. Ordinary code, no model involved. Each rule covers output
// llama3.2 has actually returned for these two reports, which is why none of them
// look defensive until you see the run that needs them (../../expected-output.md).
function validate(f: TripFacts, source: string): Verdict[] {
  return [
    // A name the source text never contains does not get to be a fact, no
    // matter how plausible it reads.
    grounded("trail_name", f.trail_name, source),
    grounded("park", f.park, source),
    validDate("date_hiked", f.date_hiked),
    // 0 is the dangerous near-miss: it is a value, so a pipeline stores it and
    // nothing downstream ever questions it. The honest answer is null.
    inRange("distance_mi", f.distance_mi, 100, "mi"),
    inRange("elevation_gain_ft", f.elevation_gain_ft, 20000, "ft"),
  ];
}

const cleanList = (items: string[] | null | undefined) => (items ?? []).filter((s) => s && s.trim()).map((s) => s.trim());

// Rejected fields become null. Storing nothing beats storing a plausible lie:
// null is a gap a human can fill, 0 is a number nobody will ever question.
function clean(f: TripFacts, verdicts: Verdict[]): TripFacts {
  const v = (name: string) => verdicts.find((x) => x.field === name)!;
  const ok = (name: string) => v(name).passed;
  return {
    trail_name: ok("trail_name") ? f.trail_name?.trim() ?? null : null,
    park: ok("park") ? f.park?.trim() ?? null : null,
    date_hiked: ok("date_hiked") ? v("date_hiked").normalized ?? f.date_hiked?.trim() ?? null : null,
    distance_mi: ok("distance_mi") ? f.distance_mi : null,
    elevation_gain_ft: ok("elevation_gain_ft") ? f.elevation_gain_ft : null,
    // The arrays get no rule of their own: substring matching cannot ground a
    // free-text condition the way it grounds a name, so they are only cleaned.
    wildlife: cleanList(f.wildlife),
    conditions: cleanList(f.conditions),
    hazards: cleanList(f.hazards),
  };
}

const reportPaths = process.argv.length > 2
  ? process.argv.slice(2)
  : [resolve(DATA, "tr-0007.md"), resolve(DATA, "tr-0011.md")];

for (const reportPath of reportPaths) {
  const report = stripFrontMatter(readFileSync(reportPath, "utf8"));

  // Prose in, typed object out. Nothing here strips a markdown fence or a
  // preamble, because .parse() with a zod schema makes the shape the model's
  // problem: the schema goes up as response_format, and the reply comes back
  // already validated against it.
  const response = await client.chat.completions.parse({
    model: MODEL,
    messages: [{
      role: "user",
      content: `Extract the trail facts from this trip report.
Use null for any field the report does not state, and empty arrays
when nothing applies. Do not guess.

${report}`,
    }],
    response_format: zodResponseFormat(TripFacts, "trip_facts"),
  });
  const raw = response.choices[0].message.parsed!;

  console.log(`== ${basename(reportPath)} ==\n`);
  console.log("-- what the model gave us --");
  show(raw);

  // The schema guarantees the JSON parses, not that it is true. Every scalar
  // goes through a rule here, and anything that fails is coerced to null before
  // it could reach a database. Do not shortcut this to store `raw` directly.
  const verdicts = validate(raw, report);

  console.log("\n-- what the validator says --");
  for (const v of verdicts) {
    console.log(`  ${v.passed ? "PASS  " : "REJECT"}  ${v.field.padEnd(18)} ${v.value ?? "null"}`);
    if (!v.passed) console.log(`          reason: ${v.reason}`);
    else if (v.normalized && v.normalized !== v.value) console.log(`          normalized to: ${v.normalized}`);
  }

  const rejected = verdicts.filter((v) => !v.passed).length;
  console.log();
  console.log(rejected === 0
    ? "-- what we would store (nothing rejected this run) --"
    : `-- what we would store (${rejected} ${rejected === 1 ? "field" : "fields"} coerced to null) --`);
  show(clean(raw, verdicts));
  console.log();
}
