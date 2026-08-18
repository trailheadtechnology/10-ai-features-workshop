# 01 · Summarization

Module 1: Understanding · Runs on Ollama (`llama3.2`)

## The User Problem

A hiker planning this weekend's trip opens Trailhead Guides and finds forty trip reports for Avalanche Lake Trail, each one 1,200 words of trail diary, gear opinions, and granola recipes. Somewhere in there is the one thing they need to know: is the bridge out, and are the mosquitoes bad? Nobody reads forty essays; they skim three, miss the warning in the fourth, and have a bad Saturday.

## The Concept

Summarization is the simplest possible LLM feature: one chat-completion call with a document and an instruction. You don't need fine-tuning, a vector database, or any pipeline at all. That makes it the right first feature, because by the end of the lab everyone in the room has called a model and built something useful.

The craft is all in the instruction, because "summarize this" produces a book report and real products ask for a summary with a purpose: "In 3 bullets, tell a hiker planning a trip this week about current conditions, hazards, and crowding. Ignore gear talk." The second lesson is that summaries can be shaped to fit the UI slot that needs them: plain prose, bullets, a fixed template, or a single headline. A small local model handles all of this well, which is why this feature never touches the cloud.

## The Lab

The hands-on lab is [F01-lab.md](F01-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is the Recommended lab for its module: start here unless you have a reason not to.

## Leadership Beat

- **When to reach for this:** anywhere users face long content they don't want to read. Reviews, tickets, reports, meeting notes, email threads.
- **Rough cost & effort:** days, not months. One API call per document, running on free local models, with no training and no new infrastructure.
- **The one-liner for your CTO:** "Our users are drowning in text we already have. A weekend of prompt work turns it into answers."

This card is row 1 of the [decision framework](../../../docs/decision-framework.md).
