# White-box LUT clustering review (in progress)

Scope: ten main routines and eight accessors in giaIff.c, implemented in
map/if/gia_iff.go to avoid an import cycle. The following giaIiff.c contains
two explicit source stubs; retain those return contracts rather than invent
a Boolean matcher. This is a sequential GIA family review, not the full IF
package audit. Original package completion remains2/131.

## Plan

- [x] Read original state, support counting, timing, selection and wrapper flow.
- [x] Add red Go-only regressions for initialization, float/empty boundaries,
  traversal ownership and source diagnostics before repairing differences.
- [ ] Reconcile strict first-candidate ties and exact packed traversal order.
- [ ] Exercise pairs/triples, reconvergence, shared outputs, phases, unequal
  delays, CI arrival profiles and white-box timing/ownership in Go fixtures.
- [ ] Review compactness separately, including count/selection scratch; retain
  original enumeration and clustering policy.
- [ ] Cover both giaIiff.c source stubs with unchanged-output assertions.
- [ ] Freeze files, run affected race/vet and full Go-only module integration,
  reconcile all source routines and update the sequential audit.

Initial inspection: Vec_FltStartFull uses all-one float bits (NaN), while Go
stores numeric -1. Source TimeOne compares arrivals explicitly, unlike Go max
which propagates NaN. Empty maxima use -ABC_INFINITY, not -MaxFloat32. Source
count/selection use GIA traversal epochs; Go allocates separate maps/flags and
omits those state changes. Wrapper rejection diagnostics also differ.


## Boundary repairs and source traversal storage

Twelve new Go-only regressions fail before production changes; log
/tmp/abcgo-iff-boundaries-red.log. Restore all-one float bits for unset arrivals,
source explicit TimeOne comparisons and early returns, -ABC_INFINITY empty
maxima, the nonnegative triple-boundary assertion, timing-before-count validation
order, Abc_MaxFloat output aggregation and exact unsupported-library messages.
A fixture initially omitted its LUT library and would have mistaken a nil panic
for the source guard; correct it before recording the final behavioral red log.

Count and selection now advance/use GIA traversal epochs, preserving observable
marks and source packed postorder. Remove the per-count map and per-selection
bool array; reserve the source-sized packing vector. The warmed sixteen-input
count regression goes from5 allocations to0, while retaining exact boundary
counts. Repeated selection checks absorbed roots, selected roots/terminals and
stale CO marks independently, as well as exact packed words.

Frozen full Go-only module exits0:130 tested packages pass,3 commands have no
tests. Full IF race77.344s and IF/GIA vet exit0; logs
/tmp/abcgo-iff-boundaries-final-{module.jsonl,race.log,vet.log}. Race compilers
are hidden from tests' PATH, with absolute CC/CXX and -B/usr/bin/ used solely
for Go's race-runtime build as in the mapping-family gate. All sessions ended.
Selection assertions, strict ties, broader timing/packing matrices and the
following Boolean-matching source stubs remain to review; no family closure.
