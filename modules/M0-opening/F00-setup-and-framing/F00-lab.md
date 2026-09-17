# Lab 00: Environment Check

*Everyone does this lab, in [Module 0](../M0-overview.md), before anything else. It takes about ten minutes. Every other lab today needs this one to have worked. Do not skip it and do not save it for later.*

- **Goal:** prove your laptop can talk to three things: a chat model running on your machine, an embedding model running on your machine, and the workshop's Azure OpenAI endpoint in the cloud.
- **Input:** `http/smoke-test.http`, three requests, and the API key handed out in the room.
- **How:** send the three requests in order. Each one gets a JSON response back. Each one proves that one piece a later feature needs is working.
- **Model:** `llama3.2` and `nomic-embed-text` on Ollama, both local, no key. `gpt-4.1` on Azure OpenAI, with the key from the instructor.

**The User Problem:** the most expensive twenty minutes of any hands-on workshop is the twenty minutes in Module 1 when a third of the room discovers their environment does not work. This lab spends that time up front instead: everyone runs one smoke test, broken setups surface while there is still slack to fix them, and the room starts Module 1 together.

Every model today is an HTTP endpoint. Ollama is a local server on `localhost:11434` (native `/api/chat` and `/api/embed`, plus an OpenAI-compatible `/v1`); Azure OpenAI serves `gpt-4.1` and `gpt-5.5` at `openai.azure.com/openai/deployments/<name>/chat/completions` with an `api-key` header. Same request shape both places, so if your language can make an HTTP request, you are equipped, and the SDKs used later today are wrappers over these same calls.

There is no code to write. Every step below is one request to send and one thing to look for in the response. You can send the requests from VS Code with the REST Client extension, from a JetBrains IDE, or with the curl command shown in each step. All three ways send exactly the same request.

### Step 1: Open the request file

1. Open `http/smoke-test.http` in your editor. This file ships with the workshop and was written by hand for this lab. It holds three plain HTTP requests, one per model you need today. Each request is a block: a `###` comment line, then the method and URL, then headers, then a JSON body. The bodies are the same ones the curl commands below send. The file is the whole request; nothing sits behind it.
2. Make sure Ollama is running. On macOS and Windows, look for the Ollama icon in the menu bar or system tray. On Linux, run `ollama serve` in a terminal if nothing is already listening on port 11434.
3. If you use curl instead of the editor, open a terminal in this folder.

**Check:** the file shows three requests separated by `###` lines. Request 3 has the placeholder `<KEY FROM INSTRUCTOR>` in its `api-key` header.

### Step 2: Local chat

This request proves three things at once: Ollama is installed, it is running, and the `llama3.2` model is downloaded. Modules 1 and 3 use this model.

1. Send request 1 in `http/smoke-test.http`. It is a `POST` to `http://localhost:11434/api/chat` with model `llama3.2`, one user message, and `"stream": false`.
2. The curl version is the same request:

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD OK" }],
  "stream": false
}'
```

3. Read the response. Find the `message` object and read its `content` field.

**Check:** JSON with `"done": true` and a `message.content` that contains `TRAILHEAD OK`. A small model sometimes wraps the phrase in a polite sentence. That still passes. Compare against `expected-output.md`, which shows a trimmed example of a passing response for each of the three requests and names the fields to look for. Shape is what matters there, not exact wording.

**If it fails:** a "connection refused" error means Ollama is not running. Start the Ollama app and send again. An error that names the model means the model is not downloaded. Run `ollama pull llama3.2`, or copy it from a helper's USB drive, then send again.

### Step 3: Local embeddings

This request proves the embedding model is downloaded. Module 2 and the anomaly detection feature use this model.

1. Send request 2 in `http/smoke-test.http`. It is a `POST` to `http://localhost:11434/api/embed` with model `nomic-embed-text` and one string as `input`.
2. The curl version is the same request:

```bash
curl http://localhost:11434/api/embed -d '{
  "model": "nomic-embed-text",
  "input": "A quiet trail along the lake shore."
}'
```

3. Read the response. Find the `embeddings` field.

**Check:** JSON with an `embeddings` array that holds one inner array of 768 floating point numbers. Any numbers pass. Nobody reads embeddings by eye.

**If it fails:** an error that names the model means the model is not downloaded. Run `ollama pull nomic-embed-text` and send again.

### Step 4: Cloud chat

This request proves you can reach the Azure OpenAI deployment using the key from the room. Sentiment comparison, RAG generation, and the capstone use this model.

