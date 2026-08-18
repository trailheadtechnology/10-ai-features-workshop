# What Passing Looks Like

Wording varies run to run; the checks below are what has to be true. These samples came from actual `llama3.2` runs of `http/ollama.http`.

## Request 1: Naive Prompt on tr-0001 (The Failure You're Supposed to Get)

A faithful book report, mostly about the author's gear:

> The author recently completed the Avalanche Lake Trail in Glacier National Park on July 12th with a new, lighter gear setup. [...] The author's new gear consisted of: a lightweight 24-liter frameless day pack (14 oz empty), trail runners instead of boots [...]

The summary isn't wrong, just useless to a hiker planning this weekend. Right model, wrong instruction.

## Request 2: Improved Prompt on tr-0001

Three bullets, and nothing the report didn't say:

> - **Current trail conditions**: The trail was in fine shape, packed dirt and roots, with a couple of muddy patches.
> - **Hazards or closures**: No closures or hazards reported.
> - **Crowding**: Heavy by 10am, but if you want to have the lake to yourself, be walking by 7:30.

**Check:** the gear debrief (most of the source text) is gone, mud and crowds made it in, and nothing is closed. tr-0001 contains no closure and no trail hazard. The only bear in it is a ranger's remark that a black bear crossed a road near Lake McDonald earlier in the week. A run that keeps the bear as a plain sighting is fine. A run that closes a trail because of it is a hallucination, and that is the failure this prompt is written to avoid (see the note below).

## Request 3: Improved Prompt on tr-0004 (The Success Check)

> - Trail to Avalanche Lake is currently closed due to the washed-out footbridge over the creek at the gorge.
> - No closures or hazards reported for the remaining trails, including the Trail of the Cedars boardwalk.
> - Moderate crowding, particularly at the trailheads and popular spots along Going-to-the-Sun Road.

**Check:** the washed-out bridge appears, and appears first. In the source report that fact is buried in the fourth paragraph between airport sandwiches and huckleberry ice cream. If your bullets mention the sister, the deer, or Moby the rental SUV, your prompt needs work. A second bullet reading "no other closures reported" is normal; the model is answering the hazards slot after already spending the closure.

## The Bug That Produced the Last Three Lines of This Prompt

The first version of the briefing prompt was four lines and stopped after "If the report mentions a closure or hazard, it must appear in the first bullet." It reads fine. It also hallucinated.

Run against tr-0001 twenty-four times, that prompt asserted a hazard, a closure, or a warning the report never made in **11 of 24 runs (46%)**. About half of those were flatly invented closures:

> - **Hazards or closures:** The trailhead had been closed due to a black bear crossing the road near Lake McDonald earlier in the week [...]

> - The Avalanche Lake Trail has been closed to any further development due to an avalanche risk at the lake outlet (no further information is provided in the report).

The model was handed a slot labeled "hazards or closures" and told the hazard goes first. With no hazard in the report, it promoted the nearest available noun (a bear, a creek, the word "avalanche" in the trail's name) into one. The instruction that made the good version work is the same instruction that made this failure likely.

Two lines fixed it:

    Report only what the trip report states. Do not turn a wildlife sighting into a
    hazard or a closure, and write "no closures or hazards reported" when it says none.

Same twenty-four runs on tr-0001: **1 of 24 (4%)**. tr-0004 still leads with the washed-out bridge in **12 of 12** runs, so the success check did not regress.

Two things worth taking away. Naming the specific wrong move ("do not turn a wildlife sighting into a closure") beat a general plea to be accurate; a version that only asked for every bullet to be supported by the report's own words still invented closures in 2 of 12 runs. And giving the model a legal way to say nothing matters, because a required slot with no honest filler is an invitation to invent one. Also note that 4% is not 0%. Small local models stay probabilistic, so a feature that ships this way needs the output checked or a bigger model, not just a better paragraph.

The first prompt you write is rarely the one you ship. This one took several measured revisions, and some of them were worse than the original. One rewrite that labeled the bullets and put closures first read beautifully on tr-0001 and then answered "Closures or hazards: none reported" on tr-0004, directly above a bullet describing the closed trail.

## Stretch Goal

Same report, two audiences. A hiker summary keeps closures, mud, and crowds; a ranger/maintenance summary should instead surface things like the washed-out footbridge's location, the barricade, and visitor turnaround volume. If both summaries are identical, the audience isn't reaching the prompt.
