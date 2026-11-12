# Glucose core review

Scope: SolverTypes.go, Solver.go, Glucose2.go, Heap.go and BoundedQueue.go.
The package-wide C++ bridge, circuit reasoning, preprocessing and utilities are
still separate work; explicit-file checks below do not claim the whole package builds.

## Compactness and readability pass

Clause references are stable Go pointers, not packed arena offsets or pointer
arithmetic. Literal and three-valued encodings retain the original bit-level
semantics. Heap length and capacity derive from its slice. Its position vector
is a genuine indexed-heap lookup, not duplicated length metadata. The bounded
queue keeps logical count (physical storage is capacity), insertion position,
and a sum cache maintained by every mutation. Solver variable arrays retain
separate algorithmic data; no cached variable count is stored. Clause allocator
collection removes dead ownership and dead watches rather than relocating pointers.

The core is organized around propagation, analysis, simplification/reduction,
and search. Configuration fields and public operations are exported Go names.
Proof addition/deletion remains explicit, visible source in proofClause.

## Independent bug pass within the core port

Checked all four LBool raw encodings against original truth masks and undefined
equality; checked self-subsumption with an opposite literal preceding a direct
match. Clause allocator Free now accounts once even when its mark was already
set by deletion. Bounded queue resize resets its sum (the C growTo leaves a stale
sum on reuse); regression covers resize and clear. Heap randomized priority
updates check ordering and membership after removal/rebuild.

Solver regression exercises 1,500 random formulas with five repeated assumption
queries each against exhaustive truth evaluation, checks returned models, and
uses aggressive reduction/restart thresholds. Separate tests cover contradictory
assumptions without poisoning the base formula, incremental contradiction,
zero budgets, asynchronous interruption, dead-clause collection and reset reuse.

Compiled unchanged native sources Glucose2.cpp, Options2.cpp and System2.cpp
with -O2 -ffunction-sections -fdata-sections -DABC_USE_STDINT_H and linked
--gc-sections. Native statuses for a reproducible LCG stream of 500 formulas
and five assumption queries per formula have FNV64 hash 4376511358172924066;
the Go regression reproduces that sequence and hash.

Validation: go test -race over the five core source files and their three test
files passes. Complete-package tests await remaining original source groups.

DRUP regression generates pigeonhole4/3 UNSAT and independently validates each
proof addition by reverse unit propagation, including the final empty clause.

Native oracle reproduction from repository root (the harness is checked in at
`testdata/native_status_oracle.cpp`):

```
g++ -O2 -ffunction-sections -fdata-sections -DABC_USE_STDINT_H -I src \
  abcgo/sat/glucose2/testdata/native_status_oracle.cpp \
  src/sat/glucose2/Glucose2.cpp src/sat/glucose2/Options2.cpp \
  src/sat/glucose2/System2.cpp -Wl,--gc-sections -o /tmp/glucose2-native-oracle
/tmp/glucose2-native-oracle
```
