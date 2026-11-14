# Source interpolation drivers

These drivers own dependencies on AIG graphs, mapped/simple CNF and the
original BSAT engines. The graph package cannot import its solver/CNF users.
The source entry points therefore live here rather than in `aig/aig`:

- `Aig_ManInter`: stored-clause SAT refutation and Inta interpolation.
- `Aig_ManInterFast`: incremental SAT checks for corresponding output pairs.
- `Aig_ManInterTest`: two mapped-CNF copies checked by solver2.
- `Aig_ManInterRepar`: the original experimental reparameterization driver.

Reparameterization currently fails the upstream solver2 assertion forbidding
clause ID zero. Its original C setup is tested by a compiled oracle; fixing
only clause IDs exposes a second failure (SAT where the driver asserts UNSAT).
The Go port preserves that behavior. Later source steps remain visible for
review but are not validated as a working algorithm. The former Go Shannon
projection was a substitute and has been removed.

Tests cover Craig conditions, relation-variable ordering, all two-input
predicate pairs, incremental constant/complemented outputs, source insertion
failures, solver2 self-comparison, and the original C reparameterization
failures. Passing them does not establish a working reparameterizer.
