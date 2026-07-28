# Lab assets for 01 Summarization

Everything the lab spec in [../F01-spec.md](../F01-spec.md) references:

- `ollama.http`: three ready-to-run requests against Ollama (`llama3.2`): the naive prompt, the improved prompt on the same report, and the improved prompt on the buried-hazard report. Report text is inlined, so each request runs as-is or ports to your language by copying the JSON body.
- `tr-0001.md`: the clean report (a gear obsessive hikes Avalanche Lake in July 2025; the bridge is fine, the mud and crowds are real)
- `tr-0004.md`: the buried-hazard report (June 2026; the washed-out footbridge hides mid-report between airport sandwiches and huckleberry ice cream)
- `expected-output.md`: real `llama3.2` outputs for all three requests, plus the success checks
