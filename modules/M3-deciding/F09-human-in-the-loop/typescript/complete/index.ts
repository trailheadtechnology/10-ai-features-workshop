// Review queue: the model drafts, a human decides, and every decision is logged
// to decisions.jsonl.
//
//   npm run complete                                  review the queue: [a]pprove / [e]dit / [r]eject / [s]kip
//   npm run complete -- --policy                      print the routing policy table and exit
//   npm run complete -- --auto-approve-dry-run        non-interactive run for testing and CI
//   npm run complete -- ../data/inquiries.jsonl       any queue file works
//
// SAFETY INVARIANT, load-bearing, do not weaken:
// emergencies never reach the model. The policy table below routes them to
// human-only and the loop skips the API call entirely, in code, before any
// request is built. The system prompt also tells the model to escalate instead
// of drafting, but that instruction is a request and this lane is a guarantee.
// A model that ignores the instruction and writes a warm, fluent, confident
// reply to someone reporting an overdue hiker is not a hypothetical; it is the
// documented behavior of the model this demo ships with. Anyone editing this
// file must keep the emergency path free of model calls. Adding a "just draft
// it and let the reviewer catch it" shortcut here puts a reassuring lie in front
// of a person who needed a dispatcher.

import { appendFileSync, existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { userInfo } from "node:os";
import { dirname, relative, resolve } from "node:path";
import { createInterface } from "node:readline/promises";
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

// The lane is chosen by what a wrong answer costs, not by how good the model is
// at the category. Everything reversible can be drafted; emergency is
// irreversible and stays human-only. Note the lookup below defaults an unknown
// category to human-only, so a category added upstream fails closed rather than
// quietly acquiring a draft lane.
const POLICY: Record<string, string> = {
  "trail-condition": "draft-for-approval",
  "permit": "draft-for-approval",
  "complaint": "draft-for-approval",
  "general": "draft-for-approval",
  "lost-and-found": "draft-for-approval",
  "emergency": "human-only",
};

const DATA = resolve(import.meta.dirname, "../../data");
const HERE = import.meta.dirname;
let inquiriesPath = resolve(DATA, "inquiries.jsonl");
let outboxDir = resolve(HERE, "outbox");
let decisionsPath = resolve(HERE, "decisions.jsonl");
let autoApprove = false;

function printPolicy(): void {
  console.log("Routing policy (error cost decides the lane):");
  for (const [category, lane] of Object.entries(POLICY)) console.log(`  ${category.padEnd(16)} ${lane}`);
  console.log();
}

const args = process.argv.slice(2);
for (let i = 0; i < args.length; i++) {
  switch (args[i]) {
    case "--auto-approve-dry-run": autoApprove = true; break;
    case "--outbox": outboxDir = resolve(args[++i]); break;
    case "--decisions": decisionsPath = resolve(args[++i]); break;
    case "--policy": printPolicy(); process.exit(0);
    default: inquiriesPath = resolve(args[i]); break;
  }
}

const dataDir = dirname(inquiriesPath);
mkdirSync(outboxDir, { recursive: true });

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const reviewer = autoApprove ? "auto-approve-dry-run" : userInfo().username;
const counts: Record<string, number> = {};
const rl = autoApprove ? null : createInterface({ input: process.stdin, output: process.stdout });

type Inquiry = { id: string; channel: string; received: string; category: string; doc: string; text: string };
type Decision = { at: string; inquiryId: string; category: string; lane: string; decision: string; reviewer: string; draft: string | null; final: string | null; editDistance: number };

const indent = (text: string) => text.split("\n").map((l) => "  | " + l.trimEnd()).join("\n");
const log = (d: Decision) => appendFileSync(decisionsPath, JSON.stringify(d) + "\n");
const rel = (p: string) => relative(process.cwd(), p) || ".";
const record = (inquiry: Inquiry, lane: string, decision: string, draft: string | null, final: string | null, editDistance: number): Decision =>
  ({ at: new Date().toISOString(), inquiryId: inquiry.id, category: inquiry.category, lane, decision, reviewer, draft, final, editDistance });
const bump = (key: string) => { counts[key] = (counts[key] ?? 0) + 1; };

async function readEdited(draft: string): Promise<string> {
  console.log("  Type the reply you want to send. End with a single '.' on its own line.");
  console.log("  Press Enter on the first line to start from the draft text instead.\n");
  const lines: string[] = [];
  let first = true;
  for (;;) {
    const line = await rl!.question("");
    if (line === ".") break;
    if (first && line.length === 0) {
      lines.push(draft);
      console.log("  (draft copied in; keep typing to append, '.' to finish)");
    } else {
      lines.push(line);
    }
    first = false;
  }
  const edited = lines.join("\n").trim();
  return edited.length === 0 ? draft : edited;
}

// Logged on every decision so the promotion question has data behind it rather
// than a feeling. It measures how much someone typed, not whether they were
// fixing a comma or preventing a lawsuit, so it can support an argument for
// promoting a lane and must never be the only evidence for one.
// O(a*b) and unbounded by draft length; fine for a review queue, not for bulk.
function editDistance(a: string, b: string): number {
  let previous = Array.from({ length: b.length + 1 }, (_, j) => j);
  for (let i = 1; i <= a.length; i++) {
    const current = [i];
    for (let j = 1; j <= b.length; j++) {
      const cost = a[i - 1] === b[j - 1] ? 0 : 1;
      current[j] = Math.min(current[j - 1] + 1, previous[j] + 1, previous[j - 1] + cost);
    }
    previous = current;
  }
  return previous[b.length];
}

printPolicy();
if (autoApprove) console.log("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.\n");

for (const line of readFileSync(inquiriesPath, "utf8").split("\n")) {
  if (!line.trim()) continue;
  const inquiry: Inquiry = JSON.parse(line);
  const lane = POLICY[inquiry.category] ?? "human-only";

  console.log("-".repeat(72));
  console.log(`${inquiry.id}  ·  ${inquiry.category}  ·  ${inquiry.channel}  ·  lane: ${lane}`);
  console.log("-".repeat(72));
  console.log(indent(inquiry.text));
  console.log();

  // THE GATE. This must stay above the API call, and the API call must stay
  // below it. A human-only message is escalated and logged without a single
  // token being spent on it, so there is no draft to leak, no reviewer fatigue
  // to survive, and no sampling luck involved. The ESCALATE handling further
  // down is a backstop for emergencies that arrive miscategorized; it is never
  // the control, because it runs after the model has already had its say.
  if (lane === "human-only") {
    console.log("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.\n");
    log(record(inquiry, lane, "escalated", null, null, 0));
    bump("escalated");
    continue;
  }

  const snippetPath = resolve(dataDir, "snippets", inquiry.doc || "");
  const snippet = inquiry.doc && existsSync(snippetPath) ? readFileSync(snippetPath, "utf8").trim() : "(none on file for this message)";

  process.stdout.write("  drafting...");
  const response = await client.chat.completions.create({
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
  const draft = (response.choices[0].message.content ?? "").trim();
  console.log("\r  draft:      \n");
  console.log(indent(draft));
  console.log();

  // Second layer, for an emergency that reached here under the wrong category.
  // An ESCALATE prefix is a hard stop: the draft is logged for the audit trail
  // but never offered for approval, because a reviewer presented with a
  // sendable-looking reply may send it. Do not soften this into a warning.
  if (draft.toUpperCase().startsWith("ESCALATE")) {
    console.log("  Model asked to escalate. Draft discarded, routing to a human.\n");
    log(record(inquiry, lane, "escalated", draft, null, 0));
    bump("escalated");
    continue;
  }

  let decision: string;
  let final: string | null = null;
  if (autoApprove) {
    decision = "approved";
    final = draft;
    console.log("  [auto] approved\n");
  } else {
    const key = (await rl!.question("  [a]pprove  [e]dit  [r]eject  [s]kip > ")).trim().toLowerCase();
    console.log();
    switch (key) {
      case "a": decision = "approved"; final = draft; break;
      case "e": decision = "edited"; final = await readEdited(draft); break;
      case "r": decision = "rejected"; break;
      default: decision = "skipped"; break;
    }
  }

  if (final !== null) {
    const path = resolve(outboxDir, `${inquiry.id}.txt`);
    writeFileSync(path, final + "\n");
    console.log(`  -> ${decision}, queued at ${rel(path)}\n`);
  } else {
    console.log(`  -> ${decision}, nothing queued\n`);
  }

  log(record(inquiry, lane, decision, draft, final, editDistance(draft, final ?? "")));
  bump(decision);
}

rl?.close();
console.log("=".repeat(72));
console.log("Queue done: " + Object.entries(counts).map(([k, v]) => `${v} ${k}`).join(", "));
console.log(`Audit trail: ${rel(decisionsPath)}   ·   Outbox: ${rel(outboxDir)}/`);
