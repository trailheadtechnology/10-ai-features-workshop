// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete                          "more like this" for Avalanche Lake Trail
//   npm run complete -- trail-0008            any trail id works
//   npm run complete -- Trail of the Cedars   so does any name (or part of one)
//   npm run complete -- --gear Cascade 65     the same trick on gear, from review text
//
// Vectors are cached in embeddings.json / gear-embeddings.json next to this
// script. The cache is only checked for missing keys, so an edited description
// or a different embedding model leaves the stale vectors in place. Delete the
// cache file whenever the source text or the model changes.

import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const EMBED_MODEL = "nomic-embed-text";
const DATA = resolve(import.meta.dirname, "../../data");
const HERE = import.meta.dirname;

type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };
type Review = { id: string; product: string; rating: number; reviewer: string; text: string };

function cosine(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}

// Embed each text once and cache the vectors to disk. Note the staleness trap:
// the cache is accepted whenever it holds every key, so changed text under an
// existing key keeps its old vector. Delete the file to force a re-embed.
async function embedWithCache(cacheName: string, texts: Record<string, string>): Promise<Record<string, number[]>> {
  const cachePath = resolve(HERE, cacheName);
  if (existsSync(cachePath)) {
    const cached: Record<string, number[]> = JSON.parse(readFileSync(cachePath, "utf8"));
    if (Object.keys(texts).every((k) => k in cached)) return cached;
  }
  const keys = Object.keys(texts);
  const response = await client.embeddings.create({ model: EMBED_MODEL, input: keys.map((k) => texts[k]) });
  const vectors = Object.fromEntries(keys.map((k, i) => [k, response.data[i].embedding]));
  writeFileSync(cachePath, JSON.stringify(vectors));
  return vectors;
}

// Products have no descriptions, so each vector comes from that product's reviews
// concatenated: "similar" here means "reviewers describe them the same way".
//
// Content similarity finds substitutes, not complements. The nearest neighbor to
// a backpack is usually another size of the same backpack, which is the one item
// its owner will never buy. Complements come from behavior data (what people buy
// or mention together), and no embedding of the product text can supply it.
async function recommendGear(query: string): Promise<void> {
  const reviews: Review[] = readFileSync(resolve(DATA, "gear-reviews.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
  const texts: Record<string, string> = {};
  for (const r of reviews) texts[r.product] = texts[r.product] ? `${texts[r.product]}\n${r.text}` : r.text;

  const vectors = await embedWithCache("gear-embeddings.json", texts);

  const target = Object.keys(texts).find((p) => p.toLowerCase().includes(query.toLowerCase()));
  if (!target) throw new Error(`No product matches '${query}'.`);

  console.log(`You bought: ${target}`);
  console.log("Goes well with:\n");
  const hits = Object.entries(vectors)
    .filter(([p]) => p !== target)
    .map(([p, v]) => ({ product: p, score: cosine(vectors[target], v) }))
    .sort((a, b) => b.score - a.score)
    .slice(0, 5);
  for (const { product, score } of hits) console.log(`  ${score.toFixed(4)}  ${product}`);
}

const args = process.argv.slice(2);
if (args[0] === "--gear") {
  await recommendGear(args.slice(1).join(" "));
  process.exit(0);
}

// Same catalog and same embedding model as feature 04. Recommendations need no
// new model and no new data, only the vectors search already produced.
const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails.json"), "utf8"));
const vectors = await embedWithCache("embeddings.json", Object.fromEntries(trails.map((t) => [t.id, t.description])));

const query = args.join(" ") || "trail-0117";
const target = trails.find((t) => t.id.toLowerCase() === query.toLowerCase() || t.name.toLowerCase().includes(query.toLowerCase()));
if (!target) throw new Error(`No trail matches '${query}'.`);

// "More like this" is the search from feature 04 with the query vector replaced
// by an item's own vector. That means it ranks on what the descriptions talk
// about, so whatever the prose leaves out is invisible: difficulty and distance
// live in structured fields and are never mentioned in the text, and a moderate
// family hike will cheerfully return a list of hard all-day climbs. If those
// fields matter to the user, filter or re-rank on them after the similarity pass.
console.log(`You liked: ${target.name} (${target.park})`);
console.log("You might also like:\n");

const hits = trails
  .filter((t) => t.id !== target.id)
  .map((t) => ({ trail: t, score: cosine(vectors[target.id], vectors[t.id]) }))
  .sort((a, b) => b.score - a.score)
  .slice(0, 5);
for (const { trail, score } of hits) {
  console.log(`  ${score.toFixed(4)}  ${trail.name} (${trail.park}, ${trail.difficulty}; ${trail.features.join(", ")})`);
}
