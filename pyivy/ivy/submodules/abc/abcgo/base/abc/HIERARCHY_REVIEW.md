# Hierarchy and barrier-buffer review

## Readability / idiomatic pass

Ported all 14 original abcBarBuf.c functions (the initial inventory counted 13).
Added Design in hierarchy.go: named modules, ordered tops, model membership and
box-to-model references. Network maintains separate ordered terminal lists while
its object table keeps stable IDs and deletion holes. SetTerminalOrder validates
permutations before applying them; PI/PO slots map positionally across each box
instance, exactly as in ABC. CopyEdge retains complements alongside compatibility
Copy pointers. Interface-copy, net creation and module-copy helpers consolidate
ownership and avoid repeated container manipulation.

The complete API group covers hierarchical instance checks, topological boundary
collection, flattening to barrier latches, reconstruction of module boundaries,
barrier-friendly object order, conversion between barriers and transparent buffer
nodes, counting, and the roundtrip convenience routine. It retains ABC's
single-instance/combinational-model restriction rather than silently flattening
reused/recursive models. Modules not visited by the optimized root cone are
copied into the restored design, including extra tops, so library definitions
are preserved. No C pointer sentinels, container dependencies, hidden drafts or
stub algorithms were introduced.

## Separate bug / fidelity pass

Observed failing tests for two issues found during review:

* Missing optimized barrier buffers were rejected only after changing the
  source's copy mapping. Added preflight barrier counts, terminal shape, output
  arity and optimized graph checks before clone/copy mutation.
* Renaming a registered module allowed the same module identity to be added
  twice. AddModule now rejects duplicate identities independently of names.

Also verified and fixed preservation of barrier-count metadata through the
buffer roundtrip and preservation of all original design modules/tops. Explicit
optimized net names override reconstructed interface nets; otherwise original
module net names are kept. Source and destination graphs retain separate owned
objects and model references.

Tests cover one-level and nested hierarchy flatten/reconstruction, barrier-to-
buffer roundtrips and Boolean behavior for every input assignment, ordered model
PI mappings, independently reordered model/root PO mappings, stable IDs through
terminal permutation/deletion, unused primary inputs, complemented edges,
unused module/top preservation, shared/recursive-model rejection, and validation
before mutation. Full selected original/core tests pass with -race. Package-wide
build now advances beyond abcBarBuf.go to the remaining abcBlifMv.go draft;
this is not a full base/abc completion claim.
