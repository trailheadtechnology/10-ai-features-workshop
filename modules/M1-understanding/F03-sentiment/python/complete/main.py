"""Finished demo, matching the demo script in docs/slides/outlines:
  uv run main.py            both sets, both models, table + accuracy + disagreements
  uv run main.py --easy     easy set only (demo steps 3 and 4)
  uv run main.py --hard     hard set only (demo step 5)

The big model is Azure OpenAI when AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_KEY,
and AZURE_OPENAI_DEPLOYMENT are set. When they aren't, llama3.2 on Ollama
stands in so the whole comparison runs offline. Either way, the swap is the
few lines building `big` below; nothing downstream changes.
"""

import json
import os
import sys
from dataclasses import dataclass
from pathlib import Path

from openai import AzureOpenAI, OpenAI

DATA = Path(__file__).resolve().parents[2] / "data"

sets = ["easy"] if "--easy" in sys.argv else ["hard"] if "--hard" in sys.argv else ["easy", "hard"]

# Model 1: the small local model. Free, private, 2GB.
ollama = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")
small = (ollama, "phi3")

# Model 2: the big model, or its local stand-in.
endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
key = os.environ.get("AZURE_OPENAI_KEY")
deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
if endpoint and key and deployment:
    big = (AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment)
    big_name = f"azure:{deployment}"
else:
    print("AZURE_OPENAI_* not set; using llama3.2 on Ollama as the big-model stand-in.\n")
    big = (ollama, "llama3.2")
    big_name = "llama3.2"


# Same function as the starter: one prompt, one word back, any client.
def classify(target: tuple[OpenAI, str], text: str) -> str:
    client, model = target
    # Both models get this exact prompt, and it is byte-identical to the one in
    # ../../http/ollama.http and ../../http/azure.http, line breaks included. Reflowing these
    # four lines into one costs phi3 measured accuracy on both sets while leaving
    # llama3.2 unchanged, so varying the prompt shape and the model in the same
    # run measures nothing. See ../../expected-output.md.
    prompt = f"""Classify this gear review as exactly one word: positive, negative, or mixed.
Positive means the reviewer is happy with the product, negative means unhappy,
mixed means genuinely both. Judge the review text only; ignore any star rating
it mentions. Reply with only the label.

Review: {text}"""
    response = client.chat.completions.create(model=model, messages=[{"role": "user", "content": prompt}], temperature=0)
    raw = (response.choices[0].message.content or "").lower()
    found = [(raw.index(l), l) for l in ("positive", "negative", "mixed") if l in raw]
    return min(found)[1] if found else raw.strip()


@dataclass
class Result:
    review: dict
    set: str
    reference: str
    small: str
    big: str


labels = json.loads((DATA / "reference-labels.json").read_text())
results: list[Result] = []

for name in sets:
    print(f"── {name} set ──")
    print(f"{'id':<9} {'reference':<10} {'phi3':<10} {big_name:<10}")
    for line in (DATA / f"{name}.jsonl").read_text().splitlines():
        if not line.strip():
            continue
        review = json.loads(line)
        reference = labels[review["id"]]["label"]
        s = classify(small, review["text"])
        b = classify(big, review["text"])
        results.append(Result(review, name, reference, s, b))
        flag = "  <- disagree" if s != b else ""
        print(f"{review['id']:<9} {reference:<10} {s:<10} {b:<10}{flag}")
    print()

print("── accuracy vs. reference labels ──")
for name in sets:
    batch = [r for r in results if r.set == name]
    small_ok = sum(1 for r in batch if r.small == r.reference)
    big_ok = sum(1 for r in batch if r.big == r.reference)
    print(f"{name:<5}  phi3 {small_ok}/{len(batch)}   {big_name} {big_ok}/{len(batch)}")
print()

disagreements = [r for r in results if r.small != r.big]
print(f"── disagreements ({len(disagreements)} of {len(results)}) ──")
for d in disagreements:
    verdict = f"{big_name} right" if d.big == d.reference else "phi3 right" if d.small == d.reference else "both wrong"
    print(f"{d.review['id']} [{d.set}] ref={d.reference} phi3={d.small} {big_name}={d.big}  ({verdict})")
    text = d.review["text"]
    print(f'  "{text if len(text) <= 100 else text[:100].rstrip() + "..."}"')
if not disagreements:
    print("(none this run)")
