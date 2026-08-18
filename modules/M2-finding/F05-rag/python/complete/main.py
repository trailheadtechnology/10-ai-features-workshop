"""Finished demo, matching the demo script in docs/slides/outlines:
  python main.py                                    the Sperry Chalet question, grounded
  python main.py "Is the Avalanche Lake Trail open right now?"
  python main.py --no-context                       step 1: the confident wrong answer
  python main.py --alpha 1.0                        pure cosine, the wrong-park neighbors
  python main.py --retrieval-only                   print the score table and stop
  python main.py --top-k 8 "your question"          vary retrieval depth
  python main.py --model qwen3:32b                  a bigger model, if you have the memory

Retrieval is hybrid: normalized cosine similarity blended with a BM25-lite
lexical score, so a distinctive proper noun like "Sperry" counts for something.
Chunks are one numbered subsection each wherever a section ran long enough to
hold a rule and the exception that overrides it; see ../../expected-output.md.
Every generated answer's [chunk-id] citations are validated against the chunks
that were actually retrieved.

Retrieval always runs locally (nomic-embed-text). Generation uses Azure OpenAI
when AZURE_OPENAI_ENDPOINT / AZURE_OPENAI_KEY / AZURE_OPENAI_DEPLOYMENT are set,
and falls back to the local model chosen just below when they are not.
"""

import json
import math
import os
import re
import sys
from collections import Counter
from pathlib import Path

from openai import AzureOpenAI, OpenAI

# ---------------------------------------------------------------------------
# Which local model generates the answer.
#
# Everything else in this file is identical either way: same retrieval, same
# chunks, same prompt. Only the model changes. All numbers below were measured
# on this pipeline and are written up in ../../expected-output.md.
#
#                            llama3.2 (3B)      qwen3:32b
#   answers correctly           97%               100%
#   opens with "Yes" before
#     saying fires are banned    40%                0%
#   refuses when it shouldn't    8%                 0%
#   invalid citations         9 per 200 runs      0 per 80 runs
#   download / memory           2 GB / any        20 GB / ~24 GB free
#   time per answer             0.9 s             16.6 s
#
# The big model is better at everything and costs 18x the wall clock and a
# machine most people do not have. Switch only if you have the memory; on 16 GB
# it will not load, and on 32 GB it will make you wait. This setting only affects
# generation. Retrieval is identical either way, so no model here can rescue an
# answer that is not in the retrieved chunks.
#
# local_model = "qwen3:32b"    # see the memory note above before switching
local_model = "llama3.2"       # the default, because it runs on any laptop in the room
# ---------------------------------------------------------------------------

DATA = Path(__file__).resolve().parents[2] / "data"
chunks_path = DATA / "chunks.jsonl"
cache_path = Path(__file__).with_name("embeddings.json")
top_k = 3
alpha = 0.6            # weight on the semantic signal; 1.0 = cosine only
no_context = False
retrieval_only = False
question_parts: list[str] = []
args = sys.argv[1:]
i = 0
while i < len(args):
    a = args[i]
    if a == "--no-context":
        no_context = True
    elif a == "--retrieval-only":
        retrieval_only = True
    elif a == "--top-k":
        i += 1
        top_k = int(args[i])
    elif a == "--alpha":
        i += 1
        alpha = float(args[i])
    elif a == "--model":
        # Swap models without editing the file, for demoing the contrast live.
        i += 1
        local_model = args[i]
    else:
        question_parts.append(a)
    i += 1
question = " ".join(question_parts) or "Can I have a campfire at Sperry Chalet in September?"

ollama = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")

# Generation: Azure OpenAI if configured, local llama3.2 otherwise. This is the
# one-line client swap from step 9 of the demo; everything downstream is identical.
endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
key = os.environ.get("AZURE_OPENAI_KEY")
deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
if endpoint and key and deployment:
    chat_client, chat_model = AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment
    print(f"[generation: Azure OpenAI, deployment '{deployment}']")
else:
    chat_client, chat_model = ollama, local_model
    print(f"[generation: AZURE_OPENAI_* not set, falling back to local {local_model}]")


def generate(prompt: str) -> str:
    response = chat_client.chat.completions.create(model=chat_model, messages=[{"role": "user", "content": prompt}])
    return response.choices[0].message.content or ""


print(f"Q: {question}\n")

if no_context:
    # Step 1 of the demo: no retrieval, no context, pure model memory.
    print(generate(question))
    raise SystemExit

# Load the pre-chunked park docs.
chunks = [json.loads(line) for line in chunks_path.read_text().splitlines() if line.strip()]


def embed(texts: list[str]) -> list[list[float]]:
    return [d.embedding for d in ollama.embeddings.create(model="nomic-embed-text", input=texts).data]


# Embed the chunks with local nomic-embed-text, caching to embeddings.json so
# only the first run pays the ~40 seconds of embedding time.
if cache_path.exists():
    index = json.loads(cache_path.read_text())
