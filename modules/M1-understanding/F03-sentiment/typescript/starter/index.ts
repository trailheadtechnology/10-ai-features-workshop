// Demo starting point: one chat client, one classify function, one review.
// Run: npm run starter [-- review-id]
// Ids come from ../../data/easy.jsonl and ../../data/hard.jsonl. The default is
// gr-0007, a hard-set review whose sarcasm ("five-star experience, truly") points
// the opposite way from its two-star rating.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const DATA = resolve(import.meta.dirname, "../../data");

type Review = { id: string; product: string; rating: number; reviewer: string; text: string };
const lines = (name: string) => readFileSync(resolve(DATA, name), "utf8").split("\n").filter((l) => l.trim());
const reviews: Review[] = [...lines("easy.jsonl"), ...lines("hard.jsonl")].map((l) => JSON.parse(l));

const wanted = process.argv[2] ?? "gr-0007";
const review = reviews.find((r) => r.id === wanted)!;

console.log(`${review.product} (${review.rating} stars), reviewed by ${review.reviewer}`);
console.log(review.text);
console.log();

// The whole feature is this function. The prompt carries it; the model is
// swappable because everything upstream only sees the client and a model name.
async function classify(client: OpenAI, model: string, text: string): Promise<string> {
  // Keep this prompt byte-identical to the one in ../../http/ollama.http, line breaks
  // included. Reflowing these four lines into one costs phi3 measured accuracy
  // on both sets, so a comparison run against a reflowed prompt is not
  // comparing models. See ../../expected-output.md.
  const prompt = `Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: ${text}`;
  const response = await client.chat.completions.create({ model, messages: [{ role: "user", content: prompt }], temperature: 0 });
  const raw = (response.choices[0].message.content ?? "").toLowerCase();
  // Small models sometimes wrap the label in a sentence; keep the first label mentioned.
  const found = ["positive", "negative", "mixed"]
    .map((label) => ({ label, at: raw.indexOf(label) }))
    .filter((x) => x.at >= 0)
    .sort((a, b) => a.at - b.at);
  return found[0]?.label ?? raw.trim();
}

console.log(`phi3 says: ${await classify(client, "phi3", review.text)}`);
