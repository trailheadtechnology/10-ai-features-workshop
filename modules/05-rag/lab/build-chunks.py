"""Rebuild chunks.jsonl from data/park-docs/.

You do not need to run this for the lab; chunks.jsonl ships ready to use. It is
here because chunking is the decision that quietly broke this module, and a lab
about RAG should let you change that decision and measure what happens.

The strategy, and why each part exists:

  Parent sections stay whole under MAX_SECTION_WORDS, because a short section is
  one idea and splitting it strands half of that idea.

  Longer sections split at the numbered subsection boundaries the document
  already provides, since regulations state one rule per subsection.

  MIN_CHUNK_WORDS packs consecutive subsections together until a chunk is big
  enough to stand alone. Without this floor, a 19-word subsection ("Fuel is not
  available anywhere within the Park") retrieves as if it were an answer.

  Every chunk carries the document title and its section heading, so a retrieved
  fragment still says where it came from.

  Every chunk ends with the first sentence of whatever comes NEXT in the
  document. That forward pointer is the fix for the bug this module documents:
  Section 4.1 said fires are permitted when fire danger is low, Section 4.2 said
  they are prohibited at Sperry year round, they landed in different chunks, and
  the model confidently told visitors to light a fire. Overlap is forward only
  on purpose. Carrying the PREVIOUS section's tail would paste "fires permitted"
  onto the top of the prohibition chunk and reintroduce the bug from the other
  side.

Usage:
    python3 build-chunks.py [out.jsonl] [max_section_words] [min_chunk_words]

Try it with different numbers, re-run the lab, and see what moves. Sensible
experiments: drop MIN_CHUNK_WORDS to 0 and watch the unanswerable question start
getting answered; raise MAX_SECTION_WORDS past 400 and watch precision fall.
"""
import json, os, re, sys

HERE = os.path.dirname(os.path.abspath(__file__))
SRC = os.path.normpath(os.path.join(HERE, "..", "..", "..", "data", "park-docs"))
OUT = sys.argv[1] if len(sys.argv) > 1 else os.path.join(HERE, "chunks.jsonl")

# A section under this many words is one idea; keep it whole. Over it, split at the
# subsection boundaries the document already provides.
MAX_SECTION_WORDS = int(sys.argv[2]) if len(sys.argv) > 2 else 200
MIN_CHUNK_WORDS = int(sys.argv[3]) if len(sys.argv) > 3 else 50

ABBREV = re.compile(r"(?:\b[A-Z]|No|Nos|Sec|Secs|Fig|e\.g|i\.e|approx|a\.m|p\.m|Mr|Mrs|Ms|Dr|St|Ft|vs|U\.S)$")


def first_sentence(text, max_words=45):
    """First sentence of a passage, for the forward-overlap tail."""
    flat = " ".join(text.split())
    # strip a leading subsection number and markdown so the carried text reads as prose
    flat = re.sub(r"^\d+\.\d+\s+", "", flat)
    flat = flat.replace("**", "").replace("*", "")
    pos = 0
    while True:
        m = re.compile(r"[.!?](?=[\s\)]|$)").search(flat, pos)
        if not m:
            return " ".join(flat.split()[:max_words])
        head = flat[: m.start()]
        last = head.split()[-1] if head.split() else ""
        # a period after a digit or a known abbreviation is not a sentence end
        if re.search(r"\d$", last) or ABBREV.search(last):
            pos = m.end()
            continue
        sent = flat[: m.end()]
        words = sent.split()
        if len(words) > max_words:
            return " ".join(words[:max_words])
        return sent


def split_sections(body):
    """[(heading, section_number, body_text)] for '## N. Title' headings."""
    parts = re.split(r"^(##\s+(\d+)\.\s+.*)$", body, flags=re.M)
    preamble = parts[0]
    out = []
    for i in range(1, len(parts), 3):
        out.append((parts[i].strip(), int(parts[i + 1]), parts[i + 2].strip()))
    return preamble.strip(), out


def split_subsections(section_body, number):
    """[(label, text)] per numbered subsection, or [] if the section has none."""
    pat = re.compile(r"^(%d\.(\d+))\s" % number, re.M)
    marks = list(pat.finditer(section_body))
    if len(marks) < 2:
        return []
    out = []
    for i, m in enumerate(marks):
        end = marks[i + 1].start() if i + 1 < len(marks) else len(section_body)
        out.append((m.group(1), section_body[m.start():end].strip()))
    # anything before the first subsection (a lead-in paragraph) rides along with it
    lead = section_body[: marks[0].start()].strip()
    if lead:
        out[0] = (out[0][0], lead + "\n\n" + out[0][1])
    return out


def pack(subs):
    """Merge consecutive subsections until each chunk clears the word floor."""
    if not subs:
        return []
    out, buf_labels, buf_text, count = [], [], [], 0
    for label, text in subs:
        buf_labels.append(label.split(".")[1])
        buf_text.append(text)
        count += len(text.split())
        if count >= MIN_CHUNK_WORDS:
            out.append((label_range(buf_labels), "\n\n".join(buf_text)))
            buf_labels, buf_text, count = [], [], 0
    if buf_text:  # a short tail rides along with the chunk before it
        if out:
            prev_label, prev_text = out[-1]
            merged = label_range([prev_label.split("-")[0]] + buf_labels)
            out[-1] = (merged, prev_text + "\n\n" + "\n\n".join(buf_text))
        else:
            out.append((label_range(buf_labels), "\n\n".join(buf_text)))
    return out if len(out) > 1 else []


def label_range(labels):
    return labels[0] if len(labels) == 1 else "%s-%s" % (labels[0], labels[-1])


def build(path):
    slug = os.path.splitext(os.path.basename(path))[0]
    raw = open(path, encoding="utf-8").read()
    title = re.search(r"^#\s+(.*)$", raw, flags=re.M).group(1).strip()
    preamble, sections = split_sections(raw)

    # units: (chunk_id_suffix, chunk_text_parts, pointer_label, body_for_overlap)
    units = []
    units.append(("00", [preamble], "Section 1", preamble))
    for heading, number, body in sections:
        name = re.sub(r"^##\s+\d+\.\s+", "", heading).strip()
        subs = split_subsections(body, number) if len(body.split()) > MAX_SECTION_WORDS else []
        subs = pack(subs)
        if subs:
            for label, text in subs:
                units.append(("%02d.%s" % (number, label), [heading, text],
                              "%d.%s" % (number, label), text))
        else:
            units.append(("%02d" % number, [heading, body],
                          "Section %d, %s" % (number, name), body))

    chunks = []
    for i, (suffix, parts, _, _) in enumerate(units):
        parts = ["[%s]" % title] + [p.strip() for p in parts]
        if i + 1 < len(units):
            nxt_label, nxt_body = units[i + 1][2], units[i + 1][3]
            parts.append("(continues in %s) %s" % (nxt_label, first_sentence(nxt_body)))
        chunks.append({
            "chunk_id": "%s:%s" % (slug, suffix),
            "source": os.path.basename(path),
            "text": "\n\n".join(parts),
        })
    return chunks


all_chunks = []
for name in sorted(os.listdir(SRC)):
    if name.endswith(".md"):
        all_chunks += build(os.path.join(SRC, name))

with open(OUT, "w", encoding="utf-8") as f:
    for c in all_chunks:
        f.write(json.dumps(c, ensure_ascii=False) + "\n")

lens = [(len(c["text"].split()), c["chunk_id"]) for c in all_chunks]
lens.sort(reverse=True)
print("chunks:", len(all_chunks))
print("longest:", lens[:5])
print("shortest:", lens[-3:])
