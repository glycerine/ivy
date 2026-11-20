# ABC performance investigation

## Starting inventory

The C source has ten subsystem families. The Go module mirrors each family.

| Family | C subdirectories | Responsibilities |
| --- | ---: | --- |
| aig | 7 | Graph representations including MiniAIG, Hop, AIG and GIA, traversal, simulation and graph I/O |
| base | 14 | Network and word-level models, shell, file formats and command dispatch |
| bdd | 9 | Decision diagrams, reordering and decomposition |
| bool | 7 | Boolean functions, truth tables, covers and decomposition |
| map | 9 | LUT and standard-cell technology mapping, libraries and timing |
| misc | 16 | Vectors, allocators, hashing, utilities and compression |
| opt | 20 | Logic optimization including rewriting, factoring and retiming |
| phys | 1 | Physical design |
| proof | 12 | Equivalence checking, abstraction and other proof engines |
| sat | 13 | SAT solvers and solver interfaces |

The Go commands are abc-demo (read, balance, rewrite, write, verify) and abc-starter (parallel shell command runner). The integration package connects MiniAIG and SAT. The checkout contains 54 Go benchmark functions, including focused graph traversal, mapping, simulation and truth-table workloads. Historical build results in PORTING.md are dated.

## Baseline workloads

The bundled C executable reads i10.aig with 257 inputs, 224 outputs, 2,675 AND nodes and depth 50 using the command `&r i10.aig; &ps`. This circuit supplies a shared input for an AIGER load benchmark. BenchmarkDfsSupportWalks exercises support, input collection, cone size and sequential collection on a 1,024-AND chain. Measure both allocation volume and runtime. Candidate experiments include graph construction allocation locality, traversal scratch storage and integer decoding. Select changes using profiles.

### Counting a shared GIA cone

On the Threadripper 3960X (linux/amd64, Go 1.27.3), three runs of the existing 1,024-AND cone benchmark took 33.15–34.08 µs/op, 25,200 B/op and 11 allocations. The CPU profile across all four DFS benchmarks attributed 58.5% cumulative CPU to giaDfsCollect. A C driver linked against the checkout's object files constructs the same chain, calls Gia_ManConeSize 200,000 times and checks the sum is 204,800,000. It took 5.08 µs/op of process CPU. A gprof flat profile of that driver attributed all 100 samples to Gia_ManConeSize_rec.

Count ANDs at first visit when the caller only needs a count. Ordered node collection still emits on the return visit. Eight local stack slots cover small traversal frontiers. The revised Go cone benchmark takes about 10 µs/op with zero heap allocations. Support also drops from one allocation to zero. The added test exercises nested and repeated roots plus a constant. The C and Go timing methods differ (process CPU and elapsed respectively), so small timing differences should be checked in matched runs.

### Loading i10 through Gia_AigerRead

BenchmarkAigerLoadI10 reads the checked-in file through the file API, requests default structural hashing, verifies graph dimensions and stops the manager each iteration. A matched C file-API driver did 4,000 iterations at 135 µs/op elapsed. The gprof flat profile put 27% of samples in Gia_ManHashFind, 24% in Gia_AigerReadFromMemory and 20% in Gia_ManHasDangling. Go initially took 2.15 ms/op, 807 KB/op and 8,081 allocations/op. Its CPU profile showed map operations throughout the parser and hash pass. Gia_ManAppendObj accounted for 49% of sampled allocation bytes.

Binary AIGER variables have the same IDs as the GIA objects during initial decoding. Using these IDs directly and sharing a single definition/remapping table for sparse ASCII inputs reduced allocation to 654 KB/op and 8,023 allocations/op. Three elapsed measurements were 1.38, 1.63 and 1.23 ms/op. Tests cover sparse and dense numbering, polarity and undefined outputs. Other package tests ran concurrently, so use isolated repetitions for precise time comparisons. Manager object allocation is the next candidate.

### Allocate GIA objects in bounded blocks

Gia_ManAppendObj now takes the next zero-initialized slot from a block of up to 256 objects. The manager still exposes stable object pointers and the ID-indexed pointer slice. The cap limits speculative allocation for graphs that start with a large capacity hint. A regression test holds a CI pointer through multiple blocks and a pointer-table reallocation.

