# Deco port and reviews

## Complete header coverage

`src/bool/deco/deco.h` defines 30 unique static-inline functions and two iterator
macros. All are available in this package; the iterator macros become Go range
sequences retaining their original index ranges and early-break behavior.
The edge, node and graph types alias `bool/dec` so there is one authoritative
graph implementation and no copied graph metadata. The header's distinct
`Dec_Man_t` factoring scratch and NPN buffers are explicit named Go fields;
no manager algorithm was defined by this header.

## Separate compactness / idiomatic Go / readability review

Reviewed facade necessity, storage duplication and callback/iterator behavior.
Only narrow forwarding functions remain. No function variables or second graph
representation are introduced. Native slices and per-node stable objects belong
to bool/dec; deco adds no capacity or node-count mirrors. Manager scratch buffers
are fields required by the original header, rather than generic cache maps.
`Dec_NodeSetLevel`/`Dec_NodeLevel` preserve the 14-bit logical level boundary.

## Separate bug review

Checked C bitfield assignment, packed conversions, graph constants, root phase,
all four graph constructors in both XOR/MUX variants, caller input phases,
node identity during growth, iterator ranges and invalid-variable access.

- C bitfields truncate the index to 30 bits and complement to one bit. The
  facade normalizes these explicitly, including negative C integer inputs and
  packed bit31 (padding). The underscore union variants use the same portable
  encoding, avoiding native bitfield layout/aliasing assumptions.
- The exported Node member can be populated manually by Go callers, so root and
  constructor boundaries normalize its high bits too. Regression checks that a
  high-bit leaf reference correctly selects the intended low30 index.
- A regression observed `Dec_GraphVar` returning an internal node rather than
  enforcing the source assertion. Added explicit variable-root guards to both
  GraphVar and GraphVarInt, including constants. Reported the authoritative
  bool/dec issue to its owner for the corresponding fix.

Validation: exhaustive Boolean assignments for AND/OR/XOR/MUX with all input
complement phases and both constructions, packed boundary tests, stable growth,
constants/leaves, iterators and header manager metadata. `go test -race
./bool/deco ./bool/dec -count=1` and `go vet ./bool/deco` pass.
