// Five tools over the workshop's mock APIs, wired into a hand-written tool-calling
// loop. Every tool call prints as it happens, the permit step waits for a human
// yes, and a step budget bounds the loop.
//
//   npm run complete                                            the capstone request
//   npm run complete -- Plan me a trip on Avalanche Lake Trail in September
//   npm run complete -- --yes <request>                         auto-approve the permit gate
//
// Model: Azure OpenAI when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY /
// AZURE_OPENAI_DEPLOYMENT are set; otherwise Ollama llama3.2, which is much
// weaker at sequencing five tools. See ../F10-typescript.md before judging a local run.
//
// There is no agent framework here on purpose. The loop is the same one the lab's
// http/azure.http walks by hand: send the messages with the tools array, read the tool
// calls out of the reply, run them, append the results, repeat.

import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { createInterface } from "node:readline/promises";
import OpenAI, { AzureOpenAI } from "openai";
import type { ChatCompletionMessageParam, ChatCompletionTool } from "openai/resources/chat/completions";

const DATA = resolve(import.meta.dirname, "../../data");

function createChatClient(): { client: OpenAI; model: string } {
  const { AZURE_OPENAI_ENDPOINT: endpoint, AZURE_OPENAI_KEY: key, AZURE_OPENAI_DEPLOYMENT: deployment } = process.env;
  if (endpoint && key && deployment) {
    return { client: new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment }), model: deployment };
  }
  console.log("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.");
  return { client: new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" }), model: "llama3.2" };
}

// ---------------------------------------------------------------------------
// The five tools: ordinary functions over the workshop's fixture files. The
// descriptions in TOOLS below are the model's only documentation for each tool
// and parameter, so rewording them changes which tools get called and with what
// arguments. Treat that prose as behavior, not commentary. Every function prints
// itself on entry so the loop is visible while it runs.
// ---------------------------------------------------------------------------
let autoApprovePermits = false;
const called = new Set<string>();
let lastResultIds: string[] = [];

type Trail = { id: string; name: string; park: string; distance_mi: number; elevation_ft: number; difficulty: string; features: string[]; description: string };
type Report = { id: string; trail_id: string; date: string; text: string };

const load = (name: string) => JSON.parse(readFileSync(resolve(DATA, name), "utf8"));

function narrate(tool: string, args: unknown): void {
  called.add(tool);
  console.log(`[tool] ${tool} ${JSON.stringify(args)}`);
}

function result(payload: unknown): string {
  const text = typeof payload === "string" ? payload : JSON.stringify(payload);
  console.log(`  [result] ${text.length > 120 ? text.slice(0, 120) + "..." : text}`);
  return text;
}

function parkEntry(table: Record<string, unknown>, park: string): unknown {
  const first = park.split(" ")[0].toLowerCase();
  for (const [name, value] of Object.entries(table)) {
    if (name.startsWith("_")) continue;
    const n = name.toLowerCase();
    if (n.includes(park.toLowerCase()) || park.toLowerCase().includes(n) || n.includes(first)) return value;
  }
  return undefined;
}

function searchTrails({ park = "Glacier National Park", features = null, max_difficulty = null }: { park?: string; features?: string[] | null; max_difficulty?: string | null }): string {
  narrate("search_trails", { park, features, max_difficulty });
  // Small models sometimes send a string where the schema says array.
  if (typeof features === "string") features = (features as string).trim() ? [features as string] : null;
  const rank = (d: string) => (d === "easy" ? 0 : d === "moderate" ? 1 : 2);
  const maxRank = max_difficulty ? rank(max_difficulty.toLowerCase()) : 2;
  const found = (load("trails.json") as Trail[])
    .filter((t) => t.park.toLowerCase().includes(park.toLowerCase()))
    .filter((t) => rank(t.difficulty) <= maxRank)
    .filter((t) => !features || features.length === 0 || features.some((f) => t.features.some((x) => x.toLowerCase().includes(f.toLowerCase()))))
    .slice(0, 8)
    .map(({ id, name, park, distance_mi, elevation_ft, difficulty, features }) => ({ id, name, park, distance_mi, elevation_ft, difficulty, features }));
  lastResultIds = found.map((t) => t.id);
  return result(found);
}

