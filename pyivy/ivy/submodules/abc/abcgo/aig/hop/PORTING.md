# Hop Go port

All 181 original Hop/Cudd2 function bodies now have Go implementations. The
package builds in the actual abcgo module and uses Go-managed memory. Original
ABC algorithms and function names remain recognizable; no generated source files
were hidden or excluded from the build.

## Representation and integration

Complemented graph edges are stable cached proxy pointers referencing regular
nodes. Hop_Regular/Hop_Not/Hop_IsComplement preserve pointer-API identity without
unsafe address tagging. Node pages are typed slices linked through pNext while
free; live nodes reuse pNext for structural hashing. Structural buckets are typed
slices. Temporary data, integer data and traversal IDs have distinct typed fields
instead of reinterpretation through C unions.

Manager-local Vec_Ptr_t preserves the original int32 count/index API with []any
storage. Vec_Int_t is an alias of the common misc/vec implementation. Consumers
that use shared vec.Vec_Ptr_t should explicitly adapt returned pointer vectors;
they are currently separate named types. Core Hop graph constructor APIs retain
*Hop_Man_t and *Hop_Obj_t parameters/results.

EquationOperations/CompileEquation/ParseEquation connect the generic misc/parse
compiler to Hop. Manager methods satisfy bool/dec's Network and VariableNetwork
interfaces and have a concrete factoring parity test. No package cycle is needed.

The Cudd2 recorder uses Hop's graph backend, a Go map and comparable opaque BDD
identities. BDDRef carries an explicit complement flag; uintptr handles retain
the low-bit convention without being converted to Go pointers. Recorder state is
serialized with a mutex. Quit writes the recorded graph as aig_temp.blif, matching
the original recorder's output role.

## Verification

Actual-workspace commands pass:

    GOCACHE=/tmp/abcgo-cache go -C abcgo test ./aig/hop
    GOCACHE=/tmp/abcgo-cache go -C abcgo test -race ./aig/hop
    GOCACHE=/tmp/abcgo-cache go -C abcgo test -cover ./aig/hop

Statement coverage is 70.8%. Tests exercise node-pool ownership/free lists,
complement identity, Boolean constructors, structural sharing and resize,
reference cleanup, graph consistency, traversal order and cycles, levels,
transfer/composition/remapping, balancing, 32-/64-bit truth tables and large
supports, equation and SOP factoring adapters, MUX recognition, BLIF output and
BDD-operation recording. This establishes Hop package behavior; the entire
ABC application still requires integration and parity testing by the main port.
