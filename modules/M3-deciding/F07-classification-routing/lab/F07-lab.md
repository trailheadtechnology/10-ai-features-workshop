# Lab assets for 07 Classification & Routing

Everything the lab spec in [../F07-spec.md](../F07-spec.md) references:

- `inquiries-slice.jsonl`: 20 messages pulled from `data/inquiries.jsonl`, one JSON object per line with the original `id`, `channel`, `received`, and `text`. The mix is representative of the real inbox: permit requests, trail-condition questions, complaints, lost-and-found reports, a couple of general questions, both emergencies (`inq-0013`, an overdue hiker, and `inq-0041`, an injured ankle mid-trail), and one deliberately ambiguous message (`inq-0035`).
- `reference-labels.json`: the category for each id from the taxonomy the feature uses (`permit | conditions | complaint | lost-and-found | emergency | general | unsure`), the queue each category routes to, and notes on the two emergencies and on why `inq-0035` is labeled `unsure`.
- `ollama.http`: the classify request against Ollama (`llama3.2`), shown on three inlined messages: a routine conditions question, an emergency, and the ambiguous one. The taxonomy is plain-language category descriptions in the prompt, and the `format` field is a JSON schema whose enum is the seven categories, so the model cannot invent a label. Copy request 1 and swap the message text to work through the rest of the slice.
- `expected-output.md`: a real `llama3.2` run over all 20, scored against the reference labels, with the accuracy it actually got, the emergency recall, what it did with the ambiguous message, and the success checks.
- `answer-key.md`: instructor notes on the full 100-message corpus. Not for handout.
