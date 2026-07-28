# Lab assets for 02 Extraction

Everything the lab spec in [../README.md](../README.md) references:

- `ollama.http`: two ready-to-run requests against Ollama (`llama3.2`), both carrying the same JSON schema in the `format` field: the fact-rich report and the fact-sparse one. Report text is inlined, so each request runs as-is or ports to your language by copying the JSON body.
- `tr-0007.md`: the fact-rich report (a note-taker hikes Sperry Chalet on July 4, 2026; 12.8 miles, 3,400 feet, bears, goats, and pie, all stated outright)
- `tr-0011.md`: the fact-sparse report (a Yosemite trip written up a month late; no trail name, no date, no mileage, no elevation, because the little book is in the car at the mechanic's)
- `expected-output.md`: real `llama3.2` outputs for both requests and for the .NET demo, plus the success checks and an honest list of where the model still slips
