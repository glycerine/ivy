# Full extrab package completion and validation

The normal root `bdd/extrab` package builds and all of its source families are
implemented. Original `src/bdd/extrab` files remain visible and unchanged; there
are no build tags, renamed drafts, placeholder algorithms or suppressed files
used to obtain the package gate. Authoritative graph primitives live in
`bdd/cudd/core`; cascade, primes and sets have their single importable leaf
implementations with root API forwarding.

## Separate readability and compactness review

Family reviews cover unateness/threshold, set filters, max/min antichains,
auto/image/cascade, K-map/miscellaneous, and symmetry/deadlines. Repeated validation,
BDD cofactor splitting, positive cubes and subset selection live in the shared
header. Global C scratch/cache state is replaced by per-operation memo maps and
persistent image-tree objects. Temporary image results are reused, and ordinary
and timed core operations share algorithms rather than keeping divergent ports.
See `AUTO_IMAGE_CASCADE_REVIEW.md`, `KMAP_MISC_REVIEW.md`,
`SYMMETRY_TIME_REVIEW.md`, and the existing family/leaf review documents.

## Separate bug review

The reviews checked ownership, complemented phases, shuffled level/index maps,
paired-literal domains, constant/empty-family cases, caller buffers and explicit
reference contracts. Findings fixed with regressions include partition IDs
outnumbering BDD variables, cost/count overflow, stale global ISOP phase counts,
strict constant cube limits, bounded allocation rollback, malformed tuple/pair
inputs, impossible random-family requests, and K-map relation dimension overflow.
Native semantic fixtures preserve actual source behavior such as signed-space
recursions, empty-set singleton families and the scalar symmetry prefilter.
Deliberate boundary changes are documented, especially Go `time.Time` deadlines,
writer errors, invalid-domain rejection, and explicit BDD-only timed transfer
rather than corrupting numeric ADD constants as the C routine could.

## Final ordinary gates

Both commands passed on the final code (2026-11-12):

```sh
GOCACHE=/tmp/abcgo-go-cache go -C abcgo test -race ./bdd/extrab ./bdd/extrab/cascade ./bdd/extrab/primes ./bdd/extrab/sets ./bdd/cudd/core ./bdd/cas ./bdd/reo -count=1
GOCACHE=/tmp/abcgo-go-cache go -C abcgo vet ./bdd/extrab ./bdd/extrab/cascade ./bdd/extrab/primes ./bdd/extrab/sets ./bdd/cudd/core ./bdd/cas ./bdd/reo
```

Race results: extrab 7.402s; cascade 2.072s; primes 1.062s; sets 13.849s;
core 25.674s; CAS 4.238s; REO 1.112s. The initial ordinary
`go test ./bdd/extrab ./bdd/cudd/core` also passed. No explicit-file filtering is
needed or used for the final gate. Normal Go tests require no C compiler.
Native fixture regeneration commands/checksums are in `test_vectors/README.md`.

These are package/dependency gates, not a claim that every other ABC package or
the unfinished root `bdd/cudd` package is complete. Managers remain single-thread
objects. Cooperative deadlines do not hard-preempt allocator/validation work.
Valid highly skewed random-set distributions may require many samples. K-map
output uses portable ASCII rather than Windows console colors. No benchmark or
large demo is automatically run by the tests.
