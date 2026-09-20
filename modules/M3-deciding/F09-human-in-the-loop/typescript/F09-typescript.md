<!-- Generated file: built from the lab source by build_labs.py. Edit the source and rerun; hand edits to this file are overwritten. -->
# Lab 09: Human-in-the-Loop (TypeScript)

*You are on the TypeScript track. Other tracks: [.NET](../dotnet/F09-dotnet.md), [Python](../python/F09-python.md). Lab overview: [F09-lab.md](../F09-lab.md).*

**The User Problem:** Feature 07 routed the inbox, so now a ranger stares at forty messages that all need replies. Most answers are boilerplate the ranger has typed a hundred times, and typing them eats the afternoon. The obvious move is to let the AI answer, and the obvious disaster is the AI telling a visitor that campfires are fine during a burn ban, on official park letterhead. The ranger's problem is drudgery; the park's problem is that full automation of outbound communication is how you end up apologizing publicly.

*A Challenge lab. Do it if you finished [Module 3](../../M3-overview.md)'s Recommended lab and want another, or skip it without guilt: you will have seen this feature demonstrated either way.*

- **Goal:** put a human between the model and the visitor. The model drafts, a ranger approves, edits, or rejects, every decision is logged, and emergencies never reach the model at all.
- **Input:** `data/inquiries.jsonl`, six inquiries routed by feature 07, with category and park doc; `data/snippets/`, the four excerpts they cite (`glac-bc-2025-04.md`, `glac-cl-2026-01.md`, `yose-cl-2026-01.md`, `zion-nar-2026-01.md`); `policy-worksheet.md`, the lane table you fill in.
- **How:** the starter already loops over the six inquiries and sends each one to the chat model with a drafting prompt. You add five things: a policy table, a gate above the model call, a review prompt, an outbox, and an audit log.
- **Model:** `llama3.2`, local. No key.

## The Concept

Human-in-the-loop is a product pattern, not a model feature, and it's the difference between AI features that ship and AI features that get killed in legal review. The core move: the AI drafts, the human approves, edits, or rejects, and the system remembers what happened. The user-facing risk drops to near zero while most of the typing still disappears.

The design question is where to put the human, and the answer comes from error cost and reversibility, which connects straight back to feature 07's asymmetry lesson. A sensible policy has three lanes: full automation for cheap, reversible, low-stakes replies; draft-plus-approval for the middle; human-only for the expensive and irreversible (in Trailhead Guides terms, emergencies never get an AI draft at all). Two practical details do a lot of work in real systems. Keep an audit trail of what was drafted, who approved it, and what they changed. And measure the gap between draft and final text, because how much humans edit tells you whether trust in each lane is earned, and edit patterns show you exactly where the drafts fall short.

Every step below is one thing to make the program do. The `starter/` drafts and "sends" every inquiry with no review. Edit it until it does all seven steps. Compare against `complete/` when stuck; its comments say why each piece is there.

## Running

`outbox/` and `decisions.jsonl` are run artifacts and are gitignored.

Two scripts, both using the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), so swapping the provider later is a different constructor and nothing else. `tsx` runs the `.ts` files directly, so there is no build step. `starter/index.ts` is deliberately unsafe: every draft goes straight out, nothing is logged, and nothing treats an emergency differently except a sentence in the system prompt that the model is free to ignore. `complete/index.ts` is the finished demo as shown on stage.

Run everything from the `typescript/` folder, where `package.json` lives. Run `npm install` there once, then:

```bash
npm run starter
```

npm runs scripts from the package folder, so `typescript/` is the working directory, not `starter/`. That matters in step 4.

`complete/` adds flags for the stretch goals. npm flags go after `--`:

```bash
npm run complete                                  # review the queue interactively
npm run complete -- --policy                      # print the routing policy table and exit
npm run complete -- --auto-approve-dry-run        # non-interactive run for testing
```

### Step 0: Run the starter and read every draft as an editor

**Do:** run `starter/` as it is. Read the code first. It is the loop you are about to change:

1. It opens `../../data/inquiries.jsonl`: `id`, `channel`, `received`, `category`, `doc`, `text` per line.

   The starter already does this. `import.meta.dirname` is the folder holding `index.ts`, so `DATA` is the `data/` folder two levels above it. `process.argv[2]` is the first argument you pass after `npm run starter --`. `readFileSync(path, "utf8")` reads the whole file as a string, `.split("\n")` cuts it into lines, and `JSON.parse` turns one line of JSON into an object, so later code reads fields as `inquiry.id` or `inquiry.text`. The `Inquiry` type names those fields:

   ```typescript
   const DATA = resolve(import.meta.dirname, "../../data");
   const inquiriesPath = process.argv[2] ? resolve(process.argv[2]) : resolve(DATA, "inquiries.jsonl");
   const dataDir = dirname(inquiriesPath);
   ```

   ```typescript
   type Inquiry = { id: string; channel: string; received: string; category: string; doc: string; text: string };

   for (const line of readFileSync(inquiriesPath, "utf8").split("\n")) {
     if (!line.trim()) continue;
     const inquiry: Inquiry = JSON.parse(line);
   ```

