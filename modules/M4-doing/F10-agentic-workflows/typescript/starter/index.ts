// The trip request goes to a plain chat completion, with no tools and no loop.
// Nothing here can reach the trail catalog, the weather feed, or the condition
// reports, so the itinerary that comes back is fluent, generic, and books
// nothing. That gap is the point of this script; the fix lives in ../complete.
//
// Run: npm run starter [-- your own trip request]

import OpenAI, { AzureOpenAI } from "openai";

function createChatClient(): { client: OpenAI; model: string } {
  const { AZURE_OPENAI_ENDPOINT: endpoint, AZURE_OPENAI_KEY: key, AZURE_OPENAI_DEPLOYMENT: deployment } = process.env;
  if (endpoint && key && deployment) {
    return { client: new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment }), model: deployment };
  }
  console.log("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.");
  return { client: new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" }), model: "llama3.2" };
}

const { client, model } = createChatClient();

const request = process.argv.slice(2).join(" ") || "Plan me a 3-day trip in Glacier National Park for September 14-16.";

console.log(`Request: ${request}`);
console.log("-".repeat(60));

const response = await client.chat.completions.create({
  model,
  messages: [{ role: "user", content: `You are the trip planner for Trailhead Guides, a hiking app.\n\n${request}` }],
});

console.log(response.choices[0].message.content);
console.log("-".repeat(60));
console.log("Note: zero tool calls were made. No weather was checked, no");
console.log("conditions were read, nothing was booked. Fluent and useless.");
