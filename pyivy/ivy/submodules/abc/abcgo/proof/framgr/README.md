# Shared FRAIG manager checkpoint

This package currently contains a source-faithful copy of the FRA manager
needed by `fracomb.Fra_FraigCec`. The FRA CEC wrapper uses `SweepForCEC` here,
allowing CEC to use the same real sweep without a `cec -> fra -> ssw -> cec`
import cycle. General FRA entry points still use the original manager in
`proof/fra`, so these implementations are temporarily duplicated. Consolidate
their shared manager and tests before treating this checkpoint as complete;
fixes to either copy must be evaluated against the other in the meantime.
