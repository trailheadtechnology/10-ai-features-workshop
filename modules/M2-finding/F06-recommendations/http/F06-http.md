# HTTP Walkthrough for 06 Recommendations

The lab steps in [`../F06-lab.md`](../F06-lab.md), done from raw requests. Nothing here needs a language: run the requests as they are, or copy their bodies into whatever HTTP client you have.

- `ollama.http`: the embeddings requests, needed only if you want fresh vectors. One description, a batch of three for a cosine sanity check, and the gear variant that embeds a product's reviews.

## Running the Requests

Open the `.http` file in VS Code with the REST Client extension (or a JetBrains IDE; the built-in HTTP client reads the same files), put the cursor in a request, and click **Send Request** above it. The response opens in a side pane. Requests are separated by `###` lines and numbered; run them in order. Every request body is plain JSON, so porting one to your language is copying the body into whatever HTTP client you already have: `requests` in Python, `fetch` in Node, `HttpClient` in .NET, curl in a shell.

### Lab Step 1: Vectors, Then Nearest Neighbors

If you did feature 04, reuse its vectors; if not, `data/trail-embeddings.json` has all 30 precomputed, or request 2 shows how to embed a batch fresh. Take target trail `trail-0117`'s vector, cosine-score every other trail against it, skip the target, print the top 5.

Check: Gunsight Lake Approach at 0.7849 on top. Read the difficulty column against the target's.

### Lab Step 2: The Other Targets

`trail-0003` Trail of the Cedars and `trail-0008` Highline Trail. Same code, different target vector.

Check: substantial overlap with the acceptable sets in [`../expected-output.md`](../expected-output.md); there is more than one right answer, and the file says which neighbors are defensible.

### Lab Step 3: Would You Ship It?

Read your own lists as a product owner. One target has no real neighbors at all; a shipping product shows nothing rather than five weak guesses.

### Stretch

Average two trails' vectors and rank against the average, or filter to the same park or an easier difficulty before ranking. Request 3 embeds a product's reviews for the gear variant, where the top hit for the Cascade 65 is the Cascade 40.
