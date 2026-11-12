# Glucose utility review

Scope: Alg, Alloc, Vec, Queue, Map, Sort, XAlloc, integer portability/constants,
Options/Options2, ParseUtils, Dimacs and System/System2.

## Separate design / compactness pass

All containers use Go slices, generics and checked indices. Vector/region/heap
length and capacity derive from slices. FIFO indices and bounded-queue logical
count describe occupied storage. Map's count is an aggregate across independent
buckets needed for constant-time size/rehash checks; it duplicates no individual
slice length. Custom hash/equality and original prime-growth bucket traversal
remain supported, including collision-safe removal and move semantics.

Region references are integer offsets. The inverse lookup compares Go pointers
without unsafe pointer subtraction. Original endian/compiler/stdint portability
branches become exact-width Go aliases; no C files or cgo remain in these groups.
Go owns allocation lifetime; explicit allocation-limit failures retain a typed
error. Option values/ranges/registry and parsing remain available, while library
errors and help requests return to callers instead of exiting their process.

Input buffering supports plain and gzip readers, retains streaming behavior and
reports absolute offsets plus line/column. DIMACS count mismatches remain warnings.
CPUTime and memory diagnostics read Linux process statistics without subprocesses;
unsupported hosts report zero rather than inventing elapsed-wall-time CPU values.
The source's Linux virtual-memory/peak metric is preserved (not renamed RSS).

## Separate bug pass

Checked vector reserve/grow/shrink ownership, queue wrap/reallocation/reuse,
region references across growth and move, custom map collisions, and range
boundaries/invalid option values. Replaced text-format default key hashing with
Go comparable hashing: pointer keys must retain identity when pointees mutate.
A regression mutates a pointee and verifies lookup and map move.

Exact string match now advances UTF-8 byte length, not rune count. Buffered read
failures and truncated gzip checksums propagate through DIMACS instead of looking
like clean EOF. Invalid integers, overflowing literals, negative headers and
unterminated clauses return errors. Region inverse lookup includes the final
allocated element (the original strict-end pointer assertion rejected it).
Tests retain each of these failures. Existing core/circuit/preprocessing oracle
and exhaustive tests are run together with utilities to detect interaction.

## Validation

The exact race command below passes; replacing `test -race` with `vet` also passes.
This explicit-file check excludes only the three unfinished ABC bridge source
files owned by another worker. They remain visible in the package and the full
package build is still expected to fail until their port is complete.

```sh
GOCACHE=/tmp/abcgo-cache go test -race abcgo/sat/glucose2/Alg.go abcgo/sat/glucose2/Alloc.go abcgo/sat/glucose2/BoundedQueue.go abcgo/sat/glucose2/CGlucose.go abcgo/sat/glucose2/CGlucoseCore.go abcgo/sat/glucose2/Constants.go abcgo/sat/glucose2/Dimacs.go abcgo/sat/glucose2/Glucose2.go abcgo/sat/glucose2/Heap.go abcgo/sat/glucose2/IntTypes.go abcgo/sat/glucose2/Map.go abcgo/sat/glucose2/Options.go abcgo/sat/glucose2/Options2.go abcgo/sat/glucose2/ParseUtils.go abcgo/sat/glucose2/Queue.go abcgo/sat/glucose2/SimpSolver.go abcgo/sat/glucose2/SimpSolver2.go abcgo/sat/glucose2/Solver.go abcgo/sat/glucose2/SolverTypes.go abcgo/sat/glucose2/Sort.go abcgo/sat/glucose2/System.go abcgo/sat/glucose2/System2.go abcgo/sat/glucose2/Vec.go abcgo/sat/glucose2/XAlloc.go abcgo/sat/glucose2/assert.go abcgo/sat/glucose2/circuit_test.go abcgo/sat/glucose2/heap_queue_test.go abcgo/sat/glucose2/pstdint.go abcgo/sat/glucose2/simp_solver_test.go abcgo/sat/glucose2/solver_core_test.go abcgo/sat/glucose2/solver_types_test.go abcgo/sat/glucose2/utilities_test.go
```
