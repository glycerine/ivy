# `map/cov` port status

## Implemented scope

- Native cover-manager lifecycle and object-ID-indexed node state.
- Variable-width Go cube storage, bit operations, cover collection, checking,
  serialization, and support counting.
- Exact SOP minimization through `map/cov/minimize` for supports up to 16
  variables when the exact-cover bounds are met.
- Exact ESOP truth-table preservation: minimum-literal/term search through four
  variables and algebraic-normal-form conversion for larger supports.
- `Abc_NtkSopEsopCover` creates an ABC logic network with exact local SOPs for
  each two-input AIG AND. It preserves object sharing, fanin ordering, edge
  complement phases, PI/PO names, and source function on every input vector.

The network derivation intentionally exploits the source invariant: a STRASH
AIG is a graph of two-input AND nodes plus complemented edges. Each gate's
SOP is therefore one exact product term; there is no need to recover a large
multi-level truth table. This keeps the AIG's sharing while returning a normal
ABC SOP logic network.

## Differences from the C implementation

The C implementation iteratively grows bounded covers across cuts, mixes SOP
and ESOP costs, and may restructure the network. The current native derivation
preserves the original AIG structure and emits local exact SOP gates; it does
not reproduce those cover-growth and restructuring heuristics. The `fUseEsop`
flag does not currently select an ESOP network representation in the public
conversion path. These are optimization differences, not changes to Boolean
function semantics, and must remain visible in review rather than being
described as a line-for-line C port.

## Review notes

- `minimize` has exhaustive truth-table and independently brute-forced
  minimum-cost tests for every three-variable function.
- ESOP conversion toggles each cube's truth contribution, rather than treating
  cubes as an OR cover. Larger ESOPs use the invertible Boolean Mobius
  transform, so overlapping terms retain XOR semantics.
- Network conversion carries source complement bits on graph edges and checks
  the target graph before returning it.
- Cube storage uses Go slices and the collector; no C memory manager or pointer
  reinterpretation remains in these APIs.

## Verification status

The nested package gate is a normal package test:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/cov/minimize -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/cov/minimize
```

The whole `map/cov` files now pass `gofmt` parsing. Normal package tests and vet
are still pending because imported `base/abc` currently fails to parse in its
own generated `abcHieCec.go` sources. After that dependency builds, run:

```sh
GOCACHE=/tmp/abcgo-cache go -C abcgo test ./map/cov -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./map/cov -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./map/cov
```
