# ACB source review

## Findings

- `Abc_TtGetBit` and `Abc_TtSetBit` accept pointers to truth-table words in
  `acbFunc.c`, where `NWORDS` is 256. The pointer bridge in `acb_types.go`
  previously exposed only one `uint64`; indexing a bit above 63 could panic or
  read the wrong memory. It now exposes the words through the queried bit
  (`i >> 6`), matching the source's word indexing. The regression test checks
  bit 70 across the word boundary.
- CNF packed-clause and SAT bridge tests exercise the ACB call contracts through
  the owning package APIs, including packed offsets/literals, source literal
  polarity, model access, final conflict, and source assumption minimization.

## Verification

- `gofmt -d base/acb/*.go` produced no output.
- `git diff --check -- base/acb` passed.
- `go test ./base/acb -run '^$'` is currently blocked before ACB type checking:
  `base/abci/abcMap.go` has parser errors at lines 209, 222, 246, 248, 255,
  257, 302, 304, 308, and 316. Those files are outside this review's scope.

This records the source-sensitive ACB fixes and the current build boundary; it
does not claim a completed whole-package runtime review while the dependency
cannot be parsed.
