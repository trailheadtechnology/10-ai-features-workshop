# The Trailhead Guides corpus

One shared, deliberately messy dataset for a fictional national-park trip-planning app, reused across all ten modules. The theme gives the day a narrative without coupling the modules together: each lab stands alone, but they all live in the same park.

**This file is the dataset contract, and the corpus below it is built.** Everything described here exists and is what the labs actually read. The file names and fields are the contract, because module lab specs reference them, so if a schema changes, update the consuming modules' READMEs in the same commit.

The data is synthetic, generated for this workshop. Trailhead Guides is not a real company, the trails and reviewers are invented, and the park documents imitate the register of real regulations without reproducing any. Where a real park name appears, treat every rule attached to it as fiction: do not plan an actual trip from this corpus.

## Corpora

### `trip-reports/*.md`: long, rambling trip reports

About 40 reports, 500 to 1,500 words each, written in inconsistent first-person styles: trail-diary rambling, gear obsessing, off-topic tangents. Markdown with YAML front matter:

```yaml
---
id: tr-0031
author: display name
date: 2026-06-14
park: Glacier National Park
---
```

Trail names, dates hiked, distances, wildlife sightings, and trail conditions appear only in the prose, and inconsistently, because the summarization and extraction labs need realistic mess to work against.

Powers: 01 Summarization (the "conditions TL;DR"), 02 Extraction (prose to clean JSON).

### `gear-reviews.jsonl`: product reviews

About 300 reviews across roughly 25 products. One JSON object per line:

```json
{"id": "gr-0042", "product": "Cascade 65 Backpack", "rating": 4, "reviewer": "display name", "text": "..."}
```

Include sarcasm, mixed feelings, and reviews whose star rating contradicts the text. The sentiment lab needs hard cases.

Powers: 03 Sentiment, 06 Recommendations.

### `trails.json`: trail catalog

About 200 trails across a handful of parks, one JSON array:

```json
{"id": "trail-0117", "name": "Avalanche Lake Trail", "park": "Glacier National Park",
 "distance_mi": 4.6, "elevation_ft": 730, "difficulty": "moderate",
 "features": ["waterfall", "lake", "dog-friendly"], "description": "2 to 4 sentences of prose"}
```

Descriptions must contain semantic content that keyword search misses. A trail described as "a gentle grade shaded by cedars" should match a "not too steep" query even though the words never overlap.

Powers: 04 Semantic Search, 06 Recommendations.

### `park-docs/*.md`: regulations, guides, FAQs

25 markdown documents: campfire regulations, permit rules, seasonal closures, backcountry guides, per-park FAQs. Long enough to need chunking. Parks have overlapping but different rules, so retrieval quality is visible: "Can I have a campfire at Sperry Chalet in September?" has one right answer (no, wood fires are prohibited there year round), and several documents are near misses that would produce a different answer if retrieved instead. The authoritative statement lives in `glacier-backcountry-camping-guide.md` section 4.2. Two other documents touch the same fact without owning it: `glacier-bear-safety-advisory.md` mentions the no-wood-fire status in passing while discussing food waste, and `glacier-seasonal-closures-2026.md` defers to the camping guide by document number. That is deliberate. Real corpora repeat themselves imperfectly, and a retriever that lands on a cross-reference instead of the source is a case worth seeing.

Powers: 05 RAG.

### `inquiries.jsonl`: visitor messages

About 100 inbound visitor messages with a realistic mix:

```json
{"id": "inq-0058", "channel": "email|web-form|voicemail-transcript", "received": "2026-07-02T09:14:00Z", "text": "..."}
```

Categories present in the data (unlabeled, since labeling is the lab): permit requests, trail-condition questions, complaints, lost and found, and a few genuine emergencies that make misrouting feel consequential.

Powers: 07 Classification & Routing, 09 Human-in-the-Loop.

### `condition-reports.jsonl`: trail conditions over time

About 500 short reports over two seasons across the trails in `trails.json`:

```json
{"id": "cr-0311", "trail_id": "trail-0117", "date": "2026-06-20", "text": "Muddy near the lake, otherwise clear."}
```

Mostly routine, with planted anomaly clusters: a sudden burst of "bridge washed out" reports on one trail, and a spike of bear-activity mentions on another.

Powers: 08 Anomaly Detection.

### `mock-apis/`: JSON fixtures for the capstone

Small canned responses standing in for real services, shaped like tool-call results:

- `weather.json`: forecast by park and date range
- `permits.json`: permit availability plus a fake "submit request" response
- `campsites.json`: campsite availability by park and date

Powers: 10 Agentic Workflows ("Plan me a 3-day trip in Glacier for mid-September", where the agent checks conditions, finds trails, drafts an itinerary, and requests a permit).

## Summary

| Corpus | Format | Actual size | Powers |
|---|---|---|---|
| `trip-reports/*.md` | markdown + front matter | 40 reports, 500 to 1,500 words | 01, 02 |
| `gear-reviews.jsonl` | JSONL | 300 reviews across 25 products | 03, 06 |
| `trails.json` | JSON array | 200 trails across 6 parks | 04, 06 |
| `park-docs/*.md` | markdown | 25 docs | 05 |
| `inquiries.jsonl` | JSONL | 100 messages | 07, 09 |
| `condition-reports.jsonl` | JSONL | 500 reports | 08 |
| `mock-apis/*.json` | JSON fixtures | 3 fixtures | 10 |
