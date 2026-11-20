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

### Reduce candidate allocation in MF

The eight-input LUT limit bounds truth tables at four words. The mapper now expands the two or three input truths in local arrays, computes the result in the first array, and copies a truth into permanent storage only when interning a new function. A four-word map key replaces serialization to a byte string. Tests check that reusing a candidate buffer preserves older functions, including ones that differ only in the last word.

With truth scratch and reused exact-area traversal scratch, i10 took 48.5–48.8 ms and about 159,200 allocations. Moving leaf-array scratch into each candidate evaluation regressed allocations to about 328,735. The profile showed those arrays escaped. Reusing them across candidate pairs of the same node, preparing fanins once and retaining owned copies of candidate leaves reduced the final run to 43.8–44.2 ms, 11.7 MB and about 100,620 allocations. Utility and MF race tests passed, including full i10 native CNF parity.

The module-wide run started at the Satoko watcher checkpoint passed 130 packages, with three packages reporting no tests.

### Exercise the synthesis demo through its CLI

Built src/demo.c against the same native objects. On i10 the native resynthesis goes from 2,396 ANDs at depth 37 after the first balance to 1,851 at depth 35. C's final equivalence proof passed. One run took 0.13 s for rewriting and 0.22 s for verification. The original Go executable failed on its first read command. Linked the actual I/O and ABC command modules to frame initialization and shutdown. The BLIF dispatcher now converts an AIG or logic network to a netlist using ABC's conversion, and file-read AIGER networks inherit the source filename and default port names when needed.

Three regression tests failed against the preceding revision via a Go build overlay. They reproduced the unknown read command, failure to write a shared AND to BLIF, and missing AIGER network and port names. All three passed with the fixes, as did the root and base/io package race tests. The small NAND CLI test exercises the complete demo with real command handlers. The full module run at the preceding MF checkpoint passed 130 packages, with three packages having no tests.

Native CEC independently checked a BLIF after each Go i10 stage. Reading and both initial balancing passes preserved equivalence. The first rewrite -l changed output V1613(0) from one to zero on all-zero inputs. Subsequent passes retained errors. Go's CEC also panicked on a prover parameter type assertion for the nontrivial miter. These findings preclude a valid end-to-end timing comparison. A diagnostic Go profile across all synthesis stages with intermediate BLIF writes put 52% of CPU in Network.lookupAnd, which scans the object table on a hash miss.

### Check nontrivial miters through the CEC command path

A regression test compares the two associations of a three-input AND. Before the fix, its nonconstant structural miter panicked in Ivy proving when CEC passed a proof/fraig parameter record to the base/abci prover. CEC, sensitivity analysis and redundancy removal now use the command-facing parameter type and their original limits. Both proving entry points accept the concrete type, so the compiler checks every caller. The regression now reports equivalence. The exhaustive two-input prover tests passed alongside it.

### Preserve Ivy FRAIG representatives during refinement

The Go i10 prover reached an unmapped simulation representative. A six-input unit test exposed the underlying class-list error without running a SAT solver. After splitting a class into three pairs, its heads pointed to later members. The pattern scores were [1, 2] where successive source-order splits give [2, 1]. Both assertions failed before the fix. The Go refinement now partitions the linked list in place like ivyFraig.c, and iterates over the remaining nonmatching group. This also eliminates the map and grouped slices formerly created for every refinement. The complete Ivy package passed under the race detector.

### Repair incremental rewriting and measure the synthesis path

The zero-input regression on i10 failed before the rewrite fix. Two initial balances followed by rewrite -l changed V512 from true to false. C computes cuts from live fanins on demand, frees cached cuts when removing an AND from its hash table, and counts a hash hit in the current MFFC as a cost while still sharing the actual AND during reconstruction. The Go pass now follows those rules. The regression covers all 224 output bits. Two more red checks caught incorrect use of MarkA in prospective cone cost and unnecessary duplication of a marked AND. Fixing sharing also exposed a double polarity adjustment and an unordered merge in DSD restructuring. Its four-input OR depth regression failed at depth three and now passes at two.

Native CEC proved the final Go i10 BLIF equivalent to the input. Go produces 1,838 ANDs at depth 33 versus C's 1,851 at depth 35. Both Go and native CEC now report equivalence. With concurrent tests running, one complete Go demo took 27.82 seconds, including 9.31 seconds in rewriting and 18.31 seconds in verification. Eight native demo profiles put the largest flat samples in SAT propagation (21%), solver search (7%) and cut merging (5%).

