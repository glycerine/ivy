# Shared FRAIG manager checkpoint

This package currently contains a source-faithful copy of the FRA manager
needed by `fracomb.Fra_FraigCec`. The FRA CEC wrapper uses `SweepForCEC` here,
allowing CEC to use the same real sweep without a `cec -> fra -> ssw -> cec`
import cycle. The general `Fra_FraigPerform` and choice entry points also use
this manager, including mapped sweeps with callbacks and the built-in BMC
implication callback. The manager, parameters, simulation, class, and BMC
state have shared type identity. The initialized BMC frame unroller and
counterexample reconstruction are also shared with `proof/fra`.

The BMC drivers remain duplicated in `proof/fra`.
In particular, `Fra_BmcPerform` takes different routes to the shared mapped
sweep and implication callback; the existing differential test exercises the
callback, but does not establish equivalence for every BMC option or state.
Consolidation of that driver still needs broader behavioral tests. The C-to-Go
review is not yet a complete proof of source fidelity across every manager
option and rare timeout or error path; fixes should be checked against both
active BMC copies until they are consolidated.
