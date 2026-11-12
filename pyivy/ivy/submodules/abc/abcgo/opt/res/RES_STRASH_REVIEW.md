# Window strashing review

## Readability and compactness pass

`Res_WndStrash` now constructs a current `base/abc.Network` and uses explicit
`abc.Edge` values throughout. It records temporary copies in one local map and
only mirrors them into `Object.CopyEdge` while calling the existing Hop-to-AIG
converter. A deferred restore prevents stale copy values from escaping the
call. TFO membership is collected into a local set from fanouts, avoiding
mutation of traversal IDs on the source graph.

## Independent bug pass

Reviewed complemented source fanins, center inversion, multiple root miter
pairs, roots that are absent from the copied window, fanin/divisor outputs,
foreign vector entries, missing Hop managers, nil window vectors, and cleanup
of temporary source copy state. The exact regression fixture checks every
assignment of a three-input window and verifies the miter, rewritten center,
and center fanin outputs.

## Dependency boundary

The implementation uses `base/abc` `Network`, `Edge`, `Abc_ConvertAigToAig`,
and `Abc_AigMiter` APIs. The package test currently cannot reach these tests
because `base/abc/abcHie.go` still has parser errors; once that dependency is
green, this file's focused regression should be run under race detection and
vet.
