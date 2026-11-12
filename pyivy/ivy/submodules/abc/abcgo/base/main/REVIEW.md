# Framework port review

## Readability and compactness pass

The package is importable as `abcmain`; all former C implementation files remain
visible. `Frame` owns one command/alias/flag/history registry and one network
backup chain. Dispatch owns command timing and backup creation, so the full cmd
parser can reuse the framework without a second registry or duplicate accounting.
Typed bridge interfaces isolate concrete network, timing, interchange, resource
limit, model-checking, and extension-loader dependencies. Missing computational
backends return explicit errors. The fallback dispatcher implements useful shell
commands and rejects unknown commands.

The original empty restart extension hook remains empty. Header-only API files
contain declarations documentation rather than fabricated implementations. CLI,
competition sequencing, version/date utilities, module lifecycle and Go extension
loading are implemented with standard library facilities. Registered Go extension
loading replaces native dlopen; a host loader can supply other implementations.
Returned collection snapshots prevent callers mutating framework registries.

## Independent bug pass

Fixed and regression-tested reentrant initialization/destruction/close callbacks,
alias expansion limits applying separately to each command, buffered interactive
input preserving read-ahead, recall cloning rather than sharing backup ownership,
and closed-frame setters retaining state or panicking on cleared maps. CEX
replacement also clears the stale CEX vector. Initializer teardown is reverse
initialization order, Close is idempotent, and owned resources shared across slots
are closed once when identity is comparable. Input/output streams remain owned by
the caller.

Tests cover lifecycle reset and callback ordering, registry snapshots, command
errors and quoting, dispatch injection, network backup/recall/ownership, timing,
resource transfer, source paths with spaces, CLI execution, date formatting,
model-check sequencing and extension initialization/teardown. Frame command and
lifecycle execution is serialized by the caller; registry/state access uses locks.
Public adapter fields must be configured before concurrent use.

Validation: `go test -race -timeout 30s ./base/main` and `go vet ./base/main`
pass for the entire package. This does not claim the unfinished whole abcgo tree
builds. Concrete network/interchange/model-check adapters remain responsibilities
of their consuming modules; no unavailable backend is replaced by a success stub.
