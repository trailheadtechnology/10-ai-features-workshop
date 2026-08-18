# HTTP Walkthrough for 07 Classification & Routing

The lab steps in [`../F07-lab.md`](../F07-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: the classify request against Ollama (`llama3.2`), shown on three inlined messages: a routine conditions question, an emergency, and the ambiguous one. The `format` field is a JSON schema whose enum is the seven categories, so the model cannot invent a label.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Requests 1 to 3, Then All 20

Send the three requests: request 1 should come back `conditions`, request 2 `emergency`, and request 3 `unsure`. Then copy request 1 and swap the message text for each of the 20 in `data/inquiries-slice.jsonl` (or loop the body in your language) and score against `data/reference-labels.json`. Keep `temperature` at 0.

Check: overall accuracy, and separately, emergency recall. Recorded: 17/20 and 2/2.

### Lab Step 2: Fix a Miss by Editing a Description

`inq-0030` (a wedding photographer asking whether a special-use permit is required) lands in `general` because the `permit` description talks about reserving and paying. Widen the description by a clause and re-run. Two rules constrain any edit: the ordering paragraph that makes emergency win stays, and `unsure` stays narrow.

### Lab Step 3: The Success Check

Both emergencies classified `emergency`, `inq-0035` in `unsure` rather than confidently wrong, and your accuracy at or above where it started. Missing an emergency fails the lab at 19/20. Compare [`../expected-output.md`](../expected-output.md).

### Stretch

Add a `priority` field to the schema, or a confidence score and route anything below a threshold to `unsure`.
