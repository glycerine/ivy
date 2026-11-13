# FRA port status

The FRA package ports the active ABC `src/proof/fra` implementation, using
Go graph and SAT interfaces and the shared `proof/framgr` manager. Some C
helpers are folded into graph-scoped implementations rather than retained as
separate functions. Source comparison and tests are ongoing; a passing test
suite does not establish full equivalence to ABC on every circuit or option.

Timed `Fra_FraigLatchCorrespondence` uses the calling thread's CPU clock on
Linux and pins the goroutine for the timed run, matching ABC's `Abc_Clock`.
Other supported Unix targets use process CPU time; Windows uses elapsed CRT
clock-style time. Unsupported targets explicitly reject timed runs. The
timeout is checked inside the partition loop, so an empty partition list can
finish even after the nominal deadline, as in C.

## Correctness and validation

### Source comparison and coverage review

- Rechecked implication candidate ordering, cost-threshold selection, the
  zero-use-limit tie behavior, SAT clause polarity, missing frame mappings, and
  solver simplification against `src/proof/fra/fraImp.c`. The zero-limit path
  intentionally returns the maximum-cost tie group, matching the C loop's
  append-before-limit-check order.
- Rechecked initialized BMC frame construction and implication filtering
  against `src/proof/fra/fraBmc.c`. The Go path retains zero-initialized
  latches, frame-by-frame transfer, C's native EXOR-as-AND behavior, and the
  source's direct-input fast-path limitation.
- Added a regression to `fra_test.go` for mapped-child complement propagation
  and the missing second fanin of a buffer, corresponding to the child helpers
  in `fra.h`.
- The FRA package gates pass with `GOCACHE=/tmp/abcgo-gocache`: build, tests,
  vet, and race tests. The test suite reports 75.2% statement coverage.
  Coverage includes compiled-C source fixtures; optional full-source ABC
  differential modes remain opt-in because of their build cost.
- This pass found no verified implementation discrepancy in these reviewed
  paths. The review and package coverage do not establish equivalence across
  every FRA algorithm, graph shape, or parameter combination.

- Bounded model checking is compared with exhaustive input traces on generated
  sequential graphs, with and without rewriting. Latch correspondence is
  likewise checked across generated graphs and multiple prefix settings.
- Induction and sequential equivalence have generated-circuit tests and
  counterexample replay tests. Source-compatible `Fra_FraigSec` preserves
  ABC's status and unchecked witness behavior; callers needing replay may use
  `Fra_FraigSecVerified`, which returns undecided if no witness can be validated.
- Invariant verification only supports one induction frame. As in the C
  routine, other depths return success without any check; callers must not
  interpret that return as a proof. Malformed clause encodings are rejected
  when the supported depth is checked.
- Clause generation does not promote unchecked candidates to proved storage
  after an induction solver failure. Its legacy integer return is not itself a
  proof certificate.
- When induction refinement makes no progress, the source prints a warning,
  stops refining, and emits the current reduction. The Go port follows that
  behavior; the source warning says the result may not verify.

Exhaustive tests cover small graphs and bounded traces, not all graph sizes,
parameter combinations, time limits, or interactions among ABC's algorithms.
The SAIG retiming implementation and its C differential test limitations are
described in `aig/saig/PORTING_PROGRESS.md`. Independent differential testing
against ABC for FRA's full range of options remains useful future work.

Clause SAT calls pass signed conflict limits using ABC solver arithmetic; the
source accepts negative limits, whose behavior is not simply unlimited. The
checked-in `test_vectors/claus_limit_source.c` characterizes the undefined
result and assertion behavior against compiled C.

Compiled-C SEC option fixtures cover single-output, separate and merged interpolation,
disabled interpolation, and signed interpolation conflict budgets. In the
fixture, separate interpolation preserves ABC's frame-zero counterexample,
while merged interpolation returns a frame-one counterexample; a negative
budget returns undecided. The shared INT solver passes the signed budget to
BSAT unchanged, as the C implementation does. The single-output regression also replays its
counterexample on the original graph after cleanup removes a register.

On the source SEC sequential-simulation failure path, the function returns
before assigning its optional result graph pointer, leaving the caller's
previous value intact. A negative global conflict limit can instead stop the
induction stage with an undecided result graph; a zero frame bound skips it.
The checked-in `test_vectors/sec_options_source.c` fixture characterizes these
boundary cases against C.

The SEC `TimeLimit` option is currently ineffective in the active ABC C
implementation: its global deadline checks are commented out. This port
preserves that behavior. SEC passes its separate `nPdrTimeout` to PDR.

An optional differential test runs the ABC executable's network `ssweep`
wrapper on generated AND/inverter graphs and compares bounded output traces
with the Go induction result. From `abcgo`, run:

```sh
ABC_REFERENCE="$(pwd)/../abc" go test ./proof/fra -run TestABCInductionDifferential -v
```

The test exercises induction depth, partitioning, latch correspondence,
rewriting, prefix, level-limit, and one-hot options. The wrapper may clean or
reorder state, so the test checks behavior rather than internal structure. It
cannot establish that the selected binary was built from this checkout.

ABC's FRA routines do not consistently preserve native EXOR nodes. In
particular, induction frame construction and extension and latch partition
construction call `Aig_And` for every internal node, including EXOR nodes.
The Go ports follow those source operations; do not assume native EXOR retains
XOR semantics through these paths. The ABC differential above uses ordinary
AND/inverter graphs and does not cover native EXOR.

`Fra_SmlSelectMaxCost` preserves an ABC edge case: with a zero selection limit,
it still returns all candidates tied at the maximum cost because the C loop
checks the limit after appending. Normal implication derivation requires a
positive limit, but direct callers can observe this behavior.
# Partitioned AIG and timeframe entry points

Partitioned FRAIG and choicing use `fra.Aig_ManFraigPartitioned` and
`fra.Aig_ManChoicePartitioned`; both pass the original `Fra_FraigChoice`
backend and its conflict and level limits to the AIG partition code. The
lower-level AIG package cannot import FRA without an import cycle, so its
source-named and `WithFraig` variants require an explicit callback at
compile time. `fra.Saig_ManTimeframeSimplify` similarly
supplies `Fra_FraigEquivence` to the SAIG timeframe helper. No alternative
equivalence algorithm is selected when a backend is missing.


The partition bridge matrix has an independent checked-in-C build mode. From
`abcgo`, run:

```sh
FRA_BUILD_ORACLE=1 go test -race ./proof/fra -run TestPartitionMatrixAgainstCompiledABC -count=1 -v
```

This mode requires GCC, G++, make, ar, zlib headers and libraries, and the
full checked-in ABC source tree. It copies the sources to a temporary build,
excludes existing `.o` and `.d` files, and builds the archive there. The test
skips by default because compiling ABC is comparatively expensive. The older
`TestPartitionFraigAgainstCompiledABC` uses checkout object files and does not
independently establish their freshness.