2. For each inquiry, if `doc` is not empty and `data/snippets/<doc>` exists, it reads and trims that file as the excerpt. Otherwise the excerpt is the literal text `(none on file for this message)`.

   The starter already does this, inside the loop. `inquiry.doc || ""` uses an empty name when `doc` is missing, `existsSync` checks that the file is there, and `condition ? x : y` picks one of two values:

   ```typescript
     const snippetPath = resolve(dataDir, "snippets", inquiry.doc || "");
     const snippet = inquiry.doc && existsSync(snippetPath) ? readFileSync(snippetPath, "utf8").trim() : "(none on file for this message)";
   ```

3. It builds a system message (the drafting prompt, unchanged across inquiries) and a user message (excerpt + channel + received + text), sends both to `llama3.2`, and trims the reply.

   The starter already does this. `client` is created once above the loop. `client.chat.completions.create` takes the model name and the list of messages and returns a response object; `await` waits for it (a `.ts` file run as a module may use `await` outside any function). The reply text is `draft.choices[0].message.content`, which step 0 item 4 prints. The user message is a backtick template string, so `${...}` inserts a value and the line breaks are part of the text:

   ```typescript
   const client = new OpenAI({ baseURL: "http://localhost:11434/v1", apiKey: "ollama" });
   ```

   ```typescript
     const draft = await client.chat.completions.create({
       model: "llama3.2",
       messages: [
         { role: "system", content: SYSTEM_PROMPT },
         { role: "user", content: `Reference excerpt:
   ${snippet}

   Visitor message (${inquiry.channel}, received ${inquiry.received}):
   ${inquiry.text}

   Draft the reply.` },
       ],
     });
   ```

4. It prints `=== SENT to visitor · <id> (<category>) ===`, the draft, then a blank line. After the loop: `All replies sent. Nobody read them. Nothing was logged.`

   The starter already does this. The first three lines are the end of the loop body; `console.log()` with nothing inside prints the blank line, and `?? ""` uses an empty string if the model returned no content. The last line is after the loop's closing `}`:

   ```typescript
     console.log(`=== SENT to visitor · ${inquiry.id} (${inquiry.category}) ===`);
     console.log((draft.choices[0].message.content ?? "").trim());
     console.log();
   ```

   ```typescript
   console.log("All replies sent. Nobody read them. Nothing was logged.");
   ```

The system prompt:

```text
You are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers. Exception: if the visitor's message reports an emergency, an injury, a possible fire, or a missing or overdue person, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch.
```

```bash
npm run starter
```

