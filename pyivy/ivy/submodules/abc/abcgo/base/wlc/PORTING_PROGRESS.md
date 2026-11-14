# WLC port progress

The package is **not buildable**: many draft files still contain untranslated
C/C++ constructs and unresolved cross-package interfaces. The full package
has not passed tests or vet, and no end-to-end WLC behavior is claimed.

Isolated extracted bit helpers, comparators, adders, multipliers and signed
and unsigned restoring dividers have focused race-tested unit tests. The
compiled-C fixtures in `test_vectors/` cover 392 small multiplier cases and
680 signed divider cases. The original C signed multiplier returns `001` for
zero times zero with operand widths 1 and 2; the Go port reproduces this
source behavior rather than changing the algorithm. The C CNF multiplier miter
also reports SAT for widths 1 through 3; the corresponding Go test preserves
that observed result rather than assuming the miter proves commutativity. The
source adder CNF table forbids the correct carry/sum tuple and permits the
other three tuples; a compiled C probe and exhaustive Go test characterize
this behavior.

C iterator headers and brace-delimited iterator macros have been translated
mechanically in the remaining drafts. These conversions are not evidence that
the containing functions or files build or preserve the full source behavior.
