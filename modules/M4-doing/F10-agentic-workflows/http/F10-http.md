# HTTP Walkthrough for 10 Agentic Workflows

The lab steps in [`../F10-lab.md`](../F10-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `azure.http`: the tool-calling round-trip against Azure OpenAI, one request per turn, with you playing the loop. Step 1 is fully written out; step 2 leaves a marked hole for your `get_weather` definition; step 3 has all five tools and the closed-trail request; the stretch is the permit gate. Every tool result you paste is the real output of the demo's tools over `../data/`. There is no `ollama.http` for this feature: the loop mechanics are identical, but the lab runs on the cloud model on purpose.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

In this file you are the agent loop: the model asks for a tool, you copy the call out of the response, "run" the tool by pasting the canned result, send it back, repeat until you get an itinerary. That copy-paste is what every `complete/` project automates, and doing it once by hand is the fastest way to see what agent frameworks hide. Paste the key handed out in the room over `<KEY FROM INSTRUCTOR>` first; the endpoint and deployment are already filled in.

### Lab Step 1: Requests 1a to 1c, the Two-Tool Round-Trip

Send 1a. Expect `finish_reason: "tool_calls"` and a `tool_calls` entry naming `search_trails` or `check_campsites`; if you get an itinerary instead, the model never saw your tools. Copy the assistant message out of your response into 1b (its `id` goes where `CALL-ID` is), keep the tool message that hands back the real search result, send. Repeat with 1c for the second tool.

Check: an itinerary naming real trails and real campgrounds from the payloads you handed back. Invented names mean a tool result did not reach the model.

### Lab Step 2: Request 2, Add Get_weather

Write the `get_weather` definition yourself (name, description, a `park` parameter) over the marker, run the same loop, and hand back the forecast in the comment when it asks. September 16 is the rain day.

Check: the forecast shapes the plan; the hard day lands on the 14th or 15th and the 16th gets something short, with a sentence saying why. Compare [`../expected-output.md`](../expected-output.md).

### Lab Step 3: Request 3, the Washed-Out Bridge

All five tools, a request that names `trail-0117`. When the model asks for `get_trail_conditions` with that id, hand back the four reports in the comment.

Check: the itinerary drops or flags the trail, for a reason that came from the tool result. The three failure modes to watch for are listed in `../expected-output.md`, in ascending order of embarrassment.

### Stretch: The Human Gate

Do not send the last request until you have said yes out loud. Read the arguments the model chose for `request_permit`; that is the summary a human approves. Decline and send the cancelled result instead, and watch what the model does. You are also the step budget: decide up front how many round-trips you will do before you stop.
