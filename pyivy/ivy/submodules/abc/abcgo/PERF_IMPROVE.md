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

### Store DAR cut boundaries inline and stream them to CNF

Each cached slot now holds four leaves and a signature. A bounded pool allocates slots as needed. Merging sorted boundaries also determines both truth-table placements. CNF evaluates candidates by value and copies the chosen cut once. Callers of Slots and Cuts retain independently editable slices. A cached-root regression initially measured ten and fourteen allocations for those APIs. Both now need at most two, including their leaf storage.

With 300 passes, the 768-gate caller-owned enumeration reached 500 µs with minimization and 985 µs with full support, at 1,634 allocations in either case. For the node-owned operation, 500 passes on CPU 11 took 343 and 869 µs in Go, versus 144 and 605 µs in C. Go allocated 29 objects per pass. Ten fixed i10 proofs on CPU 12 averaged 437 ms and 2.01 million allocations, compared with 454 ms and 2.27 million allocations on the pre-DAR comparison build.

Two tests exposed incorrect area-flow choices. Replacing an earlier AIG node with a gate that depends on a later object made CNF read a zero area for that fanin. The mapper now visits fanins first. Already topologically numbered graphs need only the returned slice, down from 23 allocations in the regression. A second circuit tied at cost 40 on two cuts with different reference sums. Original C truncated the sums to its 11-bit field, choosing the two-input cut with value four over the three-input cut with value two. Go previously chose the latter. Both regressions passed after the fixes. Full race tests passed for AIG, SAIG, DAR, CSW, CNF, SSW, abstraction, FRA and PDR.

### Preserve clause identity when loading SAT frames

Go used to insert all frame clauses before linking the registers and reversed the order of the two implications. A red proof-store test recorded different root clause IDs. The loader now reserves variables together, inserts latch equalities in native order before each new frame, and uses one temporary buffer to lift clauses. The test checks all six proof roots and the original CNF after loading. A second red test found that Cnf_DataWriteIntoSolver2 returned the regular solver. It now returns the proof-logging Solver2, and an unsatisfiable four-clause formula yields a four-root resolution core.

The CNF allocator had lost its graph pointer and both declared dimensions, and duplicate flip offsets were applied only once. Regression tests reproduced all three discrepancies. Clause copies now share a bounded backing array across clauses in each copy. Each clause has its own capacity boundary and both public and private duplicates retain their independent data. Packed import of 8,192 ternary clauses fell from 277 to 70 µs and from 8,193 allocations to two.

The paired native driver is sat/cnf/test_vectors/perf_loading.c. Across 500 single-core runs, full Go data duplication took 269 µs and five allocations, down from 590 µs and 16,387 allocations. Native took 8.45 µs and makes one literal copy. Loading and disposing a regular solver took 1.75 ms in Go, down from 2.07 ms, compared with 1.17 ms in C. The Go profile attributed 54% of sampled allocated objects to BSAT clause storage and 27% to copying input clauses for normalization. Native sampling put 36% of flat samples in Sat_MemAppend. CNF, BSAT, ACB, ABCI, PDR, BMC, abstraction and FRA race tests passed.

### Store short SAT clauses with their literal arrays

Normalizing a repeated 66-literal input allocated a temporary slice on every call. The solver now reuses its private normalization vector as in C. A separate test measured 1,043 allocations for 512 short clauses. Both regular and learned streams now allocate bounded blocks of short records containing the literals and optional ID. The block size grows with actual short-clause use. Tests exercise independent input and neighbor storage, retained clause pointers, IDs and restarts. A high-bit watch-tag test also exposed negative decoded literals for valid Go variables at and above 2²⁹. An unsigned shift now preserves all encodable bits.

In 750 loads of the 8,192-clause workload on CPU 12, Go averaged 1.45 ms and 6,220 allocations, down from 1.75 ms and 30,727 allocations. The seven-pigeon solve took 2.95 ms and 300,432 bytes in 1,814 allocations, with the same native search counters. An earlier block-sizing trial used 308,944 bytes because long learned clauses advanced its growth estimate. BSAT, CNF, BMC and PDR race tests passed.

### Count minterms across sparse and wide BDDs

The public Go counting routines returned zero for a function of variables 6 and 7 when asked for its count over two variables. Native CUDD returned one, and returned 0.25 when asked for the count over zero variables. Both ordinary and extended Go counters now scale by the requested variable count independently of the indices. A second red test found that the core lost the sole assignment of a complemented 80-variable cube. Following complemented cofactors directly preserves that assignment. For widths beyond the double exponent range, the extended calculation also recovers a single assignment among 1,200 variables. The ADD background setter now accepts manager-owned constants. With infinity as background, ordinary ADD zero counts as 16 at width four in both C and Go.

The paired queens fixtures construct row-major BDDs using the same operation order and reference lifetimes. Five queens gives ten solutions and 167 DAG nodes. Six queens gives four and 130. The six-queen baseline took 14.66 ms, 8.18 MB and 20,402 allocations per solve in Go across 80 runs. C took 2.60 ms across 100 runs. The Go CPU profile attributed 86% cumulatively to the general ITE recursion, and its computed map accounted for 63% of allocated bytes. Native sampling put 42% of flat time in unique-node lookup and 17% in the dedicated AND recursion. All 13 BDD packages passed under the race detector after the counting fixes. The preceding complete module run passed all 130 tested packages.

### Use CUDD's two-operand conjunction recursion

AND now handles terminals and complemented cofactors directly, orders the operands for commutative cache reuse and consults the computed table when either operand's regular reference count differs from one. OR uses the same recursion through De Morgan's law. The context-aware calls retain cancellation checks inside the recursion. A regression for swapped AND followed by its dual OR initially took eight lookups and one hit. It now takes two lookups and two hits.

Cache hits for BDD ITE, ADD ITE and generic node operations previously returned released nodes with too few child references. Tests now check that each hit restores the children while leaving the returned root's reference count at zero. A constant-only lookup preserves its native read-only behavior. Separate red tests found leaked references after failed bounded AND and ISOP operations. Three retries raised two surviving nodes' reference counts by three each. The two operations now undo the references acquired by new and revived nodes.

The six-queen Go workload took 5.95 ms and 3.26 MB in 180 pinned runs, down from 14.66 ms and 8.18 MB. All 13 BDD packages passed under the race detector, including exhaustive three-variable fused abstraction and reorder tests. The updated profile attributed 68% of CPU time and 89% of allocated bytes to canonical node lookup and construction.

### Count incoming decision-diagram references in one pass

Checking external ownership after releasing the six-queen root scanned every canonical node once for each node being checked. The original Go profile placed 99.7% of samples in the per-node reference query. The whole-manager check now counts incoming arcs once across BDD, ADD and ZDD nodes. Garbage collection uses the same count, and BDD reordering retains each external root's multiplicity across trial permutations. A red test also reported three saturated nodes as leaking. The check now follows CUDD's rule for saturated reference counts.

Three checks averaged 1.084 s each before the change. Afterward, 3,000 checks averaged 223 µs with zero allocated bytes on the released graph. Native averaged 49.6 µs in 6,000 checks. Both returned zero outstanding roots. The native queens driver accepts a final refs argument to time the check on a single constructed and released graph.

A regular handle and its complement now share one Go allocation. The allocation regression originally measured 1,085 allocations to create 512 projection pairs. The full queens build now uses 10,159 allocations and 3.10 MB, down from 20,158 and 3.26 MB. Its 6.08 ms runtime was close to the previous 5.95 ms measurement. All 13 BDD packages passed under the race detector.

### Honor CUDD table sizes and memory hints

Four native initialization cases exposed fixed defaults in the Go constructor. With three BDD and two ZDD variables, a unique-table request of nine and a cache request of 70, C reports 96 unique slots and 64 cache slots. Go reported 1,536 and 262,144. The constructors now round the unique subtable size up and the computed cache size down using the native minimums. A 1 MiB memory hint yields a hard cache threshold of 8,738 entries and a loose unique-table threshold of 5,242 on the 64-bit build. Resetting either threshold uses ABC's default 64 MiB data limit.

Tests also reproduced wrong slot counts when growing variables through the public APIs. Inserting a BDD or ADD variable at the top now moves the previous levels' capacities along with their variables. Aligning six ZDD variables updates the inventory to 160 slots. Growing an empty eight-slot unique table to ten variables doubles the cache once, to eight slots. Both growth cases were compared with native CUDD. All 13 BDD packages and misc/extra passed their race suites in the sizing run.

### Share bounded computed slots across CUDD operations

The old caches let 64 insertions occupy 64 entries in a four-slot manager. The ZDD and constant-only lookups also missed entries inserted by the matching operation. A public flush retained BDD results, a shuffle retained generic results from the old variable order, and a closed manager still returned cached nodes. Each failure was reproduced in a unit test before the port.

BDD conjunction, BDD/ADD ITE and external operators now use the same direct-mapped slot space, with exact operand and operation checks on a hit. Sparse storage allocates only touched slots. The native miss threshold can double the slot count and migrate valid entries while retaining lifetime hit and lookup counters. Flushing invalidates values, drops their Go pointers and preserves the used-slot statistic. Tests check collisions, actual returned values, a four-to-eight-slot growth after ten hits, live entries across resizing and cleanup on quit.

The six-queen graph ended with 3,092 valid entries in 16,384 slots. A first version that retained a copy of each hash took 5.73 ms and 3.36 MB. Recomputing hashes only at resize and releasing manager maps on quit took 5.39 ms and 3.29 MB across 400 runs. The previous unbounded implementation took 6.08 ms and 3.10 MB. All BDD packages and misc/extra passed under the race detector, and both CUDD packages passed again after compacting the entries.

### Count exact minterms using CUDD digits

With infinity as the ADD background, Go's exact counter returned zero for ADD zero and for a variable choosing between those two values. The native terminal count at width four is 16 and the choice has eight assignments. The extended counter also lost the complement of the designated background when that background was BDD one. The extended regression follows the complement calculation in cuddUtil.c. The native debug build reached an assertion in EpdSubtract3 on this particular input, so its output was unavailable for comparison.

The exact count now memoizes one base-2^32 value per regular node. A path makes at most one decision per manager variable, so entire low digits that must stay zero can be restored at the root. The tests compare all 256 three-variable functions against independently scaled truth-table populations at 11 widths. Comparing APA digit slices also dropped from ten allocations to zero in a red allocation regression.

For five queens scaled to 4,121 variables, Go took 23.2 µs and 19,784 bytes per count, down from 137 µs and 316,360 bytes. The checked native driver took 33.5 µs over 24,000 counts, with 65% of its samples in digit shifting and 24% in addition. For six queens at 4,096 variables Go took 28.6 µs and 21,208 bytes, down from 171 µs and 247,624 bytes. The Go profile puts 52% of cumulative samples in memo table growth. In the native six-queen case, two 32-bit operands in Cudd_ApaAdd wrap before widening. At width 36 the native answer was hexadecimal 70000004 for the known four solutions. The Go benchmark checks the full exact result against four shifted to the requested width. All 13 BDD packages and misc/extra passed under the race detector.

### Match CUDD live-node counts and manager pins

A mixed manager with three BDD and two ZDD variables reported nine application BDD/ADD nodes in Go on initialization. Native CUDD reports one. Its seven BDD/ADD keys include isolated projections and unused constants. The ZDD live count starts at four, including both terminals. Go now maintains the dead BDD/ADD key count as canonical nodes release or reclaim their children, and the public counters distinguish these inventories. Nine native snapshots cover creating, releasing, reclaiming and collecting both a conjunction and a ZDD singleton. The mixed manager's garbage collection threshold is 384, up from the erroneous reported value of one.

Red tests also showed that releasing a manager-pinned projection removed two references to one, and that exact lookup of a previously released ADD constant did not revive it. Both paths now retain the correct references. The peak count survives both dereference and collection when the peak was never explicitly queried. The raw node allocator previously admitted ten live keys under a nine-key limit and rejected ten live keys under a ten-key limit after fifteen conjunctions had died. Both cases are now covered.

The six-queen construction took 5.71 ms with 10,141 allocations after correcting the pin lifetimes and accounting. Checking the released graph took 237 µs and zero allocated bytes. A separate three-run experiment moved the child pointers first in the Go map key. Its 5.45–5.65 ms times and identical allocation counts did not justify the layout change on their own.

### Partition canonical decision nodes by variable

Three table implementations ran against the same six-queen construction. Partitioning Go maps by variable reduced allocation to 2.74 MB and took 5.30–5.37 ms. Per-variable arrays with linear probing took 4.48–4.56 ms and 3.07 MB. Reusing the location from a lookup when inserting a missing node took 4.27–4.35 ms in three runs of 500. The final 600-run check during the full-module build took 4.92 ms. The preceding global map took 5.71 ms and 3.29 MB. Native took 2.60 ms on the original paired baseline. The Go profile still places 62% of cumulative CPU time in canonical node lookup and construction.

The selected tables store one key inline for an isolated projection, retain the original child identities during graph rewiring, and allow deletion during iteration. Tests force 96 distinct keys to share a hash and exercise 1,280 canonical functions through growth, partial collection and reinsertion. A profiling run exposed two new allocations in the reference audit. A red allocation test now covers this path. Passing the incoming-arc map from the caller restored zero heap allocations, with the six-queen audit taking 205 µs.

Releasing the three-variable ZDD universe left five live nodes in Go. Native releases three decisions and keeps its two terminals. Both implementations now reproduce the six native free, rebuild and external-ownership snapshots. Additional red tests exposed two out-of-range reads during the exported Go free/rebuild lifecycle. Reading an absent level now returns nil, and requesting a projection can rebuild its universe. The prior full-module race run completed with all 130 tested packages passing, including GIA in 744 seconds.

