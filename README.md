# Building Pragmatic AI: 10 AI Features Your Users Actually Want

A full-day, hands-on masterclass at WeAreDevelopers World Congress North America, September 23, 2026.

Most AI conversations start with the technology: "we should add a chatbot," "we need an LLM strategy." This workshop starts from the other end, with ten problems your users actually have and the AI feature that solves each one. By the end of the day you will have built working versions of all ten, and you'll leave with a decision framework for the question that matters back at the office: which of these problems do my users have, and what would it cost to solve them?

The organizing idea is problems, not features. Every feature opens with a user who is stuck, and the AI only shows up as the answer to that user's problem.

## The day

| Time | Segment |
|---|---|
| 9:00-9:30 | **[Module 0: Opening](modules/M0-opening/M0-overview.md)** · environment check, the day's thesis, a tour of the data |
| 9:30-11:00 | **[Module 1: Understanding](modules/M1-understanding/M1-overview.md)** · making sense of messy content |
| | Core [01 Summarization](modules/M1-understanding/F01-summarization/F01-spec.md) · [02 Extraction](modules/M1-understanding/F02-extraction/F02-spec.md) · [03 Sentiment](modules/M1-understanding/F03-sentiment/F03-spec.md) |
| 11:00-11:15 | Break |
| 11:15-12:45 | **[Module 2: Finding](modules/M2-finding/M2-overview.md)** · surfacing the right thing |
| | Core [04 Semantic Search](modules/M2-finding/F04-semantic-search/F04-spec.md) · [05 RAG](modules/M2-finding/F05-rag/F05-spec.md) · [06 Recommendations](modules/M2-finding/F06-recommendations/F06-spec.md) |
| 12:45-13:45 | Lunch |
| 13:45-15:15 | **[Module 3: Deciding](modules/M3-deciding/M3-overview.md)** · triage and judgment |
| | Core [07 Classification & Routing](modules/M3-deciding/F07-classification-routing/F07-spec.md) · [08 Anomaly Detection](modules/M3-deciding/F08-anomaly-detection/F08-spec.md) · [09 Human-in-the-Loop](modules/M3-deciding/F09-human-in-the-loop/F09-spec.md) |
| 15:15-15:30 | Break |
| 15:30-16:30 | **[Module 4: Doing](modules/M4-doing/M4-overview.md)** · the capstone |
| | [10 Agentic Workflows](modules/M4-doing/F10-agentic-workflows/F10-spec.md), everyone attempts it |
| 16:30-17:00 | **Closing** · the [decision framework](docs/decision-framework.md), pitching AI features to leadership, and Q&A |

**Ten features, five modules, and you are not expected to build all ten.** Each module presents three related features with live demos, then hands you roughly 45 minutes. One feature per module is marked **Core** and everyone does that lab; the rest are **Challenge** labs for anyone with time left, and every lab has a stretch goal beyond that. Finishing one lab properly is the intended outcome for most people. You will see all ten features demonstrated either way.

Each module's overview is the menu: what the three features are, which is Core, what each lab costs you, and the thread that ties them together.

## How to use this repo

All ten features work through one fictional product: Trailhead Guides, a national-park trip-planning app with a messy, realistic corpus of trip reports, gear reviews, trail descriptions, park regulations, and visitor inquiries (see [`data/`](data/corpus.md)). Each feature stands alone, but they all live in the same park.

During the workshop, open the module overview to pick a feature, then open that feature's spec, read the lab section, and work from its `lab/` folder. Every doc is named for what it is, so `F05-spec.md`, `F05-lab.md`, and `F05-dotnet.md` are the three files for RAG. Every lab ships raw HTTP request files (`.http` and curl) against Ollama and Azure OpenAI, so if you can make an HTTP request in your language, you can do every lab. Python, JavaScript, Java, Go, and Rust are all welcome here. If you'd rather follow along in .NET, each feature's `dotnet/` folder has a `starter/` project (the demo's starting point) and a `complete/` project (the finished demo).

Before the workshop, do the pre-work in [`SETUP.md`](SETUP.md). It's mostly "install Ollama and pull three models."

```
├── SETUP.md                  # attendee pre-work, please do this before Sept 23
├── docs/
│   ├── decision-framework.md # the closing leadership framework
│   └── slides/               # slide decks
├── data/                     # the shared Trailhead Guides corpus (corpus.md)
└── modules/
    └── MN-theme/             # a module: three related features, one hands-on period
        ├── MN-overview.md    # the menu: which feature is Core, what each lab costs
        └── FNN-feature/      # one AI feature
            ├── FNN-spec.md   # problem, concept, demo outline, lab spec, leadership beat
            ├── lab/          # FNN-lab.md plus the assets: .http files, data, expected output
            └── dotnet/       # FNN-dotnet.md plus starter/ and complete/ projects
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
