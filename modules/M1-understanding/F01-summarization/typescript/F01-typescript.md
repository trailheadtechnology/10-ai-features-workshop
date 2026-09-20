<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 01: Summarization (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F01-dotnet.md), [Python](../python/F01-python.md). Lab overview: [F01-lab.md](../F01-lab.md).*

**The User Problem:** A hiker planning this weekend's trip opens Trailhead Guides and finds forty trip reports for Avalanche Lake Trail, each one 1,200 words of trail diary, gear opinions, and granola recipes. Somewhere in there is the one thing they need to know: is the bridge out, and are the mosquitoes bad? Nobody reads forty essays; they skim three, miss the warning in the fourth, and have a bad Saturday.

*This is the Recommended lab for [Module 1](../../M1-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** `data/tr-0001.md`, a clean report (mud and crowds, no closure), and `data/tr-0004.md`, the report with the washed-out footbridge buried mid-text. The other 38 files in `data/` are the rest of the synthetic Avalanche Lake corpus described in the repo's [README.md](../../../../README.md); no build script produced them.
- **How:** one chat call per report. Read the file, put an instruction above its text, send the whole thing to Ollama as a single user message, and print the reply.
- **Model:** `llama3.2`, local. No key.

## The Concept

Summarization is the simplest possible LLM feature: one chat-completion call with a document and an instruction. You don't need fine-tuning, a vector database, or any pipeline at all. That makes it the right first feature, because by the end of the lab everyone in the room has called a model and built something useful.

The craft is all in the instruction, because "summarize this" produces a book report and real products ask for a summary with a purpose: "In 3 bullets, tell a hiker planning a trip this week about current conditions, hazards, and crowding. Ignore gear talk." The second lesson is that summaries can be shaped to fit the UI slot that needs them: plain prose, bullets, a fixed template, or a single headline. A small local model handles all of this well, which is why this feature never touches the cloud.

Each step below is one thing to make the program do. The starter already reads a report and sends the naive prompt. Edit it until it does all three steps. Look at `complete/` when you get stuck; its comments say why each piece is there.

## Running

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`). Switching to Azure OpenAI later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. Run everything from the `typescript/` folder, where `package.json` lives:

```bash
npm install                                     # once
npm run starter                                 # tr-0001.md, the default
npm run starter -- ../data/tr-0004.md           # any report path works
```

Paths you pass are relative to `typescript/`, even if your shell is in `starter/`, because npm runs scripts from the package folder. That is why the TypeScript paths have one fewer `../` than the other tracks.

`complete/` defaults to `tr-0004.md` and adds flags for the stretch goals:

```bash
npm run complete -- --briefing                  # 3 bullets; the bridge surfaces
npm run complete -- --headline
npm run complete -- --briefing --audience ranger
```

### Step 1: Run the starter and read the book report

**Do:**
1. Review how the existing code from the starter project reads the whole report file into one string. It defaults to `data/tr-0001.md`.

   It happens on lines 19 and 20 of `starter/index.ts`. `process.argv[2]` is the first word typed after `npm run starter --` (`process.argv[0]` and `[1]` are Node and the script itself). When nothing was typed it is `undefined`, and `??` falls back to `tr-0001.md` in the `DATA` folder. `readFileSync(path, "utf8")` returns the whole file as one string, which goes straight into `stripFrontMatter` (item 2):

   ```typescript
   const reportPath = process.argv[2] ?? resolve(DATA, "tr-0001.md");
   const report = stripFrontMatter(readFileSync(reportPath, "utf8"));
   ```

   `readFileSync` and `resolve` come from Node's built-in modules, imported at the top of the file. `DATA` is built on line 12: `import.meta.dirname` is the folder this `.ts` file sits in (`starter/`), and `resolve` joins it with `../../data` into one absolute path:

   ```typescript
   import { readFileSync } from "node:fs";
   import { resolve } from "node:path";
   ```

   ```typescript
   const DATA = resolve(import.meta.dirname, "../../data");
   ```

2. Review how it strips the front matter: split on the `---` lines, keep the third part, trim it. What is left starts at the report's title.

   A helper function on lines 14 to 17, above the code that calls it, does this. `split("---")` cuts the text at every `---`, so part 0 is the empty text before the first `---` and part 1 is the front matter. `slice(2)` keeps every part from the third on, and `join("---")` glues them back together in case the report body itself contains a `---`. `.trim()` removes the blank lines at the ends:

   ```typescript
   function stripFrontMatter(markdown: string): string {
     const parts = markdown.split("---");
     return parts.length >= 3 ? parts.slice(2).join("---").trim() : markdown.trim();
   }
   ```

3. Review how it builds the prompt: the one line below, a blank line, then the report text.

   ```text
   Summarize this trip report.
   ```

   It happens inside the call, on line 24. The backtick string is a template literal: `\n\n` is a line break plus a blank line, and `${report}` drops the report text in:

   ```typescript
     messages: [{ role: "user", content: `Summarize this trip report.\n\n${report}` }],
   ```

4. Review how it sends the prompt as a single user message to `llama3.2`: the model name, plus one message with role `user` and that content. No system message, no temperature, no streaming.

   The client is created once near the top of `index.ts` (line 10); `apiKey: "ollama"` is a placeholder because Ollama needs no key, but the package refuses to start without one:

   ```typescript
   const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
   ```

   The call on lines 22 to 25 passes one object holding the model name and a `messages` array with one message, an object with a `role` and a `content`. `await` waits for the reply; a file with `"type": "module"` in `package.json` may use `await` at the top level, outside any function:

   ```typescript
   const response = await client.chat.completions.create({
     model: "llama3.2",
     messages: [{ role: "user", content: `Summarize this trip report.\n\n${report}` }],
   });
   ```

5. Review how it prints the reply text.

   It happens on the last line. The reply text sits in the first choice's message:

   ```typescript
   console.log(response.choices[0].message.content);
   ```

All five are already in place. Run the starter twice:

```bash
npm run starter
```

**Why:** `Summarize this trip report.` is the prompt most people would start with, but we can do better. Run it more than once and notice what changes and what doesn't: the words come out different every time, but they are fundamentally the same.

**Check:** the model summarized the report faithfully; the report is mostly about gear, so the summary is too. What a hiker planning Saturday needs is missing: trail conditions, hazards, closures, and crowding. The fix has to come from the instruction.

### Step 2: Replace the instruction with the briefing prompt

Step 2 only changes the instruction string. Reading the file, stripping the front matter, and the call all stay as they are.

**Do:**
1. Replace `Summarize this trip report.` with the seven lines below, line breaks exactly as shown.

   ```text
   You are helping a hiker planning to hike this trail within the next week.
   From the trip report below, produce exactly 3 bullets covering:
   current trail conditions, hazards or closures, and crowding.
   Ignore gear talk, personal stories, and scenery.
   Report only what the trip report states. Do not turn a wildlife sighting into a
   hazard or a closure, and write "no closures or hazards reported" when it says none.
   If the report does state a closure or hazard, it must appear in the first bullet.
   ```

   This replaces the starter's whole call on lines 22 to 25, from `const response = await client.chat.completions.create({` down to its closing `});`, with the two statements below. The `console.log(...)` line under it stays. Paste the prompt lines starting at the left edge with no indentation, because spaces at the start of a line inside the backticks become part of the prompt.

   Use a template literal (backticks) so the embedded quotes and line breaks survive as-is. `${report}` drops the report text in after the blank line. The call is the same call written on one line, with `prompt` as the message content:

   ```typescript
   const prompt = `You are helping a hiker planning to hike this trail within the next week.
   From the trip report below, produce exactly 3 bullets covering:
   current trail conditions, hazards or closures, and crowding.
   Ignore gear talk, personal stories, and scenery.
   Report only what the trip report states. Do not turn a wildlife sighting into a
   hazard or a closure, and write "no closures or hazards reported" when it says none.
   If the report does state a closure or hazard, it must appear in the first bullet.

   ${report}`;
   const response = await client.chat.completions.create({ model: "llama3.2", messages: [{ role: "user", content: prompt }] });
   ```

2. Run it on `data/tr-0001.md` four or five times, not once. When I ran it 20 times, every run gave me three bullets, but only about half held to everything in the check below: 4 runs invented a closure outright, 3 more filed the bear as a hazard, and 4 dropped the mud patches. The grounding lines cut that a lot, and they do not eliminate it. If every run invents a closure, those lines did not make it into your prompt.

   From `typescript/`, with no path so the default `tr-0001.md` is used:

   ```bash
   npm run starter
   ```

**Why:** keep the line breaks where they are; the prompt behaves differently once you reflow it. The `Report only what the trip report states.` sentence and the one after it are there because a prompt that demands a hazards bullet will invent a hazard when the report has none, promoting a bear sighting or the word "avalanche" in the trail name into a closure. Those two lines give the model a legal way to say nothing instead. Without them, `tr-0001.md` came back with an invented hazard or closure in 11 of 24 runs; with them, 1 of 24, and the full measurement is in [`expected-output.md`](../expected-output.md). A later 20-run measurement on a newer `llama3.2` pull came in higher, 4 of 20 inventing a closure outright, so treat the lines as a large reduction rather than a fix.

**Check:** three bullets. A one-line lead-in such as "Here are three bullets" is normal for `llama3.2`; anything more than that is not. The gear debrief is gone. Mud patches and the 10am crowds are in. The hazards bullet says "no closures or hazards reported". The only bear in `tr-0001.md` is a ranger's remark about a road near Lake McDonald; keeping it as a plain sighting is fine, closing the trail over it is the failure the grounding lines exist to stop.

### Step 3: Run the same prompt on the buried-hazard report

**Do:**
1. Pass `tr-0004.md` as the report path:

   ```bash
   npm run starter -- ../data/tr-0004.md
   ```

2. Leave the step 2 prompt exactly as it is.
3. Run `tr-0001.md` once more with the same prompt, to confirm the clean report still passes.

   ```bash
   npm run starter
   ```

**Why:** `tr-0004.md` has the same front matter and rambling shape as `tr-0001.md`; the difference is a washed-out footbridge and a closed trail buried in its fourth paragraph. That placement is deliberate, because a summarizer that misses it fails the feature.

**Check:** the first bullet is the washed-out footbridge and the closed trail, a fact buried between airport sandwiches and huckleberry ice cream in the source. A second bullet reading "no other closures reported" is normal. Bullets about the sister, the deer, or Moby the rental SUV mean the prompt in your program is not the one above; diff it against the seven lines, and the usual cause is a reflowed or missing line. `tr-0001.md` still comes back with nothing closed.

### Stretch goals

Pick any. `complete/` already has each one built in, behind the flag named below.

- **Brief a different reader.** Add an audience variable. When it is `ranger`, replace the first line of the step 2 prompt with the line below; otherwise keep the hiker line. Run `data/tr-0004.md` through both. `complete/` does this with `--briefing --audience ranger`.

  ```text
  You are helping a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery.
  ```

  The starter has no argument parsing, so hard-code the audience. Put this line below `const report = ...` and above `const prompt = ...`. It is the exact line from `complete/index.ts`, where `--audience` changes the variable instead (that is why it is `let`, not `const`); change `"hiker"` to `"ranger"` to switch readers:

  ```typescript
  let audience = "hiker";
  ```

  Put these lines right below it. This is one statement: the text after `?` if `audience === "ranger"`, otherwise the text after `:`:

  ```typescript
  const audienceFocus = audience === "ranger"
    ? "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
    : "a hiker planning to hike this trail within the next week";
  ```

  Then replace the first line of your step 2 prompt, the one that starts ``const prompt = `You are helping a hiker``, with the line below, so `${audienceFocus}` is filled in when the string is built. The other lines of the prompt and the call stay as they are:

  ```typescript
  // Hint: this is only the first line; the rest of the prompt follows it unchanged
  const prompt = `You are helping ${audienceFocus}.
  ```

  Run `npm run starter -- ../data/tr-0004.md` once with `"hiker"` and once with `"ranger"`.

  **Check:** the ranger version leads with the washed-out bridge as a maintenance or closure item and drops the crowding; the hiker version keeps the crowding. Identical output means the audience line is not reaching the prompt.

- **Shrink the summary to a headline.** Same file, same call, different instruction. Replace the whole step 2 prompt with the three lines below and run `data/tr-0004.md`. `complete/` does this with `--headline`.

  ```text
  From the trip report below, write ONE line of at most 12 words,
  suitable for a status badge on a trail card in an app.
  Lead with the most important condition or closure. No preamble.
  ```

  This replaces your whole step 2 ``const prompt = `...`;`` statement, all nine lines from ``const prompt = `You are helping`` down to ``${report}`;``. The `const response = ...` and `console.log(...)` lines below it stay as they are. If you did the audience goal, the `audience` lines can stay; this prompt just does not use them. Paste the lines starting at the left edge. Run `npm run starter -- ../data/tr-0004.md`:

  ```typescript
  const prompt = `From the trip report below, write ONE line of at most 12 words,
  suitable for a status badge on a trail card in an app.
  Lead with the most important condition or closure. No preamble.

  ${report}`;
  ```

  **Check:** one line, at most 12 words, that leads with the closure. No bullets, no preamble. Only the instruction changed. A new spot in the UI costs a new prompt rather than new infrastructure.

- **See the hallucination the grounding lines prevent.** If you did the headline stretch goal, put the step 2 prompt back first. Then delete the two lines of the step 2 prompt that begin "Report only what the trip report states" and end "when it says none." Run `data/tr-0001.md` (the starter's default) ten or more times. Put the lines back when done.

  These are the two lines to delete from inside your ``const prompt = `...` `` string. Delete the whole lines, so no blank line is left behind. Then run `npm run starter` repeatedly:

  ```typescript
  Report only what the trip report states. Do not turn a wildlife sighting into a
  hazard or a closure, and write "no closures or hazards reported" when it says none.
  ```

  **Check:** some runs now invent a closure from the bear, the creek, or the word "avalanche" in the trail's name. Measured over 24 runs on `tr-0001.md`: 11 of 24 (46%) without the lines, 1 of 24 (4%) with them. `tr-0004.md` led with the bridge in 12 of 12 runs either way. Full numbers in [`expected-output.md`](../expected-output.md).

## What Is in This Folder

- `data/tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `data/tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `data/tr-*.md`: all 40 Avalanche Lake trip reports, synthetic Markdown with a four-field front matter and a prose body; `tr-0001` and `tr-0004` are the lab's two named inputs
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