else:
    print(f"[embedding {len(chunks)} chunks, one-time; caching to {cache_path.name}]")
    index = {}
    for start in range(0, len(chunks), 32):
        batch = chunks[start:start + 32]
        for chunk, vector in zip(batch, embed([c["text"] for c in batch])):
            index[chunk["chunk_id"]] = vector
    cache_path.write_text(json.dumps(index))


def cosine_similarity(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    return dot / (math.sqrt(sum(x * x for x in a)) * math.sqrt(sum(y * y for y in b)))


# Question filler. Without this, "can" and "have" carry as much weight as "Sperry"
# simply because no park document says "can I have".
STOP_WORDS = {
    "the", "and", "for", "are", "but", "not", "you", "your", "with", "that", "this", "these",
    "those", "from", "have", "has", "had", "was", "were", "been", "being", "can", "could",
    "will", "would", "shall", "should", "may", "might", "must", "does", "did", "doing",
    "what", "when", "where", "which", "who", "whom", "why", "how", "any", "all", "some",
    "there", "here", "then", "than", "them", "they", "their", "its", "his", "her", "our",
    "get", "got", "still", "now", "right", "just", "about", "into", "onto", "over", "under",
    "out", "off", "per", "via", "one", "two", "also", "more", "most", "much", "many", "each",
    "other", "such", "only", "own", "same", "too", "very", "let", "need", "want",
}


# Lowercase, split on non-alphanumerics, drop question filler, and knock the
# plural off so "campfires" in a document matches "campfire" in a question.
# Crude on purpose: an attendee can read all of it in ten seconds.
def tokenize(text: str) -> list[str]:
    out = []
    for t in re.findall(r"[a-z0-9]+", text.lower()):
        if len(t) > 2 and t not in STOP_WORDS:
            out.append(t[:-1] if len(t) > 3 and t.endswith("s") and not t.endswith("ss") else t)
    return out


def min_max(raw: dict[str, float]) -> dict[str, float]:
    lo, hi = min(raw.values()), max(raw.values())
    span = hi - lo
    return {k: (v - lo) / span if span > 1e-9 else 0.0 for k, v in raw.items()}


# ---------------------------------------------------------------------------
# Signal 1: semantic. Feature 04's cosine search, verbatim, over the chunk index.
# ---------------------------------------------------------------------------
question_vector = embed([question])[0]
cosine = {c["chunk_id"]: cosine_similarity(question_vector, index[c["chunk_id"]]) for c in chunks}

# ---------------------------------------------------------------------------
# Signal 2: lexical. BM25-lite over the chunk text. The embedder collapses
# "campfire regulations" from five different parks onto nearly the same point;
# the word "Sperry" appears in exactly one document, and IDF makes that count.
# ---------------------------------------------------------------------------
tokenized = {c["chunk_id"]: tokenize(c["text"]) for c in chunks}
avg_length = sum(len(t) for t in tokenized.values()) / len(tokenized)
doc_freq: Counter[str] = Counter()
for terms in tokenized.values():
    doc_freq.update(set(terms))

K1, B = 1.2, 0.3
n = len(chunks)
query_terms = list(dict.fromkeys(tokenize(question)))
# IDF: a term in 1 of 250 chunks is worth far more than one in 200 of them.
idf = {t: math.log(1 + (n - doc_freq[t] + 0.5) / (doc_freq[t] + 0.5)) for t in query_terms}

lexical: dict[str, float] = {}
for c in chunks:
    terms = tokenized[c["chunk_id"]]
    counts = Counter(terms)
    score = 0.0
    for t in query_terms:
        tf = counts.get(t)
        if not tf:
            continue
        score += idf[t] * (tf * (K1 + 1)) / (tf + K1 * (1 - B + B * len(terms) / avg_length))
    lexical[c["chunk_id"]] = score

# Rescale both signals to 0..1 across the corpus for this question, so alpha
# means what it looks like it means and the two numbers are comparable on screen.
semantic_norm = min_max(cosine)
lexical_norm = min_max(lexical)

scored = sorted(
    (
        {
            "chunk": c,
            "cosine": cosine[c["chunk_id"]],
            "semantic_norm": semantic_norm[c["chunk_id"]],
            "lexical": lexical[c["chunk_id"]],
            "lexical_norm": lexical_norm[c["chunk_id"]],
            "combined": alpha * semantic_norm[c["chunk_id"]] + (1 - alpha) * lexical_norm[c["chunk_id"]],
        }
        for c in chunks
    ),
    key=lambda h: -h["combined"],
)
top = scored[:top_k]

distinctive = "  ".join(f"{t}({idf[t]:.2f}/{doc_freq[t]} chunks)" for t in sorted(query_terms, key=lambda t: -idf[t])[:5])
print(f"[query terms by IDF]  {distinctive}")
print(f"[retrieved top {top_k}]  combined = {alpha:.2f} * semantic + {1 - alpha:.2f} * lexical")
print("  rank  combined   semantic (cos)     lexical (bm25)     chunk_id")
for r, h in enumerate(top, 1):
    print(f"  {r:>4}  {h['combined']:8.4f}   {h['semantic_norm']:5.3f} ({h['cosine']:.4f})   {h['lexical_norm']:5.3f} ({h['lexical']:5.2f})   {h['chunk']['chunk_id']}")
margin = scored[0]["combined"] - scored[1]["combined"] if len(scored) > 1 else 0
print(f"  margin over rank 2: {margin:.4f}\n")

if retrieval_only:
    raise SystemExit

# Grounded prompt: context in, citations out, refusal when the context is silent.
#
# The corpus is full of dated notices ("CLOSED effective June 20, 2026, until further
# notice"). A model with no idea what day it is treats "is the trail open right now?" as
# a question its documents cannot speak to, and refuses. So we tell it the date.
#
# The date belongs in the refusal clause specifically, not in the Rules block above.
# Stated broadly, it invites the model to apply effective-date reasoning to rules that
# have no expiry, and it starts deciding a year-round fire ban has lapsed.
#
# Production passes date.today() here; this demo pins a date so the recorded outputs in
# ../../expected-output.md stay reproducible.
TODAY = "September 23, 2026"
REFUSAL = "The provided documents don't say."

retrieved_ids = {h["chunk"]["chunk_id"] for h in top}
context = "\n\n".join(f"chunk_id: {h['chunk']['chunk_id']}\nsource: {h['chunk']['source']}\n{h['chunk']['text']}" for h in top)

prompt = f"""You are a park information assistant. Answer the visitor's question using ONLY the context below.
Rules:
- Base every statement on the context. Do not use outside knowledge.
- Cite the chunk_id of each chunk you relied on, in square brackets, e.g. [glacier-visitor-faq:02].
- Copy chunk_ids exactly as they appear above the context. Do not add section numbers to them,
  and do not combine parts of two chunk_ids.
- If, and only if, none of the context is relevant to the question, reply exactly: "{REFUSAL}"
  A question about "right now" is answered from the context, not refused: today is {TODAY},
  and a notice that is in effect "until further notice" is still in effect right now.

Context:
{context}

Question: {question}

Answer:
"""


# Any bracketed token containing a colon is a citation attempt, including the
# comma-separated lists llama3.2 sometimes writes. Being generous about what
# counts as an attempt is the point: we want to catch the near-misses.
def citations(text: str) -> list[str]:
    out = []
    for m in re.finditer(r"\[([^\]]*:[^\]]*)\]", text):
        out.extend(c.strip() for c in m.group(1).split(",") if ":" in c)
    return out


def invalid_citations(text: str, valid: set[str]) -> list[str]:
    return list(dict.fromkeys(c for c in citations(text) if c not in valid))


answer = generate(prompt)
bad = invalid_citations(answer, retrieved_ids)

# Citation validation. An invalid citation is a product defect, not a style
# issue: it is a receipt pointing at a document nobody retrieved, and sometimes
# at a document that does not exist. Retry once with the valid ids spelled out,
# then strip whatever is still wrong so a bad receipt never reaches the visitor.
if bad and REFUSAL in answer:
    # A refusal that cites a source is already correct; by definition it has no sources,
    # so the only defect is the citation itself. Do not send this back to the model. Asked
    # to rewrite using valid ids, it rewrites the refusal too, and the exact string the
    # product matches on comes back paraphrased. Deleting a citation is a string operation.
    print(f"!! CITATION CHECK FAILED: {', '.join(f'[{c}]' for c in bad)} not in the retrieved set")
    print("!! the answer was a refusal with a citation attached; dropping the citation, no retry needed\n")
    answer = REFUSAL
    bad = []
elif bad:
    print(f"!! CITATION CHECK FAILED: {', '.join(f'[{c}]' for c in bad)} not in the retrieved set")
    print("!! retrying once with the valid chunk_ids spelled out\n")
    retry_prompt = prompt + f"""

Your previous answer cited {', '.join(f'[{c}]' for c in bad)}, which is not a real chunk_id.
The only chunk_ids you may cite are, exactly:
{chr(10).join('  ' + cid for cid in retrieved_ids)}
Rewrite the answer using only those.
"""
    answer = generate(retry_prompt)
    bad = invalid_citations(answer, retrieved_ids)
    if bad:
        print(f"!! STILL INVALID after retry: {', '.join(f'[{c}]' for c in bad)}")
        print("!! stripping them; the answer below is unverified where the citation was removed\n")
        for c in bad:
            answer = answer.replace(c, "invalid-citation-removed")

print(answer)

cited = list(dict.fromkeys(c for c in citations(answer) if c in retrieved_ids))
print(f"\n[citations: {len(cited)} valid ({', '.join(cited)}), {len(bad)} invalid]")
