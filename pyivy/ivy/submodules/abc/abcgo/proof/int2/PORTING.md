# INT2 port progress

`Int2_ManSetDefaultParams` is a native Go reset of the parameter block. The
old draft called C `memset`; the Go implementation assigns the zero value and
then applies the three non-zero defaults. A focused test checks that every
field, including `pFileName`, is cleared before defaults are applied.

Focused verification for this completed component (run from the repository
root):

```sh
GOCACHE=/tmp/abcgo-cache go test -race abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
GOCACHE=/tmp/abcgo-cache go vet abcgo/proof/int2/int2.go abcgo/proof/int2/params.go abcgo/proof/int2/params_test.go
```

The full package is not yet buildable. The remaining source-derived files still
contain invalid C expressions and casts. More fundamentally, the C code directly
reads and writes `Gia_Man_t`/`Gia_Obj_t` private fields and consumes C-style SAT
and CNF helpers. The Go port must move through public GIA accessors and the
exported `bsat.Solver` API; it must not recreate private field access or invent
solver semantics. GIA is currently being ported in parallel, and its owner is
adding the required pData accessors. The source itself also contains several
unresolved identifiers/typos in `Int2_ManCreateFrames` and interpolation setup;
these need to be resolved against the intended algorithm, not mechanically
copied.

No full-package or algorithm-fidelity claim is made yet. Next steps are to
replace C-style object loops with indexed accessors, translate solver calls to
the public BSAT API, validate CNF data ownership, then add original-C fixtures
for unrolling, initialization checks, BMC frames, refinement cubes and
preimage enumeration.