BenchmarkI10ReadBalanceRewrite reads the input network, balances twice, runs the first rewrite -l, records node count and depth and releases the network. Two Go iterations at one CPU took 2.99 seconds/op, 930 MB/op and 937,243 allocations/op. The native driver in base/abci/test_vectors/perf_rewrite.c took 42.52 ms/op over 30 iterations. The Go and C stages finish at depth 36 with 2,004 and 2,039 ANDs respectively. These are equivalent synthesis tasks with different internal choices. Go's stage profile attributed 55% cumulative CPU time to lookupAnd and 93% of allocation bytes to initializing reference counts. These are the next performance targets.

### Remove whole-network scans during rewrite lookup and cone costing

Temporary MFFC reference counts now keep signed adjustments only for touched objects. The old implementation repeatedly copied every object's fanout count into a new map. A red regression also showed that the snapshot hid edges attached while an adjustment was active. It now returns the live count plus the adjustment. In the same read/balance/rewrite stage this change alone cut allocation from 930 MB to 69.1 MB and runtime from 2.99 s to 1.76 s.

On a structural hash miss the network now checks the shorter of the two input fanout lists. Every matching AND must appear in both. The test covers duplicate unhashed ANDs, opposite edge polarity, rewiring that changes list order, and a manager excluding the first match. Three final one-thread trials took 166–171 ms/op, with 2,004 ANDs at depth 36 throughout. This stage was 42.52 ms/op in C. Its native flat profile favored Rwr_NodeRewrite (18%), Abc_AigAndLookup (16%) and Dec_GraphToNetworkCount (11%). The complete Go demo reached 1.83 s in one run, with 0.64 s rewriting and 1.10 s verification, and proved equivalence. Focused race tests passed; the preceding live-cut snapshot also passed complete base/abc, base/abci and opt/rwr race tests.

### Avoid copying network tables to count live objects

Native ABC keeps separate live-object counts and an upper bound for object IDs. Go now reads its existing per-kind and terminal indexes directly for counts and indexed primary-port access. A sparse-ID simulation regression failed before the fix: the AND after a deleted ID produced zero on two true inputs. Verification now walks the ID range and advances the traversal epoch as in C. The new simulation and sparse-count tests passed under the race detector.

Three one-thread i10 read/balance/rewrite trials took 148–150 ms/op and 49.3 MB/op, down from 166–171 ms and 69.1 MB. All trials preserved the 2,004 ANDs at depth 36.

### Match cut enumeration order under a budget

With two cut slots, the Go enumerator returned [1,2,3] twice. A new regression reproduced the second cut coming from parents (0,1). The C loop enumerates each small right cut against all full left cuts first, which retains [1,2,4] from (1,0). The regression passed after restoring this loop order.

Walk the size-ordered fanin lists directly, use their signatures to rule out incompatible full cuts and track retained cut counts during dominance filtering. Bounded scratch lets oversize merges fail without allocating a leaf slice. Tests check the sorted union and independent ownership across all cut limits from 3 through 12, including modulo-31 signature collisions. Three i10 stage runs took 124.6–126.9 ms/op, 42.9 MB and 599,077 allocations, versus 148–150 ms, 49.3 MB and 859,246 allocations previously. The result still had 2,004 ANDs at depth 36. The earlier full module run covering the live-rewrite repairs passed all 130 tested packages, with three reporting no tests.

### Record the canonical form of each output polarity

The three-input AND regression for Cut_TruthNCanonicize initially returned 0x01010101 for both polarities and left the direct phase at zero. Native cutTruth.c canonizes the direct and complemented truth independently and stores the first minimizing input phase for each. Reusing the Go port of Extra_TruthCanonFastN gives direct 0x01010101 at phase seven and complement 0x7f7f7f7f at phase zero. The cut and extra package race tests passed. The preceding enumeration changes passed complete cut, rwr and abci race tests.

### Expand cut truth tables a word at a time

Cut truth computation now places each child's support using adjacent wordwise swaps, combines the two expanded functions, then replicates across unused variables. Tests check each assignment of randomly generated cuts through twelve variables, including compact inputs, all input/output phase flags and both AND and XOR. The shared adjacent-swap, stretch and shrink routines work directly in the caller's output storage. An in-place eight-variable round-trip regression initially measured 34 heap allocations, and now measures zero. Partially overlapping input/output slices and disjoint buffers also pass.

