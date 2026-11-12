# Complement family review

Ported compl.go completely: complement, simplify, simultaneous simp_comp,
special-case dispatch, split/merge, distance-one merging, all four lifting modes,
and single-cube De Morgan complement. The implementation retains the source
unate-recursive algorithm and shared cubeList/cover representation.

## Design, compactness, and readability pass

Reviewed separately from correctness. Repeated ceiling construction is shared;
source operation-specific special cases remain separate so each ownership and
result contract is visible. Recursive functions dispose list vectors after both
children exist. Merge reuses borrowed row slices only within the owned branch
covers, then copies rows into the result before releasing those covers. Wide
arithmetic prevents overflow in the source heuristic's count product. The
source's debug-only progress output is omitted. No minterm enumeration is used
in the implementation.

## Bug review

A red randomized regression exposed sorting the nil sentinel returned by
sf_list: d1_order panicked on an empty slice. Merge now sorts only actual rows;
standalone merge/lifting helpers also honor optional sentinel termination.
Directed tests cover empty sentinel lists.

The C complex-lifting helper allocates xlower once but frees it inside its
active-row loop, causing reuse/double-free on multiple active rows. Go allocates
independent lower-bound scratch per active row; a two-row regression verifies
both remain orthogonal to the onset. Other checks cover empty/universal covers,
singleton De Morgan complement, absent-column factoring, unate conversion,
binate branching, both default lifting choices, complex lifting, no lifting,
and borrowed input immutability.

## Original-C differential and checks

The checked-in fixture contains 180 deterministic multiple-valued cases. Its
reproducer, test_vectors/build_compl_oracle.py, compiles the unchanged original
src/misc/espresso C implementations into a temporary standalone oracle (only
ABC portability declarations/header glue are adapted). It records complement,
simplify, and both simp_comp results. Tests compare their Boolean functions at
every mixed-valued minterm, allowing equally valid cover ordering/heuristic
choices. An additional 250 random cases check all four returned functions
against independent minterm evaluation of the input and verify original row
bits remain unchanged. All lifting modes have direct semantic tests.

Combined actual converted-source irred/cofactor/complement/unate/contain/cvrm,
set/geometry and sparse/mincov subset passes tests, race tests, and vet. Full
Espresso package testing is still blocked by unrelated raw-C drafts, including
cubehack.go and cvrin.go; no files were hidden or excluded in place.
