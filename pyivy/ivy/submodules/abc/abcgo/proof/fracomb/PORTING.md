# FRA combinational port

`fracomb` contains the shared combinational portions of ABC's
`src/proof/fra/fraCore.c` and `src/proof/fra/fraCec.c`:

- `Fra_FraigMiterStatus` and `Fra_FraigMiterAssertedOutput` inspect the miter
  using ABC's graph phase conventions.
- `Fra_FraigSat` translates a combinational miter to CNF and runs either ABC
  solver implementation. It supports polarity transformation, OR or AND of
  outputs, resource limits, and the original solver's learned-clause settings.
  `cnf.Cnf_Derive` uses the source cut mapping and CNF writer, including the
  free-output encoding needed here. Both solver paths simplify at the root
  before the bounded search, as in the C implementation.
- `Aig_ManCountXors` recognizes structural and native Go AIG XOR nodes.
- `Fra_FraigCec` runs the initial SAT attempt, rewriting, FRAIG sweeps, and a
  final SAT attempt. The caller supplies the actual FRAIG sweep to keep this
  package independent of the manager package.

Partitioned and top-level CEC entry points remain in `proof/fra`; FRAIG
manager operations remain in `proof/framgr`. This package requires a
combinational graph for SAT and CEC. A caller must supply a non-nil sweep if
CEC proceeds past the initial SAT attempt.

The source-named entry points retain ABC's witness behavior. When the initial
structural miter check reports failure, `Fra_FraigCec` replaces any model with
an all-zero assignment without replaying it. For a direct positive input this
assignment does not assert the output. `Fra_FraigSat` also returns raw SAT
variable values after polarity transformation without undoing nonzero input
phases. Callers requiring a replayable witness must account for these source
limitations. Unknown solver results stay unknown.

The package's full race suite covers both SAT solver variants, OR and AND
output modes, all two-input single-output truth functions under all input
phases, all pairs of two-input output functions, empty outputs, and witness
replay. It also checks miter status, direct-input CEC witnesses, and both XOR
representations. `proof/framgr` contains an integration test that forces the
initial CEC SAT attempt to exhaust its budget and exercises the real FRAIG
sweep; a second satisfiable fixture exercises repeated real sweeps and replays
the resulting witness against both the rewritten and original miter. A
free-output CNF fixture checks the exact variable and clause order
against C. Independent direct comparisons with the original C implementation
also matched 100 generated all-free and 100 mixed asserted/free output graphs;
the source oracle cases use AND/inverter graphs, not Go's native XOR extension.

Remaining coverage gaps: the CEC integration tests do not exercise every
rewrite/sweep result. Resource exhaustion and learned-clause limits are not
exhaustively characterized, and no broad performance comparison with C has been
performed. Verbose output is not byte-for-byte identical to ABC's diagnostic
text.
