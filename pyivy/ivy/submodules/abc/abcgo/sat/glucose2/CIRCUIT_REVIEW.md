# Circuit reasoning review

Scope: CGlucose.go and CGlucoseCore.go, plus hooks in normal core propagation,
conflict analysis, enqueue, cancellation, activity updates, reset and model output.
Original attribution to He-Teng Zhang is retained.

## Design / compactness review

The original high-bit CRef encoding becomes an explicit gate-reason object with
a gate variable. Conflict and implication materialization retain original AND/XOR
clause ordering. No uintptr, unsafe pointers or scratch arena aliasing is needed.
Fanin arrays and fanout slice lengths derive from storage; there are no separate
cached edge counts. Fanout edges retain original newest-first traversal order.
Justification heap priorities, current-justification flags, cone traversal marks,
and intrusive watch links represent real algorithm state. Public mode, fanin,
round and cone operations remain available; mode zero uses ordinary CDCL.

## Separate bug review

Reviewed XOR parity and phase handling, opposite signs between reason and conflict
clauses, AND conflict selection when both fanins are false, and rollback of
justification watches. Replaced recursive cone traversal with an explicit stack.
Gate replacement now unlinks old fanout edges before inserting new ones, fixing
the original forward-list setter's repeated-installation corruption. Regression
checks exact old/new fanout membership. Activity rescaling updates all live
justification priorities, not only the most recently bumped variable.

Tests enumerate all assignments and phases for both gate kinds; exercise modes
one and two, repeated solves and partial justification models; compare 500 random
nine-variable cones with ten repeated assumption requests each against exhaustive
truth evaluation of independently constructed gate CNF.

A separate unchanged native C++ oracle, testdata/native_circuit_oracle.cpp,
performs another 5,000 reproducible circuit solves. Its FNV64 status hash is
2660800957835849238, matched by TestCircuitNativeStatusOracle. Compile with the
same command as CORE_REVIEW.md, substituting the circuit harness. Core and circuit
explicit-file race tests pass. Full package remains pending utility/preprocessing/
ABC bridge completion; none of the draft files were hidden or excluded.
