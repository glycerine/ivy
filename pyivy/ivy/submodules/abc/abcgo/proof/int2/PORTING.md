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

Focused verification for this completed component (run from the repository
root):

```sh
GOCACHE=/tmp/abcgo-cache go test -race abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
GOCACHE=/tmp/abcgo-cache go vet abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./proof/int2/sat_bridge.go ./proof/int2/sat_bridge_test.go -count=1
GOCACHE=/tmp/abcgo-cache go -C abcgo vet ./proof/int2/sat_bridge.go ./proof/int2/sat_bridge_test.go
```

The full package is not yet buildable. The remaining source-derived files still
contain invalid C expressions and casts. The C code also relies on solver and
CNF APIs that must be bridged through the exported `bsat.Solver` API. GIA now
provides `Data`, `SetData`, `TakeData`, `Identity`, and `SetIdentity` accessors
along with object value/mark APIs. The source itself also contains unresolved
identifiers/typos in `Int2_ManCreateFrames` and interpolation setup; these need
to be resolved against the intended algorithm, not mechanically copied.

No full-package or algorithm-fidelity claim is made yet. Next steps are to
finish converting C-style loops/expressions, translate solver calls to the
public BSAT API, validate CNF data ownership, then add original-C fixtures
for unrolling, initialization checks, BMC frames, refinement cubes and
preimage enumeration.
