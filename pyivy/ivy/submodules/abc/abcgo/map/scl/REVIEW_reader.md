# SCL reader/model review

Scope: sclLib.go, sclLiberty.go, sclLibScl.go, sclLibUtil.go, formula.go,
genlib.go and liberty_write.go. Network sizing/buffering and command files still
contain original drafts and are outside this completed reader/model increment.
The full package therefore does not yet compile.

## Readability and compactness pass

Native slices replace embedded C vectors, a name map replaces name hash chains,
and one exported SC_Lib model is shared by readers, timing lookup and MIO bridges.
The generic Liberty tree retains unknown attributes, source positions and groups.
The model reader interprets combinational cells, ordered input/output pins,
Boolean formulas, templates, wire loads and conditional timing groups. Boolean
formulas are parsed once into expression nodes before truth enumeration.
SCL binary encoding uses one checked codec with the original version-8 layout.
No synthetic parser or library-success placeholder is used. GENLIB import retains
ABC's deliberately unit-delay fallback rather than pretending its linear timing
parameters are measured NLDM surfaces. Printing emits formulas and timing
statistics; the original optional Kit DSD pretty-printer is not reproduced.

## Independent bug/source review

Compared against sclLib.h, sclLiberty.c, sclLibScl.c and sclLibUtil.c. Corrected
class order to input-count/capacitance/name rather than area. Class identity follows
the original first-output truth table plus input/output count. Related conditional
timing tables merge by pointwise maximum when axes match; mismatched grids retain
the first table as in C. Missing rise/fall tables use their paired counterpart.
Quoted comment markers and punctuation remain literal. Numeric and table-shape
errors include useful context rather than producing malformed surfaces.

The original two-point load tangent selects the same first sample twice, giving
0/0. The Go implementation selects the second sample only for this two-point case;
a real MIO conversion regression exposed and now covers the NaN. Missing-backend
and invalid-delay paths return errors. Exported read APIs use errors instead of
C null pointers/assertions. Truth enumeration currently rejects more than20 inputs
explicitly rather than permitting unbounded allocations.

Integer tables preserve C's construction order: fixed-point copies are populated
before floating-point library unit normalization. The binary reader validates
bounds, version, formula/truth agreement, trailing bytes and count consistency.
Every prefix of a valid fixture is rejected as truncated without a panic.

## Validation

Selected reader/model file set passes race tests. Tests compile the unchanged
original C interpolation function and compare196 points including extrapolation.
A separate test invokes the original ABC executable on a three-function Liberty
fixture and compares the complete serialized SCL output byte-for-byte. These
oracle tests skip only if their external executable/compiler is unavailable.
Production Go code does not require either. Additional tests cover formula truth,
Liberty serialization roundtrip, SCL binary roundtrip/truncation, skip rules,
unit normalization, class rings and actual MIO simple/delay-based conversion.
This is focused validation, not a claim that the unfinished full package builds.
