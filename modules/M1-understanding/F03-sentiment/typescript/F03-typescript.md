# TypeScript Demo for 03 Sentiment

Two scripts, both reading from [`../data/`](../data/):

- `starter/index.ts`: one client, one `classify` function, one review (`gr-0007`, the sarcastic two-star). Prints the review and what `phi3` says.
- `complete/index.ts`: the finished demo. Both review sets through both models with the byte-identical four-line prompt, a table with disagreements flagged, accuracy per set, and the disagreement list with a verdict on who was right.

Setup once (`npm install` in this directory), then:

```bash
npm run complete             # both sets, both models
npm run complete -- --easy   # easy set only (lab steps 3-4 shape)
npm run complete -- --hard   # hard set only (lab step 5 shape)
```

Set `AZURE_OPENAI_ENDPOINT`, `AZURE_OPENAI_KEY`, `AZURE_OPENAI_DEPLOYMENT` (endpoint `https://trailhead-ai-workshop.openai.azure.com`, deployment `gpt-4.1`, key handed out in the room) to use Azure OpenAI as the big model via the SDK's `AzureOpenAI` client; leave them unset and `llama3.2` on Ollama stands in, so the whole comparison runs offline. The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`); swapping providers later is a different constructor. `tsx` runs the `.ts` files directly, no build step.

## Lab Walkthrough: From `starter/` to `complete/`

The steps are in [`../F03-lab.md`](../F03-lab.md); this maps each onto `starter/index.ts` → `complete/index.ts`. Edit the starter in place (or copy it first); `complete/`'s comments say why each piece is there.

### Lab step 0: run the starter

```bash
npm run starter
```

Check: `phi3 says: negative` on `gr-0007`.

### Lab steps 1-2: load reference labels, loop the easy set, score it

```typescript
const labels: Record<string, { label: string }> = JSON.parse(readFileSync(resolve(DATA, "reference-labels.json"), "utf8"));
let correct = 0, total = 0;
for (const line of lines("easy.jsonl")) {
  const review: Review = JSON.parse(line);
  const label = await classify(client, "phi3", review.text);
  const reference = labels[review.id].label;
  console.log(`${review.id.padEnd(9)} ${reference.padEnd(10)} ${label.padEnd(10)}`);
  total++; if (label === reference) correct++;
}
console.log(`phi3 ${correct}/${total}`);
```

Check: 9/10 on the easy set in the recorded runs; yours may differ by one.

### Lab steps 3-4: second client, hard set through both models

```typescript
const big: Target = { client: new AzureOpenAI({ endpoint, apiKey: key, apiVersion: "2024-10-21", deployment }), model: deployment };
// or, offline: { client: ollama, model: "llama3.2" }
const small = await classify(smallTarget, review.text);
const bigLabel = await classify(big, review.text);
```

Check: 7/10 for `phi3` on the hard set, 10/10 for `gpt-4.1`, 8/10 for the `llama3.2` stand-in.

### Lab step 5: disagreement list

```typescript
for (const d of results.filter((r) => r.small !== r.big)) {
  const verdict = d.big === d.reference ? "big right" : d.small === d.reference ? "phi3 right" : "both wrong";
  console.log(`${d.review.id} [${d.set}] ref=${d.reference} phi3=${d.small} big=${d.big}  (${verdict})`);
}
```

Check: your version of the two tables in `../expected-output.md`. Stretch: change the label to `{overall, aspects: {comfort, durability, price}}` with structured output.