1. Get the key from the whiteboard or the card on your table.
2. In `http/smoke-test.http`, find request 3. Replace `<KEY FROM INSTRUCTOR>` in its `api-key` header with the key. Leave the endpoint and the deployment name `gpt-4.1` alone. They are already correct.
3. Send request 3. It is a `POST` to `https://trailhead-ai-workshop.openai.azure.com/openai/deployments/gpt-4.1/chat/completions?api-version=2024-10-21` with one user message.
4. The curl version is the same request. Paste the key over `<KEY FROM INSTRUCTOR>` here too:

```bash
curl "https://trailhead-ai-workshop.openai.azure.com/openai/deployments/gpt-4.1/chat/completions?api-version=2024-10-21" \
  -H "Content-Type: application/json" \
  -H "api-key: <KEY FROM INSTRUCTOR>" \
  -d '{ "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD CLOUD OK" }] }'
```

5. Read the response. Find the `choices` array, take its first element, and read `message.content`.

**Check:** JSON with a `choices` array whose first `message.content` contains `TRAILHEAD CLOUD OK`. The `model` field says `gpt-4.1-2025-04-14`. The `usage` object shows a handful of tokens each way.

**If it fails:** a `401` means the key was typed wrong. The keys are long, so type it again and send again. A DNS error means the URL still has a placeholder in it. A timeout usually means you are on the venue guest network. Switch to the workshop network printed on your card.

### Step 5: Raise a hand if anything is still red

1. Count your passes. You need three.
2. If all three passed, your machine is ready for all ten features. Do a stretch goal or help the person next to you.
3. If any request still fails after you tried its fix, flag a helper now, during the opening. We have backups ready for exactly this moment: USB copies of the models and a shared endpoint. Module 1 starts on time either way, so do not wait until then to say something.

**Check:** three JSON responses and no red text in your editor or terminal.

### Stretch goals

Pick any. None of them is needed to pass this lab.

- **Say the day's thesis out loud.** Every company right now is asking "where can we add AI?" This workshop spends the day practicing the better question: "what problems can AI best solve for my users?" Each of the ten features that follow opens with a user who is stuck, and the AI only shows up as the answer to that user's problem. "Users" is defined broadly on purpose: the product manager who cannot read every review counts, and so does the ranger staring down a full inbox. Three of the ten features (03, 07, and 08) solve problems for the people running the product rather than the people using it. Spotting problems early is a user problem too; it just belongs to a user on your payroll. **Check:** you can name, in one sentence, the user and the problem for the feature you are about to try next.

- **Read one trip report.** Open `../../M1-understanding/F01-summarization/data/` and read one `tr-*.md` file from start to finish, granola recipes and all. That folder holds 40 fictional trip reports for Trailhead Guides, the made-up park app every feature today uses. Each file is a short YAML front matter block (`id`, `author`, `date`, `park`) followed by a rambling Markdown trip diary of about a thousand words. They ship with the workshop corpus; there is no script that builds them. Feature 01 exists because nobody wants to read forty of those. **Check:** you can say in one sentence what a hiker needs to know from it, and you can say how long it took you to find that sentence.
- **Confirm the third model.** `SETUP.md` also asks you to pull `phi3`. Feature 03 compares it against `llama3.2` and `gpt-4.1`. Run `ollama list` in a terminal. **Check:** `llama3.2`, `nomic-embed-text`, and `phi3` all appear. If `phi3` is missing, run `ollama pull phi3` now, while the workshop wifi is still quiet.
- **Set the Azure environment variables.** If you plan to write code in .NET, Python, or TypeScript later today, those code tracks read the cloud settings from environment variables, not from the `.http` file. Set `AZURE_OPENAI_ENDPOINT` to `https://trailhead-ai-workshop.openai.azure.com`, `AZURE_OPENAI_DEPLOYMENT` to `gpt-4.1`, and `AZURE_OPENAI_KEY` to the key from the room. **Check:** printing each variable in a new terminal shows the value you set.

## Pick a Track

This lab has only one track. Every language sends the same three requests to the same two servers and checks the answers against the same [`expected-output.md`](expected-output.md).

| Track | Start here | What you run |
|---|---|---|
| VS Code REST Client or JetBrains HTTP Client | `http/smoke-test.http` | the three requests in the file, in order |
| curl, or any HTTP client in your language | this page | the three curl commands above, or a port of them |

## What Is in This Folder

- `http/smoke-test.http`: the three hand-written requests, in a format that VS Code's REST Client extension and JetBrains IDEs can run directly. Request 3 needs the key pasted in first.
- `expected-output.md`: a trimmed passing response for each of the three requests, and what the common failures mean.