function getWeather({ park = "Glacier National Park" }: { park?: string }): string {
  narrate("get_weather", { park });
  const entry = parkEntry(load("mock-apis/weather.json"), park);
  return result(entry ?? { error: `No forecast available for '${park}'.` });
}

function getTrailConditions({ trail_id = null }: { trail_id?: string | null }): string {
  narrate("get_trail_conditions", { trail_id });
  // Every tool parameter here has a default and every failure returns an error
  // string instead of throwing. A model that supplies a missing or malformed
  // argument gets a correctable message back naming the valid ids, rather than
  // crashing the process mid-loop.
  if (!trail_id || ["", "null", "string"].includes(trail_id.trim())) {
    const candidates = lastResultIds.length > 0 ? lastResultIds.join(", ") : "call search_trails first";
    return result({ error: `trailId is required. Call this tool again with one of these ids: ${candidates}.` });
  }
  // A model may pass the trail name where an id is expected, so resolve names
  // too instead of returning nothing found.
  let id = trail_id;
  if (!id.toLowerCase().startsWith("trail-")) {
    const byName = (load("trails.json") as Trail[]).find((t) => t.name.toLowerCase().includes(id.toLowerCase()));
    if (byName) id = byName.id;
  }
  const reports: Report[] = readFileSync(resolve(DATA, "condition-reports.jsonl"), "utf8").split("\n").filter((l) => l.trim()).map((l) => JSON.parse(l));
  const mine = reports.filter((r) => r.trail_id.toLowerCase() === id.toLowerCase()).sort((a, b) => b.date.localeCompare(a.date)).slice(0, 4);
  if (mine.length === 0) return result({ error: `No condition reports found for '${id}'.` });
  return result(mine.map((r) => ({ date: r.date, report: r.text })));
}

function checkCampsites({ park = "Glacier National Park" }: { park?: string }): string {
  narrate("check_campsites", { park });
  const entry = parkEntry(load("mock-apis/campsites.json"), park);
  return result(entry ?? { error: `No campsite data for '${park}'.` });
}

async function requestPermit({ park = "Glacier National Park", zone = "Lake McDonald / Sperry", dates = "unspecified", group_size = 2 }: { park?: string; zone?: string; dates?: string; group_size?: number }): Promise<string> {
  narrate("request_permit", { park, zone, dates, group_size });
  // Filing a permit is the one irreversible action in this agent, so it never
  // runs on the model's say-so; a human confirms first. --yes bypasses the
  // prompt and exists for demo runs only.
  console.log(`  [gate] About to file a permit request: ${park}, zone '${zone}', ${dates}, group of ${group_size}.`);
  let approved: boolean;
  if (autoApprovePermits) {
    console.log("  [gate] --yes supplied; auto-approved.");
    approved = true;
  } else {
    const rl = createInterface({ input: process.stdin, output: process.stdout });
    const answer = (await rl.question("  [gate] File it? [y/N] ")).trim().toLowerCase();
    rl.close();
    approved = answer === "y" || answer === "yes";
  }
  if (!approved) {
    return result({ status: "cancelled", message: "The user declined to file the permit request. Do not retry; finish the itinerary and note that no permit was filed." });
  }
  return result(load("mock-apis/permits.json").submit_response);
}

const FUNCTIONS: Record<string, (args: any) => string | Promise<string>> = {
  search_trails: searchTrails,
  get_weather: getWeather,
  get_trail_conditions: getTrailConditions,
  check_campsites: checkCampsites,
  request_permit: requestPermit,
};

// The same five definitions the lab's data/tool-definitions.json ships, so what
// the model sees here is exactly what it sees from the .http file.
const TOOLS: ChatCompletionTool[] = load("tool-definitions.json").tools;

