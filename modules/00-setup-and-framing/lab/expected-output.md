# What passing looks like

You're checking shape, not exact text. Model output wording can vary; the structure can't.

## 1. Local chat

JSON with a `message` object whose `content` contains `TRAILHEAD OK`:

```json
{
  "model": "llama3.2",
  "message": { "role": "assistant", "content": "TRAILHEAD OK" },
  "done": true
}
```

(Extra fields like `total_duration` are fine. Small models sometimes add a polite sentence around the phrase; that still passes.)

## 2. Local embeddings

JSON with an `embeddings` array containing one array of a few hundred floats:

```json
{
  "model": "nomic-embed-text",
  "embeddings": [[0.0123, -0.0456, 0.0789, "... 768 numbers ..."]]
}
```

If you get numbers, any numbers, you pass. Nobody reads embeddings by eye; that's the whole point of module 08.

## 3. Cloud chat

JSON with a `choices` array whose first message contains `TRAILHEAD CLOUD OK`:

```json
{
  "choices": [
    { "message": { "role": "assistant", "content": "TRAILHEAD CLOUD OK" }, "finish_reason": "stop" }
  ]
}
```

A `401` means the key was mistyped. A DNS error means the ENDPOINT placeholder is still in the URL.
