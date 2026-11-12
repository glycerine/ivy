# Delay and timing review

Scope: `ifDelay.go` and `ifTime.go`, with `ifDelay_test.go` and `ifTime_test.go`.

The separate readability pass replaced static C scratch arrays with bounded local slices, kept the original public algorithm names, and separated pin-profile validation from manager bookkeeping. SOP complement metadata is explicit rather than hidden in a vector capacity field.

The separate bug review checked pin permutations, signed sentinel handling, truth/profile consistency, external timing boxes, lifted leaves, and reverse required-time propagation. Regressions cover a shared output driver (take the minimum required time), CI self-cuts (no extra gate delay), lifted-ID decoding, variable pin delays, complemented ISOP covers, generated AIG truth, and invalid profiles leaving output untouched. The integration tests also exercise the real DSD engine and literal-encoded DSD permutations.

The selected-file race and vet gates passed using `test_vectors/delay-time-files.txt`:

```
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race -count=1 $(cat abcgo/map/if/test_vectors/delay-time-files.txt)
GOCACHE=/tmp/abcgo-cache go -C abcgo vet $(cat abcgo/map/if/test_vectors/delay-time-files.txt)
```

These are selected-file gates, not evidence that all remaining IF drafts compile. The manifest includes separately owned native DSD/cut/truth/manager sources to test actual integration rather than substitutes.
