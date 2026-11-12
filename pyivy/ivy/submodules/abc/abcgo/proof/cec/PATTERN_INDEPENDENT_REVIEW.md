# Independent CEC pattern and simulation review

Scope: cecPat.go, cecClass.go, cecSim.go, cecSimBack.go, simulation_words.go and their tests. Reviewer: AIG package owner, independent of implementation owner.

## Design and readability pass

The manager-local simulation pool, phase-normalized word helpers and shared sensitization traversal have clear ownership and preserve the original algorithm structure. The pool keeps reference counts adjacent to reusable signatures; class operations keep ordered native IDs and explicit scratch vectors. Compression and packing are separated from justification. No blocking readability findings. The optional compactness finding (cecSimSetPhases duplicating Gia_ManSetPhase) was resolved by delegation to the native GIA function.

## Separate bug pass

Reviewed packed-number bounds and overflow, compatible-pattern atomic writes, pool reuse and expansion, class partition recursion, normalized signatures and hashes, stale simulation cleanup, complemented pins and explicit rich gate handling. Found one concrete issue: backward pattern paths silently applied AND semantics to explicit XOR/MUX nodes. These source algorithms require normalized AND graphs. Implementation owner added explicit guards rather than producing unsound patterns. Forward class simulation retains actual XOR/MUX/buffer semantics.

Independent regression simulation_review_test.go exercises XOR and MUX rejection in signature computation, left/right sensitization, ternary verification traversal, scalar backward justification, and implication insertion. Added the test to test_patterns.sh. No additional concrete correctness findings remained at review completion.

Validation: `abcgo/proof/cec/test_patterns.sh -race` passes (3.384 seconds), including independent rejection regressions, existing pool/class/codec tests, and original C differential tests. Full package compilation remains separately blocked by unrelated untranslated GIA files; this focused gate compiles actual current source files in a temporary integration module.
