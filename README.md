# Building Pragmatic AI: 10 AI Features Your Users Actually Want

A full-day, hands-on masterclass at WeAreDevelopers World Congress North America, September 23, 2026.

Most AI conversations start with the technology: "we should add a chatbot," "we need an LLM strategy." This workshop starts from the other end, with ten problems your users actually have and the AI feature that solves each one. By the end of the day you will have built working versions of all ten, and you'll leave with a decision framework for the question that matters back at the office: which of these problems do my users have, and what would it cost to solve them?

The organizing idea is problems, not features. Every module opens with a user who is stuck, and the AI only shows up as the answer to that user's problem.

## The day

| Time | Segment |
|---|---|
| 9:00-9:30 | **Opening.** The "problems, not features" framing, an environment check, and how the day works |
| 9:30-11:00 | **Block 1: UNDERSTANDING** (making sense of messy content) |
| | [01 Summarization](modules/01-summarization/) · [02 Extraction](modules/02-extraction/) · [03 Sentiment](modules/03-sentiment/) |
| 11:00-11:15 | Break |
| 11:15-12:45 | **Block 2: FINDING** (surfacing the right thing) |
| | [04 Semantic Search](modules/04-semantic-search/) · [05 RAG](modules/05-rag/) · [06 Recommendations](modules/06-recommendations/) |
| 12:45-13:45 | Lunch |
| 13:45-15:15 | **Block 3: DECIDING** (triage and judgment) |
| | [07 Classification & Routing](modules/07-classification-routing/) · [08 Anomaly Detection](modules/08-anomaly-detection/) · [09 Human-in-the-Loop](modules/09-human-in-the-loop/) |
| 15:15-15:30 | Break |
| 15:30-16:30 | **Block 4: DOING** (capstone) |
| | [10 Agentic Workflows](modules/10-agentic-workflows/), a double-length capstone |
| 16:30-17:00 | **Closing.** The [decision framework](docs/decision-framework.md), pitching AI features to leadership, and Q&A |

Every 30-minute module follows the same beat: 2 minutes on the user problem, a 13-minute live demo, a 13-minute timed lab with a stretch goal for fast finishers, and 2 minutes on the leadership angle. The modules are standalone, so Blocks 1 and 2 can swap if we decide to open the day with search's wow factor instead of an easy first win.

## How to use this repo

All ten modules work through one fictional product: Trailhead Guides, a national-park trip-planning app with a messy, realistic corpus of trip reports, gear reviews, trail descriptions, park regulations, and visitor inquiries (see [`data/`](data/)). Each module stands alone, but they all live in the same park.

During the workshop, open the module's `README.md`, read the lab spec, then work from the module's `lab/` folder. Every lab ships raw HTTP request files (`.http` and curl) against Ollama and Azure OpenAI, so if you can make an HTTP request in your language, you can do every lab. Python, JavaScript, Java, Go, and Rust are all welcome here. If you'd rather follow along in .NET, each module's `dotnet/` folder has a `starter/` project (the demo's starting point) and a `complete/` project (the finished demo).

Before the workshop, do the pre-work in [`SETUP.md`](SETUP.md). It's mostly "install Ollama and pull three models."

```
├── SETUP.md                  # attendee pre-work, please do this before Sept 23
├── docs/
│   ├── decision-framework.md # the closing leadership framework
│   └── slides/               # slide decks
├── data/                     # the shared Trailhead Guides corpus
└── modules/
    └── NN-feature/
        ├── README.md         # problem, concept, demo outline, lab spec, leadership beat
        ├── lab/              # language-agnostic lab assets (.http files, data, expected output)
        └── dotnet/           # starter/ and complete/ demo projects
```

## Model strategy

The workshop leans local (Ollama) wherever a small model does the job well, and reaches for Azure OpenAI / AI Foundry where quality genuinely matters. That choice pays off twice. It shows you where small, free, private models are enough, which is a finding your leadership will care about on its own. And it means most of the day survives conference wifi. Cloud modules use pre-provisioned API keys handed out at the door.

| Module | Runs on | Why |
|---|---|---|
| 01 Summarization | Ollama (`llama3.2`) | Local models summarize well |
| 02 Extraction | Ollama (`llama3.2`, JSON mode) | Structured output works locally |
| 03 Sentiment | Ollama (`phi3`) vs. Azure OpenAI | The point of this module is comparing small and specialized against big and general, live |
| 04 Semantic Search | Ollama embeddings (`nomic-embed-text`) | Fully local |
| 05 RAG | Local embeddings + Azure OpenAI generation | Local first, then cloud for the quality contrast |
| 06 Recommendations | Ollama embeddings (`nomic-embed-text`) | Reuses module 04's infrastructure to show how far embeddings go |
| 07 Classification & Routing | Ollama (`llama3.2`) | Local is fine |
| 08 Anomaly Detection | Ollama embeddings + plain math | Outlier distance in embedding space barely needs a model |
| 09 Human-in-the-Loop | Whatever the demo needs | It's a pattern, not a model feature |
| 10 Agentic Workflows | Azure OpenAI / AI Foundry | Tool-calling reliability matters, and the capstone doesn't get to be flaky |

The .NET demos use [Microsoft.Extensions.AI](https://learn.microsoft.com/en-us/dotnet/ai/microsoft-extensions-ai) as the abstraction layer, so the same C# code targets Ollama or Azure OpenAI by swapping the client registration. That one-line swap is the whole provider-flexibility story, and it gets a slide of its own.

## About the instructor

J. Tower is the owner and a principal consultant at [Trailhead Technology Partners](https://trailheadtechnology.com), a Microsoft MVP, and a .NET Foundation board member. He demos in .NET/C#, but the labs are for everyone.
