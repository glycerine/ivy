# opt/sbd native port

The complete package remains the target. Native components below pass explicit
file tests, but ordinary `go test ./opt/sbd` still fails first in the active
`sbdCore.go` draft. Core, the alternate cut server, LUT/CEGIS synthesis,
SAT-topology synthesis, window SAT and simulation remain to be ported. No draft
has been hidden behind build tags or substituted with a placeholder.

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

## Verification checkpoint

Completed components pass race tests (4.722 seconds) and vet together:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo/opt/sbd test -race sbd.c.go sbd_topology_test.go sbdCnf.go sbdCnf_test.go network.go sbdPath.go sbdPath_test.go sbdCut.go sbdCut_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo/opt/sbd vet sbd.c.go sbd_topology_test.go sbdCnf.go sbdCnf_test.go network.go sbdPath.go sbdPath_test.go sbdCut.go sbdCut_test.go
```

Ordinary package verification has been rerun and currently stops at
`sbdCore.go:49`. All remaining original source files are still active. No
experiments were tested, no git commands were used, and no testdata directory
was created.
