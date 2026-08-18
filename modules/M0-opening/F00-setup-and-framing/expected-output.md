# What Passing Looks Like

You're checking shape rather than exact text, because model wording varies between runs but the structure does not.

## 1. Local Chat

JSON with a `message` object whose `content` contains `TRAILHEAD OK`:

```json
{
  "model": "llama3.2",
  "message": { "role": "assistant", "content": "TRAILHEAD OK" },
  "done": true
}
```

(Extra fields like `total_duration` are fine. Small models sometimes add a polite sentence around the phrase; that still passes.)

## 2. Local Embeddings

JSON with an `embeddings` array containing one array of a few hundred floats:

```json
{
  "model": "nomic-embed-text",
  "embeddings": [[0.0123, -0.0456, 0.0789, "... 768 numbers ..."]]
}
```

If you get numbers, any numbers, you pass. Nobody reads embeddings by eye; that's the whole point of feature 08.

## 3. Cloud Chat

JSON with a `choices` array whose first message contains `TRAILHEAD CLOUD OK`:

```json
{
  "choices": [
    { "message": { "role": "assistant", "content": "TRAILHEAD CLOUD OK" }, "finish_reason": "stop" }
  ]
}
```

The `model` field in the response names the deployment's underlying model (for the workshop deployment, `gpt-4.1-2025-04-14`), and `usage` shows a handful of tokens each way. A `401` means the key was mistyped. A DNS error means the ENDPOINT placeholder is still in the URL.
