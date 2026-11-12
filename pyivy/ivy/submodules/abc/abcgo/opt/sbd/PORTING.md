# opt/sbd native port

The full package is native Go, including the main optimizer, incremental and
alternate cuts, window SAT, simulation, LUT/CEGIS synthesis and fixed-topology
synthesis. All original product files participate in ordinary package builds.
The optimizer uses `aig.Graph`, Kit truth/ISOP algorithms, Dau decomposition,
native Vec/util helpers and the public BSAT solver; it does not import Gia or
CNF packages.

## Graph and mapping boundary

`network.go` owns typed SBD metadata around the existing authoritative
`aig.Graph`: LUT root/support/truth, mapped levels, references, mirror literals,
and box-input CO links. A distinct BoxLevelInputs list filters disabled timing
arcs during level computation; path backtracing retains all structural box
inputs exactly as the source does. The Gia owner added matching neutral
BoxInputObjects/BoxLevelInputObjects and LUT-reference accessors. mf.Mapping's
LUT interface uses the same owned support/truth shape.

ABC/Gia literal zero is false, while native Graph regular constant zero is true.
Boundary conversion changes only constant polarity, preserving original integer
node IDs and SBD's internal literal/mirror convention (-1 means no mirror).
Mapping setters copy caller slices. Default binary cuts retain complemented
fanin truth and explicit XOR functions. Box setters validate CO/CI kinds and
separate the enabled mask from the structural inputs.

## Completed components and reviews

### CNF (`sbdCnf.go`)

Ported exact dual-polarity Kit ISOP cover-to-CNF encoding, printing, and formal
fanin-literal/pivot-polarity translation. Clauses retain byte literals and 255
separators; translated vectors retain the source trailing empty level.

Readability review: separated printing from the stdout wrapper and made the
output-equivalence encoding and separator convention explicit. The ISOP cover
is caller-visible native Vec storage rather than reinterpreted C pointers.

Bug review: a red exhaustive test exposed the original constant-truth shortcut
binding variable zero even when the declared support was nonempty. Constant
clauses now constrain the actual output variable nVars. Zero-input constants
retain their source encoding. Dimension bounds prevent out-of-byte literal
indices and unsupported truth widths.

Validation: all 65,536 four-input functions, including constant functions with
nonempty declared support, are checked over every input/output assignment.
Clause counts, signed literal translation and text printing have regressions.

### Topology enumeration (`sbd.c.go`)

Ported level-composition enumeration, stable interning/insertion order,
parameter counting and the original diagnostic. Owned slices/maps replace raw
vector-hash lifetimes.

Readability review: factored numeric state keys and exposed the finite
level-expansion helper to tests. Kept source capacity rules and traversal order.

Bug review: empty configurations return zero instead of indexing an empty
vector. Negative levels/dimensions are rejected; parameter arithmetic uses
wider intermediates and checks int32 results.

Validation: all level profiles through five binary nodes match a separate
positive-composition generator, with capacity and parameter-count checks.

### Critical paths (`sbdPath.go`, `network.go`)

Ported mapped LUT levels, one-level-slack critical traversal, timing-box
backtracing, source internal-path pass and delay diagnostics. Uses the typed
mapping adapter rather than flattening LUTs into binary AIG levels.

Readability review: separate traversal state replaces global graph marks;
metadata ownership is explicit. Box structural and enabled-timing inputs remain
distinct, following Gia_ManLutLevelWithBoxes and Sbc_ManCriticalPath_rec.

Bug review: validates node kinds, mapping references, box arity and cyclic
mapped timing. Constant literal conversion is explicit rather than assuming the
two graph packages share polarity. Default truth tables include fanin phases.

Validation: deep versus one-level-short branches, mapped support replacing a
binary cone, traversal across a timing box, disabled timing arcs, constant
roundtrips, mirror sentinels and metadata slice ownership have regressions.

### Incremental cuts (`sbdCut.go`)

Ported sorted cut merge, containment/signature filters, both cost comparisons,
dominance insertion, normalized truth interning/support reduction up to ten
variables, incremental delay/spec computation, references/mirrors and best-cut
selection. Native cut slices replace packed C records; truth words remain exact.
The source Sbd_StoDerefObj unconditionally returns before its disabled code; its
native empty body is documented source behavior, not a substitute algorithm.

