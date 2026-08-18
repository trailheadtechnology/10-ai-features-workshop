// Demo starting point: keyword search over the trail slice. Split the query into
// words and count how many appear, as whole words, in each trail's name and
// description.
// Run: npm run starter -- <query>          (defaults to the demo query)
//
// Matching on words alone has no access to meaning. A description saying the
// trail avoids the steep section still matches "steep", and a trail whose text
// reads "leashed dogs are welcome" never matches "dog-friendly". That gap is
// what the complete script replaces with embeddings.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";

const DATA = resolve(import.meta.dirname, "../../data");

type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };

const query = process.argv.slice(2).join(" ") || "dog-friendly waterfall hike, not too steep";
const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails-slice.json"), "utf8"));

const tokens = [...new Set((query.toLowerCase().match(/[a-z]+/g) ?? []).filter((w) => w.length >= 3))];

const results = trails
  .map((t) => {
    const haystack = `${t.name} ${t.description}`.toLowerCase();
    const hits = tokens.filter((w) => new RegExp(`\\b${w}\\b`).test(haystack));
    return { trail: t, hits };
  })
  .filter((r) => r.hits.length > 0)
  .sort((a, b) => b.hits.length - a.hits.length || a.trail.id.localeCompare(b.trail.id))
  .slice(0, 5);

console.log(`Keyword search: "${query}"`);
console.log(`Query words: ${tokens.join(", ")}\n`);

if (results.length === 0) {
  console.log("No results. Not one trail contains those words.");
  process.exit(0);
}

for (const { trail, hits } of results) {
  console.log(`${hits.length} word(s) [${hits.join(", ")}]  ${trail.id}  ${trail.name} (${trail.difficulty}, ${trail.distance_mi} mi)`);
}
