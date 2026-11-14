# LPK port review

## Source comparison

Reviewed the Go routines against all translation units in `src/opt/lpk`:
`lpkAbcDec.c`, `lpkAbcDsd.c`, `lpkAbcMux.c`, `lpkAbcUtil.c`, `lpkCore.c`,
`lpkCut.c`, `lpkMan.c`, `lpkMap.c`, `lpkMulti.c`, `lpkMux.c`, and `lpkSets.c`.
The source-only `lpk_.c` contains no routines. Shared C `static` print helpers
are private Go helpers in their respective files. Traversal order, strict tie
selection, cut signature collision checks, packed truth-word order, support
bit numbering, and MUX/DSD split polarity were checked against the C bodies.

`Abc_NtkLutMinDecompose` now carries the temporary AIG fanin complement bit
when adding each copied edge; the C source passes complemented object handles, so
copying only `edge.Object` lost phase. A truth-reconstruction regression covers
XOR, whose AIG needs complemented edges.

`Lpk_MapPrime` now passes manager-owned reusable truth-graph scratch to
`Kit_TruthToGraph`, matching C's reuse of its `vCover` workspace while keeping
the Go kit API's slice representation. `Lpk_ManStop` releases that scratch.
No generator or experimental replacement algorithm was introduced.

## Tests

Existing focused tests cover LUT sizing, support/delay helpers, function
allocation and support minimization, cofactor supports, DSD set enumeration,
truth packing, cut signature/dominance, MUX split truth reconstruction,
recursive decomposition implementation, cleanup, wide truth decomposition,
and impact tracking. Added `test_vectors/cofactor_choice.json` to lock down the
C routine's first-variable tie behavior, plus function metadata/duplication,
manager workspace layout/cleanup, and graph scratch reuse tests. `Lpk_FunComputeCofSupps` also preserves the
source assertion against recomputing already-cached cofactor supports. Added
focused `Lpk_CreateCommonOrder` tests for cofactored-variable priority and
equal-score grouping, matching the C scoring loops and tie behavior.

## Verification status

`gofmt` and `git diff --check -- opt/lpk` pass. `go test ./opt/lpk` and focused
LPK tests currently stop before type checking because the imported
`base/abci` package has unrelated parse errors in `abcMap.go` (currently
reported from line 424 onward), including pseudo-C constructs. The exact line
frontier shifts as that package is edited; earlier blockers in `abcDar.go` and
`abcBidec.go` were repaired. No
workaround or edit to those unrelated files was made. Rerun focused and full
LPK tests after the `base/abci` source is syntactically valid.