Readability review: shared merge/containment helpers remove duplicate pointer
walks while retaining source sort ties, dominance and reserved unit-cut capacity.
Native stable truth keys and owned word slices replace raw memory/hash handles.

Bug review: temporary truth and cut storage has owned lifetimes across slice
appends; fanin complements include mirror phases. Cut dimensions are checked,
source packed bitfield counters retain full int32 values, and Free is idempotent.

Validation: all pairs of six-element subsets and capacities are checked against
independent set unions and containments. Random reconvergent graphs are tested
at cut sizes 4/6/8/10: every stored cut truth agrees with direct graph evaluation
for every primary-input assignment, with mapped delay and reference checks.

### Alternate cuts, simulation and window SAT

The alternate cut server retains the four ordered top-reduction attempts and
lower-level retry. Simulation uses owned packed words and a reproducible random
source per network. Window SAT retains the original dual-network observability
construction, core-based truth/SOP extraction and replacement relation.

Readability review: explicit owned slices replace graph scratch marks and raw
pointer lifetimes. Original experimental backward simulation is labeled as such
and is not used as a rewrite proof. Solver inputs and source literal polarity
are documented at their boundaries.

Bug review: solver phase randomization now visits every variable rather than
skipping indices through the source shift. Constant SAT variables are explicitly
constrained. Native buffers receive transparent equations, observation roots,
reference counts and critical-path traversal, preventing an unsound rewrite
when an otherwise observed branch passes through a buffer.

Validation: alternate-cut feasibility, mirrors and topology; 192-pattern packed
simulation against direct graph evaluation; justification and complement
classes; SAT truth, missing-divisor counterexamples, observability, replacement
relations and care models all have behavioral regressions.

### Structure and fixed-topology synthesis

Ported LUT/selector CNF, parameter decoding, structure printing, the full
counterexample-guided synthesis loop, topology constraints, delay constraints,
random connected topology generation and fixed-topology function synthesis.

Readability review: typed topology/delay records and explicit structure input
indices replace address arithmetic. Parameter layouts and the root-last
feed-forward order are stated at the API boundary.

Bug review: the source standalone experimental setup overlapped output and
parameter variables, used parameters as data inputs, omitted selector coverage
and inverted substitution polarity. Native constraints implement the intended
feed-forward cascade with distinct outputs and active-low selectors. The delay
loop constrains every forbidden delay instead of repeating the first one.
Truth allocation depends on actual inputs, avoiding the source maximum-size
allocation. Random topology generation handles the exhausted-frontier boundary.

Validation: all two-input LUT functions, selector one-hot constraints, actual
CEGIS success/failure, all two-input fixed-topology functions, positive AND3 and
negative majority3 synthesis, multiword AND8 and connected random topologies
have independent evaluation checks.

### Main optimization and mapping derivation

Ported windows/TFO collection, divisor sorting, simulation columns, dominance
and cover search, support pruning, one/two/three-LUT exploration, truth-based
implementation, mirror/reference updates and graph/mapping derivation. Timing
box structural links and enabled delay arcs remain separate. LUT targets 2..6
are supported; support cuts remain bounded at ten variables.

Readability review: split the original core into manager/window work,
cover/exploration and graph derivation files. The authoritative graph and local
mapping arrays retain source object identities until final derivation. Native
ownership and public solver calls replace unsafe cross-package field access.

Bug review: selected cover-column indices are translated back to SAT variables;
cover cardinality follows the requested LUT size. Support queries beyond six
variables use exact SOP/core existence proofs instead of indexing six-variable
masks. Structural-hash self mirrors and cyclic or nonimproving replacements are
rejected. Mirror phases feed simulation consistently. SAT conflict limits also
apply to constant checks. Mapping setters reject duplicate/self support and
nonlogic roots. These fixes have targeted regressions.

Validation: exhaustive four-column minimum covers plus larger heuristic covers;
random full-optimizer equivalence with mapping, extra cuts, divisors and critical
path options; internal don't-care elimination; actual two-LUT, three-parallel
LUT and three-cascade LUT resynthesis; five/six-input targets; timing-box
remapping; buffer observability and structural-hash self-cycle rejection.
Resynthesis output is checked exhaustively through ten primary inputs.

## Verification

Run from the repository root:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./opt/sbd
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -timeout 120s ./opt/sbd
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./opt/sbd
```

The public `bsat.SetPolarity` bridge is separately covered by API tests and the
whole BSAT package race/vet gates. No experiments are included in these checks.
