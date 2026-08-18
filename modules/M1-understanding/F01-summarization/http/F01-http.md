# HTTP Walkthrough for 01 Summarization

The lab steps in [`../F01-lab.md`](../F01-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: three requests against Ollama (`llama3.2`), each with a trip report inlined.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Request 1, the Naive Prompt

Send request 1, which is "Summarize this trip report" on `tr-0001.md`, the clean report. What comes back is faithful, generic, and useless to a hiker, and that is the baseline you are improving on.

### Lab Step 2: Request 2, the Improved Prompt on the Same Report

Request 2 is the 3-bullet briefing prompt on the same report. Send it four or five times, not once: the point of the clean report is that there is no closure in it, and a prompt that demands a hazards bullet will invent one. If yours does, look at the last three lines of request 2's prompt, which give the model a legal way to report nothing, and make sure your own version keeps them.

Check: three bullets, and the hazards bullet says nothing is closed, every run.

### Lab Step 3: Request 3, the Buried-Hazard Report

Request 3 is the same improved prompt on `tr-0004.md`, where the washed-out footbridge is mentioned in passing halfway down.

Check: the first bullet is the closure. Compare the sample in [`../expected-output.md`](../expected-output.md).

### Stretch

Copy request 3 and change the first line of the prompt so it briefs a park ranger who cares about maintenance, closures, and safety incidents rather than a hiker. Same report, different audience, one line changed.
