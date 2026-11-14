# Algebraic factoring and support-minimization review

## Pass 1 — representation/readability

All 22 source helper names are implemented using Graph/Edge and native typed
slices. Cone frontiers are explicit collections; recursive reconstruction copies
nodes before creating gates so slice growth cannot invalidate a node pointer.
The public traversal helpers preserve ABC's current/previous generation protocol.
The complete support cut is bounded to sixteen variables, as in the source.
Local boolean truth tables replace Kit and untyped scratch-pointer storage;
reconstruction uses the graph's typed object-copy map. Cofactor Edge literals represent
assignments forced true, so a complemented literal forces its node to zero.
Named status constants retain the original 0/-1/-2/-3 outcomes. The two original
example helpers return their constructed graph/result/status for inspection
instead of spawning a GUI viewer or destroying the graph immediately.

## Pass 2 — correctness

Separate tests compare both original source examples exhaustively on their care
sets. Algebraic factoring is checked under both assigned polarities, including
negative implications and constant conditions. Native XOR reconstruction and
complemented function roots preserve cared behavior. Randomly generated cones
are exhaustively evaluated after each successful reduction and Graph.Check
validates every reconstructed graph.

Review fixes include:

- Accepted cofactors are checked progressively against the updated function on
  the care set. Individually legal choices can otherwise combine into an invalid
  assignment; a direct helper regression demonstrates that case.
- Empty highlighted cones and non-node functions return explicit nonapplication
  statuses instead of dereferencing missing simulation results.
- Constant nodes and duplicate cut boundaries are rejected as independent truth
  variables; source assumptions are explicit rather than silently corrupting
  the truth-table mapping.
- Support-membership checks retain the public traversal-mark side effects.
- Traversal generation rollover preserves the generation becoming Previous.
  Clearing all marks between a seed step and cone walk would lose the seeds.
  Regressions cross the boundary during overlap and highlighted-cone collection.

The converted-file AIG test suite including utility and factoring tests passes
with the race detector. Full-package compilation remains dependent on other
workers' outstanding files; no draft sources were hidden or disabled.

## Sequential audit checkpoint

The preceding report is historical, not completion evidence for the current
package audit. New source comparison found that support collection ignored
already-current frontier marks and reconstruction failed to publish C's
object-copy side effects. Both regressions were observed failing before their
repairs. Support collection now uses traversal marks directly, eliminating its
extra membership map. Reconstruction uses the graph copy map and preserves
unrelated entries. A compiled original-C frontier fixture verifies support
order and resulting marks. See `../../SEQUENTIAL_AUDIT.md` for current scope
and verification; the remaining factoring helpers still require review.
