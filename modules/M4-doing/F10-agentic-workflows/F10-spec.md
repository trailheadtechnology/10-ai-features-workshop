# 10 · Agentic Workflows (Capstone)

Module 4: Doing · Runs on `gpt-5.5` on Microsoft Foundry (Ollama fallback) · A 60-minute module of its own: 10 minutes of demo, 50 of building

## The User Problem

A Trailhead Guides user types: "Plan me a 3-day trip in Glacier for mid-September." Fulfilling that today means the user does everything themselves: check the weather forecast, search for trails, cross-reference current conditions, check campsite availability, and file a permit request in a separate system. That is five tools and one afternoon, and the app helped with exactly one step. The user didn't want search results; they wanted the trip planned.

## The Concept

An agent is an LLM given tools and a goal, running in a loop: the model decides which tool to call, your code executes it, the result goes back to the model, and it continues until the goal is met. Tool calling is the mechanism underneath, where you describe your functions (name, parameters, what they do) and the model responds with "call `check_weather` with `park=Glacier, dates=Sept 14-16`" instead of prose. Your code stays in charge of actually doing things; the model only ever chooses.

This is the capstone because it composes the day. The agent searches trails semantically (04), grounds itself in current condition reports (08's data), works with structured tool results (02's lesson in reverse), and pauses for human confirmation before filing the permit (09's policy, applied). It's also the feature where model choice stops being negotiable. A dropped or malformed tool call doesn't degrade the experience, it breaks the loop, so this feature runs on Azure OpenAI rather than gambling the finale on a local model. And because an agent acts instead of answering, guardrails are part of the design: a step budget so it can't loop forever, and a confirmation gate before anything irreversible.

## The Lab

The hands-on lab is [F10-lab.md](F10-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is a Challenge lab, for anyone who finished the module's Recommended lab and wants another.

## Leadership Beat

- **When to reach for this:** multi-step workflows that span systems, where users state a goal and then do the clicking themselves. Trip planning, onboarding, procurement, incident response, report assembly.
- **Rough cost & effort:** the highest of the ten. Weeks to months, frontier-model API costs, and real design work on guardrails, confirmation gates, and observability. Do one of features 01-09 first; do this when the goal-shaped problem is worth it.
- **The one-liner for your CTO:** "Users state the goal in one sentence, the software does the clicking, and a human still signs off on anything that matters."

This card is row 10 of the [decision framework](../../../docs/decision-framework.md).
