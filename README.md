# Building Pragmatic AI: 10 AI Features Your Users Actually Want

A full-day, hands-on masterclass at WeAreDevelopers World Congress North America, September 23, 2026.

Most AI conversations start with the technology: "we should add a chatbot," "we need an LLM strategy." This workshop starts from the other end, with ten problems your users actually have and the AI feature that solves each one. By the end of the day you will have built working versions of all ten, and you'll leave with a decision framework for the question that matters back at the office: which of these problems do my users have, and what would it cost to solve them?

The organizing idea is problems, not features. Every feature opens with a user who is stuck, and the AI only shows up as the answer to that user's problem.

## The day

| Time | Segment |
|---|---|
| 9:00-9:30 | **[Module 0: Opening](modules/M0-opening/)** · environment check, the day's thesis, a tour of the data |
| 9:30-11:00 | **[Module 1: Understanding](modules/M1-understanding/)** · making sense of messy content |
| | Core [01 Summarization](modules/M1-understanding/F01-summarization/) · [02 Extraction](modules/M1-understanding/F02-extraction/) · [03 Sentiment](modules/M1-understanding/F03-sentiment/) |
| 11:00-11:15 | Break |
| 11:15-12:45 | **[Module 2: Finding](modules/M2-finding/)** · surfacing the right thing |
| | Core [04 Semantic Search](modules/M2-finding/F04-semantic-search/) · [05 RAG](modules/M2-finding/F05-rag/) · [06 Recommendations](modules/M2-finding/F06-recommendations/) |
| 12:45-13:45 | Lunch |
| 13:45-15:15 | **[Module 3: Deciding](modules/M3-deciding/)** · triage and judgment |
| | Core [07 Classification & Routing](modules/M3-deciding/F07-classification-routing/) · [08 Anomaly Detection](modules/M3-deciding/F08-anomaly-detection/) · [09 Human-in-the-Loop](modules/M3-deciding/F09-human-in-the-loop/) |
| 15:15-15:30 | Break |
| 15:30-16:30 | **[Module 4: Doing](modules/M4-doing/)** · the capstone |
| | [10 Agentic Workflows](modules/M4-doing/F10-agentic-workflows/), everyone attempts it |
| 16:30-17:00 | **Closing** · the [decision framework](docs/decision-framework.md), pitching AI features to leadership, and Q&A |

**Ten features, five modules, and you are not expected to build all ten.** Each module presents three related features with live demos, then hands you roughly 45 minutes. One feature per module is marked **Core** and everyone does that lab; the rest are **Challenge** labs for anyone with time left, and every lab has a stretch goal beyond that. Finishing one lab properly is the intended outcome for most people. You will see all ten features demonstrated either way.

Each module's README is the menu: what the three features are, which is Core, what each lab costs you, and the thread that ties them together.

## How to use this repo

All ten features work through one fictional product: Trailhead Guides, a national-park trip-planning app with a messy, realistic corpus of trip reports, gear reviews, trail descriptions, park regulations, and visitor inquiries (see [`data/`](data/)). Each feature stands alone, but they all live in the same park.

During the workshop, open the module README to pick a feature, then open that feature's `README.md`, read the lab spec, and work from its `lab/` folder. Every lab ships raw HTTP request files (`.http` and curl) against Ollama and Azure OpenAI, so if you can make an HTTP request in your language, you can do every lab. Python, JavaScript, Java, Go, and Rust are all welcome here. If you'd rather follow along in .NET, each feature's `dotnet/` folder has a `starter/` project (the demo's starting point) and a `complete/` project (the finished demo).

Before the workshop, do the pre-work in [`SETUP.md`](SETUP.md). It's mostly "install Ollama and pull three models."

```
├── SETUP.md                  # attendee pre-work, please do this before Sept 23
├── docs/
│   ├── decision-framework.md # the closing leadership framework
│   └── slides/               # slide decks
├── data/                     # the shared Trailhead Guides corpus
└── modules/
    └── MN-theme/             # a module: three related features, one hands-on period
        ├── README.md         # the menu: which feature is Core, what each lab costs
        └── FNN-feature/      # one AI feature
            ├── README.md     # problem, concept, demo outline, lab spec, leadership beat
            ├── lab/          # language-agnostic lab assets (.http files, data, expected output)
            └── dotnet/       # starter/ and complete/ demo projects
```

## Model strategy

The workshop leans local (Ollama) wherever a small model does the job well, and reaches for Azure OpenAI / AI Foundry where quality genuinely matters. That choice pays off twice. It shows you where small, free, private models are enough, which is a finding your leadership will care about on its own. And it means most of the day survives conference wifi. Cloud features use pre-provisioned API keys handed out at the door.

| Feature | Runs on | Why |
|---|---|---|
| 01 Summarization | Ollama (`llama3.2`) | Local models summarize well |
| 02 Extraction | Ollama (`llama3.2`, JSON mode) | Structured output works locally |
| 03 Sentiment | Ollama (`phi3`) vs. Azure OpenAI | The point of this feature is comparing small and specialized against big and general, live |
| 04 Semantic Search | Ollama embeddings (`nomic-embed-text`) | Fully local |
| 05 RAG | Local embeddings + Azure OpenAI generation | Local first, then cloud for the quality contrast |
| 06 Recommendations | Ollama embeddings (`nomic-embed-text`) | Reuses feature 04's infrastructure to show how far embeddings go |
| 07 Classification & Routing | Ollama (`llama3.2`) | Local is fine |
| 08 Anomaly Detection | Ollama embeddings + plain math | Outlier distance in embedding space barely needs a model |
| 09 Human-in-the-Loop | Whatever the demo needs | It's a pattern, not a model feature |
| 10 Agentic Workflows | Azure OpenAI / AI Foundry | Tool-calling reliability matters, and the capstone doesn't get to be flaky |

The .NET demos use [Microsoft.Extensions.AI](https://learn.microsoft.com/en-us/dotnet/ai/microsoft-extensions-ai) as the abstraction layer, so the same C# code targets Ollama or Azure OpenAI by swapping the client registration. That one-line swap is the whole provider-flexibility story, and it gets a slide of its own.

## About the instructor

J. Tower is the owner and a principal consultant at [Trailhead Technology Partners](https://trailheadtechnology.com), a Microsoft MVP, and a .NET Foundation board member. He demos in .NET/C#, but the labs are for everyone.
