# Lab assets for 02 Extraction

Everything the lab spec in [../F02-spec.md](../F02-spec.md) references:

- `ollama.http`: two ready-to-run requests against Ollama (`llama3.2`), both carrying the same JSON schema in the `format` field: the fact-rich report and the fact-sparse one. Report text is inlined, so each request runs as-is or ports to your language by copying the JSON body.
- `tr-0007.md`: the fact-rich report (a note-taker hikes Sperry Chalet on July 4, 2026; 12.8 miles, 3,400 feet, bears, goats, and pie, all stated outright)
- `tr-0011.md`: the fact-sparse report (a Yosemite trip written up a month late; no trail name, no date, no mileage, no elevation, because the little book is in the car at the mechanic's)
- `expected-output.md`: real `llama3.2` outputs for both requests and for the .NET demo, plus the success checks and an honest list of where the model still slips

The last step of the lab is the validator. The schema gets you JSON that parses; it does not get you JSON that is true. Once you have an extracted record, write the rejection rules:

- a date that no parser accepts is not a date, and "last month" belongs in `null`
- `0` miles and `0` feet are not measurements, they are a missing fact wearing a number
- an empty string is not a trail name
- a trail or park whose words never appear in the report is a name the model supplied

Reject means coerce to `null`, not throw. A gap is something a human can fill; a plausible wrong value is something nobody ever notices. The .NET version in `../dotnet/complete/Program.cs` is roughly seventy lines of ordinary code, and `expected-output.md` has real rejections from four consecutive runs so you know what you are aiming at.