On i10, three measurements took 0.90–1.08 ms/op and 1,737 allocations/op, down from 8,023. Total allocated bytes stayed around 654 KB/op. The reduction is in heap object count. The targeted pointer and variable-numbering tests passed; a module-wide test run was started for this cross-cutting change.

### Hash gates during AIGER decoding

The C reader hashes each gate as it decodes it, then checks the finished graph for dangling logic. The former Go reader built an unoptimized graph and copied reachable gates into a second manager. Reachability before simplification missed gates made dangling by a constant fold. The Go reader now uses the GIA structural hash during decoding, creates a binary-literal remap only on the first changed gate, and runs cleanup on the simplified graph before installing metadata. Object-indexed extensions still require original IDs. Simple mode also applies the constant and equal-input simplifications found in C.

A two-output fixture checks a gate made dangling by a later simplification. Calling the C file API directly produced (AND count, output driver literals) of (1, 1, 7) in default mode, (2, 1, 9) in simple mode, and (3, 9, 11) with strashing skipped. The Go regression test matches each result and checks indexed metadata. Driver marking follows backward edges in one scan, including a real MUX's third input. A separate test starts with stale marks on every object.

Three i10 reads took 652–701 µs/op, 343 KB/op and 1,687 allocations/op. The initial single-manager experiment reached 854–875 µs/op with 7,261 allocations/op before removing temporary fanin slices from driver marking. The AIGER and dangling-graph tests passed. Reproduce the Go benchmark from abcgo with `go test ./aig/gia -run '^$' -bench BenchmarkAigerLoadI10 -benchmem -count=3`.

### Structural hash insertion and reproducible C workloads

C reserves the collision-link vector to the bucket count and retains the insertion slot if appending a gate cannot move the links. Ported both operations. The hot lookup now indexes the two integer slices directly. A forced single-bucket test checks the three colliding entries across link growth, including the insertion immediately after reallocation. The hash tests include a comparison with the original native primitives.

With one Go execution thread (`-cpu=1`), five repetitions of 4,096 unique hash insertions, each followed by a commuted hit, fell from 388–394 µs to 333–341 µs and from 502,464 to 474,064 B/op. The single-thread i10 load improved from 421–431 µs to 378–382 µs. These are separate before/after measurements with the same thread setting. A fresh monotonic-wall-clock C run took 122.5 µs for the hash workload, 127.9 µs for i10, and 5.18 µs for the cone. The Go benchmarks report wall time too. Full GIA tests at the first cone-change snapshot passed in 491 s. Targeted hash tests passed at this snapshot.

The reference driver lives at `abcgo/aig/gia/test_vectors/perf_gia.c`. Build and run from the ABC directory on Linux with the default native library configuration:

```sh
make libabc.a
cc -O2 -DABC_USE_STDINT_H -Isrc -c abcgo/aig/gia/test_vectors/perf_gia.c -o /tmp/perf_gia.o
c++ /tmp/perf_gia.o libabc.a -lreadline -lpthread -lm -ldl -o /tmp/perf_gia
/tmp/perf_gia load i10.aig
/tmp/perf_gia hash
/tmp/perf_gia cone
cd abcgo
go test ./aig/gia -run '^$' -bench 'Benchmark(HashAndChain|AigerLoadI10)$' -cpu=1 -count=5 -benchmem
```

For the timings above, the driver was linked directly against the existing native object files with the executable entry point omitted.

### Count-only traversal on every i10 output

Size queries now use a stack of object IDs and mark each dependency when first queued. This avoids loading a shared object again to discover it was already visited. The ordered collectors still produce their original postorder. The count tests cover traversal marks, overlapping roots, typed gates, constants, and a deep graph with more pending siblings than the local stack can hold.

BenchmarkI10OutputCones sums cone sizes separately across all 224 outputs. The native driver (`perf_gia cones i10.aig`) counted 23,773 AND visits per sweep in 169.4 µs. At one Go execution thread, the old walk took 355–360 µs and 91 allocations per sweep. The new walk verified the same visit count in 121–123 µs with zero allocations/op reported. For the 1,024-AND chain, three Go cone timings were 4.58–4.59 µs against C's 5.18 µs. Go is ahead of C on both of these counting workloads. File loading and hash construction still trail C.

The module-wide run started at commit 450d13fd9 passed 130 packages, with three additional packages reporting no tests. It predates the single-pass reader and the new traversal.

### Satoko CDCL propagation