const SYSTEM_PROMPT = `You are the trip-planning agent for Trailhead Guides, a hiking app.
Today's date is September 11, 2026.

Plan trips using your tools; never invent trails, weather, availability,
or conditions. Every trail name, forecast, campground, and condition in
your answer must have come back from a tool call in this conversation.

Call the tools one at a time, in this order, and do not write any part of
the itinerary until all of them have been called:
1. get_weather for the park.
2. search_trails for candidate trails that fit the request.
3. get_trail_conditions for EVERY trail you intend to recommend, one call
   per trail, using the trail id returned by search_trails.
   If the newest reports for a trail mention a closure, a washout, a bridge
   that is out, or any other reason hikers are turning around, that trail is
   CLOSED. Do not schedule a day on a closed trail. Replace it with another
   trail from search_trails and state plainly, in the itinerary, that the
   original trail is closed and why.
4. check_campsites for where to stay each night.
5. request_permit once, only if a backcountry site or permit zone is involved.

If you have not yet called search_trails and get_trail_conditions, your
next move is a tool call, not prose.

Then write the final itinerary: one section per day with trail, campsite,
and how the forecast shaped the choice (put harder or more exposed hiking
on the drier days). End with the permit status.`;

// ---------------------------------------------------------------------------
// The loop. Send everything so far, read the reply; if it asked for tools, run
// them and append the results; repeat until it answers in prose or the step
// budget runs out. That budget is the only bound on the loop: a model that keeps
// deciding to call one more tool has no other stopping condition.
// ---------------------------------------------------------------------------
const MAX_ITERATIONS = 12;

async function runAgent(client: OpenAI, model: string, messages: ChatCompletionMessageParam[]): Promise<string> {
  for (let i = 0; i < MAX_ITERATIONS; i++) {
    const response = await client.chat.completions.create({ model, messages, tools: TOOLS, tool_choice: "auto" });
    const message = response.choices[0].message;
    if (!message.tool_calls || message.tool_calls.length === 0) {
      messages.push({ role: "assistant", content: message.content ?? "" });
      return message.content ?? "";
    }
    // The assistant turn that made the calls has to be replayed before the
    // tool results, in the shape the API returned it.
    messages.push({ role: "assistant", content: message.content, tool_calls: message.tool_calls });
    for (const call of message.tool_calls) {
      if (call.type !== "function") continue;
      const fn = FUNCTIONS[call.function.name];
      let args: Record<string, unknown> = {};
      try { args = JSON.parse(call.function.arguments || "{}"); } catch { args = {}; }
      const output = fn ? await fn(args) : JSON.stringify({ error: `Unknown tool '${call.function.name}'.` });
      messages.push({ role: "tool", tool_call_id: call.id, content: output });
    }
  }
  console.log(`[budget] stopped after ${MAX_ITERATIONS} iterations`);
  return "";
}

const args = process.argv.slice(2);
autoApprovePermits = args.includes("--yes");
const request = args.filter((a) => a !== "--yes").join(" ") || "Plan me a 3-day trip in Glacier National Park for September 14-16.";

const { client, model } = createChatClient();
const messages: ChatCompletionMessageParam[] = [{ role: "system", content: SYSTEM_PROMPT }, { role: "user", content: request }];

console.log(`Request: ${request}`);
console.log("=".repeat(60));

// A model can stop mid-plan believing it is done and start writing the itinerary
// from tools it never called. When a required tool is still uncalled, name it and
// let the loop continue. Capped at three nudges so a stuck model cannot spin here.
// A frontier model should need none of these; [nudge] lines mean the model is
// underpowered for the task, not that the app is broken.
const REQUIRED = ["get_weather", "search_trails", "get_trail_conditions", "check_campsites"];
let answer = await runAgent(client, model, messages);

for (let nudge = 0; nudge < 3; nudge++) {
  const missing = REQUIRED.filter((t) => !called.has(t));
  if (missing.length === 0) break;
  console.log(`[nudge] still missing: ${missing.join(", ")}`);
  const hint = missing.includes("get_trail_conditions") && lastResultIds.length > 0 ? ` Use one of these trail ids: ${lastResultIds.join(", ")}.` : "";
  messages.push({ role: "user", content: `You have not called these tools yet: ${missing.join(", ")}. Call the next one now with real arguments.${hint} Do not write the itinerary yet.` });
  answer = await runAgent(client, model, messages);
}

// The mirror failure: the model announces it has finished calling tools and then
// stops without ever writing the plan. One turn asking for it directly.
if (!answer.toLowerCase().includes("day")) {
  console.log("[nudge] tools are done but no itinerary was written; asking for it.");
  messages.push({ role: "user", content: "Every tool you need has been called. Write the final itinerary now, using only what the tools returned. Do not call any more tools." });
  answer = await runAgent(client, model, messages);
}

console.log("=".repeat(60));
console.log(answer);
