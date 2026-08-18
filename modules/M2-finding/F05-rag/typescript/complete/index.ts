// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete                                       the Sperry Chalet question, grounded
//   npm run complete -- "Is the Avalanche Lake Trail open right now?"
//   npm run complete -- --no-context                       step 1: the confident wrong answer
//   npm run complete -- --alpha 1.0                        pure cosine, the wrong-park neighbors
//   npm run complete -- --retrieval-only                   print the score table and stop
//   npm run complete -- --top-k 8 "your question"          vary retrieval depth
//   npm run complete -- --model qwen3:32b                  a bigger model, if you have the memory
//
// Retrieval is hybrid: normalized cosine similarity blended with a BM25-lite
// lexical score, so a distinctive proper noun like "Sperry" counts for something.
// Chunks are one numbered subsection each wherever a section ran long enough to
// hold a rule and the exception that overrides it; see ../../expected-output.md.
// Every generated answer's [chunk-id] citations are validated against the chunks
// that were actually retrieved.
//
// Retrieval always runs locally (nomic-embed-text). Generation uses Azure OpenAI
// when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY / AZURE_OPENAI_DEPLOYMENT are set,
// and falls back to the local model chosen just below when they are not.

import { existsSync, readFileSync, writeFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI, { AzureOpenAI } from "openai";

// ---------------------------------------------------------------------------
// Which local model generates the answer.
//
// Everything else in this file is identical either way: same retrieval, same
// chunks, same prompt. Only the model changes. All numbers below were measured
// on this pipeline and are written up in ../../expected-output.md.
//
//                            llama3.2 (3B)      qwen3:32b
//   answers correctly           97%               100%
//   opens with "Yes" before
//     saying fires are banned    40%                0%
//   refuses when it shouldn't    8%                 0%
//   invalid citations         9 per 200 runs      0 per 80 runs
//   download / memory           2 GB / any        20 GB / ~24 GB free
//   time per answer             0.9 s             16.6 s
//
// The big model is better at everything and costs 18x the wall clock and a
// machine most people do not have. Switch only if you have the memory; on 16 GB
// it will not load, and on 32 GB it will make you wait. This setting only affects
// generation. Retrieval is identical either way, so no model here can rescue an
// answer that is not in the retrieved chunks.
//
// let localModel = "qwen3:32b";   // see the memory note above before switching
let localModel = "llama3.2";       // the default, because it runs on any laptop in the room
// ---------------------------------------------------------------------------

const DATA = resolve(import.meta.dirname, "../../data");
const chunksPath = resolve(DATA, "chunks.jsonl");
const cachePath = resolve(import.meta.dirname, "embeddings.json");
let topK = 3;
let alpha = 0.6;          // weight on the semantic signal; 1.0 = cosine only
let noContext = false;
let retrievalOnly = false;
const questionParts: string[] = [];
const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case "--no-context": noContext = true; break;
    case "--retrieval-only": retrievalOnly = true; break;
    case "--top-k": topK = parseInt(args[++i], 10); break;
    case "--alpha": alpha = parseFloat(args[++i]); break;
    // Swap models without editing the file, for demoing the contrast live.
    case "--model": localModel = args[++i]; break;
    default: questionParts.push(args[i]); break;
  }
}
const question = questionParts.join(" ") || "Can I have a campfire at Sperry Chalet in September?";

const ollama = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });

// Generation: Azure OpenAI if configured, local llama3.2 otherwise. This is the
// one-line client swap from step 9 of the demo; everything downstream is identical.
const { AZURE_OPENAI_ENDPOINT: endpoint, AZURE_OPENAI_KEY: key, AZURE_OPENAI_DEPLOYMENT: deployment } = process.env;
let chatClient: OpenAI;
let chatModel: string;
if (endpoint && key && deployment) {
  chatClient = new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment });
  chatModel = deployment;
  console.log(`[generation: Azure OpenAI, deployment '${deployment}']`);
} else {
  chatClient = ollama;
  chatModel = localModel;
  console.log(`[generation: AZURE_OPENAI_* not set, falling back to local ${localModel}]`);
}

async function generate(prompt: string): Promise<string> {
  const response = await chatClient.chat.completions.create({ model: chatModel, messages: [{ role: "user", content: prompt }] });
  return response.choices[0].message.content ?? "";
}

console.log(`Q: ${question}\n`);

if (noContext) {
  // Step 1 of the demo: no retrieval, no context, pure model memory.
  console.log(await generate(question));
  process.exit(0);
}

