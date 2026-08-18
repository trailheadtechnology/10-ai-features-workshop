// Finished demo, matching the demo script in docs/slides/outlines:
//   npm run complete                                the naive prompt (the book report)
//   npm run complete -- --briefing                  3-bullet hiker briefing
//   npm run complete -- --headline                  one-line trail status for a card UI
//   npm run complete -- --briefing --audience ranger
//   Any non-flag argument is a path to a different trip report.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const MODEL = "llama3.2";
const DATA = resolve(import.meta.dirname, "../../data");

function stripFrontMatter(markdown: string): string {
  const parts = markdown.split("---");
  return parts.length >= 3 ? parts.slice(2).join("---").trim() : markdown.trim();
}

let reportPath = resolve(DATA, "tr-0004.md");
let mode = "naive";
let audience = "hiker";
const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case "--briefing": mode = "briefing"; break;
    case "--headline": mode = "headline"; break;
    case "--audience": audience = args[++i]; break;
    default: reportPath = args[i]; break;
  }
}

const report = stripFrontMatter(readFileSync(reportPath, "utf8"));

const audienceFocus = audience === "ranger"
  ? "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
  : "a hiker planning to hike this trail within the next week";

let prompt: string;
switch (mode) {
  // Deliberately the weak prompt. It is kept so the two prompts can be run
  // back to back against the same report; nothing about it should be fixed.
  case "naive":
    prompt = `Summarize this trip report.\n\n${report}`;
    break;

  // The last two lines are load-bearing, not politeness. The bullets require a
  // hazards slot and require any hazard to come first, so on a report with no
  // hazard the model will promote the nearest noun (a bear, a creek, the word
  // "avalanche" in the trail name) into a closure. Giving it a legal way to
  // report nothing is what stops that. Measurements in ../../expected-output.md.
  case "briefing":
    prompt = `You are helping ${audienceFocus}.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.

${report}`;
    break;

  // Same client, same report, same call. Only the instruction changes to fit a
  // different UI slot, so no new infrastructure is needed for a new surface.
  case "headline":
    prompt = `From the trip report below, write ONE line of at most 12 words,
suitable for a status badge on a trail card in an app.
Lead with the most important condition or closure. No preamble.

${report}`;
    break;

  default:
    throw new Error(`unknown mode ${mode}`);
}

const response = await client.chat.completions.create({ model: MODEL, messages: [{ role: "user", content: prompt }] });
console.log(response.choices[0].message.content);
