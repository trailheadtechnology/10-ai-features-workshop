# 01 · Summarization

Module 1: Understanding · Runs on Ollama (`llama3.2`)

## The user problem

A hiker planning this weekend's trip opens Trailhead Guides and finds forty trip reports for Avalanche Lake Trail, each one 1,200 words of trail diary, gear opinions, and granola recipes. Somewhere in there is the one thing they need to know: is the bridge out, and are the mosquitoes bad? Nobody reads forty essays. They skim three, miss the warning in the fourth, and have a bad Saturday.

## The concept

Summarization is the simplest possible LLM feature: one chat-completion call with a document and an instruction. You don't need fine-tuning, a vector database, or any pipeline at all. That makes it the right first feature, because by the end of the lab everyone in the room has called a model and built something useful.

The craft is all in the instruction. "Summarize this" produces a book report. Real products ask for a summary with a purpose: "In 3 bullets, tell a hiker planning a trip this week about current conditions, hazards, and crowding. Ignore gear talk." The second lesson is that summaries can be shaped to fit the UI slot that needs them: plain prose, bullets, a fixed template, or a single headline. A small local model handles all of this well, which is why this feature never touches the cloud.

## Demo outline (about 12 min, .NET)

1. Open a raw trip report from `data/trip-reports/` on screen and scroll through it slowly, so the room feels the problem before seeing the fix.
2. Starter project: `IChatClient` wired to Ollama via Microsoft.Extensions.AI. One method, one prompt: "Summarize this trip report."
3. Run it and get a faithful, generic, useless book report. Name the failure out loud: right model, wrong instruction.
4. Iterate the prompt live into the hiker-focused version (conditions, hazards, crowding, ignore everything else). Run again. This is the payoff: three bullets, and the bridge warning surfaces.
5. Point at the last two lines of that prompt, the grounding lines. Without them, this exact prompt run against the clean report invented a bear-related trail closure in roughly half of runs, because a required "hazards" slot with nothing honest to put in it is an invitation to make something up. The numbers and the failing outputs are in `lab/expected-output.md`. First prompt written is rarely the prompt shipped.
6. Change the shape: same call, but output a one-line "trail status" headline for a card UI. Same feature, different product surface.
7. Point at the Ollama endpoint in the code. This ran entirely on the laptop, with no API key and no data leaving the room.

## Lab spec (Core lab, any language)

*Everyone does this one. It is the Core lab for [Module 1](../M1-overview.md), and the hands-on period runs about 45 minutes, so there is room to do it properly rather than fast.*

- **Goal:** turn a raw trip report into a 3-bullet "conditions briefing" for hikers.
- **Input:** two trip reports provided in `lab/`, one with a buried hazard warning, drawn from `data/trip-reports/`.
- **How:** POST to Ollama's chat endpoint (`llama3.2`). `lab/ollama.http` has the exact request; port it to your language or run it as-is.
- **Steps:**
  1. Send report #1 with the naive prompt ("summarize this") and read the book report you get back.
  2. Rewrite the prompt to demand exactly 3 bullets covering conditions, hazards, and crowding, and nothing else. Run it a few times, not once: report #1 has no closure in it, and a prompt that demands a hazard bullet will happily invent one.
  3. Run report #2 through your improved prompt. Success check: your 3 bullets surface the buried hazard (compare `lab/expected-output.md`), and report #1 still comes back with nothing closed.
- **Stretch goal:** make the summary audience-switchable. The same report, summarized for a hiker and then for a park ranger who cares about maintenance issues rather than scenery.

## Leadership beat

- **When to reach for this:** anywhere users face long content they don't want to read. Reviews, tickets, reports, meeting notes, email threads.
- **Rough cost & effort:** days, not months. One API call per document, running on free local models, with no training and no new infrastructure.
- **The one-liner for your CTO:** "Our users are drowning in text we already have. A weekend of prompt work turns it into answers."

This card is row 1 of the [decision framework](../../../docs/decision-framework.md).
