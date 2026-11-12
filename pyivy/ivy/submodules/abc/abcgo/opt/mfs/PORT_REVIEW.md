# Native multi-fanout synthesis port

## Representation and algorithm scope

MFS operates on multi-input local functions, not just two-input AIG vertices.
`Network` therefore owns stable integer node IDs, ordered fanins, expanded uint32
truth functions, CI/CO order and a trailing register partition. Local AIGs retain
bi-decomposition cost. `FromAIG` and `ToAIG` bridge the committed AIG APIs. A
future base/abc adapter can preserve its object identities; this package does
not import that still-unbuildable application package.

The eight files in upstream `src/opt/mfs/module.make` are ported:

* `mfsCore`: parameter defaults; node/depth/arity scheduling; resubstitution,
  level-ordered don't-care BDC, area/redundancy/dry-run options and the source's
  three thresholded power passes.
* `mfsWin` / `mfsDiv`: fanout and level-bounded roots, TFI support, TFO exclusion,
  divisor expansion, source level/ID sorting, original-fanin tail and private
  MFFC reference closure.
* `mfsStrash`: original/target-flipped window construction, root observability,
  relevant external care clauses and target/divisor or fanin projections.
* `mfsInter` / `mfsSat`: paired independent window copies, opposite target
  values, divisor-XOR projection, one-hot input constraints, SAT care enumeration
  with a cumulative conflict budget and exact projected-function separation.
* `mfsResub`: original one- and two-divisor counterexample-cover selection,
  removal-first search, area-critical fanin ordering, power filters, replacement
  and dangling-cone deletion.
* `mfsMan`: owned per-window scratch, counters, levels and reverse requirements.

The obsolete duplicate `mfsCore_.c` / `mfsResub_.c` are excluded from upstream
module.make and define conflicting copies of the same APIs. Their Go files
explicitly record that provenance; algorithms live in the current implementation.
The upstream `mfs_.c` has no definitions. The optional `mfsGia.c` is also excluded
from upstream module.make; its native two-copy miter and query entry points share
the same checked projection solver instead of maintaining a second tagged graph.
No drafts have been moved, hidden or excluded using build constraints.

## Backend and interface differences

Satoko replaces the old BSAT/interpolation-manager coupling. Two independently
constrained copies must have opposite target values and equal chosen divisors;
UNSAT proves functional separation. A SAT model records which other divisors
separate the counterexample. Projected SAT enumeration derives the separating
truth function rather than emitting a resolution proof trace. It preserves every
reachable observable on/off minterm; unreachable projections complete to zero.
No timeout is accepted as a proof, care set or usable function.

The native BDC implementation minimizes local functions under the enumerated
care set. Power-aware local decomposition fills don't-cares with the more common
value when the source's `2p(1-p) >= .4` rule applies. Power sampling follows the source
48-frame/16-prefix, ten-word sequential protocol, including packed 16-bit
previous/current lanes for transition probability. Registers retain their actual
RO/RI partition and zero initial state. RNG state is local rather than shared
globally; heuristic structures need not be byte-identical to ABC.

Updates preserve the target's stable ID and delete newly unreferenced former
fanins, rather than allocating a new target object. More-effort processing
rebuilds its window after a successful update so it cannot reuse stale divisors.
A wholly dangling target conservatively uses itself as a root; the original root
routine asserts a nonempty observable cone. Dangling branches of an otherwise
observable target are ignored exactly as source.

The native truth-table network supports up to the MFS twelve-fanin optimization
limit. The source caps resubstitution at eight and don't-care BDC at twelve;
both pass filters are retained. Delay and GiaSat fields are retained for source
parameter compatibility; upstream active core does not select a separate delay
or GIA solver algorithm from these flags.

## Source-based validation

`test_vectors/mfs/oracle.c` includes the actual `mfsWin.c`, `mfsDiv.c` and ABC
comparison code. It constructs real ABC objects/vectors, then compares root and
divisor IDs/order on 250 generated graphs (discarding target cones outside the
source's observable-cone precondition). Cases vary fanout cutoff, TFO levels,
divisor level bound, duplicate fanins and window/divisor caps including zero.
ELF section collection removes unrelated application dependencies; no stand-in
algorithm replaces the routines being tested.

Independent exhaustive checks cover 210 random four-CI networks, across BDC,
default resubstitution and more-effort modes. Every CO is evaluated for every CI
assignment before and after optimization. Additional enumeration checks paired
SAT feasibility for hundreds of candidate projections and verifies each returned
interpolant against every observable care assignment. Dedicated tests cover
one-divisor gain and counterexamples, observability versus zero-level windows,
external care, one-hotness, power/area/redundancy modes, dry-run immutability,
register-preserving AIG roundtrips and invalid-cycle rejection.

## Separate design/readability pass

Reviewed ownership, hot paths, public domain types and source algorithm boundaries.
Typed IDs/truth words replace scratch pointer casts and global manager unions.
Window recursion memoizes both polarities and does not depend on node-ID order.
SAT clauses and projection indices are owned by the prepared window. Pruning
uses private reference counts; traversal never repurposes graph fanout storage.
Private helpers consolidate the one- and two-divisor proof/update operations.

## Separate bug pass

Original-C differential testing found an extra observability root: an output-
unreachable fanout branch was included in an otherwise observable window. The C
comparison failed before the fix; `TestWindowIgnoresDanglingFanoutBranches` now
locks down the correction. The source divisor comparator's level/ID tie break
and blocked-fanin expansion condition are retained and compared directly.

Additional checks focus on literal complement polarity, independent copy variable
offsets, projected minterm blocking clauses, duplicate fanins, empty projections,
UNSAT care and interpolation, replacement cone deletion and stale windows after
updates. Independent reviewer `dau_review` separately identified four mismatches:

* The BDC pass incorrectly inherited the resubstitution-only `WinMax` check.
  `TestBdcDoesNotApplyResubWindowSizeLimit` failed before the gate was restricted
  to resubstitution, and now passes.
* Power passes stopped after the first successful fanin instead of continuing.
  They now continue and rebuild projections after a mutation;
  `TestPowerPassContinuesAfterSuccessfulReplacement` requires four removals
  from a five-fanin constant rather than stopping one removal per pass.
* Probability sampling treated ROs as independent random PIs. The initial
  frozen-zero-register regression returned .4956 before the fix. The complete
  source 48/16 ten-word sequential protocol now returns the reachable zero/one
  probabilities, with actual state transfer and packed transition lanes.

* The AIG import adapter omitted dangling internal gates. It now imports every
  live AND/XOR/buffer cone before CO construction. The dedicated regression was
  observed red (zero imported nodes) and now preserves the dangling gate through
  a complete AIG roundtrip.

The independent readability pass also replaced quadratic insertion sorting with
stable O(V log V) level sorting. Invalid parameters now return an error from
`Optimize` before manager allocation; the negative-conflict-limit regression was
observed red as a panic and now passes. Direct checked manager allocation retains
its constructor panic contract. Dry runs no longer count unapplied mutations.
Packed counterexample-mask regressions confirm that the source all-one-padding
criterion means every witness must be separated, including across word boundaries.

Final normal tests pass in 1.602s, race in 3.765s, and vet is clean.
Independent reviewer `dau_review` also reran all three gates and signed off in
[INDEPENDENT_REVIEW.md](INDEPENDENT_REVIEW.md), with separate readability and
source-fidelity/bug passes. No findings remain open.

## Package gates

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -count=1 ./opt/mfs
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 ./opt/mfs
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/mfs
```
