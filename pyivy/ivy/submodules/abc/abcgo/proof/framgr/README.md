# Shared FRAIG manager checkpoint

This package currently contains a source-faithful copy of the FRA manager
needed by `fracomb.Fra_FraigCec`. The FRA CEC wrapper uses `SweepForCEC` here,
allowing CEC to use the same real sweep without a `cec -> fra -> ssw -> cec`
import cycle. The general `Fra_FraigPerform` and choice entry points also use
this manager, including mapped sweeps with callbacks and the built-in BMC
implication callback. The manager, parameters, simulation, class, and BMC
state have shared type identity. The initialized BMC frame unroller and
counterexample reconstruction are also shared with `proof/fra`.

The BMC accessors, implication filter and drivers now use the shared manager.
Before consolidating the main driver, differential tests covered both callback
paths over multiple frame windows; the earlier implementations also shared the
same mapped sweep, and the temporary callback state contained the exact fields
read by the filter. The C-to-Go review is not yet a complete proof of source
fidelity across every manager option and rare timeout or error path.
