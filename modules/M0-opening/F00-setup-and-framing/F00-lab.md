# Lab 00 · Environment Check

Three requests, three JSON responses, and you're cleared for the whole day. Run them in order; each one proves a different dependency that later features rely on.

- **Goal:** prove your machine can reach a local model, an embedding model, and the venue's Azure OpenAI endpoint.
- **How:** run the three requests in `http/smoke-test.http` (or the curl equivalents below), in order:
  1. Chat completion against Ollama (`llama3.2`), which proves Ollama is installed, running, and the model is pulled.
  2. Embedding against Ollama (`nomic-embed-text`), which proves the embedding model that powers Modules 2 and 3 is ready.
  3. Chat completion against Azure OpenAI using the key handed out at the door, which proves the cloud path for sentiment, RAG, and the capstone.
- **Success check:** three JSON responses, no red text. Compare `expected-output.md`.
- **If something fails:** flag a helper now. Fallbacks (USB model copies, shared endpoints) exist precisely for this moment. Do not wait until Module 1 to mention it.
- **Stretch goal:** finished early? Pull up `modules/M1-understanding/F01-summarization/data/trip-reports/` and skim one trip report end to end. Feel the problem feature 01 is about to solve.

The exact requests live in `http/smoke-test.http` (VS Code REST Client / JetBrains HTTP Client format). The curl equivalents are below.

## 1. Local Chat (Powers Modules 1 and 3)

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD OK" }],
  "stream": false
}'
```

**Pass:** JSON containing a message with `TRAILHEAD OK`.
**Fail fixes:** Ollama not running → start the Ollama app. Model missing → `ollama pull llama3.2` (or grab it from a helper's USB drive).

## 2. Local Embeddings (Powers Module 2 and Anomaly Detection)

```bash
curl http://localhost:11434/api/embed -d '{
  "model": "nomic-embed-text",
  "input": "A quiet trail along the lake shore."
}'
```

**Pass:** JSON containing an `embeddings` array of numbers.
**Fail fix:** `ollama pull nomic-embed-text`.

## 3. Cloud Chat (Powers Sentiment Comparison, RAG Generation, and the Capstone)

The endpoint and deployment name are already filled in below; paste the key handed out in the room over `<KEY FROM INSTRUCTOR>`.

```bash
curl "https://trailhead-ai-workshop.openai.azure.com/openai/deployments/gpt-4.1/chat/completions?api-version=2024-10-21" \
  -H "Content-Type: application/json" \
  -H "api-key: <KEY FROM INSTRUCTOR>" \
  -d '{ "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD CLOUD OK" }] }'
```

**Pass:** JSON containing `TRAILHEAD CLOUD OK`.
**Fail fixes:** 401 → re-type the key (they're long; typos happen). Timeout → you're probably on the venue guest network; switch to the workshop network on your card.

## Done?

All three passed: you're fully provisioned for all ten features. Stretch goal: skim one trip report from feature 01's `data/trip-reports/` end to end, granola recipes and all. Feature 01 exists because nobody wants to read forty of those.

Something failed and the fixes above didn't help: flag a helper **now**, during the opening. This is the moment we've reserved for fixing environments; Module 1 starts on schedule either way.
