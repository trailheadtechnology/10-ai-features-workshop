"""Demo starting point: one chat client, one classify function, one review.
Run: python main.py [review-id]
Ids come from ../../data/easy.jsonl and ../../data/hard.jsonl. The default is
gr-0007, a hard-set review whose sarcasm ("five-star experience, truly") points
the opposite way from its two-star rating.
"""

import json
import sys
from pathlib import Path

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
DATA = Path(__file__).resolve().parents[2] / "data"

wanted = sys.argv[1] if len(sys.argv) > 1 else "gr-0007"
reviews = [json.loads(line) for name in ("easy.jsonl", "hard.jsonl") for line in (DATA / name).read_text().splitlines() if line.strip()]
review = next(r for r in reviews if r["id"] == wanted)

print(f"{review['product']} ({review['rating']} stars), reviewed by {review['reviewer']}")
print(review["text"])
print()


# The whole feature is this function. The prompt carries it; the model is
# swappable because everything upstream only sees the client and a model name.
def classify(client: OpenAI, model: str, text: str) -> str:
    # Keep this prompt byte-identical to the one in ../../http/ollama.http, line breaks
    # included. Reflowing these four lines into one costs phi3 measured accuracy
    # on both sets, so a comparison run against a reflowed prompt is not
    # comparing models. See ../../expected-output.md.
    prompt = f"""Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: {text}"""
    response = client.chat.completions.create(model=model, messages=[{"role": "user", "content": prompt}], temperature=0)
    raw = (response.choices[0].message.content or "").lower()
    # Small models sometimes wrap the label in a sentence; keep the first label mentioned.
    found = [(raw.index(l), l) for l in ("positive", "negative", "mixed") if l in raw]
    return min(found)[1] if found else raw.strip()


print(f"phi3 says: {classify(client, 'phi3', review['text'])}")
