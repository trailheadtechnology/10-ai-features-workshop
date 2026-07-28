# What passing looks like

Every number below came from a real run against Ollama with `nomic-embed-text` on
2026-07-27, using `dotnet/complete`. Embeddings are deterministic for a fixed model,
so if your pipeline matches you should reproduce these distances to about four
decimal places. If your numbers differ in the third decimal, check your
normalization before you check anything else.

The pipeline, in full:

1. Embed `"classification: " + report.text` for all 40 trail-0117 reports.
2. L2-normalize every vector.
3. Average them component-wise and normalize the result. That is the centroid.
4. Score each report as `1 - dot(vector, centroid)`, which is cosine distance.
5. Sort descending.

## The thing you need to know before you start

Skip step 1's prefix and this lab does not work. That is not a footnote, it is the
main lesson, so here it is up front with both rankings side by side.

`nomic-embed-text` is trained with task prefixes (`search_query:`,
`search_document:`, `clustering:`, `classification:`). Feed it bare text and you
get a usable-looking vector that is quietly off-distribution. On 40 short trail
reports, that is the difference between a working detector and a broken one.

| | no prefix | `classification: ` prefix |
|---|---|---|
| rank of `cr-0429` ("the footbridge over the gorge is completely gone") | 11 of 40 | 2 of 40 |
| washout reports in the top 5 | 0 | 3 |
| alerts fired | 2, both false positives | 1, all three reports genuine |

Both runs are reproducible: `dotnet run` and `dotnet run -- --raw` in
`../dotnet/complete/`.

## The ranking (trail-0117, `classification:` prefix)

Mean distance 0.1935, standard deviation 0.0274. `!` marks reports above the
threshold, which the program sets at mean + 1 sd = 0.2208.

```
  dist    id       date        report
 !0.2561  cr-0496  2026-07-22  Trail is lovely as far as the creek, but with the bridge gone…
 !0.2399  cr-0429  2026-06-18  The footbridge over the gorge is completely gone. Creek is ra…
 !0.2371  cr-0282  2025-09-15  Parking lot was full by 8am, arrive early. Take your time on …
 !0.2347  cr-0464  2026-07-05  Bridge is still out at the gorge. Rangers say no repair timel…
 !0.2260  cr-0354  2026-05-08  The avalanche chutes are green and full of glacier lilies.
 !0.2257  cr-0436  2026-06-21  Crossing at the gorge is impassable, do not attempt. The whol…
 !0.2219  cr-0329  2025-10-10  Dusty in the exposed sections but otherwise perfect. Started …
  0.2206  cr-0125  2025-07-03  Took the kids up. Skeeters bad at dusk, tolerable at midday. …
  0.2190  cr-0039  2025-05-19  Quick morning lap. Blowdown about two miles in, a short scram…
  0.2139  cr-0371  2026-05-16  Frost on the boardwalks first thing in the morning, slick as …
  0.2115  cr-0009  2025-05-04  Waterbars are doing their job, tread is in great shape. Paint…
  0.2100  cr-0088  2025-06-12  Mosquitoes are out in force near the water, bring repellent.
  0.2093  cr-0412  2026-06-08  Deer flies chased us for a solid mile in the open stretch. Sp…
  0.2072  cr-0352  2026-05-07  Did an early start. Seasonal streams mostly dry, carry more w…
  0.2067  cr-0431  2026-06-19  Bridge washed out about two miles in after last night's storm…
  0.1986  cr-0133  2025-07-04  Took the kids up. Trail crew has cleared most of the winter d…
  0.1985  cr-0067  2025-06-01  Some erosion on the outer edge of the switchbacks, stay insid…
  0.1956  cr-0355  2026-05-08  Muddy in the usual low spots, gaiters not a bad idea. Sunscre…
  0.1951  cr-0177  2025-07-26  Loose gravel on the steep pitch, poles recommended. Everythin…
  0.1946  cr-0309  2025-10-01  Deep mud where the horses have churned things up.
  0.1945  cr-0480  2026-07-14  FYI the washed-out crossing has NOT been fixed. Saw a couple …
  0.1926  cr-0216  2025-08-13  Took the kids up. The falls are absolutely roaring with runof…
  0.1920  cr-0028  2025-05-14  A dusting of fresh snow up high, nothing serious yet.
  0.1914  cr-0438  2026-06-22  Heads up: the bridge is OUT. Water is still high and fast. A …
  0.1841  cr-0048  2025-05-23  Sunset hike. Clear and dry the whole way. Take your time on t…
  0.1796  cr-0290  2025-09-21  One big downed tree before the junction, well-worn path aroun…
  0.1752  cr-0173  2025-07-22  The falls are absolutely roaring with runoff.
  0.1727  cr-0344  2026-05-04  Boots got heavy in the meadow section, otherwise fine. Wildfl…
  0.1722  cr-0357  2026-05-10  Boots got heavy in the meadow section, otherwise fine.
  0.1715  cr-0185  2025-07-29  Trail crew has cleared most of the winter deadfall, nice work…
  0.1705  cr-0443  2026-06-24  Confirmed the washout everyone is posting about. Wreckage of …
  0.1658  cr-0042  2025-05-21  Late afternoon walk. Sloppy for the first mile, then it dries…
  0.1639  cr-0040  2025-05-19  Solo trip midweek. The switchbacks are greasy after the rain,…
  0.1634  cr-0410  2026-06-07  Muddy in the usual low spots, gaiters not a bad idea. Balsamr…
  0.1619  cr-0366  2026-05-14  Water levels dropping, the ford is only shin deep now. The sw…
  0.1605  cr-0395  2026-05-31  Best conditions I've seen this trail in.
  0.1535  cr-0340  2026-05-02  Icy spots in the trees where the sun never hits, watch your f…
  0.1535  cr-0376  2026-05-18  Icy spots in the trees where the sun never hits, watch your f…
  0.1496  cr-0102  2025-06-20  Creek crossings are all rock-hoppable at the moment. Started …
  0.1484  cr-0014  2025-05-07  Out and back before lunch. Icy spots in the trees where the s…
```

