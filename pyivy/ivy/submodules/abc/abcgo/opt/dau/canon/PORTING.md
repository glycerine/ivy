# Shared ABC truth canonicalization

This acyclic leaf package implements the active `Abc_TtCanonicize` algorithm in `src/opt/dau/dauCanon.c`. GIA cannot import the parent DAU package because DAU itself contains a GIA bridge. Future DAU callers can delegate to `Canonicalize` without changing the truth representation or algorithm.

The input is a normalized little-endian 64-bit truth table for zero through sixteen variables. The output permutation maps new variable positions to their old positions. Input phase bits refer to the new positions; bit `nVars` records output complementation. The algorithm's limited refinement passes and tie handling are retained exactly; this is not a replacement with Kit's different semi-canonicalization.

The optional Linux C oracle compiles the actual C source with unreachable-section removal and compares all truth words, the permutation and phase for all 256 three-variable functions plus deterministic random and structured tied-count tables through sixteen variables. Production code has no C or external dependencies. Standalone Go reconstruction tests verify the returned transformation metadata even when no C compiler is installed.
