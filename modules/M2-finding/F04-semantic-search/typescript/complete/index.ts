// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete -- dog-friendly waterfall hike, not too steep
//   npm run complete -- somewhere quiet to take my kids
// Embeds every trail description once, embeds the query, ranks by cosine
// similarity, prints the top 5.

import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

// Ollama serves embeddings on the same OpenAI-compatible endpoint as chat.
const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const EMBED_MODEL = "nomic-embed-text";
const DATA = resolve(import.meta.dirname, "../../data");

type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };

const query = process.argv.slice(2).join(" ") || "dog-friendly waterfall hike, not too steep";
const trails: Trail[] = JSON.parse(readFileSync(resolve(DATA, "trails-slice.json"), "utf8"));

async function embed(texts: string[]): Promise<number[][]> {
  const response = await client.embeddings.create({ model: EMBED_MODEL, input: texts });
  return response.data.map((d) => d.embedding);
}

// Embedding the catalog takes seconds, so the vectors are cached next to this
// script. The cache is keyed by trail id and nothing else: if a description
// changes, or the embedding model changes, delete embeddings.json. Otherwise
// every later query is ranked against vectors for text that no longer exists.
const cachePath = resolve(import.meta.dirname, "embeddings.json");
let vectors: Record<string, number[]>;
if (existsSync(cachePath)) {
  vectors = JSON.parse(readFileSync(cachePath, "utf8"));
  console.log(`Loaded ${Object.keys(vectors).length} cached vectors from embeddings.json`);
} else {
  const started = performance.now();
  const embeddings = await embed(trails.map((t) => t.description));
  vectors = Object.fromEntries(trails.map((t, i) => [t.id, embeddings[i]]));
  writeFileSync(cachePath, JSON.stringify(vectors));
  console.log(`Embedded ${trails.length} trail descriptions in ${Math.round(performance.now() - started)} ms`);
}

// The query has to go through the same model that produced the cached vectors.
// Vectors from two different models are not comparable, and cosine similarity
// will still return confident-looking numbers if you mix them.
const [queryVector] = await embed([query]);

function cosineSimilarity(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}

// This ranks on topic, not on suitability. An embedding cannot tell "great for
// kids" from "dangerous for kids" or "easy" from "never uses the word easy but
// is a cliff", so a top result can be about the right subject and still be the
// worst possible recommendation. Read the absolute scores as well: a top 5
// bunched together at a low score means nothing in the catalog is a real match
// and the order is mostly noise. Anything you can express as a filter over the
// metadata you already have (difficulty, features) is cheaper and more reliable
// than hoping the vector carries it.
const results = trails
  .map((t) => ({ trail: t, score: cosineSimilarity(queryVector, vectors[t.id]) }))
  .sort((a, b) => b.score - a.score)
  .slice(0, 5);

console.log(`\nSemantic search: "${query}"\n`);
for (const { trail, score } of results) {
  console.log(`${score.toFixed(4)}  ${trail.id}  ${trail.name} (${trail.difficulty}, ${trail.distance_mi} mi)  [${trail.features.join(", ")}]`);
}