### Pack CUDD cache records and count projection references

In C, AND on two isolated projections performs no cache lookups. With an additional owner for each projection it performs one lookup per call and hits on the second and third calls. Go produced the opposite results in all six cases. The cache policy now accounts for the manager's implicit reference. A separate red diagnostic test found that the consistency checker accepted a missing or misdirected inverse on both a projection and the one terminal. It now checks canonical pairs. A regular and complemented handle share one reciprocal pointer and a phase flag, saving one pointer per handle.

Correcting the cache policy increased the number of occupied slots at the end of the six-queen solve from 3,092 to 7,061 out of 16,384. Storing those records directly in a Go map took 5.42–5.51 ms and 3.76 MB. A fully dense array after promotion took 4.94–5.15 ms and 3.70 MB. The selected layout stores the first four records inline, then uses blocks of 64. Its slot index starts as a sparse map of integers and becomes a direct integer array when enough slots have been used. The comparison trial took 3.74–3.88 ms and 2.85 MB. Tests verify all 160 entries across index promotion, flush and rehash, including removal of invalidated node references.

In a fresh back-to-back run on CPU 12, C took 1.68 ms over 500 solves and profiled Go took 4.64 ms over 800. Both returned four solutions and 130 DAG nodes. Native flat samples attributed 39% to unique-node interning and 28% to conjunction recursion. Go attributed 59% cumulatively to canonical-node lookup and construction, including 22% in handle allocation. The earlier C baseline was 2.60 ms. All 13 BDD packages, misc/extra and base/abci passed the race detector at this checkpoint.

### Allocate mapped CNF clauses in two flat literal arrays

Both drivers wrote 2,051 SAT variables, 3,074 clauses and 7,170 literals for 1,024 binary ANDs. Native took 82.7 µs per write over 8,000 iterations. Its largest flat costs were SOP conversion (43%) and counting literals (31%). Go initially took 368 µs, 571,928 bytes and 6,172 allocations. A 128-gate regression measured 792 allocations for the regular writer and 793 for the object-numbered writer.

The writers now count first, reserve both clause tables and emit capacity-bounded slices of contiguous literals. A single bulk copy preserves the private version of each clause. A byte lookup decodes each ternary SOP cube. Two covering tests exercise every input byte and all 16 signed literal-pair positions. Eliminating per-clause allocations took the benchmark to 193–194 µs. Building both clause tables during emission took 173–175 µs, with 34.9% of CPU samples still in garbage collection's object scan. The final compact-truth version took 182–196 µs, 286,568 bytes and 13 allocations across three 6,000-iteration runs.

A red SAT test found that compact two-input truth 1 rejected its true row, inputs 0 and 0. Both writers, the mapping export and the small-cut lookup now repeat the significant truth bits over unused inputs. The SAT test checks all 16 functions against every assignment in both output formats. Another regression reproduced the loss of an explicitly assigned cut on a fresh manager and after its graph grew. Cut storage now grows for valid new object IDs.

Native area-flow measurements also exposed two scoring errors. A two-input AND with an internal leaf of area 83 and two fanouts scored 71 in C and 44 in Go. Empty and two-input supergates scored 100 and 341 in C versus 0 and 41 in Go. The Go helpers now return the native scores. The main DAR mapping path already used the SOP score and 11-bit fanout tie break.

All nine affected packages in the final CNF race run passed, including both writers, PDR, BMC, abstraction, DAR and the ABC integration package.

### Count rewrite candidates through the owning AIG hash

A fresh i10 proof profile attributed 23% of Go's allocated objects to boxing complemented decomposition operands. A miss on an AND of two inverted inputs allocated two objects. Passing phases into a typed lookup eliminates both. A generic adapter still runs the original counting path. Tests compare both paths on hits, misses, both input phases, root reuse, constant and operand reductions, deletion-cone membership and both budgets.

Regressions reproduced a deleted hash entry reappearing in decomposition lookup, construction and balancing. These operations now consult the owning AIG manager. A second test showed that a nontrivial pair still returned an AND after either operand's effective fanout count reached zero. The count now follows the native check after trivial reductions. An Ivy rewrite with two missing intermediate ANDs panicked when a typed nil passed as an interface value reached the next lookup. It now counts two new nodes and rejects a one-node budget. Both the balance search and the complemented decomposition miss have zero-allocation regressions.

Both languages proved the same 5,051-AND miter equivalent. C took 221.6 ms per proof over 14 iterations, with 37.5% of flat samples in SAT propagation and 9.4% in solver search. Go initially took 438–452 ms, 120 MB and 1.95–1.97 million allocations. The phased lookup trial took 396–402 ms, 109 MB and 1.27–1.30 million allocations. After updating effective fanout checks, stable sorting and scratch handling in balancing, two seven-iteration runs took 408–411 ms, 107–108 MB and 1.22–1.24 million allocations. Race tests passed for the six affected decomposition, ABC, rewrite and Ivy packages.

### Evaluate only the selected refactoring cone

The i10 profile attributed 6.8% of flat CPU samples to saving, clearing and restoring every object mark for each local truth table. The new paired cone benchmark has three gates in a network with 4,096 unrelated inputs. Both drivers require the same 64-bit XOR truth and three visited gates. Go took 13.9 µs and allocated a 4,864-byte mark array per call. Native took 36.2 ns across 20 million calls, with 54% of samples in truth evaluation and 42% in collection.

A local function table now serves as the visit set, preserving caller marks and the traversal epoch. A two-pass version took 180 ns. Evaluating each gate on its first depth-first visit took 129–130 ns with zero allocations. A two-word regression checks all 128 assignments, node order, shared fanins and a preexisting mark.

Separate red tests exposed incorrect refactoring gains and deletion-cone marks. A constant replacement counted two removable gates with a cut leaf held, although all three ANDs disappear. A nonconstant replacement marked pinned inputs as deletable, advanced the traversal epoch twice and changed the caller's marks. Refactoring now uses the full MFFC size for a constant and the native temporary leaf references plus one labeling traversal for a nonconstant. The standalone ABC helper and all six affected packages passed their race tests.

The complete i10 equivalence proof took 357–363 ms with two-pass cone evaluation. The final single-pass code took 363–376 ms and about 99 MB during the full-module test build. Each proof still checked the 5,051-AND miter. The preceding version took 408–411 ms and 107–108 MB. The native proof measurement for this pair was 221.6 ms.

### Check the i10 network and its ordered object lists

The fresh i10 check benchmark validates 2,675 ANDs per iteration, with file checks enabled in C. Native took 382 µs per check over 3,000 iterations. Its largest flat costs were checking each object's connections (26%) and traversing for cycles (13%). Go initially took 4.70 ms, 2.32 MB and 673 allocations. Grouping expected destinations by source allowed one exact comparison of every reciprocal pin, bringing this to 1.13 ms. Reusing that scratch for a cycle check and indexing the object lists by ID brought it to 446–448 µs. Cycle failures still report the original DFS back edge.

The final checker verifies all per-kind and combined lists in one pass, compares CI and CO names in a shared table, and enumerates stored name IDs without sorting. The AIG check recognizes a cached AND after checking that the pair is irreducible and both inputs have effective fanouts. Three 3,000-iteration runs took 314–322 µs, 134,335 bytes and seven allocations. The full i10 proof took 341–348 ms and 83 MB for the 5,051-AND miter, versus 363–376 ms and 99 MB before this change. C previously took 222 ms for the proof.

Red tests reproduced a duplicate object occupying two ID slots, a removed AND accepted by the whole-network check, missing or duplicate gate entries, and malformed box lists that could pass validation or panic. A separate red test caught the standalone latch checker dereferencing a nil object. The retained tests also cover exchanged reciprocal pins with equal degrees, a high-degree gate with duplicate pins, stale names, direct CI-to-CO namesakes and cached reducible ANDs. All nine affected packages passed their race tests. The full module run started at the preceding refactoring checkpoint passed all 130 tested packages, including GIA in 465 seconds.

### Retain DAR choices and allocate only populated cut slots

The paired 1,024-AND chain now measures full CNF derivation with the same 1,710 clauses and 4,442 literals in C and Go. Native took 446 µs over 1,000 derivations, with 38% of CPU samples computing DAR cuts and 16% scoring area flow. Go initially took 1.19–1.20 ms, 1.39 MB and 3,458 allocations. Half the allocated bytes reserved DAR slots, including ten slots for every input.

A red lookup test found that the prepared manager always returned a nil DAR choice. Another showed that replacing a CNF cut changed the supposedly original DAR scan to area five and kept an absorbed AND referenced. The manager now retains the original chosen leaves, truth, area flow and 11-bit tie value. DAR scanning and cut creation consult that choice. Explicit transfer copies referenced choices and drops absorbed nodes. Tests cover distinct managers with overlapping IDs, mutable CNF cuts, complemented lookups, zero-leaf contradictions and freeing one manager independently.

Packing initial CNF cuts into per-gate records brought allocation down to 387 objects. DAR now reserves slots after visiting both children and keeps only the prefix touched by the merge. Holes inside it retain their original physical positions and values. The public physical-slot API still returns the requested number. A 64-slot test originally cached 64 records for both an input and a two-input AND. It now caches one and two respectively. Both native 768-gate slot digests still match.

Complete chain derivation now takes 934–938 µs, 847,120 bytes and 372 allocations. On the layered graph, node-owned enumeration took 255–330 µs with support minimization and 647–661 µs with full support. The preceding node-owned Go measurements were 343 and 869 µs, and native took 144 and 605 µs. Eight affected packages passed their race tests on the combined changes.

### Validate compact AIGs through their existing hash tables

After compacting the DAR slots, 16% of the full chain derivation profile was in compact AIG validation. On the same 1,024-AND chain, the Go checker took 150–151 µs, 75,304 bytes and 22 allocations. Native Aig_ManCheck took 14.5 µs across 180,000 checks, with 39% of flat samples in table lookup and 27% in hashing. The Go check also audits fanout counts, cached levels and acyclicity.

The Go checker now verifies the identity of each gate in its owning table and compares the gate counts with the table sizes. This detects duplicate and stale entries without building another pair of hash tables. Strictly increasing IDs across all structural edges prove acyclicity on the ordinary path. Graphs containing backward edges still receive a full cycle check. Three 8,000-pass measurements took 48.6–49.9 µs, 20,736 bytes and two allocations. Whole chain CNF derivation took 806–846 µs with 352 allocations.

Red tests reproduced malformed PI pins, an extra PO pin, missing and duplicate interfaces, a deleted constant, and an AND stored with both phases of one regular fanin. All now fail validation. A separate test showed that reordering legal repeated raw outputs rejected an exact permutation. The reorder check now verifies multiplicities, preserving complemented and repeated raw drivers. All eleven affected packages in the race run passed.

### Fold singly referenced CNF cuts and verify the composed truth

Five red postprocessing cases exposed the Go implementation's cost and reference rules. It compared a folded cut with the root cost alone, tried leaves in their stored order and could fold an unreferenced root. It now visits eligible leaves in stable cost order, permits a fold up to the combined cost of both cuts, rejects saturated covers and updates references while removing the absorbed cut. Tests compare the resulting counts with a fresh mapped scan and check that each root absorbs at most one leaf.

A red four-input test found that truth 0x0018 cost seven clauses instead of the native minimum of six. Cut composition now uses the precomputed SOP table for small results. The regression checks all 65,536 functions. Another red test exposed reversed arguments in the support-placement helper. A separate eight-variable test checks 256 rows with overlapping supports and a substituted variable in the middle of the top truth table.

On the three-leaf and seven-leaf AND compositions, Go initially took 1.97–2.04 µs and 7.93–8.99 µs respectively. Compiling the input row toggles once per union and using small SOP lookup brought this to 144–148 ns with three allocations and 3.71–3.72 µs with 62 allocations. The seven-leaf case spends most of its remaining work deriving the two ISOP covers.

The original C driver took 65 ns and 877 ns but produced different Boolean functions. Its three-leaf truth was 0x14 in the low byte, differing from the expected AND on three of eight rows. Its seven-leaf truth differed on eight of 128 rows. In cnfCut.c, the native routine decrements the leaf count before calling Cnf_CutTruth, whose address calculation uses that count. A temporary one-word pointer correction changed both results and still failed to recover the correct tables. The Go regressions check every assignment against the AND definition. The native measurement driver reports its mismatches explicitly. This optional C routine is outside the standard CNF derivation path. Nine affected packages passed their race tests.

### Restore scalar ISOP recursion and reuse its private store

The native reference forms 256 pseudo-random five-variable covers per batch, trying both polarities on alternating functions. Both languages produce digest aa1cc4a627bf6337 over the selected polarity, cube count and every cube in order. C took 312 µs per batch over 9,000 batches. Its profile attributed 64% of samples to scalar recursion and 27% to the two cofactor operations. The original Go code delegated each scalar recursive call to the slice-based implementation. It took 1.49–1.50 ms, 608 KB and 37,024 allocations per batch.

A red test showed that the scalar entry point exhausted its store for a zero cover with zero words available, a constant cube with one word available and a single literal with two words available. All three fit in the native KIT implementation. Another case with seven declared variables and one essential variable verifies that the truth fills all four output words. Scalar recursion now holds cofactors and intermediate truth values in registers. The wide routine reuses its reserved output words for temporary cofactors, and both paths append cubes in native order.

