# HTTP Walkthrough for 03 Sentiment

The lab steps in [`../F03-lab.md`](../F03-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: five requests against Ollama. Two easy reviews and two hard ones through `phi3`, then the review phi3 misses rerun on `llama3.2`.
- `azure.http`: the same prompt, byte for byte, against Azure OpenAI. The endpoint and deployment are filled in; paste the key handed out in the room over `<KEY FROM INSTRUCTOR>`.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Requests 1 and 2, the Easy Set on phi3

Send requests 1 and 2 and check the labels against `data/reference-labels.json`. Then copy request 1 and swap the review text for the other eight in `data/easy.jsonl` (or loop the same body in your language). Keep the prompt's four line breaks exactly as they are: reflowing it onto one line costs `phi3` measured accuracy.

Check: your score for `phi3` on the easy set. Recorded: 9/10.

### Lab Step 2: Requests 3 to 5 and Azure.http, the Hard Set on Both Models

Requests 3 and 4 are hard-set reviews on `phi3`; request 5 is request 4's review on `llama3.2`, the offline stand-in for the big model, so a disagreement shows up without leaving localhost. With the card values in `azure.http`, run its three requests for the real comparison, then extend to the rest of `data/hard.jsonl` the same way as step 1.

Check: a label per review per model.

### Lab Step 3: The Disagreement List

Write down every review where the two models disagree, with the reference label and your call on who was right. That list is the deliverable.

Check: your version of the two tables in [`../expected-output.md`](../expected-output.md). Recorded: `phi3` 7/10 on the hard set, `gpt-4.1` 10/10, and the local stand-in 7/10.

### Stretch

Change the prompt to ask for `{"overall": ..., "aspects": {"comfort": ..., "durability": ..., "price": ...}}` and add a `format` schema for it, then see which model can go deeper than one label.
