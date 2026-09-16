# Lab 04: Semantic Search

*This is the Recommended lab for [Module 2](../M2-overview.md): start here unless you have a reason not to. The hands-on period runs about 60 minutes, so there is room to do it properly rather than fast.*

- **Goal:** rank trails by similarity to a natural-language query, and beat keyword search on three queries it fails.
- **Input:** `data/trails-slice.json`, 30 trails with three keyword traps; `data/queries.json`, the three test queries and their checks.
- **How:** two embedding calls to `nomic-embed-text` through your track's embeddings client, one for the catalog and one for the query. Cosine similarity and ranking are code you write.
- **Model:** `nomic-embed-text`, local. No key. 768 floats per input.

The big idea: turn every trail description into a list of numbers (a vector), turn the search query into the same kind of list, and rank trails by how close the two lists are. Every step below is one thing to make the program do. Every track's `starter/` is a plain keyword search with no AI in it. Edit it until it does all four steps. Compare against `complete/` when you get stuck.

### Step 0: Run the starter and watch keyword search fail

**Do:** run `starter/` as it is with each of these queries. It lowercases the query, keeps words of three letters or more, and counts how many appear in each trail's name and description.

```text
dog-friendly waterfall hike, not too steep
```

```text
somewhere quiet to take my kids
```

**Check:** the first query puts `trail-0007` Upper Yosemite Falls Trail first, a hard 7.2-mile climb that matched on `waterfall` and `too`. The second query returns exactly one trail, `trail-0004` Ocean Path, because its description happens to contain the word `kids`. Both are in `expected-output.md`'s keyword blocks. The rest of the lab replaces that scoring with embeddings.

### Step 1: Load the trails

**Do:**
1. Open `../../data/trails-slice.json`: one JSON array of 30 objects, each with `id`, `name`, `park`, `distance_mi`, `elevation_ft`, `difficulty`, `features` (array of strings), and `description` (a three- or four-sentence blurb).
2. Parse it into a list of trail objects (the starter already does this — keep that code).

**Why:** the 30 were hand-picked from feature 10's 200-trail catalog to make keyword matching fail on purpose: four dog-friendly waterfall trails whose descriptions never say so in those words (`trail-0011`, `trail-0027`, `trail-0055`, `trail-0068`), plus three keyword traps — `trail-0074` Easy Creek Trail is a hard 2,610-foot climb named after homesteader Elias Easy, `trail-0187` Dog Lake Trail prohibits pets, and `trail-0058` Panorama Cliffs Bypass contains the word "steep" while describing how it avoids the steep sections.

**Check:** the list has 30 entries. The first `id` is `trail-0003`, Trail of the Cedars.

### Step 2: Embed the 30 descriptions once and cache the vectors

**Do:**
1. Collect the 30 `description` strings, in list order.
2. Embed all 30 in **one call** to `nomic-embed-text` (a single batch, not a loop). The response holds one 768-float vector per input, in the order sent.
3. Walk the list and the response together and store each vector in a dictionary keyed by that trail's `id`.
4. Write the dictionary to `embeddings.json` next to your program. At program start, load it and skip the embed call if it exists. Print how many vectors were embedded or loaded.

**Why:** `embeddings.json` is a cache your program creates, not a shipped data file — gitignored so the first run always embeds live. It is keyed only by `id`, so delete it whenever a description or the model changes, or every later query ranks against vectors for text that no longer exists.

**Check:** 30 keys, each holding 768 floats. The first run embeds in under two seconds; the second run prints that it loaded 30 cached vectors. Fewer than 30 vectors, or one not 768 long, means the batch did not go through.

### Step 3: Embed query 1, write cosine similarity, print the top 5

**Do:**
1. Take the query from the command-line arguments, joined with spaces; default to this when there are none:

   ```text
   dog-friendly waterfall hike, not too steep
   ```

