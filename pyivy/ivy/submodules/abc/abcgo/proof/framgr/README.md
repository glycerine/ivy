# Shared FRAIG manager checkpoint

This package currently contains a source-faithful copy of the FRA manager
needed by `fracomb.Fra_FraigCec`. The FRA CEC wrapper uses `SweepForCEC` here,
allowing CEC to use the same real sweep without a `cec -> fra -> ssw -> cec`
import cycle. The general `Fra_FraigPerform` and choice entry points also use
this manager, including mapped sweeps with callbacks and the built-in BMC
implication callback. The manager, parameters, simulation, class, and BMC
state have shared type identity. Some FRA manager helper implementations
remain duplicated in `proof/fra`; consolidate those helpers and their tests
before treating this checkpoint as complete. Fixes to duplicated helpers must
be evaluated against both copies in the meantime.
