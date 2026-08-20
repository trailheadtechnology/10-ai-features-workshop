"""Build the plain .pptx decks from the outlines.

    python3 build.py            # writes pptx/*.pptx from outlines/*.md

Each outline is markdown: `## N. Title` starts a slide, `- ` lines are bullets
(two-space indent nests one level), a `Notes:` paragraph becomes the speaker
notes, a `Flow: A -> B -> C` line draws a left-to-right box-and-arrow diagram
(`Flow (label): ...` puts a label to its left; several Flow lines stack as rows),
and any other paragraph is body text. Once a slide has a `Notes:` line,
everything after it up to the next `##` is also speaker notes (that is where the
per-feature demo scripts live). Everything before the first `##` is the outline's
own preamble (module runsheet) and is skipped. Requires python-pptx.
"""

import re
import sys
from pathlib import Path

from pptx import Presentation
from pptx.dml.color import RGBColor
from pptx.enum.shapes import MSO_SHAPE
from pptx.enum.text import MSO_ANCHOR, PP_ALIGN
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
            cur = {"title": clean(m.group(1)), "items": [], "paras": [], "notes": [], "flows": []}
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
        m = re.match(r"^Flow(?:\s*\((.*?)\))?:\s*(.*)$", line)
        if m:
            boxes = [clean(b) for b in m.group(2).split("->")]
            cur["flows"].append((clean(m.group(1) or ""), boxes))
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


def draw_flows(slide, prs, flows, top):
    """Rows of rounded boxes joined by arrows, black on white like the rest of the deck.
    Returns the bottom edge so the body text can start below the diagram."""
    margin = Inches(0.67)
    label_w = Inches(1.6) if any(lbl for lbl, _ in flows) else 0
    left0 = margin + label_w
    avail = prs.slide_width - margin - left0
    gap_h = Inches(0.3)
    y = top
    for label, boxes in flows:
        n = len(boxes)
        arrow_w = Inches(0.45)
        box_w = (avail - arrow_w * (n - 1)) / n
        pt = 15 if n <= 4 else 13
        # rough fit: ~9 chars per inch at 13pt, ~8 at 15pt; grow the row for wordy boxes
        cpl = max(6, int((box_w / 914400 - 0.16) * (8 if pt == 15 else 9)))
        lines = max(-(-len(b) // cpl) for b in boxes)
        row_h = max(Inches(0.9), Inches(0.35 + 0.26 * lines))
        if label:
            tb = slide.shapes.add_textbox(margin, y, label_w - Inches(0.1), row_h)
            tf = tb.text_frame
            tf.word_wrap = True
            tf.vertical_anchor = MSO_ANCHOR.MIDDLE
            tf.text = label
            for r in tf.paragraphs[0].runs:
                r.font.size = Pt(14)
                r.font.italic = True
        x = left0
        for i, text in enumerate(boxes):
            shp = slide.shapes.add_shape(MSO_SHAPE.ROUNDED_RECTANGLE, x, y, box_w, row_h)
            shp.fill.background()
            shp.line.color.rgb = RGBColor(0, 0, 0)
            shp.line.width = Pt(1.5)
            shp.shadow.inherit = False
            tf = shp.text_frame
            tf.word_wrap = True
            tf.vertical_anchor = MSO_ANCHOR.MIDDLE
            tf.margin_left = tf.margin_right = Inches(0.08)
            tf.text = text
            para = tf.paragraphs[0]
            para.alignment = PP_ALIGN.CENTER
            for r in para.runs:
                r.font.size = Pt(pt)
                r.font.color.rgb = RGBColor(0, 0, 0)
            x += box_w
            if i < n - 1:
                ar = slide.shapes.add_shape(MSO_SHAPE.RIGHT_ARROW, x + Inches(0.06), y + row_h / 2 - Inches(0.16), arrow_w - Inches(0.12), Inches(0.32))
                ar.fill.solid()
                ar.fill.fore_color.rgb = RGBColor(0, 0, 0)
                ar.line.fill.background()
                ar.shadow.inherit = False
                x += arrow_w
        y += row_h + gap_h
    return y - gap_h


def place(slide, prs, title_slide, body_top=None):
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
        top = body_top if body_top is not None else Inches(1.75)
        body.left, body.top, body.width, body.height = margin, top, width, Inches(7.1) - top


def build(outline, target):
    prs = Presentation()
    prs.slide_width = Inches(13.333)  # 16:9; python-pptx's default template is 4:3
    prs.slide_height = Inches(7.5)
    title_layout = prs.slide_layouts[0]
    body_layout = prs.slide_layouts[1]

    for i, s in enumerate(parse(outline)):
        body_top = None
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
            base = 20
            if s["flows"]:
                room = Inches(7.1) - (draw_flows(slide, prs, s["flows"], Inches(1.7)) + Inches(0.3))
                base = 16 if room > Inches(2.0) else 14
                body_top = Inches(7.1) - room
            first = True
            for para in s["paras"]:
                p = tf.paragraphs[0] if first else tf.add_paragraph()
                first = False
                p.text = para
                p.level = 0
                for r in p.runs:
                    r.font.size = Pt(base)
            for level, text in s["items"]:
                p = tf.paragraphs[0] if first else tf.add_paragraph()
                first = False
                p.text = text
                p.level = level
                for r in p.runs:
                    r.font.size = Pt(base if level == 0 else base - 2)
        place(slide, prs, i == 0, body_top)
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
