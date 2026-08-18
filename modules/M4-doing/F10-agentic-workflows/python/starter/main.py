"""The trip request goes to a plain chat completion, with no tools and no loop.
Nothing here can reach the trail catalog, the weather feed, or the condition
reports, so the itinerary that comes back is fluent, generic, and books
nothing. That gap is the point of this script; the fix lives in ../complete.

Run: python main.py [your own trip request]
"""

import os
import sys

from openai import AzureOpenAI, OpenAI


def create_chat_client() -> tuple[OpenAI, str]:
    endpoint = os.environ.get("AZURE_OPENAI_ENDPOINT")
    key = os.environ.get("AZURE_OPENAI_KEY")
    deployment = os.environ.get("AZURE_OPENAI_DEPLOYMENT")
    if endpoint and key and deployment:
        return AzureOpenAI(azure_endpoint=endpoint, api_key=key, api_version="2024-10-21"), deployment
    print("[note] AZURE_OPENAI_* not set; falling back to Ollama llama3.2.")
    return OpenAI(base_url="http://localhost:11434/v1", api_key="ollama"), "llama3.2"


client, model = create_chat_client()

request = " ".join(sys.argv[1:]) or "Plan me a 3-day trip in Glacier National Park for September 14-16."

print(f"Request: {request}")
print("-" * 60)

response = client.chat.completions.create(
    model=model,
    messages=[{"role": "user", "content": f"You are the trip planner for Trailhead Guides, a hiking app.\n\n{request}"}],
)

print(response.choices[0].message.content)
print("-" * 60)
print("Note: zero tool calls were made. No weather was checked, no")
print("conditions were read, nothing was booked. Fluent and useless.")