The scalar implementation alone took 492–496 µs and 1,662 allocations per batch. Reusing the private recursion store across calls brought that to 234–236 µs, 96 bytes and two allocations per 256 functions. Stores larger than 256 KiB are released. Caller-owned results survive reuse, including concurrent cover and incompletely specified truth-table requests. Seven-leaf CNF composition fell from 3.72 µs and 62 allocations to 1.68–1.74 µs and 12 allocations. All eleven affected packages passed their race tests.

### Accept existing unary AIG buffer records

The full module suite exposed failures in hierarchy flattening and SBD observability after the AIG checker change. Both callers use a buffer with a single driver and Go's zero value in the unused second pin. A new unit test failed on construction, reconnection and a historical raw buffer record. The checker rejected all three and the second-fanin accessor reported the constant object as a driver. Buffer construction and reconnection now store the absent-edge sentinel, and accessors recognize existing unary records. The new red test and both affected integration tests passed after the fix. All four affected packages passed under the race detector.

The subsequent full module run passed all 130 packages with tests after the buffer fix.

### Report native CUDD occupied buckets and per-level density

A four-stage native fixture grows BDD, ZDD and numeric constant tables independently. The final native manager has 192 buckets, of which 69 are occupied. Go originally reported 214 occupied buckets out of 152 and returned the same value for theoretical occupancy. The regression failed at every stage. Go now counts nonempty native hash buckets using the ABC build's node identity and complement-bit hash. BDD and ADD keys share a logical level, ZDD keys use separate levels and numeric constants use their IEEE-754 words. The expected fraction uses each level's key count and width separately. Explicitly dereferenced keys remain in both diagnostics until collection.

Additional coverage exercises the native pre-insertion density threshold, an existing-key hit, a permuted mixed BDD/ADD level and a sparse table with two billion logical buckets. A red mixed-family test showed the shared subtable staying at eight slots when a new ADD key should have grown it to sixteen. Capacity changes update the computed-cache limits. Go's underlying open-addressed storage still grows according to its own probe density. The native decision to collect under dead-node pressure needs a separate adaptation for Go's unpinned intermediate results.

The C driver took 222 ns per actual/expected query pair. The Go reproduction took 1.81–2.18 µs and three allocations initially. Small stack bitmaps brought this to 1.58–1.85 µs and zero allocations. The Go profile attributed 40% cumulatively to marking occupied buckets, 23% to iteration over the constant map and 13% to exponentiation. Six-queens construction took 3.86–4.16 ms across three samples. All thirteen BDD package race suites passed.

### Build variable interactions once per BDD or ADD root

The old interaction builder walked every BDD node as a separate root, built two maps per search and enumerated the support pairs repeatedly for each nested cone. It also omitted ADD roots. A red regression identified eight missing pairs across two ADDs with a shared tail. A separate red test added four variables to a manager and observed existing pairs (2,4) and (9,14) reappear at (1,12) and (6,8). The old packed matrix interpreted its bits using the current variable count.

The builder now scans levels from the top, searches each root once and revisits a shared subgraph when collecting a different root's support. A byte of existing node padding holds the local and global visit state. Consecutive variable indices within a support can be written to each packed row as bit ranges. Indexing rows by the larger index keeps pair positions stable as variables are added, and removal discards the corresponding rows. A 139-variable test compares the whole matrix twice against 25 independently constructed mixed supports in reverse level order.

For 32 independent 24-variable cones, 8,832 true pairs, Go went from 2.01–2.02 ms and 5,001 allocations to 28.6–29.3 µs and 14 allocations. On a 128-variable chain, 8,128 pairs, it went from 3.13 ms and 2,054 allocations to 6.08–6.12 µs and 18 allocations. The intermediate root-search version still set individual pairs and took 59 µs and 36 µs. The original C routine took 857 µs and 65.2 µs in the long runs. All pairs were also checked in separate C and Go runs. C profiles attributed 96% and 71% respectively to interaction initialization itself. All thirteen BDD package race suites passed.

### Validate and compact groups of retiring BDD variables

Red tests showed that destroying a BDD subtable could invalidate a referenced ADD, while a released BDD parent prevented valid removal. Replacement variables inherited the binding, type, pair index and lazy group of their predecessors. Further failures covered a saturated isolated projection, an uncounted collection and a pre-collection hook veto that left the manager partially resized. The native fixture confirms rejection for a live ADD or BDD parent, successful removal after dereference and one counted collection.

Go checks the live keys at each selected BDD/ADD level and validates the entire request before changing a projection. A saturated projection requires a scan for live parents. Collection respects active Go temporaries and removes explicitly released keys. One pass compacts the level widths and both permutations, retaining any capacity reserved by InitTable. Removed handles lose their manager, surviving handles keep their identity and newly allocated variables start with default attributes.

Inserting and removing 128 variables at level 64 in a 256-variable manager initially took 74.0–74.4 µs, 60.7 KB and 153 allocations per cycle. Table and collection changes brought this to 19.4–19.6 µs and 128 allocations. Initializing fresh projection pairs in groups of up to 128, then clearing the retired level records together, brought it to 10.9–11.0 µs and one 18,432-byte allocation. C took 10.2–10.8 µs. A test covers reuse of a preexisting projection inside a batch, partial removal and reserved capacity. Six-queens construction took 3.75–3.82 ms and 10,410 allocations after the change. All thirteen BDD package race suites passed.

### Collect released decision nodes under unique-table pressure

A red BDD test filled fourteen levels with 448 explicitly released keys. The old Go table doubled its width and retained 447 dead keys after reusing one as a child. Native C collected once, retained the eight-bucket level and finished with 31 BDD/ADD keys including constants. Red ZDD and numeric tests showed that even an existing-key lookup can trigger collection in those families. Disabling GC preserves the original keys and expands the table. The Go implementation now matches all four native cases, preserves unpinned intermediate Go results and removes cache entries referring to retired handles.

Additional red tests covered a collection hook that interns the same key as its caller, a failed bounded allocation that grew a subtable during rollback and an overbroad slowdown that affected constants, ZDDs and BDD variables created after the slowdown. A successful rehash restores the native density threshold for the affected BDD level. A separate red memory test found 256 probe slots retained for one surviving projection and another 256 for seven nodes. Sweeping directly in probe order permits inline storage of a lone survivor and downsizing of sparse backing arrays.

Go initially took 100–105 µs for the complete pressure workload, including manager creation. Sweeping in probe order brought this to 84.6–85.0 µs and about 115 KB per manager. C took 35.2 µs, with 71% of its sampled time in table initialization. The initial Go profile attributed 37% cumulatively to allocation. The thirteen BDD package race suites and base/abci passed on the final implementation.

The full-module run started with automatic table-pressure collection passed all 130 packages with tests.

### Register present/next-state maps for BDD reachability

Both network and AIG initializers built the initial cube without registering the state-variable swap advertised by their C counterparts. The two native public entry points were missing. Once the core API was present, separate red integration tests showed each initializer still leaving the manager unmapped. Both reachability drivers now register once and reuse mapped functions across image iterations. Direct callers that supply their own initial cube still get the default latch swap when needed.

Tests cover complement phase, replacement and invalid registration, computed-cache lifetime, heap shuffling, variable creation and subtable destruction. All 256 three-variable functions are checked for an overlapping mapping. The initial registration preserves unrelated cached operations, subsequent registrations flush them and variables created later map to themselves.

Mapping all 256 eight-bit state cubes originally took 277–279 µs, 30.7 KB and 768 allocations per batch through general permutation. The registered version took 10.9–11.1 µs and zero allocations. C took 7.62 µs. Its profile attributed 24% of samples to unary cache lookup and 20% to mapping recursion. The Go profile also identified cache lookup and canonical node probing.

A six-bit enabled counter reaches its first all-ones state at frame 63. Both Go image modes replay the returned trace. Partitioned mode went from 2.02–2.09 ms and 5,202 allocations to 1.89–1.99 ms and 4,778 allocations. Monolithic mode went from 1.79–1.80 ms and 4,976 allocations to 1.62–1.72 ms and 4,553 allocations. The original C partitioned solver took 675 µs. Its monolithic solver returned unknown before image iteration because bbrReach.c tests the partitioned tree pointer after constructing the monolithic tree. A temporary native comparison build checked the selected pointer, replayed the same trace and took 515 µs. The repository's C source remains the benchmark reference. In the new Go profile, fused conjunction and abstraction account for 46% of samples and counterexample construction for 52%, including backward images. The thirteen BDD packages, base/abci and misc/extra passed their race tests.

### Restore numerical ADD composition and ordered substitutions

Seven retained tests failed against the preceding implementation. Applying weights 3 and −2 to leaves 5 and 7 returned 5 where C computes 1. A nested case returned 7 instead of −14. Sequential substitution followed variable index in the wrong direction, the internal entry ignored its remaining cube, and four APIs accepted foreign-manager inputs on terminal shortcuts. Two permutation regressions covered an incomplete final triple and listing the same variable swap twice.

The composition routines validate all participating edges and use the manager's current level order. Simultaneous composition stops below the deepest substitution. Sequential composition applies the deepest replacement first. Weighted composition multiplies each child by its own weight and sums the products. The native driver and the Go benchmark both check the result on a reverse-ordered manager. A separate Go test checks all eight assignments with nested, overlapping, zero and negative weights.

The first correct implementation took 841–855 ns and 18 allocations per weighted composition. Its profile attributed 93% of time to generic arithmetic application, including allocations for operand pointers passed to callbacks. Giving the two arithmetic operations distinct computed-cache keys and passing operands directly reduced this to 109–111 ns and zero allocations. Original C took 106 ns over three million calls. The thirteen BDD package race tests passed.

### Fuse cached conjunction and existential abstraction

A bounded image operation returned nil after adding a unique key and advancing the next ID. A second red test showed an inner operation overriding its parent's tighter ceiling. A third found that existential, universal and XOR abstraction returned nil for a negative cube without setting the native invalid-argument error. A fourth showed that swapping the operands of a repeated image query produced zero cache hits. All four tests now pass. Independent truth tables check existential and universal abstraction on all 256 three-variable functions and all eight cubes after shuffling levels; the existing 524,288-case test checks the fused result.

Fused image recursion and existential abstraction now share the manager's computed table across calls, skip irrelevant cube levels and stop when a quantified branch determines the result. Universal abstraction uses the complement dual. Limited image operations share the same rollback and nested-ceiling logic as limited conjunction. Cache hits still check the caller's cancellation context.

The commuted query fell from 449–459 ns to 39–44 ns, with zero allocations. Original C took 9.5 ns with one lookup and one hit, and spent 62% of its CPU samples in the cache lookup. On the six-bit counter, partitioned Go took 1.72–1.84 ms on a same-core repeat (1.79–2.15 ms in the profiled run), down from 1.89–1.99 ms. Allocations fell from 4,778 to 4,152, while bytes rose from 1.04 MB to 1.12 MB. Monolithic Go fell from 1.62–1.72 ms to 1.36–1.46 ms and from 4,553 to 3,879 allocations. Conjunction and abstraction still account for 38% of the counter profile, and trace replay for 17%. The thirteen BDD packages, base/abci and misc/extra passed their race tests.

### Reuse the AIG evaluation order and buffers during trace replay

Two red replay tests exposed incorrect handling of trace dimensions. A graph with one register and an additional output returned two next-state bits. An oversized trace with an int32-wrapped bit count began simulating its first frame before discovering invalid data. Next-state extraction now reads exactly the registered latch inputs and validates the bit count with 64-bit arithmetic before evaluating a frame.

The old trace replay copied the graph and its terminal lists, computed an all-object order and allocated evaluation buffers on every frame. It spent 33% of its samples obtaining the order and 31% allocating. Replay now obtains the order once and uses the AIG's scalar simulation loop with reusable inputs and values. Tests check dirty buffer reuse, rewritten logic with non-topological IDs, a deletion hole, a dangling AND, explicit XORs, inverted buffer fanins and final graph marks. Existing randomized tests compare all final outputs and registers to independent graph evaluation.

Replaying the six-bit counter's 64 frames took 117–123 µs, 176,136 bytes and 705 allocations before the change. It now takes 8.85–9.12 µs, 272 bytes and five allocations. C took 6.92 µs, with 99% of its samples in the replay function. The full partitioned Go solver took 1.49–1.55 ms on a same-core repeat and the monolithic solver took 1.10–1.13 ms. Both replay the trace twice and now use 1,400 fewer allocations per solve. AIG, SAIG, BBR, PDR, SSW and base/abci passed their race tests.

### Pack sparse computed-cache positions into integer arrays

A red allocation test found two heap allocations when inserting four computed records, although the records themselves fit inline. The first four slot numbers now also fit inline. Larger sparse indexes use open-addressed 64-bit words that contain both the logical cache slot and its packed record position. Dense indexes retain direct integer lookup. Cache hits compare the stored key in place. A 300-entry strided-slot test covers sparse growth, unrelated misses, partial invalidation, and replacement of both live and invalidated entries.

The Go map implementation took 39–44 ns for the commuted image hit. The packed index took 28.3–28.4 ns in the final trial, with zero allocations in both cases. The six-queen solve took 3.69–3.76 ms and 10,396 allocations, compared with the earlier 3.75–3.82 ms and 10,410 allocations. A batch of 256 state maps stayed at 10.9–11.0 µs. The C references for these workloads were 9.5 ns, 1.68–2.60 ms and 7.62 µs respectively.

Across paired 650-solve runs on the same core, partitioned reachability went from 1.50–1.67 ms to 1.35–1.37 ms and used 33 fewer allocations. Monolithic reachability went from 1.10 ms to 1.03–1.09 ms and used 22 fewer allocations. Bytes per solve stayed near 768 KB in the partitioned case and fell from 682 KB to 672 KB in the monolithic case. All thirteen BDD package race suites, base/abci and misc/extra passed.

