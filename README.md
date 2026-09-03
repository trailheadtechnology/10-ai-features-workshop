# Building Pragmatic AI: 10 AI Features Your Users Actually Want
A full-day, hands-on workshop.

Most AI conversations start with the technology: "we should add a chatbot," "we need an LLM strategy." This workshop starts from the other end, with ten problems your users actually have and the AI feature that solves each one. By the end of the day you will have built at least four of them yourself, and seen all ten working. More importantly, you'll leave with a decision framework for when to use AI and how to identify the correct AI solution to real-world problems?

## Who It Is For, and What You Leave Able to Do

You are involved in the creation of software for a living, you can make an HTTP request in your language of choice, and you want to know what shipping AI features looks like. No machine learning background is assumed, and no particular language or stack is required: the labs run from raw HTTP files or from .NET, Python, or TypeScript starters. The instructor will demo in C# and .NET.

By the end of the day you will be able to identify a problem that is a good match for an AI solution, and:

1. Turn a long document into a purpose-shaped summary and a validated structured record using a local model, and explain why the instruction and the schema matter more than the model.
2. Build embedding search over your own catalog, and ground a model's answer in your own documents with citations you have verified and a refusal path for questions the documents don't cover.
3. Route a shared inbox with a classifier tuned so the expensive class is never missed, and put a human approval gate, with an audit trail, in front of anything irreversible.
4. Wire a handful of tools into a bounded agent loop, read its trace, and say what would have to be true before you'd let it act on a real system.

We will cover ten features grouped into 4 modules.

## The Day

| Length | Segment |
|---|---|
| 30 min | **[Module 0: Opening](modules/M0-opening/M0-overview.md)** · environment check, the day's thesis, a tour of the data |
| 90 min | **[Module 1: Understanding](modules/M1-understanding/M1-overview.md)** · making sense of messy content |
| | [01 Summarization](modules/M1-understanding/F01-summarization/F01-spec.md) [RECOMMENDED] |
| | [02 Extraction](modules/M1-understanding/F02-extraction/F02-spec.md) [CHALLENGE] |
| | [03 Sentiment](modules/M1-understanding/F03-sentiment/F03-spec.md) [CHALLENGE] |
| 15 min | Break |
| 90 min | **[Module 2: Finding](modules/M2-finding/M2-overview.md)** · surfacing the right thing |
| | [04 Semantic Search](modules/M2-finding/F04-semantic-search/F04-spec.md) [RECOMMENDED] |
| | [05 RAG](modules/M2-finding/F05-rag/F05-spec.md) [CHALLENGE] |
| | [06 Recommendations](modules/M2-finding/F06-recommendations/F06-spec.md) [CHALLENGE] |
| 60 min | Lunch |
| 90 min | **[Module 3: Deciding](modules/M3-deciding/M3-overview.md)** · triage and judgment |
| | [07 Classification & Routing](modules/M3-deciding/F07-classification-routing/F07-spec.md) [RECOMMENDED] |
| | [08 Anomaly Detection](modules/M3-deciding/F08-anomaly-detection/F08-spec.md) [CHALLENGE] |
| | [09 Human-in-the-Loop](modules/M3-deciding/F09-human-in-the-loop/F09-spec.md) [CHALLENGE] |
| 15 min | Break |
| 60 min | **[Module 4: Doing](modules/M4-doing/M4-overview.md)** · the capstone |
| | [10 Agentic Workflows](modules/M4-doing/F10-agentic-workflows/F10-spec.md) [EVERYONE] |
| 30 min | **Closing** · the [decision framework](docs/decision-framework.md), pitching AI features to leadership, and Q&A |

**Ten features, four modules, but you are not expected to build all ten.** Each 90-minute module opens with 30 minutes in which I introduce the theme and demo all three features, and then hands you 60 minutes to build. One feature per module is marked **Recommended**, and that is the lab to start with unless you have a reason not to; the other two are **Challenge** labs for anyone with time left, and every lab has a stretch goal beyond that. The capstone is the one lab everyone does. Finishing one lab properly is the intended outcome for most people. You will see all ten features demonstrated either way.

Each module's overview is the menu: what the three features are, which one is Recommended, what each lab costs you, and the thread that ties them together.

## How to Use This Repo

All ten features work through one fictional product: Trailhead Guides, a national-park trip-planning app with a messy, realistic corpus of trip reports, gear reviews, trail descriptions, park regulations, and visitor inquiries. Every feature folder carries the data its lab reads in its own `data/`, described in that feature's lab doc, so you meet each dataset when its feature does. The data is synthetic; where a real park name appears, every rule attached to it is fiction, and nobody should plan an actual trip from it.

During the workshop, open the module overview to pick a feature, then open that feature's `FNN-lab.md`: the goal, the steps, the success checks, and a table of four languages. `http/` is raw requests you can run from VS Code or port to any language; `dotnet/`, `python/`, and `typescript/` each have a `starter/` to edit and a `complete/` answer key. Every language has its own `FNN-<language>.md` walkthrough of the same steps, and every language checks against the same `expected-output.md`. The spec (`FNN-spec.md`) is the short read before the lab: the user problem, the concept, and the leadership beat. The instructor's demo scripts live with the slides, in `docs/slides/outlines/`.

