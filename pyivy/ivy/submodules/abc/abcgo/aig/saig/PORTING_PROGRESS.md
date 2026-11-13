# SAIG porting progress

## Minimum-area retiming

`Saig_ManRetimeMinArea` is not yet a faithful port of ABC's algorithm in
`src/aig/saig/saigRetMin.c`. Its current forward step uses the existing
single-step retimer, and its backward step uses a local frontier heuristic
instead of `Nwk_ManDeriveRetimingCut` from `src/opt/nwk/nwkAig.c` and
`nwkFlow.c`. A candidate-cut coverage check prevents the heuristic from
constructing an invalid initial-state problem; it can consequently skip
retiming opportunities that ABC finds. This guard does not complete the port.

Remaining work:

- Port and validate the forward and backward maximum-flow/minimum-cut
  procedures, including ABC's PI/PO and constant-node boundary rules.
- Restore the backward initial-state retry loop, including bad-register
  hiding, unsatisfiable-core selection, and subsequent interface restoration.
- Restore the source's backward-only and no-initial-state control paths, and
  verify their original semantics explicitly.
- Compare retimed graphs with the C implementation and exhaustively validate
  their reachable outputs and reconstructed initial states on small graphs.
- Run the SAIG race suite and the dependent FRA SEC integration tests.
