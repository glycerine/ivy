# Pattern, simulation and equivalence-class port

The native implementations in `cecPat.go`, `cecClass.go`, `cecSim.go` and `cecSimBack.go` preserve the algorithms in their corresponding C files. Original pattern/simulation constructors from `cecMan.c` and their manager types from `cecInt.h` live with these implementations. `simulation_words.go` contains the shared signature comparisons, first differing bit, scoring, and source-prime hash calculation.

## Representations and behavior

Patterns use CEC's CI-index literal convention and the original seven-bit varint/delta encoding. Both first-zero and second-zero sensitizing paths are explored, with the shorter cube selected, sorted and compressed. Compatible cubes share simulation bits; presence masks make conflicting insertions atomic. Packing reserves the first bit of each initial-word chunk and doubles row widths as in C. Random words come from the exact mutex-protected GIA MWC generator.

Forward simulation retains the original reference-counted recyclable word arena and class reference lifetimes. Classes are phase-normalized against their first bit, split recursively, and represented using actual GIA representative/next metadata. Native `[]uint32` views replace pointer arithmetic. Arena dimensions derive from slice lengths. No shadow circuit replaces the mutable GIA metadata. Forward simulation also handles real XOR, MUX and buffer objects; MUX control references are counted and released explicitly.

Backward pattern minimization and implication propagation consume normalized AND-based graphs, as the original algorithms assume. Explicit XOR/MUX nodes are rejected at traversal boundaries rather than silently evaluated as ANDs. High-level CEC proof drivers use the separate full-gate `core.Model` path. Backward greedy justification has no backtracking: a failed attempt does not prove UNSAT. Word-parallel implication retains both the one-word split and the multiple-word split, random justification choices, contradiction masks, level scheduling, and cleanup.

The unfinished C `Cec_ManCheckSat2` wrapper returns zero without running its recursive routine; the Go wrapper preserves that behavior. The separate recursive API retains traversal state, including assignments left by a failed greedy attempt.

## Correctness fixes retained as regressions

- Invalid/truncated/overflowing compressed patterns panic before out-of-bounds reads or signed wraparound.
- One-input pattern packing works without relying on a two-row pointer-difference trick.
- Packing grows and retries when no compatible slot remains.
- Simulation releases zero-use signatures from dangling gates and permits arena growth while records remain live.
- Buffer fanins consume one reference; real MUX control fanins consume their own reference.
- Small simulation round counts avoid the source's `nRounds/5` modulo-by-zero condition.
- Supplied initial simulation widths can grow row storage safely.
- Printed arena usage counts bytes rather than labeling a word count as megabytes.

## Verification and integration scope

Run `./test_patterns.sh -race`. It uses copies of the actual native CEC/GIA files in a temporary module and links the existing native dependency packages. Product files are never hidden, excluded by build tags, or replaced by stubs. Other CEC/GIA draft files remain visible blockers to the full package build, which is still a separate required integration gate.

Go tests cover compressed round trips and maximum integers; atomic conflicting insertions and multi-series packing; shorter sensitizing cubes and ternary verification; sorted class refinement/removal; live arena growth, width changes and record release; independently replayed counterexamples; rich forward gates; backward witnesses; and repeated one/multiple-word implication runs.

The optional Linux source oracle compiles the actual `cecClass.c`, `cecPat.c`, `cecSim.c` and GIA random generator. It compares signature tests, scoring and hash values, exact compressed bytes, exact packed simulation words, representative/next class chains, and every state word/counter after backward implication and forward implication steps. It does not invoke an installed ABC executable or substitute another implementation as an oracle.

Independent compactness/readability and separate bug reviews are complete; see `PATTERN_INDEPENDENT_REVIEW.md`. The identified unsupported-gate interpretation was fixed with explicit guards and independent regressions. The combined native CEC/pattern race suite and focused vet pass.