The native driver in opt/cut/test_vectors/perf_truth.c uses the same leaf placements, functions and complemented first edge as BenchmarkCutTruthMerge. A Go regression compares whole-table hashes against the native results for all three widths. Times are elapsed nanoseconds per merge.

| Variables | Previous Go | Wordwise Go | Native C |
| ---: | ---: | ---: | ---: |
| 4 | 480–498 | 60.9–61.0 | 41.8 |
| 7 | 3,771–3,989 | 152.2–152.4 | 120.9 |
| 12 | 294,756–299,156 | 2,662–2,834 | 2,796 |

The one-thread i10 read/balance/rewrite stage took 120.9–123.4 ms over three trials, with 2,004 ANDs and depth 36. Cut and extra package race tests passed, including the native small-function truth and phase fixtures.

### Score the rewrite manager's own decomposition graphs

The rewrite path previously rebuilt its canonicalization tables, decoded a second forest and constructed a second set of candidate graphs on the first node. A winning four-input XOR gave gain two while the initialized manager's winning node still had zero score, gain and added nodes. A separate regression showed the score report omitted candidate gains and cleanup erased the class count while retaining the candidate scores. Both tests failed on the previous implementation.

The evaluator now uses the manager's own class lists, graphs, permutation table and per-node scores. Score reporting and clearing follow rwrEva.c. The native class-order fixture checks all 222 classes against the actual manager, and an independent test decoder checks the candidate graph functions. Complete rewrite package race tests passed.

The i10 read/balance/rewrite stage reached 103.8–107.7 ms/op, 36.2 MB and 571,640 allocations, down from 120.9–123.4 ms and 42.9 MB. The complete Go demo finished in 1.37 s in one run, including 0.41 s rewriting and 0.89 s verification. Native CEC independently proved the exported BLIF equivalent to i10.

### Prove the same nontrivial equivalence problem in both ports

Added BenchmarkI10NativeResynthesisCEC and the matching base/abci/test_vectors/perf_cec.c. The inputs are the original i10 and the named AIGER produced by the native demo's synthesis commands. Generate the second file from the checkout root with `./abc -c 'read_aiger i10.aig; balance; balance; rewrite -l; rewrite -lz; balance; rewrite -lz; balance; write_aiger -s abcgo/base/abci/test_vectors/i10_native_demo.aig'`. Both drivers load the input pair outside the timer and build a fresh miter, check that structural hashing alone does not solve it, prove it under default CEC parameters and delete it each iteration.

Both programs built miters with 5,051 ANDs and proved equivalence. Native took 210.4 ms/op across 18 calls. Go took 596.8–603.5 ms/op, 165–166 MB/op and roughly four million allocations/op. In the native gprof run, SAT propagation accounted for 40% of samples and solver search for another 12%. In Go, propagation took 18% cumulatively, rewriting 33%, mapped CNF derivation 12% and refactoring 12%. Go sampled allocations included boxing ABC edges, cut enumeration and comparing generic vector entries.

### Label rewrite cones with live references on the nodes

The concrete rewrite adapter's cut-boundary test initially saw one reference on a pinned leaf (expected two) and failed to label the unpinned input when its last reference disappeared. Applying temporary fanout increments directly, advancing the traversal epoch and calling Abc_NodeMffcLabelAig reproduces the native labeling. A signed count adjustment on each object replaces the temporary map. The field uses space freed by grouping Boolean fields. A four-AND reconvergent cone regressed at two heap allocations on the previous revision and now labels without heap scratch.

The i10 read/balance/rewrite stage measured 90.7–92.6 ms, 31.7 MB and 504,270 allocations. The fixed-pair equivalence benchmark measured 560.4–569.2 ms and 151–152 MB across two trials, still proving the same 5,051-AND miter. Complete network and rewrite package race tests passed.

### Size bsat variable storage for the actual formula

The new bsat benchmark and native driver both solve seven pigeons in six holes with 604 conflicts, 788 decisions and 7,409 propagations. A regression fixes those native search counters. Before optimization, Go used 6.23 MB and 6,050 allocations per solve. Native averaged 2.10 ms in an initial 700-iteration profile dominated by propagation (62.5%). Go initially took 5.53–5.69 ms, spending 20% of sampled CPU in integer-key map lookups and 95% of allocated bytes on preallocating variable arrays.

