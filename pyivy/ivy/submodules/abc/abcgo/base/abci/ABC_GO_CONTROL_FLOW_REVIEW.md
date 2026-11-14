# `abc.go` control-flow port review

Reviewed against `src/base/abci/abc.c` for the command paths touched while
clearing Go parser errors.

## Findings

- `Abc_CommandIProve` previously relied on `Abc_NtkForEachCo` to introduce a
  C loop and then used `break` outside a Go loop. It now explicitly scans CO
  indices in order, records the first simulated failing output, and exits that
  loop. This matches the source loop's first-match behavior.
- The single-output miter verification path used `Gia_ManForEachCo` in the same
  macro style. It now scans `Gia_ManCoNum` outputs in order and breaks on the
  first phase mismatch, preserving the source's output index and early exit.
- In commands with `usage:` labels, file handles, errors, optional simulation
  vectors, CEX pointers, and the FF-test selected model are declared before
  option parsing where a C `goto usage` could otherwise jump over their Go
  declarations. The usage branches still return the same command status and
  print the existing usage text.
- `Abc_CommandIfif` now routes option failures directly through a local usage
  closure. The closure reloads the current parameter record before printing,
  matching the C label's state at every `goto usage` site while avoiding a
  jump over later declarations such as the LUT library.
- `gofmt` completes for `base/abci/abc.go`. A package build proceeds beyond
  that file; it currently encounters independent parser errors in sibling
  files such as `abcBidec.go` and `abcDar.go`.

## Validation

The `Abc_CommandAbcLoad` and `Abc_CommandIfif` help-path tests cover early
usage exits without a frame. The full `base/abci` test command cannot currently compile the package
because of the sibling-file parser errors above, so the new test awaits that
frontier being cleared.
