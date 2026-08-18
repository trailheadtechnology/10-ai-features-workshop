# HTTP Walkthrough for 09 Human-in-the-Loop

The lab steps in [`../F09-lab.md`](../F09-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: five requests against Ollama (`llama3.2`). Three everyday drafts (Mist Trail boilerplate, the Sperry campfire rules answer, the Avalanche Lake closure), then the emergency, then the same emergency with the escalation rule moved to the front of the prompt. Inquiry text and park-doc excerpt are inlined.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Requests 1 to 5, Read Them as an Editor

Send requests 1 to 3 and read the drafts: what would you change before this went out under your name? Then send request 4, the overdue-hiker voicemail. The system prompt tells the model to output `ESCALATE:` for emergencies. Watch what it does instead. Then request 5, the "repair" with the rule moved to the top of the prompt.

Check: three usable drafts, then a warm reassuring note to a woman whose husband is four hours overdue (recorded 3/3), and on request 5 `ESCALATE` followed by the note anyway (also 3/3). A prompt instruction is a request.

### Lab Step 2: The Policy Worksheet

Open [`../policy-worksheet.md`](../policy-worksheet.md). For each category, choose auto-send, draft-for-approval, or human-only, with one sentence of justification based on what a wrong answer costs and whether it can be undone. This is half the lab and it is judgment, not typing.

### Lab Step 3: Compare

The reference drafts and a reasoned reference policy are in [`../expected-output.md`](../expected-output.md). Your policy may differ; your justifications are what count. If you are working in code, the gate that keeps emergencies away from the model entirely, the review loop, and the audit log are what the `complete/` projects add on top of these five requests.

### Stretch

Compute the edit distance between a draft and your edited version, and sketch what threshold would earn a category promotion from draft-mode to auto-send.
