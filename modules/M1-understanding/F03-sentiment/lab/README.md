# Lab assets for 03 Sentiment

Everything the lab spec in [../README.md](../README.md) references:

- `ollama.http`: five ready-to-run requests against Ollama. Two easy reviews and two hard ones through `phi3`, then the review phi3 misses rerun on `llama3.2` so a disagreement shows up without leaving localhost. Review text is inlined, so each request runs as-is or ports to your language by copying the JSON body.
- `azure.http`: the same prompt, byte for byte, against Azure OpenAI. Fill in ENDPOINT, DEPLOYMENT, and YOUR-KEY from the card handed out at the door.
- `easy.jsonl`: 10 straightforward reviews from `data/gear-reviews.jsonl`. The text says what it means and the star rating agrees.
- `hard.jsonl`: 10 reviews where the text and the rating fight. Sarcasm ("Absolutely love it when the mesh blew out"), five stars aimed at a return process, two stars aimed at an instruction manual, one star aimed at an ex-partner.
- `reference-labels.json`: hand labels for all 20, `positive | negative | mixed`, with a one-phrase rationale on each hard case explaining what the rating is really about.
- `expected-output.md`: real measured accuracy for both models on both sets, the honest disagreement list, and one finding about prompt formatting that nobody went looking for.

The reviews keep their original ids, so any of them can be traced back to the full corpus.
