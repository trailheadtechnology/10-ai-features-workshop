# Lab 04: Semantic Search

*This is the Recommended lab for [Module 2](../M2-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** rank trails by similarity to a natural-language query, and beat keyword search on three queries it fails.
- **Input:** `data/trails-slice.json`, 30 trails with three keyword traps; `data/queries.json`, the three test queries and their checks.
- **How:** two embedding calls to `nomic-embed-text` through your track's embeddings client, one for the catalog and one for the query. The HTTP track sends them as `http/ollama.http` (requests 1 to 3). Cosine similarity and ranking are code you write.
- **Model:** `nomic-embed-text`, local. No key. 768 floats per input.

The big idea: turn every trail description into a list of numbers (a vector), turn the search query into the same kind of list, and rank trails by how close the two lists are. Each step below is one thing to make the program do. Each code track's `starter/` is a plain keyword search with no AI in it. Edit it until it does all four steps. Compare against `complete/` when you get stuck. If you are on the HTTP track, run the numbered requests in `http/ollama.http` where a step names one, and write the ranking in any language you like.

### Step 0: Run the starter and watch keyword search fail

Run `starter/` as it is. It lowercases the query, keeps the words of three letters or more, and counts how many of those words appear in each trail's name and description. Run it twice, once with each of these queries. If you are on the HTTP track, there is no starter; read the keyword blocks in `expected-output.md` instead.

```text
dog-friendly waterfall hike, not too steep
```

```text
somewhere quiet to take my kids
```

**Check:** the first query puts `trail-0007` Upper Yosemite Falls Trail first, a hard 7.2-mile climb that matched on `waterfall` and `too`. The second query returns exactly one trail, `trail-0004` Ocean Path, because its description happens to contain the word `kids`. The keyword blocks in `expected-output.md` show both. The rest of the lab replaces that scoring with embeddings.

### Step 1: Load the trails

1. Open `../../data/trails-slice.json`. It is one JSON array of 30 objects. Each object has `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (an array of strings), and `description`, a prose blurb of three or four sentences. The 30 were picked by hand, verbatim and in catalog order, from the 200-trail `trails.json` that feature 10's agent searches (`../../M4-doing/F10-agentic-workflows/data/trails.json`), a synthetic catalog with real park names and invented trails; no script built the slice, it ships with the workshop as is. The picks are deliberate: four dog-friendly waterfall trails with gentle grades whose descriptions never say so in those words (`trail-0011`, `trail-0027`, `trail-0055`, `trail-0068`), plus three keyword traps that make word matching fail on purpose: `trail-0074` Easy Creek Trail is a hard 2,610-foot climb named after homesteader Elias Easy, `trail-0187` Dog Lake Trail prohibits pets, and `trail-0058` Panorama Cliffs Bypass contains the word "steep" while describing how it avoids the steep sections.
2. Parse the file into a list of trail objects. The starter already does this, so keep that code.

**Check:** the list has 30 entries. The first `id` is `trail-0003`, Trail of the Cedars.

### Step 2: Embed the 30 descriptions once and cache the vectors

1. Collect the 30 `description` strings from the list, in list order.
2. Embed all 30 strings in one call to `nomic-embed-text` through your track's embeddings client, passing the 30 strings as a single batch rather than looping. Your track's walkthrough shows the call. For HTTP-track readers, `http/ollama.http` holds three ready-to-send requests against the local Ollama embeddings endpoint, the only network call this lab makes; each response carries one vector per input string, in the order you sent them. Request 2 is this batch call with two of the slice's descriptions (Carlon Falls and Easy Creek) as its input. Request 1 is the one-string version, and its input is:

```text
a gentle grade shaded by cedars
```

3. The response holds one vector per input, in the same order you sent them. Each vector is 768 floats. Walk the list and the response together, and store each vector in a dictionary keyed by that trail's `id`.
4. Write the dictionary to `embeddings.json` next to your program. At the top of the program, if that file exists, load it and skip the embed call. Print how many vectors you embedded or loaded. That file is a cache your program creates, not a data file that ships with the lab: one JSON object mapping each trail `id` to its 768-float vector, gitignored so the first run always embeds live. It is keyed by id and nothing else, so delete it whenever a description or the model changes, or every later query is ranked against vectors for text that no longer exists.

**Check:** 30 keys, each holding 768 floats. Request 1's vector starts `[0.0611, -0.0001, -0.2007, -0.0624, -0.022, -0.0063, 0.0232, -0.0023]`. The first run embeds in under two seconds; the second run prints that it loaded 30 cached vectors. Fewer than 30 vectors, or one not 768 long, means the batch did not go through.

### Step 3: Embed query 1, write cosine similarity, print the top 5

1. Take the query from the command line arguments joined with spaces. Default to query 1 when there are none. The starter already does this. Query 1 is:

```text
dog-friendly waterfall hike, not too steep
```

2. Embed the query as a single string through the same client and the same model (HTTP track: request 3 in `http/ollama.http`). Keep the one vector that comes back. The query has to go through the same model as the descriptions. Vectors from two different models cannot be compared.
3. Write a `cosine_similarity(a, b)` function. Loop `i` from 0 to 767. Add `a[i] * b[i]` to `dot`, `a[i] * a[i]` to `magA`, and `b[i] * b[i]` to `magB`. Then return `dot / (sqrt(magA) * sqrt(magB))`. The pseudocode is in `expected-output.md`. The result is a number near 1 when the two texts mean similar things and near 0 when they do not.
4. For every trail, compute the cosine between the query vector and the vector stored under that trail's `id`.
5. Sort the trails by that score, highest first, and keep the top 5.
6. Print each of the 5 on one line: the score to four decimals, `id`, `name`, `difficulty`, `distance_mi`, and the `features` list.

**Check:** at least two of `trail-0068`, `trail-0055`, `trail-0027`, `trail-0011` in your top 3; recorded first is `trail-0068` Carlon Falls at `0.7733`. `trail-0007` Upper Yosemite Falls first means you are still counting keyword hits, the step 0 baseline.

### Step 4: Run the other two queries and read the scores

1. Run the program again with each of these as the query (or change request 3's `input`). Both come from `data/queries.json`, which holds the lab's three test queries as an array of records with `id` (`q1` to `q3`), `query`, `why` (what that query is in the set to prove), and `success_check` (the same check repeated below). It was written by hand for this lab and ships with the workshop; the program never reads it, you do.

```text
somewhere quiet to take my kids
```

```text
an easy hike to a great view
```

2. Look at the scores themselves, not just the order. Query 2's top 5 sits between 0.44 and 0.49. Query 1's top hit was 0.77. When the whole top 5 is low and close together, nothing in the catalog is a strong match, and the order is mostly noise.

**Check:** query 2 puts Taft Point (`trail-0020`) first at `0.4876` (correct, and still a cliff edge), with short family-friendly trails filling the rest of the top 5; query 3 puts `trail-0058` Panorama Cliffs Bypass at or near the top (`0.6481`). `trail-0074` Easy Creek Trail in query 3's top 5 means you are matching words, not meaning.

### Stretch goals

Pick either. Each one uses information you already have, the trail metadata or the keyword hits, instead of hoping the vector carries it. The embedding fixed recall, meaning it finds the right trails. These fix precision, meaning they drop the wrong ones.

- **Filter before you rank.** Between step 3's item 3 and item 4, drop trails whose `difficulty` is `hard`. For query 1, also drop trails whose `features` list does not contain `dog-friendly`. Then rank whatever is left. **Check:** query 1 returns only dog-friendly trails; query 3 loses Beehive Loop (`trail-0017`) and Chimney Tops (`trail-0005`) from the top 5. Either still there means the filter ran after the top 5 was taken.
- **Blend in the keyword score.** Keep the keyword-hit count from step 0 alongside the cosine score, rescale both to 0..1, and rank on a weighted sum of the two. Print both scores on each row. **Check:** every row shows a cosine score and a keyword score. A trail that appears only because of its keyword count, such as Easy Creek Trail (`trail-0074`) on query 3, means the keyword weight is too high. Feature 05 measures the same blend on a bigger corpus.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| Raw HTTP, any language | [`http/F04-http.md`](http/F04-http.md) | the requests in `http/ollama.http`, or a port of them in your language |
| .NET | [`dotnet/F04-dotnet.md`](dotnet/F04-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F04-python.md`](python/F04-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F04-typescript.md`](typescript/F04-typescript.md) | `typescript/starter/index.ts` |

Every code track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/trails-slice.json`: 30 trails lifted verbatim from feature 10's 200-trail [`trails.json`](../../M4-doing/F10-agentic-workflows/data/trails.json), chosen so the catalog has four right answers for query 1 that never use the query's words, and three keyword traps. Step 1 names them.
- `data/queries.json`: the three test queries, why each one is in the set, and the success check for each. Step 4 describes the record shape.
- `http/ollama.http`: the three embedding requests the HTTP track sends. Every other track makes the same two calls through a client library.
- `embeddings.json`: not in the folder until you run the program. Step 2 explains the cache.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