### Preserve the structural AND hash across network edits

Seven red cases showed that an unrelated edit could replace the newest duplicate AND with the oldest. Other failures covered a deletion callback seeing the removed entry, losing the native predecessor after deletion or object reuse, bucket order after rehash, accepting a cached non-node, and unchanged pin or data writes displacing a valid entry. A miss on an excluded node could repopulate the network's hash and spuriously resize an empty manager. The native fixture checks insertion, deletion, output editing, rehash and resize, including growth from three to eleven bins on the eighth insertion.

Network edits now invalidate the affected fanin pair. A raw node rewired into a cached pair still participates in the existing lowest-object-ID fallback. Explicitly registered duplicates retain their insertion order, reverse on rehash, and reveal their predecessor when removed. Unique pairs use the normal hash entry alone. Deletion unlinks the node before calling the cut callback. The tests distinguish edited functions, complemented pins, excluded nodes, and the source ordering of registered duplicates.

Toggling an output and looking up its shared-input AND took 4.71–4.79 µs in Go. Fanout scanning accounted for 98% of the old profile. The updated lookup took 60–65 ns and zero allocations. C took 5.4–5.5 ns and spent 94% of its samples in lookup. A paired i10 CEC run with 5,051 miter ANDs went from 327 ms to 307–310 ms; C previously took 222 ms. Network checking was within the prior range at 324 µs versus 318–325 µs. The base/abc, base/abci, base/io, bool/dec and map/scl race tests passed.

The full-module snapshot containing cached BDD abstraction and reusable trace replay passed all 130 packages with tests, including GIA in 463 seconds.

### Retire collected decision handles and group short-lived allocations

Four red correctness tests exposed dangling Go handles after explicit collection, reordering, variable deletion and bounded rollback. A collected BDD could revive references to its old children. A rolled-back handle could enter a manager after its numeric ID had been reused. Releasing a newly created key before rollback also dropped two unrelated external pins from one to zero. The manager now detaches both handle phases and their child links at retirement, and accounts for child references once.

The pressure fixture previously allocated 541 objects for a batch containing 448 released keys. Allocating every decision pair in groups of 31 brought that down to 111. After collecting 4,096 keys with one survivor every 64, however, each manager retained 287 KB versus 33 KB with individual allocations. Unconditional groups of four still retained 58 KB. The selected version starts a four-pair group only when the most recently issued pair has released its children. A 40-manager comparison retained 33.2 KB in the strided case and 41.5 KB after six queens, versus 33.0 KB and 40.5 KB for individual pairs. The retained-heap benchmark in cuddNodeLifetime_test.go reproduces both workloads with a fixed 40 iterations.

Pressure collection improved from 83.1–84.0 µs and 541 allocations to 76.7–79.1 µs and 205 allocations, with 115,200 allocated bytes in each case. C took 35.2 µs. Six queens took 3.76–3.79 ms versus 3.67–3.72 ms in this comparison, with allocations falling from 10,396 to 10,308. The six-bit reachability solver took 1.46–1.53 ms partitioned and 1.02–1.07 ms monolithic. The thirteen BDD packages, base/abci and misc/extra passed their race tests.

### Index network ANDs by the two phased object IDs

Hashing pairs of Go edges used two pointers and two padded Boolean flags per key. The index now stores two full-width integers. Inverting all bits of an ID records its complement phase. The cached result still checks the live object's exact fanins. Tests cover all four phases, commutation, maximum-width IDs, and duplicate order after compacting a network with an ID hole.

A pair of 31-bit IDs packed into one uint64 gave the fastest local lookup at 36–39 ns, but discarded upper bits of the Go ID. Full-width keys measured 46–50 ns, down from 60–65 ns for pointer pairs. A paired i10 run reduced network checking from 321–326 µs to 303–315 µs. CEC stayed within the previous range at 315 ms versus 313–316 ms, with about 0.7 MB less allocated per solve. The base/abc, base/abci, base/io, bool/dec and map/scl race tests passed.

### Preserve AIGER reset and output-property transformations

Six red tests found divergent behavior in the legacy reader and duplication functions. Removing an invariant left its private AND dangling. An unused unknown-initialized register created three unnecessary ANDs. Reset conversion and named constraints both merged duplicates with structural hashing disabled. Direct constraint duplication panicked on the native invariant type. The reset routine left stale fanout marks, and combining named and explicit properties lost the bad-output count. A seventh regression checks raw opposite-phase pins on a skipped AND during reset conversion.

The legacy reader now uses the shared constraint and zero-initialization routines. It cleans dangling logic after filtering named properties when hashing is enabled. Header constraints keep their polarity and bad outputs keep their class. The conversion still creates the extra initial-choice PI and reset register for an unused unknown state, as in C. The native fixture checks all three reader modes for a skipped invariant, the unused register case and both duplicate-gate transformations.

The reset duplication scan now marks drivers in place and clears its scratch marks as the original does. On a 4,096-AND chain with three registers and a mixture of reset types, this reduced time from 367–374 µs to 227–237 µs, memory from 456,864 to 372,880 bytes and allocations from 8,225 to 28. Native C took 106 µs, with 29% of samples in AND append and 28% in driver marking. The targeted GIA race tests passed. The preceding full-module snapshot with decision-handle retirement and network hash preservation passed all 130 packages with tests, including GIA in 466 seconds.

### Read AIGER records into reusable storage

Two red tests found a short output-name vector on a sequential graph and an accepted duplicate symbol when the first label was empty. The name vector now covers every combinational output, including register inputs, while user PO IDs still cover primary outputs only. The symbol strings share an owned copy of the symbol section. Tests check that names survive mutation of the caller's input and that an empty first name still reserves its index.

Reading 1,024 unnamed decimal output records previously allocated 2,078 objects. The parser now reads bounded numeric records directly from the input into a three-integer buffer and preserves whitespace, leading zeros, signed-edge phases and CRLF input. One- and two-byte gate deltas have direct decode paths. Tests cover the transitions across all five uint32 encoding widths and cursor position on decode failure.

In matched 4,000-read i10 trials the previous reader took 355–369 µs and 1,683 allocations. Decimal scratch gave 344 µs and 1,007 allocations. Adding the common delta paths gave 300 µs. Shared symbol text and lazy user-ID vectors brought the complete reader to 301–305 µs and 525 allocations. The release C driver took 142 µs in a fresh run. Its separate profile put 33% of samples in structural hash lookup and 24% in the dangling-node check. The GIA AIGER tests passed with the race detector.

### Sweep explicit CUDD collections in variable order

The explicit collector previously constructed incoming-arc and reachability maps and a DFS stack on each call. Collecting a fully pinned 1,024-key table failed a zero-scratch regression at 29 allocations. A second red test showed that collecting an unrelated node discarded a cached nonconstant query with live operands. C's cache explicitly preserves its distinguished nonconstant result.

The sweep now visits each family's current variable order, including shuffled orders, and visits shared constants last. Once unowned parents release their references, each node can determine whether it survives from its remaining references and a mark passed down from surviving roots. The mark fits in existing node padding. A twelve-seed independent traversal compares every surviving BDD, ADD, ZDD and terminal handle against reachability from explicit pins, manager roots, a nonconstant background and a saturated node.

With 4,096 keys, freeing 3,072 and retaining 1,024 took 479–487 µs and 119,680 bytes of scratch before the change. The ordered sweep took 100–107 µs and zero allocated bytes. The native reference took 193 µs for the same release pattern. Both timers exclude graph construction. The whole-process native profile spent 34% in collection and 49% initializing managers. The thirteen BDD packages, base/abci and misc/extra passed their race tests. Testing shuffled universes also exposed incorrect reference ownership in the existing ZDD swap routine on both collector revisions.

### Keep the GIA output-type validation contract

The full GIA run exposed two failures following the invariant change. The earlier rejection test classified type 2 as invalid, although the native AIGER reader emits it for invariants. Moving that case to type 3 reproduced a real input-validation regression alongside the existing negative case. Constraint duplication now validates the complete vector before copying and accepts the three types produced by the reader. Both invalid-input cases and the independent type-2 output and register test pass under the race detector. The earlier whole-package race run also reached its ten-minute timeout, so the rerun has a longer timeout.

### Move only dependent ZDD keys during adjacent swaps

Reversing ten variables previously rebuilt the entire unique table 45 times. A Go profile attributed 74% of CPU samples to rebuilding tables. The swap now visits the upper subtable, removes the keys whose children depend on the lower variable, and interns the replacement nodes against the surviving keys. It transfers references before releasing old children and updates each handle's complementary index.

Red tests exposed spurious external references after a shuffle, changed explicit pins, and stale complementary indices. A six-variable universe retained 17 decision keys after collection. It now retains exactly six and releases all six after dropping its root and the manager's universe. A native C fixture checks the same ownership transitions. Further red tests showed that an identity shuffle left cached answers behind, reordering retained explicitly released keys from both families, and the minimum-size test counted a released node. Preprocessing now flushes the cache and collects released keys while preserving active Go temporaries. The size threshold counts live ZDD keys.

Building all 120 three-element sets over ten variables and reversing the order took 750–758 µs, 995,256 bytes and 2,263 allocations before this change. The complete replacement took 69–75 µs, 70,520 bytes and 445 allocations. Release C took 38.6 µs for the same build and shuffle. Its process profile put 69% in manager initialization and 20% in adjacent swaps. All thirteen BDD packages, base/abci and misc/extra passed race testing.

### Pack GIA's six single-bit fields into bytes

A GIA object now takes 40 bytes on amd64, down from 56. Its two backward edge distances, object ID and scratch literal retain their original widths. A size regression first failed at 56 bytes. A behavioral test exercises fanins more than 512 objects apart, all four complemented-fanin combinations, independent mark and phase updates, a negative scratch value and stable complemented handles.

In paired runs on the same CPU, constructing and querying a 4,096-AND hash chain fell from 350–359 to 306–325 µs and from 474,064 to 387,504 bytes. Reset duplication on the same chain length fell from 256–272 to 205–215 µs and from 372,880 to 279,152 bytes. The i10 read dropped from 311,720 to 244,248 bytes. Its timing ranges overlapped at 313–358 and 321–356 µs. Native reference times from the corresponding workloads were 123, 106 and 125–142 µs respectively. Targeted flag, property, and fanin tests passed under the race detector. The full module on the preceding flag snapshot reported only the two subsequently fixed property-validation cases.

### Align existing ZDD variables with the native shuffle

The previous BDD-to-ZDD alignment reconstructed every family in a scratch manager, even for identical orders. Red tests found missing complementary handles, phantom external pins and 14 surviving keys where only the eight-key universe remained reachable. Repeating an identity alignment also skipped the native shuffle's garbage collection hook. Both C and Go fixtures now grow four ZDD variables to eight, reorder two blocks, test all 256 sets against two pinned original covers and the new universe, then check collection.

The implementation now builds an empty universe in its final order. For an existing universe it extends the table as needed and applies the normal adjacent shuffle. Existing single-variable groups use the BDD order directly. Repeatedly aligning a twelve-variable manager containing 66 externally pinned pair roots dropped from 208–214 µs, about 150 KB and 221 allocations to 527–583 ns and zero allocations. The old Go profile spent half its samples in map assignment. C took 714 ns, with 94% of its samples in cache flushing. The thirteen BDD packages, base/abci and misc/extra passed race testing, and the final zero-allocation revision passed the full core race suite.

### Reuse phased edge values in decomposition and rewriting

Profiles of the i10 equivalence proof showed repeated heap boxing in decomposition lookup, rewrite lookup and complementation. Six repeated queries on an already built AND allocated seven times, including a copy of the entire PI list. The red allocation test now passes at zero. Each queried network object lazily stores immutable Go interface values for its two edge phases. Both keep the original dynamic Edge type. The PI lookup accesses the existing ordered index. Tests check renumbering, both phases, concurrent first queries and zero edges.

Paired fourteen-proof trials gave 307–315 ms and about 986,130 allocations for the previous implementation, versus 298–299 ms and 692,713 allocations with the shared pair. Bytes fell from 75.1 to 70.6 MB per proof. The rewrite pipeline kept the same 2,004 ANDs and 36 levels. A separate 30-proof profile measured 305 ms, against 326 ms in the preceding profile. Propagation used 24% of CPU, and cut allocation and leaf merging together accounted for 23% of allocation samples. The native i10 proof previously took 222 ms.

An alternative with two atomic interface slots inside every object was also measured with the subsequent cut-storage prototype. It saved another 10,071 allocations per proof while adding about 170 KB per proof, and timings overlapped. The shared pair keeps the smaller per-object footprint for networks that never enter decomposition. The base/abc, bool/dec and opt/rwr race tests passed. The full module with the corrected GIA property validation and the shared edge pair passed all 130 packages with tests.

### Store common cut leaves and truth with their handle

Merging and recycling a four-leaf cut formerly allocated three objects. The Go profile spent 47% of sampled CPU in cut allocation and 26% in leaf merging. Sorted unions now use a stack buffer until the merge succeeds. Each cut can hold four leaves and one truth word in its own allocation. Wider cuts use separate storage. A shared lifetime token enforces ownership and rejects double recycling without registering every cut in a map. A second red test found a stopped manager retained the entire input network inventory. Stopping now releases the inventory and its fanout counts.

Tests check independent leaf and truth storage for each supported cut width from three through twelve, rejection of over-wide unions, recycled and foreign handles, and stopped managers. Paired 20-million-merge runs fell from 172–173 ns, 132 bytes and three allocations to 71–72 ns, 128 bytes and one allocation. The native fixed-size allocator took 12 ns for the same sorted inputs, with 50% of its samples in merging and 17% in the allocator's free-list fetch.

