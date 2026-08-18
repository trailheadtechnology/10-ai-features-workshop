# 07 · Classification & Routing

Module 3: Deciding · Runs on Ollama (`llama3.2`)

## The User Problem

Every message to Trailhead Guides lands in one inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, and, occasionally, someone reporting an actual emergency. A ranger triages the pile by hand once or twice a day. The permit request waits behind the granola questions, the complaint goes to the wrong person twice, and the emergency sits unread for four hours. The users' real problem is that their message goes into a hole, and how fast it comes out depends on luck.

## The Concept

This is classification again (feature 03 was the warm-up), but now the label has consequences: it decides where the message goes and how fast. An LLM makes a solid zero-shot classifier. You describe the categories in plain language and it labels messages with no training data, which is exactly the situation most teams are in on day one.

Two design decisions carry the feature. The first is that the taxonomy is the product: category names and one-sentence descriptions in the prompt are where accuracy lives, and when the model misfiles something, you usually fix the description rather than the model. The second is that errors are not symmetric, since misrouting a complaint costs a day of annoyance and misrouting an emergency is a headline. So the system needs an "unsure" route to a human, and it should be tuned so the expensive class never slips through, even at the cost of extra false alarms. Recall on the class that matters, not overall accuracy, is the number to watch. A local model does this fine, and at inbox volume, free matters.

## The Lab

The hands-on lab is [F07-lab.md](F07-lab.md): the goal, the steps, the success checks, and the stretch goal, with a walkthrough for each track in `http/`, `dotnet/`, `python/`, and `typescript/`. It is the Recommended lab for its module: start here unless you have a reason not to.

## Leadership Beat

- **When to reach for this:** any shared inbox, ticket queue, or intake form where a human sorts before anyone acts. Support, sales leads, HR requests, incoming documents.
- **Rough cost & effort:** days. No training data needed to start; the ongoing work is refining the taxonomy as real traffic reveals edge cases. Local models make per-message cost roughly zero.
- **The one-liner for your CTO:** "Every message reaches the right person in seconds, and the urgent ones stop waiting in line."

This card is row 7 of the [decision framework](../../../docs/decision-framework.md).
