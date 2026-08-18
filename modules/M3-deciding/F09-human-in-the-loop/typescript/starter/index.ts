// Starting point, and deliberately unsafe. Every draft goes straight out with no
// human between the model and the visitor, and nothing is logged, so there is no
// record of what was sent or any way to find out later.
// Run: npm run starter [-- path-to-inquiries.jsonl]
//
// Nothing here treats an emergency differently. The system prompt asks the model
// to escalate instead of drafting, and the model is free to ignore that and
// often does, which is why complete/ moves the decision out of the prompt and
// into a policy check in code. Do not use this shape on a real inbox.

import { existsSync, readFileSync } from "node:fs";
import { dirname, resolve } from "node:path";
import OpenAI from "openai";

const SYSTEM_PROMPT = `You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides.
A human ranger reviews your draft before anything is sent, so write it ready to send:
friendly, plain, professional, at most two short paragraphs, signed
'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure,
state the rule and cite the source document number and section (for example
GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided;
if the excerpt does not answer the question, say a ranger will follow up with
specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
Exception: if the visitor's message reports an emergency, an injury, a possible fire,
or a missing or overdue person, do not draft a reply at all. Output exactly one line
beginning with ESCALATE: followed by a one-line reason, so the message goes straight
to dispatch.`;

const DATA = resolve(import.meta.dirname, "../../data");
const inquiriesPath = process.argv[2] ? resolve(process.argv[2]) : resolve(DATA, "inquiries.jsonl");
const dataDir = dirname(inquiriesPath);

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });

type Inquiry = { id: string; channel: string; received: string; category: string; doc: string; text: string };

for (const line of readFileSync(inquiriesPath, "utf8").split("\n")) {
  if (!line.trim()) continue;
  const inquiry: Inquiry = JSON.parse(line);

  const snippetPath = resolve(dataDir, "snippets", inquiry.doc || "");
  const snippet = inquiry.doc && existsSync(snippetPath) ? readFileSync(snippetPath, "utf8").trim() : "(none on file for this message)";

  const draft = await client.chat.completions.create({
    model: "llama3.2",
    messages: [
      { role: "system", content: SYSTEM_PROMPT },
      { role: "user", content: `Reference excerpt:
${snippet}

Visitor message (${inquiry.channel}, received ${inquiry.received}):
${inquiry.text}

Draft the reply.` },
    ],
  });

  // No review step and no audit record. In a real deployment this line is the
  // send call, and by the time anyone reads the output the mail has gone.
  console.log(`=== SENT to visitor · ${inquiry.id} (${inquiry.category}) ===`);
  console.log((draft.choices[0].message.content ?? "").trim());
  console.log();
}

console.log("All replies sent. Nobody read them. Nothing was logged.");
