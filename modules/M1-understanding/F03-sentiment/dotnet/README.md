# .NET demo for 03 Sentiment

Two console projects, both built on Microsoft.Extensions.AI:

- `starter/`: the demo's starting point. One `IChatClient` over Ollama (`phi3`), one `Classify` method, one review, one word back.
- `complete/`: the finished demo. Both sets through both models, with a label table, accuracy per set per model, and the disagreement list.

From `starter/`:

```bash
dotnet run                # gr-0007, the sarcastic trail runner review
dotnet run -- gr-0034     # any id from ../../lab/easy.jsonl or hard.jsonl
```

From `complete/`:

```bash
dotnet run                # both sets, both models, table + accuracy + disagreements
dotnet run -- --easy      # easy set only (demo steps 3 and 4)
dotnet run -- --hard      # hard set only (demo step 5)
```

## The two models

`complete/` builds two `IChatClient`s and then never mentions a provider again. The first is `phi3` on Ollama. The second is Azure OpenAI when `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, and `AZURE_OPENAI_DEPLOYMENT` are all set:

```bash
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
export AZURE_OPENAI_KEY=...
export AZURE_OPENAI_DEPLOYMENT=gpt-4o-mini
```

With those unset, the app prints a note and falls back to `llama3.2` on Ollama as a stand-in, so the whole comparison runs offline. That fallback is a few lines around one constructor, and everything downstream, the classify method included, is untouched. Point at those lines on stage: it is the provider-flexibility claim made real, and the same swap works in the other direction when you move to production.

Both models get a byte-identical prompt. That is deliberate and it is load-bearing. The same prompt reflowed onto one line costs `phi3` a measurable amount of accuracy while leaving `llama3.2` unchanged, so a comparison that varies prompt shape along with the model is measuring noise. See [../lab/expected-output.md](../lab/expected-output.md).
