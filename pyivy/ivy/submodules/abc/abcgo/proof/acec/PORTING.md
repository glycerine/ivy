# ACEC port status

The C implementation in `src/proof/acec` has been translated to Go. The
package builds, and `go test -race -count=1 ./proof/acec` and
`go vet ./proof/acec` pass. A fresh `go test -count=1 -cover ./proof/acec`
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

Review fixed several correctness bugs, including discarded final outputs in
matched ACEC, source-derived scratch cleanup failures, native XOR/MUX/buffer
copying errors, missing arithmetic-box handling, polarity ambiguity, signed
signature multiplication, and polynomial handling of XOR, MUX, and constants.

Intentional conservative paths and limits:

- Arithmetic-box normalization can change individual output functions: the
  original routine replaces extra roots at a rank with zero. After the boxed
  CEC run, `Acec_Solve` separately verifies the original circuits before
  reporting either equivalence or a counterexample. It also replays a reported
  assignment on the original circuits and records the first differing output.
  This additional proof can cost time and may return unknown under the conflict
  limit even when the boxed run finishes.
- Arithmetic-box detection assumes two-input AND/XOR cuts. For native MUX or
  buffer graphs, `Acec_ProduceBox` declines the optimization and `Acec_Solve`
  uses ordinary CEC; `Acec_DeriveBox` similarly returns no box. Normalization
  duplicates the graph when no box can be derived.
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
