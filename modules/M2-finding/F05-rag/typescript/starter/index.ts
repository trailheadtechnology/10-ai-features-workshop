// Demo starting point: one chat client, one question, no context.
// This is the "plain chatbot" from step 1 of the demo script. It answers
// confidently. It is also wrong: the model puts Sperry Chalet in California
// and guesses at fire rules the park wrote down years ago.
// Run: npm run starter [-- "your question"]

import OpenAI from "openai";

const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });

const question = process.argv.slice(2).join(" ") || "Can I have a campfire at Sperry Chalet in September?";

console.log(`Q: ${question}\n`);
const response = await client.chat.completions.create({ model: "llama3.2", messages: [{ role: "user", content: question }] });
console.log(response.choices[0].message.content);
