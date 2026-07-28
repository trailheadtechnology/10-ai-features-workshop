# What passing looks like

Scores below came from a real `nomic-embed-text` run over the 30 descriptions in
`trails.json`, and they match `trail-embeddings.json` byte for byte. If you
compute your own vectors, expect the same ordering and scores within a few
thousandths. Different embedding models will shuffle the tail; the head should
survive.

There is more than one right answer here. Each target lists the neighbors this
run produced plus the wider set that is defensible, so grade yourself on
overlap, not on an exact match.

## Cosine sanity check

Before ranking anything, confirm your similarity function works. Using the
vectors from request 2 in `ollama.http`:

| pair | cosine |
| --- | --- |
| Avalanche Lake (trail-0117) vs Gunsight Lake (trail-0086) | 0.7849 |
| Avalanche Lake (trail-0117) vs Checkerboard Wash (trail-0041) | 0.6117 |

Two Glacier lake hikes score high, a Zion desert wash scores low. Vectors are
768 floats. If your two numbers are equal, or either is above 0.99, you are
comparing a vector to itself somewhere.

## Target 1: trail-0117, Avalanche Lake Trail

```
0.7849  trail-0086  Gunsight Lake Approach (Glacier National Park, hard)
0.7758  trail-0168  Black Lake via Glacier Gorge (Rocky Mountain National Park, hard)
0.7692  trail-0100  Piegan Pass Trail (Glacier National Park, hard)
0.7642  trail-0010  Alum Cave Trail (Great Smoky Mountains National Park, moderate)
0.7590  trail-0091  Otokomi Lake Trail (Glacier National Park, hard)
```

**Check:** at least three of your five come from {trail-0086, trail-0168,
trail-0100, trail-0091, trail-0080 Chasm Lake, trail-0186 Snyder Lake,
trail-0196 Fern Lake, trail-0141 Akaiyan Falls}. Ranks 6 through 8 in this run
were Chasm Lake (0.7589), Snyder Lake (0.7544), and Fern Lake (0.7508), all
within 0.01 of the trails that made the cut, so a slightly different order is
not a bug.

**Read the results before you celebrate.** Two things are wrong with this list:

- **Alum Cave Trail** at #4 is a Smokies hike with no lake, no waterfall, and
  nothing in common with Avalanche Lake except a description full of rock
  formations and a steady grade. The model matched prose texture, not the thing
  a hiker cares about.
- **Four of the five are rated hard.** Avalanche Lake is a moderate 4.6-mile
  family hike, and this list hands a family a 20-mile approach to Gunsight Lake.
  The descriptions never mention difficulty, so the embedding cannot see it.

Both are the module's honest caveat made concrete. The fix is not a better
embedding model, it is metadata: filter or re-rank by difficulty and distance
after the similarity pass. That is the stretch goal, and it is also the answer
to "would you ship this?"

## Target 2: trail-0003, Trail of the Cedars

```
0.7858  trail-0027  Grotto Falls Trail (Great Smoky Mountains National Park, moderate)
0.7422  trail-0068  Carlon Falls Trail (Yosemite National Park, moderate)
0.7282  trail-0131  Gatlinburg Gateway Greenway (Great Smoky Mountains National Park, easy)
0.7230  trail-0048  Coalpits Wash Trail (Zion National Park, moderate)
0.7228  trail-0039  Ship Harbor Nature Trail (Acadia National Park, easy)
```

**Check:** Grotto Falls should be #1 by a wide margin, and at least two of
{trail-0068, trail-0131, trail-0039, trail-0070 Oconaluftee River, trail-0150
Lily Lake} should appear. The top of this list is the good case: old-growth
shade and dog-friendly flat walking cluster on their own.

Note the drop after #1. Grotto Falls sits 0.04 above second place, then the
scores flatten into a crowd separated by thousandths. **Coalpits Wash** at #4 is
a Zion desert route with no trees at all, and it is there because the tail of
this ranking is noise. A production feature would set a floor (say 0.74) and
show three results instead of padding to five.