On top of shared decomposition edge values, the i10 read/balance/rewrite pipeline took 55–57 ms and 160,688 allocations, down from 61–76 ms and 212,051 allocations. CEC fell from 302–303 to 290–291 ms, 70.5 to 67.6 MB and about 691,541 to 547,219 allocations. The cut, rewrite and base/abci race tests passed. The full GIA race suite also completed successfully in 767 seconds after the earlier property validation and object-layout changes.

### Preserve ZDD coordinate transforms and score live keys

The adjacent linear transform changed membership incorrectly for a pinned family that bypassed the upper variable. It also mishandled an ancestor pointing into the lower subtable. The regression pins all sixteen two-variable families with a nonconstant suffix, checks each of sixteen assignments after one and two transforms, checks every external reference, and releases everything back to the four-key universe. The C fixture verifies the same coordinate mapping. The new two-level rewrite reserves the special lower nodes first, then resolves new cofactors for both layers. It acquires new arcs before releasing the old ones and reuses the two subtable probe arrays.

Two applications over sixteen pinned families took 3.07–3.11 µs and sixteen allocations on the incorrect implementation. The first correct implementation took 5.65–5.86 µs and fourteen allocations. Keeping the probe arrays and sizing local rewrite storage brought this to 1.90–2.05 µs with zero allocations. Native C took 358 ns. A zero-allocation regression and an independent node census run alongside the truth tests.

A second pair of red tests showed a six-key universe scoring 17 in the reorder heuristic and adjacent swaps leaving logical subtable widths attached to the old levels. Eagerly retiring the eleven obsolete handles broke an existing exhaustive test that retains Go handles through a swap. The score now uses a maintained count of released ZDD keys. This preserves the handles and yields the same live size as C. Public node-count queries on 306 live nodes fell from 2.15–2.16 µs to 1.88–1.95 ns, versus 1.43 ns in C. Repeated twelve-variable alignment also fell to 90–94 ns because preprocessing can skip an all-live table. The thirteen BDD packages, base/abci and misc/extra passed race testing.

### Attach both AND pins in one validated operation

Recreating an AND retained its previous level of 57 when the new inputs required level 5. The raw manager constructors also reordered equal-ID pins when only the first was inverted. Three more red cases found partial mutations when the second driver was an illegal PO, including a newly allocated object, an attached first fanout and a changed hash-exclusion flag. The C fixture confirms level recomputation, equal-ID pin order and duplicate fanout multiplicity on valid inputs.

Both constructors now validate their inputs before mutation. The shared attachment operation allocates both pins together, adds each fanout occurrence, and sets the level and phase from the new inputs. Explicit manager insertion preserves the native last-inserted hash head. On a 1,024-AND chain, paired Go builds fell from 603–613 to 475–487 µs, 671,288 to 622,136 bytes and 8,272 to 5,200 allocations. The native construction and teardown took 265 µs. The old Go profile spent 22% of CPU inside slice growth. Native profiling put 82% of samples in fixed-entry allocation.

The i10 rewrite pipeline kept 2,004 ANDs and depth 36, with 141,235 allocations versus 160,688. The fixed proof kept the 5,051-AND miter and used 503,026 allocations versus roughly 547,216, with about 0.7 MB less allocated. Its 291–304 ms timing overlapped the prior 291–314 ms. The base/abc, base/abci, base/io, bool/dec, opt/rwr and map/scl race suites passed.

### SAT assignment storage and frequency tracking

Native ABC stores assignments, saved polarities and analysis tags in three bytes per variable. A red storage test found twelve bytes in the Go port. Both CDCL solvers now share byte slices for these fields. A growth regression checks that large clause-reason handles, saved polarity and tags survive a resize. Another red test caught direct, conflicting and propagated enqueues overwriting existing frequency bytes of 23, 255 and 17 with 1. The shared enqueue now sets only previously zero frequencies, as in C.

Pinned, interleaved runs of 560 iterations each took 3.44–3.53 ms before and 2.15–2.26 ms after for allocating, initializing and deleting a 65,536-variable solver. Go allocations fell from 8,681,416 to 8,091,592 bytes, exactly nine bytes per variable. The equivalent native fixture took 8.32 ms, with 78% of sampled CPU in setup and 22% in deletion. The Go setup and pigeonhole profile showed 21% in heap update, 13% directly in propagation and 9% in clearing memory. The pigeonhole solve preserved 604 conflicts, 788 decisions and 7,409 propagations, taking 3.07–3.11 ms before and 3.08–3.12 ms after. Native took 2.10–2.17 ms. The full module passed after the related AND changes, including the GIA package in 461 s.

### Match the native ZDD linear sifting path

Four five-variable covers initially gave different final permutations, coordinate truth tables or heap sizes from C. The sifter was looking up a variable by its starting level again after intervening swaps. It now ranks variable IDs by their initial subtable populations and looks up each variable's current level when its turn arrives. A race test also found that independent managers shared a swap counter. Each manager now owns its counter. The native fixture and all four fixed Go cases agree on the full 32-bit membership bitmap, permutation and live key count.

Recording moves chronologically allows the sifter to reuse the same array when walking backward. Wide transforms reuse their rewrite workspace and clear its child handles after each operation. A regression measured two allocations per warm transform pair before this change and zero afterward. On a complete build, shuffle and sift of the four covers in rotation, workspace reuse reduced 640 allocations and 232,090 bytes to 460 allocations and 198,336 bytes. Two pinned 2,400-operation samples took 331–336 µs before and 331–341 µs after. Native took 39.77 µs, with 63% of samples in manager initialization. The Go allocation profile put 42% in subtable resizing, 26% in rewrite workspace growth and 25% in node-pair creation. All thirteen BDD packages passed race testing. The base/abci and misc/extra race suites passed on retry.

### Retire unreachable linear-sifting trials

A five-variable direct sift left 593 newly created, unreachable trial nodes in the unique tables. An independent traversal starting at every preexisting Go handle and every still-live node identified the unused keys. The sifter now sweeps in variable order when released keys exceed a quarter of resident keys, and once at the end. Existing handles and their entire transitive child graphs survive, including handles that currently own no references. The four native permutation, membership and live-size oracles still pass.

In interleaved runs of 3,200 complete builds, shuffles and sifts, Go fell from 332–335 to 118–121 µs. Allocated bytes fell from 198,336 to 68,640 and allocations fell from 460 to 310. Native C took 39.77 µs. The remaining Go allocations come mainly from node pairs (46%) and subtable resizing (34%). The CUDD core passed its full normal and race suites.

### Reserve the decision heap during bulk SAT growth

The remaining per-variable load flag also occupied four bytes, versus one in the C structure. Its regression now checks storage, restart, shrink and regrowth into the former allocation. The growth routine reserves the decision heap for the requested range, clears reused metadata in contiguous slices and skips heap repairs when equal scores are appended to an empty heap. A separate reference insertion test checks every heap entry and inverse position across all three activity modes after old variables have left the heap or had their scores reduced.

Two pinned runs of 600 operations each took 2.21–2.29 ms before and 1.28–1.44 ms after for 65,536 variables. Bytes fell from 8,091,592 to 7,145,120 and allocations from 51 to 27. The native setup took 8.32 ms. The pigeonhole search kept its native counters, reduced six allocations, and ranged from 3.07 to 3.69 ms under concurrent builds. The i10 proof kept its 5,051-AND miter and took 283–287 ms versus 284 ms before. Race suites passed for bsat, pdr, base/abci and the CUDD core.

### Share CUDD's portable random stream across both Go APIs

The public CUDD utility package kept a second copy of the generator used by the reorder core. Seeding the public API and reading from the core produced 744,052,179 where native seed 1 produces 1,902,191,391. A second red test reported races when eight goroutines drew from the generator. Both APIs now use one synchronized copy of CUDD's portable recurrence and shuffle table. The regressions check an interleaved native sequence and compare the entire multiset of 3,200 concurrent draws against a serial run. All thirteen BDD packages and misc/extra passed race testing.

### Print disjoint minterm rows with native terminal precision

A reentrant output writer printed a second manager's cover while the first was in progress. The first manager then emitted an extra, false minterm because its zero and background terminals had been overwritten globally. The second regression found that 12,345.6789 printed with all its digits, where C prints 12345.7. Finite values now use six significant digits and IEEE infinity spellings match the native IEEE build. The stock C build clamps infinity to ±1e302, as documented in the existing ADD numeric oracle. Linking the two native CUDD table and initialization objects with HAVE_IEEE_754 verified all five Go terminal strings.

The traversal keeps its own ASCII path and emits a complete row with one writer call. Two pinned, interleaved 14,000-operation runs on eight-variable parity (128 rows and 1,536 bytes) fell from 71.82–71.89 to 5.00–5.06 µs. Allocations fell from 129 to one, 1,088 to 48 bytes. The matching native IEEE formatter took 25.62 µs to an in-memory stream. The previous Go profile spent 27% of sampled CPU in the formatting engine's doPrint, and native gprof attributed all its non-library samples to the recursive formatter. The CUDD utility and extra BDD race suites passed. A full-module run after resuming also passed all 130 tested packages, including GIA in 466 s.

### Mark fanin cones by object ID during traversal and replacement

In the i10 proof allocation profile, 83 MB of 1.55 GB across 24 proofs went to the pointer map used to mark fanin cones. AIG replacement only needs to know whether the target cone is valid and contains the old node. It now reuses byte marks with a fresh generation for each check. The read-only public traversal owns its marks per call and uses sparse pages for small root sets in large ID spaces. It still emits both terminals and internal nodes in fanin postorder. Tests cover holes, sequential feedback through a box output, combinational cycles, invalid roots, concurrent traversals and generation wrap. A warm rejected replacement fell from three allocations to zero. Separate red tests showed that removed fanins, fanouts, terminals and boxes still occupied backing-array slots. Deletion now clears the vacated pointers and visits the relevant object-kind lists.

On a 1,024-AND chain, the pointer map took 154–156 µs and 104,080 bytes. Direct marks took 27–28 µs and 30,968 bytes. Sizing the result from a bounded root-depth estimate brought this to 15.42–15.61 µs, 10,624 bytes and two allocations. Native Abc_NtkDfsNodes took 16.94 µs and emits the 1,024 internal nodes. Go additionally emits both inputs and detects cycles. The old Go profile attributed 29% of cumulative CPU to map lookups. Native gprof attributed 69% to DFS and 28% to traversal-ID checks.

The i10 proof used about 62.04 MB and 493,686 allocations versus 66.70–66.73 MB and 503,013–503,016 before. Its 272–277 ms range overlapped the previous 276–285 ms. The rewrite retained 2,004 ANDs at depth 36 and used 1.54 MB less. Race suites passed for base/abc, base/abci, base/io, bool/dec and map/scl.

### Filter cut frontiers before spending the last available slot

A sequential node with one free slot accepted a cut it had already found, reached its cutoff and missed the next distinct cut. The network driver could then report convergence with a valid three-input root cut missing. Both entry points now check the old and current frontiers during enumeration, remove dominated supersets and make the recovered slots available to later candidates. Choice merging follows the native source-list order for sublimit cuts and defers the full-width suffixes. The native fixture verifies ordering, representative simulation phase, pruning and the last-slot case.

Leaf-count bins keep tail pointers to their last elements. A regression showed that appending to a copy of an empty bin wrote into the original. Merged candidates now acquire owned storage after filtering while preserving logical allocation and recycling statistics. Independent sequential inputs can be shifted in place and restored. Shifted frontiers sharing a list suffix must be copied. A red regression found that the common leaf received both input shifts. Additional tests exercise occupied destinations and overflow halfway through a list.

Two pinned runs took 391–408 ns and 128 bytes in one allocation per complete two-candidate sequential node, down from 714 ns and 864 bytes in six allocations in the first correct implementation. The native C fixture took 1.20–1.38 µs and includes ABC's per-merge timing calls. A rejected four-input merge fell from 106–112 ns and 128 bytes to 44–48 ns and zero allocations, versus 34.49 ns in C. On i10, the full proof retained its 5,051-AND miter and used 60.00 MB and 481,534 allocations, down from 61.97 MB and about 491,454. The interleaved proof runs varied between 295 and 340 ms across both versions, so they establish the memory savings but no end-to-end speedup. The final shared-suffix regression and cut package passed.

### Evaluate four-input rewrite cuts directly from the cut manager

The i10 allocation profile charged about 2.3 MB per proof to copying linked cuts into growing slices. The concrete entry now walks the cut manager's list in order, and the public slice entry uses the same evaluation loop. A red callback test showed that the old loop read fanout counts for the valid prefix of a cut containing a deleted leaf. Native ABC first completes the leaf lookup and ignores the entire bad cut. The concrete network now reuses a four-entry manager buffer, clears it on return and preserves separate buffers for caller-provided adapters.

The 16-input native fixture reports five valid cuts. Without shared fanouts all five are skipped. Pinning the four middle nodes produces 15 graph evaluations, zero gain and complemented inputs 25, 26, 27 and 28, all matched in Go. In two interleaved two-million-node runs the unpinned Go case fell from 749–837 ns and 520 bytes in five allocations to 368–376 ns with zero allocations. Native Rwr_NodeRewrite measured 2.22 µs of process CPU with its internal clock calls enabled. The pinned native case took 4.94 µs.