Before the workshop, do the pre-work in [`SETUP.md`](SETUP.md). It's mostly "install Ollama and pull three models."

```
├── SETUP.md                  # attendee pre-work, please do this before Sept 23
├── workshop.slnx             # every .NET starter and complete project; dotnet build once
├── docs/
│   ├── decision-framework.md # the closing leadership framework
│   ├── runsheet.md           # instructor: what to have open, what to run, what to cut
│   ├── soak-2026-09-03.md    # measured results against the real Azure deployment
│   └── slides/               # decks, outlines, and the instructor demo scripts
└── modules/
    └── MN-theme/             # a module: three related features, one hands-on period
        ├── MN-overview.md    # the menu: which feature is Recommended, what each lab costs
        └── FNN-feature/      # one AI feature
            ├── FNN-spec.md   # the read-first page: problem, concept, leadership beat
            ├── FNN-lab.md    # attendee: goal, steps, success checks, stretch, what is in data/
            ├── expected-output.md
            ├── data/         # everything the lab and the code read
            ├── http/         # FNN-http.md walkthrough plus ollama.http / azure.http: raw requests, any language
            ├── dotnet/       # FNN-dotnet.md walkthrough plus starter/ and complete/ projects
            ├── python/       # FNN-python.md walkthrough, requirements.txt, starter/main.py, complete/main.py
            └── typescript/   # FNN-typescript.md walkthrough, package.json, starter/index.ts, complete/index.ts
```

## How the Models Are Called

Every model in this workshop sits behind an HTTP endpoint, and everything else is a wrapper around a POST.

- **Ollama** is a local server. Once it is running, it listens on `http://localhost:11434` and answers two families of requests: its native API (`/api/chat`, `/api/embed`), which is what the `http/ollama.http` files use, and an OpenAI-compatible API under `/v1`, which is what the Python and TypeScript starters use through the official `openai` SDK. The model runs on your laptop; the request is a local HTTP call.
- **Microsoft Foundry** serves the workshop's `gpt-4.1` and `gpt-5.5` deployments the same way, at `https://trailhead-ai-workshop.openai.azure.com/openai/deployments/<name>/chat/completions` with an `api-key` header. That is the Azure OpenAI API, and it is the same request shape as Ollama's `/v1`, which is why the SDKs need only a different constructor to switch.

So the `http/` track is not a simplification of the "real" way; it is the real way, with the SDK removed. Microsoft.Extensions.AI in .NET and the `openai` package in Python and TypeScript are conveniences over the same calls, and swapping a feature from local to cloud is a URL and a key.

## Model Strategy

The workshop leans local (using Ollama) wherever a small model does the job well, and reaches for frontier models on Microsoft Foundry where model size genuinely matters. That choice pays off twice in the labs and shows where small, free, private models are enough and where they are not. Cloud features use pre-provisioned API keys handed out during the workshop, pointing at a `gpt-4.1` deployment in Azure AI Foundry.

Two names show up for the cloud side, on purpose. **Microsoft Foundry** is the platform: the resource, the model catalog, and the deployments the workshop runs. **Azure OpenAI** is the API those OpenAI models are served through, which is why the endpoint is `…openai.azure.com`, the SDK class is `AzureOpenAI`, and the env vars are `AZURE_OPENAI_*`. The labs talk to the API; the instructor manages the platform.

| Feature | Runs on | Why |
|---|---|---|
| 01 Summarization | Ollama (`llama3.2`) | Local models summarize well |
| 02 Extraction | Ollama (`llama3.2`, JSON mode) | Structured output works locally |
| 03 Sentiment | Ollama (`phi3`) vs. `gpt-4.1` on Foundry | The point of this feature is comparing small and specialized against big and general, live |
| 04 Semantic Search | Ollama embeddings (`nomic-embed-text`) | Fully local |
| 05 RAG | Local embeddings + `gpt-4.1` on Foundry for generation | Local first, then cloud for the quality contrast |
| 06 Recommendations | Ollama embeddings (`nomic-embed-text`) | Reuses feature 04's infrastructure to show how far embeddings go |
| 07 Classification & Routing | Ollama (`llama3.2`) | Local is fine |
| 08 Anomaly Detection | Ollama embeddings + plain math | Outlier distance in embedding space barely needs a model |
| 09 Human-in-the-Loop | Whatever the demo needs | It's a pattern, not a model feature |
| 10 Agentic Workflows | `gpt-5.5` on Foundry | Tool-calling reliability matters, and the capstone doesn't get to be flaky |

The .NET demos use [Microsoft.Extensions.AI](https://learn.microsoft.com/en-us/dotnet/ai/microsoft-extensions-ai) as the abstraction layer, so the same C# code targets Ollama or Azure OpenAI by swapping the client registration. That one-line swap gets a slide of its own.

## About the Instructor

J. Tower is the owner and a principal consultant at [Trailhead Technology Partners](https://trailheadtechnology.com), a Microsoft MVP, and a .NET Foundation board member. He demos in .NET/C#, but the labs are for everyone.
