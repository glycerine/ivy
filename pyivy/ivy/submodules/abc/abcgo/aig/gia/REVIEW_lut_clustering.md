# White-box LUT clustering and matching stubs (reviewed)

Scope: ten main routines and eight accessors in giaIff.c, implemented in
map/if/gia_iff.go to avoid an import cycle. The following giaIiff.c contains
two explicit source stubs; retain those return contracts rather than invent
a Boolean matcher. This is a sequential GIA family review, not the full IF
package audit. Original package completion remains2/131.

## Plan

- [x] Read original state, support counting, timing, selection and wrapper flow.
- [x] Add red Go-only regressions for initialization, float/empty boundaries,
  traversal ownership and source diagnostics before repairing differences.
- [x] Reconcile strict first-candidate ties and exact packed traversal order.
- [x] Exercise pairs/triples, reconvergence, shared outputs, phases, unequal
  delays, CI arrival profiles and white-box timing/ownership in Go fixtures.
- [x] Review compactness separately, including count/selection scratch; retain
  original enumeration and clustering policy.
- [x] Cover both giaIiff.c source stubs with unchanged-output assertions.
- [x] Freeze files, run affected race/vet and full Go-only module integration,
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


## Selection guards, independent cluster oracle and timing ownership

Four invalid packed-root combinations and unordered NaN triple timing fail
before restoring the missing source positive-root/sentinel and strict-improvement
assertions (/tmp/abcgo-iff-selection-guards-red.log). Validate packed roots
before recursive child visits, as in C. Retain the source assertions rather
than silently constructing invalid packing records.

An independent set oracle enumerates root-plus-direct-child cluster sets,
computes their external boundary, minimizes delay with smaller/earlier ties,
and collects cluster dependencies in postorder. It does not call production
count/time/selection helpers. All1152 generated cases agree exactly: eight DAGs,
four edge phases, both degrees, three CI profiles, three delay libraries and
two output orders. Include reconvergence, dangling cuts, repeated/complemented
outputs, direct CI/constant outputs, unequal pin counts and size-dependent costs.
Check every stored time/match, packed word, CO time, unchanged mapping ownership
and every summary row; JSON confirms1152 leaf cases execute.

Separate explicit tie fixtures prove source preference for a single over equal
clusters, the first equally good pair, a pair over an equal triple, and the
first triple in deliberately permuted fanin order. White-box fixtures put a CO
before its box-output CI, propagate4 +2.5 +1 =7.5, and repeat wrappers without
stale timing state. Borrowed managers survive; temporary timing stays detached;
packing is replaced while mapping identity is retained. Four valid wrapper size/
library combinations cover the4-/6-LUT branches, alongside earlier rejection
messages. All eight accessors and independent manager/match storage are checked.
The shared wide fixture is completed before attaching mapping, avoiding manual
index-vector rebuilding after growing its output interface.

Frozen full Go-only module exits0 (130 tested packages,3 commands without tests),
full IF race77.317s and IF/GIA vet exit0. Logs:
/tmp/abcgo-iff-matrix-final-{module.jsonl,race.log,vet.log}. All sessions ended.
The giaIff.c clustering routines are reconciled below; following giaIiff.c stub
checks and the combined final documentation gate remain next. This is not a
complete IF or GIA package audit; original completion remains2/131.

| Original giaIff.c portion | Go representation and review |
| --- | --- |
| Start/Stop | Own times/match slices, source FF float fill and -1 matches; borrowed graph/library lifetime retained. Go assigns the graph in Start rather than waiting for Perform, so its object accessors are immediately usable; nil-safe Stop resets owned state. |
| Eight time/match accessors | Direct ID/object access with independent manager and pair/triple storage; both setter/getter directions tested. |
| Count | Source ordered cluster boundary and unique traversal marks; warmed zero-allocation count. |
| TimeOne | Root/second/third source early returns, explicit comparisons, empty bound and triple precondition; separate loops preserve its asymmetry for third-only skips. |
| TimeTwo/TimeThree | Source ordered direct-child enumeration, timing-before-count order, library lookup and strict ties; do not expand nested child clusters opportunistically. |
| Perform | Object-order CI/box/CO timing, mapped-node-only visits, all-cut counters, degree policy, source maxima and strict triple guard. |
| Select_rec/Select | Recursive closure with source epoch marks, pre-recursion match guards, dependency/group order and packing header; absorbed roots remain unmarked unless visited through another output. |
| Wrapper |4-/6-LUT library/degree mapping, exact failures/verbose prefix, temporary versus borrowed timing and packing replacement. Go keeps temporary timing local instead of attaching/detaching a transient graph pointer. |

Keep source clustering policy even when alternative packing could be better.
No optimizer, matcher or C fixture/compiler dependency is added. Compactness
removes repeated map/flag scratch and reuses traversal storage while preserving
kind2/3 accessor meaning and the source-shaped match representation.


## Source Boolean-matching stubs and final family gate

Both routines in giaIiff.c are explicit stubs: Gia_ManDeriveMatches returns0;
Gia_ManIiffTest returnsNULL. The Go implementations already preserve that
behavior. Add comments explaining the source limitation, not an invented
matcher. New Go-only tests cover nil arguments and populated pointer/vector/
paged-memory outputs, preserving identities and contents. The test entry point
accepts an absent input file with all flag values -1/0/1, returnsnil, preserves
the borrowed graph's size/mapping/packing and all four independent AND-not
truth assignments, and creates no files. No C fixture or compiler is needed.

All ten giaIff.c main routines, eight accessors and both giaIiff.c stubs are
reconciled. A final separate inspection confirms source enumeration, strict
comparison/order, NaN/empty bounds, traversal ownership, partial-cluster
selection, box timing and wrapper lifetime policies. Keep source clustering
heuristics and the source's unimplemented matcher; do not optimize them into
other algorithms. Earlier pending statements above are historical checkpoints.

Final frozen full Go-only module exits0:130 tested packages pass,3 commands have
no tests; both new stub tests execute. Affected full IF race77.317s from the
preceding frozen chunk remains valid (IF unchanged); new GIA stub race1.021s
and IF/GIA vet exit0. Logs:
/tmp/abcgo-iff-matrix-final-race.log,
/tmp/abcgo-iff-reviewed-final-{module.jsonl,stub-race.log,vet.log}.
All handles are terminal. This closes these two GIA source families; GIA/IF
package audits and the full131-package objective remain open,2/131 complete.
