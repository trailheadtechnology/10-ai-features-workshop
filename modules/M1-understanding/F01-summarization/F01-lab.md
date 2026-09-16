# Lab 01: Summarization

*This is the Recommended lab for [Module 1](../M1-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** `data/tr-0001.md`, a clean report (mud and crowds, no closure), and `data/tr-0004.md`, the report with the washed-out footbridge buried mid-text. The other 38 in `data/` are the rest of the Avalanche Lake corpus.
- **How:** one chat call per report. Read the file, put an instruction above its text, send the whole thing to Ollama as a single user message, and print the reply. `http/ollama.http` holds all three requests with the reports pasted in; the code tracks read the same files from `data/`.
- **Model:** `llama3.2`, local. No key.

Each step below is one thing to make the program do. Every code track's `starter/` already reads a report and sends the naive prompt. Edit it until it does all three steps. Look at `complete/` when you get stuck. If you are on the HTTP track, send the numbered request in `http/ollama.http` that each step names.

### Step 1: Run the starter and read the book report

The starter already makes the whole call. Read the code before you run it, so you know what you are about to change. Here is what it does:

1. Open `../../data/tr-0001.md` and read the whole file into one string. The file is one of the 40 trip reports in `data/`, `tr-0001.md` through `tr-0040.md`, that make up the Avalanche Lake corpus of the fictional Trailhead Guides app. Each report is a Markdown file: a YAML front matter block with `id`, `author`, `date`, and `park`, then a title line and 600 to 1,000 words of first-person prose in a deliberately inconsistent style, where trail conditions and hazards appear only in the prose and never in a field. The reports are synthetic, written for this workshop to the corpus contract described in the repo's [`README.md`](../../../README.md). No build script ships with them, and no real trail day is behind any of them. (HTTP track: request 1 in `http/ollama.http` already has the report text pasted in. That file holds the lab's three requests, one per step, with the report bodies inlined and the prompts kept byte-identical to the ones in `complete/`, so it runs as-is against a local Ollama.)
2. Strip the front matter. The file starts with a block between two `---` lines. That block holds `id`, `author`, `date`, and `park`. Split the text on `---` into three parts, keep the third part, and trim it. What is left starts at the report's title.
3. Build the prompt. It is the one-line instruction below, then a blank line, then the report text.

```text
Summarize this trip report.
```

4. Send the prompt as a single user message to `llama3.2` through your track's chat client. Whatever the client looks like, the call is the same shape everywhere: the model name, plus a list of messages holding one message whose role is `user` and whose content is the prompt. Nothing else goes in. No system message, no temperature, no streaming. (HTTP track: request 1 is exactly that body.)
5. Take the text of the reply message and print it.

Run it twice. The wording changes between runs. The shape does not.

**Check:** a paragraph or two about the author's gear and their day. Nothing a hiker planning Saturday could act on. The summary is faithful and useless. Right model, wrong instruction.

### Step 2: Replace the instruction with the briefing prompt

1. Keep everything from step 1 except the instruction line.
2. Replace `Summarize this trip report.` with the seven lines below. Keep the line breaks exactly as shown. If you reflow them, `llama3.2` behaves differently. (HTTP track: request 2 is this prompt on `tr-0001.md`.)

```text
You are helping a hiker planning to hike this trail within the next week.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.
```

3. The prompt is now those seven lines, a blank line, then the report text. Nothing else about the call changes. Same client, same model, same single user message.
4. Run it on `data/tr-0001.md` four or five times, not once. The checks below have to hold on every run.

**Check:** three bullets and nothing else. The gear debrief, which is most of the source text, is gone. The mud patches and the 10am crowds are in. The hazards bullet says "no closures or hazards reported". The only bear in `tr-0001.md` is a ranger's remark about a road near Lake McDonald, so a bullet that keeps it as a plain sighting is fine. A bullet that closes the trail because of the bear is the failure that the two lines beginning "Report only what the trip report states" exist to stop. The measured rate with and without those lines is in [`expected-output.md`](expected-output.md).

### Step 3: Run the same prompt on the buried-hazard report

1. Change the report path from `tr-0001.md` to `tr-0004.md`. Every starter takes the path as its one optional argument, so pass it on the command line (`dotnet run -- ../../data/tr-0004.md`, `uv run main.py ../../data/tr-0004.md`, `npm run starter -- ../data/tr-0004.md`) or change the default in the code. `tr-0004.md` has the same front matter and the same rambling shape as `tr-0001.md`; the difference is that its author buries a washed-out footbridge and a closed trail in the fourth paragraph. It was written that way on purpose, because a summarizer that misses it fails the feature. (HTTP track: request 3.)
2. Leave the step 2 prompt exactly as it is. Run it.
3. Run `tr-0001.md` once more with the same prompt. This confirms the clean report still passes.

**Check:** the first bullet is the washed-out footbridge and the closed trail. In the source that fact is buried in the fourth paragraph, between airport sandwiches and huckleberry ice cream. A second bullet reading "no other closures reported" is normal. Bullets about the sister, the deer, or Moby the rental SUV mean the prompt in your program is not the one above. `tr-0001.md` still comes back with nothing closed.

### Stretch goals

Pick any. Every code track's `complete/` already has each one built in, behind the flag named below.

- **Brief a different reader.** Add an audience variable to the program. When it is `ranger`, replace the first line of the step 2 prompt with the line below. Otherwise keep the hiker line. Run `data/tr-0004.md` through both. `complete/` does this with `--briefing --audience ranger`.

  ```text
  You are helping a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery.
  ```

  **Check:** the ranger version leads with where the bridge went out and the barricade; the hiker version keeps the crowding. Identical output means the audience line is not reaching the prompt.

- **Shrink the summary to a headline.** Same file, same call, different instruction. Replace the whole step 2 prompt with the three lines below and run `data/tr-0004.md`. `complete/` does this with `--headline`.

  ```text
  From the trip report below, write ONE line of at most 12 words,
  suitable for a status badge on a trail card in an app.
  Lead with the most important condition or closure. No preamble.
  ```

  **Check:** one line, at most 12 words, that leads with the closure. No bullets and no preamble. Only the instruction changed, and that is the point. A new spot in the UI costs a new prompt, not new infrastructure.

- **See the hallucination the grounding lines prevent.** Delete the two lines of the step 2 prompt that begin "Report only what the trip report states" and end with "when it says none." Run `data/tr-0001.md` ten or more times. Put the lines back when you are done.

  **Check:** some runs now invent a closure from the bear, the creek, or the word "avalanche" in the trail's name. Measured over 24 runs on `tr-0001.md`, the prompt without those lines asserted a hazard or closure the report never made 11 times (46%). With the lines, 1 of 24 (4%). `tr-0004.md` led with the bridge in 12 of 12 runs either way. The full story is in [`expected-output.md`](expected-output.md).

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F01-http.md`](http/F01-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F01-dotnet.md`](dotnet/F01-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F01-python.md`](python/F01-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F01-typescript.md`](typescript/F01-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `data/tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `data/tr-*.md`: all 40 Avalanche Lake trip reports, synthetic Markdown with a four-field front matter and a prose body; `tr-0001` and `tr-0004` are the lab's two named inputs
- `http/ollama.http`: the three lab requests with the report text pasted in, for the HTTP track
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
