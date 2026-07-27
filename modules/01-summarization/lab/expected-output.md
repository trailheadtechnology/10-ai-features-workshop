# What passing looks like

Wording varies run to run; the checks below are what has to be true. These samples came from an actual `llama3.2` run of `ollama.http`.

## Request 1: naive prompt on tr-0001 (the failure you're supposed to get)

A faithful book report, mostly about the author's gear:

> The author recently completed the Avalanche Lake Trail in Glacier National Park on July 12th with a new, lighter gear setup. [...] The author's new gear consisted of: a lightweight 24-liter frameless day pack (14 oz empty), trail runners instead of boots [...]

The summary isn't wrong. It's just useless to a hiker planning this weekend, which is the point. Right model, wrong instruction.

## Request 2: improved prompt on tr-0001

Three bullets, conditions first:

> - **Trail Conditions:** The trail was in fine shape, with muddy patches about 1 mile in where a seep crosses the tread.
> - **Hazards:** A ranger reported a black bear crossing the road near Lake McDonald earlier in the week.
> - **Crowding:** Heavy by 10am; be walking by 7:30 if you want the lake to yourself.

**Check:** the gear debrief (most of the source text) is gone, and mud, bear, and crowds each made it in.

## Request 3: improved prompt on tr-0004 (the success check)

> - **Closure:** The trail to Avalanche Lake is currently closed due to the washed-out footbridge over the creek at the gorge, until a crew can repair it.
> - ...

**Check:** the washed-out bridge appears, and appears first. In the source report that fact is buried in the fourth paragraph between airport sandwiches and huckleberry ice cream. If your bullets mention the sister, the deer, or Moby the rental SUV, your prompt needs work.

## Stretch goal

Same report, two audiences. A hiker summary keeps closures, mud, and crowds; a ranger/maintenance summary should instead surface things like the washed-out footbridge's location, the barricade, and visitor turnaround volume. If both summaries are identical, the audience isn't reaching the prompt.
