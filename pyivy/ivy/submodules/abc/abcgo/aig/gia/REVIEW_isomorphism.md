# GIA isomorphism review (in progress)

Scope: giaIso.c, giaIso2.c and giaIso3.c; Go implementations are split among
iso_refine/canonical/reduce/diagnostics, iso2_refine/compare and giaIso3.go.
Review these sequentially within the open GIA package audit. The original
131-package objective remains active, with2/131 packages complete.

## Plan

- [x] Locate original routines, actual Go implementations and existing tests.
- [ ] Reconcile first-generation level classes, packed fields, unsigned
  signatures, sorting/ties, forward/backward refinement and convergence.
- [ ] Review canonical traversal, PI/register permutations, serialized keys,
  output reduction, dual-output and estimate/ownership boundaries.
- [ ] Reconcile second-generation hash/unique labels, cone traversals, sequential
  comparison and quality/skip reduction policies with the original source.
- [ ] Review synchronous/reverse signatures, source RNG and diagnostic drivers.
- [ ] Add red Go-only regressions before uncovered bug fixes, then independent
  scalar/classification/permutation/sequence tests and compactness checks.
- [ ] Reconcile all source names, freeze edits, run affected race/vet and full
  module Go-only integration before claiming family closure.

Initial findings to investigate: first-generation level allocation only uses
output depth despite including all nodes, making deeper dangling cones unsafe.
The source explicitly resets terminal levels before computing node levels.
Second-generation uniquification currently uses a new map instead of its
allocated reusable hash workspace; examine ordering, mark/scratch ownership and
compactness before deciding whether this is a faithful representation change.
Several verbose reports/diagnostic drivers omit source behavior. Existing simple
tests do not prove this whole family correct; no new C fixtures will be used.


## First level-preparation repair

Inventory totals:33 routines in giaIso.c (including four packed-word accessors),
23 in giaIso2.c and12 in giaIso3.c. Closure requires all68, not just their
currently green simple wrappers. Main refinement/canonicalization and second-/
third-generation runtime reconciliation remain open.

A new depth-five dangling-chain regression fails at the output-sized level
array (/tmp/abcgo-iso-dangling-levels-red.log). C's Gia_IsoPrepare similarly
allocates from output depth while classifying every object, so this is an
explicit repair of inherited out-of-bounds behavior: include all AND levels in
the maximum. Retain all nodes, bucket order and source refinement algorithm;
do not silently prune unused cones. Separately, a new terminal-reset regression
fails with stale negative levels (/tmp/abcgo-iso-terminal-levels-red.log).
Restore C's explicit constant/CI level initialization before computing ANDs.

Twenty-four chain-depth/output/sequential combinations cover depths1/2/5/32,
constant/input/shallow-AND outputs and RI/RO boundaries. Independent stable
buckets check every level, class record and packed object ID; low signature
words and graph node Value fields must stay unchanged. The terminal-reset case
checks the source initialization contract separately. All fixtures are Go-only.

Frozen full Go-only module exits0 (130 tested packages,3 commands without
tests), focused existing/new isomorphism race1.025s and GIA vet exit0. Logs:
/tmp/abcgo-iso-prepare-final-{module.jsonl,race.log,vet.log}. All sessions ended.
Only this preparation checkpoint is closed; the family and full GIA package
remain in progress, original package completion2/131.
