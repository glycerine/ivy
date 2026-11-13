# Independent bug review

Reviewed AIG literal polarity, constant-one encoding, output-driver unwrapping, CI/CO order, simultaneous latch updates, SAVE timing, event accumulation, restart capture, shadow subset validation, fairness conjunction, safety assumptions, GF sharing, parser ownership, implication polarity, guarded monotonicity, barrier violation accumulation, absorber sequencing, SAT model validity, and proof/unknown classification.

Regressions added and findings fixed:

- Source CLI `-s`/`-l` meanings were initially reversed; a red command test caught this. Flags now preserve source semantics and default abstraction size.
- The simulation command initially inherited full-mode safety output creation; a separate red test caught it. Simulation stays liveness-only. The source one-step constructor emits a safety output only when it receives safety assertions.
- Missing LTL names initially reached an assertion-style panic; a red test now requires an ordinary error from the native transformation API.
- A preprocessing-only monotone-discovery run cannot claim a liveness proof. Likewise, finite-k failures remain unknown at the top level.
- Original hint indexes are preserved even when an unrelated PO separates them. Known-monotone previous-state lookup avoids the source's negative index calculation.
- Constant false fairness disables unfair-loop counterexamples; no-live/no-fair cases use explicit identity values.
- Full input/state witnesses are replayed independently in randomized tests.
- Invalid graphs and negative SAT limits return errors. Expired deadlines stop k-liveness before invoking a checker; resource exhaustion never becomes a safety proof.
- Zero-width combinations have a valid base case and count overflow is explicit.

The original-C kernel oracle exposed a separate source buffer overflow in the monotone manager display name: allocation allowed one underscore but formatting wrote two. Its generator repairs only that byte count; native strings have no such fixed allocation. All graph equations remain original C in the reference.

Final evidence: full-package normal tests, race tests, and vet pass. Exact C traces cover all major graph-construction families; independent reachable-state enumeration checks PDR against independently enumerated states. Defined source limitations (one-step loops, abstraction, GF-only compilation, and bounded proof resources) are documented in `PORTING.md`.
