// Starting point: ranks trail reports by distance from the trail's baseline.
// Run: npm run starter
//
// Makes no model calls and needs no network. The vectors in
// ../../data/embeddings-0117.json were precomputed with nomic-embed-text so this
// runs with Ollama down, which is the fallback when the room's network is not
// cooperating. Those vectors were embedded with the "classification: " task
// prefix nomic requires, so anything you add to this corpus must be embedded the
// same way or its distances will not be comparable to these.
//
// complete/ embeds live and adds the alert rule on top of this ranking.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const DATA = resolve(import.meta.dirname, "../../data");
type Report = { id: string; trail_id: string; date: string; text: string };

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

const reports: Report[] = readFileSync(resolve(DATA, "reports-0117.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
const raw: Record<string, number[]> = JSON.parse(readFileSync(resolve(DATA, "embeddings-0117.json"), "utf8")).embeddings;
const vectors = Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, normalize(v)]));

// The centroid is this trail's definition of normal, and it is built from the
// same reports it is about to judge. Anomalies that appear often enough pull the
// centroid toward themselves and stop looking anomalous.
const all = Object.values(vectors);
const dimensions = all[0].length;
const centroid = new Array<number>(dimensions).fill(0);
for (const vector of all) for (let i = 0; i < dimensions; i++) centroid[i] += vector[i] / all.length;
const center = normalize(centroid);

const scored = reports
  .map((r) => ({ report: r, distance: cosineDistance(vectors[r.id], center) }))
  .sort((a, b) => b.distance - a.distance);

console.log(`trail-0117 · ${reports.length} reports · ${dimensions}-dim nomic-embed-text vectors\n`);
console.log("  dist    id       date        report");
for (const { report, distance } of scored) {
  console.log(`  ${distance.toFixed(4)}  ${report.id}  ${report.date}  ${truncate(report.text, 62)}`);
}
