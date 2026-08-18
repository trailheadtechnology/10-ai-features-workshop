// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete             both sets, both models, table + accuracy + disagreements
//   npm run complete -- --easy   easy set only (demo steps 3 and 4)
//   npm run complete -- --hard   hard set only (demo step 5)
//
// The big model is Azure OpenAI when AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_KEY,
// and AZURE_OPENAI_DEPLOYMENT are set. When they aren't, llama3.2 on Ollama
// stands in so the whole comparison runs offline. Either way, the swap is the
// few lines building `big` below; nothing downstream changes.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI, { AzureOpenAI } from "openai";

const DATA = resolve(import.meta.dirname, "../../data");
const args = process.argv.slice(2);
const sets = args.includes("--easy") ? ["easy"] : args.includes("--hard") ? ["hard"] : ["easy", "hard"];

type Target = { client: OpenAI; model: string };

// Model 1: the small local model. Free, private, 2GB.
const ollama = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const small: Target = { client: ollama, model: "phi3" };

// Model 2: the big model, or its local stand-in.
const { AZURE_OPENAI_ENDPOINT: endpoint, AZURE_OPENAI_KEY: key, AZURE_OPENAI_DEPLOYMENT: deployment } = process.env;
let big: Target;
let bigName: string;
if (endpoint && key && deployment) {
  big = { client: new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment }), model: deployment };
  bigName = `azure:${deployment}`;
} else {
  console.log("AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.\n");
  big = { client: ollama, model: "llama3.2" };
  bigName = "llama3.2";
}

// Same function as the starter: one prompt, one word back, any client.
async function classify({ client, model }: Target, text: string): Promise<string> {
  // Both models get this exact prompt, and it is byte-identical to the one in
  // ../../http/ollama.http and ../../http/azure.http, line breaks included. Reflowing these
  // four lines into one costs phi3 measured accuracy on both sets while leaving
  // llama3.2 unchanged, so varying the prompt shape and the model in the same
  // run measures nothing. See ../../expected-output.md.
  const prompt = `Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: ${text}`;
  const response = await client.chat.completions.create({ model, messages: [{ role: "user", content: prompt }], temperature: 0 });
  const raw = (response.choices[0].message.content ?? "").toLowerCase();
  const found = ["positive", "negative", "mixed"]
    .map((label) => ({ label, at: raw.indexOf(label) }))
    .filter((x) => x.at >= 0)
    .sort((a, b) => a.at - b.at);
  return found[0]?.label ?? raw.trim();
}

type Review = { id: string; product: string; rating: number; reviewer: string; text: string };
type Result = { review: Review; set: string; reference: string; small: string; big: string };

const labels: Record<string, { set: string; label: string }> = JSON.parse(readFileSync(resolve(DATA, "reference-labels.json"), "utf8"));
const results: Result[] = [];

for (const set of sets) {
  console.log(`── ${set} set ──`);
  console.log(`${"id".padEnd(9)} ${"reference".padEnd(10)} ${"phi3".padEnd(10)} ${bigName.padEnd(10)}`);
  for (const line of readFileSync(resolve(DATA, `${set}.jsonl`), "utf8").split("\n")) {
    if (!line.trim()) continue;
    const review: Review = JSON.parse(line);
    const reference = labels[review.id].label;
    const s = await classify(small, review.text);
    const b = await classify(big, review.text);
    results.push({ review, set, reference, small: s, big: b });
    const flag = s !== b ? "  <- disagree" : "";
    console.log(`${review.id.padEnd(9)} ${reference.padEnd(10)} ${s.padEnd(10)} ${b.padEnd(10)}${flag}`);
  }
  console.log();
}

console.log("── accuracy vs. reference labels ──");
for (const set of sets) {
  const batch = results.filter((r) => r.set === set);
  const smallOk = batch.filter((r) => r.small === r.reference).length;
  const bigOk = batch.filter((r) => r.big === r.reference).length;
  console.log(`${set.padEnd(5)}  phi3 ${smallOk}/${batch.length}   ${bigName} ${bigOk}/${batch.length}`);
}
console.log();

const disagreements = results.filter((r) => r.small !== r.big);
console.log(`── disagreements (${disagreements.length} of ${results.length}) ──`);
for (const d of disagreements) {
  const verdict = d.big === d.reference ? `${bigName} right` : d.small === d.reference ? "phi3 right" : "both wrong";
  console.log(`${d.review.id} [${d.set}] ref=${d.reference} phi3=${d.small} ${bigName}=${d.big}  (${verdict})`);
  const t = d.review.text;
  console.log(`  "${t.length <= 100 ? t : t.slice(0, 100).trimEnd() + "..."}"`);
}
if (disagreements.length === 0) console.log("(none this run)");
