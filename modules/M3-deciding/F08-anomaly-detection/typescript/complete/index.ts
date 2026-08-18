// Ranks trail reports by distance from the trail's own baseline and alerts when
// several outliers land close together in time. Needs Ollama running with
// nomic-embed-text pulled.
//
//   npm run complete                    trail-0117, live embeddings, distance table + cluster alerts
//   npm run complete -- --raw           same trail with the task prefix removed
//   npm run complete -- --trail 0042    the other trail in the data folder
//   npm run complete -- --sigma 1.5     tighter threshold
//   npm run complete -- --window 30     wider clustering window, in days
//
// Embeddings are the only model calls. Everything after them is arithmetic, so
// the cost of this feature is one embedding per report and nothing per query.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const DATA = resolve(import.meta.dirname, "../../data");
type Report = { id: string; trail_id: string; date: string; text: string };

let trail = "0117";
let sigma = 1.0;
let window = 14;
// nomic-embed-text is trained with task prefixes (search_query:, search_document:,
// clustering:, classification:) and expects one on every input. Embedding bare text
// still returns a well-formed vector, which is why --raw fails silently rather than
// throwing, but the vectors land off-distribution and the ranking degrades badly.
// Do not drop this prefix, and if you change it, change it for every input in the
// same corpus: vectors embedded under different prefixes are not comparable.
let prefix = "classification: ";
const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case "--trail": trail = args[++i]; break;
    case "--sigma": sigma = parseFloat(args[++i]); break;
    case "--window": window = parseInt(args[++i], 10); break;
    case "--raw": prefix = ""; break;
    default: throw new Error(`unknown argument: ${args[i]}`);
  }
}

function normalize(vector: number[]): number[] {
  const length = Math.sqrt(vector.reduce((s, v) => s + v * v, 0));
  return vector.map((v) => v / length);
}

// Correct only for unit-length inputs, where the dot product is already the cosine
// similarity. Every vector reaching this function has been through normalize; pass an
// unnormalized one and it returns a number that still looks plausible.
function cosineDistance(a: number[], b: number[]): number {
  let dot = 0;
  for (let i = 0; i < a.length; i++) dot += a[i] * b[i];
  return 1 - dot;
}

const truncate = (text: string, n: number) => (text.length <= n ? text : text.slice(0, n - 1) + "…");
const days = (a: string, b: string) => Math.round((Date.parse(a) - Date.parse(b)) / 86_400_000);

const reports: Report[] = readFileSync(resolve(DATA, `reports-${trail}.jsonl`), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));

const response = await client.embeddings.create({ model: "nomic-embed-text", input: reports.map((r) => prefix + r.text) });
const vectors = response.data.map((d) => normalize(d.embedding));

// The centroid is this trail's definition of normal, and it is built from the
// same reports it is about to judge. Anomalies that appear often enough pull the
// centroid toward themselves and stop looking anomalous, so a long-running
// detector should rebuild this from a trailing window rather than the full history.
const dimensions = vectors[0].length;
const centroid = new Array<number>(dimensions).fill(0);
for (const vector of vectors) for (let i = 0; i < dimensions; i++) centroid[i] += vector[i] / vectors.length;
const center = normalize(centroid);

const scored = reports
  .map((r, i) => ({ report: r, distance: cosineDistance(vectors[i], center) }))
  .sort((a, b) => b.distance - a.distance);

// The threshold is derived from this corpus rather than hard-coded, so it travels
// to a trail with a different spread of distances. It is still a business choice,
// not a boundary the data hands you: sigma decides how much review you are willing
// to pay for, and there is no value of it that separates incidents from oddities.
const mean = scored.reduce((s, x) => s + x.distance, 0) / scored.length;
const deviation = Math.sqrt(scored.reduce((s, x) => s + (x.distance - mean) ** 2, 0) / scored.length);
const threshold = mean + sigma * deviation;

console.log(`trail-${trail} · ${reports.length} reports · nomic-embed-text (${dimensions} dims)`
  + (prefix.length === 0 ? " · NO task prefix" : ` · prefix "${prefix.trim()}"`));
console.log(`mean distance ${mean.toFixed(4)} · sd ${deviation.toFixed(4)} · threshold mean+${sigma}sd = ${threshold.toFixed(4)}\n`);

console.log("  dist    id       date        report");
for (const s of scored) {
  console.log(`${s.distance > threshold ? " !" : "  "}${s.distance.toFixed(4)}  ${s.report.id}  ${s.report.date}  ${truncate(s.report.text, 62)}`);
}

// Corroboration is what makes this alertable. A single report far from normal is
// usually just an unusual subject, not an incident; two or more inside the window
// mean several people independently noticed the same thing. Requiring two is what
// keeps the alert queue small enough that someone still reads it.
const flagged = scored.filter((s) => s.distance > threshold).sort((a, b) => a.report.date.localeCompare(b.report.date));
console.log(`\n${flagged.length} of ${scored.length} reports above threshold. Clustering them within ${window} days:\n`);

let alerts = 0;
for (let i = 0; i < flagged.length;) {
  let j = i + 1;
  while (j < flagged.length && days(flagged[j].report.date, flagged[j - 1].report.date) <= window) j++;
  const group = flagged.slice(i, j);
  if (group.length >= 2) {
    alerts++;
    console.log(`  ALERT trail-${trail}: ${group.length} anomalous reports between ${group[0].report.date} and ${group[group.length - 1].report.date}`);
    for (const s of group) console.log(`        ${s.report.id} ${s.report.date}  ${truncate(s.report.text, 70)}`);
  } else {
    console.log(`  (ignored) ${group[0].report.id} ${group[0].report.date} is a lone outlier, not an event`);
  }
  i = j;
}
console.log(`\n${alerts} alert(s). Model calls: ${reports.length} embeddings, 0 chat completions.`);
