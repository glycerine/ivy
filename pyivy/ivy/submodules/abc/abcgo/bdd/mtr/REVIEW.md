# MTR review record

Two distinct review passes were performed after the initial port (2026-11-12).

## Compactness, idiomatic Go, and readability pass

- Kept the original Mtr_* entry points and exported node fields for integration.
  The implementation uses ordinary pointers, uint32 intervals and Go GC; it has
  no allocator emulation or unsafe pointer arithmetic.
- Group insertion validates all overlaps before mutating links, replacing the
  original duplicated insertion branches with one common reparenting operation.
- Reader/writer-based APIs centralize parsing, output and error handling; original
  printing/reader entry points delegate to them. No graph algorithm is duplicated.
- Small interval/shift helpers make overflow bounds explicit. Source copyright
  and BSD terms remain in mtr.go. All files are gofmt-formatted.

## Bug review pass

Findings were corrected with regressions before integration:

1. Copy expansion checked fields individually but could still produce an end
   beyond MTR_MAXHIGH. Check the full interval and use division-before-multiply
   bounds to avoid uint64 overflow. Tests cover oversized expansion and the
   maximum valid leaf interval.
2. Large positive shifts could overflow int64 before the range comparison. Check
   the allowable shift interval before arithmetic. A max-int shift regression
   confirms failure leaves the tree unchanged.
3. Group validation could recurse forever on cycles and did not reject an
   out-of-range root interval. Track visited nodes and validate interval ends;
   cycle and invalid-bound regressions confirm errors.

Existing tests also verify nested groups, crossing rejection, insertion gaps,
parent/elder/younger links, copy independence, expansion, dissolution, adjacent
swaps and descendant offsets, failure atomicity, input failures, flags and output.

Verification: `GOCACHE=/tmp/abcgo-go-cache go -C abcgo test ./bdd/mtr -race -count=1`.
The complete MTR package is ported and tested; no files are hidden or build-tagged.
