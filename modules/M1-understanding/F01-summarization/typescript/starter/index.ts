// Demo starting point: one chat client, one call, one naive prompt.
// Run: npm run starter [-- path-to-trip-report.md]

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

// Ollama speaks the OpenAI chat API on /v1, so the official SDK is the client.
// Swapping providers later is a different constructor here and nothing else.
const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });

const DATA = resolve(import.meta.dirname, "../../data");

function stripFrontMatter(markdown: string): string {
  const parts = markdown.split("---");
  return parts.length >= 3 ? parts.slice(2).join("---").trim() : markdown.trim();
}

const reportPath = process.argv[2] ?? resolve(DATA, "tr-0001.md");
const report = stripFrontMatter(readFileSync(reportPath, "utf8"));

const response = await client.chat.completions.create({
  model: "llama3.2",
  messages: [{ role: "user", content: `Summarize this trip report.\n\n${report}` }],
});
console.log(response.choices[0].message.content);