A capacity regression failed because the 42-variable instance reserved space for 50,000 variables and 100,000 watch lists. Variable storage now grows geometrically from the requested size. The test checks repeated growth to 1,000 variables and preservation of an earlier implication. Allocation fell to 327 KB per solve. During high system load a sequential trial pinned both Go binaries and C to CPU 12 for 300 iterations each. Elapsed times were 7.83 ms (previous Go), 5.64 ms (new Go) and 2.70 ms (C); process CPU times per solve were about 7.73, 5.60 and 2.73 ms. The fixed i10 proof used about 140 MB in this version. Complete bsat, fracomb, fra and Ivy race tests passed.

### Look up bsat clauses by page and word offset

The solver now resolves a handle through its page and aligned offset. Learned-clause compaction validates every relocation before installing the surviving pointers. Two regressions failed on the previous revision. A handle with its high bit set decoded to a negative page, and a rejected relocation overwrote the first survivor's ID before reporting the second survivor's invalid ID. Both tests pass after the fixes. Further coverage checks stale handles after rollback and restart. The complete bsat and proof package race tests passed.

In consecutive 300-solve trials on CPU 12, Go fell from 5.23 to 3.26 ms per pigeonhole proof. Native took 2.17 ms with identical search counters. The direct index increased allocation from 327 KB to 438 KB per solve, with 6,050 allocations in both versions. The 5,051-AND i10 proof averaged 551 ms and 142 MB over four iterations.

### Reuse conflict and backtracking storage in bsat

The allocation profile attributed 55% of sampled objects to conflict analysis and 13% to backtracking. Conflict analysis now writes directly into the search's reusable learned clause and uses the solver's stack to minimize it. Backtracking restores the heap in reverse order from the propagated segment of the existing trail. Watch vectors acquire backing storage on their first insertion.

New allocation tests first measured four allocations per repeated conflict, one per backtrack and four reserved entries per unwatched literal. All pass after the changes. The initial trail implementation also panicked when an assumption contradicted a pending root unit. A one-variable regression reproduced the failure before the correction and verifies a subsequent unrestricted solve. Complete race tests passed for bsat, proof, fra, fracomb and Ivy.

The seven-pigeon solve now allocates 301,256 bytes in 1,967 objects, down from 438,112 bytes in 6,050 objects, with identical search counters. Under concurrent system load, alternating 500-solve wall times ranged from 8.52–9.79 ms before and 8.10–9.30 ms after. The timing ranges overlap. The fixed i10 proof averaged 132.5 MB and 3.74 million objects over five calls.

### Build four-input NPN classes by swapping cofactors

Building the canonicalization tables spent 85% of Go's sampled CPU time in Transform. Native's two permutation functions took 87% of flat samples. Go now permutes each representative once per column and computes its 16 input phases by swapping the cofactors of each permuted variable. Visiting the resulting table in ABC's phase and permutation order preserves its chosen witnesses. All four result arrays still belong to each caller.

Consecutive 300-iteration runs on CPU 12 measured 4.986 ms for previous Go, 0.835 ms for revised Go and 10.817 ms for the original uninstrumented C source compiled with GCC -O3. Both Go revisions allocate 327,680 bytes in four arrays. The native driver is misc/extra/npn/test_vectors/perf_npn.c. Exhaustive reconstruction, the native table digest and independent mutation of all output arrays passed. Race tests passed for npn, dec, rwr and rwt.

### Count mapped CNF references during traversal

The fixed i10 proof spent 7.4% of its sampled CPU time searching a vector for nodes already visited in its CNF mapping. Regression tests also found zero area for a seven-clause mapping, a complemented root in the collected gates, reversed DAR traversal order, stale reference counts and ignored supergate marks. The area-only scan allocated and filled a vector that C leaves null. The tests failed before changing the scan.

The traversal now counts incoming mapped references on each regular object, collects a gate only on its first visit and sums the actual cut costs. It handles the native DAR preorder, both transferred-cut orders, marked supergates and recursive calls that continue an existing traversal. Top-level CNF derivations restore structural fanout counts before returning. Complete cnf, pdr, bmc and fracomb race tests passed. The preceding whole-module checkpoint passed all 130 tested packages, with three packages reporting no tests.

