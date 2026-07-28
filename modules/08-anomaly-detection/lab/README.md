# Lab assets for 08 Anomaly Detection

Everything the lab spec in [../README.md](../README.md) references:

- `reports-0117.jsonl`: all 40 condition reports for trail-0117, straight out of `data/condition-reports.jsonl`. Five of them (`cr-0429`, `cr-0431`, `cr-0436`, `cr-0438`, `cr-0443`) describe a footbridge washing out between 2026-06-18 and 2026-06-24, and three July follow-ups (`cr-0464`, `cr-0480`, `cr-0496`) say it is still out. The other 32 are mud, ice, bugs, and blowdown.
- `reports-0042.jsonl`: the 25 reports for trail-0042, for the stretch goal. The bear cluster runs `cr-0446` through `cr-0455`, late June into early July 2026.
- `embeddings-0117.json`: real `nomic-embed-text` vectors for all 40 trail-0117 reports, 768 dimensions each, exactly as Ollama returned them. Use these if you would rather skip the embedding step and get straight to the arithmetic. They are unnormalized, so L2-normalize before you use them.
- `ollama.http`: embedding requests against `nomic-embed-text`, single and batched, plus one deliberately unprefixed request for comparison.
- `expected-output.md`: a real run's full distance ranking, the threshold, the alert output, and an honest account of how well this actually works on this data.

## One thing that will cost you an hour if you miss it

Every input in `ollama.http` and every vector in `embeddings-0117.json` was
produced from `"classification: " + report.text`, not from the report text alone.
`nomic-embed-text` is trained with task prefixes, and omitting one does not fail
loudly, it just quietly hands you worse vectors. On this data that drops the first
washout report from rank 2 to rank 11 and makes the detector fire on October mud
instead. `expected-output.md` shows both rankings side by side.
