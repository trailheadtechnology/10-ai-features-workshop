# Lab 01: Summarization

*This is the Recommended lab for [Module 1](../M1-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** `data/tr-0001.md`, a clean report (mud and crowds, no closure), and `data/tr-0004.md`, the report with the washed-out footbridge buried mid-text. The other 38 files in `data/` are the rest of the synthetic Avalanche Lake corpus described in the repo's [README.md](../../../README.md); no build script produced them.
- **How:** one chat call per report. Read the file, put an instruction above its text, send the whole thing to Ollama as a single user message, and print the reply.
- **Model:** `llama3.2`, local. No key.

Each step below is one thing to make the program do. Every track's `starter/` already reads a report and sends the naive prompt. Edit it until it does all three steps. Look at `complete/` when you get stuck.

### Step 1: Run the starter and read the book report

**Do:**
1. Open `../../data/tr-0001.md` and read the whole file into one string.
2. Strip the front matter: split on the `---` lines, keep the third part, trim it. What is left starts at the report's title.
3. Build the prompt: the one line below, a blank line, then the report text.

   ```text
   Summarize this trip report.
   ```

4. Send the prompt as a single user message to `llama3.2` through your track's chat client: the model name, plus one message with role `user` and that content. No system message, no temperature, no streaming.
5. Print the reply text.

**Why:** this is the naive version on purpose. It stays in the lab as the baseline you visibly improve on in step 2.

Run it twice. The wording changes between runs; the shape does not.

**Check:** a paragraph or two about the author's gear and their day. Nothing a hiker planning Saturday could act on. The summary is faithful and useless. Right model, wrong instruction.

### Step 2: Replace the instruction with the briefing prompt

**Do:**
1. Keep everything from step 1 except the instruction line.
2. Replace `Summarize this trip report.` with the seven lines below, line breaks exactly as shown.

   ```text
   You are helping a hiker planning to hike this trail within the next week.
   From the trip report below, produce exactly 3 bullets covering:
   current trail conditions, hazards or closures, and crowding.
   Ignore gear talk, personal stories, and scenery.
   Report only what the trip report states. Do not turn a wildlife sighting into a
   hazard or a closure, and write "no closures or hazards reported" when it says none.
   If the report does state a closure or hazard, it must appear in the first bullet.
   ```

3. Nothing else about the call changes: same client, same model, same single user message.
4. Run it on `data/tr-0001.md` four or five times, not once. The check below has to hold on every run.

**Why:** the reflowed prompt behaves differently, so keep the line breaks. The last two lines exist because a prompt that demands a hazards bullet will invent one (a bear sighting, the word "avalanche" in the trail name) when the report has no real hazard; they give the model a legal way to report nothing. Measured rate with and without those lines is in [`expected-output.md`](expected-output.md).

**Check:** three bullets and nothing else. The gear debrief is gone. Mud patches and the 10am crowds are in. The hazards bullet says "no closures or hazards reported". The only bear in `tr-0001.md` is a ranger's remark about a road near Lake McDonald; keeping it as a plain sighting is fine, closing the trail over it is the failure step 2's grounding lines exist to stop.

### Step 3: Run the same prompt on the buried-hazard report

**Do:**
1. Change the report path from `tr-0001.md` to `tr-0004.md`. Every starter takes the path as its one optional argument: `dotnet run -- ../../data/tr-0004.md`, `uv run main.py ../../data/tr-0004.md`, `npm run starter -- ../data/tr-0004.md`.
2. Leave the step 2 prompt exactly as it is. Run it.
3. Run `tr-0001.md` once more with the same prompt, to confirm the clean report still passes.

**Why:** `tr-0004.md` has the same front matter and rambling shape as `tr-0001.md`; the difference is a washed-out footbridge and a closed trail buried in its fourth paragraph, on purpose, because a summarizer that misses it fails the feature.

**Check:** the first bullet is the washed-out footbridge and the closed trail, a fact buried between airport sandwiches and huckleberry ice cream in the source. A second bullet reading "no other closures reported" is normal. Bullets about the sister, the deer, or Moby the rental SUV mean the prompt in your program is not the one above. `tr-0001.md` still comes back with nothing closed.

### Stretch goals

Pick any. Every track's `complete/` already has each one built in, behind the flag named below.

- **Brief a different reader.** Add an audience variable. When it is `ranger`, replace the first line of the step 2 prompt with the line below; otherwise keep the hiker line. Run `data/tr-0004.md` through both. `complete/` does this with `--briefing --audience ranger`.

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

  **Check:** one line, at most 12 words, that leads with the closure. No bullets, no preamble. Only the instruction changed: a new spot in the UI costs a new prompt, not new infrastructure.

- **See the hallucination the grounding lines prevent.** Delete the two lines of the step 2 prompt that begin "Report only what the trip report states" and end "when it says none." Run `data/tr-0001.md` ten or more times. Put the lines back when done.

  **Check:** some runs now invent a closure from the bear, the creek, or the word "avalanche" in the trail's name. Measured over 24 runs on `tr-0001.md`: 11 of 24 (46%) without the lines, 1 of 24 (4%) with them. `tr-0004.md` led with the bridge in 12 of 12 runs either way. Full numbers in [`expected-output.md`](expected-output.md).

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F01-dotnet.md`](dotnet/F01-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F01-python.md`](python/F01-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F01-typescript.md`](typescript/F01-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `data/tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `data/tr-*.md`: all 40 Avalanche Lake trip reports, synthetic Markdown with a four-field front matter and a prose body; `tr-0001` and `tr-0004` are the lab's two named inputs
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