BenchmarkPigeonhole7Into6 constructs the 42-variable pigeonhole formula through the public clause API, solves it, checks UNSAT and destroys the solver. The C driver uses identical variable numbering, clause insertion order and polarity. Compile from the ABC directory with `cc -O2 -DABC_USE_STDINT_H -Isrc abcgo/sat/satoko/test_vectors/perf_pigeonhole.c src/sat/satoko/{solver,solver_api,cnf_reader}.c -lm -o /tmp/perf_pigeonhole` in bash. Run the Go equivalent with `go test ./sat/satoko -run '^$' -bench BenchmarkPigeonhole7Into6 -cpu=1 -count=3` from abcgo.

Go initially took 10.89 ms/op, about 1.28 MB/op and 9,366 allocations/op. Propagation accounted for 70% cumulative CPU. Hashing the (literal, clause ID) key for blocker lookups was the largest flat sample. C took 2.57 ms/op, with 57% flat CPU in solver_propagate in a separate gprof build. Porting C's watcher pair and binary prefix puts the blocker next to its clause ID and avoids scanning long clauses for binary implications. This also aligns marked-cone blocker checks and inspection accounting with the C propagation loop. A regression test moves one watch, then exercises the opposite watch with and without cone restriction.

Three Go runs now take 4.76–4.79 ms/op, 793 KB/op and 9,306 allocations/op. Before and after, Go took 695 conflicts and 892 decisions per solve. C took 710 conflicts and 875 decisions. Both return UNSAT, but their search traces differ. The race-enabled Satoko subtree and integration tests passed, including the optional native comparison of 900 formula/assumption cases. Investigation of the remaining search cost includes conflict analysis and restart timing.

### Match the native restart trace

Instrumented temporary builds of both solvers to record each decision and each learnt clause with its backtrack level and LBD (number of distinct decision levels). The first difference was event 469, after 204 learnt clauses. Go restarted while another conflict was waiting to propagate. C first finishes propagation, then checks the integer moving average against the global average. The revised search follows that ordering and uses the existing bounded queue. Database reduction likewise runs at a propagation fixed point and checks the conflict threshold before counting learnt clauses.

The revised Go and C traces compared byte-for-byte equal over all 1,584 events. Both solved the benchmark with 710 conflicts, 875 decisions and three search starts. The regression test checks those counters. Three Go runs took 4.49–4.52 ms/op; the differential and integration tests passed.

### Reuse Satoko's conflict-analysis memory

Ported the C timestamp array for counting distinct decision levels and performing binary resolution. Stamps grow to cover the current index domain, including duplicate assumptions that push decision depth beyond the number of variables. Explicitly clear on generation wrap. On its own this change reduced the seven-pigeon run to 4.01–4.02 ms and 8,035 allocations/op.

Learnt-clause minimization now uses Satoko's iterative reason traversal. Tentative marks roll back if a proof fails and remain available to later literals if it succeeds. The solver reuses scratch for learnt literals, the proof stack, tagged variables and last-level variables. Added regressions for mark rollback and retention, deep decision levels and stamp wrap. The race-enabled Satoko subtree, native differential and integration tests passed. Rechecking all 1,584 trace records still matched C exactly. Three final runs took 3.42–3.43 ms, 411 KB and 2,512 allocations per solve.

### End-to-end mapped SAT on i10

BenchmarkI10MappedSAT loads i10, coarsens GIA, maps cuts with an eight-input LUT limit, generates the output-disjunction CNF, initializes and simplifies Satoko, and solves. The matched native workload is `perf_gia sat i10.aig`. C took 21.26 ms and returned SAT with zero conflicts. Its flat profile put 37% in Mf_ObjMergeOrder and 27% in Abc_Tt6Isop. The initial Go benchmark took 171.7 ms/op and allocated 25.8 MB/op in 460,878 allocations. More than half the CPU samples were in the per-assignment truth expansion.

MF now uses the same truth expansion and minimum-base utilities as giaMf.c. The shared Go utilities swap low variables with word masks, swap high variables in blocks of words, and test dependence by comparing cofactors a word at a time. A random-table test checks swapped assignments through ten variables. The MF differential suite now compares the entire i10 CNF against native ABC, including clause and literal order. MF and utility package tests passed.

Three Go pipeline measurements took 56.5–57.1 ms/op, about 22.8 MB/op and 406,782–406,783 allocations/op. The updated allocation profile points to temporary truth tables and leaf arrays during candidate evaluation.
