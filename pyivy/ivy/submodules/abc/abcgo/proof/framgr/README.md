# Shared FRAIG manager checkpoint

This package currently contains a source-faithful copy of the FRA manager
needed by `fracomb.Fra_FraigCec`. The FRA CEC wrapper uses `SweepForCEC` here,
allowing CEC to use the same real sweep without a `cec -> fra -> ssw -> cec`
import cycle. The general `Fra_FraigPerform` and choice entry points also use
this manager, as do mapped FRA sweeps without a callback and the built-in BMC
implication callback. General callback-based FRA entry points still use the
original manager in `proof/fra`, so these implementations are temporarily
duplicated. Consolidate
their shared manager and tests before treating this checkpoint as complete;
fixes to either copy must be evaluated against the other in the meantime.
The two implementations now share the `CombOptions` parameter type and
`Simulation` state. Their manager and class types are still distinct.
