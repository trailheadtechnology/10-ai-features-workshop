# What Passing Looks Like

Every number below came from a real run: `nomic-embed-text` on Ollama, the 30 descriptions in `trails-slice.json`, cosine similarity in plain code. Your scores should land within a few thousandths of these. Embedding models are deterministic in a way chat models are not, so unlike feature 01, you can compare digits.

The keyword results come from `dotnet/starter`, which is the baseline this feature exists to beat: lowercase the query, keep words of three letters or more, count whole-word hits in each trail's name plus description.

## Cosine Similarity, in Case You Want the Pseudocode

```
dot = 0; magA = 0; magB = 0
for i in 0..len(a):
    dot  += a[i] * b[i]
    magA += a[i] * a[i]
    magB += b[i] * b[i]
return dot / (sqrt(magA) * sqrt(magB))
```

That is the entire algorithm. `nomic-embed-text` returns 768 floats per input, so `a` and `b` are 768 long. The first eight floats of "a gentle grade shaded by cedars" look like this:

```
[0.0611, -0.0001, -0.2007, -0.0624, -0.022, -0.0063, 0.0232, -0.0023]
```

## Query 1: "Dog-Friendly Waterfall Hike, Not Too Steep"

**Keyword search (the failure):**

```
2 word(s) [waterfall, too]  trail-0007  Upper Yosemite Falls Trail (hard, 7.2 mi)
1 word(s) [waterfall]       trail-0033  Running Eagle Falls Trail (easy, 0.6 mi)
1 word(s) [friendly]        trail-0047  Laurel Falls Trail (easy, 2.6 mi)
1 word(s) [steep]           trail-0058  Panorama Cliffs Bypass (moderate, 5.6 mi)
1 word(s) [waterfall]       trail-0068  Carlon Falls Trail (moderate, 3.8 mi)
```

The number one result is a hard 7.2-mile switchback climb that scored well partly on the word "too." Number three matched "friendly" out of "family-friendly," which says nothing about dogs. Number four matched "steep" from a description that says the trail *avoids* the steep sections; the word is there, the meaning is inverted. Three of the four trails that actually allow dogs are missing entirely, because their descriptions say "leashed dogs are welcome," and no user types that.

**Semantic search (the payoff):**

```
0.7733  trail-0068  Carlon Falls Trail (moderate, 3.8 mi)   [waterfall, dog-friendly, river, swimming]
0.7334  trail-0055  Copeland Falls Path (easy, 1.2 mi)      [waterfall, dog-friendly, river, family-friendly]
0.7296  trail-0027  Grotto Falls Trail (moderate, 2.6 mi)   [waterfall, dog-friendly, old-growth]
0.6907  trail-0011  Hadlock Falls Loop (easy, 2.1 mi)       [waterfall, dog-friendly, carriage-road, bridge]
0.6787  trail-0033  Running Eagle Falls Trail (easy, 0.6 mi)
```

All four dog-friendly waterfall trails, in the top four, none of which contain the words "dog-friendly" or "steep." The fifth is a half-mile paved waterfall walk, which is a defensible near-miss. The hard 7.2-mile climb that keyword search loved has dropped off the list.

**Check:** at least two of trail-0068, trail-0055, trail-0027, trail-0011 in your top 3.

## Query 2: "Somewhere Quiet to Take My Kids"

**Keyword search:**

```
1 word(s) [kids]  trail-0004  Ocean Path (easy, 4.4 mi)
```

One result out of thirty trails, and only because a description happens to use the word "kids." A user seeing this concludes the app has one family trail.

**Semantic search:**

```
0.4876  trail-0020  Taft Point Trail (easy, 2.3 mi)         [viewpoint, fissures, sunset]
0.4586  trail-0055  Copeland Falls Path (easy, 1.2 mi)      [waterfall, dog-friendly, river, family-friendly]
0.4565  trail-0108  Virginia Falls Trail (moderate, 3.6 mi) [waterfall, creek, family-friendly]
0.4411  trail-0011  Hadlock Falls Loop (easy, 2.1 mi)
0.4364  trail-0027  Grotto Falls Trail (moderate, 2.6 mi)
```

Results two through five are exactly right: short, gentle, family-friendly. Result one is the honest surprise of this lab, and it is worth ten seconds on stage. Taft Point's description reads "Keep children close; the ground gives no second chances here," next to a three-thousand-foot drop. The embedding caught "children" and "relaxed forest stroll" and scored it top of the list. Semantically it is about kids. As a recommendation it is the worst trail in the slice for this user.

Two lessons come out of that one row. First, embeddings capture topic, not sentiment or safety; nothing in cosine similarity knows the difference between "great for kids" and "dangerous for kids." Second, look at the scores: the whole list sits around 0.44 to 0.49, well below query 1's 0.77. Low absolute scores with a tight spread mean the catalog has no strong match and the ranking is mostly noise. A production search box should treat that as "no confident results" rather than showing the top hit with a straight face. This is the argument for a score floor, for metadata filters on top of the vector score, and eventually for reranking.

**Check:** family-friendly short trails fill most of your top 5. If you got Taft Point at number one, your code is correct.

## Query 3: "An Easy Hike to a Great View"

**Keyword search (the trap fires):**

```
1 word(s) [easy]  trail-0017  Beehive Loop (hard, 1.5 mi)
1 word(s) [easy]  trail-0055  Copeland Falls Path (easy, 1.2 mi)
1 word(s) [easy]  trail-0074  Easy Creek Trail (hard, 7.8 mi)
1 word(s) [view]  trail-0146  Canyon Overlook Trail (easy, 1 mi)
```

Two of the four hits are hard trails that matched on the word "easy." Beehive Loop matched from "a fear of heights will find you halfway up with no easy retreat." Easy Creek Trail matched on its name, which honors homesteader Elias Easy and not the 2,610 feet of loose rubble the trail actually climbs. The slice also contains Dog Lake Trail (trail-0187), which forbids pets outright and will happily match a search for "dog."

**Semantic search:**

```
0.6481  trail-0058  Panorama Cliffs Bypass (moderate, 5.6 mi)  [viewpoint, forest]
0.6235  trail-0004  Ocean Path (easy, 4.4 mi)                  [coastline, tide-pools, family-friendly]
0.5932  trail-0017  Beehive Loop (hard, 1.5 mi)                [iron-rungs, exposure, viewpoint]
0.5841  trail-0087  Sprague Lake Loop (easy, 0.9 mi)           [lake, accessible, family-friendly, reflections]
0.5790  trail-0005  Chimney Tops Trail (hard, 3.8 mi)          [viewpoint, rock-scramble]
```

Number one is the trail whose entire description is about avoiding the steep route while keeping the views, and it never uses the word "easy." Easy Creek Trail is gone from the list, name trap and all. Beehive Loop survives at number three, though, and it is a cliff with iron rungs. Same lesson as query 2: the model reads "easy" and "views" as topics in the text without judging whether the trail is easy.

That is the honest shape of this feature. Semantic search fixes recall, which is the thing keyword search is catastrophically bad at. It does not fix precision by itself. The stretch goal below is where precision comes from.

**Check:** trail-0058 at or near the top, and trail-0074 nowhere in your top 5.

## Stretch Goal

Filter before you rank, or blend scores. Both queries 2 and 3 clean up immediately if you drop trails whose `difficulty` is `hard` before ranking, or if you require the `dog-friendly` feature for query 1 instead of hoping the vector carries it. Metadata you already have is cheaper and more reliable than the embedding for anything expressible as a filter. Use the vector for the part users can only say in words.
