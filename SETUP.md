# Before the Workshop

Fifteen minutes of pre-work saves you from fighting conference wifi on the day. Everything below is optional in the sense that we have fallbacks, but please try. The whole room moves faster when the models are already on your laptop.

## 1. Install Ollama and Pull Three Models

Install Ollama from [ollama.com/download](https://ollama.com/download) (macOS, Windows, and Linux), then pull the three models the labs use:

```bash
ollama pull llama3.2
ollama pull phi3
ollama pull nomic-embed-text
```

That's roughly 5GB of downloads total, which is exactly why we're asking you to do it on hotel or home wifi instead of the venue's.

These three are the only models the workshop needs, and they are deliberately small: every lab runs on a laptop with 8GB of RAM, and nothing here needs a GPU.

### Optional, and Only if Your Machine Is Big

Feature 05 (RAG) records what happens when you swap the 3B model for one ten times larger. The numbers are written up either way, so this is for people who would rather run the comparison than read it. If that's you, pull it now, while you're here and on good wifi:

```bash
ollama pull qwen3:32b
```

Check your hardware first, because this one is demanding: a **20GB download** (four times everything above combined), roughly **24GB of free memory** to run, and about **17 seconds per answer** against under one second for `llama3.2`. On a 16GB machine it will not load at all.

Skipping it costs you nothing. The demo runs `llama3.2` regardless, and `dotnet run -- --model qwen3:32b` is how you'd switch if you did pull it.

**Either way, please don't pull this at the venue.** Twenty gigabytes times a roomful of people is how the wifi dies for everyone, demos included.

Heads up: we'll re-verify these model choices in the weeks before the event, since small-model quality moves fast. Check this file the week before the workshop in case a model name changed.

## 2. Verify It Works

With Ollama running, this should return a chat response:

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [{ "role": "user", "content": "Say hello in five words." }],
  "stream": false
}'
```

If you got JSON back with a message in it, you're done with the required pre-work.

## 3. An Editor That Runs `.http` Files

Every lab ships its requests as `.http` files, and the quickest way to run them is [VS Code](https://code.visualstudio.com/) with the [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client) extension (`humao.rest-client`): open the file, click "Send Request" above a request, read the response in a side pane. JetBrains IDEs open the same files with their built-in HTTP client, so if that's home for you, nothing to install.

Add the extension for whatever language you'll write your own code in: [C# Dev Kit](https://marketplace.visualstudio.com/items?itemName=ms-dotnettools.csdevkit) for the .NET path, [Python](https://marketplace.visualstudio.com/items?itemName=ms-python.python), or nothing extra for TypeScript, which VS Code supports out of the box. Then clone the repo:

```bash
git clone https://github.com/trailheadtechnology/10-ai-features-workshop.git
```

If you'd rather use curl or your language's HTTP client, that works too; every `.http` request is plain enough to port by hand.

## 4. Azure OpenAI: Nothing to Do

Three features (sentiment comparison, RAG generation, and the agent capstone) use frontier models on Microsoft Foundry, through a deployment the instructor runs; the requests go to the Azure OpenAI API, which is why the endpoint below ends in `openai.azure.com`. There is nothing to sign up for and nothing to pay for. The endpoint and deployment names are already in the lab files; the one thing you get in the room is the API key, which goes wherever a file says `<KEY FROM INSTRUCTOR>`:

- Endpoint: `https://trailhead-ai-workshop.openai.azure.com`
- Deployments: `gpt-4.1` (features 03, 05, and 07's cloud steps) and `gpt-5.5` (feature 10, the capstone)
- Key: `<KEY FROM INSTRUCTOR>`, handed out on the day and revoked afterward

If you are working from the .NET, Python, or TypeScript starters, the same three values go in `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_DEPLOYMENT`, and `AZURE_OPENAI_KEY`.

## 5. Optional: The .NET, Python, or TypeScript Path

The instructor demos in C#, and every feature ships `starter/` and `complete/` projects in three stacks if you want to follow along in code rather than from the `.http` files:

- **.NET**: install the current .NET SDK ([dot.net](https://dot.net); the repo pins the SDK in `global.json`) and confirm `dotnet --version` runs. `dotnet build workshop.slnx` at the repo root builds all twenty projects once, so the first `dotnet run` in the room is not a compile wait.
- **Python**: Python 3.11 or newer. Each feature's `python/requirements.txt` is one line (`openai`). Modern macOS and Linux Pythons refuse `pip install` outside a virtual environment, so the setup is `python3 -m venv .venv`, `source .venv/bin/activate`, `pip install -r requirements.txt` in the feature's `python/` folder, and every Python walkthrough repeats it.
- **TypeScript**: Node 20 or newer. Each feature's `typescript/` folder has a `package.json`; `npm install` there pulls `openai` and `tsx`, and `npm run complete` runs the demo with no build step.

The Python and TypeScript versions use the official `openai` package against Ollama's OpenAI-compatible endpoint, so they need nothing beyond the Ollama install in section 1. If you'd rather work in Java, Go, or anything else with an HTTP client, skip this entirely. The labs don't require any of the three.

## If You Did None of This Before Arriving

You'll still be fine. We bring the Ollama models on USB drives for local copying, the venue keys are handed out either way, and every lab can be completed with nothing but the provided `.http` request files and any HTTP client you already have. You'll just spend the first coffee break copying models instead of chatting, which is a fate we'd like to help you avoid.
