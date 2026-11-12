# Auto, image and cascade-root review

Completed `extraBddAuto.c`, `extraBddImage.c` and the root API of
`extraBddCas.c`. Original C files remain visible and unchanged. The cascade root
forwards to the already authoritative, oracle-tested `cascade` package rather
than duplicating its search algorithms.

## Readability and compactness pass

Auto preserves the original cofactor translation-space recursion and paired
positive/negative recursions for signs, equations and matrices. Each operation
has its own memo key/domain; shared code removes refcount/cache scaffolding, not
mathematical branches. The fast path still interleaves original variables and
shift variables, simultaneously composes XORs, abstracts original variables,
and maps shifts back. Its general simultaneous-vector substitution is now the
single core `Cudd_bddVectorCompose` implementation, not an extrab-only duplicate.

Image computation retains source scheduling: quantify leaf-local variables,
choose the variable minimizing the sum of squared partition sizes, combine two
smallest participating partitions, and quantify variables whose remaining
occurrences are exactly that pair. Persistent tree nodes save quantification
cubes and referenced intermediate images for later care sets. Temporary support
metadata preserves the source's support overapproximation rather than silently
changing the schedule after Boolean cancellations. Monolithic image remains a
separate conjunction-plus-abstraction implementation.

Partition membership uses Go sets of partition IDs. It does not encode these IDs
as unrelated BDD variables. Canonical graph operations stay authoritative in core.
During construction image functions equal temporary partition functions, so the
same computed result is reused instead of performing an identical conjunction
and abstraction twice. Graph refs are owned only by persistent nodes/cubes.

Cascade wrappers preserve scalar counts, caller buffers, referenced cofactor/path
outputs, width limits and the original constant-node special case. Pointer-key
`st` node collections become Go maps. Global private encoding/cache helper state
is encapsulated by the existing cascade implementation, not exposed as unsafe
standalone functions with implicit process-global parameters.

## Bug review

Auto validates BDD ownership and matrix domains (even paired-literal indices
within manager range), support-space scratch capacity, and singleton-combination
counts. Complement phase is propagated through every BDD split. EXOR equation
outputs must each contain exactly one reduced-function variable; malformed or
repeated assignments fail and release partially referenced results. Valid outputs
retain the original referenced-result ownership contract.

The image tree validates projected retained variables and rejects a replacement
care set whose support exceeds the scheduling care support, preserving the old
image on rejection. The monolithic implementation deliberately allows broader
care support as in C. Images/cubes are referenced across manager reorder, deletion
is idempotent, and closed trees reject further reads/computation.

Source image scheduling encodes partition numbers as BDD indices and may access
beyond the variable array when partitions outnumber BDD variables. Go membership
sets remove that restriction. Cost squares use float64 before multiplication,
avoiding the C signed-int overflow; selection uses infinity instead of fixed
sentinels that could fail on large BDDs. Every leaf/intermediate graph is protected
by a persistent image reference. Metadata maps use stable variable indices, so
manager reordering does not invalidate support membership.

Profile validation now distinguishes BDD/ADD and ZDD levels. This extends the
source's nominal ZDD support safely instead of indexing its BDD permutation with
ZDD indices. Width traversal remains the single interval-based cascade routine.
Root path wrappers reject undersized output buffers before writing, and preserve
source references on both nodes and path sums.

## Evidence

- All 256×256 three-variable function translation spaces match direct
  `forall x: F(x)==G(x xor shift)` truth evaluation; fast and recursive paths agree.
- Original Auto checksum `4b8b81243ecf473a`: all function pairs, fast/sign paths,
  canonical variables, positive/negative equation extraction, and all 256
  three-variable equation matrices.
- Canonical reduction, EXOR gate ownership, equation reconstruction, invalid
  matrix/ownership/scratch domains, and shuffled BDD order have regressions.
- Image tests use 128 seeded partition sets (including more partitions than BDD
  variables), all 16 care sets, independent direct existential images, care
  support rejection, deletion, invalid inputs, and manager reorder.
- Original tree+monolithic image checksum `2c5b1f2faff964eb` matches 128 seeded
  source-valid schedules and all care sets.
- Root cascade tests cover both encoders, path reconstruction/references,
  constant false-path behavior, buffer validation, profile cutoff and node maps.
  Existing cascade independent/original-C tests remain required.
- Core vector composition tests exhaustive truth tables and simultaneous cyclic
  substitution, ownership and buffer/domain validation.

All root families are now ported and ordinary root race tests/vet pass; see
`FULL_PACKAGE_REVIEW.md` for the exact complete gates. Original sources stay
visible. Native generators are optional and live in `test_vectors`; normal Go
tests need no C compiler.
