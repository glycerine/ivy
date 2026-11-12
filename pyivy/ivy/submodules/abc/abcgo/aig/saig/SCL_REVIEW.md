# SAIG sequential cleanup review

## Readability and representation

`Saig_ManReportUselessRegisters` and
`Saig_ManReportComplements` operate on the Go AIG `Graph`/`Edge` representation
directly. This removes C pointer iteration and bit-field assumptions. The
terminal counts come from the graph's input/output slices and register count;
the temporary complement scan uses a short-lived list of marked edges rather
than retaining C-style scratch state.

## Bug review

The complement scan now clears only marks that it set. This preserves unrelated
caller marks and prevents temporary SAIG marks from leaking on return. The
report routines use the graph's reference counts and the sequential terminal
partition (`PI/PO` prefix, latch `LO/LI` suffix) to classify register drivers.
Tests cover repeated regular drivers with opposite edge phase, mark cleanup,
preservation of unrelated marks, and both useless-register report categories.

## Verification and scope

Verified with:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/saig/saigScl.go ./aig/saig/saigScl_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./aig/saig/saigScl.go ./aig/saig/saigScl_test.go
```

The full `aig/saig` package remains unbuildable because other generated source
files still contain C syntax. This review covers only the two sequential
cleanup reporting routines.