The 20-proof i10 runs used 57.05–57.07 MB and about 450,142 allocations, down from 60.01–60.03 MB and 482,130. They took 271–294 ms versus 279–321 ms and retained the 5,051-AND miter. The full suite from the preceding cut checkpoint passed all 130 tested packages. The rewrite, cut and network orchestration packages passed targeted tests, with native comparisons covering both scoring and rejection.

### Evaluate decomposition truth tables a word at a time

Dec_GraphDeriveTruth copied the graph and evaluated 32 separate assignments. Both DEC graph representations now propagate a 32-bit truth word per gate in topological order. Small graphs use stack storage. Tests compare the result to all 32 scalar evaluations across graph sizes up to 704 nodes, including complemented and nonfinal roots. They also check that opaque node payloads survive and that an invalid forward edge is rejected. KIT already used word arithmetic but wrote every word twice. A red test found that even a literal truth query erased the stored functions of all four leaves. Native KIT returns a literal directly. Its Go evaluator now uses the existing intermediate words for internal graphs and preserves literal payloads.

On a four-input, seven-gate decomposition, the old profile spent 76% of CPU in scalar evaluation. Two pinned runs took 904–906 ns and 352 bytes per truth. A word evaluator using a callback took 39 ns, and reading the actual node type during the traversal took 30.2–30.5 ns, both with zero allocations. The native fixture took 18.86 ns for the same 0x333f333f truth, with all sampled CPU in Dec_GraphDeriveTruth. Its literal side-effect check also passed. The full i10 proof used 56.19–56.23 MB and about 448,590 allocations versus 57.08–57.10 MB and 450,870. These pairs took 270–271 and 264–267 ms respectively, so the isolated gain did not shorten this proof. DEC, KIT, RWR, RWT and base/abci passed their unit tests.

### Keep SAT literals and proof identifiers independently addressable

Appending a literal to the slice returned by clause_begin or clause_end overwrote the hidden proof ID. A second red regression found that assigning partition 2 to a proof clause or variable read back 2. The native one-bit field reads back 0. The Go header now stores its ID separately, its literal slice covers exactly the literals, and both partition setters retain the low bit. The header uses byte fields for the flags and LBD, and its four inline literal slots now fit a four-literal clause. Native handle offsets, stream IDs and compaction accounting still use ABC word sizes. Additional regressions caught signed overflow in word counts and a page transition wrapping onto a live handle. Both failures now precede any allocation or state change.

Two pinned runs inserting 1,024 mixed clauses of two to eight literals fell from 99–103 µs and 203,568 bytes in 1,501 allocations to 82–83 µs and 184,464 bytes in 1,209 allocations. The matching native fixture took 5.03 µs and checked seven partition assignments. The pigeonhole proof retained its 604 conflicts, 788 decisions and 7,409 propagations. It used 2,240 bytes and 21 allocations less, with overlapping timing ranges around 3.0–3.3 ms. The i10 proof used about 3,145 fewer allocations with little change in bytes or elapsed time. A census of 12,672 actual clause insertions during i10 found 5,029 length-three clauses, 1,562 length-four, 1,057 length-five and 327 of length 32 or more. This provides a distribution for testing larger clause blocks. The preceding DEC and KIT changes also passed the final race runs for DEC, KIT, RWR, RWT and base/abci.

### Index SAT clause headers by occupied blocks of the native page

A regression placed a three-literal clause after 65,536 literals. Growing the flat pointer table for that single new header used 270,356 bytes. The new index allocates blocks of 256 pointer slots only where a clause starts. The first block grows with actual use so a fresh solver can store its binary scratch clause cheaply. Clauses of five through 128 literals share batches of headers and integers with exact-capacity slices. Larger clauses own their payload directly. Tests retain individual pointers through allocation, rollback and restart. Two further red tests showed that shrinking into the middle of a clause corrupted the cursor and that a stale learned bookmark deleted problem clauses before failing. Rollback verifies both stream boundaries before shrinking either stream.

On 1,024 mixed clauses, fixed inline capacities of eight and sixteen cut allocations from 1,209 to 59 but raised allocated bytes to 198 KB and 222 KB. Growing the flat pointer index by doubling used 158 KB. Separate batches of header and literal storage with a segmented index gave the best combination. The version with a small initial block took 39.1–39.7 µs and 121,456 bytes in 93 allocations, versus 85.0–87.9 µs and 184,464 bytes in 1,209 allocations. Native took 5.03 µs. The seven-pigeon proof fell from 297,144 bytes and 1,787 allocations to 184,048 bytes and 671 allocations with identical search counters. Its elapsed ranges overlapped around 3.1–3.2 ms.

In two interleaved 18-proof i10 runs, allocated bytes fell from 56.13 MB to 54.44 MB and allocations from about 445,449 to 433,920. Timing remained within the shared machine's variation at 274–280 ms before and 278–282 ms after. Checking every header lookup for an empty index block slowed the seven-pigeon proof. The accepted implementation performs those extra checks when validating a shrink boundary and keeps direct indexing on the propagation path. The full suite at the earlier linked-cut rewrite checkpoint passed all 130 tested packages.

### Compare propagation costs and refresh the native i10 baseline

Interleaving the native and Go executables on the same CPU for two runs of 16 proofs each gave 213.95–217.96 ms in C and 263.27–263.37 ms in Go. Both proved equivalence on the 5,051-AND miter. Whole-process CPU including initialization was 216–220 ms in C and 264 ms in Go. Native gprof put 44.8% of its flat samples in propagation, 8.1% in solver search and 4.7% in node rewriting.

The Go seven-pigeon profile put 67% of cumulative CPU in propagation. Four isolated variants preserved every native decision, conflict and propagation count. Direct indexing of a known clause handle took 3.01–3.02 ms, compacting the local watch slice took 3.00 ms, and a population count of the 32-bit decision-level mask took 2.97–2.98 ms, compared with 3.04 ms initially. Together they took 2.92–2.93 ms in that interleaved run. Tests cover collisions at bit 31 across 96 levels and a watch move followed by a conflict with a pending implication. Additional trials that skipped redundant writes, cached page directory fields or changed the candidate comparison gave overlapping results and were left out.

The production version shares handle decoding with the public clause accessor. A second, busier run put it at 3.16–3.17 ms and the prior version at 3.07–3.21 ms. The i10 ranges overlapped at 274–281 ms and 279–280 ms respectively. The evidence establishes the isolated improvement and exact counters, while a full-proof timing gain remains unconfirmed. Race suites passed for BSAT, FRA combination, PDR and base/abci.

### Reuse structural AND lookups while scoring a single rewrite root

Several candidate graphs ask for the same phased pair of inputs. A 256-entry table on the concrete rewrite manager now records the object ID, output phase and intrinsic level for each match, as well as misses. A new generation begins at every root. Cut boundaries still relabel the deletion cone before each evaluation. The table stores integers and clears its network reference when the call returns. Tests compare all four phases, swapped operands, constants and colliding keys against the network lookup. They also cover stale misses after insertion, generation wrap and identical input IDs on two networks.

On the 16-input tree with four pinned leaves, the same five cuts yielded 15 candidates and the same four complemented winning leaves in both ports. Over 250,000 calls, the previous Go implementation took 2.57–2.58 µs per root, the new implementation 2.13–2.19 µs, and C took 4.90 µs. Both Go implementations performed zero allocations per warm call. On i10, a trial caching only misses took 45.3–45.6 ms per balance and rewrite versus 46.2–46.4 ms initially. Caching matches as well took 43.7–43.8 ms with 64 entries and 43.1–43.2 ms with 256. The production version allocated the table on the first concrete rewrite, adding 10,880 bytes and one allocation per pass. It took 43.1–43.4 ms versus 45.4–45.8 ms, producing 2,004 ANDs at 36 levels in every run. Interleaved batches of 18 full i10 equivalence proofs took 262.2–268.6 ms after the change and 269.9–275.6 ms before it. Both retained the 5,051-AND miter. The RWR, RWT, DEC, KIT and base/abci race suites passed.

### Match native decomposition field widths and measure graph construction

Two red tests reproduced an error in how DEC scores deep graphs. ABC stores a leaf's level in a 14-bit field before using it. It checks a prospective gate's full level against the limit, then stores the accepted level in the same width. Go kept 16,384 through both stores and rejected two legal rewrites. The regression probes both boundaries and a rejection at a limit of 16,383. A separate red test found that KIT serialized a 31st node-index bit, including for negative indices. Its edge constructor now retains the native 30-bit node index. The C fixture verifies the same edge words and all three scoring outcomes. DEC, KIT and base/abci passed their normal and race suites.

A one-proof census of i10 counted 4,795 DEC graphs and 2,517 KIT graphs. Of these, 901 in each package had two leaves and one gate. DEC also created 2,512 four-leaf graphs, primarily through the rewriting library. On four leaves and seven ANDs the initial Go builders took 387–403 ns with 1,360 bytes and 12 allocations. At 64 ANDs they took 2.31–2.42 µs with 6,944 bytes and 70 allocations. Native construction took 140 ns and 330 ns for DEC, and 115 ns and 515 ns for KIT, on these two sizes.

### Build the exact DEC topology when converting a KIT refactoring graph

The native refactorer reads the KIT result directly through the compatible C graph layout. A red Go test showed that copying its gates as ANDs changed an OR's printing flags from [1 1 0] to [0 0 1]. Selecting a literal as the output also discarded every internal gate. The conversion now copies the original operation flags and phased edges, retains the complete topology and sets the selected root independently. The destination constructor allocates exactly the known node count in blocks of at most 32, with an indexed pointer directory. Tests hold nodes across block boundaries, subsequent appends and graph release. Invalid dimensions are rejected before allocation.

For a four-leaf, seven-gate copy, two pinned batches of 500,000 calls fell from 464–481 ns and 1,408 bytes in 13 allocations to 288–303 ns and 1,040 bytes in three. Full i10 equivalence used 53.38–53.39 MB and about 412,539 allocations per proof versus 54.48–54.52 MB and about 433,925. Timing in these batches fell from 260.6–261.6 ms to 259.2–259.4 ms. DEC, KIT and base/abci passed their normal and race suites.

Separate trials on graphs with unknown final size tested a smaller pointer directory, compact printing fields, and three initial block sizes. The best seven-gate construction result was 290–343 ns and three allocations. For a full proof, the compact individually allocated variant used 50.4 MB. Blocks used 51.9–52.2 MB and removed roughly 58,500–64,100 allocations. Elapsed time varied across the machine, including 247–283 ms for the same compact variant, so these general-constructor trials remain experimental. The initial DEC construction profile attributed 70% of cumulative time to small scanned allocations.

The full package suite passed at the preceding field-width checkpoint, covering 130 tested packages. A fresh native comparison after the graph conversion interleaved two 16-proof batches per implementation on CPU 12. C took 209.76–213.59 ms per proof and Go took 242.05–249.64 ms. Each implementation proved equivalence on the 5,051-AND miter. Including process startup gave 212.0–216.0 ms for C and 243.0–250.7 ms for Go. The Go proof remains about 15–17% slower in these batches.

### Remove logic nodes without shifting the remaining directory

The 24-proof Go profile put 6.1% of cumulative CPU in removing an object from an ordered pointer slice. A comparison on 4,096 nodes showed that binary-searching the deletion position still cost 6.59 ms per network, almost identical to linear search. The repeated pointer moves dominated. The logic-node directory now records each node's slot, clears the slot on deletion, and compacts surviving entries in order when at least half the slots are empty. Object ID reassignment rebuilds the slots. Counts account for holes immediately, and callers reading the public list receive independent ordered snapshots. The AIG cleanup pass now reads its before and after counts directly.

With matching public creation, deletion and teardown operations, creating 4,096 nodes, deleting alternate IDs and checking all 2,048 survivors took 6.64 ms before and 844–853 µs after, with unchanged allocated bytes. C took 98.33 µs. Tests cover repeated holes and compaction with interleaved terminals, new nodes after deletion, snapshot independence, and deleting a gate after ID reassignment. The base/abc, base/abci, RWR and RWT race suites passed.

On i10, interleaving the original implementation, binary search and indexed deletion gave 244–255 ms, 244–246 ms and 238.2–238.3 ms respectively. A busier pair of runs after eliminating the two count snapshots gave 243–260 ms versus 278–291 ms originally. Allocated bytes fell from 53.37–53.39 MB to 52.97–52.99 MB. The proof still established equivalence on 5,051 miter ANDs.

### Keep temporary sequential cut copies outside the cut inventory

A red test duplicated sixteen live cuts. Go reported 32 allocated, 32 current and a peak of 32. Releasing the copy left 32 allocations and 16 recycles in the statistics. C duplicates and releases temporary input lists directly through the fixed allocator. Its fixture retained 16 allocated, 16 current, a peak of 16 and zero recycles throughout a million snapshots. Go now distinguishes these temporary handles from the regular cuts used as buffer and oracle outputs. Recycling still checks ownership and detects a repeated release.

The temporary list owns its leaf and truth storage and allocates handles in blocks of at most 32. Tests cover 85 cuts spanning multiple blocks, including four-word truths, long boundaries, flags, indices and proof references. Two pinned runs copying and releasing sixteen cuts took 526–631 ns and one allocation versus 980–1,003 ns and sixteen allocations. Both allocated 2,048 bytes per list. Native took 119 ns. The cut, RWR and base/abci normal and race suites passed.

### Share four-input truth expansion between CUT and DAR

The i10 profile attributed 4.3% of cumulative CPU to computing cut truths. DAR already tabulates the placement of a truth table with up to four inputs. CUT now uses that same KIT helper for small merged supports, including cuts that reserve more truth words than their actual support needs. A red in-place merge test produced 0xaaaaaaaa instead of 0x88888888 when storing the result into its right operand. Native C copied both inputs before writing the result. Go now reads the right operand before overwriting the output. The same in-place case passed in both implementations at widths 4, 7 and 12. The established Go choice-polarity and XOR tests also passed.