**Check:** `cr-0496`, `cr-0429`, `cr-0464`, and `cr-0436` are four of the top six.
Four of the eight washout-related reports sit above threshold. Every report at the
very bottom of the list is boilerplate mud, ice, and deadfall, which is what you
want: the densest part of the cluster scores lowest.

## Being honest about the ranking

The module card says the washout reports top the list "by a wide margin." On this
data, the first half of that is true and the second half is not. Three things are
worth saying out loud in the room, because they are what this technique actually
looks like in production:

**1. There is no clean gap.** Distances run from 0.1484 to 0.2561 with no visible
break anywhere. `cr-0496` at the top is 0.2561; `cr-0125` ("skeeters bad at dusk")
is 0.2206. The washout reports rank high, but they do not fall off a cliff away
from everything else. Any threshold you pick is a business decision about how much
review you can afford, not a mathematical boundary the data hands you. Mean + 1 sd
flags 7 reports; mean + 1.5 sd flags 4. Neither is "correct."

**2. Routine reports are not one tight cluster.** Diagnostics on these vectors:

| | average pairwise cosine distance |
|---|---|
| washout report to washout report | 0.261 |
| routine report to routine report | 0.342 |
| washout report to routine report | 0.401 |

The washout reports are genuinely tighter and genuinely separated. The problem is
that "normal" on this trail is not one thing. It is mud plus ice plus mosquitoes
plus wildflowers plus parking, and those subjects are as far from each other as
they are from a washed-out bridge. That is why `cr-0282` (parking) and `cr-0354`
(glacier lilies) rank third and fifth. They are not errors; they really are unusual
for this trail. They are just not incidents.

**3. The anomalies pollute their own baseline.** Eight of forty reports are about
the bridge, so the washout cluster drags the centroid toward itself and then
measures its distance from a centroid it helped build. Statisticians call this
masking. It is why `cr-0443` ("confirmed the washout everyone is posting about")
lands at rank 31, below reports about icy boardwalks. It is far from mud, but by
the time it was written it was close to normal, because seven other bridge reports
had redefined normal. Recomputing the centroid from only the 32 reports dated
before 2026-06-18 puts all eight washout reports in the top ten. That is the
sliding-window stretch goal, and it is not a trick; it is the correct way to build
a streaming detector.

## The alert rule is what actually works

Ranking alone is mediocre here. Ranking plus the cluster rule from step 4 of the
demo is good, and it is the part worth showing:

```
7 of 40 reports above threshold. Clustering them within 14 days:

  (ignored) cr-0282 2025-09-15 is a lone outlier, not an event
  (ignored) cr-0329 2025-10-10 is a lone outlier, not an event
  (ignored) cr-0354 2026-05-08 is a lone outlier, not an event
  ALERT trail-0117: 3 anomalous reports between 2026-06-18 and 2026-07-05
        cr-0429 2026-06-18  The footbridge over the gorge is completely gone. Creek is raging and…
        cr-0436 2026-06-21  Crossing at the gorge is impassable, do not attempt. The whole span w…
        cr-0464 2026-07-05  Bridge is still out at the gorge. Rangers say no repair timeline yet,…
  (ignored) cr-0496 2026-07-22 is a lone outlier, not an event

1 alert(s). Model calls: 40 embeddings, 0 chat completions.
```

