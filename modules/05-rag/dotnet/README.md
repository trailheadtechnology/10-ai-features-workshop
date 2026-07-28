# .NET demo for 05 RAG

Two console projects, both built on Microsoft.Extensions.AI:

- `starter/`: the demo's starting point. One `IChatClient`, one question, no context. It produces the confident wrong answer that puts Sperry Chalet in the San Gabriel Mountains of California.
- `complete/`: the finished demo as shown on stage. Embeds `../../lab/chunks.jsonl` with `nomic-embed-text`, retrieves top-k by cosine similarity, and generates a cited answer from the retrieved context.

Retrieval always runs locally. Generation goes to Azure OpenAI when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are all set, and falls back to local `llama3.2` when they are not, printing which one it picked. That fallback is the whole of step 7 in the demo outline: one client construction changes, the prompt and the retrieval do not.

From `starter/`:

```bash
dotnet run                                  # the Sperry question, no context, no receipts
dotnet run -- "Is the Avalanche Lake Trail open right now?"
```

From `complete/`:

```bash
dotnet run                                                    # the Sperry question, grounded and cited
dotnet run -- "What is the maximum group size on a Glacier backcountry permit?"
dotnet run -- "Are there EV charging stations in Glacier National Park?"   # expect a refusal
dotnet run -- --no-context                                    # step 1 of the demo, inside the finished app
dotnet run -- --top-k 8                                       # stretch goal: watch the citation degrade
```

The first run of `complete/` embeds all 241 chunks (roughly 40 seconds) and caches the vectors to `complete/embeddings.json`. Every run after that is instant. Delete that file if you change `chunks.jsonl`.

Both projects run fully offline against Ollama. See [../lab/expected-output.md](../lab/expected-output.md) for what the answers actually look like.
