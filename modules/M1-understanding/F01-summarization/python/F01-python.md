# Python Demo for 01 Summarization

Two scripts, both reading the trip reports in [`../data/`](../data/):

- `starter/main.py`: the demo's starting point. One `OpenAI` client pointed at Ollama, one call, the naive "Summarize this trip report" prompt. Run it on `tr-0001.md` and read the book report.
- `complete/main.py`: the finished demo as shown on stage. Same call, three prompts selected by flag: the naive one, the 3-bullet hiker briefing with the grounding lines, and the one-line headline for a card UI. `--audience ranger` swaps who the briefing is for.

Setup once, from this `python/` directory. A virtual environment is not optional on a modern macOS or Linux Python (`pip install` outside one is refused), and activating it is what puts `python` and `pip` on your path:

```bash
python3 -m venv .venv
source .venv/bin/activate        # Windows: .venv\Scripts\activate
pip install -r requirements.txt
```

Then, with the venv active, from `complete/`. (`starter/main.py` takes no flags, at most the one positional argument its header comment names, same as the .NET starter.)

```bash
python main.py                              # naive prompt on the buried-hazard report
python main.py --briefing                   # 3 bullets; the bridge surfaces
python main.py --headline
python main.py --briefing --audience ranger
python main.py ../../data/tr-0001.md        # any report path works
```

The client is the official `openai` package pointed at Ollama's OpenAI-compatible endpoint (`http://localhost:11434/v1`), which is the same trick the TypeScript version uses and the Python equivalent of the .NET demo's `IChatClient`: switching to Azure OpenAI later is a different constructor and nothing else. Real output and the measured hazard-invention rate behind the briefing prompt's last two lines are in [`../expected-output.md`](../expected-output.md).

## Lab Walkthrough: From `starter/` to `complete/`

The steps in [`../F01-lab.md`](../F01-lab.md), done in Python: start from `starter/main.py` and end where `complete/main.py` is. Edit the starter in place (or copy it first); `complete/` is the answer key, and its comments say why each piece is there. Run from the `starter/` directory with the venv active; the flags shown for later steps are the ones `complete/` supports, so add the same argument parsing or hard-code the value.

### Step 1: Run the Starter As-Is and Read the Book Report

The starter is spec step 1: the naive prompt on the clean report, `tr-0001.md`. Run it twice. It is faithful, generic, and useless, and that is the baseline you are improving on.

Run:

```bash
python main.py
```

Check: A paragraph or two about the author's gear and their day. Nothing a hiker planning Saturday could act on.

### Step 2: Rewrite the Prompt Into the 3-bullet Briefing (lab step 2)

Replace the naive prompt with one that demands exactly three bullets (conditions, hazards or closures, crowding) and nothing else. Start with just that, and run it four or five times on the clean report before adding anything: a prompt that requires a hazard bullet will invent one from a bear sighting or the word "avalanche" in the trail name. When you see that happen, add the last three lines below. They give the model a legal way to report nothing, and they are the only reason the finished prompt is trustworthy.

```python
prompt = f"""You are helping a hiker planning to hike this trail within the next week.
From the trip report below, produce exactly 3 bullets covering:
current trail conditions, hazards or closures, and crowding.
Ignore gear talk, personal stories, and scenery.
Report only what the trip report states. Do not turn a wildlife sighting into a
hazard or a closure, and write "no closures or hazards reported" when it says none.
If the report does state a closure or hazard, it must appear in the first bullet.

{report}"""
response = client.chat.completions.create(model="llama3.2", messages=[{"role": "user", "content": prompt}])
```

Run:

```bash
python main.py   # several times
```

Check: Three bullets, and on `tr-0001.md` the hazards bullet says nothing is closed, every run. The measured invention rate without the last three lines is in `../expected-output.md`.

### Step 3: Run the Buried-Hazard Report Through the Same Prompt (lab step 3)

The prompt stays the same and the report changes: `tr-0004.md` mentions the washed-out footbridge in passing, halfway down; a good briefing puts it first.

Run:

```bash
python main.py ../../data/tr-0004.md
```

Check: The first bullet is the closure. Compare the sample in `../expected-output.md`. If the bridge is missing or lands third, tighten the "must appear in the first bullet" line.

### Step 4: Stretch: Switch the Audience

Same call, same report, one variable swapped into the first line of the prompt. This is what `--audience ranger` does in `complete/`.

```python
audience_focus = (
    "a park ranger who cares about maintenance issues, closures, safety incidents, and visitor impacts, not scenery"
    if audience == "ranger" else "a hiker planning to hike this trail within the next week")
# then: You are helping {audience_focus}.
```

Check: The ranger version drops the crowding chatter and leads with the bridge as a maintenance item. `complete/` also has the one-line `--headline` shape: same call, a 12-word instruction, a different UI slot.
