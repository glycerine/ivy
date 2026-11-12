# Cut sweeping

All active algorithms in `src/opt/csw` are native Go. `Csw_Sweep` retains bounded ordered-cut enumeration, reference-cost eviction, signature and dominance filtering, truth stretching/support minimization, constant/single-leaf/two-leaf resubstitution, phase-aware cut hashing, recursive sweeping of replacement nodes, and dangling-node cleanup. The header helpers and diagnostic counters remain available under their original names.

The native graph uses packed `aig.Edge` values, with `aig.InvalidEdge` for an absent result; zero is the valid constant-one edge. Cuts use sorted leaf slices and normalized 32-bit truth words. Stable cut records share one leaves allocation and one truth allocation per node; a separate record link replaces C byte-stride traversal. Slice lengths replace redundant leaf/word/table counts and Go owns allocation lifetimes. Metadata grows when resubstitution adds new nodes, avoiding dependence on the source graph's original allocation bound.

Ordinary AND graphs follow the original algorithm. Native graph extensions preserve explicit XOR truth semantics and buffer barriers; topological DFS also accepts live nodes whose IDs are not in dependency order. Register, constraint, assertion, and barrier counts are preserved. These changes do not attempt sequential optimization.

## Validation

- `TestSweepOriginalC` compiles a small driver and links the repository's actual ABC objects when available. Sixty deterministic random 50-node AND networks, cut sizes 2–8 and leaf limits 2–6, match the original C's resulting node counts and all output truth tables exactly.
- Random five-input graphs, including explicit XOR gates, are checked on every input assignment with leaf limits 2, 3, 5, and 6.
- Opposite associative trees demonstrate actual duplicate removal. Sixteen-input trees exercise multiword cut truth, shared support, and buffer preservation over all 65,536 assignments.
- All pairs of nonempty six-variable supports exercise ordered union, dominance, truth placement, and rejection beyond the leaf limit.

Run from `abcgo`:

```
GOCACHE=/tmp/abcgo-cache go test ./opt/csw -count=1
GOCACHE=/tmp/abcgo-cache go test -race ./opt/csw -count=1
GOCACHE=/tmp/abcgo-cache go vet ./opt/csw
```

The original-C test skips explicitly if compilers or original object files are unavailable. The remaining semantic tests are unconditional.
