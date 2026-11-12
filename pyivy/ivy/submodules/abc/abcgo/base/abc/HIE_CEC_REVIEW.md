# Hierarchical Boolean graph derivation

## Readability and compactness pass

Ported all active abcHieCec.c routines. Direct expansion and bottom-up cached
module derivation remain distinct real algorithms. Each instantiation owns its
net-signal environment, so multiply used models do not alias intermediate
signals. Typed aig.Graph/Edge replace the original GIA manager/literal integers,
consistent with the other exported ABC graph conversion helpers. Cached import
handles explicit outputs, buffers, XORs and complemented edges.

Module dependency collection, instance/node counting and recursive-definition
checks use local caches instead of overloading pointer-sized Data slots. Source
Step/DTemp scratch results remain available, and caller-owned Network.Data is
not destroyed. Flat-network count wrappers retain the original special cases.

## Separate bug and fidelity pass

Tests compare direct/cached derivation, reused models and wire aliases against
complete truth tables. All 256 combinations of eight distinct boundary phase
locations are checked on every two-input assignment, including flattening via
abcHie and importing cached child graphs. This exposed missing phase transfers
in the earlier general hierarchy copier; typed unary bridges now preserve
formal/actual input and output inversions. Nonbinary complements are rejected.

Indirect recursion and combinational box cycles fail explicitly rather than
recursing forever. Models supplied out of dependency order still derive through
actual cached dependencies. The returned graph is the requested root even if
it is not the last supplied model. Terminal ordering/permutation and SAT-model
snapshot APIs requested by RET/FRET consumers have regressions. Binary symbolic
MV domain storage requested by the IO owner has an owned validated setter.

## Reader integration

HierarchyReader(filename,multiValued,check) is installed through
SetHierarchyReader by the actual IO package, avoiding an import cycle. The
filename wrapper selects binary parsing with checking and invokes the real
native cached hierarchy derivation. The old private test-toggle alternative
Au manager representation is a separate remaining original source file; no
unported parser is substituted. Tests use a supplied in-memory reader solely
to exercise the service contract. Without a reader the API returns an explicit
error. The actual base/io owner is implementing/registering this dependency;
filename integration is not claimed complete until that service lands.

Selected complete ported ABC scope passes race tests and vet. Full package
frontier is now abcHieGia.go; all remaining originals stay visible.
