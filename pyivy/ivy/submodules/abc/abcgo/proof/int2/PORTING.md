# INT2 port progress

`Int2_ManSetDefaultParams` is a native Go reset of the parameter block. The
old draft called C `memset`; the Go implementation assigns the zero value and
then applies the three non-zero defaults. A focused test checks that every
field, including `pFileName`, is cleared before defaults are applied.

The C-object access in `Int2_ManDupInit`, `Int2_ManFrameInit`, and
`Int2_ManRefineCube` has been rewritten to use the public GIA accessors. This
removes invalid C-style assignment expressions from the first two construction
paths. The Go helpers follow the source's initialization muxing and
cube-justification behavior, including its documented defects. Unrolling and
incremental frame construction include inferred behavior where the C source is
incomplete; their Go tests do not establish equivalence to executable C.

`sat_bridge.go` adapts INT2's local SAT calls to the public `bsat.Solver`
interface. Exhaustive assumption tests cover all input and output phases for
AND and buffer constraints. CNF loading also preserves the source literal
encoding, including immediate contradiction detection. The bridge component
passes its explicit-file race test and vet gate.

`int2Util.go` retains the source SOP indexing, including its skipped-literal
behavior and nil-SOP dereference. The incomplete preimage helper returns
undecided. See SOURCE_FIDELITY.md for the exact known defects.

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
manager fields absent from its signature. `Int2_ManPerformInterpolation`
therefore always returns the existing undecided status; it cannot produce a
proof or a counterexample until a faithful driver is available. The previously
invented bounded reachability substitute has been removed.

The GIA dependency and the complete Go INT2 package now build. The original
C source still contains unresolved identifiers and typos in interpolation
setup; these need to be resolved against the intended algorithm, not
mechanically copied. See [SOURCE_FIDELITY.md](SOURCE_FIDELITY.md) for the
source search and the remaining fidelity gap.

The full package passes `go test -race ./proof/int2` and `go vet ./proof/int2`
from the `abcgo` module. Original-C fixtures for unrolling, initialization
checks, BMC frames, refinement cubes and preimage enumeration remain useful.
Interpolation fidelity remains incomplete until a reliable source or contract
for the original image interpolation driver becomes available.
