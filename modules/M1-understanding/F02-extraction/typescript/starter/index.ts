// Demo starting point: the naive approach. Ask for JSON in the prompt, get
// back whatever the model feels like: prose preamble, markdown fences, drifting
// field names. This is what the parsed, typed response replaces.
// Run: npm run starter [-- path-to-trip-report.md]

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const DATA = resolve(import.meta.dirname, "../../data");

function stripFrontMatter(markdown: string): string {
  const parts = markdown.split("---");
  return parts.length >= 3 ? parts.slice(2).join("---").trim() : markdown.trim();
}

const reportPath = process.argv[2] ?? resolve(DATA, "tr-0007.md");
const report = stripFrontMatter(readFileSync(reportPath, "utf8"));

const response = await client.chat.completions.create({
  model: "llama3.2",
  messages: [{ role: "user", content: `Extract the details of this trip report as JSON.\n\n${report}` }],
});
console.log(response.choices[0].message.content);
