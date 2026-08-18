// Classifies every inquiry in ../../data/inquiries-slice.jsonl and scores the
// result against ../../data/reference-labels.json.
// Run: npm run complete
//
// The category comes back as a zod enum through structured output, so the
// model can only return a label the routing table already knows how to handle.
// Adding a category here means adding a routing destination for it too, or the
// lookup on the routing table will come back undefined.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import OpenAI from "openai";
import { zodResponseFormat } from "openai/helpers/zod";
import { z } from "zod";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
const MODEL = "llama3.2";
const DATA = resolve(import.meta.dirname, "../../data");

const Category = z.enum(["permit", "conditions", "complaint", "lost-and-found", "emergency", "general", "unsure"]);
type Category = z.infer<typeof Category>;
const TriageResult = z.object({ category: Category });

// These descriptions are the taxonomy, and editing them changes behavior more
// than any code below. Two rules constrain any rewrite. Emergency wins over
// every other category, including messages that also mention a permit or a lost
// item, so the ordering paragraph at the end must stay. And unsure has to stay
// narrow: it means two queues must both act on one message, not that the model
// found the message hard. Widen it and it fills up with ordinary traffic, which
// is the unsorted inbox this system replaced.
const prompt = (text: string) => `You are the triage system for the Trailhead Guides shared inbox.
Classify the visitor message into exactly one category.

- permit: reserving, changing, canceling, or paying for a permit, pass,
  or reservation, including billing problems and missing confirmations
  for a permit application.
- conditions: asking whether a trail, road, or area is open, safe, or
  passable right now: snow, water levels, washouts, wildlife activity,
  closures.
- complaint: unhappy about a park facility, service, or staff member
  and wants it acknowledged or fixed.
- lost-and-found: reporting a lost or found physical item.
- emergency: a person may be hurt, missing, or in danger right now and
  needs immediate human attention.
- general: anything else: park rules, fees, trip planning, questions
  that fit none of the above.
- unsure: two different queues both have to act before this message can
  be resolved, so no single queue owns it. The case that qualifies: the
  sender asks about trail conditions AND asks someone to change, refund,
  or cancel a booking. Trail info cannot issue a refund, and the permits
  office does not decide whether a trail is passable, so a human reads
  this queue and splits the work. Also use unsure when the message fits
  none of the categories above.

Decide in this order. First, if anyone might be hurt, missing, or in
danger, answer emergency and stop; never answer unsure for those, even
when the message also mentions permits, conditions, or a lost item.
Second, if one queue can resolve the whole message on its own, answer
that queue; a booking or reservation problem with nothing else attached
is permit, not unsure. Third, only if two queues must both act, answer
unsure. Unsure is not a catch-all for anything hard.

Message:
${text}`;

const clip = (text: string, n: number) => (text.length <= n ? text : text.slice(0, n) + "...");

type Inquiry = { id: string; channel: string; received: string; text: string };
const inquiries: Inquiry[] = readFileSync(resolve(DATA, "inquiries-slice.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
const reference: { routing: Record<string, string>; labels: Record<string, string> } = JSON.parse(readFileSync(resolve(DATA, "reference-labels.json"), "utf8"));

const results: { inquiry: Inquiry; category: Category }[] = [];
for (const inquiry of inquiries) {
  const response = await client.chat.completions.parse({
    model: MODEL,
    messages: [{ role: "user", content: prompt(inquiry.text) }],
    response_format: zodResponseFormat(TriageResult, "triage"),
    // Anything above 0 makes the same message land in different queues on
    // different runs, which makes a scored comparison against fixed reference
    // labels meaningless.
    temperature: 0,
  });
  results.push({ inquiry, category: response.choices[0].message.parsed!.category });
  process.stdout.write(".");
}
console.log("\n");

// Emergencies print before the routing table and are sorted to the top of it.
// A person scanning this output under time pressure must not have to read past
// the first screen to find one.
const emergencies = results.filter((r) => r.category === "emergency");
if (emergencies.length > 0) {
  console.log("!!! EMERGENCY: route to dispatch, page the duty ranger now !!!");
  for (const { inquiry } of emergencies) console.log(`!!! ${inquiry.id}  ${clip(inquiry.text, 70)}`);
  console.log();
}

console.log(`${"id".padEnd(10)} ${"category".padEnd(15)} routed to`);
console.log("-".repeat(62));
for (const { inquiry, category } of [...results].sort((a, b) => Number(a.category !== "emergency") - Number(b.category !== "emergency"))) {
  console.log(`${inquiry.id.padEnd(10)} ${category.padEnd(15)} ${reference.routing[category]}`);
}

// Two scores, and they are not equally important. Overall accuracy is the
// headline number; recall on the emergency class is the one that decides
// whether this taxonomy is safe to ship. A missed emergency is a person waiting
// in a queue nobody is watching, and no amount of accuracy elsewhere offsets it.
// If you tune the category descriptions, judge the change on emergency recall
// first and treat a drop there as a failure even when accuracy improves.
const labels = reference.labels;
const correct = results.filter((r) => r.category === labels[r.inquiry.id]).length;
const emergencyIds = Object.entries(labels).filter(([, v]) => v === "emergency").map(([k]) => k);
const caught = emergencies.filter((e) => emergencyIds.includes(e.inquiry.id)).length;

console.log();
console.log(`Accuracy vs reference labels: ${correct}/${results.length}`);
console.log(`Emergency recall: ${caught}/${emergencyIds.length} ` +
  (caught === emergencyIds.length ? "(all caught; the metric that matters)" : "(MISSED ONE; this fails, whatever the accuracy says)"));
for (const { inquiry, category } of results.filter((r) => r.category !== labels[r.inquiry.id])) {
  console.log(`  miss: ${inquiry.id} got ${category}, reference says ${labels[inquiry.id]}`);
}