// Load the pre-chunked park docs.
type Chunk = { chunk_id: string; source: string; text: string };
const chunks: Chunk[] = readFileSync(chunksPath, "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));

async function embed(texts: string[]): Promise<number[][]> {
  const response = await ollama.embeddings.create({ model: "nomic-embed-text", input: texts });
  return response.data.map((d) => d.embedding);
}

// Embed the chunks with local nomic-embed-text, caching to embeddings.json so
// only the first run pays the ~40 seconds of embedding time.
let index: Record<string, number[]>;
if (existsSync(cachePath)) {
  index = JSON.parse(readFileSync(cachePath, "utf8"));
} else {
  console.log(`[embedding ${chunks.length} chunks, one-time; caching to embeddings.json]`);
  index = {};
  for (let start = 0; start < chunks.length; start += 32) {
    const batch = chunks.slice(start, start + 32);
    const vectors = await embed(batch.map((c) => c.text));
    batch.forEach((c, i) => { index[c.chunk_id] = vectors[i]; });
  }
  writeFileSync(cachePath, JSON.stringify(index));
}

function cosineSimilarity(a: number[], b: number[]): number {
  let dot = 0, magA = 0, magB = 0;
  for (let i = 0; i < a.length; i++) {
    dot += a[i] * b[i];
    magA += a[i] * a[i];
    magB += b[i] * b[i];
  }
  return dot / (Math.sqrt(magA) * Math.sqrt(magB));
}

// Question filler. Without this, "can" and "have" carry as much weight as "Sperry"
// simply because no park document says "can I have".
const STOP_WORDS = new Set([
  "the", "and", "for", "are", "but", "not", "you", "your", "with", "that", "this", "these",
  "those", "from", "have", "has", "had", "was", "were", "been", "being", "can", "could",
  "will", "would", "shall", "should", "may", "might", "must", "does", "did", "doing",
  "what", "when", "where", "which", "who", "whom", "why", "how", "any", "all", "some",
  "there", "here", "then", "than", "them", "they", "their", "its", "his", "her", "our",
  "get", "got", "still", "now", "right", "just", "about", "into", "onto", "over", "under",
  "out", "off", "per", "via", "one", "two", "also", "more", "most", "much", "many", "each",
  "other", "such", "only", "own", "same", "too", "very", "let", "need", "want",
]);

// Lowercase, split on non-alphanumerics, drop question filler, and knock the
// plural off so "campfires" in a document matches "campfire" in a question.
// Crude on purpose: an attendee can read all of it in ten seconds.
function tokenize(text: string): string[] {
  return (text.toLowerCase().match(/[a-z0-9]+/g) ?? [])
    .filter((t) => t.length > 2 && !STOP_WORDS.has(t))
    .map((t) => (t.length > 3 && t.endsWith("s") && !t.endsWith("ss") ? t.slice(0, -1) : t));
}

function minMax(raw: Record<string, number>): Record<string, number> {
  const values = Object.values(raw);
  const lo = Math.min(...values), hi = Math.max(...values);
  const span = hi - lo;
  return Object.fromEntries(Object.entries(raw).map(([k, v]) => [k, span > 1e-9 ? (v - lo) / span : 0]));
}

// ---------------------------------------------------------------------------
// Signal 1: semantic. Feature 04's cosine search, verbatim, over the chunk index.
// ---------------------------------------------------------------------------
const [questionVector] = await embed([question]);
const cosine: Record<string, number> = Object.fromEntries(chunks.map((c) => [c.chunk_id, cosineSimilarity(questionVector, index[c.chunk_id])]));

// ---------------------------------------------------------------------------
// Signal 2: lexical. BM25-lite over the chunk text. The embedder collapses
// "campfire regulations" from five different parks onto nearly the same point;
// the word "Sperry" appears in exactly one document, and IDF makes that count.
// ---------------------------------------------------------------------------
const tokenized: Record<string, string[]> = Object.fromEntries(chunks.map((c) => [c.chunk_id, tokenize(c.text)]));
const avgLength = Object.values(tokenized).reduce((sum, t) => sum + t.length, 0) / chunks.length;
const docFreq = new Map<string, number>();
for (const terms of Object.values(tokenized)) {
  for (const term of new Set(terms)) docFreq.set(term, (docFreq.get(term) ?? 0) + 1);
}
const df = (t: string) => docFreq.get(t) ?? 0;

const K1 = 1.2, B = 0.3;
const n = chunks.length;
const queryTerms = [...new Set(tokenize(question))];
// IDF: a term in 1 of 250 chunks is worth far more than one in 200 of them.
const idf: Record<string, number> = Object.fromEntries(queryTerms.map((t) => [t, Math.log(1 + (n - df(t) + 0.5) / (df(t) + 0.5))]));

const lexical: Record<string, number> = {};
for (const c of chunks) {
  const terms = tokenized[c.chunk_id];
  const counts = new Map<string, number>();
  for (const t of terms) counts.set(t, (counts.get(t) ?? 0) + 1);
  let score = 0;
  for (const t of queryTerms) {
    const tf = counts.get(t);
    if (!tf) continue;
    score += idf[t] * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * terms.length / avgLength));
  }
  lexical[c.chunk_id] = score;
}

// Rescale both signals to 0..1 across the corpus for this question, so alpha
// means what it looks like it means and the two numbers are comparable on screen.
const semanticNorm = minMax(cosine);
const lexicalNorm = minMax(lexical);

const scored = chunks
  .map((c) => ({
    chunk: c,
    cosine: cosine[c.chunk_id],
    semanticNorm: semanticNorm[c.chunk_id],
    lexical: lexical[c.chunk_id],
    lexicalNorm: lexicalNorm[c.chunk_id],
    combined: alpha * semanticNorm[c.chunk_id] + (1 - alpha) * lexicalNorm[c.chunk_id],
  }))
  .sort((a, b) => b.combined - a.combined);
