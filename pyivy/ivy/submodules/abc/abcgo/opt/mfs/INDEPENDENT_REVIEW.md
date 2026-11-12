# Independent MFS review

## Readability and compactness pass

The package has a clear boundary between the stable-ID truth-table network,
window construction, SAT projection, interpolation and resubstitution. The
network owns its fanin/truth slices, while the AIG remains an adapter and local
cost model. Private reference counts avoid mutating graph fanout storage during
pruning. The production implementation does not retain C pointer-tagging or
temporary global traversal marks. The level-order scheduler now uses a stable
sort rather than quadratic insertion sorting.

## Separate source-fidelity and bug pass

Compared the native path to `src/opt/mfs/mfsCore.c`, `mfsWin.c`, `mfsDiv.c`,
`mfsSat.c`, `mfsInter.c`, and `mfsResub.c`, including the bit-packed
counterexample criterion. The independent pass found and the owner fixed:

* Non-resub BDC inherited the resub-only `WinMax` rejection. A regression now
  proves BDC still handles a window larger than that cap.
* Power passes stopped at the first successful fanin. They now continue with a
  fresh projection after mutation; a five-fanin constant regression requires
  four removals.
* Probability-one sampling treated latch outputs as independent random PIs.
  The simulator now follows the source's 48-frame/16-prefix, ten-word
  sequential protocol; frozen-state tests cover both constant latch values.
* Invalid negative parameters panicked through `Optimize` despite its error
  return contract. The public optimization path now returns an error before
  mutation, while the low-level manager constructor remains checked.
* The level ordering used quadratic insertion sorting; it now uses stable
  sorting.

The actual C root/divisor oracle compares 250 generated cases. Additional tests
check projected SAT and interpolation against exhaustive assignments, exact
packed counterexample coverage across word boundaries, and output preservation.
Current package gates pass:

```
GOCACHE=/tmp/abcgo-gocache go test -count=1 ./opt/mfs
GOCACHE=/tmp/abcgo-gocache go test -race -count=1 ./opt/mfs
GOCACHE=/tmp/abcgo-gocache go vet ./opt/mfs
```

The AIG adapter initially omitted output-unreachable internal nodes. Since the
C sweep includes every internal node and the adapter's purpose is to retain the
network being optimized, `FromAIG` now converts every live AND/XOR/buffer before
adding COs. `TestAigAdapterRetainsDanglingLogic` checks both node accounting and
the roundtrip. I found no remaining functional mismatch in the window, SAT,
projection, resubstitution or adapter paths.