At four inputs the internal Go merge fell from 65.3–66.3 ns to 12.6–12.7 ns with zero allocations. The public Go call, which also accumulates truth-computation time, fell from 136–137 ns to 81.8–83.1 ns. The native public function took 33.0 ns. Seven-input Go merges took 165–166 ns versus 159 ns before, and twelve-input merges took 2.85–2.90 µs versus about 3.0 µs. Native measured 109 ns and 2.81 µs at those widths. All three native truth hashes matched. Interleaved batches of 18 i10 proofs fell from 245.1–246.3 ms to 237.3–240.4 ms with the same miter size and allocation counts. CUT, DAR, DAR core, KIT and base/abci passed their race suites. The full suite completed all 130 tested packages at the node-directory checkpoint.

### Restore cut timing counters and reduce timer overhead

Three red tests found zero elapsed time after 128 sequential merges, 128 mapped nodes and 128 choice unions in each of the combinational and sequential modes. The Go manager now times the two merge paths over the same list-building interval as C, times unions through result placement, and records mapping when enabled. It keeps Go's per-truth duration. Reading elapsed monotonic time relative to manager creation at both ends of an interval reduced a public four-input truth merge from 82.1–82.2 ns to 63.7–66.5 ns in two interleaved eight-million-call runs. A native fixture checked the sequential merge counter over four million iterations. The CUT, RWR and base/abci race suites passed.

On one pinned CPU, batches of 18 i10 proofs took 237.6–241.0 ms after the change and 244.9–259.2 ms before it. A separate C/Go comparison on a shared CPU measured 216.6–221.4 ms per proof in C and 241.9–244.4 ms in Go. Both produced an equivalent 5,051-AND miter. The Go profile attributed 29.6% of cumulative CPU to SAT propagation. Cut allocation accounted for 16.5% of allocated bytes, network object construction for 9.8%, and DEC graph construction and append for another 11.1%. Native gprof put 36.7% of flat samples in propagation and 11.0% in solver search.

### Release AIG update references and reduce deletion snapshots

A red unit test found that clearing the AIG update lists left all deleted and updated objects reachable from their backing arrays. Reset now clears the occupied entries before reusing the buffers. Recursive AND deletion captures its two original fanin pointers on the stack as C does. A test rewires a pin inside the cut-release callback and checks that the recursion visits precisely the original cone. Network teardown scans live slots in place for mark checks and then releases edges, avoiding two directory-sized allocations.

Building and deleting a 1,024-gate chain with a distinct primary input per gate fell from 8,284 to 7,258 allocations and from 1,043,072 to 973,440 bytes. Pinned Go times overlapped at 1.27–1.31 ms before and 1.20–1.29 ms after. C took 142 µs. The 4,096-node alternating deletion fixture allocated 81,920 fewer bytes, although this run was slower at 905–912 µs versus 834–863 µs. The full i10 proof saved about 0.32 MB and 3,017 allocations. Its timing ranges also overlapped at 234–238 ms and 233–242 ms. Both paths produced the same miter. Normal and race suites passed for base/abc, base/abci, RWR and RWT.

### Walk Ivy gates in creation order when the topology permits

Building fanin-first order consumed 4.4% of i10's allocated bytes. The old walk grew both a pointer hash set and an output slice. On 4,096 ordered ANDs it took 685–760 µs and allocated 450,672 bytes in 62 allocations. A trial reserving the output and visiting by dense ID took 60–61 µs and two allocations. The production path checks fanin IDs during the object-directory scan. A forward reference triggers a fanin-first walk from that point with the existing prefix marked. It took 37.9–38.2 µs and one 32,768-byte allocation on the ordered chain. A C fixture collecting the same pointer vector took 5.95 µs.

The walk preserves client traversal marks, original ordering among independent gates, complemented inputs and dangling gates across a replacement with a newer fanin. The fast mapper reuses the order through its two recovery passes. Interleaved batches of 18 i10 proofs took 227.6–233.5 ms after and 233.6–245.2 ms before, with about 2.0 MB fewer allocated bytes per proof and the same 5,051-AND result. Ivy, FRA and base/abci passed their race suites, including the native mapping and simulation comparisons.

### Simulate Ivy nodes from the current graph and honor the word count

Red tests found that rebinding a primary input's simulation record still produced 0x88888888 from its old words. Toggling the gate phase also left the result unchanged. ABC's graph simulator reads the current operands and phases at the call, while the linked-record simulator follows the snapshot from initialization. A native fixture verifies the two distinct outcomes. Another red test gave a one-word manager an extra nonzero output word. Go reported a false counterexample, changed the extra input and output words, and split identical signatures into different classes. Assignment, comparisons, scoring, hashing and output checks now use the manager's word count.

The shared word kernel computes phase masks once per gate. On 1,024 gates with 32 simulation words, the Go graph simulator took 40.0–40.3 µs per round versus 77.0–78.3 µs previously. The linked simulator took 23.2–24.4 µs versus 64.8–70.3 µs, with zero allocations. C measured 9.18 µs and 16.49 µs respectively. The full i10 proof showed no material timing change across interleaved batches of 20 at 226–227 ms. Ivy, FRA and base/abci passed their race suites. The full suite at the cut-timing checkpoint completed all 130 tested packages.

### Group FRAIG signatures without copying every word into a string

A red regression placed two different nonzero signatures and the all-zero signature in the same hash bucket. With the constant node's stored signature changed, Go placed the wrong input in its class. ABC assigns all-zero signatures to the constant representative independently of its stored words. The native fixture verifies every member and representative through four classes. Go now uses ABC's signature hash, checks full signatures on collisions, and releases the temporary bucket links before returning the ordered class list. It keeps its own typed storage for fanouts and class links.

Building 512 pairs from 1,024 inputs with 32 words per signature took 84.9–86.0 µs, 4,864 bytes and one allocation in Go. The string-based version took 297–344 µs, 325,832 bytes and 2,072 allocations. C took 28.5 µs. Twenty-proof i10 batches used about 0.82 MB and 5,270 fewer allocations per proof. Times overlapped at 227–231 ms versus 229–235 ms, with the same 5,051-AND miter. Ivy, FRA and base/abci passed their race suites.

### Issue Ivy node storage as a graph grows

The Go allocator populated its recycle stack with every address in a new 4,096-slot page. It allocated 777,840 bytes for a 16-slot pool. A red test also recycled a value copy of a published input. The next fetch returned the copy even though it lived outside the manager's pages. A slot index now identifies the original address, survives object overwrite and reuse, and permits an ownership check without pointer arithmetic. Clients still get the most recently recycled slot first and retain the same complemented handle across reuse. Tests cover explicit page reservation, LIFO reuse across two pages and invalidation of retained handles at shutdown.

Physical blocks of 32, 128, 512 and 4,096 nodes were tested under the same logical 4,096-slot pages. At 16 slots, 128-node blocks took 6.7–7.0 µs and 21,224 bytes, versus 207–258 µs and 777,840 bytes before. C took 2.37 µs. At 5,000 slots, the same Go variant took 224–230 µs and 820,456 bytes, versus 538–707 µs and 1,400,480 bytes before. C took 267 µs. Full-page physical blocks took 162–163 µs on the larger case but allocated 623,088 bytes on the small one. The 128-node choice cut about 0.73 MB per i10 proof. Full-proof times overlapped at 223.9–224.9 ms after and 224.4–224.6 ms before. Ivy, FRA and base/abci passed normal and race suites after the final ownership checks.

A fresh interleaved C/Go comparison after the Ivy changes measured 213.6–215.5 ms per i10 proof in C and 223.4–228.5 ms in Go. Both produced the equivalent 5,051-AND miter. Each run used 16 proofs on the same CPU.

### Match native decomposition edges and constant graph storage

Red tests found that DEC panicked when the requested node exceeded its 30-bit field. Both packages also retained excess bits on edges assembled through the exported Node field. In KIT, passing one as an AND operand raised a spurious out-of-range panic. The graph APIs now truncate at construction, serialization and insertion as ABC does. The standalone DECO wrappers delegate to DEC. The native graph fixture checks negative inputs, high input bits and the null node directory on both KIT constants.

The constant-graph regression originally found a 50-slot node directory behind each KIT constant. Building a pair of escaping constant graphs now allocates 96 bytes in two allocations, down from 928 bytes in four allocations. Two pinned batches of three million pairs took 45.8–46.3 ns per pair versus 178–190 ns previously. Normal tests passed in DEC, DECO, KIT, RWR and base/abci. DEC, DECO and KIT also passed with the race detector.

### Index KIT nodes and compact decomposition node metadata

Looking up each KIT node scanned the entire node directory. A leaf-depth walk on a 1,024-gate chain took 692 µs in Go and 5.16 µs in C, which locates nodes by pointer subtraction. Go now verifies a cached node index by address and reads the two fanins after one membership check. A caller can overwrite a whole exported node by value, so mismatched indices still fall back to the directory. Tests cover overwrites, copied pointers, growth, released graphs and foreign nodes. The same traversal took 6.36–6.72 µs after the change and allocated nothing. At 64 gates Go fell from 4.20 µs to 442–444 ns. C took 375 ns.

DEC and KIT store their three printing flags in one byte and three five-bit latch counts in three bytes. Tests check the original input phases of AND and OR for every polarity pair. Seven-gate construction fell from 1,360 to 1,008 bytes and 64-gate construction from 6,944 to 4,768 bytes in both packages. Twenty-proof i10 batches allocated about 2.24 MB less per proof with overlapping elapsed times of 228–232 ms after and 226–239 ms before. DEC, DECO, KIT, DAR, RWR and base/abci passed normal and race tests. The full suite at the Ivy allocator checkpoint passed all 130 tested packages.

### Test graph allocation strategies and check cut ownership

Nine experimental layouts varied initial directory capacity, the number of nodes issued per block, and whether slots were prepared up front. Grouping only the leaves lowered an 18-proof i10 batch from about 46.88 to 45.26 MB and saved 14,300 allocations per proof. Packing directories together with the first three or eleven nodes saved 46,600 allocations per proof but used 46.54 MB. It built four-leaf, seven-gate DEC and KIT graphs in about 200 ns and two allocations. Earlier native measurements took 120 ns for DEC and 76 ns for KIT. The end-to-end Go timings overlapped, from 222–233 ms for the baseline to 223–228 ms for the packed variant. These layouts remain experimental.

A 30-proof Go profile attributed 18.1% of allocated bytes to cut creation and 31.0% of cumulative CPU to SAT propagation. A red cut test showed that recycling value copies of two live handles counted four recycles and drove the manager's live count to minus two. Copies of temporary snapshots were also accepted. Each handle now records its allocation address and validates it before changing counters. Creating, merging and recycling a four-leaf cut still used one 128-byte allocation. Two four-million-merge runs took 70.8–72.1 ns versus 71.8–72.3 ns before. The proof kept its 5,051-AND miter and spent 228–229 ms after the fix versus 230–232 ms before. CUT, RWR and base/abci passed normal and race tests.

### Share the rewrite frame's canonical tables

The native rewrite manager borrows four read-only four-input NPN tables from the ABC frame. A dual-manager C fixture confirmed that the tables have identical addresses and the mutable practical-class table and forests are independent. The Go regression initially found four separate canonical arrays per manager. Go now initializes these tables once. Tests mutate a candidate, a leaf binding and a practical-class entry, then stop one manager and evaluate the other.

The startup benchmark constructs and releases the 1,785-node, 222-class library. C averaged 408 µs across 3,000 warm starts. Its profile placed 36% of flat samples in preprocessing and 23% in recursive graph collection. The initial Go heap profile placed 52% of allocated bytes in constructing and appending DEC nodes and 14% in NPN tables. Across two 1,200-start batches, initial Go took 1.97–2.59 ms and about 2.38 MB per manager, falling to 1.39–1.64 ms and 2.05 MB with the frame tables. The i10 proof saved about 0.65 MB per run with overlapping timing ranges. RWR, DEC and base/abci passed normal tests, and RWR and base/abci passed with the race detector.

### Reuse the built-in rewrite library's decomposition descriptions

The first standard manager records the topology and printing flags produced by ABC's decomposition traversal, together with its final visit marks and volume fields. Later standard managers instantiate fresh nodes and class lists from these descriptions. Arbitrary forests still pass through the source traversal. A C fixture produces digest `ddc262750c62bab1` across every forest node's traversal state, class order, root, input edge and printing flag. Four concurrent Go starts matched it under the race detector.

Exact-size reconstruction took 0.96–1.11 ms per manager in two 1,500-start batches versus 1.35–1.42 ms rebuilding the graphs. Allocations fell from 17,873 to 5,366 and bytes from 2.05 MB to 1.65 MB. A heap profile with every allocation sampled showed 177 KiB of additional persistent descriptions. Across 12 fresh processes per version, the median first startup rose from 2.07 to 2.29 ms. In 22-proof i10 batches, the descriptions saved about 0.79 MB and 24,400 allocations per proof. Times overlapped at 223–224 ms and 224–226 ms with the same miter. Normal and race tests passed for RWR and base/abci.

### Allocate rewrite truth-table chains on first insertion

The native library fixture has 134 occupied keys in a domain of 65,536 four-input functions. Go previously allocated a 512 KiB pointer directory per manager. Experiments with pages of 16, 64, 128, 256 and 1,024 pointers found the lowest startup allocation at 64. Its 27 occupied pages and outer index use 22,016 bytes before allocator rounding. Startup allocated 1.145 MB versus 1.646 MB and took 0.84–0.91 ms versus 1.49–1.72 ms of process CPU during the same interleaved experiment. The paged index adds 27 allocations per manager.