The 1,024-gate binary chain in sat/cnf/test_vectors/perf_scan.c has 3,072 clauses of mapped area. Go's scan fell from 14.13 ms and 524,766 allocations to 89.2 µs and 964 allocations. The native driver took 15.1 µs. Consecutive eight-proof runs pinned to CPU 12 measured 521.8 ms and 3.73 million allocations before the change and 474.3 ms and 2.27 million allocations after. Both proved the 5,051-AND i10 miter.

### Restore the alternate mapped CNF used by PDR

The public alternate derivation produced six variables, eight clauses and an empty mapping for a complemented three-input AND. Direct calls to original C with both truth-minimization settings produced seven variables, seven clauses, fifteen literals and a twelve-entry cut map. A regression checks the exact clause order, object ranges and mapping. A separate red test showed that the alternate writer ignored a supplied five-input ISOP. Both writers now share clause emission, and both alternate derivation entry points select DAR cuts using the requested minimization setting. An explicit inverting buffer had returned nil in both mapped encodings. Its new one-leaf cut preserves its polarity.

The mapped format exposed a PDR model-preparation panic on an internal AND absorbed by a cut. A focused test reproduced the panic. Dynamic clauses now use AIG object IDs directly. When a model consumer requests an absorbed physical gate, the frame solver installs its definition so the object can be read after solving. Tests cover absorbed buffers, ANDs and XORs on every input assignment. Complete CNF, PDR, abstraction and bounded-model-checking race tests passed, including the PDR differential against local ABC.

BenchmarkPairedShiftRegisters and proof/pdr/test_vectors/perf_shift.c construct two twelve-stage shift registers driven by a shared unconstrained input. Both prove the final stages always agree through frame 12. Sequential 150-proof runs on CPU 12 took 9.04 ms for the previous Go encoding, 8.81 ms for the restored mapping and 18.29 ms for C. The mapped Go path used 2.52 MB and 8,373 allocations per proof. Native spent 51% of samples loading SOP covers and 19% in SAT search. Go spent 28% in the SAT decision random-number generator, and 77% of its sampled allocation bytes were in variable-array growth.

### Reserve PDR frame variables together

Building a PDR frame solver allocated 118 objects in a new regression. Reserving all 500 initial variables in one call passes the allocation bound and preserves an existing larger solver's SAT model. For the default integral seed, BSAT now reduces the exact integer product modulo 2³¹−1 with a fold. Tests compare the state and result bit for bit with the original recurrence over 10,000 steps per seed, including fractional and signed-zero cases. The standalone generator took 7.14 ns per draw, down from 10.85 ns.

Consecutive 350-proof runs pinned to CPU 12 took 8.73 ms and 2.50 MB per paired-register proof before the changes and 7.38 ms and 1.40 MB after. Both reached frame 12. The seven-pigeon solve measured 3.04 and 3.07 ms respectively, with 604 conflicts, 788 decisions and 7,409 propagations in both runs. Complete BSAT and PDR race tests passed.

### Expand four-input DAR cut truths by support placement

A native driver and a Go test build the same 768-gate layered graph. The test compares a digest of every active slot, its physical position, leaves, value and truth against the C implementation with support minimization both on and off. A separate regression found that both whole-graph Go entry points retained an unused two-gate cone. Native removes it before counting references and scoring cuts. Both Go entry points now do the same.

On the baseline, the Go CPU profile attributed 33% of samples to merging cuts, including sorting and allocation. A rejected union allocated four objects in a red allocation test. Bounded insertion now rejects it without allocating. A small table expands compact functions onto each subset of four variables, and support minimization compares word-sized cofactors. Tests cover every truth function at every four-variable placement and exhaustive support reduction.

Across 200 passes pinned to CPU 12, the public physical-slot benchmark fell from 1,193 to 875 µs with minimization and from 4,109 to 2,184 µs with full support. The original C driver took 142 and 596 µs respectively. C retains the physical slots in node-owned storage. The public Go benchmark includes a separate mutable copy for the caller. Native sampling attributed 43% to computing node cuts, 29% to dominance checks and 21% to ordered merge. DAR, CNF, FRA and PDR race tests passed. The previous alternate-CNF checkpoint passed all 130 tested packages in the complete module.
