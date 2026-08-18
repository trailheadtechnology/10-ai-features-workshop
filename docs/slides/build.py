"""Build the plain .pptx decks from the outlines.

    python3 build.py            # writes pptx/*.pptx from outlines/*.md

Each outline is markdown: `## N. Title` starts a slide, `- ` lines are bullets
(two-space indent nests one level), a `Notes:` paragraph becomes the speaker
notes, and any other paragraph is body text. Once a slide has a `Notes:` line,
everything after it up to the next `##` is also speaker notes (that is where the
per-feature demo scripts live). Everything before the first `##` is the outline's
own preamble (module runsheet) and is skipped. Requires python-pptx.
"""

import re
import sys
from pathlib import Path

from pptx import Presentation
from pptx.util import Inches, Pt

HERE = Path(__file__).parent
OUTLINES = HERE / "outlines"
OUT = HERE / "pptx"

DECKS = [
    ("00-opening.md", "00-opening.pptx"),
    ("M1-understanding.md", "M1-understanding.pptx"),
    ("M2-finding.md", "M2-finding.pptx"),
    ("M3-deciding.md", "M3-deciding.pptx"),
    ("M4-doing.md", "M4-doing.pptx"),
    ("05-closing.md", "05-closing.pptx"),
]


def clean(text):
    """Drop markdown emphasis and code markers; slides don't render them."""
    text = re.sub(r"\*\*(.+?)\*\*", r"\1", text)
    text = re.sub(r"(?<!\w)\*(.+?)\*(?!\w)", r"\1", text)
    text = text.replace("`", "")
    return text.strip()


def parse(path):
    """Return a list of slides: {title, items:[(level,text)], paras:[str], notes:str}."""
    slides = []
    cur = None
    in_notes = False
    for raw in path.read_text().splitlines():
        line = raw.rstrip()
        m = re.match(r"^## \d+\.\s+(.*)$", line)
        if m:
            cur = {"title": clean(m.group(1)), "items": [], "paras": [], "notes": []}
            slides.append(cur)
            in_notes = False
            continue
        if cur is None or not line.strip() or line.strip() == "---":
            continue
        if line.startswith("Notes:"):
            cur["notes"].append(clean(line[len("Notes:"):]))
            in_notes = True
            continue
        if in_notes:
            cur["notes"].append(clean(line))
            continue
        m = re.match(r"^(\s*)-\s+(.*)$", line)
        if m:
            level = 1 if len(m.group(1)) >= 2 else 0
            cur["items"].append((level, clean(m.group(2))))
            continue
        cur["paras"].append(clean(line))
    for s in slides:
        s["notes"] = "\n".join(s["notes"])
    return slides


def place(slide, prs, title_slide):
    """The default template's placeholders are laid out for 4:3, and their
    geometry is inherited from the layout, so set all four sides explicitly
    (setting only left/width writes a zero top/height and the text collapses)."""
    margin = Inches(0.67)
    width = prs.slide_width - 2 * margin
    title, body = slide.placeholders[0], slide.placeholders[1]
    if title_slide:
        title.left, title.top, title.width, title.height = margin, Inches(2.2), width, Inches(1.5)
        body.left, body.top, body.width, body.height = margin, Inches(3.9), width, Inches(2.0)
    else:
        title.left, title.top, title.width, title.height = margin, Inches(0.4), width, Inches(1.2)
        body.left, body.top, body.width, body.height = margin, Inches(1.75), width, Inches(5.25)


def build(outline, target):
    prs = Presentation()
    prs.slide_width = Inches(13.333)  # 16:9; python-pptx's default template is 4:3
    prs.slide_height = Inches(7.5)
    title_layout = prs.slide_layouts[0]
    body_layout = prs.slide_layouts[1]

    for i, s in enumerate(parse(outline)):
        if i == 0:
            slide = prs.slides.add_slide(title_layout)
            slide.shapes.title.text = s["title"]
            sub = slide.placeholders[1].text_frame
            lines = [t for _, t in s["items"]] + s["paras"]
            sub.text = lines[0] if lines else ""
            for extra in lines[1:]:
                sub.add_paragraph().text = extra
        else:
            slide = prs.slides.add_slide(body_layout)
            slide.shapes.title.text = s["title"]
            tf = slide.placeholders[1].text_frame
            tf.word_wrap = True
            first = True
            for para in s["paras"]:
                p = tf.paragraphs[0] if first else tf.add_paragraph()
                first = False
                p.text = para
                p.level = 0
                for r in p.runs:
                    r.font.size = Pt(20)
            for level, text in s["items"]:
                p = tf.paragraphs[0] if first else tf.add_paragraph()
                first = False
                p.text = text
                p.level = level
                for r in p.runs:
                    r.font.size = Pt(20 if level == 0 else 18)
        place(slide, prs, i == 0)
        if s["notes"]:
            slide.notes_slide.notes_text_frame.text = s["notes"]

    OUT.mkdir(exist_ok=True)
    prs.save(OUT / target)
    return len(prs.slides)


if __name__ == "__main__":
    for src, dst in DECKS:
        n = build(OUTLINES / src, dst)
        print(f"{dst}: {n} slides")
    sys.exit(0)
