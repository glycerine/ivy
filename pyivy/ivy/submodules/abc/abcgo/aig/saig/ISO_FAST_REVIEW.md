# Fast sequential isomorphism review

Files reviewed: `saigIsoFast.go`, `saigIsoFast_test.go`.

- The compact source signature keeps the original field widths: 3/2/2/3/3/3 bits in the low 16 bits, with visit state held separately.
- Traversal is depth first, visits each regular object once per output, counts repeated positive and complemented fanouts, and follows latch inputs as additional roots.
- Signature frequencies are sorted ascending and emitted in reverse order, matching the C routine's packed `frequency << 16 | signature` vectors.
- Output groups compare signature vectors after removing the trailing PO index. Stable sorting keeps the original PO order for equal signatures.
- Regression coverage verifies that identical outputs group together while a complemented output does not, and checks the packed frequency field on a small fanin cone.

The C function computes and prints classes but returns `NULL` despite its declaration/documentation. This Go port returns the computed classes so callers can use the stated result. Explicit-file normal/race/vet gates passed for `saig.go`, `saigIsoFast.go`, and `saigIsoFast_test.go`.
