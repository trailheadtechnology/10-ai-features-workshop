# Lab assets for 09 Human-in-the-Loop

Everything the lab spec in [../README.md](../README.md) references:

- `ollama.http`: five ready-to-run requests against Ollama (`llama3.2`). Three everyday drafts (Mist Trail boilerplate, the Sperry campfire rules answer, the Avalanche Lake closure), then the emergency, then the same emergency with the escalation rule moved to the front of the prompt. Inquiry text and park-doc excerpt are inlined, so each request runs as-is or ports to your language by copying the JSON body. The system prompt is the same one both .NET projects use.
- `inquiries.jsonl`: six inquiries drawn from `data/inquiries.jsonl`, already routed by feature 07, each carrying its category and the park doc it needs. Easy boilerplate (inq-0002), a permit rules question (inq-0003), a closure with a real constraint (inq-0005), a complaint with no doc to lean on (inq-0007), an overdue hiker (inq-0013), and the Sperry campfire question (inq-0051).
- `snippets/`: the four park-doc excerpts, quoted with document and section numbers so a draft can cite its source.
- `policy-worksheet.md`: the lane table to fill in, one row per feature 07 category.
- `expected-output.md`: real `llama3.2` drafts for all six inquiries, annotated with what a ranger should approve, edit, or reject and why, plus the reference policy. It also carries the emergency result, which is the point of the lab: told plainly not to draft a reply to an overdue-hiker report, the model drafted a reassuring one three times out of three.
