# Separate design and compactness review

This is ABC's base/pla packed two-bit SOP manager, not misc/espresso's PLA_t
positional-set parser. The two packages share no representation or mutable global
variable geometry. Their tests run together to guard the package boundary.

Pla_Man_t stores dimensions and Cube slices. Input/output word widths, cube count
and divisor count derive from dimensions/slice lengths; the C cached nInWords,
nOutWords and vCubes natural-index vector are removed. CubeLits and Occurs are
materialized algorithm workspaces with explicit ConvertFromBits/ConvertToBits
boundaries, not hidden redundant counters. FULL input literals retain both
polarities when converted, permitting an exact packed roundtrip.

Parsing is line-oriented and reader-based, with errors instead of unchecked
pointer traversal. It supports all four source file types, comments, missing or
mismatched .p, .e/.end, alternate literal spellings, and retains ancillary header
directives. Writer output is deterministic, omits the C trailing NUL and returns
I/O errors. Public filename wrappers retain convenient C-style entry points.

Hashing retains source literal hashes and skip-signature matching while using
Go maps of bucket entries instead of overlaid packed array cells. It verifies
full cube equality after hash matches. Tiny covers no longer require the C hash
allocator's minimum-size assertion. Pairwise packed distance, polarity-flip
hash distance and skip-signature distance preserve their distinct source roles.

The experimental plaFxch.c is omitted from the C module.make, but its actual
single/double two-literal divisor extraction is implemented here. It recomputes
candidate costs from live cubes after extraction instead of maintaining the
source's fragile packed hash links and incremental queue. Definitions remain
ANDs; complementary occurrences implement two-cube OR factoring through De Morgan.
Go slices replace packed ownership transfers. A native fixture confirms the
selected divisor and final cubes; randomized tests check truth and literal cost.

All six source commands register directly on the committed base/main.Frame,
without importing unfinished base/cmd or GIA. |merge and |test deliberately retain
the original inactive bodies, documented with source locations. Both the actual
extraction and distance algorithms have separate working entry points. Lifecycle
cleanup clears the frame resource before reuse and is safe to repeat.

No C source is excluded behind build tags or ignored Go filenames. Original C
harnesses and input fixtures are in test_vectors, per the repository convention.
