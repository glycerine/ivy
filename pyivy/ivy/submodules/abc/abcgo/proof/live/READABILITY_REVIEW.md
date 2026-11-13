# Compactness and readability review

Reviewed the complete package after porting all nine source families. The principal reduction in size comes from factoring the source's repeated manager/CI/gate/CO duplication into `graphCopy`, and implementing full/one-step/abstract/restartable Biere constructions through one explicit options structure. Original register and monitor next-state lists remain in source order, making interface alignment inspectable.

The named `Network`/`Frame` adapters contain only information used by these algorithms. Slice-based signal/index groups replace nested untyped C vector casts. Local LTL nodes carry an explicit `HasObject` flag, so valid constant-one edge zero is distinguishable from an unassigned signal. Parser positions are local rather than package-global.

Proof obligations use the typed `SafetyChecker` boundary; the default implementation is a native PDR engine with explicit result, witness and limit fields. K-liveness proof-model duplication is shared between the source-shaped helper and native driver. Combination enumeration streams borrowed indexes and materializes copies only for requested stored groups.

The graph-copy helper snapshots nodes once, while interface counts and output
lookup use direct accessors instead of repeatedly copying graph-wide slices.
Safety-model construction shares the output and register-copy path for both
constrained and unconstrained graphs; the constraint-history latch remains last
in the register interface.

Removed C-only allocation/free plumbing, duplicate static helpers across translation units, disabled demonstration code, unconditional debug dumps, and unused source-null analysis placeholders. Their actual analysis is supplied by the native monotonicity APIs. No build tags, source hiding, unsafe operations, or dummy algorithms are used.

Remaining source-shaped private names aid comparison with original code. Public APIs have Go data ownership and errors, documented in `PORTING.md`. Scope and proof-backend differences are explicit rather than hidden behind a PDR name.