Boundary tests check both sides of a page, the last key, chain order, empty reads and stopped managers. Two initially failing checks caught an invalid negative key returning an empty chain and a stopped manager still accepting a query. The i10 proof saved about 1.0 MB per run. Two pinned experiments had wide timing ranges on the shared host, from 277–373 ms for the new version and 299–397 ms for the previous version. RWR and base/abci passed normal and race tests.

### Keep RWT canonical tables live for exactly the manager lifetime

A concurrent regression reproduced a live residual-rewriting manager with four empty canonical tables. A frame stop could run between initialization and the separate lock acquisition that copied its slices. The manager now acquires all four references under the initialization lock. Another red test found a stopped manager still holding 320 KiB of table storage after frame shutdown. Teardown now releases its references. Other live managers retain their original tables, and a later frame start creates a new generation with identical truth functions.

The concurrent stop/start and release regressions passed 20 repetitions with the race detector, followed by ten more after adding a frame-generation check. RWT and base/abci passed their normal and race suites.

### Store DEC nodes in directly indexed blocks

DEC now keeps the first nodes in a slice that stays at its original address. Growth uses indexed blocks of 32 nodes. Exact-size construction puts the remainder first so the following blocks are full. This removes the per-node pointer directory and preserves saved pointers through growth and teardown. Tests exercise ten starting sizes, unused slots, block boundaries, graph indices and retained payloads. A trial that supplied truth words through a callback failed the existing zero-allocation test. Dispatching once on graph representation keeps the scratch words on the stack.

Seven-gate construction took 157–165 ns, 656 bytes and two allocations, down from 380–386 ns, 1,008 bytes and twelve allocations. For 64 gates Go took 1.19–1.47 µs and six allocations versus 2.28 µs and 70 allocations. Native construction took 120 ns and 374 ns respectively. In 22-proof i10 batches Go took 231–234 ms versus 247–249 ms and saved about 0.26 MB and 5,266 allocations per proof. An interleaved native comparison on the same core took 216–229 ms in C and 233–241 ms in Go. Both produced the equivalent 5,051-AND miter. DEC, DECO, RWR, RWT and base/abci passed normal and race tests.

### Bound borrowed word records before sharing mapper storage

The i10 mapped-SAT run took 45.30 ms in Go and allocated 11.59 MB in 99,461 allocations. Native C took 25.85 ms and solved with zero conflicts. Go's heap profile attributed 49.8% of allocated bytes to interning truth functions, including duplicate keys and independently allocated records. Both CPU profiles show cut enumeration and irredundant sum-of-products computation among the largest costs.

The native mapper keeps truths in fixed-width paged records. A red test against Go's existing paged vector found that appending a word to one borrowed record overwrote the first word of the next hashed record. Limiting each slice's capacity to its record width isolates append and still lets callers edit the original words in place and rehash. The vector and mapped-CNF suites passed normally and under the race detector. The preceding full-module run at the RWT frame checkpoint passed all 130 packages with tests.

### Intern MF cut functions in fixed-width word pages

The Go mapper now uses the existing Vec_Mem implementation of the native truth store. Its hash chains contain integer IDs and compare the stored words in place. Enumeration reads each candidate directly from the store. At completion a single exact-size slice exposes the public truth views and the hash index is released. The initial bucket and page reservations are bounded for small and sparse networks.

Three interleaved experiments distinguished the storage costs. Using pages alone reduced allocated bytes from 11.59 to 8.86 MB per i10 solve. Reserving initial slice capacity reduced that to 8.72 MB. Building public views only at completion brought it to 7.45 MB and 84,225 allocations, down about 15,236 allocations per solve. The production binary took 43.8–45.9 ms against 46.9–52.6 ms before the change in 160-solve batches. The new profile attributes 16.4% of allocated bytes to truth interning and its callees, down from 49.8%. Native C took 25.85 ms in the preceding comparison.

Tests hold a truth across page and hash growth, distinguish the last word, verify public slice bounds at all six LUT widths, and check that generation releases the index. MF and vector normal and race tests passed, including comparison of every i10 CNF clause and literal in native order. Targeted GIA mapper and Satoko tests also passed.

### Reuse exact-area reference scratch across MF nodes

The post-interning heap profile attributed 7.5% of allocated bytes to a 32-element array in best-cut selection. Go put this array on the heap at every gate, including area-flow rounds. An exact-area trial on a five-level binary tree reproduced two allocations per visit. Mapping now passes a reusable slice between serial gate visits and rounds. The test checks zero allocations after warmup, restoration of all 62 fanin reference counts and that the area-flow path needs no scratch.

The i10 solve fell from 7.45 to 6.91 MB and from 84,225 to 80,048 allocations. Two 120-solve batches took 40.3–41.0 ms versus 41.3–42.2 ms with per-node scratch. The mapped-CNF native differential and MF package tests passed normally and under the race detector.

### Copy retained MF cuts after ranking

A failing choice-network test showed that editing a saved sibling's leaves also changed the selected cut of a different node. Native Mf_ManSaveCuts copies the leaf IDs of each retained cut. Go now ranks a candidate before copying its leaves, which also avoids copying candidates displaced immediately. Insertion preserves tie order and moves the retained suffix once. The published cut set clears unused slots.

A trial that reused displaced leaf arrays reached 50,069 allocations but had no consistent speed benefit over copying just the retained leaves. The chosen version ran i10 in 38.1–38.3 ms versus 40.3 ms, in matched 150-solve batches. Allocation fell from 6.91 MB and 80,048 allocations to 6.26 MB and 59,008. The choice ownership test, a full-set rejection and tie-order test, and the complete MF suite passed normally and under the race detector. The full-module truth-storage checkpoint passed all 130 tested packages.

### Memoize three-input ISOP counts and result words

For three variables, each of eight assignments is off, unspecified or on. There are 6,561 possible ISOP subproblems. The C fixture computes a checksum over all of them, including the exact result word and the caller's cumulative cube count. The Go test checks the same checksum from eight concurrent callers. Count-only queries can reuse a result after its first computation. Explicit covers continue to come from the ordered recursive implementation.

Trials compared a 256-byte table for two-variable cofactors, a dense three-variable index, a compact ternary index, and eager, per-row and per-entry initialization. Filling the complete compact table on first use took a median 308 µs on the cold query fixture. Populating one entry at a time took 9 µs including first-call overhead. The selected table reserves 26,244 bytes for atomic entries and 512 bytes for mask indices, and accepts concurrent computation of the same deterministic answer. All 6,561 warm pairs averaged 5.4–5.6 ns each in the production Go binary, versus 38.5–38.7 ns previously and 23.4 ns for the C fixture. Each produced the checksum `a40a7d1f15a0d833`.

The i10 pipeline took 36.2–36.6 ms versus 40.2–41.2 ms before this change with 6.26 MB per solve. Four same-core, interleaved 30-solve batches gave 36.1–36.8 ms for Go and 21.8–22.5 ms for C. Utility and MF package tests passed normally and under the race detector, including the exhaustive checksum from a cold process and exact native i10 CNF comparison.

### Restore ordered cut signatures in the MF enumeration loop

Native MF computes a 64-bit leaf signature once per fanin cut and carries it with each ranked candidate. The Go port now computes those masks before the Cartesian product, rejects unions with too many set bits, and checks for possible containment before comparing the actual sorted leaves. The masks move with the ranked cuts and disappear when enumeration finishes. Truth minimization can remove a leaf, so it also recomputes a changed mask. Disabling that update reproduced an unwanted second cut for `a AND (a OR b)`. Tests also cover distinct leaves with identical signatures and moves through the ranked list.

In 150-solve batches, preparing fanin masks alone took 34.2–34.4 ms, keeping ranked-candidate masks alone took 31.7–32.0 ms, and combining both took 30.3–30.4 ms against 35.3–35.9 ms before. The production binary repeated at 30.2–30.3 ms versus 35.4 ms, with 6.26 MB and 59,008 allocations. Four separate same-core native comparisons took 21.5–22.9 ms in C and 32.1–33.7 ms in Go. MF tests and the native CNF differential passed normally and under the race detector.

### Reuse data cuts across MUX control candidates

Each MF data-cut pair now merges once. Every candidate control cut merges its own leaves with that saved pair, in the original enumeration order. The candidate gets a separate buffer because support minimization can edit it in place. The MUX test checks both retained control alternatives against every primary-input assignment at LUT widths three through eight.

In 160-solve i10 batches the production version took 29.08–29.09 ms versus 29.83–29.95 ms. Hoisting only the small array of cut arguments had little effect, and adding leaf endpoint checks after signature filtering did not improve the result. Four same-core comparisons measured 28.8–29.3 ms in Go and 21.1–21.3 ms in C. Both solve the i10 output disjunction. The native clause-order differential and MF race tests passed. The full-module ISOP checkpoint passed all 130 tested packages.

### Specialize seven- and eight-input ISOP cofactors

Wide ISOP and its callees accounted for 17.4% of the i10 mapped-SAT CPU samples. The Go implementation now performs the C seven-input split on individual words and the eight-input split on pairs of words. Count and cover entry points share the fixed-size operations. The mapper calculates small-support CNF costs from one word and feeds full-width records directly to the wide cost function.

A new native fixture hashes the ordered covers, the caller's initial two cubes, the counts and all output words for 96 functions with six through eight inputs. Go matches the C digest `c5fa775abb4222c8` with distinct buffers and with output sharing either input. Both Go versions allocated zero bytes in this workload. The final Go binary took 2.26–2.30 µs per table against 2.30–2.36 µs previously. C took 2.48–2.62 µs in two runs. Initial i10 comparisons put the combined change at 28.1 ms versus 29.2–29.3 ms. The production repeat on the shared host gave 29.2–30.6 ms versus 30.4–31.2 ms. All LUT storage widths have direct cost versus cover tests. Utility and MF normal and race suites passed, including the native i10 CNF comparison.

### Save only the final MF cut leaves for each node

Mapping now reuses a bounded array of candidate records throughout enumeration. Ranking moves byte-sized slot indices, and publication copies the surviving cuts into an exact-size array with one shared leaf allocation for that node. Each cut's capacity ends at its own final leaf. A regression test caught two deliberate faults. Saving the temporary leaves corrupted an earlier node during the next enumeration, and allowing spare slice capacity let append overwrite the neighboring cut. Empty cuts retain a nil leaf slice.

Pointer order, byte-index order and pointer-free candidate records all took about 27.0–27.2 ms versus 28.4 ms in the same i10 experiment. The production byte-index version saved 0.55 MB and 25,318 allocations, down to 5.71 MB and 33,690 allocations per solve. Copying the entire incoming cut struct made two scratch buffers per gate escape even after replacing its leaves. Copying the published fields explicitly recovered those allocations. The final run took 27.6–29.0 ms versus 29.2–29.8 ms. A separate busier-core comparison took 41.6–42.4 ms in Go and 32.2–32.6 ms in C. All MF tests passed normally and under the race detector.

### Match and accelerate paged-vector record hashing

The native fixture exposed a public hash-key difference at small capacities. Two regression tests failed before the fix. The C helper chooses one bucket for hints zero and one, and three buckets for hint two. Go had chosen two buckets in all three cases. The fixed vector matches the C digest `fce2fa148b5a01ac` over 128 records at six widths and seven initial capacities, including repeated growth and duplicate insertion.

Record hashing now applies the eight native multipliers directly in four-word groups. Interleaved 1,200-iteration Go runs took 106–110 ns per record versus 117–124 ns before, with identical allocations. A separate native comparison took 90–92 ns per record in Go and 93–115 ns in C. The i10 mapped-SAT cross-check took 27.9–28.2 ms versus 28.3–28.7 ms. Both affected packages passed normal and race tests. The preceding wide-ISOP full-module run passed all 130 tested packages.

### Reuse MF truth operations across different leaf IDs

Two candidates with the same gate kind, signed input functions and relative input positions have the same output truth. During enumeration, MF now caches that result and the positions of the leaves retained by support minimization. A hit selects the current candidate's leaf IDs. The cache is local to the mapping and is released before flow refinement. A one-slot test forces collisions between AND, XOR and MUX, repeats each operation with different IDs, and checks every assignment in both object-ID modes at LUT widths six and eight. The normal and race MF suites passed, including the i10 native clause-order differential.

The 1,024-entry trial hit on 17,020 of 44,269 candidate operations. A 4,096-entry limit gave a better time with 96 KiB of scratch, while 16,384 entries used four times as much scratch for similar timing. Small networks size the table to their object count rounded up to a power of two. The production i10 run took 24.29–24.49 ms versus 27.05–27.19 ms in two interleaved 160-solve batches per binary. Allocation rose from 5.71 MB to 5.81 MB and from 33,690 to 33,691 allocations. Four additional same-core batches gave 24.27–24.54 ms for Go and 21.21–21.57 ms for C.

### Release Satoko rollback storage and preserve unit counters

The mapped-SAT allocation profile put 47% of allocated objects in clause storage and watch registration. While inspecting clause lifetime, a regression test found that rollback retained pointers to removed clauses in the backing array. It also raised both original and learnt literal counts from two to three when restoring unit clauses. Satoko excludes units from both counters on insertion. The fix clears the removed pointers and applies the same counting rule during rollback. The regression first failed on both checks and now passes with the normal and race Satoko suites. The full module run at the vector-hash checkpoint also finished green across all 130 tested packages.
