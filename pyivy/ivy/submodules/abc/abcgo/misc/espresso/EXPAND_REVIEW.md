# Expansion and cover-helper review

Scope: shared cube setup/declarations (`espresso.h.go`, `globals.go`,
`cubestr.go`), expansion (`expand.go`), containment (`contain.go`), cover
manipulation (`cvrm.go`, `cvrm_output.go`), unate covers (`unate.go`), dense/sparse
adapter (`sminterf.go`), cube/PLA formatting (`cvrout.go`), PLA lifecycle
(`pla.go`), and high-level orchestration (`espresso.c.go`, `exact.go`).

## Separate readability, compactness, and idiomatic Go review

- Replaced C pointer-list layout with a named `cubeList` containing an owned
  cofactor mask and borrowed cube slices. Sorting/containment lists retain an
  explicit nil sentinel where existing algorithms need it; no unsafe pointer
  arithmetic is used to recover original row indices.
- Set-family reconstruction copies rows. The indexed containment helper locates
  borrowed slices by Go element identity and preserves original index metadata.
- Isolated single-output minimization callbacks in `cvrm_output.go`; core cover
  operations can be tested without depending on the entire minimization driver.
- Shared one formatter between `pc1`, `pc2`, and `print_cube`, replacing the C
  diagnostics' fixed 256-byte static buffers with strings.Builder.
- Kept the original expansion phases, traversal order, lookahead tie breaking,
  portable deterministic random-order fallback, and balanced exact-cover stack.
- Cube context remains explicitly process-local and nonreentrant, as documented
  on ConfigureCube; tests intentionally do not run these operations in parallel.
- Go collection ownership replaces C allocator/free loops; PLA lifecycle clears
  all owned references rather than walking and manually freeing linked nodes.

## Separate bug and boundary review

- Expansion's multivalued unravel threshold now checks division before
  multiplication, avoiding signed overflow before the 500-cube cutoff.
  General unravel checks expansion against 1,000,000 and total int32 capacity.
- Cube setup checks inconsistent dimensions, absent/zero part sizes, negative
  symbolic sizes (including int32 minimum), and packed-set width overflow.
- Found original-C exact-cover edge: in the dual representation, the empty
  intersection encodes a hitting set selecting the entire universe. Ordinary
  `unate_intersect` intentionally discards empty products, but using that rule
  for exact enumeration loses such valid covers. The internal exact path keeps
  empty intersections; public ordinary intersection retains source behavior.
  An empty row family now returns the mathematically correct empty hitting set.
- Found original-C all-primes edge: after essential lowering removes every
  blocker, find_all_primes copied RAISE without its remaining FREESET, returning
  nonprime cubes. The result now uses RAISE union FREESET. Exhaustive comparison
  for all 256 three-input truth tables caught and verifies this correction.
- Output headers guard the no-output case before indexing its part vector,
  correcting a source ordering that could index variable -1.
- Output formatting uses dynamically sized buffers; long cube strings do not
  overflow the original fixed diagnostic buffers.

## Validation

- 120 original-C expansion fixtures with binary and multiple-valued variables:
  exact output masks and row order match. Generator copies unchanged source
  algorithms and changes compatibility declarations only; see
  `test_vectors/generate_expand_oracle.py` and `expand_oracle.json`.
- 450 random independent binary truth tables: expansion preserves ON/OFF
  semantics and each output cube is independently proved prime.
- 600 random set families: both recursive unate complement and balanced exact
  enumeration match independently enumerated inclusion-minimal hitting sets.
- All 256 three-input truth tables: all_primes matches independently enumerated
  prime implicants, including constant functions.
- Mixed-valued unravel ordering, indexed containment metadata, distance-one
  merging, cube context save/restore, and PLA/cube formatting regressions.
- Focused combined set/geometry/sparse/minimum-cover/expansion helper build and
  tests pass. Race and vet validation run on that same real helper group.
- Full package validation still depends on remaining Espresso source families
  being ported; raw C fragments in cubehack.go/cvrin.go currently prevent the
  full package from parsing. High-level espresso/exact orchestration is real Go,
  but its complete integration tests await those families and related workers.

## Follow-up: consensus primes and PLA input

Added `primes.go` as the exact-minimization driver's real consensus dependency.
Its recursive cofactors, ceiling factoring, unate shortcut, bounded consensus
batches and containment merging are preserved. All 256 three-input truth tables
now check both consensus generation and expansion-based generation against the
same independent exhaustive prime enumeration.