2. Embed the query as a single string, through the same client and same model. Keep the one vector that comes back.
3. Write `cosine_similarity(a, b)`: loop `i` 0 to 767, accumulate `dot += a[i]*b[i]`, `magA += a[i]*a[i]`, `magB += b[i]*b[i]`, then return `dot / (sqrt(magA) * sqrt(magB))`.
4. For every trail, compute the cosine between the query vector and that trail's stored vector.
5. Sort by score, highest first, and keep the top 5.
6. Print each of the 5 on one line: score to four decimals, `id`, `name`, `difficulty`, `distance_mi`, and the `features` list.

**Why:** the query has to go through the same model as the descriptions — vectors from two different models are not comparable, and cosine similarity will still return confident-looking numbers if you mix them. The result is a number near 1 when two texts mean similar things and near 0 when they do not.

**Check:** at least two of `trail-0068`, `trail-0055`, `trail-0027`, `trail-0011` in your top 3; recorded first is `trail-0068` Carlon Falls at `0.7733`. `trail-0007` Upper Yosemite Falls first means you are still counting keyword hits, the step 0 baseline.

### Step 4: Run the other two queries and read the scores

**Do:** run the program again with each of these:

```text
somewhere quiet to take my kids
```

```text
an easy hike to a great view
```

**Why:** these come from `data/queries.json`, which holds all three test queries with `id`, `query`, `why` (what that query proves), and `success_check`. Look at the scores themselves, not just the order — when the whole top 5 is low and close together, nothing in the catalog is a strong match, and the order is mostly noise.

**Check:** query 2 puts Taft Point (`trail-0020`) first at `0.4876` (correct, and still a cliff edge), with short family-friendly trails filling the rest of the top 5, all between 0.44 and 0.49; query 3 puts `trail-0058` Panorama Cliffs Bypass at or near the top (`0.6481`), top score `0.77` on query 1. `trail-0074` Easy Creek Trail in query 3's top 5 means you are matching words, not meaning.

### Stretch goals

Pick either. Each uses information you already have — trail metadata or the keyword hits — instead of hoping the vector carries it. The embedding fixed recall (finding the right trails); these fix precision (dropping the wrong ones).

- **Filter before you rank.** Between step 3's item 3 and item 4, drop trails whose `difficulty` is `hard`. For query 1, also drop trails whose `features` list does not contain `dog-friendly`. Then rank whatever is left. **Check:** query 1 returns only dog-friendly trails; query 3 loses Beehive Loop (`trail-0017`) and Chimney Tops (`trail-0005`) from the top 5. Either still there means the filter ran after the top 5 was taken.
- **Blend in the keyword score.** Keep the keyword-hit count from step 0 alongside the cosine score, rescale both to 0..1, and rank on a weighted sum of the two. Print both scores on each row. **Check:** every row shows a cosine score and a keyword score. A trail that appears only because of its keyword count, such as Easy Creek Trail (`trail-0074`) on query 3, means the keyword weight is too high. Feature 05 measures the same blend on a bigger corpus.

## Pick a Track

Every track does the same steps against the same data and checks against the same [`expected-output.md`](expected-output.md). Each folder's walkthrough maps the steps above onto that track.

| Track | Start here | What you edit |
|---|---|---|
| .NET | [`dotnet/F04-dotnet.md`](dotnet/F04-dotnet.md) | `dotnet/starter/Program.cs` |
| Python | [`python/F04-python.md`](python/F04-python.md) | `python/starter/main.py` |
| TypeScript | [`typescript/F04-typescript.md`](typescript/F04-typescript.md) | `typescript/starter/index.ts` |

Every track has a `complete/` next to its `starter/`, which is the answer key.

## What Is in This Folder

- `data/trails-slice.json`: 30 trails lifted verbatim from feature 10's 200-trail [`trails.json`](../../M4-doing/F10-agentic-workflows/data/trails.json), chosen so the catalog has four right answers for query 1 that never use the query's words, and three keyword traps. Step 1 names them.
- `data/queries.json`: the three test queries, why each one is in the set, and the success check for each. Step 4 describes the record shape.
- `embeddings.json`: not in the folder until you run the program. Step 2 explains the cache.
- `expected-output.md`: real `nomic-embed-text` rankings for all three queries, the keyword baseline they beat, cosine-similarity pseudocode, and two results that are honestly wrong and worth talking about.
