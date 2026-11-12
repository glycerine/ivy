# DAU enumeration, divisor diagnostics and graph conversion

Ported every active function in dauEnum.c, dauDivs.c and dauGia.c, preserving
source-derived filenames. Conversion uses the authoritative aig.Graph and
bool/kit truth factoring. Dsm_ManDeriveGia accepts the live mf.Network+LUT
contract also implemented by mf.Mapping; no second mapped-network model is
introduced. Graph results use native aig.Edge types, including its constant-one
node convention, with explicit constant translation at Kit's GIA-literal boundary.

## Compactness/readability pass

Reviewed separately from correctness. DSD conversion reuses DAU's parsed dsExpr
representation rather than maintaining another parser. A shared evaluator has
explicit balanced/folded construction options; the original recursive cursor
entry points remain adapters. Prime nodes use the real Kit factorer and reuse
caller workspace. Sorting fanins by existing graph levels preserves source tie
order and balanced pairing. Existing native graph operations own structural
hashing, levels and cleanup.

Enumeration uses Go strings instead of three static 64/256-byte buffers; shared
operand formatting handles star/complement stripping and variable lifting.
Typed divisor rows expose diagnostic counts without inventing an algorithmic
return that the source does not provide. Statistics use atomic counters and
reset together with report output. No new manual allocator or redundant graph
metadata is retained.

## Bug review

Audited constant convention translation, input complements, repeated fanins,
fold/balance order, MUX substitution scope, prime width/argument agreement,
Shannon cofactors, residual-prime factoring, temporary truth ownership, cursor
position, buffers and sequential boundaries. Prime fanins retain their source
order. Mapped buffers are explicitly created rather than silently aliased away;
register counts and optional name/spec provider metadata are preserved. Source
cleanup runs through the authoritative graph manager.

The existing graph's native XOR representation is used when requested. Its MUX
operation represents the exact MUX as AND/complement edges; there is no separate
native MUX kind in this authoritative graph. This changes representation for
the source's optional real-MUX storage mode, while preserving the exact Boolean
function and graph API. Default source AIG mode matches truth, object count and
root level in the original-C comparisons below.

Empty/malformed expressions and out-of-range leaves/prime widths reject with
explicit failures instead of reading beyond C arrays. Enumeration preserves the
source complement-pruning decisions and ordered partitions. The divisor API is
honestly diagnostic: its source NULL result is retained; the extracted count
API returns useful owned rows. The source test000 early return is preserved.

## Original-C and independent evidence

* Original dauEnum.c is compiled with only the diagnostic maximum-variable
  parameter changed from ten to five. Its entire ordered five-variable dump
  matches the port byte-for-byte. Original dauDivs.c diagnostic output also
  matches byte-for-byte for a six-variable truth table.
* A temporary driver links the original existing ABC object files and invokes
  Dau_DsdToGia. Ten expressions, including constants, complemented expressions,
  balanced n-ary operators, nested MUX substitution and prime truth tables match
  original truth, root level and object count. This optional test skips when
  the original built objects/toolchain are unavailable.
* Every assignment of eleven representative DSD expressions is compared with
  the independent DAU truth evaluator in folded, balanced and native-XOR modes.
  A six-input AND verifies balanced level three versus folded level five.
* Fifty-six random truth tables with two through eight leaves are mapped through
  real DSD decomposition and Kit factoring and checked at every assignment.
* Real mf.Mapping integration covers combinational outputs/complements, one
  register's CI/CO boundary, sequential metadata, and both recursive cursor APIs.

Full actual `go -C abcgo test -race ./opt/dau -count=1` passes (15.377 seconds),
and `go -C abcgo vet ./opt/dau` passes. No experiments, build tags, hidden source
files or testdata directories are used.