**Why:** `data/inquiries.jsonl` is six of the 100 messages in feature 07's inbox, with two fields feature 07's output would supply: `category` (feature 07's `conditions` is spelled `trail-condition` here) and `doc` (the park-doc excerpt file name in `data/snippets/`, or empty for `inq-0007` and `inq-0013`). Each snippet file quotes only the sections that answer its inquiry, word for word from the full documents in feature 05's corpus, so the model sees the source section and nothing else.

**Check:** read every draft the way a ranger would.
- `inq-0002` says the Mist Trail is closed when the excerpt says it reopened (reject).
- `inq-0051` and `inq-0005` are accurate and cited (approve).
- `inq-0003` gets both rules right but pins the flash flood rule on `GLAC-BC-2026-01`, a Glacier document number, on a Zion question (edit).
- `inq-0007` apologizes and decides nothing (edit).
- `inq-0013`, the overdue-hiker report, gets a warm reply to Diane with no `ESCALATE` line, even though the prompt told the model not to draft one. In the recorded runs that happened 3 times out of 3.

The rest of the lab exists because of that last draft.

### Step 1: Fill in policy-worksheet.md

**Do:** open `policy-worksheet.md`. It has one row per feature 07 category (`permit`, `conditions`, `complaint`, `lost-and-found`, `general`, `emergency`; `unsure` has no row), five columns: Category, Lane, Worst plausible error, Reversible?, Justification.

1. For each row, pick a lane: `auto-send`, `draft-for-approval`, or `human-only`.
2. Write the worst plausible wrong reply for that category.
3. Write whether that error can be taken back after it's sent.
4. Write a one-sentence justification.
5. Answer the two follow-up questions under the table (the second asks where your lane lives in the code; steps 2 and 3 are the answer).

**Check:** every row says what a wrong reply costs, and emergency is `human-only`. If your reason for a lane is "the prompt tells it to," reread the `inq-0013` draft from step 0. Your lanes may differ from the reference in `expected-output.md`. Your justifications are what count.

### Step 2: Add the policy table and print the lane

**Do:**
1. Above the loop, build a dictionary from category to lane:

```text
trail-condition   draft-for-approval
permit            draft-for-approval
complaint         draft-for-approval
general           draft-for-approval
lost-and-found    draft-for-approval
emergency         human-only
```

A `Record<string, string>` is an object that maps each category (the key) to its lane (the value). Put it right below the closing `` `; `` of `SYSTEM_PROMPT`, above the `const DATA` line. This is exactly what `complete/` has:

```typescript
const POLICY: Record<string, string> = {
  "trail-condition": "draft-for-approval",
  "permit": "draft-for-approval",
  "complaint": "draft-for-approval",
  "general": "draft-for-approval",
  "lost-and-found": "draft-for-approval",
  "emergency": "human-only",
};
```

2. Print the table once at the top of the run, under `Routing policy (error cost decides the lane):`, one row per entry, two-space indent, category left-aligned in a 16-character column, then the lane, then a blank line after the table.

   Write the printout as a function so the `--policy` stretch goal can reuse it. Put it directly below the `POLICY` object, with a blank line between them. `Object.entries(POLICY)` walks the object and gives each key and value as `[category, lane]`, and `category.padEnd(16)` pads the value with spaces to 16 characters, left-aligned:

   ```typescript
   function printPolicy(): void {
     console.log("Routing policy (error cost decides the lane):");
     for (const [category, lane] of Object.entries(POLICY)) console.log(`  ${category.padEnd(16)} ${lane}`);
     console.log();
   }
   ```

   Then call it once, with no indent, on its own line directly above `for (const line of readFileSync(inquiriesPath, "utf8").split("\n")) {`:

   ```typescript
   printPolicy();
   ```

3. Inside the loop, right after parsing the inquiry, look up its `category`. If not found, use `human-only`.

   `POLICY[inquiry.category]` is `undefined` when the category is not in the object, and `?? "human-only"` replaces `undefined` with the second value, which is the fail-closed lookup. Add it on the line directly below `const inquiry: Inquiry = JSON.parse(line);`, at the same indent:

   ```typescript
     const lane = POLICY[inquiry.category] ?? "human-only";
   ```

4. Print a per-inquiry header right after the lane lookup, before the excerpt is read (the old `SENT` line after the model call goes away in step 4): 72 dashes, `<id>  ·  <category>  ·  <channel>  ·  lane: <lane>`, another 72 dashes, the visitor's `text` prefixed with `  | ` per line, blank line.

   `"-".repeat(72)` builds a string of 72 dashes. Add these lines directly below the `const lane` line, above `const snippetPath`:

   ```typescript
     console.log("-".repeat(72));
     console.log(`${inquiry.id}  ·  ${inquiry.category}  ·  ${inquiry.channel}  ·  lane: ${lane}`);
     console.log("-".repeat(72));
     console.log(indent(inquiry.text));
     console.log();
   ```

   `indent` is a small helper written as an arrow function (`(text: string) => ...` takes `text` and returns the expression after `=>`). It splits the text on newlines, `.map` puts `  | ` in front of each line, and `.join("\n")` glues the lines back together. Put it above the loop, directly below the `type Inquiry = ...` line:

   ```typescript
   const indent = (text: string) => text.split("\n").map((l) => "  | " + l.trimEnd()).join("\n");
   ```

```bash
npm run starter
```

**Why:** feature 07 can return `unsure`, and the worksheet has no row for it. Falling back to `human-only` means a category you did not plan for goes to a ranger rather than out as a draft.

**Check:** the run opens with the six-row table. Every inquiry header ends in its lane. `inq-0013` shows `lane: human-only` and the other five show `lane: draft-for-approval`. The drafts still print for all six, including the emergency. The next step fixes that.

### Step 3: Gate the emergency above the model call

**Do:** this is the whole feature. The check must run before any message is built; the model call must stay below it.

1. Right after the lane lookup and header, and before the excerpt is read or a message is built, test `lane == "human-only"`.
2. If true, print `  NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and move to the next inquiry without reading the excerpt, building a request, or spending tokens.
3. Only below that test, read the excerpt and call the model as before.

Add this block directly below the header's final `console.log();`, above the `const snippetPath` line, at the same 2-space indent. `===` is TypeScript's equality test. `continue` jumps straight to the next inquiry, so nothing below it in the loop runs: no excerpt, no message, no `client.chat.completions.create` call. The `\n` at the end of the text prints the blank line after it.

```typescript
  if (lane === "human-only") {
    console.log("  NO DRAFT. Policy routes this straight to a human. Paging dispatch.\n");
    continue;
  }
```

Step 5 adds the log line inside this block, above `continue;`.

```bash
npm run starter
```

**Why:** the system prompt already told the model not to draft a reply for a message like `inq-0013`, and in the recorded runs it drafted one anyway, 3 times out of 3. A prompt instruction is a request; this lane is a guarantee. The test sits above the model call, so the emergency never depends on the model cooperating. Leave the `ESCALATE` sentence in the system prompt where it is.

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` and the program makes no model call for it. The other five still get drafts.

### Step 4: Ask the reviewer and queue approved text in an outbox

**Do:**
1. Replace the `SENT` printout: after the model call, print `  draft:`, blank line, the draft prefixed with `  | ` per line, blank line.

   First, in the model call from step 0, rename the result from `draft` to `response`, so the name `draft` is free for the trimmed string. Change the first line of the call to:

   ```typescript
     const response = await client.chat.completions.create({
   ```

   Then delete the starter's three `SENT` lines after the call:

   ```typescript
     console.log(`=== SENT to visitor · ${inquiry.id} (${inquiry.category}) ===`);
     console.log((draft.choices[0].message.content ?? "").trim());
     console.log();
   ```

   In their place, directly below the `});` that ends the model call, show the draft. `indent` is the helper from step 2:

   ```typescript
     const draft = (response.choices[0].message.content ?? "").trim();
     console.log("\r  draft:      \n");
     console.log(indent(draft));
     console.log();
   ```

   `complete/` also prints `  drafting...` on the line above `const response`, and the `\r` in the draft line moves the cursor back to overwrite it. `process.stdout.write` prints without a newline. Optional:

   ```typescript
     process.stdout.write("  drafting...");
   ```

2. Print `  [a]pprove  [e]dit  [r]eject  [s]kip > ` and read one line; trim and lowercase it.

   Node has no one-line "wait for Enter" call, so you build one reader for the whole run with `node:readline/promises` and ask it a question at each prompt. Add the import at the top of the file, below the other imports:

   ```typescript
   import { createInterface } from "node:readline/promises";
   ```

   Create the reader once, above the loop, directly below the `const client = ...` line. `complete/` skips the reader when the `--auto-approve-dry-run` stretch flag is on, so it needs an `autoApprove` switch; declare it directly below the `const dataDir` line and leave it `false` until you add that flag. `let` (not `const`) means the stretch goal can change it later:

   ```typescript
   let autoApprove = false;
   ```

   ```typescript
   const rl = autoApprove ? null : createInterface({ input: process.stdin, output: process.stdout });
   ```

   Inside the loop, below the draft display, ask the question. `rl.question(prompt)` prints the prompt without a newline and gives back the typed line once Enter is pressed; `await` waits for it. `rl` may be `null` in the stretch goal, and the `!` in `rl!` tells TypeScript it is not `null` here:

   ```typescript
     const key = (await rl!.question("  [a]pprove  [e]dit  [r]eject  [s]kip > ")).trim().toLowerCase();
     console.log();
   ```

   The open reader keeps the program running after the last inquiry, so close it after the loop: add this line directly below the loop's closing `}`, above the closing `console.log` line. `?.` skips the call when `rl` is `null`:

   ```typescript
   rl?.close();
   ```

3. Map the key to a decision and final text: `a` → `approved`, final = draft. `e` → `edited`, final = what the reviewer types. `r` → `rejected`, no final text. Anything else → `skipped`, no final text.

   Declare the two results first. `let` means the value can change; `string | null` means `final` may be `null`, which is how "no final text" is stored. A `switch` picks the `case` that matches `key`, `break` ends that case, and `default` catches everything else. Add below the `rl!.question` lines:

   ```typescript
     let decision: string;
     let final: string | null = null;
     switch (key) {
       case "a": decision = "approved"; final = draft; break;
       case "e": decision = "edited"; final = await readEdited(draft); break;
       case "r": decision = "rejected"; break;
       default: decision = "skipped"; break;
     }
   ```

   `readEdited` does not exist yet. Item 4 writes it.

4. For `e`, read lines until a line that is just `.`. If the first line is empty, copy the draft in first and keep reading (so the reviewer can append instead of retyping). Join, trim; if empty, fall back to the draft.

   `readEdited` is a helper function, so it goes above the loop, directly below the `const indent = ...` line from step 2. It is `async` because it waits for typed lines, which is why the `switch` calls it with `await`. It asks `rl!.question("")` (no prompt text) once per line and collects the lines in an array inside a `for (;;)` loop, which runs until `break`. `lines.join("\n")` glues the lines back together:

   ```typescript
   async function readEdited(draft: string): Promise<string> {
     console.log("  Type the reply you want to send. End with a single '.' on its own line.");
     console.log("  Press Enter on the first line to start from the draft text instead.\n");
     const lines: string[] = [];
     let first = true;
     for (;;) {
       const line = await rl!.question("");
       if (line === ".") break;
       if (first && line.length === 0) {
         lines.push(draft);
         console.log("  (draft copied in; keep typing to append, '.' to finish)");
       } else {
         lines.push(line);
       }
       first = false;
     }
     const edited = lines.join("\n").trim();
     return edited.length === 0 ? draft : edited;
   }
   ```

5. At startup create `outbox/` in `starter/`, the folder holding the source file you are editing, and put `decisions.jsonl` beside it. Anchor those two paths so they land there under your track's run command, never next to a compiled binary. When there's a final text, write it plus a newline to `outbox/<id>.txt` and print `  -> <decision>, queued at <path>`, where `<path>` is that file relative to the folder you ran from (`outbox/<id>.txt` when you run from `starter/`). Otherwise print `  -> <decision>, nothing queued`.

   `npm run starter` runs with `typescript/` as the working directory, so a bare `"outbox"` would land next to `package.json`. Anchor both paths to the script's own folder instead: `import.meta.dirname` is the folder holding `index.ts`, and `resolve(folder, name)` joins them into a full path. First replace the starter's two `node:fs` and `node:path` import lines at the top of the file with these (they add the file-writing functions and `relative`; this step and step 5 use them all):

   ```typescript
   import { appendFileSync, existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
   import { dirname, relative, resolve } from "node:path";
   ```

   Declare the paths directly below the `const inquiriesPath = ...` line, above `const dataDir` (`decisionsPath` is used in step 5):

   ```typescript
   const HERE = import.meta.dirname;
   let outboxDir = resolve(HERE, "outbox");
   let decisionsPath = resolve(HERE, "decisions.jsonl");
   ```

   Create the folder once, directly below the `let autoApprove = false;` line. `{ recursive: true }` means "also create missing parent folders, and do not fail if it already exists":

   ```typescript
   mkdirSync(outboxDir, { recursive: true });
   ```

   `rel` is one more helper for above the loop (directly below `const indent` is fine). `relative(process.cwd(), p)` turns a full path into one relative to the folder you ran from:

   ```typescript
   const rel = (p: string) => relative(process.cwd(), p) || ".";
   ```

   Then, inside the loop below the `switch`, write the file only when there is final text. `` `${inquiry.id}.txt` `` builds the file name, `writeFileSync(path, final + "\n")` writes the text with a trailing newline (replacing any older file), and `rel(path)` prints `starter/outbox/inq-0051.txt` under `npm run starter`:

   ```typescript
     if (final !== null) {
       const path = resolve(outboxDir, `${inquiry.id}.txt`);
       writeFileSync(path, final + "\n");
       console.log(`  -> ${decision}, queued at ${rel(path)}\n`);
     } else {
       console.log(`  -> ${decision}, nothing queued\n`);
     }
   ```

   That `if`/`else` is the last thing in the loop body for now. Step 5 adds the log line below it.

```bash
npm run starter
```

**Check:** approve `inq-0051` and `outbox/inq-0051.txt` appears holding the draft. Reject `inq-0002` and no file appears for it. Edit `inq-0003` by pressing Enter, typing one line, then `.`, and `outbox/inq-0003.txt` holds the draft with your line appended.

The file is `starter/outbox/inq-0051.txt`, and the process exits on its own after the sixth inquiry.

### Step 5: Log every decision to decisions.jsonl with an edit distance

**Do:**
1. Write a Levenshtein edit-distance function (single-character inserts/deletes/substitutions to turn one string into another; two-row DP is fine).

   Another helper function for above the loop, directly below `readEdited`. `previous` and `current` are the two rows: `previous[j]` holds the distance from the first `i - 1` characters of `a` to the first `j` characters of `b`. `Array.from({ length: b.length + 1 }, (_, j) => j)` builds the array `[0, 1, 2, ...]`, and `a[i - 1]` is one character of `a`. Each cell takes the cheapest of insert, delete, or substitute (`Math.min`), and `previous = current` makes the row just filled the one the next pass reads:

   ```typescript
   function editDistance(a: string, b: string): number {
     let previous = Array.from({ length: b.length + 1 }, (_, j) => j);
     for (let i = 1; i <= a.length; i++) {
       const current = [i];
       for (let j = 1; j <= b.length; j++) {
         const cost = a[i - 1] === b[j - 1] ? 0 : 1;
         current[j] = Math.min(current[j - 1] + 1, previous[j] + 1, previous[j - 1] + cost);
       }
       previous = current;
     }
     return previous[b.length];
   }
   ```

2. After every review decision, append one JSON object per line to `decisions.jsonl`: `at` (UTC ISO 8601), `inquiryId`, `category`, `lane`, `decision`, `reviewer` (OS username), `draft`, `final` (null when nothing queued), `editDistance` (distance from draft to final, or to empty string when there's no final text). Match the key names exactly.

   Add one import at the top of the file, with the other imports. `userInfo().username` is the OS username:

   ```typescript
   import { userInfo } from "node:os";
   ```

   Declare a `Decision` type with the lab's exact key names, directly below the `type Inquiry = ...` line:

   ```typescript
   type Decision = { at: string; inquiryId: string; category: string; lane: string; decision: string; reviewer: string; draft: string | null; final: string | null; editDistance: number };
   ```

   Set the reviewer above the loop, directly below `const client = ...`. `complete/` picks the name from the `--auto-approve-dry-run` stretch flag, which is why it reads `autoApprove` from step 4:

   ```typescript
   const reviewer = autoApprove ? "auto-approve-dry-run" : userInfo().username;
   ```

   Two helpers, above the loop with the others (directly below `const rel = ...` is fine). `record` builds a `Decision` object; `new Date().toISOString()` is the UTC timestamp, and writing `lane,` alone is short for `lane: lane,`. `log` turns the object into one line of JSON with `JSON.stringify` (which writes `null` as JSON `null`) and `appendFileSync` adds it to the end of the file, creating the file on first use:

   ```typescript
   const log = (d: Decision) => appendFileSync(decisionsPath, JSON.stringify(d) + "\n");
   const record = (inquiry: Inquiry, lane: string, decision: string, draft: string | null, final: string | null, editDistance: number): Decision =>
     ({ at: new Date().toISOString(), inquiryId: inquiry.id, category: inquiry.category, lane, decision, reviewer, draft, final, editDistance });
   ```

   Inside the loop, directly below the outbox `if`/`else` from step 4, log the decision. `final ?? ""` measures against an empty string when `final` is `null`:

   ```typescript
     log(record(inquiry, lane, decision, draft, final, editDistance(draft, final ?? "")));
   ```

3. In the step 3 gate, before moving on, append the same shape with `decision: "escalated"`, `draft`/`final` both null, `editDistance` 0.

   Inside the step 3 `if (lane === "human-only") {` block, between the `NO DRAFT` line and `continue;`, at the same 4-space indent:

   ```typescript
       log(record(inquiry, lane, "escalated", null, null, 0));
   ```

4. Count decisions by name. After the loop print 72 equals signs, then `Queue done: ` plus the counts (e.g. `1 escalated, 3 approved, 1 edited, 1 rejected`), then `Audit trail: decisions.jsonl   ·   Outbox: outbox/` (both paths relative to the folder you ran from, as in step 4).

   A `Record<string, number>` object counts each decision name. Declare it above the loop, directly below the `const reviewer` line:

   ```typescript
   const counts: Record<string, number> = {};
   ```

   `bump` is one more helper, directly below `const record`. `counts[key] ?? 0` is 0 for a name not seen yet:

   ```typescript
   const bump = (key: string) => { counts[key] = (counts[key] ?? 0) + 1; };
   ```

   Call it right below each `log(...)` call. In the gate, above `continue;`:

   ```typescript
       bump("escalated");
   ```

   At the bottom of the loop, below the other `log(...)`:

   ```typescript
     bump(decision);
   ```

   Replace the starter's closing `All replies sent` line (below `rl?.close();`) with the summary. `Object.entries(counts)` gives each `[name, count]` pair, `.map` turns each pair into text like `3 approved`, and `.join(", ")` puts `, ` between them:

   ```typescript
   console.log("=".repeat(72));
   console.log("Queue done: " + Object.entries(counts).map(([k, v]) => `${v} ${k}`).join(", "));
   console.log(`Audit trail: ${rel(decisionsPath)}   ·   Outbox: ${rel(outboxDir)}/`);
   ```

```bash
npm run starter
```

**Check:** one full run adds six lines to `decisions.jsonl`. One is `inq-0013` with `"decision":"escalated"` and `"draft":null`. Approved lines have `"editDistance":0`. A rejected line's `editDistance` equals its draft's length. (`decisions.jsonl` and `outbox/` are run artifacts. Delete them between runs for a clean take.)

### Step 6: Run the queue and compare with expected-output.md

**Do:**
1. Run your program and review all six inquiries. Approve at least one, edit at least one, reject at least one.
2. Read the annotated drafts in `expected-output.md` and compare your choices against what a ranger should do and why.
3. Read the Reference Policy table in `expected-output.md` against your worksheet from step 1.

```bash
npm run starter
```

**Check:** `inq-0013` prints `NO DRAFT. Policy routes this straight to a human. Paging dispatch.` with no model call; every other draft offers `[a]pprove [e]dit [r]eject [s]kip`, logs to `decisions.jsonl`, and queues approved text in `outbox/`. Your lanes may differ from `expected-output.md`. Your justifications are what count.

### Stretch goals

Pick any. Each one is already built in `complete/`, with the measured reason for it in [`expected-output.md`](../expected-output.md).

- **Try the prompt repair and watch it fail.** Move the escalation rule to the front of the system prompt and add "write nothing after that line":

  ```text
  FIRST, before anything else, check the visitor's message for an emergency: an injury, a possible fire, or a missing or overdue person. If you see one, do not draft a reply at all. Output exactly one line beginning with ESCALATE: followed by a one-line reason, so the message goes straight to dispatch. Write nothing after that line.
  Otherwise, you are drafting a reply to a park visitor on behalf of a ranger at Trailhead Guides. A human ranger reviews your draft before anything is sent, so write it ready to send: friendly, plain, professional, at most two short paragraphs, signed 'Trailhead Guides Ranger Desk'. When your answer involves a park rule or a closure, state the rule and cite the source document number and section (for example GLAC-BC-2025-04, Section 4.2). Use only facts from the reference excerpt provided; if the excerpt does not answer the question, say a ranger will follow up with specifics rather than guessing. Never invent dates, fees, policies, or phone numbers.
  ```

  Put this in your program's system message. Your step 3 gate stops `inq-0013` before the model call, so on the real queue it prints `NO DRAFT` and the model never sees it. To watch what the prompt does with an emergency, make a copy of `data/inquiries.jsonl` named `data/inquiries-miscategorized.jsonl`, change `inq-0013`'s `category` from `emergency` to `general` in the copy, and run the whole queue from the copy once:

  ```bash
  npm run starter -- ../data/inquiries-miscategorized.jsonl
  ```

  **Check:** on the copy, `inq-0013` usually prints `ESCALATE:` and then the reply anyway (3 runs out of 3 in the record), and the reply invents an active search; some runs skip the `ESCALATE:` line and just invent the search. Under the same prompt the model also escalates a routine message or two, such as `inq-0002`, `inq-0003`, or `inq-0005`; which ones changes from run to run. Tightening the prompt trades one failure for the other. Put the original prompt back.

  The system message is the `SYSTEM_PROMPT` string at the top of index.ts. A backtick string keeps every line between the opening and closing backticks, so replace only the text inside them, and keep the two lines of the new prompt on two lines. The prompt text has no backticks in it, so it can be pasted as is. Keep the original somewhere (a copy of the file is fine) so you can put it back:

  ```typescript
  // Hint:
  const SYSTEM_PROMPT = `FIRST, before anything else, ... Write nothing after that line.
  Otherwise, you are drafting a reply ... Never invent dates, fees, policies, or phone numbers.`;
  ```

  `npm run starter -- <file>` passes the file to the program, and the path is relative to `typescript/`, where npm runs. The starter still reads that first argument as the queue file.

- **Add the ESCALATE backstop.** After the model call and before the review prompt, test whether the draft starts with `ESCALATE` (case-insensitive). If it does, print `  Model asked to escalate. Draft discarded, routing to a human.`, log a line with `decision: "escalated"`, draft kept, `final` null, `editDistance` 0, then move on. **Why:** this catches an emergency that arrived under the wrong category. It runs after the model has already answered, so it's a backup, never the main control. **Check:** run against `data/inquiries-miscategorized.jsonl`, the copy from the previous stretch goal with `inq-0013` labeled `general` (make it now if you skipped that goal). Most runs offer Diane's warm reply for approval, because the model doesn't escalate. On a run where it does, the backstop line prints and the reviewer never sees the draft. That gap is why step 3 exists.

  The test goes directly below the draft display from step 4 (below the `console.log(indent(draft));` and `console.log();` lines) and above `const key = ...`. It looks like the step 3 gate, with the draft passed to `record` instead of `null`. `draft.toUpperCase()` makes the test case-insensitive:

  ```typescript
    if (draft.toUpperCase().startsWith("ESCALATE")) {
      console.log("  Model asked to escalate. Draft discarded, routing to a human.\n");
      log(record(inquiry, lane, "escalated", draft, null, 0));
      bump("escalated");
      continue;
    }
  ```

  To run against the copy, from `typescript/` (the starter still reads its first argument as the queue file):

  ```bash
  npm run starter -- ../data/inquiries-miscategorized.jsonl
  ```

- **Add the flags `complete/` supports.** `--policy` prints the routing table and exits without reading the queue. `--auto-approve-dry-run` sets the reviewer to `auto-approve-dry-run`, approves every draft without asking, and prints `--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.` under the table. `--outbox <dir>` and `--decisions <file>` move the run artifacts. Any other argument is the path to a queue file. **Check:** `--policy` prints six rows and nothing else. `--auto-approve-dry-run` writes six lines to `decisions.jsonl`, five `approved` and one `escalated` (or four and two, if you added the backstop and `llama3.2` wrote a spurious `ESCALATE` on a routine message), and one file in `outbox/` for each approved draft.

  `complete/index.ts` shows the argument loop over `process.argv.slice(2)`. First change the starter's `inquiriesPath` line so it no longer reads `process.argv[2]`, and make it `let` so the loop can change it:

  ```typescript
  let inquiriesPath = resolve(DATA, "inquiries.jsonl");
  ```

  The loop has to come after the four names it sets exist and before `const dataDir` and `mkdirSync(...)`, which use them. Check that `inquiriesPath`, `HERE`, `outboxDir`, `decisionsPath`, and `autoApprove` sit together above `const dataDir` (move `let autoApprove = false;` up there), then put the loop directly below them. `process.argv.slice(2)` is the list of arguments after `npm run starter --`, `args[++i]` moves on to the value after the flag and reads it, and `process.exit(0)` ends the program:

  ```typescript
  const args = process.argv.slice(2);
  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case "--auto-approve-dry-run": autoApprove = true; break;
      case "--outbox": outboxDir = resolve(args[++i]); break;
      case "--decisions": decisionsPath = resolve(args[++i]); break;
      case "--policy": printPolicy(); process.exit(0);
      default: inquiriesPath = resolve(args[i]); break;
    }
  }
  ```

  For `--auto-approve-dry-run`, print the warning directly below the `printPolicy();` call above the queue loop:

  ```typescript
  if (autoApprove) console.log("--auto-approve-dry-run: approving every draft unread. Testing only, never a shipping mode.\n");
  ```

  Then wrap step 4's prompt and `switch` in an `if`/`else`, so the reviewer is only asked when the flag is off. Everything that was there moves one level (2 spaces) to the right under `else {`, and the `let decision` and `let final` lines move up above the new `if`:

  ```typescript
    let decision: string;
    let final: string | null = null;
    if (autoApprove) {
      decision = "approved";
      final = draft;
      console.log("  [auto] approved\n");
    } else {
      const key = (await rl!.question("  [a]pprove  [e]dit  [r]eject  [s]kip > ")).trim().toLowerCase();
      console.log();
      switch (key) {
        case "a": decision = "approved"; final = draft; break;
        case "e": decision = "edited"; final = await readEdited(draft); break;
        case "r": decision = "rejected"; break;
        default: decision = "skipped"; break;
      }
    }
  ```

  Run it with `npm run starter -- --policy` or `npm run starter -- --auto-approve-dry-run`.

- **Use edit distance as the promotion signal.** Approve, edit, and reject a few drafts, then read `editDistance` in `decisions.jsonl` and argue for a threshold that would move a category from `draft-for-approval` to `auto-send`. **Check:** you name numbers. The reference gate in `expected-output.md` is 90 days of review, at least 200 reviewed messages, a median edit distance under 5 percent of draft length, and zero decisions tagged as factual corrections. Edit distance can't tell a comma from a lawsuit, so that last gate needs a field the review UI asks for.

## What Is in This Folder

- `data/inquiries.jsonl`: six inquiries drawn from feature 07's full 100-message inbox, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `data/snippets/`: the four park-doc excerpts, one file per document number, each holding only the sections that answer its inquiry, taken from the full park documents in feature 05's `data/park-docs/` and quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
- `dotnet/`, `python/`, `typescript/`: each has this lab for its language, a `starter/` to edit, and a `complete/` answer key
