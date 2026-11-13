# ACEC port status

The C implementation in `src/proof/acec` has been translated to Go. The
package builds, and the full race tests and `go vet ./proof/acec` pass.
Differential cases record source-inherited false-equivalence results described
below alongside exhaustive evaluation of the original circuits. A fresh
`go test -count=1 -cover ./proof/acec`
reported 77.0% statement coverage. This number includes source diagnostic
routines and does not establish verification correctness by itself.

The review and tests exercise half/full-adder cut recognition over input
polarities; carry-tree discovery, phase propagation, and reconstruction,
including compiled-C comparisons for one and two disconnected trees;
BDD-independent polynomial expansion on AND, native XOR, and native MUX
networks; NPN classes and LUT statistics; timing-box construction and
collapse; and end-to-end arithmetic equivalence. Compiled-C comparisons cover
all eight two-box polarity/fanout/XOR-trick timing variants and four seven-box
carry-chain configurations; collapsed seven-box graphs are checked on all 512
input assignments. The end-to-end checks exhaustively simulate unsigned
multipliers through 6×6 bits in both operand orders and 64 polarity variants
of a two-adder chain. A sign-extended 4×4 two's-complement multiplier is
exhaustively checked for all 256 operand pairs in both operand orders;
compiled ABC and Go agree on equivalence and three output-inversion cases.
An irregular 3×5 unsigned multiplier is also exhausted over all 256 operand
pairs in both operand orders. Its source comparison covers equivalence, three
output inversions, and a partial-product mutation that needs a nontrivial SAT
witness; C and Go witnesses are replayed on the original graphs.
Timing table shapes of 2×1, 4×3, and 5×4, plus a zero-box manager, are also
compared with compiled C. Compiled ABC is also compared on equivalent 6×6 graphs
and three output inversions; returned mismatch witnesses are replayed on the
original inputs. The polynomial builders also passed an
additional 1,000 deterministic mixed-DAG stress run over all input assignments.

Review identified several correctness bugs, including discarded final outputs in
matched ACEC. Other changes addressed scratch cleanup failures, native XOR/MUX/buffer
copying errors, missing arithmetic-box handling, polarity ambiguity, signed
signature parsing, and polynomial handling of XOR, MUX, and constants.

The C/Go review also found three matcher scratch-vector ownership gaps. The Go
port now frees the two equivalence maps after box matching, the temporary
literal-class copies after shift detection, and the common-count scratch vector
after each count, matching the C cleanup points. `acecCoreMatch_test.go` now
exercises positive and negative rank shifts plus the already-aligned path.
Uncached package tests, `go vet`, and race tests pass after these changes. The
uncached statement coverage run reports 77.2%; this metric includes source
diagnostic routines and does not prove the algorithm correct on all graphs.

Intentional conservative paths and limits:

- `Acec_ParseDistribute` preserves C's `left + right - 1` arithmetic on signed encoded exponents. Negative factors can therefore produce a coefficient encoding that does not represent their mathematical product.

- **Known source defect:** after matching arithmetic boxes, `Acec_Solve`
  replaces the final two outputs with constant zero before verification, as
  the C source does. It can therefore report equivalence for circuits whose
  discarded outputs differ. The returned witness belongs to this modified
  miter and may not describe a difference in the original circuits. Do not use
  this result as a proof of the original circuit interfaces.
- The C CLA output rewrite copies every internal node as a two-input AND and
  starts its output manager with zero registers. Native XOR/MUX inputs or
  sequential metadata therefore do not survive this source path unchanged.
- The C cut detectors process two-fanin logic and do not check for native MUX
  or buffer nodes. The Go port retains that behavior; these routines can omit
  native gate inputs and their results are not reliable for such graphs.
- `Acec_Normalize` requires an arithmetic box: as in C, it passes a missing
  box to the insertion routine rather than returning a duplicate input graph.
- Box matching requires equal input counts and a successful equivalence sweep;
  like the C implementation, it does not fall back when either fails.
- `Acec_ProduceBox` and therefore `Acec_Solve` require an XOR-root chain.
  A graph with no XOR roots or no linked root chain reaches a C root-ordering
  assertion; the Go port likewise panics instead of falling back to ordinary
  CEC. C root ordering also omits roots outside the first selected chain.
- `Acec_Solve` reads only the conflict limit and verbose fields of
  `Acec_ParCec_t`. The other flags, including Booth and the time limit, are
  unused in this C entry point and remain unused in its Go translation.
- The original CLA carry rewrite requires a unique polarity along every path to
  each XOR input. Reconvergent XORs commonly violate that assumption. As in C,
  detecting conflicting polarities violates the rewrite's precondition.
- `Acec_DetectXorBuildNew` initializes its mirror map with zeros in C, so
  even an empty root list can replace ordinary outputs with constant zero.
  The Go port retains this source behavior; it must not be assumed to preserve
  the original circuit.
- `Acec_CollectXorTops` rejects a plain two-input XOR because neither input
  is itself an XOR. It also rejects a discontinuous sequence of XOR outputs.
  In either case C clears the root vector and then dereferences it while
  printing; the Go port likewise panics. The compiled C two-input fixture
  exits with a segmentation fault.
- The priority-based polynomial expansion retains the C routine's explicit
  restriction against native MUX nodes. The two polynomial builders in
  `acecPo.go` do support native MUXes.
- `Acec_StructTest` is empty in C. `Gia_PolynBuild2` only allocates and frees
  temporary hash tables there; its Go body has no algorithmic effect either.
- Booth partial-product detection, larger signed and other irregular multiplier
  layouts, other timing-box shapes, and equivalence proofs across a large real-world
  corpus need further differential testing against ABC. Passing the exhaustive
  small cases does not cover all topologies or resource-limit behavior.
