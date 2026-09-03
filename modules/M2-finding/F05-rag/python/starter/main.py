"""Demo starting point: one chat client, one question, no context.
This is the "plain chatbot" from step 1 of the demo script. It answers
confidently. It is also wrong: the model puts Sperry Chalet in California
and guesses at fire rules the park wrote down years ago.
Run: uv run main.py ["your question"]
"""

import sys

from openai import OpenAI

client = OpenAI(base_url="http://localhost:11434/v1", api_key="ollama")

question = " ".join(sys.argv[1:]) or "Can I have a campfire at Sperry Chalet in September?"

print(f"Q: {question}\n")
response = client.chat.completions.create(model="llama3.2", messages=[{"role": "user", "content": question}])
print(response.choices[0].message.content)
