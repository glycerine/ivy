# FRA combinational port

`fracomb` contains the shared combinational portions of ABC's
`src/proof/fra/fraCore.c` and `src/proof/fra/fraCec.c`:

- `Fra_FraigMiterStatus` and `Fra_FraigMiterAssertedOutput` inspect the miter
  using ABC's graph phase conventions. The status check uses each CI's stored
  `CioId` together with `nTruePis`; stale terminal metadata can affect it.
- `Fra_FraigSat` translates a combinational miter to CNF and runs ABC's SAT
  solver. It supports polarity transformation, OR or AND of outputs, resource
  limits, and the original solver's learned-clause settings.
  `cnf.Cnf_Derive` uses the source cut mapping and CNF writer, including the
  free-output encoding needed here. The solver simplifies at the root before
  bounded search, as in the C implementation.
- `Aig_ManCountXors` counts structural XORs; the original's MUX-type
  precondition excludes native EXOR nodes.
- `Fra_FraigCec` runs the initial SAT attempt, rewriting, FRAIG sweeps, and a
  final SAT attempt. The caller supplies the actual FRAIG sweep to keep this
  package independent of the manager package.

Partitioned and top-level CEC entry points remain in `proof/fra`; FRAIG
manager operations remain in `proof/framgr`. This package requires a
combinational graph for SAT and CEC. A caller must supply a non-nil sweep if
CEC proceeds past the initial SAT attempt. Each fallback rewrite, balance,
and sweep consumes its input graph; a sweep must return a new graph.

The source-named entry points retain ABC's witness behavior. When the initial
structural miter check reports failure, `Fra_FraigCec` replaces any model with
an all-zero assignment without replaying it. For a direct positive input this
assignment does not assert the output. `Fra_FraigSat` also returns raw SAT
variable values after polarity transformation without undoing nonzero input
phases. Callers requiring a replayable witness must account for these source
limitations. Unknown solver results stay unknown.

If the initial bounded SAT attempt is undecided and `Dar_ManRwsat` returns
NULL (for example, when the graph deadline has expired), ABC dereferences the
failed rewrite and crashes. The Go CEC entry point preserves the failure as a
panic instead of reporting an undecided result.

ABC's `fNewSolver` path passes clause ID zero into `sat_solver2_addclause`
while loading CNF (`src/sat/cnf/cnfMan.c:486–488`); assertion-enabled ABC
aborts because `src/sat/bsat/satSolver2.c:1294` requires a nonzero ID. The
source-named Go entry point likewise panics when this path loads its first
clause. This affects both OR and AND output modes.

For a graph with no outputs, the source OR-output path attempts to add an
empty clause and hits `sat_solver_addclause`'s nonempty-clause assertion. The
Go entry point preserves that panic; the empty AND-output path remains SAT.

Signed conflict and propagation limits pass through to the source-compatible
BSAT API. ABC can return SAT, UNSAT, or undecided for negative limits,
depending on previous counters and whether the formula is already resolved.
The ordinary BSAT limit API retains its separate nonnegative validation.

The package's full race suite covers the SAT solver, OR and AND
output modes, all two-input single-output truth functions under all input
phases, all pairs of two-input output functions, empty outputs, and witness
replay. It also checks miter status, direct-input CEC witnesses, and both XOR
representations. `proof/framgr` contains an integration test that forces the
initial CEC SAT attempt to exhaust its budget and exercises the real FRAIG
sweep; a second satisfiable fixture exercises repeated real sweeps and replays
the resulting witness against both the rewritten and original miter. A
separate XOR-heavy fixture forces balancing before the first sweep, and a
failed-sweep test checks the source's manager disposal and failure path. A
free-output CNF fixture checks the exact variable and clause order
against C. Independent direct comparisons with the original C implementation
also matched 100 generated all-free and 100 mixed asserted/free output graphs;
the source oracle cases use AND/inverter graphs, not Go's native XOR extension.
The current statement coverage is 93.9% (CEC: 95.7%, SAT: 87.2%). A deadline
regression checks that a per-sweep rewrite stops at the graph deadline while
`Dar_ManRewriteDefault` returns the partial result, matching C's void rewrite
wrapper. The direct and larger C-backed stage oracles exercise early and
repeated rewrite paths.

`Dar_ManRwsat` has an exact ordered-graph C regression for a two-XOR chain
and a reducible factoring example, with and without initial balancing where
applicable. The inputs are constructed in explicit steps. C's `Aig_Exor`
itself nests two AND calls in function arguments, whose evaluation order C does
not specify; Go evaluates those calls left to right. Consequently graphs
constructed independently with those helpers can have different object order
before rewriting. A larger explicitly ordered fixture has eight inputs, four
outputs and 72 initial AND nodes. Original C and Go agree on every node ID,
fanin literal and phase after each rwsat stage; the first rewrite reduces it to
66 nodes. These fixtures do not establish exact DAR rewrite equivalence for all
larger designs.

Remaining coverage gaps: the CEC integration tests do not exercise every
rewrite/sweep result. Resource exhaustion and learned-clause limits are not
exhaustively characterized, and no broad performance comparison with C has been
performed. Verbose output is not byte-for-byte identical to ABC's diagnostic
text.

The checked-in 24-MUX stage oracle compares every node and fanin with the
original C `Dar_ManRwsat` stages. Initial duplication, both balances, rewrite,
refactor, and final rewrite now match exactly. The two-XOR and larger ordered
fixtures also match their C stage oracles. These results establish parity for
the tested structures; they do not establish exact DAR rewrite equivalence for
all larger designs.
