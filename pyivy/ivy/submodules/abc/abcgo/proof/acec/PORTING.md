# ACEC port status

The C implementation in `src/proof/acec` has been translated to Go. The
package builds, and the full race tests and `go vet ./proof/acec` pass.
Differential cases record source-inherited false-equivalence results described
below alongside exhaustive evaluation of the original circuits. An earlier
`go test -count=1 -cover ./proof/acec`
reported 70.1% statement coverage. This number includes source diagnostic
routines and does not establish verification correctness by itself.

The review and tests exercise half/full-adder cut recognition over input
polarities; carry-tree discovery, phase propagation, and reconstruction;
BDD-independent polynomial expansion on AND, native XOR, and native MUX
networks; NPN classes and LUT statistics; timing-box construction and
collapse; and end-to-end arithmetic equivalence. The end-to-end checks use
exhaustive simulation of small circuits as an oracle, including a 2×2
multiplier and 64 polarity variants of a two-adder chain. Counterexamples are
replayed on the original inputs. The polynomial builders also passed an
additional 1,000 deterministic mixed-DAG stress run over all input assignments.

Review identified several correctness bugs, including discarded final outputs in
matched ACEC. Other changes addressed scratch cleanup failures, native XOR/MUX/buffer
copying errors, missing arithmetic-box handling, polarity ambiguity, signed
signature multiplication, and polynomial handling of XOR, MUX, and constants.

Intentional conservative paths and limits:

- **Known source defect:** after matching arithmetic boxes, `Acec_Solve`
  replaces the final two outputs with constant zero before verification, as
  the C source does. It can therefore report equivalence for circuits whose
  discarded outputs differ. The returned witness belongs to this modified
  miter and may not describe a difference in the original circuits. Do not use
  this result as a proof of the original circuit interfaces.
- The C cut detectors process two-fanin logic and do not check for native MUX
  or buffer nodes. The Go port retains that behavior; these routines can omit
  native gate inputs and their results are not reliable for such graphs.
- `Acec_Normalize` requires an arithmetic box: as in C, it passes a missing
  box to the insertion routine rather than returning a duplicate input graph.
- Box matching requires equal input counts and a successful equivalence sweep;
  like the C implementation, it does not fall back when either fails.
- The original CLA carry rewrite requires a unique polarity along every path to
  each XOR input. Reconvergent XORs commonly violate that assumption. The Go
  port returns an unchanged duplicate when preflight finds ambiguous polarity;
  it does not claim to perform that rewrite in this case.
- The priority-based polynomial expansion retains the C routine's explicit
  restriction against native MUX nodes. The two polynomial builders in
  `acecPo.go` do support native MUXes.
- `Acec_StructTest` is empty in C. `Gia_PolynBuild2` only allocates and frees
  temporary hash tables there; its Go body has no algorithmic effect either.
- Booth partial-product detection, complex multiplier layouts, timing-box
  configurations, and equivalence proofs across a large real-world corpus
  need further differential testing against ABC. Passing the small exhaustive
  cases does not cover all topologies or resource-limit behavior.
