# Network checker review

## Readability / idiomatic pass

Ported the original 21 checker/comparison entry points in abcCheck.go. Shared
helpers handle ordered-list membership, edge multiplicity, unique names and
interface comparisons. Validation uses local maps rather than pointer sentinels
in user Copy fields. The Go library's Check and CheckRead entry points always
validate instead of depending on a process-global command-frame flag. Opaque
functionality managers may implement NodeFunctionValidator without introducing
import cycles. GOP Boolean functions are accepted as the port's SOP logic
representation; original string SOPs retain syntax checking, including x/n XOR
output tags. Hierarchy cycle detection uses local DFS colors and allows repeated
acyclic module references (unlike the flattening API's stricter single-instance
requirement).

## Separate bug / fidelity pass

Observed red regressions for malformed CI lists causing a nil dereference before
membership validation, and comparison sorting latch boxes by their own names
rather than latch-output names. Reordered validation and added local ownership
checks, then matched original latch-output sort semantics and rebuilt ordered
box/terminal lists consistently. Added special-XOR SOP acceptance coverage.

Tests exercise valid logic/MV netlists and latches, checking without Copy-field
mutation, illegal PI data, undriven nets, unknown reset values, duplicate/malformed
terminal lists, shared CI/CO names with intervening logic, named interface
comparison, latch sorting, and hierarchy DAG-versus-cycle behavior. Combined
ported original/core selected-file tests pass -race. Full base/abc compilation
now reaches the still-unported abcDfs.go draft. No incomplete checker was hidden
or replaced by an always-success result.
