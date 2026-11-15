# DFS and support review

Scope: all 24 entry points in `src/aig/gia/giaDfs.c`, implemented by `giaDfs.go`
and the support-count routines in `support.go`. Traversal allocation/rollover
helpers in `support.go` retain their existing Go implementation; this review adds
growth/rollover checks. Copy-DFS algorithms in `giaDup.c` have separate reviews.

## Correctness repairs

The public `Gia_ManSuppSize` wrapper returned correct counts but left different
traversal state from C: it failed to premark constant zero, including for an empty
root list, and marked CO wrappers rather than dispatching to their drivers.
Minimal red regressions and a failing native trace preceded the repair.

A separate native diagnostic comparison exposed missing/extra field padding in
`CollectTest`, `SuppSizeTest` and `CollectSeqTest`. Restore original labels,
field widths and progress characters. Normalize elapsed values for comparison;
Go continues to measure portable wall time rather than C CPU clock ticks.

## Evidence

`test_vectors/dfs_support_review.c` includes the actual original implementation,
including private helpers. Across 32 deterministic graphs, the tracked Go test
compares:

- 2,048 root selections: support/cone counts and traversal marks, CI and internal
  postorders, explicit cut leaves, and shared-epoch behavior.
- 512 sequential closures, including feedback through two register interfaces.
- 4,608 ordered root pairs for both backward and forward cone postorders.
- All-node scans, 544 single-node support counts, forward/reverse level buckets,
  reverse ordering and all three diagnostic drivers.

Native tests require Linux, a C compiler and the original ABC object files, and
skip when that environment is unavailable. Independent Go tests use a separately
constructed dependency relation and fixed-point set expansion to verify typed
XOR/MUX controls, buffers, constants, disconnected logic, cut leaves and register
feedback. They check membership, uniqueness and applicable postorder constraints.
Existing and expanded tests also cover recursive helpers appending to caller
vectors without changing epochs, repeated calls, exact threshold 16, unused
inputs, constant-output level buckets, ignored CO level values, traversal storage
growth/rollover, and every walk on a 20,000-node cone.

## Compactness

Keep iterative walks to avoid recursion-depth limits. A bounded helper pushes
fanins directly onto the existing stack in reverse order, preserving source
fanin-0-first traversal without temporary slices for each node. Support counting
shares the same walker, removing its duplicate traversal loop. CI-only walks no
longer schedule unused postorder frames. No persistent cache or manager state was
added.

For the same 1,024-AND benchmark (100 ms samples), allocations changed as follows:

| Walk | Before allocations/bytes | After allocations/bytes |
| --- | --- | --- |
| Support | 2,059 / 29,304 | 1 / 16 |
| CIs | 2,059 / 41,584 | 1 / 16 |
| Cone | 2,059 / 41,588 | 11 / 25,200 |
| Sequential | 2,065 / 54,352 | 16 / 37,968 |

Measured times decreased from approximately 78/94/95/98 microseconds to
13/12/33/35 microseconds respectively. These are local measurements, not timing
assertions. Remaining allocations are traversal-stack growth and result vectors.

## Source boundaries

Retain existing Go extensions for real MUX control dependencies, typed XORs,
buffers, constants and CO-root convenience. Original two-fanin source walks may
omit real MUX controls or assert on certain constant/direct-recursive inputs;
independent dependency tests cover the Go extensions rather than claiming native
agreement on those unsupported source inputs. Shared-epoch operations still
require callers to initialize traversal storage, and TFO requires a valid static
fanout map and internal-node roots. Public wrappers that own an epoch initialize
it; recursive helpers do not.

Final module race/vet evidence is recorded in `../../SEQUENTIAL_AUDIT.md`.