Readability review: `cvrin.go` contains stream tokenization, product decoding and
directives; `cvrin_read.go` contains the read/normalize/complement/mapping driver
and summary output. `pla.go` owns lifecycle helpers formerly mixed into cvrin.
Shared buffered readers keep their buffer; unbuffered readers use a one-byte
adapter to prevent losing later PLAs through buffering across read_pla calls.
No fixed-size token or label buffers remain.

Separate bug review: validate counts and variable indices before allocation or
indexing; terminate symbolic directives cleanly on EOF; preserve negative
symbolic part sizes through parsing; bound pair variables; guard absent output
variables. A directive token no longer consumes its final newline: otherwise an
unknown directive without arguments causes skip_line to eat the next product
line. A red regression verifies that product is retained. Parser tests cover
binary F/D/R decoding, named inputs/outputs, phase, numeric and symbolic multiple
valued parts, pairing and symbolic metadata, sequential buffered PLAs, malformed
directives, round-trip output and diagnostics longer than the original 256-byte
buffer. Focused combined parsing/core tests pass; the high-level read driver is
awaiting integration with the real mapping/phase/pair families.

## Final integration: phase, pairing, symbolic mapping and command driver

All remaining I/O dependencies are real Go: `map.go`, `opo.go`, `pair.go`,
`hack.go`, `verify.go`, `main.c.go`, `main.h.go`, `util_old.go`.

Separate readability review:

- `Run(args,input,output) error` is the embeddable command driver; EspressoMain
  adapts process streams and returns status without calling os.Exit. Original
  active subcommands, strategy/range options and compatibility spellings call
  the actual algorithms. The command and debug tables use named Go structures.
- Run serializes its own invocations and restores the prior cube geometry,
  flags, timing state and filename, including after malformed input panics.
- No external SIS allocator/process APIs are fabricated. util_old.h was only
  declarations/platform macros; slices/GC, io/os and timing replace its runtime.
- Karnaugh `map` is named mapCover because map is a Go keyword. Gray-code cell
  indices are calculated from the original independent two-bit Gray groups,
  replacing its 256-entry literal table without changing coordinates.
- Phase recursion passes depth explicitly; exhaustive pairing retains source
  generation order and callbacks. Symbolic mapping preserves linked metadata,
  mixed-radix significance, optional labels and one-hot output generation.
- Optional algorithm diagnostics retain the source's process stdout; Run's
  output parameter receives the PLA result. This is documented on its API.

Separate bug review and regression fixes:

- Widening the output part across a machine-word boundary must not read a new
  wide mask against a shorter old cube. output_phase_setup copies old membership
  within bounds; a 32-bit-to-36-bit layout regression verifies both phases.
- Paired variable parts are inserted before all preexisting MV parts, matching
  the rebuilt cube schema. Source insertion immediately before only the output
  misplaced existing MV variables. A mixed-valued pairing regression verifies
  exact membership and dense/sparse metadata.
- Pairing rejects repeated/self endpoints before modifying covers. Exhaustive
  and greedy best-cost selection now accepts entirely negative edge costs;
  source initial -1 could leave no chosen pair or reused invalid endpoints.
- Output-phase dual intersections retain empty complements, so an output choice
  requiring every prime remains feasible (same dual-set edge as exact cover).
- Verification clones queries before cofactoring: cofactor deliberately excludes
  the identical source row for redundancy tests, so verify(f,f,D) otherwise
  falsely rejects its own shared rows. A dedicated regression checks identity.
- Don't-care variable deletion rebuilds the OFF-set at the new geometry instead
  of retaining obsolete old-width rows.
- Parser symbolic sizes, one-hot expansion, output ranges and exhaustive phase
  shift widths are checked before indexing or allocation. Unbuffered successive
  reads cannot lose a later document to the internal buffered reader.

Additional validation:

- 120 original-C phase/pair/double-output fixtures match exact ordered masks.
- 100 original-C binary PLA parsing fixtures match exact ON/DC/OFF row ordering.
- 60 original-C symbolic input/output fixtures match all represented minterms.
- Independent random phase truth checks, complete pair-generation counts through
  eight inputs, negative-cost matching, mixed-radix minterms and Gray-cell
  uniqueness regressions pass.
- Every one of 256 three-input functions passes actual Run ESPRESSO and exact
  minimization round trips. Exact cube counts independently match dynamic
  programming over all valid three-input product terms.
- CLI compatibility, repeated-document minimization, malformed flags/context
  restoration, named permutation verification, symbolic inputs/outputs, KISS
  merge and phase swapping all run against the actual integrated package.
- Full `go test ./misc/espresso` now passes; earlier parser-blocker notes above
  describe the intermediate state and are resolved. Final race/vet use the full
  package, without excluded files, build tags, stubs, or experiment directories.
