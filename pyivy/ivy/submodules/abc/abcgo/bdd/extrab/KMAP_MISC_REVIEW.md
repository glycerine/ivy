# K-map and miscellaneous algorithm review

Original `extraBddKmap.c` and `extraBddMisc.c` remain unchanged and visible.

## Readability and compactness review

The K-map renderer shares one ASCII geometry/Gray-order implementation for the
ordinary and relation entrypoints. The Windows console color code is replaced
by portable writer output. Variable-selection and validation are separate from
rendering; the diagnostic and map both use the supplied writer.

Miscellaneous transfer/permutation/polarity APIs delegate to authoritative core
or cascade operations. The fused cross-manager AND retains source simultaneous
transfer/conjunction recursion, rather than transferring the entire second
operand first. Maps replace pointer-key tables and typed slices replace output
pointer arithmetic. A single traversal writes PLA rows for both the writer and
file APIs; errors are returned instead of discarded.

Cube counting extends the existing core Minato recursion with count/limit
metadata instead of duplicating ISOP or constructing a temporary ZDD cover.
Simultaneous vector substitution is shared with Auto in core. Random-set
sampling has local RNG state, with deterministic RNG injection in tests.

## Bug review

K-map variable domains reject duplicate/nonprojection/foreign variables,
incomplete support and short name arrays. Constant diagnostics honor the writer.
Custom-order truth diagnostics use selected variables, fixing the C code's
implicit first-variable assumption. Negative Gray decoding is rejected, avoiding
an infinite signed-shift loop. Relation dimension addition is bounded before
allocation by the 20-variable public limit.

Misc APIs validate ownership/mappings/caller buffer lengths before mutation.
Reference-bearing permutation outputs are built temporarily so failure cannot
leave partially written caller output. Fused transfer prevalidates every mapped
index before growing destination projections. Complement phases propagate at
each cofactor split. Generic support handles BDD/ADD/ZDD domains and levels.

The old global ISOP phase-count cache could reuse stale counts on repeated
calls. Counts now belong to one recursion; constants respect the cube budget,
phase ties prefer positive, nil vector entries preserve guide entries, and count
addition is overflow checked. The one-million-new-BDD-node budget is enforced by
the allocator, with rollback and restoration of prior nested limits on failure.

Random sampling rejects NaN/out-of-range density, impossible family sizes and
degenerate distributions requesting multiple distinct members, preventing
source nontermination in those cases. Extremely skewed valid distributions can
still take many rejection samples. PLA validates all paired-literal cubes before
opening/truncating a file, rejects contradictions and out-of-range variables,
and propagates writer/close failures. C demo routines remain callable real demos;
no test invokes their large random benchmark. StopManager uses Go ownership/GC
rather than pretending C zero-reference memory diagnostics apply unchanged.

## Evidence

Normal tests include all three-variable transforms/polarities and cross-manager
function pairs, references/buffers/support/tuples, repeated ISOP calls and strict
limits, random-family cardinality, malformed PLA covers and writer errors.
Permanent original-C fixtures: K-map text `aa69523ca10d5418`, miscellaneous
transforms `869b30057848fe11`, ISOP count/phase guide `8e62023c841daec4`, and PLA
text `3889814f756ac6b6`. Native generators live in `test_vectors`; Go tests require
no C compiler. See `FULL_PACKAGE_REVIEW.md` for ordinary package gates.