## Target 3: trail-0008, Highline Trail

```
0.7497  trail-0141  Akaiyan Falls via Sperry Junction (Glacier National Park, hard)
0.7447  trail-0152  El Capitan Meadow Walk (Yosemite National Park, easy)
0.7410  trail-0041  Checkerboard Wash Trail (Zion National Park, easy)
0.7371  trail-0127  Merced Grove Trail (Yosemite National Park, moderate)
0.7217  trail-0143  Cataloochee Divide Trail (Great Smoky Mountains National Park, moderate)
```

**Check:** this one is supposed to disappoint you. Every score is below 0.75 and
they span 0.03 top to bottom, which is another way of saying nothing in the
slice is much like the Highline. The honest read is that the catalog has no good
neighbor for a cliff-ledge traverse, and a real product would show nothing here
rather than five weak guesses.

The reason is visible in the source text. The Highline's description is about
exposure, a hand cable, and mountain goats, and the model latched onto the
wildlife-watching angle: El Capitan Meadow (people-watching climbers),
Checkerboard Wash (bighorn sheep), Cataloochee Divide (elk). Those share a
sentence shape, not a hiking experience. Any five results with scores in the
0.71 to 0.75 band pass; the point is noticing the band is low.

## Gear: bought the Cascade 65 Backpack

Products have no descriptions, so each product's vector comes from all of its
reviews concatenated. Real run over the 25 products in `../../data/gear-reviews.jsonl`:

```
0.8039  Cascade 40 Daypack
0.7854  Juniper Camp Chair
0.7643  Granite Peak 2P Tent
0.7634  Cirrus 0° Sleeping Bag
0.7633  Packrat Stuff Sack Set
```

**Check:** Cascade 40 Daypack should be #1.

And #1 is the problem. The Cascade 40 is the one product a Cascade 65 owner will
never buy, because it is the same pack in a smaller size, and it wins precisely
because reviewers describe it in nearly the same words. This is what content
similarity does: it finds substitutes. Recommending complements is a different
question that the vectors cannot answer.

Now count what reviewers actually mention alongside the Cascade 65 in the same
review text:

| co-mentioned product | reviews |
| --- | --- |
| Summit Bear Canister | 3 |
| Packrat Stuff Sack Set | 2 |
| CloudRest Sleeping Pad | 1 |
| Cirrus 20° Sleeping Bag | 1 |
| Granite Peak 2P Tent | 1 |

The bear canister is the answer a shopper wants ("fits the Summit Bear Canister
horizontally, which almost nothing does"), and the embedding ranking does not
list it at all. That signal is a `Contains` call over 300 review lines, no model
involved. It is also a preview of collaborative filtering: behavior and
co-occurrence answer questions similarity cannot, which is the "when to add
behavior data" beat at the end of the module.

## Stretch goal

- **Two trails at once:** average the vectors for trail-0117 and trail-0003 and
  rank against that. This run gave Carlon Falls (0.7947), Grotto Falls (0.7946),
  Alum Cave (0.7921), Coalpits Wash (0.7833), and Akaiyan Falls (0.7805). The
  hard alpine lakes that dominated target 1 are gone, replaced by shaded
  waterfall walks, because the average sits between the two tastes. Note that
  every score went up, which is a property of averaging vectors and not a sign
  the results got better.
- **Metadata filter:** re-run target 1 keeping only trails rated easy or
  moderate. You get Alum Cave (0.7642), Fern Lake (0.7508), Carlon Falls
  (0.7047), Coalpits Wash (0.7032), and Cataloochee Divide (0.7031). Honest
  reading: the filter removes the trails a family cannot do, and what is left is
  thin, because this 30-trail slice has almost no easy lake hikes. Filtering
  makes a bad list shorter, not longer. Deciding which of these two lists you
  would put at the bottom of the Avalanche Lake page is the judgment call the
  lab is really about.
