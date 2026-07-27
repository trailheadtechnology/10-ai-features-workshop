# Before the workshop

Fifteen minutes of pre-work saves you from fighting conference wifi on the day. Everything below is optional in the sense that we have fallbacks, but please try. The whole room moves faster when the models are already on your laptop.

## 1. Install Ollama and pull three models

Install Ollama from [ollama.com/download](https://ollama.com/download) (macOS, Windows, and Linux), then pull the three models the labs use:

```bash
ollama pull llama3.2
ollama pull phi3
ollama pull nomic-embed-text
```

That's roughly 5GB of downloads total, which is exactly why we're asking you to do it on hotel or home wifi instead of the venue's.

Heads up: we'll re-verify these model choices in the weeks before the event, since small-model quality moves fast. Check this file the week before the workshop in case a model name changed.

## 2. Verify it works

With Ollama running, this should return a chat response:

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [{ "role": "user", "content": "Say hello in five words." }],
  "stream": false
}'
```

If you got JSON back with a message in it, you're done with the required pre-work.

## 3. Azure OpenAI: nothing to do

The cloud modules (sentiment comparison, RAG generation, and the agent capstone) use Azure OpenAI. API keys and endpoint details are handed out at the door on workshop day. There is nothing to sign up for and nothing to pay for.

## 4. Optional: the .NET path

The instructor demos in C#, and every module ships `starter/` and `complete/` .NET projects if you want to follow along in the same stack. If that's you, install the current .NET SDK ([dot.net](https://dot.net); the exact version will be pinned in each module's `dotnet/` README once the code lands) and confirm `dotnet --version` runs. If you'd rather work in Python, JavaScript, Java, Go, or anything else with an HTTP client, skip this entirely. The labs don't require .NET.

## If you did none of this before arriving

You'll still be fine. We bring the Ollama models on USB drives for local copying, the venue keys are handed out either way, and every lab can be completed with nothing but the provided `.http` request files and any HTTP client you already have. You'll just spend the first coffee break copying models instead of chatting, which is a fate we'd like to help you avoid.
