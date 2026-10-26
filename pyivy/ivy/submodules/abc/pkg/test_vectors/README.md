# Permanent stand-alone test vectors

Everything under this directory is permanent Go test input. The test suite must
run without an ABC executable, ABC object/archive, C compiler, C headers, or C
oracle helper. Historical capture programs have been removed from `pkg/tools`
and must not be reintroduced. Dataset maintenance is a review operation, not a
reason to retain or rebuild a C oracle.

There is deliberately no supported “refresh from C” workflow. A dataset whose
meaning cannot be checked from its schema plus a Go model, property, replay or
certificate is not a sufficient oracle and must be replaced before the related
port slice is called validated.

`oracle/reference_build.json` is the authoritative integrity manifest. It lists
the SHA-256 of every permanent oracle artifact plus the top-level AIGER and GIA
regressions; `go test` checks every digest, rejects missing entries and rejects
untracked files below `oracle/`. The manifest does not record compiler,
executable, archive, helper-source, or checkout identity. The datasets are
reviewed inputs, not recipes that must be regenerated from C.

Dataset-backed tests must combine the permanent rows with at least one
stand-alone check appropriate to the behavior:

- an independent Go model or recurrence;
- exhaustive enumeration over a bounded domain;
- algebraic, structural, round-trip, replay, or invariant checks;
- property/model fuzzing with seeds that do not depend on C;
- an independently specified corrected reference for undefined or defective
  historical behavior.

New code must not add C or C++ generators beneath `pkg/`, shell out to ABC or a
compiler, or make a test conditional on a native executable. The stand-alone
integrity test walks the whole Go module and rejects native helper source and
directories named `testdata`. It also rejects subprocess/generator hooks in Go
tests, symlinks that could escape the module, and every file in `pkg/tools/`
except the reviewed non-oracle `source_manifest.py` inventory script.
When a permanent dataset is added or changed, update its digest in
`oracle/reference_build.json`, explain the modeled invariant in the
corresponding Go test, and retain the ordinary and fuzz tests that make the
dataset independently useful.

## Layout

- `aiger/` contains permanent AIGER parser/writer inputs.
- `oracle/` contains immutable JSON, binary, PLA, DIMACS, text and digest
  datasets used directly by Go tests.
- `oracle/source_manifest.json` is a frozen inventory of the original port
  scope. It is planning evidence only; tests never read source files through it.
- `oracle/reference_build.json` contains only stand-alone dataset policy and
  integrity hashes.

Go's built-in fuzzer may use `testdata/fuzz/` for ephemeral failure output. Do
not put permanent regressions there. Minimized permanent cases belong here and
must be loaded by an ordinary test or explicitly added as a fuzz seed.

The Go tests beside each implementation are the semantic catalog for these
datasets. This keeps schemas, intentional corrections, bounds, ownership rules
and invariants next to the code that consumes them instead of tying their
meaning to a deleted generator implementation.
