# Separate design / compactness review

Core storage is ordinary Go slices, with stable pointers for clauses/reasons and
safe integer literal encoding. Garbage collection removes dead watches and
allocator ownership without pointer relocation. No packed pointer tags or
Glucose2 circuit state remain in this legacy solver.

Heap length, vector length/capacity, region size and variable count derive from
slices. Queue indices/logical count describe occupied storage rather than cached
capacity. Map count is a genuine aggregate across independent hash buckets.
Variable-elimination occurrence counts are a maintained cost cache; touched flags
represent pending work, and the redundant C touched-count counter is removed.
Eliminated clauses use reverse-played structured reconstruction records rather
than packed reverse-length arrays.

Common utility algorithms were reused only after the source comparison recorded
in SOURCE_COMPARISON.md. Legacy behavior is separately tested; the package has no
Glucose2 dependency. Standard-library reader/writer interfaces replace C streams,
Go errors replace process exits in parsers, and generic containers retain source
operations while releasing removed references for GC.

The ABC wrappers retain original default non-simplifying behavior; real
preprocessing is exposed by SimpSolver. The legacy no-op circuit compatibility
hooks and inactive preprocessing wrapper operations are documented with source
locations, rather than presented as implemented algorithms. AIG bridge work uses
committed aig.Graph, source SAT/SOP/quantification algorithms and a Frame adapter.
No original draft source is hidden behind build tags, ignored names or exclusions.

Validation: full-package tests and vet pass. Full race tests also pass together
with the imported aig/aig and base/main packages.
