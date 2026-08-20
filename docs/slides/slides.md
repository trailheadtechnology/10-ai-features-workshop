# Slides

Six decks, one per session. The outlines in [`outlines/`](outlines/) are the source: every slide's title, on-slide bullets, and speaker notes. They are also the instructor script. Each module outline opens with a runsheet (the 30/60 or 10/50 split and what to cut if behind), and the notes under each "Demo: What to Watch" slide carry the full step-by-step demo script, which used to live in the feature specs. The specs in `modules/` are for attendees now. The `.pptx` files in [`pptx/`](pptx/) are generated from them by `build.py` (needs `python-pptx`) and stay plain (black text, white background, default template) until the content settles. Edit the outline, re-run the build, and the deck follows.

| Deck | Session | Outline |
|---|---|---|
| 0 | Opening, 30 min | [00-opening.md](outlines/00-opening.md) |
| 1 | Module 1: Understanding, 90 min | [M1-understanding.md](outlines/M1-understanding.md) |
| 2 | Module 2: Finding, 90 min | [M2-finding.md](outlines/M2-finding.md) |
| 3 | Module 3: Deciding, 90 min | [M3-deciding.md](outlines/M3-deciding.md) |
| 4 | Module 4: Doing, 60 min | [M4-doing.md](outlines/M4-doing.md) |
| 5 | Closing, 30 min | [05-closing.md](outlines/05-closing.md) |

The module decks share a shape: the module's thread, a comparison diagram of the module's features side by side, then per feature the user problem, the concept, a how-it-works flow diagram, what to watch in the demo, and the leadership card, then the hands-on menu and a debrief slide. The diagrams are drawn by `build.py` from `Flow:` lines in the outline, so they stay editable as text. The Microsoft.Extensions.AI provider-swap slide lives in the opening deck (slide 14) and is referenced from features 03 and 05.

Everything on the slides is drawn from the module overviews and feature specs; where a slide quotes a number, the run behind it is in that feature's `expected-output.md`.
