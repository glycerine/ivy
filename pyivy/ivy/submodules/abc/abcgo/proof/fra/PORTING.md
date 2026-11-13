# FRA port status

The FRA package ports the active ABC `src/proof/fra` implementation, using
Go graph and SAT interfaces and the shared `proof/framgr` manager. Some C
helpers are folded into graph-scoped implementations rather than retained as
separate functions. Source comparison and tests are ongoing; a passing test
suite does not establish full equivalence to ABC on every circuit or option.

## Correctness and validation

- Bounded model checking is compared with exhaustive input traces on generated
  sequential graphs, with and without rewriting. Latch correspondence is
  likewise checked across generated graphs and multiple prefix settings.
- Induction and sequential equivalence have generated-circuit tests and
  counterexample replay tests. SEC replays witnesses on the original graph and
  reports an undecided result when a witness cannot be validated.
- Invariant verification only supports one induction frame. Unlike the C
  routine, it rejects other depths instead of claiming success without a check;
  malformed clause encodings are also rejected.
- Clause generation does not promote unchecked candidates to proved storage
  after an induction solver failure. Its legacy integer return is not itself a
  proof certificate.
- An induction refinement loop that makes no progress returns no result rather
  than the potentially unsound reduction allowed by the original C code.

Exhaustive tests cover small graphs and bounded traces, not all graph sizes,
parameter combinations, time limits, or interactions among ABC's algorithms.
The SAIG retiming implementation and its C differential test limitations are
described in `aig/saig/PORTING_PROGRESS.md`. Independent differential testing
against ABC for FRA's full range of options remains useful future work.

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
