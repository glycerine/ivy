# INT2 port progress

`Int2_ManSetDefaultParams` is a native Go reset of the parameter block. The
old draft called C `memset`; the Go implementation assigns the zero value and
then applies the three non-zero defaults. A focused test checks that every
field, including `pFileName`, is cleared before defaults are applied.

The C-object access in `Int2_ManDupInit`, `Int2_ManFrameInit`, and
`Int2_ManRefineCube` has been rewritten to use the public GIA accessors. This
removes invalid C-style assignment expressions from the first two construction
paths and preserves the source algorithms for initialization muxing,
timeframe unrolling, and priority-guided cube justification. Those connected
algorithms still need tests once the remaining package dependencies build.

`sat_bridge.go` adapts INT2's local SAT calls to the public `bsat.Solver`
interface. Exhaustive assumption tests cover all input and output phases for
AND and buffer constraints. CNF loading also preserves the source literal
encoding, including immediate contradiction detection. The bridge component
passes its explicit-file race test and vet gate.

`int2Util.go` now uses public GIA/vector APIs to collect support cones and
construct the suffix graph. SOP parsing checks cube lengths and register
indices, and `int2_util_test.go` covers initial-state and multi-cube CO
selection plus malformed input. The preimage loop now uses distinct suffix
CI and prefix-state maps, and releases each temporary assignment/cube before
continuing.

Focused verification for this completed component (run from the repository
root):

```sh
GOCACHE=/tmp/abcgo-cache go test -race abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
GOCACHE=/tmp/abcgo-cache go vet abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./proof/int2/sat_bridge.go ./proof/int2/sat_bridge_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./proof/int2/sat_bridge.go ./proof/int2/sat_bridge_test.go
```

The checked-in C interpolation driver has unresolved names (`vImageOne`,
`vImagesAll`, `vCiMap`, `vCoMap`, and `pGiaSuff`) and the preimage helper reads
manager fields absent from its signature. Rather than return undecided after
frame zero or claim an incomplete interpolant proof, `Int2_ManPerformInterpolation`
currently uses exact reachable-state enumeration from the all-zero initial
state. It proves only after the reachable set closes; the current limits are
16 primary inputs, 262,144 distinct states, and 4,194,304 explored state/input
pairs. A configured depth or time limit can also return undecided.
Counterexamples are reported when a property output is true in a reached
state. Small cyclic, counterexample, and depth-limited machines have focused
regression tests.

The dependency GIA package also currently fails to parse in `giaBalAig.go`,
before Go can typecheck INT2. GIA now
provides `Data`, `SetData`, `TakeData`, `Identity`, and `SetIdentity` accessors
along with object value/mark APIs. The source itself also contains unresolved
identifiers/typos in interpolation setup; these need to be resolved against
the intended algorithm, not mechanically copied.

No full-package or interpolation-fidelity claim is made yet. Next steps are to
run the new transition-system tests after the GIA package compiles, add
original-C fixtures for unrolling, initialization checks, BMC frames,
refinement cubes and preimage enumeration, and reconstruct the original image
interpolation driver if a complete source contract becomes available.
