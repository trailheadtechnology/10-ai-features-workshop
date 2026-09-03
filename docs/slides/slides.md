# Slides

Six decks, one per session. The outlines in [`outlines/`](outlines/) are the source: every slide's title, on-slide bullets, and speaker notes. They are also the instructor script. Each module outline opens with a runsheet (the 30/60 or 10/50 split and what to cut if behind), and the notes under each "Demo: What to Watch" slide carry the full step-by-step demo script, which used to live in the feature specs. The specs in `modules/` are for attendees now. The `.pptx` files in [`pptx/`](pptx/) are generated from them by `build.py` (needs `python-pptx`), themed from `template.pptx` — the masters, layouts, and boilerplate slides (title, promise, about-me, thanks) extracted from the Power of Ten deck. Slides advance-to-animate: lists and flow diagrams reveal one item or box per advance, so one outline slide usually becomes several `.pptx` slides. The `[marker]` syntax (`[title]`, `[promise]`, `[about]`, `[thanks]`, `[section]`, `[demo]`, `[big]`, `[static]`) is documented in `build.py`. Edit the outline, re-run the build, and the deck follows.

| Deck | Session | Outline |
|---|---|---|
| 0 | Opening, 30 min | [M0-opening.md](outlines/M0-opening.md) |
| 1 | Module 1: Understanding, 90 min | [M1-understanding.md](outlines/M1-understanding.md) |
| 2 | Module 2: Finding, 90 min | [M2-finding.md](outlines/M2-finding.md) |
| 3 | Module 3: Deciding, 90 min | [M3-deciding.md](outlines/M3-deciding.md) |
| 4 | Module 4: Doing, 60 min | [M4-doing.md](outlines/M4-doing.md) |
| 5 | Closing, 30 min | [M5-closing.md](outlines/M5-closing.md) |

The opening deck follows the talk structure: title, setup and smoke test, the Avalanche Lake story, the promise slide, the day's agenda, the about-me/free-offer slide, then the framing content. The module decks share a shape: a module section slide, the module's thread, a comparison diagram of the module's features side by side, then per feature a numbered section slide, the user problem, the concept, a how-it-works flow diagram, a DEMO slide, what to watch in the demo, and the leadership card, then the hands-on menu and a debrief slide. The closing deck ends with an action-oriented "Your Move" section and the "Thanks! Questions?" slide. The diagrams are drawn by `build.py` from `Flow:` lines in the outline, so they stay editable as text. The Microsoft.Extensions.AI provider-swap slide ("One Line of Code, Any Provider") lives in the opening deck and is referenced from features 03 and 05.

Everything on the slides is drawn from the module overviews and feature specs; where a slide quotes a number, the run behind it is in that feature's `expected-output.md`.
