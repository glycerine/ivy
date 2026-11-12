# Concrete mapped ABC network / SCL adapter

## Readability and compactness pass

Network itself satisfies the SCL timing/mutation/phase/clone interfaces. It
exposes existing stable object IDs and adjacency without a second topology.
Object.Data remains authoritative MIO functionality. NetworkMapping stages SCL
cell IDs and MIO gate choices; SclCommitMapping validates the entire staged gate
vector before applying any node changes. ConfigureSCL supplies explicit library
bindings; an existing mapped FunctionManager can supply lazy MIO bindings.

Clones own their objects and pin edges. The topological variant allocates CIs,
then topologically sorted logic, then COs and remaps assignments to new IDs.
Cloning leaves source Copy/scratch untouched. Object names, terminal order,
barriers, edge polarity, cell staging, and timing-derived CI level snapshots are
preserved. Library handles remain shared as the timing engine expects.

## Separate bug and fidelity pass

Tests use real mapped Network objects and the real SCL timing recomputation,
then stage a resized gate and verify that the actual Object.Data changes.
Invalid staged arity is rejected before touching any gate. Further regressions
cover actual adjacency after patch/create/delete, sparse IDs, topological clone
IDs and mapping/name remaps, and disabling phases while inversions remain.

A follow-up review found phases introduced through the ordinary graph API could
escape the optional phase flag. SclHasPhases now also detects real complemented
edges, with a regression. Patching a fanin preserves its phase; disabling phase
storage cannot silently discard unmaterialized inversions. External Object.Data
changes invalidate derived SCL IDs rather than leaving two function states.

Selected ABC core/function/SCL tests pass race tests and vet. Full ABC still has
untranslated source files, so consumers cannot import the whole package yet.
