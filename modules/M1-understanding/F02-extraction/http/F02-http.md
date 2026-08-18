# HTTP Walkthrough for 02 Extraction

The lab steps in [`../F02-lab.md`](../F02-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: two requests against Ollama (`llama3.2`), both carrying the JSON schema in the `format` field.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Request 1, the Fact-Rich Report

Send request 1. The `format` field is a JSON schema: nullable scalars, arrays, and a description on every field saying when to use `null`. The reply is JSON that matches it, no prose, no fences.

Check: every field populated, matching the `tr-0007.md` block in [`../expected-output.md`](../expected-output.md).

### Lab Step 2: Request 2, the Sparse Report

Send request 2 three or four times. `tr-0011.md` never names the trail and gives no distance, elevation, or exact date. Write down every field that came back with a value the report does not contain.

Check: most missing facts are `null`, and you can name the ones that were not. `../expected-output.md` records `elevation_gain_ft: 0` and `date_hiked: "early last month"`.

### Lab Step 3: Fix the Schema, Then Validate in Code

First tighten the descriptions in the schema (`"null, never 0, if the report gives no figure"` is the fix for the zero). Then, in your language, write the validator for what the schema cannot express: reject dates that do not parse in an explicit format, measurements of 0, and names that do not appear in the source text, and coerce each to `null`. The rules are in any of the `complete/` projects; the .NET one is [`../dotnet/complete/Program.cs`](../dotnet/complete/Program.cs).