One alert. Three reports in it. All three are real. Every false positive the
threshold produced was a lone report on a quiet week, and requiring two hits inside
14 days threw all of them away for free. The parking complaint and the glacier
lilies are unusual and unimportant, and the difference between those two words is
"did anyone else say it."

The first report in the alert is `cr-0429`, dated 2026-06-18, the day the bridge
went out. That is the CTO one-liner, met.

## The same run without the prefix (`dotnet run -- --raw`)

```
mean distance 0.2444 · sd 0.0345 · threshold mean+1sd = 0.2789

 !0.3116  cr-0329  2025-10-10  Dusty in the exposed sections but otherwise perfect…
 !0.3077  cr-0177  2025-07-26  Loose gravel on the steep pitch, poles recommended…
 !0.3019  cr-0309  2025-10-01  Deep mud where the horses have churned things up.
 !0.2831  cr-0028  2025-05-14  A dusting of fresh snow up high, nothing serious yet.
 !0.2823  cr-0354  2026-05-08  The avalanche chutes are green and full of glacier lilies.
 !0.2822  cr-0371  2026-05-16  Frost on the boardwalks first thing in the morning…
  0.2646  cr-0429  2026-06-18  The footbridge over the gorge is completely gone…   <- rank 11
  ...
  0.2006  cr-0443  2026-06-24  Confirmed the washout everyone is posting about…    <- rank 37
```

Two alerts fire, one for mud in October 2025 and one for glacier lilies in May
2026. Not one washout report is flagged. Same model, same math, same threshold
rule, one missing string literal. This is the most useful five minutes of the demo:
the failure is silent, the output looks plausible, and nothing throws.

## Stretch goal: trail-0042 and the bear cluster

`dotnet run -- --trail 0042`, same pipeline, 25 reports:

```
mean distance 0.1913 · sd 0.0379 · threshold mean+1sd = 0.2292

 !0.3067  cr-0446  2026-06-25  Saw a sow with two cubs grazing near the berry patches…
 !0.2580  cr-0449  2026-06-27  Ranger at the junction told us a food-conditioned bear…
 !0.2378  cr-0127  2025-07-04  Quick morning lap. Paintbrush and lupine everywhere…
 !0.2329  cr-0262  2025-09-04  Trail runner here. Popular with families on the first mile…
  0.2005  cr-0453  2026-06-30  Bear warnings posted at the trailhead this morning…
  ...
  0.1878  cr-0455  2026-07-02  Trail was closed for a few hours today due to bear activity…

  (ignored) cr-0127 2025-07-04 is a lone outlier, not an event
  (ignored) cr-0262 2025-09-04 is a lone outlier, not an event
  ALERT trail-0042: 2 anomalous reports between 2026-06-25 and 2026-06-27
        cr-0446 2026-06-25  Saw a sow with two cubs grazing near the berry patches at the meadow…
        cr-0449 2026-06-27  Ranger at the junction told us a food-conditioned bear has been hangi…
```

This one separates better than the washout does. `cr-0446` is the clearest anomaly
in either data set: 0.3067 against a mean of 0.1913, a full 3.0 standard deviations
out, with a visible 0.049 gap to second place. Trail-0042's routine reports are
more uniform ("clear and dry the whole way" three times over), so normal is a
tighter cluster, and a tighter normal makes a sharper detector. That comparison is
worth a sentence on stage: this technique's accuracy is a property of your corpus,
not of your code.

The alert fires on the first two bear reports, five days before the trail actually
closed on 2026-07-02. `cr-0453` and `cr-0455` fall below threshold, again because
by then bears were part of this trail's normal.

## Success checks

- Your top ten contains `cr-0429`, `cr-0436`, `cr-0464`, and `cr-0496`.
- Your bottom five is all mud, ice, and easy step-overs.
- You can state your threshold and why you chose it. "Mean plus one standard
  deviation" is a fine answer. "0.22" with no reasoning is not.
- Stretch: on trail-0042, `cr-0446` is your rank 1 by a clear margin.
- If your washout reports are ranked in the 30s, you forgot the prefix.
