# Native constrained SAT sweeping

All defined algorithm families in `src/proof/ssc` have native implementations.
SSC is combinational SAT sweeping under constraints, despite its location among
verification packages: care outputs are violations and must all be false.
Register state is treated as additional unconstrained combinational inputs; the
trailing CI/CO register interface is preserved. No temporal property is inferred.

The manager finds a care-valid SAT pivot, phases candidate functions at that
pivot, refines candidate classes using packed simulation of random care-valid
patterns, and proves equivalences with two-direction incremental SAT assumption
queries. SAT counterexamples split the affected class before retrying it. Only
UNSAT proofs permit substitution. UNKNOWN retains the original function; empty
care produces a zero-output implementation with the same interface. Final
representatives, proved flags, class links, and pivot phases are published on the
source GIA. The rebuilt graph uses a local structural hash and preserves identity.

The constraint wrapper extracts trailing constraint outputs separately from
ordinary outputs and next-state functions. Append reinserts constraints before
the trailing register inputs. As in C, Append on an unconstrained graph appends
one trivial false constraint. Containment dropping uses sequential SAT implication
queries; ring optimization minimizes each output under the others being false,
preserving the disjunction of outputs.

## Native interfaces and implementation boundaries

`Ssc_Pars_t` exposes `NWords`, `BTLimit`, `SatVarMax`, `CallsRecycle`, and Boolean
`Append`, `Verbose`, `Verify` fields. Defaults equal the C defaults. Start returns
a manager plus the SAT status instead of C's invalid-pointer sentinel. Simulation
and class helpers take explicit `Simulation` and `Classes` objects, replacing
mutable GIA-embedded simulation buffers and eliminating package-global random
state. Scalar bit primitives use slices and `math/bits`; simulation doubles each
input's word row when saving patterns beyond capacity.

Local recursive graph/CNF visitors consolidate source pointer-based frontier,
supergate, and MUX helpers. CNF uses exact local gate truth clauses and supports
native XOR/MUX/buffer nodes. It does not reproduce the source's supergate clause
compression. Solver recycling rebuilds the care cone and discards only learned
state. Source FRAIG scratch-copy resimulation is replaced by independent native
simulation of the original graph on the SAT witness. This proves the same
candidate obligation without borrowing other packages' mutable metadata.

Class refinement uses deterministic full normalized simulation keys rather than
source hash buckets followed by equality splitting. Initial simulation has a
32-round cap, after which SAT still checks every remaining candidate; counterexample
refinement is immediate rather than source batch scheduling. These choices may
change node choices, SAT counts, and resource-limit outcomes, while preserving
conditional semantics. The source-only declaration `Ssc_GenerateOneHot` has no
definition in this source family and is not invented here. Memory-release helpers
use Go ownership. There are no disabled-source build tags or placeholder engines.

## Validation and current dependency status

Unconditional tests check real reduction, SAT witnesses, packed-word expansion,
zero-input and empty care, phase-aware representative metadata, register ordering,
Append, containment, ring output union, recycling, and independent concurrent
managers. 150 random five-input AND/XOR networks are compared on all care-valid
assignments. The optional original-C oracle links actual unchanged repository
objects and checks 40 further constrained networks, four outputs each, on every
care-valid assignment. It explicitly skips if compilers/objects are unavailable.

The ordinary commands remain the target:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./proof/ssc
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./proof/ssc
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./proof/ssc
```

The earlier dependency blocker involving malformed GIA drafts has been resolved;
all three commands now pass against the complete package. `test_native.sh` uses
the GIA owner's maintained native-source list in a temporary module and links
all SSC files and real dependencies, with no mocks. Previously verified focused
commands:

```
abcgo/proof/ssc/test_native.sh
abcgo/proof/ssc/test_native.sh -race
SSC_CHECK_MODE=vet abcgo/proof/ssc/test_native.sh
```

Follow-up coverage checks packed Boolean operations, care-mask and witness
extraction, pivot-filled pattern transfer, phase-aware class splitting, and
complemented simulation equivalence. Every pair of fixture nodes is checked
under both polarities against all care-valid assignments, with SAT witnesses
independently replayed. The unchanged C oracle was rerun and passed; full
package race tests and vet also pass.
