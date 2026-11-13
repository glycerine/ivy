# SAIG porting progress

## Minimum-area retiming

`Saig_ManRetimeMinArea` now uses a vertex-capacitated maximum-flow/minimum-cut
implementation for the forward and backward passes, with the same PI/PO and
constant boundaries as `Nwk_ManRetimeCutForward` and
`Nwk_ManRetimeCutBackward` in `src/opt/nwk/nwkFlow.c`. The implementation is
in SAIG to avoid an import cycle: `opt/nwk` itself imports `aig/saig`.

The backward initialized-state path now hides incompatible registers, retries
an unsatisfiable cut using the initial-state SAT core, and restores the hidden
interface. The source's backward-only and no-initial-state paths are included;
the latter retains ABC's warning that its result need not preserve the initial
state. A duplication-order bug and an invalid heuristic-cut panic were fixed.
Constraint and assertion output metadata are preserved. The top-level pass
keeps hidden-register outputs private until their registers are restored,
including when asserted outputs overlap constraints. The separately exported
backward helper reorders hidden outputs around constraints; if assertions
overlap constraints, that helper conservatively skips hidden registers because
its intermediate public interface cannot represent both roles.

Validation includes exhaustive cut-subset comparisons on small graphs; cut
sizes compared against the existing NWK port on 200 cleaned graphs in both
directions; forward/backward register-reduction examples; hidden-register and
unsatisfiable-initial-state retry cases; and reachable-output comparisons for
all three-frame traces on 32 generated two-register circuits. Run the full
SAIG race suite and the dependent FRA race suite after changes.

The reproducible reference command `go run ./aig/saig/_reference -abc ../abc`
compares against ABC's network command `retime -M 3`: 22 of 24 deterministic
two-input, three-register generated examples produced defined initial latches
in C, and all 64 three-frame traces on those examples matched the original and
the Go result. The other two C results had unknown initial latches and were
excluded from that trace comparison. Four C results had fewer registers;
sequential cleanup of the original graph gives the same counts as those C
results. This network command is a useful reference but is not a direct call
to the original SAIG function.

Remaining review: compare cut tie-breaking and end-to-end behavior directly
with the original SAIG function on larger designs. Address the exported
backward helper's remaining overlapping asserted-constraint case if a
representation for noncontiguous output roles is added. The Go NWK converter
currently omits dangling nodes, whereas the C converter iterates all nodes;
the SAIG implementation retains the source's all-node behavior.
