# NPN enumeration family review

Ported all active functions in dauNpn.c and dauNpn2.c. Source-derived files remain
and are real Go; shared parser, NP/formula and crossbar-enumeration code is split
into npn_formula.go, npn_transform.go and npn_enumeration.go. No alternate
canonicalizer or graph implementation was introduced: adaptive canonicalization
uses the authoritative DAU implementation, exact classes use lucky, and formula
AIGs use the existing aig.Graph through a neutral FormulaAIG adapter.

## Compactness, readability and memory review

Reviewed independently of correctness. Shared crossbar generation preserves
source offer order for both enumeration algorithms while avoiding repeated
truth-substitution statements. Dynamic result/frontier slices replace fixed
32-entry limits; per-round progress statistics retain cumulative source counts.
Dense permanent Dtt function marks and temporary collision bins become maps;
lookup iteration order never determines enumeration order. Transformation order
still follows the original Gray/adjacent-permutation schedules. The exhaustive
representative-table builder uses a bitset for marks, avoiding an unnecessary
byte per possible function. Optional five-variable full tables remain enormous
as in the source; no such table is constructed during package initialization.

Buffer lengths, source uint32 truncation, candidate flags and packed node/support
metadata remain explicit. Binary I/O uses the original little-endian 32/64-bit
layout. Writer errors propagate by panic in legacy void entry points. Native
EnumerateNPN/EnumerateFunctions APIs expose results without forced file output.
Formula/library input uses a bounded scanner rather than fixed C buffers and
releases files on every path.

## Bug review

Checked source NPN normalization, class marking, support-deficient classification,
shared vector scratch lifetimes, tied candidate ordering, delayed/multiplicity
levels, NP composition, output/input phase transformations, and formula truth.
Temporary orbit hash state is cleared after every collection. The two orbit
scratch vectors remain distinct during nested collection. Formula parsing
rejects trailing operators/tokens instead of silently accepting the source's
unconsumed suffix, and reports malformed/empty formulas without unsafe reads.
Generated formulas use Go strings, eliminating fixed 50/100-byte output buffers.
AIG construction reuses existing structural hashing and source input numbering.

No additional algorithmic discrepancy was found by the source comparisons.
Progress counts were reviewed separately: per-round steps, totals and fanin-pair
multiplicities are retained, including zero-new-class termination rounds.

## Original-C and independent evidence

* Original dauNpn.c plus real Extra schedules is compiled in a temporary oracle.
  Every entry of its 32,768-entry four-variable representative table matches.
  All network-enumeration frontiers match source total/partial class counts.
* All 256 three-variable truth functions match source symmetry/complement counts;
  independent minterm swapping checks random functions through six variables.
  Exhaustive table class counts for one through four variables are 2/4/14/222.
* Original ABC funenum C(f) outputs match every ordered truth and node/support
  metadata entry for four configurations, covering intermediate support,
  two-gate insertion and reduced final rounds with the real adaptive canonicalizer.
* Original L(f) library text matches byte-for-byte for three-variable 3/5-node
  and four-variable 3-node cases. Every emitted formula evaluates to its truth.
* Original delay/multiplicity runs match every frontier's fresh/total classes,
  cumulative composition steps and fanin-pair counts in four configurations.
* One hundred randomized nested formulas match the original dauNpn2.c parser;
  the graph self-test matches all 32 input assignments. Invalid formula, short
  binary table, duplicate-class, orbit reset and phase-involution tests pass.

These tests use temporary oracle directories, not a testdata directory. The
combined actual DAU header/count/hierarchy/adaptive/NPN source subset passes
normal tests, race tests and vet. Other DAU source families are being ported by
other owners and remain outside this review.