const top = scored.slice(0, topK);

const distinctive = [...queryTerms].sort((a, b) => idf[b] - idf[a]).slice(0, 5)
  .map((t) => `${t}(${idf[t].toFixed(2)}/${df(t)} chunks)`).join("  ");
console.log(`[query terms by IDF]  ${distinctive}`);
console.log(`[retrieved top ${topK}]  combined = ${alpha.toFixed(2)} * semantic + ${(1 - alpha).toFixed(2)} * lexical`);
console.log("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id");
top.forEach((h, r) => {
  console.log(`  ${String(r + 1).padStart(4)}  ${h.combined.toFixed(4).padStart(8)}   ${h.semanticNorm.toFixed(3).padStart(5)} (${h.cosine.toFixed(4)})   ${h.lexicalNorm.toFixed(3).padStart(5)} (${h.lexical.toFixed(2).padStart(5)})   ${h.chunk.chunk_id}`);
});
const margin = scored.length > 1 ? scored[0].combined - scored[1].combined : 0;
console.log(`  margin over rank 2: ${margin.toFixed(4)}\n`);

if (retrievalOnly) process.exit(0);

// Grounded prompt: context in, citations out, refusal when the context is silent.
//
// The corpus is full of dated notices ("CLOSED effective June 20, 2026, until further
// notice"). A model with no idea what day it is treats "is the trail open right now?" as
// a question its documents cannot speak to, and refuses. So we tell it the date.
//
// The date belongs in the refusal clause specifically, not in the Rules block above.
// Stated broadly, it invites the model to apply effective-date reasoning to rules that
// have no expiry, and it starts deciding a year-round fire ban has lapsed.
//
// Production passes new Date() here; this demo pins a date so the recorded outputs in
// ../../expected-output.md stay reproducible.
const TODAY = "September 23, 2026";
const REFUSAL = "The provided documents don't say.";

const retrievedIds = new Set(top.map((h) => h.chunk.chunk_id));
const context = top.map((h) => `chunk_id: ${h.chunk.chunk_id}\nsource: ${h.chunk.source}\n${h.chunk.text}`).join("\n\n");

const prompt = `You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "${REFUSAL}"
  A question about "right now" is answered from the context, not refused: today is ${TODAY},
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
${context}

Question: ${question}

Answer:
`;

// Any bracketed token containing a colon is a citation attempt, including the
// comma-separated lists llama3.2 sometimes writes. Being generous about what
// counts as an attempt is the point: we want to catch the near-misses.
function citations(text: string): string[] {
  return [...text.matchAll(/\[([^\]]*:[^\]]*)\]/g)]
    .flatMap((m) => m[1].split(","))
    .map((c) => c.trim())
    .filter((c) => c.includes(":"));
}
const invalidCitations = (text: string, valid: Set<string>) => [...new Set(citations(text).filter((c) => !valid.has(c)))];

let answer = await generate(prompt);
let bad = invalidCitations(answer, retrievedIds);

// Citation validation. An invalid citation is a product defect, not a style
// issue: it is a receipt pointing at a document nobody retrieved, and sometimes
// at a document that does not exist. Retry once with the valid ids spelled out,
// then strip whatever is still wrong so a bad receipt never reaches the visitor.
if (bad.length > 0 && answer.includes(REFUSAL)) {
  // A refusal that cites a source is already correct; by definition it has no sources,
  // so the only defect is the citation itself. Do not send this back to the model. Asked
  // to rewrite using valid ids, it rewrites the refusal too, and the exact string the
  // product matches on comes back paraphrased. Deleting a citation is a string operation.
  console.log(`!! CITATION CHECK FAILED: ${bad.map((c) => `[${c}]`).join(", ")} not in the retrieved set`);
  console.log("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed\n");
  answer = REFUSAL;
  bad = [];
} else if (bad.length > 0) {
  console.log(`!! CITATION CHECK FAILED: ${bad.map((c) => `[${c}]`).join(", ")} not in the retrieved set`);
  console.log("!! retrying once with the valid chunk_ids spelled out\n");
  const retryPrompt = prompt + `

Your previous answer cited ${bad.map((c) => `[${c}]`).join(", ")}, which is not a real chunk_id.
The only chunk_ids you may cite are, exactly:
${[...retrievedIds].map((id) => "  " + id).join("\n")}
Rewrite the answer using only those.
`;
  answer = await generate(retryPrompt);
  bad = invalidCitations(answer, retrievedIds);
  if (bad.length > 0) {
    console.log(`!! STILL INVALID after retry: ${bad.map((c) => `[${c}]`).join(", ")}`);
    console.log("!! stripping them; the answer below is unverified where the citation was removed\n");
    for (const c of bad) answer = answer.replaceAll(c, "invalid-citation-removed");
  }
}

console.log(answer);

const cited = [...new Set(citations(answer).filter((c) => retrievedIds.has(c)))];
console.log(`\n[citations: ${cited.length} valid (${cited.join(", ")}), ${bad.length} invalid]`);
