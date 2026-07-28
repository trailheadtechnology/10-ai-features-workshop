# Lab 00 · Environment check

Three requests, three JSON responses, and you're cleared for the whole day. Run them in order; each one proves a different dependency that later features rely on.

The exact requests live in `smoke-test.http` (VS Code REST Client / JetBrains HTTP Client format; the file lands with the code build-out). The curl equivalents are below.

## 1. Local chat (powers Modules 1 and 3)

```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD OK" }],
  "stream": false
}'
```

**Pass:** JSON containing a message with `TRAILHEAD OK`.
**Fail fixes:** Ollama not running → start the Ollama app. Model missing → `ollama pull llama3.2` (or grab it from a helper's USB drive).

## 2. Local embeddings (powers Module 2 and anomaly detection)

```bash
curl http://localhost:11434/api/embed -d '{
  "model": "nomic-embed-text",
  "input": "A quiet trail along the lake shore."
}'
```

**Pass:** JSON containing an `embeddings` array of numbers.
**Fail fix:** `ollama pull nomic-embed-text`.

## 3. Cloud chat (powers sentiment comparison, RAG generation, and the capstone)

Use the endpoint, deployment name, and API key from the card handed out at the door.

```bash
curl "https://ENDPOINT/openai/deployments/DEPLOYMENT/chat/completions?api-version=2024-10-21" \
  -H "Content-Type: application/json" \
  -H "api-key: YOUR-KEY" \
  -d '{ "messages": [{ "role": "user", "content": "Reply with exactly: TRAILHEAD CLOUD OK" }] }'
```

**Pass:** JSON containing `TRAILHEAD CLOUD OK`.
**Fail fixes:** 401 → re-type the key (they're long; typos happen). Timeout → you're probably on the venue guest network; switch to the workshop network on your card.

## Done?

All three passed: you're fully provisioned for all ten features. Stretch goal: open [`data/`](../../../../data/corpus.md) and skim one trip report from `trip-reports/` end to end, granola recipes and all. Feature 01 exists because nobody wants to read forty of those.

Something failed and the fixes above didn't help: flag a helper **now**, during the opening. This is the moment we've reserved for fixing environments; Module 1 starts on schedule either way.
